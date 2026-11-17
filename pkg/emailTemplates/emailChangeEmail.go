package emailTemplates

import (
	htmlTemplate "html/template"
	textTemplate "text/template"

	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
)

type EmailChangeEmailData struct {
	ConfirmURL    string
	ValidForHours int
}

const EmailChangeEmailSubject = "Confirm your new AutoReach email"

const emailChangeEmailText = `Confirm this address to use it for your AutoReach account:

{{.ConfirmURL}}

The link works for {{.ValidForHours}} hours. If you did not ask for this, ignore this email.
`

const emailChangeEmailHTML = `<!doctype html>
<html>
  <body style="margin:0;padding:24px;font-family:Helvetica,Arial,sans-serif;font-size:14px;line-height:22px;color:#18181b;">
    <p style="margin:0 0 16px;">Confirm this address to use it for your AutoReach account.</p>
    <p style="margin:0 0 16px;">
      <a href="{{.ConfirmURL}}" style="display:inline-block;padding:10px 18px;border-radius:6px;background:#2f5bd3;color:#ffffff;text-decoration:none;">Confirm new email</a>
    </p>
    <p style="margin:0 0 16px;color:#52525b;">Or open this link: <a href="{{.ConfirmURL}}">{{.ConfirmURL}}</a></p>
    <p style="margin:0;color:#52525b;">The link works for {{.ValidForHours}} hours. If you did not ask for this, ignore this email.</p>
  </body>
</html>
`

var (
	emailChangeText = textTemplate.Must(
		textTemplate.New("emailChangeText").Parse(emailChangeEmailText),
	)
	emailChangeHTML = htmlTemplate.Must(
		htmlTemplate.New("emailChangeHTML").Parse(emailChangeEmailHTML),
	)
)

func GetEmailChangeEmail(data EmailChangeEmailData) (sharedTypes.AccountEmail, error) {
	return renderAccountEmail(
		EmailChangeEmailSubject,
		data,
		emailChangeText.Execute,
		emailChangeHTML.Execute,
	)
}
