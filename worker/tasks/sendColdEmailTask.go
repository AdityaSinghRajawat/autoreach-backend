package workerTasks

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/outreachManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

type SendColdEmailTaskHandler struct {
	outreachService *outreachService.OutreachService
}

func NewSendColdEmailTaskHandler(
	service *outreachService.OutreachService,
) (*SendColdEmailTaskHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("new send cold email task handler: %w", config.ErrMissingDependency)
	}

	return &SendColdEmailTaskHandler{outreachService: service}, nil
}

func (h *SendColdEmailTaskHandler) HandleWorkerMessage(
	ctx context.Context,
	message workerConfigs.Message,
) *workerConfigs.WorkerMessageResult {
	payload := workerCustomTypes.SendColdEmailPayload{}
	if failure := decodePayload(message, &payload); failure != nil {
		return failure
	}

	return resultFor(h.outreachService.SendColdEmail(ctx, payload.MessageID))
}
