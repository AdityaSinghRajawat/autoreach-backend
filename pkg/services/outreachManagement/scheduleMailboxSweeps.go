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

	tasks := make([]utils.ConcurrentTask, 0, len(accountIDs))
	for _, item := range accountIDs {
		tasks = append(tasks, utils.ConcurrentTask{
			Name: "publish mailbox sweep",
			Run: func(ctx context.Context) error {
				_, err := s.workerClient.Publish(ctx, workerConfigs.PublishParams{
					TaskID:  workerConfigs.TaskSweepMailbox,
					Payload: workerCustomTypes.SweepMailboxPayload{AccountID: item},
				})
				if err != nil {
					utils.LogError(ctx, err, "channel_account_id", item)
				}

				return nil
			},
		})
	}

	_ = utils.RunConcurrently(ctx, tasks...)

	return nil
}
