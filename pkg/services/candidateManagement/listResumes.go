package candidateManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
)

func (s *CandidateService) ListResumes(
	ctx context.Context,
	candidateID string,
) ([]*candidateModel.Resume, *config.CustomError) {
	resumes, err := s.resumeRepo.ListByCandidate(ctx, candidateID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("list resumes: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return resumes, nil
}
