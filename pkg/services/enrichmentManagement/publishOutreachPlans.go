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

	for _, candidateID := range candidateIDs {
		_, err := s.workerClient.Publish(ctx, workerConfigs.PublishParams{
			TaskID:  workerConfigs.TaskPlanOutreach,
			Payload: workerCustomTypes.PlanOutreachPayload{CandidateID: candidateID, JobID: jobID},
		})
		if err != nil {
			utils.LogError(ctx, err, "job_id", jobID, "candidate_id", candidateID)
		}
	}
}
