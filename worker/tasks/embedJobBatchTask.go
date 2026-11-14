package workerTasks

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/jobManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

type EmbedJobBatchTaskHandler struct {
	jobService *jobService.JobService
}

func NewEmbedJobBatchTaskHandler(
	service *jobService.JobService,
) (*EmbedJobBatchTaskHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("new embed job batch task handler: %w", config.ErrMissingDependency)
	}

	return &EmbedJobBatchTaskHandler{jobService: service}, nil
}

func (h *EmbedJobBatchTaskHandler) HandleWorkerMessage(
	ctx context.Context,
	message workerConfigs.Message,
) *workerConfigs.WorkerMessageResult {
	payload := workerCustomTypes.EmbedJobBatchPayload{}
	if failure := decodePayload(message, &payload); failure != nil {
		return failure
	}

	_, customErr := h.jobService.EmbedJobs(ctx, payload.JobIDs)

	return resultFor(customErr)
}
