package service

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/rizalta/file-drop/internal/repo"
)

type mockStorage struct {
	files map[string][]byte
}

func newMockStorage() *mockStorage {
	return &mockStorage{files: make(map[string][]byte)}
}

func (ms *mockStorage) Save(storedName string, r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	ms.files[storedName] = data
	return nil
}

func (ms *mockStorage) Get(storedName string) (io.ReadCloser, error) {
	data, ok := ms.files[storedName]
	if !ok {
		return nil, os.ErrNotExist
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

func (ms *mockStorage) Delete(storedName string) error {
	if _, ok := ms.files[storedName]; !ok {
		return os.ErrNotExist
	}

	delete(ms.files, storedName)
	return nil
}

type mockQuerier struct {
	drops map[string]*repo.Drop
}

func newMockQuerier() *mockQuerier {
	return &mockQuerier{drops: make(map[string]*repo.Drop)}
}

func (mq *mockQuerier) CreateDrop(ctx context.Context, params repo.CreateDropParams) (repo.Drop, error) {
	drop := repo.Drop{
		ID:                params.ID,
		Filename:          params.Filename,
		StoredName:        params.StoredName,
		FileSize:          params.FileSize,
		MimeType:          params.MimeType,
		IsText:            params.IsText,
		TextContent:       params.TextContent,
		BurnAfterDownload: params.BurnAfterDownload,
		ExpiresAt:         params.ExpiresAt,
		CreatedAt:         time.Now().UTC(),
	}

	mq.drops[drop.ID] = &drop
	return drop, nil
}

func (mq *mockQuerier) GetDropByID(ctx context.Context, id string) (repo.Drop, error) {
	drop, ok := mq.drops[id]
	if !ok {
		return repo.Drop{}, os.ErrNotExist
	}
	return *drop, nil
}

func (mq *mockQuerier) ListActiveDrops(ctx context.Context) ([]repo.Drop, error) {
	var list []repo.Drop
	now := time.Now().UTC()
	for _, drop := range mq.drops {
		if drop.ExpiresAt.After(now) {
			list = append(list, *drop)
		}
	}
	return list, nil
}

func (mq *mockQuerier) IncrementDownloadCount(ctx context.Context, id string) error {
	drop, ok := mq.drops[id]
	if !ok {
		return os.ErrNotExist
	}
	drop.DownloadCount++
	return nil
}

func (mq *mockQuerier) DeleteDrop(ctx context.Context, id string) (string, error) {
	drop, ok := mq.drops[id]
	if !ok {
		return "", os.ErrNotExist
	}
	storedName := drop.StoredName
	delete(mq.drops, id)
	return storedName, nil
}

func (mq *mockQuerier) DeleteExpiredDrops(ctx context.Context) ([]repo.DeleteExpiredDropsRow, error) {
	var deleted []repo.DeleteExpiredDropsRow
	now := time.Now().UTC()
	for id, drop := range mq.drops {
		if !drop.ExpiresAt.After(now) {
			deleted = append(deleted, repo.DeleteExpiredDropsRow{
				ID:         drop.ID,
				StoredName: drop.StoredName,
			})
			delete(mq.drops, id)
		}
	}
	return deleted, nil
}

func TestCreateDrop(t *testing.T) {
	ctx := context.Background()

	t.Run("successful file upload", func(t *testing.T) {
		ms := newMockStorage()
		mq := newMockQuerier()
		s := NewService(mq, ms)

		fileContent := "hello world file drop"
		params := CreateDropParams{
			Filename:          "test.txt",
			FileSize:          len(fileContent),
			MimeType:          "text/plain",
			IsText:            false,
			BurnAfterDownload: false,
			ExpiresIn:         "1h",
			Reader:            bytes.NewBufferString(fileContent),
		}

		id, err := s.CreateDrop(ctx, params)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if len(id) != idLen {
			t.Errorf("expected id length %d, got %d (%q)", idLen, len(id), id)
		}

		drop, exists := mq.drops[id]
		if !exists {
			t.Fatalf("expected drop with id %q to exist in DB mock", id)
		}

		if drop.Filename != "test.txt" {
			t.Errorf("expected filename 'test.txt', got %q", drop.Filename)
		}

		storedData, blobExists := ms.files[drop.StoredName]
		if !blobExists {
			t.Fatalf("expected stored blob %q to exist in storage mock", drop.StoredName)
		}

		if string(storedData) != fileContent {
			t.Errorf("expected stored blob content %q, got %q", fileContent, string(storedData))
		}
	})

	t.Run("successful text snippet upload", func(t *testing.T) {
		ms := newMockStorage()
		mq := newMockQuerier()
		s := NewService(mq, ms)

		textContent := "this is a text snippet drop"
		params := CreateDropParams{
			Filename:          "",
			FileSize:          len(textContent),
			MimeType:          "text/plain",
			IsText:            true,
			TextContent:       textContent,
			BurnAfterDownload: true,
			ExpiresIn:         "24h",
			Reader:            nil,
		}

		id, err := s.CreateDrop(ctx, params)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		drop, exists := mq.drops[id]
		if !exists {
			t.Fatalf("expected drop with id %q to exist in DB mock", id)
		}

		if !drop.IsText {
			t.Errorf("expected IsText to be true")
		}

		if drop.StoredName != "" {
			t.Errorf("expected StoredName to be empty string for text snippet, got %q", drop.StoredName)
		}

		if !drop.TextContent.Valid || drop.TextContent.String != textContent {
			t.Errorf("expected TextContent %q, got %q", textContent, drop.TextContent.String)
		}

		if len(ms.files) != 0 {
			t.Errorf("expected 0 files in blob storage for text snippet, found %d", len(ms.files))
		}
	})
}

func TestGetDrop(t *testing.T) {
	ctx := context.Background()

	t.Run("get existing file drop", func(t *testing.T) {
		ms := newMockStorage()
		mq := newMockQuerier()
		s := NewService(mq, ms)

		fileContent := "sample data stream"
		id, err := s.CreateDrop(ctx, CreateDropParams{
			Filename:  "sample.txt",
			FileSize:  len(fileContent),
			MimeType:  "text/plain",
			ExpiresIn: "1h",
			Reader:    bytes.NewBufferString(fileContent),
		})
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		drop, rc, err := s.GetDrop(ctx, id, true)
		if err != nil {
			t.Fatalf("expected GetDrop to succeed, got %v", err)
		}
		defer func() {
			if rc != nil {
				_ = rc.Close()
			}
		}()

		if drop.Filename != "sample.txt" {
			t.Errorf("expected filename sample.txt, got %s", drop.Filename)
		}

		data, _ := io.ReadAll(rc)
		if string(data) != fileContent {
			t.Errorf("expected stream data %q, got %q", fileContent, string(data))
		}
	})

	t.Run("burn after download self-destructs drop from DB and storage", func(t *testing.T) {
		ms := newMockStorage()
		mq := newMockQuerier()
		s := NewService(mq, ms)

		id, err := s.CreateDrop(ctx, CreateDropParams{
			Filename:          "secret.txt",
			FileSize:          6,
			MimeType:          "text/plain",
			BurnAfterDownload: true,
			ExpiresIn:         "1h",
			Reader:            bytes.NewBufferString("secret"),
		})
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		drop, rc, err := s.GetDrop(ctx, id, true)
		if err != nil {
			t.Fatalf("first GetDrop failed: %v", err)
		}
		_ = rc.Close()
		_ = drop

		_, _, err = s.GetDrop(ctx, id, true)
		if err == nil {
			t.Errorf("expected 2nd GetDrop on burned file to fail, but succeeded")
		}

		if len(ms.files) != 0 {
			t.Errorf("expected 0 files in storage mock after burn-after-download, found %d", len(ms.files))
		}

		if len(mq.drops) != 0 {
			t.Errorf("expected 0 drops in DB mock after burn-after-download, found %d", len(mq.drops))
		}
	})
}

func TestListActiveDrops(t *testing.T) {
	ctx := context.Background()

	t.Run("list returns active drops", func(t *testing.T) {
		ms := newMockStorage()
		mq := newMockQuerier()
		s := NewService(mq, ms)

		_, _ = s.CreateDrop(ctx, CreateDropParams{
			Filename:  "file1.txt",
			FileSize:  5,
			ExpiresIn: "1h",
			Reader:    bytes.NewBufferString("file1"),
		})

		drops, err := s.ListActiveDrops(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(drops) != 1 {
			t.Errorf("expected 1 active drop, got %d", len(drops))
		}
	})
}

func TestDeleteDrop(t *testing.T) {
	ctx := context.Background()

	t.Run("manual delete removes file drop from DB and storage", func(t *testing.T) {
		ms := newMockStorage()
		mq := newMockQuerier()
		s := NewService(mq, ms)

		id, _ := s.CreateDrop(ctx, CreateDropParams{
			Filename:  "todelete.txt",
			FileSize:  4,
			ExpiresIn: "1h",
			Reader:    bytes.NewBufferString("data"),
		})

		err := s.DeleteDrop(ctx, id)
		if err != nil {
			t.Fatalf("expected DeleteDrop to succeed, got %v", err)
		}

		if len(mq.drops) != 0 {
			t.Errorf("expected DB drop record to be deleted")
		}

		if len(ms.files) != 0 {
			t.Errorf("expected blob file to be deleted from storage")
		}
	})

	t.Run("manual delete handles text snippet drop without storage error", func(t *testing.T) {
		ms := newMockStorage()
		mq := newMockQuerier()
		s := NewService(mq, ms)

		id, _ := s.CreateDrop(ctx, CreateDropParams{
			IsText:      true,
			TextContent: "text snippet to delete",
			ExpiresIn:   "1h",
		})

		err := s.DeleteDrop(ctx, id)
		if err != nil {
			t.Fatalf("expected DeleteDrop on text snippet to succeed, got %v", err)
		}

		if len(mq.drops) != 0 {
			t.Errorf("expected DB drop record to be deleted")
		}
	})
}

func TestCleanupExpiredDrops(t *testing.T) {
	ctx := context.Background()

	t.Run("cleanup removes expired drops from DB and storage", func(t *testing.T) {
		ms := newMockStorage()
		mq := newMockQuerier()
		s := NewService(mq, ms)

		id, _ := s.CreateDrop(ctx, CreateDropParams{
			Filename:  "expired.txt",
			FileSize:  7,
			ExpiresIn: "1h",
			Reader:    bytes.NewBufferString("expired"),
		})

		drop := mq.drops[id]
		drop.ExpiresAt = time.Now().UTC().Add(-10 * time.Minute)

		err := s.CleanupExpiredDrops(ctx)
		if err != nil {
			t.Fatalf("expected CleanupExpiredDrops to succeed, got %v", err)
		}

		if len(mq.drops) != 0 {
			t.Errorf("expected expired drop to be removed from DB")
		}

		if len(ms.files) != 0 {
			t.Errorf("expected expired blob file to be removed from storage")
		}
	})
}
