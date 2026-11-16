package outreachManagement

import (
	"context"
	"fmt"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// AnonymizeStaleContacts removes the personal details of contacts unused for the retention period
// while keeping their outreach history for metrics; a zero period keeps everything.
func (s *OutreachService) AnonymizeStaleContacts(
	ctx context.Context,
	retention time.Duration,
) (int, *config.CustomError) {
	if retention <= 0 {
		return 0, nil
	}

	now := utils.GetCurrTimeStamp()

	anonymized, err := s.companyContactRepo.AnonymizeStale(ctx, now.Add(-retention), now)
	if err != nil {
		return 0, config.NewCustomError(
			fmt.Errorf("anonymize stale contacts: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return anonymized, nil
}
