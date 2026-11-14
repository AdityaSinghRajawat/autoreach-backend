package workerTasks

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/jobManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

type EmbedJobTaskHandler struct {
	jobService *jobService.JobService
}

func NewEmbedJobTaskHandler(service *jobService.JobService) (*EmbedJobTaskHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("new embed job task handler: %w", config.ErrMissingDependency)
	}

	return &EmbedJobTaskHandler{jobService: service}, nil
}

func (h *EmbedJobTaskHandler) HandleWorkerMessage(
	ctx context.Context,
	message workerConfigs.Message,
) *workerConfigs.WorkerMessageResult {
	payload := workerCustomTypes.EmbedJobPayload{}
	if failure := decodePayload(message, &payload); failure != nil {
		return failure
	}

	_, customErr := h.jobService.EmbedJobs(ctx, []string{payload.JobID})

	return resultFor(customErr)
}
