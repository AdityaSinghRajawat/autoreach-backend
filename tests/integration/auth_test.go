//go:build integration

package integration

import (
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	authTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/authManagement"
	userTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/userManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/routes"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
)

const (
	registerPath = "/api/v1/auth/register"
	loginPath    = "/api/v1/auth/login"
	refreshPath  = "/api/v1/auth/refresh"
	logoutPath   = "/api/v1/auth/logout"
	mePath       = "/api/v1/users/me"
	testPassword = "correct-horse-battery"
)

func newRouter(t *testing.T) http.Handler {
	t.Helper()

	return routes.NewRoutes(setup.Dependencies(t))
}

func bearer(accessToken string) map[string]string {
	return map[string]string{
		config.GetAuthorizationHeader(): config.GetBearerPrefix() + accessToken,
	}
}

func register(t *testing.T, router http.Handler, email string) authTypes.SessionResponse {
	t.Helper()
	exchange := helpers.Do(t, router, http.MethodPost, registerPath, authTypes.RegisterRequest{
		Email:    email,
		Password: testPassword,
		FullName: "Alex Candidate",
	}, nil)
	require.Equal(t, http.StatusCreated, exchange.Recorder.Code, string(exchange.Body))
	helpers.AssertMatchesSpec(t, exchange)

	return helpers.DecodeData[authTypes.SessionResponse](t, exchange)
}

func TestRegisterCreatesAUserACandidateAndASession(t *testing.T) {
	router := newRouter(t)

	session := register(t, router, "Alex@Example.com")

	assert.Equal(t, "alex@example.com", session.User.Email)
	assert.Equal(t, []string{config.RoleCandidate}, session.User.Roles)
	require.NotNil(t, session.User.CandidateID)
	assert.Equal(t, "Bearer", session.TokenType)
	assert.Equal(t, int(config.GetAccessTokenTTL().Seconds()), session.ExpiresIn)
	assert.NotEmpty(t, session.AccessToken)
	assert.NotEmpty(t, session.RefreshToken)

	me := helpers.Do(t, router, http.MethodGet, mePath, nil, bearer(session.AccessToken))
	require.Equal(t, http.StatusOK, me.Recorder.Code)
	helpers.AssertMatchesSpec(t, me)
	assert.Equal(t, session.User, helpers.DecodeData[userTypes.UserResponse](t, me))
}

func TestRegisterRejectsADuplicateEmailRegardlessOfCase(t *testing.T) {
	router := newRouter(t)
	register(t, router, "alex@example.com")

	exchange := helpers.Do(t, router, http.MethodPost, registerPath, authTypes.RegisterRequest{
		Email:    "ALEX@example.com",
		Password: testPassword,
		FullName: "Someone Else",
	}, nil)

	require.Equal(t, http.StatusConflict, exchange.Recorder.Code)
	helpers.AssertMatchesSpec(t, exchange)
	assert.Equal(
		t,
		string(config.GetErrCodeConflict()),
		helpers.DecodeEnvelope(t, exchange).Error.Code,
	)
}

func TestRegisterValidatesItsPayload(t *testing.T) {
	router := newRouter(t)

	exchange := helpers.Do(t, router, http.MethodPost, registerPath, map[string]string{
		"email":    "not-an-email",
		"password": "short",
	}, nil)

	require.Equal(t, http.StatusUnprocessableEntity, exchange.Recorder.Code)
	helpers.AssertMatchesSpec(t, exchange)
	assert.ElementsMatch(t, []config.ErrorDetail{
		{Field: "email", Issue: "must be a valid email address"},
		{Field: "password", Issue: "must be at least 10"},
		{Field: "fullName", Issue: "is required"},
	}, helpers.DecodeEnvelope(t, exchange).Error.Details)
}

func TestLoginSucceedsWithCorrectCredentials(t *testing.T) {
	router := newRouter(t)
	registered := register(t, router, "alex@example.com")

	exchange := helpers.Do(t, router, http.MethodPost, loginPath, authTypes.LoginRequest{
		Email:    "Alex@Example.com",
		Password: testPassword,
	}, nil)

	require.Equal(t, http.StatusOK, exchange.Recorder.Code, string(exchange.Body))
	helpers.AssertMatchesSpec(t, exchange)
	session := helpers.DecodeData[authTypes.SessionResponse](t, exchange)
	assert.Equal(t, registered.User, session.User)
	assert.NotEqual(t, registered.RefreshToken, session.RefreshToken)
}

func TestLoginFailuresAreIndistinguishable(t *testing.T) {
	router := newRouter(t)
	register(t, router, "alex@example.com")

	wrongPassword := helpers.Do(
		t,
		router,
		http.MethodPost,
		loginPath,
		authTypes.LoginRequest{Email: "alex@example.com", Password: "wrong-password-here"},
		nil,
	)
	unknownEmail := helpers.Do(
		t,
		router,
		http.MethodPost,
		loginPath,
		authTypes.LoginRequest{Email: "nobody@example.com", Password: testPassword},
		nil,
	)

	require.Equal(t, http.StatusUnauthorized, wrongPassword.Recorder.Code)
	require.Equal(t, http.StatusUnauthorized, unknownEmail.Recorder.Code)
	helpers.AssertMatchesSpec(t, wrongPassword)
	assert.JSONEq(t, string(wrongPassword.Body), string(unknownEmail.Body))
}

func TestRefreshRotatesTheTokenAndReuseRevokesTheWholeSession(t *testing.T) {
	router := newRouter(t)
	original := register(t, router, "alex@example.com")

	first := helpers.Do(
		t,
		router,
		http.MethodPost,
		refreshPath,
		authTypes.RefreshRequest{RefreshToken: original.RefreshToken},
		nil,
	)
	require.Equal(t, http.StatusOK, first.Recorder.Code, string(first.Body))
	helpers.AssertMatchesSpec(t, first)
	rotated := helpers.DecodeData[authTypes.SessionResponse](t, first)
	assert.NotEqual(t, original.RefreshToken, rotated.RefreshToken)
	assert.Equal(t, original.User, rotated.User)

	reused := helpers.Do(
		t,
		router,
		http.MethodPost,
		refreshPath,
		authTypes.RefreshRequest{RefreshToken: original.RefreshToken},
		nil,
	)
	require.Equal(t, http.StatusUnauthorized, reused.Recorder.Code)
	helpers.AssertMatchesSpec(t, reused)

	afterReuse := helpers.Do(
		t,
		router,
		http.MethodPost,
		refreshPath,
		authTypes.RefreshRequest{RefreshToken: rotated.RefreshToken},
		nil,
	)
	assert.Equal(
		t,
		http.StatusUnauthorized,
		afterReuse.Recorder.Code,
		"the rotated token must die when its predecessor is replayed",
	)
}

func TestConcurrentRefreshesWithOneTokenEndTheWholeSession(t *testing.T) {
	router := newRouter(t)
	original := register(t, router, "alex@example.com")
	issued := make(chan string, 8)
	waiting := sync.WaitGroup{}

	for range cap(issued) {
		waiting.Go(func() {
			refreshed := helpers.Do(t, router, http.MethodPost, refreshPath,
				authTypes.RefreshRequest{RefreshToken: original.RefreshToken}, nil)
			if refreshed.Recorder.Code == http.StatusOK {
				issued <- helpers.DecodeData[authTypes.SessionResponse](t, refreshed).RefreshToken
			}
		})
	}

	waiting.Wait()
	close(issued)

	for token := range issued {
		later := helpers.Do(t, router, http.MethodPost, refreshPath,
			authTypes.RefreshRequest{RefreshToken: token}, nil)
		assert.Equal(t, http.StatusUnauthorized, later.Recorder.Code,
			"a token presented twice at once revokes every session issued from it")
	}
}

func TestRefreshRejectsAnUnknownToken(t *testing.T) {
	router := newRouter(t)

	exchange := helpers.Do(
		t,
		router,
		http.MethodPost,
		refreshPath,
		authTypes.RefreshRequest{RefreshToken: "never-issued"},
		nil,
	)

	assert.Equal(t, http.StatusUnauthorized, exchange.Recorder.Code)
}

func TestLogoutRevokesTheRefreshTokenAndIsIdempotent(t *testing.T) {
	router := newRouter(t)
	session := register(t, router, "alex@example.com")

	logout := helpers.Do(
		t,
		router,
		http.MethodPost,
		logoutPath,
		authTypes.LogoutRequest{RefreshToken: session.RefreshToken},
		nil,
	)
	require.Equal(t, http.StatusNoContent, logout.Recorder.Code)
	helpers.AssertMatchesSpec(t, logout)
	assert.Empty(t, logout.Body)

	again := helpers.Do(
		t,
		router,
		http.MethodPost,
		logoutPath,
		authTypes.LogoutRequest{RefreshToken: session.RefreshToken},
		nil,
	)
	assert.Equal(t, http.StatusNoContent, again.Recorder.Code)

	refresh := helpers.Do(
		t,
		router,
		http.MethodPost,
		refreshPath,
		authTypes.RefreshRequest{RefreshToken: session.RefreshToken},
		nil,
	)
	assert.Equal(t, http.StatusUnauthorized, refresh.Recorder.Code)
}

func TestMeRequiresAValidAccessToken(t *testing.T) {
	router := newRouter(t)

	missing := helpers.Do(t, router, http.MethodGet, mePath, nil, nil)
	garbage := helpers.Do(t, router, http.MethodGet, mePath, nil, bearer("not.a.jwt"))
	unsigned := helpers.Do(
		t,
		router,
		http.MethodGet,
		mePath,
		nil,
		bearer(
			"eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ4Iiwicm9sZSI6IkFETUlOIiwiaXNzIjoiYXV0b3JlYWNoIiwiZXhwIjo5OTk5OTk5OTk5fQ.",
		),
	)

	assert.Equal(t, http.StatusUnauthorized, missing.Recorder.Code)
	assert.Equal(t, http.StatusUnauthorized, garbage.Recorder.Code)
	assert.Equal(t, http.StatusUnauthorized, unsigned.Recorder.Code)
	helpers.AssertMatchesSpec(t, missing)
}
