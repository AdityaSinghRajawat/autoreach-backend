package jobManagement

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
)

func (s *JobService) GetJob(
	ctx context.Context,
	jobID string,
) (*jobModel.JobPosting, *config.CustomError) {
	if _, err := uuid.Parse(jobID); err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("get job: %w", config.ErrJobNotFound),
			config.GetErrCodeNotFound(),
		)
	}

	job, err := s.jobPostingRepo.FindByID(ctx, jobID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("get job: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if job == nil {
		return nil, config.NewCustomError(
			fmt.Errorf("get job: %w", config.ErrJobNotFound),
			config.GetErrCodeNotFound(),
		)
	}

	return job, nil
}
