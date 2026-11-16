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
