package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

func (s *OutreachService) recordEvent(
	ctx context.Context,
	messageID string,
	eventType modelHelpers.OutreachEventType,
) *config.CustomError {
	if err := s.eventRepo.Insert(ctx, outreachModel.NewEvent(messageID, eventType)); err != nil {
		return config.NewCustomError(
			fmt.Errorf("record outreach event: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return nil
}
