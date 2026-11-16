package dashboardManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	dashboardTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/dashboardManagement"
)

func (s *DashboardService) countCandidate(
	ctx context.Context,
	candidateID string,
) (*dashboardTypes.CandidateCounts, *config.CustomError) {
	if candidateID == "" {
		return nil, candidateNotFound()
	}

	counts, err := s.metricsRepo.CountForCandidate(ctx, candidateID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("count candidate metrics: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if counts == nil {
		return nil, candidateNotFound()
	}

	return counts, nil
}

func candidateNotFound() *config.CustomError {
	return config.NewCustomError(
		fmt.Errorf("dashboard: %w", config.ErrCandidateNotFound),
		config.GetErrCodeNotFound(),
	)
}
