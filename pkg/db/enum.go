package db

import "strings"

// Enum writes fixed enum values into SQL text instead of binding them, so the planner can still
// match partial indexes such as WHERE status = 'PENDING'. Only pass compile-time constants.
func Enum[T ~string](values ...T) string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, "'"+strings.ReplaceAll(string(value), "'", "''")+"'")
	}

	return strings.Join(quoted, ", ")
}
