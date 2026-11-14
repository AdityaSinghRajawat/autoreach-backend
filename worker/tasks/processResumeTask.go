package workerTasks

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	candidateService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/candidateManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

type ProcessResumeTaskHandler struct {
	candidateService *candidateService.CandidateService
}

func NewProcessResumeTaskHandler(
	service *candidateService.CandidateService,
) (*ProcessResumeTaskHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("new process resume task handler: %w", config.ErrMissingDependency)
	}

	return &ProcessResumeTaskHandler{candidateService: service}, nil
}

func (h *ProcessResumeTaskHandler) HandleWorkerMessage(
	ctx context.Context,
	message workerConfigs.Message,
) *workerConfigs.WorkerMessageResult {
	payload := workerCustomTypes.ProcessResumePayload{}
	if failure := decodePayload(message, &payload); failure != nil {
		return failure
	}

	_, customErr := h.candidateService.ProcessResume(ctx, payload.ResumeID)

	return resultFor(customErr)
}
