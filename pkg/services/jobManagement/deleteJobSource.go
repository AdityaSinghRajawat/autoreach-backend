package jobManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

func (s *JobService) DeleteJobSource(ctx context.Context, sourceID string) *config.CustomError {
	deleted, err := s.jobSourceRepo.Delete(ctx, sourceID)

	return config.RequireAffected("delete job source", deleted, err, config.ErrJobSourceNotFound)
}
