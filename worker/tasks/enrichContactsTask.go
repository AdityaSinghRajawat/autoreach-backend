package workerTasks

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/enrichmentManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

func EnrichContacts(
	service *enrichmentService.EnrichmentService,
) workerConfigs.WorkerMessageHandler {
	return handle(
		func(ctx context.Context, payload *workerCustomTypes.EnrichContactsPayload) *config.CustomError {
			_, customErr := service.EnrichJobContacts(ctx, payload.JobID, payload.TargetRoles)

			return customErr
		},
	)
}
