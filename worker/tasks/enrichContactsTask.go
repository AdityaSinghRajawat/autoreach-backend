package workerTasks

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/enrichmentManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

type EnrichContactsTaskHandler struct {
	enrichmentService *enrichmentService.EnrichmentService
}

func NewEnrichContactsTaskHandler(
	service *enrichmentService.EnrichmentService,
) (*EnrichContactsTaskHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("new enrich contacts task handler: %w", config.ErrMissingDependency)
	}

	return &EnrichContactsTaskHandler{enrichmentService: service}, nil
}

func (h *EnrichContactsTaskHandler) HandleWorkerMessage(
	ctx context.Context,
	message workerConfigs.Message,
) *workerConfigs.WorkerMessageResult {
	payload := workerCustomTypes.EnrichContactsPayload{}
	if failure := decodePayload(message, &payload); failure != nil {
		return failure
	}

	_, customErr := h.enrichmentService.EnrichJobContacts(ctx, payload.JobID, payload.TargetRoles)

	return resultFor(customErr)
}
