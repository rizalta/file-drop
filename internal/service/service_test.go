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

		code, err := s.CreateDrop(ctx, params)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if len(code) != codeLen {
			t.Errorf("expected code length %d, got %d (%q)", codeLen, len(code), code)
		}

		drop, exists := mq.drops[code]
		if !exists {
			t.Fatalf("expected drop with code %q to exist in DB mock", code)
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

		code, err := s.CreateDrop(ctx, params)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		drop, exists := mq.drops[code]
		if !exists {
			t.Fatalf("expected drop with code %q to exist in DB mock", code)
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
