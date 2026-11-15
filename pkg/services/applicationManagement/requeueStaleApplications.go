package applicationManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *ApplicationService) RequeueStaleApplications(
	ctx context.Context,
) (int, *config.CustomError) {
	applicationIDs, err := s.applicationRepo.ListStalePendingIDs(
		ctx,
		utils.GetCurrTimeStamp().Add(-config.GetReconcilerStaleAfter()),
		config.GetReconcilerBatchSize(),
	)
	if err != nil {
		return 0, config.NewCustomError(
			fmt.Errorf("requeue stale applications: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	for _, applicationID := range applicationIDs {
		s.publishPrepare(ctx, applicationID)
	}

	return len(applicationIDs), nil
}
