package matchingManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

func (s *MatchingService) fanOutMatches(
	ctx context.Context,
	candidateID string,
	minMatchScore float64,
	autoOutreachEnabled bool,
) int {
	if !autoOutreachEnabled {
		return 0
	}

	jobIDs, err := s.jobMatchRepo.ClaimFanOut(
		ctx,
		candidateID,
		minMatchScore,
		utils.GetCurrTimeStamp(),
	)
	if err != nil {
		utils.LogError(ctx, err, "candidate_id", candidateID)

		return 0
	}

	tasks := make([]utils.ConcurrentTask, 0, len(jobIDs))
	for _, jobID := range jobIDs {
		tasks = append(tasks, utils.ConcurrentTask{
			Name: "publish enrich contacts task",
			Run: func(ctx context.Context) error {
				s.publishEnrichContacts(ctx, candidateID, jobID)

				return nil
			},
		})
	}

	if err := utils.RunConcurrently(context.WithoutCancel(ctx), tasks...); err != nil {
		utils.LogError(ctx, err)
	}

	return len(jobIDs)
}

func (s *MatchingService) publishEnrichContacts(ctx context.Context, candidateID, jobID string) {
	_, err := s.workerClient.Publish(ctx, workerConfigs.PublishParams{
		TaskID:  workerConfigs.TaskEnrichContacts,
		Payload: workerCustomTypes.EnrichContactsPayload{JobID: jobID},
	})
	if err == nil {
		return
	}

	utils.LogWarn(
		ctx,
		"enrich contacts task publish failed, releasing the claim",
		"job_id", jobID,
		"error", err.Error(),
	)

	if err := s.jobMatchRepo.ReleaseFanOut(ctx, candidateID, jobID); err != nil {
		utils.LogError(ctx, err, "job_id", jobID)
	}
}
