package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/rizalta/file-drop/internal/repo"
	"github.com/rizalta/file-drop/internal/service"
)

type DropsService interface {
	CleanupExpiredDrops(ctx context.Context) error
	CreateDrop(ctx context.Context, params service.CreateDropParams) (string, error)
	DeleteDrop(ctx context.Context, id string) error
	GetDrop(ctx context.Context, id string) (*repo.Drop, io.ReadCloser, error)
	ListActiveDrops(ctx context.Context) ([]repo.Drop, error)
}

type dropsHandler struct {
	dropsService DropsService
}

func NewHandler(s DropsService) *dropsHandler {
	return &dropsHandler{s}
}

type UploadDropRes struct {
	ID string `json:"id"`
}

func (h *dropsHandler) UploadDrop(w http.ResponseWriter, r *http.Request) {
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
		if errors.Is(err, service.ErrInvalidExpiry) || errors.Is(err, service.ErrInvalidReader) {
			errResponse(w, err.Error(), http.StatusBadRequest)
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

func errResponse(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": msg,
	})
}
