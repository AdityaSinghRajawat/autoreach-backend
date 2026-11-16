package jobManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
)

func (s *JobService) GetJob(
	ctx context.Context,
	jobID string,
) (*jobModel.JobPosting, *config.CustomError) {
	job, err := s.jobPostingRepo.FindByID(ctx, jobID)

	return config.RequireFound("get job", job, err, config.ErrJobNotFound)
}
