package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rizalta/file-drop/internal/repo"
)

const (
	idLen      = 5
	maxRetries = 3
	CHARSET    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var (
	ErrDropNotFound  = errors.New("drop not found")
	ErrInvalidExpiry = errors.New("invalid expiration duration")
	ErrInvalidReader = errors.New("invalid file reader")
	ErrMaxRetries    = errors.New("max retries reached creating drop")
)

type Storage interface {
	Save(storedName string, r io.Reader) error
	Get(storedName string) (io.ReadCloser, error)
	Delete(storedName string) error
}

type Querier interface {
	CreateDrop(ctx context.Context, arg repo.CreateDropParams) (repo.Drop, error)
	DeleteDrop(ctx context.Context, id string) (string, error)
	DeleteExpiredDrops(ctx context.Context) ([]repo.DeleteExpiredDropsRow, error)
	GetDropByID(ctx context.Context, id string) (repo.Drop, error)
	IncrementDownloadCount(ctx context.Context, id string) error
	ListActiveDrops(ctx context.Context) ([]repo.Drop, error)
}

type service struct {
	queries Querier
	storage Storage
}

func NewService(q Querier, s Storage) *service {
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
		return "", fmt.Errorf("%w: %v", ErrInvalidExpiry, err)
	}

	storedName := ""
	if !params.IsText {
		if params.Reader == nil {
			return "", ErrInvalidReader
		}

		storedName = uuid.NewString()
		if err := s.storage.Save(storedName, params.Reader); err != nil {
			return "", fmt.Errorf("failed to save blob: %w", err)
		}
	}

	for range maxRetries {
		id, err := generateID()
		if err != nil {
			continue
		}

		drop, err := s.queries.CreateDrop(ctx, repo.CreateDropParams{
			ID:                id,
			Filename:          params.Filename,
			FileSize:          int64(params.FileSize),
			StoredName:        storedName,
			MimeType:          params.MimeType,
			IsText:            params.IsText,
			TextContent:       pgtype.Text{String: params.TextContent, Valid: params.IsText},
			BurnAfterDownload: params.BurnAfterDownload,
			ExpiresAt:         time.Now().UTC().Add(expiresIn),
		})

		if err == nil {
			return drop.ID, nil
		}
	}

	if !params.IsText && storedName != "" {
		if delErr := s.storage.Delete(storedName); delErr != nil && !errors.Is(delErr, os.ErrNotExist) {
			log.Printf("failed blob cleanup of name: %s, %v", storedName, delErr)
		}
	}

	return "", ErrMaxRetries
}

func (s *service) GetDrop(ctx context.Context, id string, isDownload bool) (*repo.Drop, io.ReadCloser, error) {
	drop, err := s.queries.GetDropByID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrDropNotFound, err)
	}

	var rc io.ReadCloser
	if !drop.IsText && isDownload {
		rc, err = s.storage.Get(drop.StoredName)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get blob: %w", err)
		}
	}

	if isDownload {
		if drop.BurnAfterDownload {
			storedName, err := s.queries.DeleteDrop(ctx, id)
			if err == nil && !drop.IsText && storedName != "" {
				if err := s.storage.Delete(storedName); err != nil && !errors.Is(err, os.ErrNotExist) {
					log.Printf("failed to delete blob %s: %v", storedName, err)
				}
			}
		} else {
			if err := s.queries.IncrementDownloadCount(ctx, id); err != nil {
				log.Printf("failed to increment download count of %s: %v", id, err)
			}
		}
	}

	return &drop, rc, nil
}

func (s *service) ListActiveDrops(ctx context.Context) ([]repo.Drop, error) {
	return s.queries.ListActiveDrops(ctx)
}

func (s *service) DeleteDrop(ctx context.Context, id string) error {
	storedName, err := s.queries.DeleteDrop(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDropNotFound, err)
	}

	if storedName != "" {
		if err := s.storage.Delete(storedName); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to delete stored blob: %w", err)
		}
	}

	return nil
}

func (s *service) CleanupExpiredDrops(ctx context.Context) error {
	drops, err := s.queries.DeleteExpiredDrops(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete expired drops from DB: %w", err)
	}

	for _, drop := range drops {
		if drop.StoredName != "" {
			if err := s.storage.Delete(drop.StoredName); err != nil && !errors.Is(err, os.ErrNotExist) {
				log.Printf("failed to delete expired blob %s: %v", drop.StoredName, err)
			}
		}
	}

	return nil
}

func generateID() (string, error) {
	b := make([]byte, idLen)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	for i := range idLen {
		b[i] = CHARSET[b[i]%byte(len(CHARSET))]
	}

	return string(b), nil
}

func parseExpiry(expiresIn string) (time.Duration, error) {
	expiresIn = strings.TrimSpace(strings.ToLower(expiresIn))
	if strings.HasSuffix(expiresIn, "d") {
		days, err := strconv.Atoi(expiresIn[:len(expiresIn)-1])
		if err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}

	return time.ParseDuration(expiresIn)
}
