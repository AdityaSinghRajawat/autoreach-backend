package jobManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
)

func (s *JobService) UpdateJobSource(
	ctx context.Context,
	sourceID string,
	request jobTypes.UpdateJobSourceRequest,
) (*jobModel.JobSource, *config.CustomError) {
	if request.IsEmpty() {
		return nil, config.NewCustomError(
			fmt.Errorf("update job source: %w", config.ErrNoFieldsSupplied),
			config.GetErrCodeBadRequest(),
		)
	}

	existing, customErr := s.findJobSource(ctx, sourceID)
	if customErr != nil {
		return nil, customErr
	}

	source := request.ToJobSourceModel(existing)
	if customErr := s.saveJobSource(ctx, source, "update job source"); customErr != nil {
		return nil, customErr
	}

	return source, nil
}
