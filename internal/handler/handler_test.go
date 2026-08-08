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

	"github.com/go-chi/chi/v5"
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

func (m *mockService) CleanupExpiredDrops(ctx context.Context) error { return nil }

func (m *mockService) DeleteDrop(ctx context.Context, id string) error {
	drop, ok := m.db[id]
	if !ok {
		return service.ErrDropNotFound
	}
	delete(m.db, id)
	if drop.StoredName != "" {
		delete(m.storage, drop.StoredName)
	}
	return nil
}

func (m *mockService) GetDrop(ctx context.Context, id string, isDownload bool) (*repo.Drop, io.ReadCloser, error) {
	drop, ok := m.db[id]
	if !ok {
		return nil, nil, service.ErrDropNotFound
	}
	var rc io.ReadCloser
	if !drop.IsText && drop.StoredName != "" {
		data := m.storage[drop.StoredName]
		rc = io.NopCloser(bytes.NewReader(data))
	}
	return drop, rc, nil
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

func TestDownloadDrop(t *testing.T) {
	ctx := context.Background()

	t.Run("download raw file stream", func(t *testing.T) {
		m := newMockService()
		h := NewHandler(m)

		fileContent := "binary file payload data"
		_, _ = m.CreateDrop(ctx, service.CreateDropParams{
			Filename: "file.txt",
			FileSize: len(fileContent),
			MimeType: "text/plain",
			Reader:   bytes.NewBufferString(fileContent),
		})

		req := httptest.NewRequest("GET", "/f/test123?download=true", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "test123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.GetDrop(rec, req)

		if rec.Code != 200 {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		if rec.Header().Get("Content-Disposition") != `attachment; filename="file.txt"` {
			t.Errorf("unexpected Content-Disposition: %s", rec.Header().Get("Content-Disposition"))
		}

		if rec.Body.String() != fileContent {
			t.Errorf("expected body %q, got %q", fileContent, rec.Body.String())
		}
	})

	t.Run("download JSON preview", func(t *testing.T) {
		m := newMockService()
		h := NewHandler(m)

		_, _ = m.CreateDrop(ctx, service.CreateDropParams{
			Filename: "file.txt",
			FileSize: 10,
			MimeType: "text/plain",
			Reader:   bytes.NewBufferString("1234567890"),
		})

		req := httptest.NewRequest("GET", "/f/test123", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "test123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.GetDrop(rec, req)

		if rec.Code != 200 {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		var res DownloadDropRes
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to unmarshal JSON preview: %v", err)
		}

		if res.Filename != "file.txt" {
			t.Errorf("expected filename file.txt, got %s", res.Filename)
		}
	})

	t.Run("non existent drop returns 404", func(t *testing.T) {
		m := newMockService()
		h := NewHandler(m)

		req := httptest.NewRequest("GET", "/f/missing", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "missing")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.GetDrop(rec, req)

		if rec.Code != 404 {
			t.Errorf("expected status 404, got %d", rec.Code)
		}
	})
}

func TestDeleteDrop(t *testing.T) {
	ctx := context.Background()

	t.Run("successful drop deletion", func(t *testing.T) {
		m := newMockService()
		h := NewHandler(m)

		_, _ = m.CreateDrop(ctx, service.CreateDropParams{
			Filename: "file.txt",
			FileSize: 5,
			Reader:   bytes.NewBufferString("hello"),
		})

		req := httptest.NewRequest("DELETE", "/api/v1/files/test123", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "test123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.DeleteDrop(rec, req)

		if rec.Code != 200 {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		if len(m.db) != 0 {
			t.Errorf("expected drop to be removed from mock DB")
		}
	})

	t.Run("deleting non existent drop returns 404", func(t *testing.T) {
		m := newMockService()
		h := NewHandler(m)

		req := httptest.NewRequest("DELETE", "/api/v1/files/missing", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "missing")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.DeleteDrop(rec, req)

		if rec.Code != 404 {
			t.Errorf("expected status 404, got %d", rec.Code)
		}
	})
}
