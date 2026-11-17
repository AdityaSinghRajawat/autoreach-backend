package unit

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/emailTemplates"
)

func TestVerificationEmailCarriesTheLinkInTextAndHTML(t *testing.T) {
	link := "https://app.autoreach.test/verify-email?token=abc.123&x=<y>"

	email, err := emailTemplates.GetVerificationEmail(emailTemplates.VerificationEmailData{
		VerifyURL:     link,
		ValidForHours: 48,
	})
	require.NoError(t, err)

	assert.Equal(t, emailTemplates.VerificationEmailSubject, email.Subject)
	assert.Contains(t, email.Text, link, "the plain text keeps the link exactly")
	assert.Contains(t, email.Text, "48 hours")
	assert.Contains(t, email.HTML, "token=abc.123")
	assert.False(t, strings.Contains(email.HTML, "<y>"), "the HTML version escapes the link")
}

func TestAccountExistsEmailRevealsNothingAboutTheAccount(t *testing.T) {
	email, err := emailTemplates.GetAccountExistsEmail()
	require.NoError(t, err)

	assert.Equal(t, emailTemplates.AccountExistsEmailSubject, email.Subject)
	assert.NotEmpty(t, email.Text)
	assert.NotEmpty(t, email.HTML)
}

func TestPasswordResetEmailCarriesTheLinkAndItsLifetime(t *testing.T) {
	link := "https://app.autoreach.test/reset-password?token=abc.123"

	email, err := emailTemplates.GetPasswordResetEmail(emailTemplates.PasswordResetEmailData{
		ResetURL:        link,
		ValidForMinutes: 60,
	})
	require.NoError(t, err)

	assert.Equal(t, emailTemplates.PasswordResetEmailSubject, email.Subject)
	assert.Contains(t, email.Text, link)
	assert.Contains(t, email.Text, "60 minutes")
	assert.Contains(t, email.HTML, "token=abc.123")
}

func TestEmailChangeEmailsCarryTheLinkAndTheNotice(t *testing.T) {
	link := "https://app.autoreach.test/confirm-email?email=alex%40example.com&token=abc"

	confirm, err := emailTemplates.GetEmailChangeEmail(emailTemplates.EmailChangeEmailData{
		ConfirmURL:    link,
		ValidForHours: 48,
	})
	require.NoError(t, err)
	assert.Equal(t, emailTemplates.EmailChangeEmailSubject, confirm.Subject)
	assert.Contains(t, confirm.Text, link)

	notice, err := emailTemplates.GetEmailChangedEmail()
	require.NoError(t, err)
	assert.Equal(t, emailTemplates.EmailChangedEmailSubject, notice.Subject)
	assert.NotEmpty(t, notice.HTML)
}
