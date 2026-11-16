package adminManagement

import (
	"context"
	"errors"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// RetryDeadTask moves a dead task back to pending so the worker runs it again.
func (s *AdminService) RetryDeadTask(
	ctx context.Context,
	queue, taskID string,
) *config.CustomError {
	if customErr := requireKnownQueue(queue); customErr != nil {
		return customErr
	}

	err := s.workerInspector.RetryDeadTask(ctx, queue, taskID)
	if errors.Is(err, config.ErrDeadTaskNotFound) {
		return config.NewCustomError(
			fmt.Errorf("retry dead task: %w", err),
			config.GetErrCodeNotFound(),
		)
	}

	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("retry dead task: %w", err),
			config.GetErrCodeServiceUnavailable(),
		)
	}

	utils.LogInfo(ctx, "dead task retried", "queue", queue, "task_id", taskID)

	return nil
}
