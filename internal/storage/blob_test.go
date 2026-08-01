package storage

import (
	"bytes"
	"io"
	"testing"
)

func TestFileStorage(t *testing.T) {
	temp := t.TempDir()

	fileStorage, err := NewFileStorage(temp)
	if err != nil {
		t.Fatalf("failed to create file storage: %v", err)
	}

	testFileName := "test_file.txt"
	testContent := "This is testing."
	contentReader := bytes.NewBufferString(testContent)

	t.Run("saving the blob", func(t *testing.T) {
		if err := fileStorage.Save(testFileName, contentReader); err != nil {
			t.Fatalf("failed to save file: %v", err)
		}
	})

	t.Run("get the blob by name", func(t *testing.T) {
		rc, err := fileStorage.Get(testFileName)
		if err != nil {
			t.Fatalf("failed to get file: %v", err)
		}
		defer func() {
			_ = rc.Close()
		}()

		readBytes, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("failed to read bytes from get: %v", err)
		}

		if string(readBytes) != testContent {
			t.Errorf("expected content %q, got %q from get", testContent, string(readBytes))
		}
	})

	t.Run("deleting blob", func(t *testing.T) {
		if err := fileStorage.Delete(testFileName); err != nil {
			t.Fatalf("failed to delete file: %v", err)
		}

		if _, err := fileStorage.Get(testFileName); err == nil {
			t.Errorf("expected no file after delete, but got %v", err)
		}
	})
}
