package workerTasks

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	applicationService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/applicationManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

func QueueApplications(
	service *applicationService.ApplicationService,
) workerConfigs.WorkerMessageHandler {
	return handle(
		func(ctx context.Context, payload *workerCustomTypes.QueueApplicationsPayload) *config.CustomError {
			_, customErr := service.QueueApplications(ctx, payload.CandidateID)

			return customErr
		},
	)
}
