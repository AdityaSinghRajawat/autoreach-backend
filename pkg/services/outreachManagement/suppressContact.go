package outreachManagement

import (
	"context"
	"strings"

	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

func (s *OutreachService) suppressContact(
	ctx context.Context,
	candidateID *string,
	contactID string,
	reason modelHelpers.SuppressionReason,
) error {
	contact, err := s.companyContactRepo.FindByID(ctx, contactID)
	if err != nil || contact == nil || contact.WorkEmail == nil {
		return err
	}

	return s.suppressAddress(ctx, candidateID, *contact.WorkEmail, reason)
}

// suppressAddress blocks an address for one candidate, or for everyone when candidateID is nil.
func (s *OutreachService) suppressAddress(
	ctx context.Context,
	candidateID *string,
	address string,
	reason modelHelpers.SuppressionReason,
) error {
	entry := outreachModel.NewSuppressionEntry()
	entry.CandidateID = candidateID
	entry.EntryType = modelHelpers.SuppressionTypeEmail
	entry.Value = strings.ToLower(address)
	entry.Reason = reason

	if err := entry.Validate(); err != nil {
		return err
	}

	return s.suppressionRepo.Insert(ctx, entry)
}
