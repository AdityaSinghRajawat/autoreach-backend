package candidateManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
)

func (s *CandidateService) GetCandidate(
	ctx context.Context,
	candidateID string,
) (*candidateModel.Candidate, *config.CustomError) {
	candidate, err := s.candidateRepo.FindByID(ctx, candidateID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("get candidate: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if candidate == nil {
		return nil, config.NewCustomError(
			fmt.Errorf("get candidate: %w", config.ErrCandidateNotFound),
			config.GetErrCodeNotFound(),
		)
	}

	return candidate, nil
}
