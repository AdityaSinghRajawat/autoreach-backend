package workerTasks

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	applicationService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/applicationManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

type QueueApplicationsTaskHandler struct {
	applicationService *applicationService.ApplicationService
}

func NewQueueApplicationsTaskHandler(
	service *applicationService.ApplicationService,
) (*QueueApplicationsTaskHandler, error) {
	if service == nil {
		return nil, fmt.Errorf(
			"new queue applications task handler: %w",
			config.ErrMissingDependency,
		)
	}

	return &QueueApplicationsTaskHandler{applicationService: service}, nil
}

func (h *QueueApplicationsTaskHandler) HandleWorkerMessage(
	ctx context.Context,
	message workerConfigs.Message,
) *workerConfigs.WorkerMessageResult {
	payload := workerCustomTypes.QueueApplicationsPayload{}
	if failure := decodePayload(message, &payload); failure != nil {
		return failure
	}

	_, customErr := h.applicationService.QueueApplications(ctx, payload.CandidateID)

	return resultFor(customErr)
}
