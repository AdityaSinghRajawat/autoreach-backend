package candidateManagement

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *CandidateService) RequeuePendingResumes(ctx context.Context) (int, *config.CustomError) {
	staleBefore := utils.GetCurrTimeStamp().Add(-config.GetReconcilerStaleAfter())

	resumeIDs, err := s.resumeRepo.ListStalePendingIDs(
		ctx,
		staleBefore,
		config.GetReconcilerBatchSize(),
	)
	if err != nil {
		return 0, config.NewCustomError(
			fmt.Errorf("requeue pending resumes: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	var requeued atomic.Int64

	tasks := make([]utils.ConcurrentTask, 0, len(resumeIDs))
	for _, resumeID := range resumeIDs {
		tasks = append(tasks, utils.ConcurrentTask{
			Name: "publish process resume task",
			Run: func(ctx context.Context) error {
				if err := s.publishProcessResume(ctx, resumeID); err != nil {
					return err
				}

				requeued.Add(1)

				return nil
			},
		})
	}

	if err := utils.RunConcurrently(ctx, tasks...); err != nil {
		return int(requeued.Load()), config.NewCustomError(
			fmt.Errorf("requeue pending resumes: %w", err),
			config.GetErrCodeServiceUnavailable(),
		)
	}

	return int(requeued.Load()), nil
}
