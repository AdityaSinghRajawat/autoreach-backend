package outreachManagement

import (
	"context"

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

	now := utils.GetCurrTimeStamp()

	message.Close(modelHelpers.OutreachStatusBounced, config.GetBounceReason())
	message.UpdateTimestamp()

	return s.closeConversation(ctx, message, modelHelpers.OutreachEventBounced,
		func(ctx context.Context) error {
			if err := s.messageRepo.CancelForContact(
				ctx,
				message.CandidateID,
				message.ContactID,
				now,
			); err != nil {
				return err
			}

			return s.suppressContact(
				ctx,
				nil,
				message.ContactID,
				modelHelpers.SuppressionReasonBounced,
			)
		})
}
