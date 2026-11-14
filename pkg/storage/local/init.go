package local

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type LocalStorage struct {
	rootDirectory string
}

func NewLocalStorage(rootDirectory string) (*LocalStorage, error) {
	absolute, err := filepath.Abs(rootDirectory)
	if err != nil {
		return nil, fmt.Errorf("new local storage: %w", err)
	}

	return &LocalStorage{rootDirectory: absolute}, nil
}

func (s *LocalStorage) resolve(key string) (string, error) {
	if err := utils.ValidateStorageKey(key); err != nil {
		return "", err
	}

	resolved := filepath.Join(s.rootDirectory, filepath.FromSlash(key))
	if !strings.HasPrefix(resolved, s.rootDirectory+string(filepath.Separator)) {
		return "", fmt.Errorf("resolve storage key: %w", config.ErrInvalidStorageKey)
	}

	return resolved, nil
}
