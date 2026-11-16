package jobManagement

import (
	"context"
	"errors"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

func (s *JobService) saveJobSource(
	ctx context.Context,
	source *jobModel.JobSource,
	operation string,
) *config.CustomError {
	source.UpdateTimestamp()

	if err := source.Validate(); err != nil {
		return config.NewCustomError(
			fmt.Errorf("%s: %w", operation, err),
			config.GetErrCodeBadRequest(),
		)
	}

	_, err := s.jobSourceRepo.Upsert(ctx, source)
	if errors.Is(err, config.ErrDuplicateRecord) {
		return config.NewCustomError(
			fmt.Errorf("%s: %w", operation, err),
			config.GetErrCodeConflict(),
		)
	}

	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("%s: %w", operation, err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return nil
}

func (s *JobService) findJobSource(
	ctx context.Context,
	sourceID string,
) (*jobModel.JobSource, *config.CustomError) {
	source, err := s.jobSourceRepo.FindByID(ctx, sourceID)

	return config.RequireFound("find job source", source, err, config.ErrJobSourceNotFound)
}

func (s *JobService) publishPollJobSource(ctx context.Context, sourceID string) {
	_, err := s.workerClient.Publish(ctx, workerConfigs.PublishParams{
		TaskID:  workerConfigs.TaskPollJobSource,
		Payload: workerCustomTypes.PollJobSourcePayload{SourceID: sourceID},
	})
	if err != nil {
		utils.LogWarn(
			ctx,
			"job source poll publish failed",
			"source_id",
			sourceID,
			"error",
			err.Error(),
		)
	}
}
