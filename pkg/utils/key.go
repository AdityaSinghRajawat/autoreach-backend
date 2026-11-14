package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

func BuildKey(parts ...string) string {
	segments := make([]string, 0, len(parts)+1)
	segments = append(segments, config.GetKeyPrefix())
	segments = append(segments, parts...)

	return strings.Join(segments, config.GetKeyDelimiter())
}

func Fingerprint(secret string) string {
	digest := sha256.Sum256([]byte(secret))

	return hex.EncodeToString(digest[:])
}
