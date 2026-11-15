package workerTasks

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/outreachManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

type SendLinkedinInviteTaskHandler struct {
	outreachService *outreachService.OutreachService
}

func NewSendLinkedinInviteTaskHandler(
	service *outreachService.OutreachService,
) (*SendLinkedinInviteTaskHandler, error) {
	if service == nil {
		return nil, fmt.Errorf(
			"new send linkedin invite task handler: %w",
			config.ErrMissingDependency,
		)
	}

	return &SendLinkedinInviteTaskHandler{outreachService: service}, nil
}

func (h *SendLinkedinInviteTaskHandler) HandleWorkerMessage(
	ctx context.Context,
	message workerConfigs.Message,
) *workerConfigs.WorkerMessageResult {
	payload := workerCustomTypes.SendLinkedinInvitePayload{}
	if failure := decodePayload(message, &payload); failure != nil {
		return failure
	}

	return resultFor(h.outreachService.SendLinkedinInvite(ctx, payload.MessageID))
}
