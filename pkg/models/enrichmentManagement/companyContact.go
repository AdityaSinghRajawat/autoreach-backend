package enrichmentManagement

import (
	"encoding/json"
	"fmt"
	"strings"
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
	LastInteractionAt       *time.Time                           `db:"last_interaction_at" json:"lastInteractionAt"`
	EmailConfirmedAt        *time.Time                           `db:"email_confirmed_at" json:"emailConfirmedAt"`
	LinkedinConfirmedAt     *time.Time                           `db:"linkedin_confirmed_at" json:"linkedinConfirmedAt"`
	AnonymizedAt            *time.Time                           `db:"anonymized_at" json:"anonymizedAt"`
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

// AdoptIdentity turns freshly enriched details into an update of the stored contact: what the
// interactions established is kept, and every value that changes goes into the contact's history.
func (c *CompanyContact) AdoptIdentity(existing *CompanyContact, now time.Time) {
	c.ID = existing.ID
	c.CreatedAt = existing.CreatedAt
	c.LastInteractionAt = existing.LastInteractionAt
	c.LinkedinConfirmedAt = existing.LinkedinConfirmedAt
	c.setHistory(existing.History())

	if c.WorkEmail != nil && existing.HasBounced(*c.WorkEmail) {
		c.SetEmail("", modelHelpers.EmailStatusInvalid)
	}

	c.recordChange(modelHelpers.ContactFieldCompany, existing.CompanyDomain, c.CompanyDomain,
		modelHelpers.ContactChangeEnrichment, now)
	c.recordChange(
		modelHelpers.ContactFieldTitle,
		existing.Title,
		c.Title,
		modelHelpers.ContactChangeEnrichment,
		now,
	)

	if c.WorkEmail != nil && existing.WorkEmail != nil {
		c.recordChange(modelHelpers.ContactFieldWorkEmail, *existing.WorkEmail, *c.WorkEmail,
			modelHelpers.ContactChangeEnrichment, now)
	}

	if c.PhoneNumber != nil && existing.PhoneNumber != nil {
		c.recordChange(modelHelpers.ContactFieldPhone, *existing.PhoneNumber, *c.PhoneNumber,
			modelHelpers.ContactChangeEnrichment, now)
	}

	if c.WorkEmail != nil && existing.WorkEmail != nil &&
		strings.EqualFold(*c.WorkEmail, *existing.WorkEmail) {
		c.EmailConfirmedAt = existing.EmailConfirmedAt
	}

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

// RecordReply notes a reply. An address that matches the stored one confirms it; a new address at
// the contact's company replaces it; any other address is kept as an alternate and the stored
// address still counts as confirmed, since the message reached the person.
func (c *CompanyContact) RecordReply(address string, at time.Time) {
	c.LastInteractionAt = &at

	address = strings.ToLower(strings.TrimSpace(address))
	switch {
	case address == "" || (c.WorkEmail != nil && strings.EqualFold(*c.WorkEmail, address)):
		c.confirmEmail(at)
	case c.isCompanyAddress(address):
		previous := ""
		if c.WorkEmail != nil {
			previous = *c.WorkEmail
		}

		c.recordChange(modelHelpers.ContactFieldWorkEmail, previous, address,
			modelHelpers.ContactChangeRepliedFromNew, at)
		c.SetEmail(address, modelHelpers.EmailStatusVerified)
		c.confirmEmail(at)
	default:
		c.appendHistory(modelHelpers.ContactChange{
			Field:     modelHelpers.ContactFieldAlternateEmail,
			Current:   address,
			Reason:    modelHelpers.ContactChangeRepliedFromOther,
			ChangedAt: at,
		})
		c.confirmEmail(at)
	}
}

func (c *CompanyContact) RecordLinkedinResponse(at time.Time) {
	c.LastInteractionAt = &at
	c.LinkedinConfirmedAt = &at
}

// RecordBounce drops an address that bounced; HasBounced keeps it from being adopted again.
func (c *CompanyContact) RecordBounce(at time.Time) {
	if c.WorkEmail == nil {
		return
	}

	c.appendHistory(modelHelpers.ContactChange{
		Field:     modelHelpers.ContactFieldWorkEmail,
		Previous:  *c.WorkEmail,
		Reason:    modelHelpers.ContactChangeBounced,
		ChangedAt: at,
	})
	c.SetEmail("", modelHelpers.EmailStatusInvalid)
	c.EmailConfirmedAt = nil
}

// ReplaceEmail adopts an address found for the contact, recording the one it replaces.
func (c *CompanyContact) ReplaceEmail(
	address string,
	status modelHelpers.EmailVerificationStatus,
	reason modelHelpers.ContactChangeReason,
	at time.Time,
) {
	previous := ""
	if c.WorkEmail != nil {
		previous = *c.WorkEmail
	}

	c.appendHistory(modelHelpers.ContactChange{
		Field:     modelHelpers.ContactFieldWorkEmail,
		Previous:  previous,
		Current:   address,
		Reason:    reason,
		ChangedAt: at,
	})
	c.SetEmail(address, status)
	c.EmailConfirmedAt = nil
}

func (c *CompanyContact) HasBounced(address string) bool {
	for _, change := range c.History() {
		if change.Reason == modelHelpers.ContactChangeBounced &&
			strings.EqualFold(change.Previous, address) {
			return true
		}
	}

	return false
}

func (c *CompanyContact) History() []modelHelpers.ContactChange {
	metadata := map[string]json.RawMessage{}
	history := []modelHelpers.ContactChange{}

	if json.Unmarshal(c.EnrichmentMetadata, &metadata) == nil {
		_ = json.Unmarshal(metadata[config.GetContactHistoryKey()], &history)
	}

	return history
}

func (c *CompanyContact) confirmEmail(at time.Time) {
	if c.WorkEmail == nil {
		return
	}

	c.IsEmailVerified = true
	c.EmailVerificationStatus = modelHelpers.EmailStatusVerified
	c.EmailConfirmedAt = &at
}

func (c *CompanyContact) isCompanyAddress(address string) bool {
	_, domain, found := strings.Cut(address, "@")

	return found && (domain == c.CompanyDomain || strings.HasSuffix(domain, "."+c.CompanyDomain))
}

func (c *CompanyContact) recordChange(
	field, previous, current string,
	reason modelHelpers.ContactChangeReason,
	at time.Time,
) {
	if previous == "" || strings.EqualFold(previous, current) {
		return
	}

	c.appendHistory(modelHelpers.ContactChange{
		Field:     field,
		Previous:  previous,
		Current:   current,
		Reason:    reason,
		ChangedAt: at,
	})
}

func (c *CompanyContact) appendHistory(change modelHelpers.ContactChange) {
	history := append(c.History(), change)
	if limit := config.GetContactHistoryLimit(); len(history) > limit {
		history = history[len(history)-limit:]
	}

	c.setHistory(history)
}

func (c *CompanyContact) setHistory(history []modelHelpers.ContactChange) {
	metadata := map[string]json.RawMessage{}
	_ = json.Unmarshal(c.EnrichmentMetadata, &metadata)

	if len(history) == 0 {
		delete(metadata, config.GetContactHistoryKey())
	} else if encoded, err := json.Marshal(history); err == nil {
		metadata[config.GetContactHistoryKey()] = encoded
	}

	if encoded, err := json.Marshal(metadata); err == nil {
		c.EnrichmentMetadata = encoded
	}
}
