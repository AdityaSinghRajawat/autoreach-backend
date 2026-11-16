package dashboardManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	dashboardTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/dashboardManagement"
)

func (s *DashboardService) GetMetrics(
	ctx context.Context,
	candidateID string,
) (*dashboardTypes.CandidateMetricsResponse, *config.CustomError) {
	counts, customErr := s.countCandidate(ctx, candidateID)
	if customErr != nil {
		return nil, customErr
	}

	metrics := counts.ToMetricsResponse(candidateID)

	return &metrics, nil
}
