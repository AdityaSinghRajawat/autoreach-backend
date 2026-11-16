package enrichmentManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

func (s *EnrichmentService) publishOutreachPlans(ctx context.Context, jobID string) {
	candidateIDs, err := s.jobMatchRepo.ListFannedOutCandidateIDs(ctx, jobID)
	if err != nil {
		utils.LogError(ctx, err, "job_id", jobID)

		return
	}

	tasks := make([]utils.ConcurrentTask, 0, len(candidateIDs))
	for _, item := range candidateIDs {
		tasks = append(tasks, utils.ConcurrentTask{
			Name: "publish outreach plan",
			Run: func(ctx context.Context) error {
				_, err := s.workerClient.Publish(ctx, workerConfigs.PublishParams{
					TaskID:  workerConfigs.TaskPlanOutreach,
					Payload: workerCustomTypes.PlanOutreachPayload{CandidateID: item, JobID: jobID},
				})
				if err != nil {
					utils.LogError(ctx, err, "job_id", jobID, "candidate_id", item)
				}

				return nil
			},
		})
	}

	_ = utils.RunConcurrently(ctx, tasks...)
}
