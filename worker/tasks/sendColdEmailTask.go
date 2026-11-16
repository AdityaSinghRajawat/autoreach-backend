package workerTasks

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/outreachManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

func SendColdEmail(service *outreachService.OutreachService) workerConfigs.WorkerMessageHandler {
	return handle(
		func(ctx context.Context, payload *workerCustomTypes.SendColdEmailPayload) *config.CustomError {
			return service.SendColdEmail(ctx, payload.MessageID)
		},
	)
}
