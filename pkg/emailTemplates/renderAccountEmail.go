package emailTemplates

import (
	"fmt"
	"io"
	"strings"

	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
)

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
