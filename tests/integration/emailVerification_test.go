//go:build integration

package integration

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	authTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/authManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/dependencies"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/emailTemplates"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/routes"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/mocks"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
)

const verifyEmailPath = "/api/v1/auth/verify-email"

var verificationTokenPattern = regexp.MustCompile(`token=([^\s]+)`)

// verifyingRouter turns email verification on, with a mailer that records what it sends.
func verifyingRouter(t *testing.T) (http.Handler, *mocks.SystemMailer, *dependencies.Dependencies) {
	t.Helper()

	deps := setup.Dependencies(t)
	original := deps.SystemMailer
	mailer := &mocks.SystemMailer{}

	deps.SystemMailer = mailer
	deps.RequireEmailVerification = true
	require.NoError(t, deps.Wire())

	t.Cleanup(func() {
		deps.SystemMailer = original
		deps.RequireEmailVerification = false
		require.NoError(t, deps.Wire())
	})

	return routes.NewRoutes(deps), mailer, deps
}

func signUp(t *testing.T, router http.Handler, email string) helpers.Exchange {
	t.Helper()

	exchange := helpers.Do(t, router, http.MethodPost, registerPath, authTypes.RegisterRequest{
		Email:    email,
		Password: testPassword,
		FullName: "Alex Candidate",
	}, nil)

	utils.DrainBackgroundGoroutines()

	return exchange
}

func linkToken(t *testing.T, body string) string {
	t.Helper()

	match := verificationTokenPattern.FindStringSubmatch(body)
	require.Len(t, match, 2, "the email carries a verification link")

	token, err := url.QueryUnescape(match[1])
	require.NoError(t, err)

	return token
}

func TestSignupAnswersAlikeForNewAndRegisteredAddresses(t *testing.T) {
	router, mailer, _ := verifyingRouter(t)

	first := signUp(t, router, "alex@example.com")
	require.Equal(t, http.StatusAccepted, first.Recorder.Code, string(first.Body))
	helpers.AssertMatchesSpec(t, first)

	again := signUp(t, router, "Alex@Example.com")
	assert.Equal(t, first.Recorder.Code, again.Recorder.Code)
	assert.JSONEq(t, string(first.Body), string(again.Body),
		"a registered address cannot be told apart from a new one")

	sent := mailer.SentTo("alex@example.com")
	require.Len(t, sent, 2, "an unverified address gets a fresh link")
	assert.Equal(t, emailTemplates.VerificationEmailSubject, sent[1].Subject)

	unverified := helpers.Do(t, router, http.MethodPost, loginPath,
		authTypes.LoginRequest{Email: "alex@example.com", Password: testPassword}, nil)
	assert.Equal(t, http.StatusForbidden, unverified.Recorder.Code)

	verified := helpers.Do(t, router, http.MethodPost, verifyEmailPath,
		authTypes.VerifyEmailRequest{Token: linkToken(t, sent[0].Body)}, nil)
	require.Equal(t, http.StatusNoContent, verified.Recorder.Code, string(verified.Body))
	helpers.AssertMatchesSpec(t, verified)

	login := helpers.Do(t, router, http.MethodPost, loginPath,
		authTypes.LoginRequest{Email: "alex@example.com", Password: testPassword}, nil)
	require.Equal(t, http.StatusOK, login.Recorder.Code, string(login.Body))

	existing := signUp(t, router, "alex@example.com")
	assert.Equal(t, http.StatusAccepted, existing.Recorder.Code)

	sent = mailer.SentTo("alex@example.com")
	assert.Equal(t, emailTemplates.AccountExistsEmailSubject, sent[len(sent)-1].Subject,
		"a verified owner is told someone tried to sign up")
}

func TestSigningUpAgainNeverChangesTheExistingAccount(t *testing.T) {
	router, _, deps := verifyingRouter(t)

	require.Equal(t, http.StatusAccepted, signUp(t, router, "alex@example.com").Recorder.Code)

	takeover := helpers.Do(t, router, http.MethodPost, registerPath, authTypes.RegisterRequest{
		Email:    "alex@example.com",
		Password: "a-different-password",
		FullName: "Mallory",
	}, nil)
	assert.Equal(t, http.StatusAccepted, takeover.Recorder.Code)

	user, err := deps.Repositories.User.FindByEmail(
		context.Background(),
		"alex@example.com",
	)
	require.NoError(t, err)
	require.NotNil(t, user)

	matches, err := utils.VerifyPassword(testPassword, user.PasswordHash)
	require.NoError(t, err)
	assert.True(t, matches, "the second signup's password was not stored")
}

func TestBrokenOrForgedVerificationLinksAreRefused(t *testing.T) {
	router, mailer, _ := verifyingRouter(t)

	require.Equal(t, http.StatusAccepted, signUp(t, router, "alex@example.com").Recorder.Code)

	token := linkToken(t, mailer.SentTo("alex@example.com")[0].Body)

	for _, forged := range []string{"not-a-token", token + "0", token[:len(token)-1]} {
		refused := helpers.Do(t, router, http.MethodPost, verifyEmailPath,
			authTypes.VerifyEmailRequest{Token: forged}, nil)
		assert.Equal(t, http.StatusBadRequest, refused.Recorder.Code, forged)
	}
}

func TestTheSystemMailerDeliversThroughTheConfiguredServer(t *testing.T) {
	setup.Config(t)

	mailer, err := utils.NewSystemMailer(utils.NewEmailUtil())
	require.NoError(t, err)

	before := len(fetchInbox(t, senderAddress, senderPassword))

	email, err := emailTemplates.GetVerificationEmail(emailTemplates.VerificationEmailData{
		VerifyURL:     "https://app.autoreach.test/verify-email?token=x",
		ValidForHours: 48,
	})
	require.NoError(t, err)
	require.NoError(t, mailer.Send(context.Background(), senderAddress, email))

	inbox := fetchInbox(t, senderAddress, senderPassword)
	require.Len(t, inbox, before+1)
	assert.Equal(t, config.GetSystemMailAddress(), inbox[len(inbox)-1].FromAddress)
}
