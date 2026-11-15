package jobManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
)

func (s *JobService) ListJobSources(
	ctx context.Context,
) ([]*jobModel.JobSource, *config.CustomError) {
	sources, err := s.jobSourceRepo.List(ctx)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("list job sources: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return sources, nil
}
