package adminManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	adminTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/adminManagement"
)

func (s *AdminService) ListQueues(
	ctx context.Context,
) ([]adminTypes.QueueSummary, *config.CustomError) {
	queues, err := s.workerInspector.ListQueues(ctx)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("list queues: %w", err),
			config.GetErrCodeServiceUnavailable(),
		)
	}

	summaries := make([]adminTypes.QueueSummary, 0, len(queues))
	for _, queue := range queues {
		summaries = append(summaries, adminTypes.NewQueueSummary(queue))
	}

	return summaries, nil
}
