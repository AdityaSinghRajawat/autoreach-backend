package workerTasks

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	matchingService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/matchingManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

func MatchResume(service *matchingService.MatchingService) workerConfigs.WorkerMessageHandler {
	return handle(
		func(ctx context.Context, payload *workerCustomTypes.MatchResumePayload) *config.CustomError {
			_, customErr := service.MatchResume(ctx, payload.ResumeID)

			return customErr
		},
	)
}
