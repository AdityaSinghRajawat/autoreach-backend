package outreachManagement

import (
	"context"
	"fmt"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *OutreachService) PurgeOutreachEvents(
	ctx context.Context,
	retention time.Duration,
) (int, *config.CustomError) {
	if retention <= 0 {
		return 0, nil
	}

	deleted, err := s.eventRepo.DeleteOccurredBefore(ctx, utils.GetCurrTimeStamp().Add(-retention))
	if err != nil {
		return 0, config.NewCustomError(
			fmt.Errorf("purge outreach events: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return deleted, nil
}
