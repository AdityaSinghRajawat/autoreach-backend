package helpers

import (
	"strings"
	"time"
)

type ContactRole string

const (
	ContactRoleEngineeringManager ContactRole = "ENGINEERING_MANAGER"
	ContactRoleHiringManager      ContactRole = "HIRING_MANAGER"
	ContactRoleRecruiter          ContactRole = "RECRUITER"
)

type EmailVerificationStatus string

const (
	EmailStatusUnverified EmailVerificationStatus = "UNVERIFIED"
	EmailStatusVerified   EmailVerificationStatus = "VERIFIED"
	EmailStatusRisky      EmailVerificationStatus = "RISKY"
	EmailStatusInvalid    EmailVerificationStatus = "INVALID"
)

type EnrichmentStatus string

const (
	EnrichmentStatusPending   EnrichmentStatus = "PENDING"
	EnrichmentStatusCompleted EnrichmentStatus = "COMPLETED"
	EnrichmentStatusFailed    EnrichmentStatus = "FAILED"
)

type ContactChangeReason string

const (
	ContactChangeEnrichment       ContactChangeReason = "ENRICHMENT"
	ContactChangeBounced          ContactChangeReason = "BOUNCED"
	ContactChangeRepliedFromNew   ContactChangeReason = "REPLIED_FROM_NEW_ADDRESS"
	ContactChangeRepliedFromOther ContactChangeReason = "REPLIED_FROM_OTHER_ADDRESS"
	ContactChangeRefreshed        ContactChangeReason = "REFRESHED_AFTER_BOUNCE"
)

const (
	ContactFieldWorkEmail      = "workEmail"
	ContactFieldAlternateEmail = "alternateEmail"
	ContactFieldTitle          = "title"
	ContactFieldPhone          = "phoneNumber"
	ContactFieldCompany        = "companyDomain"
)

// ContactChange records one value a contact used to have, or an extra one seen in an interaction.
type ContactChange struct {
	Field     string              `json:"field"`
	Previous  string              `json:"previous,omitempty"`
	Current   string              `json:"current,omitempty"`
	Reason    ContactChangeReason `json:"reason"`
	ChangedAt time.Time           `json:"changedAt"`
}

type ContactSignalType string

const (
	ContactSignalBounced           ContactSignalType = "BOUNCED"
	ContactSignalReplied           ContactSignalType = "REPLIED"
	ContactSignalLinkedinResponded ContactSignalType = "LINKEDIN_RESPONDED"
)

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
