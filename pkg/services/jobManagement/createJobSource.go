package jobManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
)

func (s *JobService) CreateJobSource(
	ctx context.Context,
	source *jobModel.JobSource,
) (*jobModel.JobSource, *config.CustomError) {
	if customErr := s.saveJobSource(ctx, source, "create job source"); customErr != nil {
		return nil, customErr
	}

	s.publishPollJobSource(ctx, source.ID)

	return source, nil
}
