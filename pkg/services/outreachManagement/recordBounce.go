package outreachManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
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

	promoted := false

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

			contact, err := s.companyContactRepo.FindByID(ctx, message.ContactID)
			if err != nil || contact == nil || contact.WorkEmail == nil {
				return err
			}

			if err := s.suppressAddress(
				ctx,
				&message.CandidateID,
				*contact.WorkEmail,
				modelHelpers.SuppressionReasonBounced,
			); err != nil {
				return err
			}

			promoted, err = s.reportContactSignal(ctx, message.CandidateID, message.ContactID,
				modelHelpers.ContactSignalBounced, *contact.WorkEmail, now)

			return err
		})
	if customErr != nil {
		return customErr
	}

	if promoted {
		s.publishContactRefresh(ctx, message.ContactID)
	}

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
