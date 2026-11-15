package applicationManagement

import (
	"context"
	"errors"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	applicationTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/applicationManagement"
	applicationModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/applicationManagement"
)

func (s *ApplicationService) CreateApplication(
	ctx context.Context,
	candidateID string,
	request applicationTypes.CreateApplicationRequest,
) (*applicationModel.Application, *config.CustomError) {
	job, err := s.jobPostingRepo.FindByID(ctx, request.JobID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("create application: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if job == nil {
		return nil, config.NewCustomError(
			fmt.Errorf("create application: %w", config.ErrJobNotFound),
			config.GetErrCodeNotFound(),
		)
	}

	score, err := s.applicationRepo.MatchScore(ctx, candidateID, job.ID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("create application: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	application := request.ToApplicationModel(candidateID, job, score)
	application.UpdateTimestamp()

	if err := application.Validate(); err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("create application: %w", err),
			config.GetErrCodeBadRequest(),
		)
	}

	created, err := s.applicationRepo.Upsert(ctx, application)
	if errors.Is(err, config.ErrDuplicateRecord) {
		return nil, config.NewCustomError(
			fmt.Errorf("create application: %w", err),
			config.GetErrCodeConflict(),
		)
	}

	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("create application: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	s.publishPrepare(ctx, created.ID)

	return created, nil
}
