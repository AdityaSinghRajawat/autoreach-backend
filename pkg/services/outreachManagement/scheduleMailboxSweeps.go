package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

func (s *OutreachService) ScheduleMailboxSweeps(ctx context.Context) *config.CustomError {
	accountIDs, err := s.channelAccountRepo.ListActiveIDs(ctx)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("schedule mailbox sweeps: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	for _, accountID := range accountIDs {
		_, err := s.workerClient.Publish(ctx, workerConfigs.PublishParams{
			TaskID:  workerConfigs.TaskSweepMailbox,
			Payload: workerCustomTypes.SweepMailboxPayload{AccountID: accountID},
		})
		if err != nil {
			utils.LogError(ctx, err, "channel_account_id", accountID)
		}
	}

	return nil
}
