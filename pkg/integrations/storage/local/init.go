package local

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

const (
	directoryMode = 0o750
	fileMode      = 0o600
)

type LocalStorageIntegration struct {
	rootDirectory string
}

func NewLocalStorageIntegration(rootDirectory string) (*LocalStorageIntegration, error) {
	absolute, err := filepath.Abs(rootDirectory)
	if err != nil {
		return nil, fmt.Errorf("new local storage: %w", err)
	}

	return &LocalStorageIntegration{rootDirectory: absolute}, nil
}

func (i *LocalStorageIntegration) resolve(key string) (string, error) {
	cleaned := filepath.Clean(filepath.FromSlash(key))
	if key == "" || strings.ContainsAny(key[:1], `/\`) || strings.Contains(key, ":") ||
		filepath.IsAbs(cleaned) || strings.HasPrefix(cleaned, "..") {
		return "", fmt.Errorf("resolve storage key: %w", config.ErrInvalidStorageKey)
	}

	resolved := filepath.Join(i.rootDirectory, cleaned)
	if !strings.HasPrefix(resolved, i.rootDirectory+string(filepath.Separator)) {
		return "", fmt.Errorf("resolve storage key: %w", config.ErrInvalidStorageKey)
	}

	return resolved, nil
}
