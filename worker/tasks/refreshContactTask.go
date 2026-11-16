package workerTasks

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/enrichmentManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

type RefreshContactTaskHandler struct {
	enrichmentService *enrichmentService.EnrichmentService
}

func NewRefreshContactTaskHandler(
	service *enrichmentService.EnrichmentService,
) (*RefreshContactTaskHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("new refresh contact task handler: %w", config.ErrMissingDependency)
	}

	return &RefreshContactTaskHandler{enrichmentService: service}, nil
}

func (h *RefreshContactTaskHandler) HandleWorkerMessage(
	ctx context.Context,
	message workerConfigs.Message,
) *workerConfigs.WorkerMessageResult {
	payload := workerCustomTypes.RefreshContactPayload{}
	if failure := decodePayload(message, &payload); failure != nil {
		return failure
	}

	return resultFor(h.enrichmentService.RefreshContact(ctx, payload.ContactID))
}
