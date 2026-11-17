package db

import "strings"

// Literal, not bound, so partial indexes still match. Pass compile-time constants only.
func Enum[T ~string](values ...T) string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, "'"+strings.ReplaceAll(string(value), "'", "''")+"'")
	}

	return strings.Join(quoted, ", ")
}
