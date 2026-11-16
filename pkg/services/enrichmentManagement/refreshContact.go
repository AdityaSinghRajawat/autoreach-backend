package enrichmentManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/enrichmentManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// RefreshContact asks the contact providers for a person's current email after theirs bounced. An
// address that bounced before, or one another contact at the company holds, is not adopted.
func (s *EnrichmentService) RefreshContact(
	ctx context.Context,
	contactID string,
) *config.CustomError {
	contact, err := s.companyContactRepo.FindByID(ctx, contactID)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("refresh contact: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if contact == nil || contact.AnonymizedAt != nil {
		return nil
	}

	details, _ := s.findContactDetails(ctx, enrichmentTypes.PersonFromContact(contact))
	if details == nil || details.Email == "" {
		utils.LogInfo(ctx, "no current email found for contact", "contact_id", contact.ID)

		return nil
	}

	address := utils.NormaliseEmail(details.Email)
	if contact.HasBounced(address) || (contact.WorkEmail != nil && *contact.WorkEmail == address) {
		return nil
	}

	inUse, err := s.companyContactRepo.EmailInUse(ctx, contact.CompanyDomain, address, contact.ID)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("refresh contact: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if inUse {
		return nil
	}

	status := s.emailVerifier.Verify(ctx, address, details.EmailVerified)
	if status == modelHelpers.EmailStatusInvalid {
		return nil
	}

	contact.ReplaceEmail(
		address,
		status,
		modelHelpers.ContactChangeRefreshed,
		utils.GetCurrTimeStamp(),
	)

	if details.Phone != "" && contact.PhoneNumber == nil {
		contact.PhoneNumber = &details.Phone
	}

	contact.UpdateTimestamp()

	if err := contact.Validate(); err != nil {
		utils.LogWarn(
			ctx,
			"refreshed contact rejected",
			"contact_id",
			contact.ID,
			"error",
			err.Error(),
		)

		return nil
	}

	if _, err := s.companyContactRepo.Upsert(ctx, contact); err != nil {
		return config.NewCustomError(
			fmt.Errorf("refresh contact: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	utils.LogInfo(ctx, "contact email refreshed", "contact_id", contact.ID)

	return nil
}
