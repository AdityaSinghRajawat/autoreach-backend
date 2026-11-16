package candidateManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
)

func (s *CandidateService) GetCandidate(
	ctx context.Context,
	candidateID string,
) (*candidateModel.Candidate, *config.CustomError) {
	candidate, err := s.candidateRepo.FindByID(ctx, candidateID)

	return config.RequireFound("get candidate", candidate, err, config.ErrCandidateNotFound)
}
