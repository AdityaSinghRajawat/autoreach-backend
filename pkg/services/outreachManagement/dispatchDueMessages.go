package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
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

	messageIDs, err := s.messageRepo.ClaimDue(ctx, now, config.GetOutreachDispatchBatchSize())
	if err != nil {
		return 0, config.NewCustomError(
			fmt.Errorf("dispatch messages: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	dispatched := 0

	for _, messageID := range messageIDs {
		_, err := s.workerClient.Publish(ctx, workerConfigs.PublishParams{
			TaskID:  workerConfigs.TaskSendColdEmail,
			Payload: workerCustomTypes.SendColdEmailPayload{MessageID: messageID},
		})
		if err == nil {
			dispatched++

			continue
		}

		utils.LogWarn(ctx, "send task publish failed, releasing the claim", "message_id", messageID)

		if err := s.messageRepo.ReleaseClaim(ctx, messageID, now); err != nil {
			utils.LogError(ctx, err, "message_id", messageID)
		}
	}

	return dispatched, nil
}
