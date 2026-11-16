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

	return config.RequireFound("count candidate metrics", counts, err, config.ErrCandidateNotFound)
}

func candidateNotFound() *config.CustomError {
	return config.NewCustomError(
		fmt.Errorf("dashboard: %w", config.ErrCandidateNotFound),
		config.GetErrCodeNotFound(),
	)
}
