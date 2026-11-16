package candidateManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
)

func (s *CandidateService) GetResume(
	ctx context.Context,
	candidateID, resumeID string,
) (*candidateModel.Resume, *config.CustomError) {
	resume, err := s.resumeRepo.FindByID(ctx, resumeID)
	if resume != nil && resume.CandidateID != candidateID {
		resume = nil
	}

	return config.RequireFound("get resume", resume, err, config.ErrResumeNotFound)
}
