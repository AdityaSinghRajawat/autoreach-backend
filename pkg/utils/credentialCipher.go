package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

type CredentialCipher struct {
	aeads         map[int]cipher.AEAD
	activeVersion int
}

func NewCredentialCipher(keys map[int][]byte, activeVersion int) (*CredentialCipher, error) {
	if len(keys) == 0 {
		return nil, fmt.Errorf(
			"new credential cipher: %w",
			config.ErrCredentialEncryptionUnavailable,
		)
	}

	aeads := make(map[int]cipher.AEAD, len(keys))

	for version, key := range keys {
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, fmt.Errorf("new credential cipher: key %d: %w", version, err)
		}

		aead, err := cipher.NewGCM(block)
		if err != nil {
			return nil, fmt.Errorf("new credential cipher: key %d: %w", version, err)
		}

		aeads[version] = aead
	}

	if _, found := aeads[activeVersion]; !found {
		return nil, fmt.Errorf("new credential cipher: %w", config.ErrUnknownCredentialKey)
	}

	return &CredentialCipher{aeads: aeads, activeVersion: activeVersion}, nil
}

// associatedData binds a ciphertext to its row, so it cannot be copied onto another record.
func (c *CredentialCipher) Encrypt(plaintext, associatedData []byte) ([]byte, int, error) {
	aead := c.aeads[c.activeVersion]

	nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+len(plaintext)+aead.Overhead())
	if _, err := rand.Read(nonce); err != nil {
		return nil, 0, fmt.Errorf("encrypt credentials: %w", err)
	}

	return aead.Seal(nonce, nonce, plaintext, associatedData), c.activeVersion, nil
}

func (c *CredentialCipher) Decrypt(
	sealed []byte,
	version int,
	associatedData []byte,
) ([]byte, error) {
	aead, found := c.aeads[version]
	if !found {
		return nil, fmt.Errorf("decrypt credentials: %w", config.ErrUnknownCredentialKey)
	}

	if len(sealed) < aead.NonceSize() {
		return nil, fmt.Errorf("decrypt credentials: %w", config.ErrInvalidToken)
	}

	nonce, ciphertext := sealed[:aead.NonceSize()], sealed[aead.NonceSize():]

	plaintext, err := aead.Open(nil, nonce, ciphertext, associatedData)
	if err != nil {
		return nil, fmt.Errorf("decrypt credentials: %w", err)
	}

	return plaintext, nil
}
