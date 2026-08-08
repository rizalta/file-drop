package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rizalta/file-drop/internal/repo"
	"github.com/rizalta/file-drop/internal/service"
)

type DropsService interface {
	CleanupExpiredDrops(ctx context.Context) error
	CreateDrop(ctx context.Context, params service.CreateDropParams) (string, error)
	DeleteDrop(ctx context.Context, id string) error
	GetDrop(ctx context.Context, id string, isDownload bool) (*repo.Drop, io.ReadCloser, error)
	ListActiveDrops(ctx context.Context) ([]repo.Drop, error)
}

type handler struct {
	dropsService DropsService
}

func NewHandler(s DropsService) *handler {
	return &handler{s}
}

type UploadDropRes struct {
	ID string `json:"id"`
}

func (h *handler) UploadDrop(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		errResponse(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	burnAfterDownload := r.FormValue("burn_after_download") == "true"
	expiresIn := r.FormValue("expires_in")
	textContent := strings.TrimSpace(r.FormValue("content"))

	file, header, err := r.FormFile("file")
	var reader io.Reader
	var filename, mimeType string
	var fileSize int
	isText := false

	if err == nil {
		defer func() {
			_ = file.Close()
		}()
		filename = header.Filename
		mimeType = header.Header.Get("Content-Type")
		fileSize = int(header.Size)
		reader = file
	} else if textContent != "" {
		mimeType = "text/plain; charset=utf-8"
		isText = true
		fileSize = len(textContent)
	} else {
		errResponse(w, "File or text content required", http.StatusBadRequest)
		return
	}

	params := service.CreateDropParams{
		Filename:          filename,
		FileSize:          fileSize,
		MimeType:          mimeType,
		IsText:            isText,
		TextContent:       textContent,
		BurnAfterDownload: burnAfterDownload,
		ExpiresIn:         expiresIn,
		Reader:            reader,
	}

	id, err := h.dropsService.CreateDrop(r.Context(), params)
	if err != nil {
		if errors.Is(err, service.ErrInvalidExpiry) {
			errResponse(w, "Invalid Expiry", http.StatusBadRequest)
			return
		} else if errors.Is(err, service.ErrInvalidReader) {
			errResponse(w, "Invalid file", http.StatusBadRequest)
			return
		}
		errResponse(w, "Failed to upload drop", http.StatusInternalServerError)
		return
	}

	res := UploadDropRes{id}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(res)
}

type DownloadDropRes struct {
	Filename      string    `json:"filename"`
	FileSize      int       `json:"file_size"`
	MimeType      string    `json:"mime_type"`
	IsText        bool      `json:"is_text"`
	DownloadCount int       `json:"download_count"`
	CreatedAt     time.Time `json:"created_at"`
}

func (h *handler) GetDrop(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	if id == "" {
		errResponse(w, "ID required", http.StatusBadRequest)
		return
	}

	isDownload := r.URL.Query().Get("download") == "true"

	drop, rc, err := h.dropsService.GetDrop(r.Context(), id, isDownload)
	if err != nil {
		if errors.Is(err, service.ErrDropNotFound) {
			errResponse(w, "Drop not found", http.StatusNotFound)
			return
		}
		errResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if rc != nil {
		defer func() {
			_ = rc.Close()
		}()
	}

	if isDownload {
		if drop.IsText {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			if _, err := w.Write([]byte(drop.TextContent.String)); err != nil {
				log.Printf("failed to write text content of id: %s, %v", id, err)
			}
			return
		}

		if drop.MimeType == "" {
			drop.MimeType = "application/octet-stream"
		}

		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, drop.Filename))
		w.Header().Set("Content-Type", drop.MimeType)
		w.Header().Set("Content-Length", strconv.FormatInt(drop.FileSize, 10))
		if _, err := io.Copy(w, rc); err != nil {
			log.Printf("Failed to download for file: %s, %v", id, err)
		}
		return
	}

	res := DownloadDropRes{
		Filename:      drop.Filename,
		FileSize:      int(drop.FileSize),
		MimeType:      drop.MimeType,
		IsText:        drop.IsText,
		DownloadCount: int(drop.DownloadCount),
		CreatedAt:     drop.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("failed to encode response")
	}
}

func (h *handler) DeleteDrop(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	if id == "" {
		errResponse(w, "ID is required", http.StatusBadRequest)
		return
	}

	if err := h.dropsService.DeleteDrop(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrDropNotFound) {
			errResponse(w, "Drop not found", http.StatusNotFound)
			return
		}
		errResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func errResponse(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": msg,
	})
}
