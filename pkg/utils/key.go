package utils

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

// Empty parts are skipped, so an optional segment never leaves a doubled delimiter.
func BuildKey(delimiter string, parts ...string) string {
	kept := make([]string, 0, len(parts))

	for _, part := range parts {
		if part != "" {
			kept = append(kept, part)
		}
	}

	return strings.Join(kept, delimiter)
}

func ValidateStorageKey(key string) error {
	if key == "" ||
		strings.HasPrefix(key, config.GetStorageKeySeparator()) ||
		strings.ContainsAny(key, `\:`) {
		return fmt.Errorf("validate storage key: %w", config.ErrInvalidStorageKey)
	}

	for _, segment := range strings.Split(key, config.GetStorageKeySeparator()) {
		if segment == "" || segment == "." || segment == config.GetStorageKeyParentToken() {
			return fmt.Errorf("validate storage key: %w", config.ErrInvalidStorageKey)
		}
	}

	return nil
}

func GenerateUUID() string {
	return uuid.NewString()
}
