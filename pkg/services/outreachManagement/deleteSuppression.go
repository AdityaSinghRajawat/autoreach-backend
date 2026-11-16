package outreachManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

func (s *OutreachService) DeleteSuppression(
	ctx context.Context,
	entryID string,
) *config.CustomError {
	deleted, err := s.suppressionRepo.Delete(ctx, entryID)

	return config.RequireAffected("delete suppression", deleted, err, config.ErrSuppressionNotFound)
}
