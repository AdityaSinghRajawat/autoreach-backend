package outreachManagement

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// EraseContact deletes a contact on their request, with their job links, messages and events, and
// adds their email address to the global suppression list so enrichment cannot bring them back.
func (s *OutreachService) EraseContact(ctx context.Context, contactID string) *config.CustomError {
	err := s.transactionManager.WithTransaction(ctx, func(ctx context.Context) error {
		contact, err := s.companyContactRepo.FindByID(ctx, contactID)
		if err != nil {
			return err
		}

		if contact == nil {
			return config.ErrContactNotFound
		}

		if _, err := s.companyContactRepo.Delete(ctx, contact.ID); err != nil {
			return err
		}

		if contact.WorkEmail == nil {
			return nil
		}

		entry := outreachModel.NewSuppressionEntry()
		entry.EntryType = modelHelpers.SuppressionTypeEmail
		entry.Value = strings.ToLower(*contact.WorkEmail)
		entry.Reason = modelHelpers.SuppressionReasonErased

		_, err = s.suppressionRepo.InsertIfAbsent(ctx, entry)

		return err
	})
	if errors.Is(err, config.ErrContactNotFound) {
		return config.NewCustomError(
			fmt.Errorf("erase contact: %w", err),
			config.GetErrCodeNotFound(),
		)
	}

	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("erase contact: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	utils.LogInfo(ctx, "contact erased on request", "contact_id", contactID)

	return nil
}
