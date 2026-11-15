package workerTasks

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/outreachManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

type PlanOutreachTaskHandler struct {
	outreachService *outreachService.OutreachService
}

func NewPlanOutreachTaskHandler(
	service *outreachService.OutreachService,
) (*PlanOutreachTaskHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("new plan outreach task handler: %w", config.ErrMissingDependency)
	}

	return &PlanOutreachTaskHandler{outreachService: service}, nil
}

func (h *PlanOutreachTaskHandler) HandleWorkerMessage(
	ctx context.Context,
	message workerConfigs.Message,
) *workerConfigs.WorkerMessageResult {
	payload := workerCustomTypes.PlanOutreachPayload{}
	if failure := decodePayload(message, &payload); failure != nil {
		return failure
	}

	_, customErr := h.outreachService.PlanOutreach(ctx, payload.CandidateID, payload.JobID)

	return resultFor(customErr)
}
