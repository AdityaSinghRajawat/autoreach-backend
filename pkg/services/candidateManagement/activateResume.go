package candidateManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	candidateTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/candidateManagement"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *CandidateService) ActivateResume(
	ctx context.Context,
	candidateID, resumeID string,
	_ candidateTypes.UpdateResumeRequest,
) (*candidateModel.Resume, *config.CustomError) {
	resume, customErr := s.GetResume(ctx, candidateID, resumeID)
	if customErr != nil {
		return nil, customErr
	}

	if !resume.IsEmbedded() {
		return nil, config.NewCustomError(
			fmt.Errorf("activate resume: %w", config.ErrResumeNotReady),
			config.GetErrCodeConflict(),
		)
	}

	resume.UpdateTimestamp()

	var activated bool

	err := s.transactionManager.WithTransaction(ctx, func(ctx context.Context) error {
		var err error

		activated, err = s.resumeRepo.Activate(ctx, candidateID, resumeID, resume.UpdatedAt)

		return err
	})
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("activate resume: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if !activated {
		return nil, config.NewCustomError(
			fmt.Errorf("activate resume: %w", config.ErrResumeNotReady),
			config.GetErrCodeConflict(),
		)
	}

	resume.IsActive = true

	if err := s.publishMatchResume(ctx, resume.ID); err != nil {
		utils.LogWarn(
			ctx,
			"match resume task publish failed",
			"resume_id", resume.ID,
			"error", err.Error(),
		)
	}

	return resume, nil
}
