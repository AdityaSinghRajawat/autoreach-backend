package outreachManagement

import (
	"context"
	"errors"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

func (s *OutreachService) saveMessage(
	ctx context.Context,
	message *outreachModel.Message,
	operation string,
) (*outreachModel.Message, *config.CustomError) {
	message.UpdateTimestamp()

	if err := message.Validate(); err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("%s: %w", operation, err),
			config.GetErrCodeBadRequest(),
		)
	}

	saved, err := s.messageRepo.Upsert(ctx, message)
	if errors.Is(err, config.ErrDuplicateRecord) {
		return nil, config.NewCustomError(
			fmt.Errorf("%s: %w", operation, err),
			config.GetErrCodeConflict(),
		)
	}

	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("%s: %w", operation, err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return saved, nil
}
