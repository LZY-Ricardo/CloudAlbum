package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) (*LocalStorage, error) {
	absPath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("resolve storage path: %w", err)
	}
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil, fmt.Errorf("create storage directory: %w", err)
	}
	return &LocalStorage{basePath: absPath}, nil
}

func (s *LocalStorage) resolvePath(key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("invalid storage key: empty")
	}
	if filepath.IsAbs(key) {
		return "", fmt.Errorf("invalid storage key %q: absolute paths are not allowed", key)
	}

	cleanKey := filepath.Clean(key)
	if cleanKey == "." {
		return "", fmt.Errorf("invalid storage key %q", key)
	}

	fullPath := filepath.Join(s.basePath, cleanKey)
	rel, err := filepath.Rel(s.basePath, fullPath)
	if err != nil {
		return "", fmt.Errorf("resolve storage key %q: %w", key, err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid storage key %q: path escapes storage root", key)
	}

	return fullPath, nil
}

func (s *LocalStorage) String() string {
	return s.basePath
}

func (s *LocalStorage) Save(_ context.Context, key string, data io.Reader) error {
	fullPath, err := s.resolvePath(key)
	if err != nil {
		return err
	}
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	f, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, data); err != nil {
		os.Remove(fullPath)
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

func (s *LocalStorage) Get(_ context.Context, key string) (io.ReadCloser, error) {
	fullPath, err := s.resolvePath(key)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found %q: %w", key, err)
		}
		return nil, fmt.Errorf("open file %q: %w", key, err)
	}
	return f, nil
}

func (s *LocalStorage) Exists(_ context.Context, key string) (bool, error) {
	fullPath, err := s.resolvePath(key)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *LocalStorage) Delete(_ context.Context, key string) error {
	fullPath, err := s.resolvePath(key)
	if err != nil {
		return err
	}

	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete file %q: %w", key, err)
	}
	return nil
}
