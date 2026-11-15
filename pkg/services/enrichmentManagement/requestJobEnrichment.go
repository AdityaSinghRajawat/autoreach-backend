package enrichmentManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/enrichmentManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

func (s *EnrichmentService) RequestJobEnrichment(
	ctx context.Context,
	candidateID, jobID string,
) (*enrichmentTypes.RequestEnrichmentResponse, *config.CustomError) {
	if customErr := s.requireMatchedJob(ctx, candidateID, jobID); customErr != nil {
		return nil, customErr
	}

	_, err := s.workerClient.Publish(ctx, workerConfigs.PublishParams{
		TaskID:  workerConfigs.TaskEnrichContacts,
		Payload: workerCustomTypes.EnrichContactsPayload{JobID: jobID},
	})
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("request job enrichment: %w", err),
			config.GetErrCodeServiceUnavailable(),
		)
	}

	return enrichmentTypes.NewRequestEnrichmentResponse(
		jobID,
		string(workerConfigs.TaskEnrichContacts),
		workerConfigs.PublishStatusEnqueued,
	), nil
}
