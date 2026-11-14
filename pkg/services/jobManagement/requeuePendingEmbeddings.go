package jobManagement

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *JobService) RequeuePendingEmbeddings(ctx context.Context) (int, *config.CustomError) {
	staleBefore := utils.GetCurrTimeStamp().Add(-config.GetReconcilerStaleAfter())

	jobIDs, err := s.jobPostingRepo.ListStalePendingEmbeddingIDs(
		ctx,
		staleBefore,
		config.GetReconcilerBatchSize(),
	)
	if err != nil {
		return 0, config.NewCustomError(
			fmt.Errorf("requeue pending embeddings: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	var requeued atomic.Int64

	tasks := make([]utils.ConcurrentTask, 0, len(jobIDs))
	for _, jobID := range jobIDs {
		tasks = append(tasks, utils.ConcurrentTask{
			Name: "publish embed task",
			Run: func(ctx context.Context) error {
				if err := s.publishEmbedTask(ctx, jobID); err != nil {
					return err
				}

				requeued.Add(1)

				return nil
			},
		})
	}

	if err := utils.RunConcurrently(ctx, tasks...); err != nil {
		return int(requeued.Load()), config.NewCustomError(
			fmt.Errorf("requeue pending embeddings: %w", err),
			config.GetErrCodeServiceUnavailable(),
		)
	}

	return int(requeued.Load()), nil
}
