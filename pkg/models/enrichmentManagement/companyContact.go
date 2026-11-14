package enrichmentManagement

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type CompanyContact struct {
	ID                      string                               `db:"id"                        json:"contactId"               validate:"required,uuid"`
	CompanyDomain           string                               `db:"company_domain"            json:"companyDomain"           validate:"required,fqdn,max=255"`
	CompanyName             string                               `db:"company_name"              json:"companyName"             validate:"required,max=255"`
	FullName                string                               `db:"full_name"                 json:"fullName"                validate:"required,max=255"`
	RoleType                modelHelpers.ContactRole             `db:"role_type"                 json:"roleType"                validate:"required,oneof=ENGINEERING_MANAGER HIRING_MANAGER RECRUITER"`
	Title                   string                               `db:"title"                     json:"title"                   validate:"required,max=255"`
	WorkEmail               *string                              `db:"work_email"                json:"workEmail"               validate:"omitnil,email,max=255"`
	PhoneNumber             *string                              `db:"phone_number"              json:"phoneNumber"             validate:"omitnil,max=50"`
	LinkedinURL             *string                              `db:"linkedin_url"              json:"linkedinUrl"             validate:"omitnil,max=500,validate_linkedin_url"`
	EnrichmentProvider      string                               `db:"enrichment_provider"       json:"enrichmentProvider"      validate:"required,oneof=APOLLO SIGNALHIRE CONTACTOUT"`
	ProviderPersonID        *string                              `db:"provider_person_id"        json:"-"                       validate:"omitnil,max=255"`
	IsEmailVerified         bool                                 `db:"is_email_verified"         json:"isEmailVerified"`
	EmailVerificationStatus modelHelpers.EmailVerificationStatus `db:"email_verification_status" json:"emailVerificationStatus" validate:"required,oneof=UNVERIFIED VERIFIED RISKY INVALID"`
	EnrichmentMetadata      json.RawMessage                      `db:"enrichment_metadata"       json:"-"                       validate:"required,validate_json_object"`
	CreatedAt               time.Time                            `db:"created_at"                json:"createdAt"               validate:"required"`
	UpdatedAt               time.Time                            `db:"updated_at"                json:"updatedAt"               validate:"required"`
}

func NewCompanyContact() *CompanyContact {
	now := utils.GetCurrTimeStamp()

	return &CompanyContact{
		ID:                      uuid.NewString(),
		EmailVerificationStatus: modelHelpers.EmailStatusUnverified,
		EnrichmentMetadata:      json.RawMessage(config.GetEmptyJSONObject()),
		CreatedAt:               now,
		UpdatedAt:               now,
	}
}

func (c *CompanyContact) UpdateTimestamp() {
	c.UpdatedAt = utils.GetCurrTimeStamp()
}

func (c *CompanyContact) Validate() error {
	if err := validations.Validate(c); err != nil {
		return fmt.Errorf("company contact validation failed: %w", err)
	}

	return nil
}

func (c *CompanyContact) IsReachable() bool {
	return (c.WorkEmail != nil && c.IsEmailVerified) || c.LinkedinURL != nil
}

func (c *CompanyContact) SetEmail(email string, status modelHelpers.EmailVerificationStatus) {
	c.EmailVerificationStatus = status
	c.IsEmailVerified = status == modelHelpers.EmailStatusVerified
	c.WorkEmail = nil

	if email != "" && status != modelHelpers.EmailStatusInvalid {
		c.WorkEmail = &email
	}
}

func (c *CompanyContact) AdoptIdentity(existing *CompanyContact) {
	c.ID = existing.ID
	c.CreatedAt = existing.CreatedAt

	if c.WorkEmail == nil && existing.WorkEmail != nil {
		c.WorkEmail = existing.WorkEmail
		c.IsEmailVerified = existing.IsEmailVerified
		c.EmailVerificationStatus = existing.EmailVerificationStatus
	}

	if c.LinkedinURL == nil {
		c.LinkedinURL = existing.LinkedinURL
	}

	if c.PhoneNumber == nil {
		c.PhoneNumber = existing.PhoneNumber
	}
}
