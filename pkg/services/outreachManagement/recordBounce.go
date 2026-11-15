package outreachManagement

import (
	"context"
	"fmt"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *OutreachService) recordBounce(
	ctx context.Context,
	message *outreachModel.Message,
) *config.CustomError {
	if message.Status == modelHelpers.OutreachStatusBounced {
		return nil
	}

	contact, err := s.companyContactRepo.FindByID(ctx, message.ContactID)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("record bounce: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	now := utils.GetCurrTimeStamp()

	message.Close(modelHelpers.OutreachStatusBounced, config.GetBounceReason())
	message.UpdateTimestamp()

	return s.closeConversation(
		ctx,
		message,
		modelHelpers.OutreachEventBounced,
		func(ctx context.Context) error {
			if err := s.messageRepo.CancelForContact(
				ctx,
				message.CandidateID,
				message.ContactID,
				now,
			); err != nil {
				return err
			}

			if contact == nil || contact.WorkEmail == nil {
				return nil
			}

			entry := outreachModel.NewSuppressionEntry()
			entry.EntryType = modelHelpers.SuppressionTypeEmail
			entry.Value = strings.ToLower(*contact.WorkEmail)
			entry.Reason = modelHelpers.SuppressionReasonBounced

			if err := entry.Validate(); err != nil {
				return err
			}

			return s.suppressionRepo.Insert(ctx, entry)
		},
	)
}
