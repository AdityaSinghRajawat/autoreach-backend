package workerTasks

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	matchingService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/matchingManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

type MatchResumeTaskHandler struct {
	matchingService *matchingService.MatchingService
}

func NewMatchResumeTaskHandler(
	service *matchingService.MatchingService,
) (*MatchResumeTaskHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("new match resume task handler: %w", config.ErrMissingDependency)
	}

	return &MatchResumeTaskHandler{matchingService: service}, nil
}

func (h *MatchResumeTaskHandler) HandleWorkerMessage(
	ctx context.Context,
	message workerConfigs.Message,
) *workerConfigs.WorkerMessageResult {
	payload := workerCustomTypes.MatchResumePayload{}
	if failure := decodePayload(message, &payload); failure != nil {
		return failure
	}

	_, customErr := h.matchingService.MatchResume(ctx, payload.ResumeID)

	return resultFor(customErr)
}
