package outreachManagement

import (
	"context"
	"strings"

	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// An address held by another contact at the company is not adopted.
func (s *OutreachService) updateContact(
	ctx context.Context,
	contactID string,
	change func(contact *enrichmentModel.CompanyContact),
) error {
	contact, err := s.companyContactRepo.FindByID(ctx, contactID)
	if err != nil || contact == nil {
		return err
	}

	previousEmail := contact.WorkEmail
	change(contact)

	if contact.WorkEmail != nil &&
		(previousEmail == nil || !strings.EqualFold(*previousEmail, *contact.WorkEmail)) {
		inUse, err := s.companyContactRepo.EmailInUse(
			ctx,
			contact.CompanyDomain,
			*contact.WorkEmail,
			contact.ID,
		)
		if err != nil {
			return err
		}

		if inUse {
			utils.LogInfo(ctx, "reply address belongs to another contact; contact email unchanged",
				"contact_id", contact.ID)

			return nil
		}
	}

	contact.UpdateTimestamp()

	if err := contact.Validate(); err != nil {
		utils.LogWarn(ctx, "contact update skipped", "contact_id", contact.ID, "error", err.Error())

		return nil
	}

	_, err = s.companyContactRepo.UpdateExisting(ctx, contact)

	return err
}
