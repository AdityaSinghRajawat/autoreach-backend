package workerTasks

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/outreachManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

type SweepMailboxTaskHandler struct {
	outreachService *outreachService.OutreachService
}

func NewSweepMailboxTaskHandler(
	service *outreachService.OutreachService,
) (*SweepMailboxTaskHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("new sweep mailbox task handler: %w", config.ErrMissingDependency)
	}

	return &SweepMailboxTaskHandler{outreachService: service}, nil
}

func (h *SweepMailboxTaskHandler) HandleWorkerMessage(
	ctx context.Context,
	message workerConfigs.Message,
) *workerConfigs.WorkerMessageResult {
	payload := workerCustomTypes.SweepMailboxPayload{}
	if failure := decodePayload(message, &payload); failure != nil {
		return failure
	}

	return resultFor(h.outreachService.SweepMailbox(ctx, payload.AccountID))
}
