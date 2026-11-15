package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

func (s *OutreachService) closeConversation(
	ctx context.Context,
	message *outreachModel.Message,
	eventType modelHelpers.OutreachEventType,
	followUp func(ctx context.Context) error,
) *config.CustomError {
	err := s.transactionManager.WithTransaction(ctx, func(ctx context.Context) error {
		if _, err := s.messageRepo.Upsert(ctx, message); err != nil {
			return err
		}

		if err := followUp(ctx); err != nil {
			return err
		}

		return s.eventRepo.Insert(ctx, outreachModel.NewEvent(message.ID, eventType))
	})
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("record %s: %w", eventType, err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return nil
}
