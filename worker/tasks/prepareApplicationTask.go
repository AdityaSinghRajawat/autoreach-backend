package workerTasks

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	applicationService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/applicationManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

type PrepareApplicationTaskHandler struct {
	applicationService *applicationService.ApplicationService
}

func NewPrepareApplicationTaskHandler(
	service *applicationService.ApplicationService,
) (*PrepareApplicationTaskHandler, error) {
	if service == nil {
		return nil, fmt.Errorf(
			"new prepare application task handler: %w",
			config.ErrMissingDependency,
		)
	}

	return &PrepareApplicationTaskHandler{applicationService: service}, nil
}

func (h *PrepareApplicationTaskHandler) HandleWorkerMessage(
	ctx context.Context,
	message workerConfigs.Message,
) *workerConfigs.WorkerMessageResult {
	payload := workerCustomTypes.PrepareApplicationPayload{}
	if failure := decodePayload(message, &payload); failure != nil {
		return failure
	}

	return resultFor(h.applicationService.PrepareApplication(ctx, payload.ApplicationID))
}
