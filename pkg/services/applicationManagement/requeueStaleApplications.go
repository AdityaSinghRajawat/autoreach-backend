package applicationManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *ApplicationService) RequeueStaleApplications(
	ctx context.Context,
) (int, *config.CustomError) {
	applicationIDs, err := s.applicationRepo.ListStalePendingIDs(
		ctx,
		utils.GetCurrTimeStamp().Add(-config.GetReconcilerStaleAfter()),
		config.GetReconcilerBatchSize(),
	)
	if err != nil {
		return 0, config.NewCustomError(
			fmt.Errorf("requeue stale applications: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	tasks := make([]utils.ConcurrentTask, 0, len(applicationIDs))
	for _, item := range applicationIDs {
		tasks = append(tasks, utils.ConcurrentTask{
			Name: "publish prepare application",
			Run: func(ctx context.Context) error {
				s.publishPrepare(ctx, item)

				return nil
			},
		})
	}

	_ = utils.RunConcurrently(ctx, tasks...)

	return len(applicationIDs), nil
}
