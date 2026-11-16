package helpers

import (
	"regexp"
	"strings"
)

type JobStatus string

const (
	JobStatusActive   JobStatus = "ACTIVE"
	JobStatusClosed   JobStatus = "CLOSED"
	JobStatusArchived JobStatus = "ARCHIVED"
)

type ATSType string

const (
	ATSTypeGreenhouse      ATSType = "GREENHOUSE"
	ATSTypeLever           ATSType = "LEVER"
	ATSTypeWorkday         ATSType = "WORKDAY"
	ATSTypeAshby           ATSType = "ASHBY"
	ATSTypeSmartRecruiters ATSType = "SMARTRECRUITERS"
	ATSTypeOther           ATSType = "OTHER"
)

var (
	bracketedPattern   = regexp.MustCompile(`[\(\[\{][^\)\]\}]*[\)\]\}]`)
	requisitionPattern = regexp.MustCompile(
		`(?i)(\b(req|requisition|job|id|ref)\s*[-#:]?\s*\d{3,}\b|#\s*\d+|\b[a-z]{1,3}-?\d{4,}\b)`,
	)
	separatorPattern = regexp.MustCompile(`[^a-z0-9+#]+`)
)

func NormalizeJobTitle(title string) string {
	normalized := strings.ToLower(title)
	normalized = bracketedPattern.ReplaceAllString(normalized, " ")
	normalized = requisitionPattern.ReplaceAllString(normalized, " ")
	normalized = separatorPattern.ReplaceAllString(normalized, " ")

	return strings.TrimSpace(normalized)
}

func NormalizeDomain(domain string) string {
	normalized := strings.ToLower(strings.TrimSpace(domain))
	normalized = strings.TrimSuffix(normalized, ".")

	return strings.TrimPrefix(normalized, "www.")
}

type JobSourceRegion string

const (
	JobSourceRegionGlobal JobSourceRegion = "GLOBAL"
	JobSourceRegionEU     JobSourceRegion = "EU"
)
