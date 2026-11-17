package emailTemplates

import (
	htmlTemplate "html/template"
	textTemplate "text/template"

	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
)

const AccountExistsEmailSubject = "Your AutoReach account"

const accountExistsEmailText = `Someone tried to create an AutoReach account with this address, which already has one.

If it was you, sign in instead. If not, you can ignore this email; nothing about your account changed.
`

const accountExistsEmailHTML = `<!doctype html>
<html>
  <body style="margin:0;padding:24px;font-family:Helvetica,Arial,sans-serif;font-size:14px;line-height:22px;color:#18181b;">
    <p style="margin:0 0 16px;">Someone tried to create an AutoReach account with this address, which already has one.</p>
    <p style="margin:0;color:#52525b;">If it was you, sign in instead. If not, you can ignore this email; nothing about your account changed.</p>
  </body>
</html>
`

var (
	accountExistsText = textTemplate.Must(
		textTemplate.New("accountExistsText").Parse(accountExistsEmailText),
	)
	accountExistsHTML = htmlTemplate.Must(
		htmlTemplate.New("accountExistsHTML").Parse(accountExistsEmailHTML),
	)
)

func GetAccountExistsEmail() (sharedTypes.AccountEmail, error) {
	return renderAccountEmail(
		AccountExistsEmailSubject,
		nil,
		accountExistsText.Execute,
		accountExistsHTML.Execute,
	)
}
