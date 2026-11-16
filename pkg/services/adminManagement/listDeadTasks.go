package adminManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	adminTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/adminManagement"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
)

func (s *AdminService) ListDeadTasks(
	ctx context.Context,
	queue string,
	params sharedTypes.PageParams,
) ([]adminTypes.DeadTask, *sharedTypes.PaginationMeta, *config.CustomError) {
	if customErr := requireKnownQueue(queue); customErr != nil {
		return nil, nil, customErr
	}

	params = params.Normalize()

	tasks, total, err := s.workerInspector.ListDeadTasks(ctx, queue, params.Page, params.PerPage)
	if err != nil {
		return nil, nil, config.NewCustomError(
			fmt.Errorf("list dead tasks: %w", err),
			config.GetErrCodeServiceUnavailable(),
		)
	}

	deadTasks := make([]adminTypes.DeadTask, 0, len(tasks))
	for _, task := range tasks {
		deadTasks = append(deadTasks, adminTypes.NewDeadTask(task))
	}

	meta := sharedTypes.NewPaginationMeta(params, total)

	return deadTasks, &meta, nil
}
