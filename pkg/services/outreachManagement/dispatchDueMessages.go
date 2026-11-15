package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

func (s *OutreachService) DispatchDueMessages(ctx context.Context) (int, *config.CustomError) {
	now := utils.GetCurrTimeStamp()

	if _, err := s.messageRepo.FailStaleSending(
		ctx,
		now.Add(-config.GetSendingStaleAfter()),
		now,
	); err != nil {
		return 0, config.NewCustomError(
			fmt.Errorf("dispatch messages: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	claimed, err := s.messageRepo.ClaimDue(ctx, now, config.GetOutreachDispatchBatchSize())
	if err != nil {
		return 0, config.NewCustomError(
			fmt.Errorf("dispatch messages: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	dispatched := 0

	for _, message := range claimed {
		_, err := s.workerClient.Publish(ctx, sendTaskFor(message.ID, message.Channel))
		if err == nil {
			dispatched++

			continue
		}

		utils.LogWarn(
			ctx,
			"send task publish failed, releasing the claim",
			"message_id",
			message.ID,
		)

		if err := s.messageRepo.ReleaseClaim(ctx, message.ID, now); err != nil {
			utils.LogError(ctx, err, "message_id", message.ID)
		}
	}

	return dispatched, nil
}

func sendTaskFor(messageID, channel string) workerConfigs.PublishParams {
	switch modelHelpers.OutreachChannel(channel) {
	case modelHelpers.OutreachChannelLinkedinConnect:
		return workerConfigs.PublishParams{
			TaskID:  workerConfigs.TaskSendLinkedinInvite,
			Payload: workerCustomTypes.SendLinkedinInvitePayload{MessageID: messageID},
		}
	case modelHelpers.OutreachChannelLinkedinDM:
		return workerConfigs.PublishParams{
			TaskID:  workerConfigs.TaskSendLinkedinDM,
			Payload: workerCustomTypes.SendLinkedinDMPayload{MessageID: messageID},
		}
	default:
		return workerConfigs.PublishParams{
			TaskID:  workerConfigs.TaskSendColdEmail,
			Payload: workerCustomTypes.SendColdEmailPayload{MessageID: messageID},
		}
	}
}
