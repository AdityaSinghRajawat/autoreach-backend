package helpers

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
