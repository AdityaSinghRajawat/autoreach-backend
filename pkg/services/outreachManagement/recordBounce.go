package outreachManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
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

	customErr := s.closeConversation(ctx, message, modelHelpers.OutreachEventBounced,
		func(ctx context.Context) error {
			if err := s.messageRepo.CancelForContact(
				ctx,
				message.CandidateID,
				message.ContactID,
				now,
			); err != nil {
				return err
			}

			if err := s.suppressContact(
				ctx,
				nil,
				message.ContactID,
				modelHelpers.SuppressionReasonBounced,
			); err != nil {
				return err
			}

			return s.updateContact(
				ctx,
				message.ContactID,
				func(contact *enrichmentModel.CompanyContact) {
					contact.RecordBounce(now)
				},
			)
		})
	if customErr != nil {
		return customErr
	}

	s.publishContactRefresh(ctx, message.ContactID)

	return nil
}

// The person may have moved on; a refresh looks for their current address.
func (s *OutreachService) publishContactRefresh(ctx context.Context, contactID string) {
	_, err := s.workerClient.Publish(ctx, workerConfigs.PublishParams{
		TaskID:  workerConfigs.TaskRefreshContact,
		Payload: workerCustomTypes.RefreshContactPayload{ContactID: contactID},
	})
	if err != nil {
		utils.LogWarn(
			ctx,
			"contact refresh publish failed",
			"contact_id",
			contactID,
			"error",
			err.Error(),
		)
	}
}
