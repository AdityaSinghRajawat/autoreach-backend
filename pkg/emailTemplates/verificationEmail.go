package emailTemplates

import (
	htmlTemplate "html/template"
	textTemplate "text/template"

	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
)

type VerificationEmailData struct {
	VerifyURL     string
	ValidForHours int
}

const VerificationEmailSubject = "Confirm your AutoReach email"

const verificationEmailText = `Confirm this address to finish creating your AutoReach account:

{{.VerifyURL}}

The link works for {{.ValidForHours}} hours. If you did not sign up, ignore this email.
`

const verificationEmailHTML = `<!doctype html>
<html>
  <body style="margin:0;padding:24px;font-family:Helvetica,Arial,sans-serif;font-size:14px;line-height:22px;color:#18181b;">
    <p style="margin:0 0 16px;">Confirm this address to finish creating your AutoReach account.</p>
    <p style="margin:0 0 16px;">
      <a href="{{.VerifyURL}}" style="display:inline-block;padding:10px 18px;border-radius:6px;background:#2f5bd3;color:#ffffff;text-decoration:none;">Confirm email</a>
    </p>
    <p style="margin:0 0 16px;color:#52525b;">Or open this link: <a href="{{.VerifyURL}}">{{.VerifyURL}}</a></p>
    <p style="margin:0;color:#52525b;">The link works for {{.ValidForHours}} hours. If you did not sign up, ignore this email.</p>
  </body>
</html>
`

var (
	verificationText = textTemplate.Must(
		textTemplate.New("verificationText").Parse(verificationEmailText),
	)
	verificationHTML = htmlTemplate.Must(
		htmlTemplate.New("verificationHTML").Parse(verificationEmailHTML),
	)
)

func GetVerificationEmail(data VerificationEmailData) (sharedTypes.AccountEmail, error) {
	return renderAccountEmail(
		VerificationEmailSubject,
		data,
		verificationText.Execute,
		verificationHTML.Execute,
	)
}
