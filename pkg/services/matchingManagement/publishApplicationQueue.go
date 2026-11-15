package matchingManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

func (s *MatchingService) publishApplicationQueue(ctx context.Context, candidateID string) {
	_, err := s.workerClient.Publish(ctx, workerConfigs.PublishParams{
		TaskID:  workerConfigs.TaskQueueApplications,
		Payload: workerCustomTypes.QueueApplicationsPayload{CandidateID: candidateID},
	})
	if err != nil {
		utils.LogError(ctx, err, "candidate_id", candidateID)
	}
}
