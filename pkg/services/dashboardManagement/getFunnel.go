package dashboardManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	dashboardTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/dashboardManagement"
)

func (s *DashboardService) GetFunnel(
	ctx context.Context,
	candidateID string,
) (*dashboardTypes.FunnelResponse, *config.CustomError) {
	counts, customErr := s.countCandidate(ctx, candidateID)
	if customErr != nil {
		return nil, customErr
	}

	funnel := counts.ToFunnelResponse(candidateID)

	return &funnel, nil
}
