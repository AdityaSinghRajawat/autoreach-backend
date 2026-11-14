package workerTasks

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	matchingService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/matchingManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

type MatchJobTaskHandler struct {
	matchingService *matchingService.MatchingService
}

func NewMatchJobTaskHandler(
	service *matchingService.MatchingService,
) (*MatchJobTaskHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("new match job task handler: %w", config.ErrMissingDependency)
	}

	return &MatchJobTaskHandler{matchingService: service}, nil
}

func (h *MatchJobTaskHandler) HandleWorkerMessage(
	ctx context.Context,
	message workerConfigs.Message,
) *workerConfigs.WorkerMessageResult {
	payload := workerCustomTypes.MatchJobPayload{}
	if failure := decodePayload(message, &payload); failure != nil {
		return failure
	}

	_, customErr := h.matchingService.MatchJob(ctx, payload.JobID)

	return resultFor(customErr)
}
