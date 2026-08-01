package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rizalta/file-drop/internal/repo"
)

const (
	codeLen    = 5
	maxRetries = 3
	CHARSET    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var ErrMaxRetries = errors.New("max retries reached creating drop")

type Storage interface {
	Save(storedName string, r io.Reader) error
	Get(storedName string) (io.ReadCloser, error)
	Delete(storedName string) error
}

type service struct {
	queries *repo.Queries
	storage Storage
}

func NewService(q *repo.Queries, s Storage) *service {
	return &service{
		queries: q,
		storage: s,
	}
}

type CreateDropParams struct {
	Filename          string
	FileSize          int
	MimeType          string
	IsText            bool
	TextContent       string
	BurnAfterDownload bool
	ExpiresIn         string
	Reader            io.Reader
}

func (s *service) CreateDrop(ctx context.Context, params CreateDropParams) (string, error) {
	expiresIn, err := parseExpiry(params.ExpiresIn)
	if err != nil {
		return "", fmt.Errorf("failed to parse expires in: %v", err)
	}

	storedName := ""
	if !params.IsText {
		if params.Reader == nil {
			return "", fmt.Errorf("invalid file reader")
		}

		storedName = uuid.NewString()
		if err := s.storage.Save(storedName, params.Reader); err != nil {
			return "", fmt.Errorf("failed to save blob: %v", err)
		}
	}

	for range maxRetries {
		code, err := generateCode()
		if err != nil {
			continue
		}

		drop, err := s.queries.CreateDrop(ctx, repo.CreateDropParams{
			ID:                code,
			Filename:          params.Filename,
			FileSize:          int64(params.FileSize),
			StoredName:        storedName,
			MimeType:          params.MimeType,
			IsText:            params.IsText,
			TextContent:       pgtype.Text{String: params.TextContent, Valid: params.IsText},
			BurnAfterDownload: params.BurnAfterDownload,
			ExpiresAt:         time.Now().Add(expiresIn),
		})

		if err == nil {
			return drop.ID, nil
		}
	}

	if !params.IsText && storedName != "" {
		if delErr := s.storage.Delete(storedName); delErr != nil {
			log.Printf("failed blob cleanup of name: %s, %v", storedName, delErr)
		}
	}

	return "", ErrMaxRetries
}

func generateCode() (string, error) {
	b := make([]byte, codeLen)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	for i := range codeLen {
		b[i] = CHARSET[b[i]%byte(len(CHARSET))]
	}

	return string(b), nil
}

func parseExpiry(expiresIn string) (time.Duration, error) {
	var ret time.Duration
	switch expiresIn {
	case "7d", "":
		ret = 7 * 24 * time.Hour
	case "3d":
		ret = 3 * 24 * time.Hour
	case "1d":
		ret = 24 * time.Hour
	default:
		var err error
		ret, err = time.ParseDuration(expiresIn)
		if err != nil {
			return 0, err
		}
	}

	return ret, nil
}
