package workerAggregators

import (
	"encoding/json"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

func AggregateEmbedJobs(
	groupKey string,
	messages []workerConfigs.Message,
) (workerConfigs.Message, error) {
	seen := make(map[string]struct{}, len(messages))
	batch := workerCustomTypes.EmbedJobBatchPayload{JobIDs: make([]string, 0, len(messages))}
	metadata := map[string]string{}

	for _, message := range messages {
		payload := workerCustomTypes.EmbedJobPayload{}
		if err := json.Unmarshal(message.Payload, &payload); err != nil {
			continue
		}

		if err := validations.Validate(&payload); err != nil {
			continue
		}

		if _, duplicate := seen[payload.JobID]; duplicate {
			continue
		}

		seen[payload.JobID] = struct{}{}
		batch.JobIDs = append(batch.JobIDs, payload.JobID)

		if len(metadata) == 0 {
			metadata = message.Metadata
		}
	}

	if len(batch.JobIDs) == 0 {
		return workerConfigs.Message{}, fmt.Errorf(
			"aggregate embed jobs: %w",
			config.ErrNoValidMessages,
		)
	}

	encoded, err := json.Marshal(batch)
	if err != nil {
		return workerConfigs.Message{}, fmt.Errorf("aggregate embed jobs: %w", err)
	}

	return workerConfigs.Message{
		TaskID:   workerConfigs.TaskEmbedJobBatch,
		Payload:  encoded,
		Metadata: metadata,
		GroupKey: groupKey,
	}, nil
}
