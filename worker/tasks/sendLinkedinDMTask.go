package workerTasks

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/outreachManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

type SendLinkedinDMTaskHandler struct {
	outreachService *outreachService.OutreachService
}

func NewSendLinkedinDMTaskHandler(
	service *outreachService.OutreachService,
) (*SendLinkedinDMTaskHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("new send linkedin dm task handler: %w", config.ErrMissingDependency)
	}

	return &SendLinkedinDMTaskHandler{outreachService: service}, nil
}

func (h *SendLinkedinDMTaskHandler) HandleWorkerMessage(
	ctx context.Context,
	message workerConfigs.Message,
) *workerConfigs.WorkerMessageResult {
	payload := workerCustomTypes.SendLinkedinDMPayload{}
	if failure := decodePayload(message, &payload); failure != nil {
		return failure
	}

	return resultFor(h.outreachService.SendLinkedinDM(ctx, payload.MessageID))
}
