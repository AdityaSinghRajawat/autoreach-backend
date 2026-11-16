package helpers

import (
	"regexp"
	"strings"
)

var (
	companyNameSeparatorPattern = regexp.MustCompile(`[^a-z0-9]+`)
	companyLegalSuffixPattern   = regexp.MustCompile(
		`(\s(inc|incorporated|llc|llp|ltd|limited|corp|corporation|co|company|plc|gmbh|ag|sa|bv|nv|pvt|private|pte|pty))+$`,
	)
)

// NormalizeCompanyName reduces a company name to the words that identify it, so "Acme, Inc." and
// "ACME Inc" compare equal.
func NormalizeCompanyName(name string) string {
	normalized := strings.TrimSpace(
		companyNameSeparatorPattern.ReplaceAllString(strings.ToLower(name), " "),
	)

	if withoutSuffix := strings.TrimSpace(
		companyLegalSuffixPattern.ReplaceAllString(" "+normalized, ""),
	); withoutSuffix != "" {
		return withoutSuffix
	}

	return normalized
}
