package workerTasks

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	applicationService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/applicationManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

func PrepareApplication(
	service *applicationService.ApplicationService,
) workerConfigs.WorkerMessageHandler {
	return handle(
		func(ctx context.Context, payload *workerCustomTypes.PrepareApplicationPayload) *config.CustomError {
			return service.PrepareApplication(ctx, payload.ApplicationID)
		},
	)
}
