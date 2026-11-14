package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

const argonEncodingFormat = "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s"

func HashPassword(password string) (string, error) {
	salt := make([]byte, config.GetArgonSaltLength())
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}

	memory, iterations, parallelism := config.GetArgonMemoryKiB(), config.GetArgonIterations(), config.GetArgonParallelism()
	key := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		parallelism,
		config.GetArgonKeyLength(),
	)

	return fmt.Sprintf(
		argonEncodingFormat,
		argon2.Version, memory, iterations, parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

func VerifyPassword(password string, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, fmt.Errorf("verify password: %w", config.ErrUnsupportedPasswordHash)
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return false, fmt.Errorf("verify password: %w", config.ErrUnsupportedPasswordHash)
	}

	var (
		memory, iterations uint32
		parallelism        uint8
	)
	if _, err := fmt.Sscanf(
		parts[3],
		"m=%d,t=%d,p=%d",
		&memory,
		&iterations,
		&parallelism,
	); err != nil {
		return false, fmt.Errorf("verify password: parse parameters: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("verify password: decode salt: %w", err)
	}

	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("verify password: decode key: %w", err)
	}

	actual := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		parallelism,
		uint32(len(expected)),
	)

	return subtle.ConstantTimeCompare(actual, expected) == 1, nil
}
