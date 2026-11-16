package enrichmentManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

func (s *EnrichmentService) requireMatchedJob(
	ctx context.Context,
	candidateID, jobID string,
) *config.CustomError {
	matched, err := s.jobMatchRepo.Exists(ctx, candidateID, jobID)

	return config.RequireAffected("require matched job", matched, err, config.ErrJobNotMatched)
}
