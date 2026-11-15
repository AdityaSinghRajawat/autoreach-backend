package unit

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func TestCredentialCipherRoundTripsAndBindsCiphertextToItsRow(t *testing.T) {
	oldKey := bytes.Repeat([]byte{1}, 32)
	newKey := bytes.Repeat([]byte{2}, 32)

	previous, err := utils.NewCredentialCipher(map[int][]byte{1: oldKey}, 1)
	require.NoError(t, err)

	sealed, version, err := previous.Encrypt([]byte("secret"), []byte("account-1"))
	require.NoError(t, err)
	assert.Equal(t, 1, version)
	assert.NotContains(t, string(sealed), "secret")

	rotated, err := utils.NewCredentialCipher(map[int][]byte{1: oldKey, 2: newKey}, 2)
	require.NoError(t, err)

	plaintext, err := rotated.Decrypt(sealed, version, []byte("account-1"))
	require.NoError(t, err, "rotation keeps old keys readable")
	assert.Equal(t, "secret", string(plaintext))

	_, err = rotated.Decrypt(sealed, version, []byte("account-2"))
	require.Error(t, err, "a ciphertext copied onto another row does not open")

	tampered := append([]byte(nil), sealed...)
	tampered[len(tampered)-1] ^= 1
	_, err = rotated.Decrypt(tampered, version, []byte("account-1"))
	require.Error(t, err)

	_, err = rotated.Decrypt(sealed, 9, []byte("account-1"))
	require.ErrorIs(t, err, config.ErrUnknownCredentialKey)

	_, newVersion, err := rotated.Encrypt([]byte("secret"), []byte("account-1"))
	require.NoError(t, err)
	assert.Equal(t, 2, newVersion, "new writes use the active key")
}

func TestCredentialCipherNeedsTheActiveKey(t *testing.T) {
	_, err := utils.NewCredentialCipher(map[int][]byte{1: bytes.Repeat([]byte{1}, 32)}, 2)
	require.ErrorIs(t, err, config.ErrUnknownCredentialKey)

	_, err = utils.NewCredentialCipher(nil, 1)
	require.ErrorIs(t, err, config.ErrCredentialEncryptionUnavailable)
}
