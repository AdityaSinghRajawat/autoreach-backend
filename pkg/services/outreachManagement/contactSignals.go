package outreachManagement

import (
	"context"
	"strings"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// reportContactSignal stores what one candidate observed about a contact everyone shares, and
// applies it only once enough candidates observed the same thing, so a single mailbox cannot
// rewrite the contact or suppress it for everyone. It reports whether the signal was applied.
func (s *OutreachService) reportContactSignal(
	ctx context.Context,
	candidateID, contactID string,
	signalType modelHelpers.ContactSignalType,
	address string,
	at time.Time,
) (bool, error) {
	signal := enrichmentModel.NewContactSignal()
	signal.ContactID = contactID
	signal.CandidateID = candidateID
	signal.SignalType = signalType
	signal.Address = modelHelpers.NormalizeEmail(address)
	signal.OccurredAt = at
	signal.UpdateTimestamp()

	if err := signal.Validate(); err != nil {
		utils.LogWarn(ctx, "contact signal skipped", "contact_id", contactID, "error", err.Error())

		return false, nil
	}

	if _, err := s.contactSignalRepo.Upsert(ctx, signal); err != nil {
		return false, err
	}

	reporters, err := s.contactSignalRepo.CountReporters(ctx, signal)
	if err != nil || reporters < config.GetContactSignalConfirmations() {
		return false, err
	}

	return true, s.applyContactSignal(ctx, signal)
}

// applyContactSignal changes the shared contact: a bounce suppresses the address for everyone and
// drops it from the contact, a reply or LinkedIn response confirms the contact.
func (s *OutreachService) applyContactSignal(
	ctx context.Context,
	signal *enrichmentModel.ContactSignal,
) error {
	var err error

	switch signal.SignalType {
	case modelHelpers.ContactSignalBounced:
		err = s.suppressAddress(ctx, nil, signal.Address, modelHelpers.SuppressionReasonBounced)
		if err == nil {
			err = s.updateContact(ctx, signal.ContactID,
				func(contact *enrichmentModel.CompanyContact) {
					if contact.WorkEmail != nil &&
						strings.EqualFold(*contact.WorkEmail, signal.Address) {
						contact.RecordBounce(signal.OccurredAt)
					}
				})
		}
	case modelHelpers.ContactSignalReplied:
		err = s.updateContact(ctx, signal.ContactID, func(contact *enrichmentModel.CompanyContact) {
			contact.RecordReply(signal.Address, signal.OccurredAt)
		})
	case modelHelpers.ContactSignalLinkedinResponded:
		err = s.updateContact(ctx, signal.ContactID, func(contact *enrichmentModel.CompanyContact) {
			contact.RecordLinkedinResponse(signal.OccurredAt)
		})
	}

	if err != nil {
		return err
	}

	return s.contactSignalRepo.MarkApplied(ctx, signal, utils.GetCurrTimeStamp())
}
