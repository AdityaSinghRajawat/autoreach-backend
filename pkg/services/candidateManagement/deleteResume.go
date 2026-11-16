package candidateManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

func (s *CandidateService) DeleteResume(
	ctx context.Context,
	candidateID, resumeID string,
) *config.CustomError {
	resume, customErr := s.GetResume(ctx, candidateID, resumeID)
	if customErr != nil {
		return customErr
	}

	deleted, err := s.resumeRepo.Delete(ctx, candidateID, resumeID)
	if customErr := config.RequireAffected(
		"delete resume",
		deleted,
		err,
		config.ErrResumeNotFound,
	); customErr != nil {
		return customErr
	}

	s.deleteStoredFile(ctx, resume.StorageKey)

	return nil
}
