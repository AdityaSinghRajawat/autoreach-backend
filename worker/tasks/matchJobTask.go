package workerTasks

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	matchingService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/matchingManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

func MatchJob(service *matchingService.MatchingService) workerConfigs.WorkerMessageHandler {
	return handle(
		func(ctx context.Context, payload *workerCustomTypes.MatchJobPayload) *config.CustomError {
			_, customErr := service.MatchJob(ctx, payload.JobID)

			return customErr
		},
	)
}
