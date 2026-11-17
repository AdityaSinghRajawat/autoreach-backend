package emailTemplates

import (
	htmlTemplate "html/template"
	textTemplate "text/template"

	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
)

const EmailChangedEmailSubject = "Your AutoReach email was changed"

const emailChangedEmailText = `The email address of your AutoReach account was changed, so this address no longer signs in.

If you did not make this change, reply to this email straight away so we can secure your account.
`

const emailChangedEmailHTML = `<!doctype html>
<html>
  <body style="margin:0;padding:24px;font-family:Helvetica,Arial,sans-serif;font-size:14px;line-height:22px;color:#18181b;">
    <p style="margin:0 0 16px;">The email address of your AutoReach account was changed, so this address no longer signs in.</p>
    <p style="margin:0;color:#52525b;">If you did not make this change, reply to this email straight away so we can secure your account.</p>
  </body>
</html>
`

var (
	emailChangedText = textTemplate.Must(
		textTemplate.New("emailChangedText").Parse(emailChangedEmailText),
	)
	emailChangedHTML = htmlTemplate.Must(
		htmlTemplate.New("emailChangedHTML").Parse(emailChangedEmailHTML),
	)
)

func GetEmailChangedEmail() (sharedTypes.AccountEmail, error) {
	return renderAccountEmail(
		EmailChangedEmailSubject,
		nil,
		emailChangedText.Execute,
		emailChangedHTML.Execute,
	)
}
