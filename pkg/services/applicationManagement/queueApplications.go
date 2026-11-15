package applicationManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

func (s *ApplicationService) QueueApplications(
	ctx context.Context,
	candidateID string,
) (int, *config.CustomError) {
	candidate, customErr := s.findCandidate(ctx, candidateID)
	if customErr != nil || !candidate.AutoApplyEnabled {
		return 0, customErr
	}

	applicationIDs, err := s.applicationRepo.QueueFromMatches(
		ctx,
		candidateID,
		candidate.MinMatchScore,
		utils.GetCurrTimeStamp(),
	)
	if err != nil {
		return 0, config.NewCustomError(
			fmt.Errorf("queue applications: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	for _, applicationID := range applicationIDs {
		s.publishPrepare(ctx, applicationID)
	}

	return len(applicationIDs), nil
}

// A failed publish leaves the row QUEUED for the reconciler, so it is only logged.
func (s *ApplicationService) publishPrepare(ctx context.Context, applicationID string) {
	_, err := s.workerClient.Publish(ctx, workerConfigs.PublishParams{
		TaskID:  workerConfigs.TaskPrepareApplication,
		Payload: workerCustomTypes.PrepareApplicationPayload{ApplicationID: applicationID},
	})
	if err != nil {
		utils.LogWarn(ctx, "prepare application publish failed", "application_id", applicationID,
			"error", err.Error())
	}
}
