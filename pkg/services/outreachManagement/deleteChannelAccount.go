package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

func (s *OutreachService) DeleteChannelAccount(
	ctx context.Context,
	candidateID, accountID string,
) *config.CustomError {
	deleted, err := s.channelAccountRepo.Delete(ctx, candidateID, accountID)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("delete channel account: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if !deleted {
		return config.NewCustomError(
			fmt.Errorf("delete channel account: %w", config.ErrChannelAccountNotFound),
			config.GetErrCodeNotFound(),
		)
	}

	return nil
}
