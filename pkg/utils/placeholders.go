package utils

import (
	"regexp"
	"slices"
)

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
