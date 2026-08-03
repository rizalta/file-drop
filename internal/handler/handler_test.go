package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rizalta/file-drop/internal/repo"
	"github.com/rizalta/file-drop/internal/service"
)

type mockService struct {
	db      map[string]*repo.Drop
	storage map[string][]byte
}

func newMockService() *mockService {
	return &mockService{
		db:      make(map[string]*repo.Drop),
		storage: make(map[string][]byte),
	}
}

func (m *mockService) CreateDrop(ctx context.Context, params service.CreateDropParams) (string, error) {
	id := "test123"
	storedName := uuid.NewString()

	m.db[id] = &repo.Drop{
		ID:                id,
		Filename:          params.Filename,
		StoredName:        storedName,
		FileSize:          int64(params.FileSize),
		MimeType:          params.MimeType,
		IsText:            params.IsText,
		TextContent:       pgtype.Text{String: params.TextContent, Valid: params.IsText},
		BurnAfterDownload: params.BurnAfterDownload,
		DownloadCount:     0,
		ExpiresAt:         time.Now().Add(1 * time.Hour),
		CreatedAt:         time.Now(),
	}

	if !params.IsText && params.Reader != nil {
		data, _ := io.ReadAll(params.Reader)
		m.storage[storedName] = data
	}

	return id, nil
}

func (m *mockService) CleanupExpiredDrops(ctx context.Context) error   { return nil }
func (m *mockService) DeleteDrop(ctx context.Context, id string) error { return nil }
func (m *mockService) GetDrop(ctx context.Context, id string) (*repo.Drop, io.ReadCloser, error) {
	return nil, nil, nil
}
func (m *mockService) ListActiveDrops(ctx context.Context) ([]repo.Drop, error) { return nil, nil }

func TestUploadDrop(t *testing.T) {
	t.Run("successful file upload", func(t *testing.T) {
		body := &bytes.Buffer{}
		mw := multipart.NewWriter(body)

		filepart, _ := mw.CreateFormFile("file", "hello.txt")
		_, _ = filepart.Write([]byte("hello world. how are you"))

		_ = mw.WriteField("expires_in", "1h")
		_ = mw.WriteField("burn_after_download", "true")
		_ = mw.Close()

		req := httptest.NewRequest("POST", "/api/v1/upload", body)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		rec := httptest.NewRecorder()

		m := newMockService()
		h := NewHandler(m)

		h.UploadDrop(rec, req)

		if rec.Code != 201 {
			t.Errorf("expected status 201 but got %d", rec.Code)
		}

		var res UploadDropRes
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to unmarshall response: %v", err)
		}

		if res.ID != "test123" {
			t.Errorf("expected \"test123\" got %s", res.ID)
		}
	})

	t.Run("successful text snippet upload", func(t *testing.T) {
		body := &bytes.Buffer{}
		mw := multipart.NewWriter(body)

		_ = mw.WriteField("content", "this is a text snippet")
		_ = mw.WriteField("expires_in", "24h")
		_ = mw.Close()

		req := httptest.NewRequest("POST", "/api/v1/upload", body)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		rec := httptest.NewRecorder()

		m := newMockService()
		h := NewHandler(m)

		h.UploadDrop(rec, req)

		if rec.Code != 201 {
			t.Errorf("expected status 201 but got %d", rec.Code)
		}

		var res UploadDropRes
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to unmarshall response: %v", err)
		}

		if res.ID != "test123" {
			t.Errorf("expected \"test123\" got %s", res.ID)
		}

		drop, exists := m.db["test123"]
		if !exists || !drop.IsText || drop.TextContent.String != "this is a text snippet" {
			t.Errorf("text snippet metadata not correctly saved in service mock")
		}
	})

	t.Run("missing file and text content returns 400", func(t *testing.T) {
		body := &bytes.Buffer{}
		mw := multipart.NewWriter(body)
		_ = mw.WriteField("expires_in", "1h")
		_ = mw.Close()

		req := httptest.NewRequest("POST", "/api/v1/upload", body)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		rec := httptest.NewRecorder()

		m := newMockService()
		h := NewHandler(m)

		h.UploadDrop(rec, req)

		if rec.Code != 400 {
			t.Errorf("expected status 400 but got %d", rec.Code)
		}
	})
}
