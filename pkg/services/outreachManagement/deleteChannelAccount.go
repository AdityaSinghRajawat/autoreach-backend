package outreachManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

func (s *OutreachService) DeleteChannelAccount(
	ctx context.Context,
	candidateID, accountID string,
) *config.CustomError {
	deleted, err := s.channelAccountRepo.Delete(ctx, candidateID, accountID)

	return config.RequireAffected(
		"delete channel account",
		deleted,
		err,
		config.ErrChannelAccountNotFound,
	)
}
