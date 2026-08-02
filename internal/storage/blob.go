package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type fileStorage struct {
	basePath string
}

func NewFileStorage(basePath string) (*fileStorage, error) {
	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &fileStorage{basePath}, nil
}

func (f *fileStorage) Save(storedName string, r io.Reader) error {
	filePath := filepath.Join(f.basePath, storedName)
	dst, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		_ = dst.Close()
	}()

	if _, err := io.Copy(dst, r); err != nil {
		_ = os.Remove(filePath)
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (f *fileStorage) Get(storedName string) (io.ReadCloser, error) {
	filePath := filepath.Join(f.basePath, storedName)
	src, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return src, nil
}

func (f *fileStorage) Delete(storedName string) error {
	filePath := filepath.Join(f.basePath, storedName)
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}
