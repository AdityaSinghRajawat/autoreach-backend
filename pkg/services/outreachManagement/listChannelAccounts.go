package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

func (s *OutreachService) ListChannelAccounts(
	ctx context.Context,
	candidateID string,
) ([]*outreachModel.ChannelAccount, *config.CustomError) {
	accounts, err := s.channelAccountRepo.ListByCandidate(ctx, candidateID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("list channel accounts: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return accounts, nil
}
