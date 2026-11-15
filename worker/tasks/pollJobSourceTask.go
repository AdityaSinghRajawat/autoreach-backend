package workerTasks

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/jobManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

type PollJobSourceTaskHandler struct {
	jobService *jobService.JobService
}

func NewPollJobSourceTaskHandler(
	service *jobService.JobService,
) (*PollJobSourceTaskHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("new poll job source task handler: %w", config.ErrMissingDependency)
	}

	return &PollJobSourceTaskHandler{jobService: service}, nil
}

func (h *PollJobSourceTaskHandler) HandleWorkerMessage(
	ctx context.Context,
	message workerConfigs.Message,
) *workerConfigs.WorkerMessageResult {
	payload := workerCustomTypes.PollJobSourcePayload{}
	if failure := decodePayload(message, &payload); failure != nil {
		return failure
	}

	_, customErr := h.jobService.PollJobSource(ctx, payload.SourceID)

	return resultFor(customErr)
}
