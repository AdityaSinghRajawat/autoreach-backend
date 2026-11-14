package utils

import (
	"fmt"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

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
