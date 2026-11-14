package unit

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
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

	principal := sharedTypes.Principal{
		UserID:      "user-1",
		CandidateID: "candidate-1",
		Role:        modelHelpers.UserRoleCandidate,
	}

	token, expiresAt, err := utils.IssueAccessToken(principal, utils.GetCurrTimeStamp())
	require.NoError(t, err)

	parsed, err := utils.ParseAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, principal, parsed)
	assert.WithinDuration(t, time.Now().Add(config.GetAccessTokenTTL()), expiresAt, 5*time.Second)
}

func TestAccessTokenIsRejectedWhenExpiredTamperedOrForeign(t *testing.T) {
	setup.Config(t)

	principal := sharedTypes.Principal{UserID: "user-1", Role: modelHelpers.UserRoleCandidate}

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
		"sub":  "user-1",
		"role": "ADMIN",
		"iss":  config.GetTokenIssuer(),
		"exp":  time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte("a-different-secret-of-sufficient-length"))
	require.NoError(t, err)
	_, err = utils.ParseAccessToken(foreign)
	assert.ErrorIs(t, err, config.ErrInvalidToken)

	unknownRole, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  "user-1",
		"role": "SUPERUSER",
		"iss":  config.GetTokenIssuer(),
		"exp":  time.Now().Add(time.Hour).Unix(),
	}).SignedString(config.GetJWTSecret())
	require.NoError(t, err)
	_, err = utils.ParseAccessToken(unknownRole)
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

func TestPrincipalCanOnlyAccessItsOwnCandidateUnlessAdmin(t *testing.T) {
	candidate := sharedTypes.Principal{
		UserID:      "u",
		CandidateID: "c-1",
		Role:        modelHelpers.UserRoleCandidate,
	}
	orphan := sharedTypes.Principal{UserID: "u", Role: modelHelpers.UserRoleCandidate}
	admin := sharedTypes.Principal{UserID: "a", Role: modelHelpers.UserRoleAdmin}

	assert.True(t, candidate.CanAccessCandidate("c-1"))
	assert.False(t, candidate.CanAccessCandidate("c-2"))
	assert.False(t, orphan.CanAccessCandidate(""))
	assert.True(t, admin.CanAccessCandidate("c-2"))
}
