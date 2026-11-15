package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

func Fingerprint(secret string) string {
	digest := sha256.Sum256([]byte(secret))

	return hex.EncodeToString(digest[:])
}

func DeriveKey(secret []byte, label string) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(label))

	return mac.Sum(nil)
}

func Sign(key []byte, parts ...string) string {
	mac := hmac.New(sha256.New, key)
	for _, part := range parts {
		mac.Write([]byte(part))
		mac.Write([]byte{0})
	}

	return hex.EncodeToString(mac.Sum(nil))
}

func VerifySignature(key []byte, signature string, parts ...string) bool {
	return hmac.Equal([]byte(signature), []byte(Sign(key, parts...)))
}

func NewRandomHex(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	return hex.EncodeToString(buffer), nil
}

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
		config.GetArgonEncodingFormat(),
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
