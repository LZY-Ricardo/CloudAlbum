package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalStorageRejectsEscapingKeys(t *testing.T) {
	baseDir := t.TempDir()
	store, err := NewLocalStorage(baseDir)
	if err != nil {
		t.Fatalf("NewLocalStorage() error = %v", err)
	}

	traversalKey := filepath.Join("..", "escape.txt")
	outsidePath := filepath.Join(filepath.Dir(baseDir), "escape.txt")

	if err := store.Save(context.Background(), traversalKey, strings.NewReader("escape")); err == nil {
		t.Fatalf("Save() with traversal key succeeded")
	}
	if _, err := os.Stat(outsidePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("outside file stat error = %v, want not exist", err)
	}

	if _, err := store.Get(context.Background(), traversalKey); err == nil {
		t.Fatalf("Get() with traversal key succeeded")
	}
	if _, err := store.Exists(context.Background(), traversalKey); err == nil {
		t.Fatalf("Exists() with traversal key succeeded")
	}
	if err := store.Delete(context.Background(), traversalKey); err == nil {
		t.Fatalf("Delete() with traversal key succeeded")
	}

	if err := store.Save(context.Background(), string(filepath.Separator)+"tmp"+string(filepath.Separator)+"escape.txt", strings.NewReader("escape")); err == nil {
		t.Fatalf("Save() with absolute key succeeded")
	}
}

func TestLocalStorageGetPreservesNotExist(t *testing.T) {
	store, err := NewLocalStorage(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocalStorage() error = %v", err)
	}

	reader, err := store.Get(context.Background(), "missing/file.txt")
	if err == nil {
		if reader != nil {
			reader.Close()
		}
		t.Fatalf("Get() error = nil, want os.ErrNotExist")
	}
	if reader != nil {
		reader.Close()
		t.Fatalf("Get() reader = %v, want nil", reader)
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Get() error = %v, want errors.Is(err, os.ErrNotExist)", err)
	}
}

func TestLocalStorageSaveAndGetValidKey(t *testing.T) {
	store, err := NewLocalStorage(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocalStorage() error = %v", err)
	}

	const key = "nested/path/file.txt"
	const content = "hello cloudalbum"

	if err := store.Save(context.Background(), key, strings.NewReader(content)); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	reader, err := store.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if got := string(data); got != content {
		t.Fatalf("Get() content = %q, want %q", got, content)
	}
}
