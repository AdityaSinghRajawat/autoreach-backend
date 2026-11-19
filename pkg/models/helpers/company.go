package helpers

import (
	"regexp"
	"strings"
)

var (
	companyNameSeparatorPattern = regexp.MustCompile(`[^a-z0-9]+`)
	companyGenericSuffixPattern = regexp.MustCompile(
		`(\s(technologies|technology|tech|enterprises|software|services|solutions|systems|consulting|consultancy|group|global|india))+$`,
	)
	companyLegalSuffixPattern = regexp.MustCompile(
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

// CompanyNamesMatch tells whether two names identify the same company: spacing and legal or generic
// trailing words ("Technologies", "Pvt Ltd") are ignored, but any other difference is not.
func CompanyNamesMatch(first, second string) bool {
	left, right := NormalizeCompanyName(first), NormalizeCompanyName(second)
	if left == "" || right == "" {
		return false
	}

	if strings.ReplaceAll(left, " ", "") == strings.ReplaceAll(right, " ", "") {
		return true
	}

	return withoutGenericSuffix(left) == withoutGenericSuffix(right)
}

func withoutGenericSuffix(name string) string {
	if stripped := strings.TrimSpace(
		companyGenericSuffixPattern.ReplaceAllString(" "+name, ""),
	); stripped != "" {
		return stripped
	}

	return name
}
