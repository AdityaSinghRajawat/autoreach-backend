package utils

import (
	"html"
	"net/url"
	"regexp"
	"slices"
	"strings"
)

var (
	blockTagPatternI    = regexp.MustCompile(`(?i)<\s*(br|/p|/div|/li|/h[1-6]|/tr)\b[^>]*>`)
	cellTagPatternI     = regexp.MustCompile(`(?i)<\s*/t[dh]\b[^>]*>`)
	anyTagPatternI      = regexp.MustCompile(`<[^>]*>`)
	blankLinesI         = regexp.MustCompile(`\n\s*\n+`)
	inlineSpacePatternI = regexp.MustCompile(`[ \t\f\v\x{00A0}]+`)
)

// Board APIs return HTML (Greenhouse escapes it once more), so it is unescaped before tags go.
func HTMLToText(markup string) string {
	text := html.UnescapeString(markup)
	text = blockTagPatternI.ReplaceAllString(text, "\n")
	text = cellTagPatternI.ReplaceAllString(text, " ")
	text = html.UnescapeString(anyTagPatternI.ReplaceAllString(text, ""))
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

var placeholderPatternI = regexp.MustCompile(`\{\{\s*([A-Za-z]+)\s*\}\}`)

func UnknownPlaceholders(text string, known []string) []string {
	unknown := []string{}

	for _, match := range placeholderPatternI.FindAllStringSubmatch(text, -1) {
		if !slices.Contains(known, match[1]) && !slices.Contains(unknown, match[1]) {
			unknown = append(unknown, match[1])
		}
	}

	return unknown
}

func RenderPlaceholders(text string, values map[string]string) string {
	return placeholderPatternI.ReplaceAllStringFunc(text, func(token string) string {
		return values[placeholderPatternI.FindStringSubmatch(token)[1]]
	})
}

func TrimOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	text := strings.TrimSpace(*value)

	return &text
}

func TrimLowerOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	text := strings.ToLower(strings.TrimSpace(*value))

	return &text
}

// A blank value becomes nil, so it is stored as absent rather than as an empty string.
func TrimOptionalStringToNil(value *string) *string {
	if value == nil {
		return nil
	}

	text := strings.TrimSpace(*value)
	if text == "" {
		return nil
	}

	return &text
}

func GetFirstName(fullName string) string {
	first, _ := SplitFullName(fullName)

	return first
}

// The first word is the first name; every other word is the last name.
func SplitFullName(fullName string) (string, string) {
	fields := strings.Fields(fullName)
	if len(fields) == 0 {
		return "", ""
	}

	return fields[0], strings.Join(fields[1:], " ")
}

// Reads as "a, b and c".
func JoinWithAnd(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	default:
		return strings.Join(items[:len(items)-1], ", ") + " and " + items[len(items)-1]
	}
}

func TruncateRunes(text string, limit int) string {
	characters := []rune(text)
	if len(characters) > limit {
		characters = characters[:limit]
	}

	return string(characters)
}

// Values are trimmed, then optionally transformed; blanks and repeats are dropped and the result is capped.
func UniqueTrimmedValues(values []string, transform func(string) string, limit int) []string {
	unique := make([]string, 0, min(len(values), limit))

	for _, value := range values {
		value = strings.TrimSpace(value)
		if transform != nil {
			value = transform(value)
		}

		if value != "" && !slices.Contains(unique, value) && len(unique) < limit {
			unique = append(unique, value)
		}
	}

	return unique
}

func FirstOrZero[T any](values []T) T {
	var zero T
	if len(values) == 0 {
		return zero
	}

	return values[0]
}

func IsHTTPURL(link string) bool {
	parsed, err := url.Parse(link)

	return err == nil && (parsed.Scheme == "https" || parsed.Scheme == "http") && parsed.Host != ""
}
