package local

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

func (i *LocalStorageIntegration) Save(_ context.Context, key string, content []byte) error {
	path, err := i.resolve(key)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), config.GetStorageDirectoryMode()); err != nil {
		return fmt.Errorf("save file: %w", err)
	}

	if err := os.WriteFile(path, content, config.GetStorageFileMode()); err != nil {
		return fmt.Errorf("save file: %w", err)
	}

	return nil
}

func (i *LocalStorageIntegration) Read(_ context.Context, key string) ([]byte, error) {
	path, err := i.resolve(key)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("read file: %w", config.ErrStoredFileNotFound)
	}

	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	return content, nil
}

func (i *LocalStorageIntegration) Delete(_ context.Context, key string) error {
	path, err := i.resolve(key)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("delete file: %w", err)
	}

	return nil
}
