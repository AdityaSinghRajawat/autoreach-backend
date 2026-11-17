package emailTemplates

import (
	htmlTemplate "html/template"
	textTemplate "text/template"

	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
)

type PasswordResetEmailData struct {
	ResetURL        string
	ValidForMinutes int
}

const PasswordResetEmailSubject = "Reset your AutoReach password"

const passwordResetEmailText = `Someone asked to reset the password of the AutoReach account for this address.

Choose a new password here:

{{.ResetURL}}

The link works for {{.ValidForMinutes}} minutes and only once. Resetting signs you out everywhere.
If you did not ask for this, ignore this email; your password has not changed.
`

const passwordResetEmailHTML = `<!doctype html>
<html>
  <body style="margin:0;padding:24px;font-family:Helvetica,Arial,sans-serif;font-size:14px;line-height:22px;color:#18181b;">
    <p style="margin:0 0 16px;">Someone asked to reset the password of the AutoReach account for this address.</p>
    <p style="margin:0 0 16px;">
      <a href="{{.ResetURL}}" style="display:inline-block;padding:10px 18px;border-radius:6px;background:#2f5bd3;color:#ffffff;text-decoration:none;">Choose a new password</a>
    </p>
    <p style="margin:0 0 16px;color:#52525b;">Or open this link: <a href="{{.ResetURL}}">{{.ResetURL}}</a></p>
    <p style="margin:0;color:#52525b;">The link works for {{.ValidForMinutes}} minutes and only once. Resetting signs you out everywhere. If you did not ask for this, ignore this email; your password has not changed.</p>
  </body>
</html>
`

var (
	passwordResetText = textTemplate.Must(
		textTemplate.New("passwordResetText").Parse(passwordResetEmailText),
	)
	passwordResetHTML = htmlTemplate.Must(
		htmlTemplate.New("passwordResetHTML").Parse(passwordResetEmailHTML),
	)
)

func GetPasswordResetEmail(data PasswordResetEmailData) (sharedTypes.AccountEmail, error) {
	return renderAccountEmail(
		PasswordResetEmailSubject,
		data,
		passwordResetText.Execute,
		passwordResetHTML.Execute,
	)
}
