package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

func (s *OutreachService) GetMessage(
	ctx context.Context,
	candidateID, messageID string,
) (*outreachModel.Message, *config.CustomError) {
	message, err := s.messageRepo.FindByCandidate(ctx, candidateID, messageID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("get outreach message: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if message == nil {
		return nil, config.NewCustomError(
			fmt.Errorf("get outreach message: %w", config.ErrMessageNotFound),
			config.GetErrCodeNotFound(),
		)
	}

	return message, nil
}
