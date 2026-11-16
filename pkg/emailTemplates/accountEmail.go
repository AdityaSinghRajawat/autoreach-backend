package emailTemplates

import (
	"fmt"
	htmlTemplate "html/template"
	"io"
	"strings"
	textTemplate "text/template"

	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
)

type VerificationEmailData struct {
	VerifyURL     string
	ValidForHours int
}

const (
	VerificationEmailSubject  = "Confirm your AutoReach email"
	AccountExistsEmailSubject = "Your AutoReach account"
)

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
	verificationText = textTemplate.Must(
		textTemplate.New("verificationText").Parse(verificationEmailText),
	)
	verificationHTML = htmlTemplate.Must(
		htmlTemplate.New("verificationHTML").Parse(verificationEmailHTML),
	)
	accountExistsText = textTemplate.Must(
		textTemplate.New("accountExistsText").Parse(accountExistsEmailText),
	)
	accountExistsHTML = htmlTemplate.Must(
		htmlTemplate.New("accountExistsHTML").Parse(accountExistsEmailHTML),
	)
)

func GetVerificationEmail(data VerificationEmailData) (sharedTypes.AccountEmail, error) {
	return renderAccountEmail(VerificationEmailSubject, data,
		verificationText.Execute, verificationHTML.Execute)
}

func GetAccountExistsEmail() (sharedTypes.AccountEmail, error) {
	return renderAccountEmail(AccountExistsEmailSubject, nil,
		accountExistsText.Execute, accountExistsHTML.Execute)
}

func renderAccountEmail(
	subject string,
	data any,
	renderText, renderHTML func(writer io.Writer, data any) error,
) (sharedTypes.AccountEmail, error) {
	var text, html strings.Builder

	if err := renderText(&text, data); err != nil {
		return sharedTypes.AccountEmail{}, fmt.Errorf("render %q text: %w", subject, err)
	}

	if err := renderHTML(&html, data); err != nil {
		return sharedTypes.AccountEmail{}, fmt.Errorf("render %q html: %w", subject, err)
	}

	return sharedTypes.AccountEmail{Subject: subject, Text: text.String(), HTML: html.String()}, nil
}
