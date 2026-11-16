package utils

import (
	"html"
	"regexp"
	"strings"
)

var (
	blockTagPatternI    = regexp.MustCompile(`(?i)<\s*(br|/p|/div|/li|/h[1-6]|/tr)\b[^>]*>`)
	anyTagPatternI      = regexp.MustCompile(`<[^>]*>`)
	blankLinesI         = regexp.MustCompile(`\n\s*\n+`)
	inlineSpacePatternI = regexp.MustCompile(`[ \t\f\v\x{00A0}]+`)
)

// Board APIs return HTML (Greenhouse escapes it once more), so it is unescaped before tags go.
func HTMLToText(markup string) string {
	text := html.UnescapeString(markup)
	text = blockTagPatternI.ReplaceAllString(text, "\n")
	text = html.UnescapeString(anyTagPatternI.ReplaceAllString(text, " "))
	text = inlineSpacePatternI.ReplaceAllString(text, " ")

	lines := strings.Split(text, "\n")
	for index, line := range lines {
		lines[index] = strings.TrimSpace(line)
	}

	return strings.TrimSpace(blankLinesI.ReplaceAllString(strings.Join(lines, "\n"), "\n\n"))
}

func NilIfEmpty(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}

func FirstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}
