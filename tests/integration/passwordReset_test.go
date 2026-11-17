//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/authManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/emailTemplates"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
)

const (
	forgotPasswordPath = "/api/v1/auth/forgot-password"
	resetPasswordPath  = "/api/v1/auth/reset-password"
	newPassword        = "a-brand-new-password"
)

func forgotPassword(t *testing.T, router http.Handler, email string) helpers.Exchange {
	t.Helper()

	exchange := helpers.Do(t, router, http.MethodPost, forgotPasswordPath,
		authTypes.ForgotPasswordRequest{Email: email}, nil)

	utils.DrainBackgroundGoroutines()

	return exchange
}

func resetPassword(t *testing.T, router http.Handler, token string) helpers.Exchange {
	t.Helper()

	return helpers.Do(t, router, http.MethodPost, resetPasswordPath,
		authTypes.ResetPasswordRequest{Token: token, Password: newPassword}, nil)
}

func TestForgotPasswordAnswersAlikeForUnknownAddresses(t *testing.T) {
	router, mailer, _ := mailingRouter(t, false)
	register(t, router, "alex@example.com")

	known := forgotPassword(t, router, "Alex@Example.com")
	require.Equal(t, http.StatusAccepted, known.Recorder.Code, string(known.Body))
	helpers.AssertMatchesSpec(t, known)

	unknown := forgotPassword(t, router, "nobody@example.com")
	assert.Equal(t, known.Recorder.Code, unknown.Recorder.Code)
	assert.JSONEq(t, string(known.Body), string(unknown.Body))

	require.Len(t, mailer.SentTo("alex@example.com"), 1)
	assert.Equal(t, emailTemplates.PasswordResetEmailSubject, mailer.Sent[0].Subject)
	assert.Empty(t, mailer.SentTo("nobody@example.com"), "unknown addresses get no email")
}

func TestResettingThePasswordWorksOnceAndEndsEverySession(t *testing.T) {
	router, mailer, _ := mailingRouter(t, false)
	session := register(t, router, "alex@example.com")

	forgotPassword(t, router, "alex@example.com")
	token := linkToken(t, mailer.SentTo("alex@example.com")[0].Body)

	reset := resetPassword(t, router, token)
	require.Equal(t, http.StatusNoContent, reset.Recorder.Code, string(reset.Body))
	helpers.AssertMatchesSpec(t, reset)

	again := resetPassword(t, router, token)
	assert.Equal(
		t,
		http.StatusBadRequest,
		again.Recorder.Code,
		"the link is spent once the password changed",
	)

	oldSession := helpers.Do(t, router, http.MethodPost, refreshPath,
		authTypes.RefreshRequest{RefreshToken: session.RefreshToken}, nil)
	assert.Equal(
		t,
		http.StatusUnauthorized,
		oldSession.Recorder.Code,
		"existing sessions are ended",
	)

	oldPassword := helpers.Do(t, router, http.MethodPost, loginPath,
		authTypes.LoginRequest{Email: "alex@example.com", Password: testPassword}, nil)
	assert.Equal(t, http.StatusUnauthorized, oldPassword.Recorder.Code)

	login := helpers.Do(t, router, http.MethodPost, loginPath,
		authTypes.LoginRequest{Email: "alex@example.com", Password: newPassword}, nil)
	assert.Equal(t, http.StatusOK, login.Recorder.Code, string(login.Body))
}

func TestAResetLinkAlsoVerifiesTheEmail(t *testing.T) {
	router, mailer, _ := verifyingRouter(t)

	require.Equal(t, http.StatusAccepted, signUp(t, router, "alex@example.com").Recorder.Code)
	forgotPassword(t, router, "alex@example.com")

	sent := mailer.SentTo("alex@example.com")
	require.Equal(t, emailTemplates.PasswordResetEmailSubject, sent[len(sent)-1].Subject)
	require.Equal(t, http.StatusNoContent,
		resetPassword(t, router, linkToken(t, sent[len(sent)-1].Body)).Recorder.Code)

	login := helpers.Do(t, router, http.MethodPost, loginPath,
		authTypes.LoginRequest{Email: "alex@example.com", Password: newPassword}, nil)
	assert.Equal(t, http.StatusOK, login.Recorder.Code, "opening the reset link proved the inbox")
}

func TestForgedOrMismatchedResetLinksAreRefused(t *testing.T) {
	router, mailer, _ := mailingRouter(t, false)
	register(t, router, "alex@example.com")

	forgotPassword(t, router, "alex@example.com")

	token := linkToken(t, mailer.SentTo("alex@example.com")[0].Body)

	for _, forged := range []string{"not-a-token", token + "0", token[:len(token)-1]} {
		assert.Equal(
			t,
			http.StatusBadRequest,
			resetPassword(t, router, forged).Recorder.Code,
			forged,
		)
	}

	verifyLink := helpers.Do(t, router, http.MethodPost, verifyEmailPath,
		authTypes.VerifyEmailRequest{Token: token}, nil)
	assert.Equal(t, http.StatusBadRequest, verifyLink.Recorder.Code,
		"a reset link cannot be used as a verification link")
}
