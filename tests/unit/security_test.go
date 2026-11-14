package unit

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
)

func TestPasswordHashVerifiesOnlyTheOriginalPassword(t *testing.T) {
	setup.Config(t)

	first, err := utils.HashPassword("correct-horse-battery")
	require.NoError(t, err)
	second, err := utils.HashPassword("correct-horse-battery")
	require.NoError(t, err)

	assert.NotEqual(t, first, second, "each hash must use a fresh salt")
	assert.NotContains(t, first, "correct-horse-battery")

	matches, err := utils.VerifyPassword("correct-horse-battery", first)
	require.NoError(t, err)
	assert.True(t, matches)

	matches, err = utils.VerifyPassword("correct-horse-batterx", first)
	require.NoError(t, err)
	assert.False(t, matches)

	_, err = utils.VerifyPassword("anything", "plain-text-not-a-hash")
	assert.Error(t, err)
}

func TestAccessTokenRoundTripsThePrincipal(t *testing.T) {
	setup.Config(t)

	principal := sharedTypes.NewUserPrincipal(
		"user-1",
		"candidate-1",
		[]string{config.RoleCandidate},
	)

	token, expiresAt, err := utils.IssueAccessToken(principal, utils.GetCurrTimeStamp())
	require.NoError(t, err)

	parsed, err := utils.ParseAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, principal, parsed)
	assert.WithinDuration(t, time.Now().Add(config.GetAccessTokenTTL()), expiresAt, 5*time.Second)
}

func TestAccessTokenIsRejectedWhenExpiredTamperedOrForeign(t *testing.T) {
	setup.Config(t)

	principal := sharedTypes.NewUserPrincipal("user-1", "", []string{config.RoleCandidate})

	expired, _, err := utils.IssueAccessToken(
		principal,
		utils.GetCurrTimeStamp().Add(-2*config.GetAccessTokenTTL()),
	)
	require.NoError(t, err)
	_, err = utils.ParseAccessToken(expired)
	assert.ErrorIs(t, err, config.ErrInvalidToken)

	valid, _, err := utils.IssueAccessToken(principal, utils.GetCurrTimeStamp())
	require.NoError(t, err)
	_, err = utils.ParseAccessToken(valid[:len(valid)-2] + "xx")
	assert.ErrorIs(t, err, config.ErrInvalidToken)

	foreign, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   "user-1",
		"roles": []string{config.RoleAdmin},
		"iss":   config.GetTokenIssuer(),
		"exp":   time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte("a-different-secret-of-sufficient-length"))
	require.NoError(t, err)
	_, err = utils.ParseAccessToken(foreign)
	assert.ErrorIs(t, err, config.ErrInvalidToken)

	withoutRoles, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   "user-1",
		"roles": []string{},
		"iss":   config.GetTokenIssuer(),
		"exp":   time.Now().Add(time.Hour).Unix(),
	}).SignedString(config.GetJWTSecret())
	require.NoError(t, err)
	_, err = utils.ParseAccessToken(withoutRoles)
	assert.ErrorIs(t, err, config.ErrInvalidToken)
}

func TestOpaqueTokensAreUniqueAndStoredOnlyAsAFingerprint(t *testing.T) {
	setup.Config(t)

	first, firstHash, err := utils.NewOpaqueToken()
	require.NoError(t, err)
	second, _, err := utils.NewOpaqueToken()
	require.NoError(t, err)

	assert.NotEqual(t, first, second)
	assert.Equal(t, utils.Fingerprint(first), firstHash)
	assert.NotEqual(t, first, firstHash)
	assert.Len(t, firstHash, 64)
}
