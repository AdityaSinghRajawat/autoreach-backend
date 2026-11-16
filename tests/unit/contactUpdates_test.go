package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
)

func storedContact(email string) *enrichmentModel.CompanyContact {
	contact := enrichmentModel.NewCompanyContact()
	contact.CompanyDomain = "acme.com"
	contact.CompanyName = "Acme"
	contact.FullName = "Riley Recruiter"
	contact.Title = "Recruiter"
	contact.EnrichmentProvider = "APOLLO"
	contact.SetEmail(email, modelHelpers.EmailStatusRisky)

	return contact
}

func TestReplyFromTheStoredAddressConfirmsIt(t *testing.T) {
	setup.Config(t)

	now := time.Now().UTC()
	contact := storedContact("riley@acme.com")
	contact.RecordReply("Riley@Acme.com", now)

	assert.Equal(t, "riley@acme.com", *contact.WorkEmail)
	assert.True(t, contact.IsEmailVerified)
	assert.Equal(t, modelHelpers.EmailStatusVerified, contact.EmailVerificationStatus)
	assert.Equal(t, now, *contact.EmailConfirmedAt)
	assert.Equal(t, now, *contact.LastInteractionAt)
	assert.Empty(t, contact.History())
}

func TestReplyFromANewCompanyAddressReplacesTheEmail(t *testing.T) {
	setup.Config(t)

	now := time.Now().UTC()
	contact := storedContact("riley@acme.com")
	contact.RecordReply("riley.r@eu.acme.com", now)

	assert.Equal(t, "riley.r@eu.acme.com", *contact.WorkEmail)
	assert.True(t, contact.IsEmailVerified)

	history := contact.History()
	require.Len(t, history, 1)
	assert.Equal(t, modelHelpers.ContactChange{
		Field:     modelHelpers.ContactFieldWorkEmail,
		Previous:  "riley@acme.com",
		Current:   "riley.r@eu.acme.com",
		Reason:    modelHelpers.ContactChangeRepliedFromNew,
		ChangedAt: now,
	}, history[0])
}

func TestReplyFromAPersonalAddressKeepsTheWorkEmail(t *testing.T) {
	setup.Config(t)

	contact := storedContact("riley@acme.com")
	contact.RecordReply("riley.personal@gmail.com", time.Now().UTC())

	assert.Equal(
		t,
		"riley@acme.com",
		*contact.WorkEmail,
		"the work address reached them, so it stays",
	)
	assert.True(t, contact.IsEmailVerified)

	history := contact.History()
	require.Len(t, history, 1)
	assert.Equal(t, modelHelpers.ContactFieldAlternateEmail, history[0].Field)
	assert.Equal(t, "riley.personal@gmail.com", history[0].Current)
}

func TestBouncedAddressIsDroppedAndNeverAdoptedAgain(t *testing.T) {
	setup.Config(t)

	now := time.Now().UTC()
	contact := storedContact("riley@acme.com")
	contact.RecordReply("riley@acme.com", now)
	contact.RecordBounce(now)

	assert.Nil(t, contact.WorkEmail)
	assert.Equal(t, modelHelpers.EmailStatusInvalid, contact.EmailVerificationStatus)
	assert.Nil(t, contact.EmailConfirmedAt)
	assert.True(t, contact.HasBounced("RILEY@acme.com"))

	enriched := storedContact("riley@acme.com")
	enriched.AdoptIdentity(contact, now)
	assert.Nil(t, enriched.WorkEmail, "enrichment cannot bring back an address that bounced")
}

func TestEnrichmentKeepsInteractionsAndRecordsWhatChanged(t *testing.T) {
	setup.Config(t)

	earlier := time.Now().UTC().Add(-time.Hour)
	now := time.Now().UTC()
	existing := storedContact("riley@acme.com")
	existing.RecordReply("riley@acme.com", earlier)
	existing.RecordLinkedinResponse(earlier)

	moved := storedContact("riley@globex.com")
	moved.CompanyDomain = "globex.com"
	moved.Title = "Head of Talent"
	moved.AdoptIdentity(existing, now)

	assert.Equal(t, existing.ID, moved.ID)
	assert.Equal(t, earlier, *moved.LastInteractionAt)
	assert.Equal(t, earlier, *moved.LinkedinConfirmedAt)
	assert.Nil(t, moved.EmailConfirmedAt, "a new address has not been confirmed yet")

	changed := map[string]string{}
	for _, change := range moved.History() {
		changed[change.Field] = change.Previous + " -> " + change.Current
	}

	assert.Equal(t, map[string]string{
		modelHelpers.ContactFieldCompany:   "acme.com -> globex.com",
		modelHelpers.ContactFieldTitle:     "Recruiter -> Head of Talent",
		modelHelpers.ContactFieldWorkEmail: "riley@acme.com -> riley@globex.com",
	}, changed)
}
