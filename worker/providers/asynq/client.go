package asynqProvider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

type Client struct {
	client *asynq.Client
}

func (c *Client) Publish(
	ctx context.Context,
	params workerConfigs.PublishParams,
) (*workerConfigs.PublishResult, error) {
	taskConfig, found := workerConfigs.GetTaskConfig(params.TaskID)
	if !found {
		return nil, fmt.Errorf("publish task %s: %w", params.TaskID, config.ErrTaskConfigMissing)
	}

	payload, err := json.Marshal(params.Payload)
	if err != nil {
		return nil, fmt.Errorf("publish task %s: encode payload: %w", params.TaskID, err)
	}

	encoded, err := encodeEnvelope(payload, buildMetadata(ctx, params.Metadata))
	if err != nil {
		return nil, err
	}

	options := []asynq.Option{
		asynq.Queue(taskConfig.QueueName),
		asynq.MaxRetry(taskConfig.MaxRetry),
	}
	if taskConfig.Timeout > 0 {
		options = append(options, asynq.Timeout(taskConfig.Timeout))
	}

	if taskConfig.Retention > 0 {
		options = append(options, asynq.Retention(taskConfig.Retention))
	}

	if params.ProcessIn > 0 {
		options = append(options, asynq.ProcessIn(params.ProcessIn))
	}

	if groupKey := resolveGroupKey(params, taskConfig); groupKey != "" && params.ProcessIn == 0 {
		options = append(options, asynq.Group(groupKey))
	}

	info, err := c.client.EnqueueContext(
		ctx,
		asynq.NewTask(string(params.TaskID), encoded),
		options...)
	if err != nil {
		return nil, fmt.Errorf("publish task %s: %w", params.TaskID, err)
	}

	return &workerConfigs.PublishResult{MessageID: info.ID, QueueName: info.Queue}, nil
}

func (c *Client) Close() error {
	if err := c.client.Close(); err != nil {
		return fmt.Errorf("close asynq client: %w", err)
	}

	return nil
}

func resolveGroupKey(
	params workerConfigs.PublishParams,
	taskConfig workerConfigs.TaskConfig,
) string {
	if params.GroupKey != "" {
		return params.GroupKey
	}

	return taskConfig.AggregatedBy
}

func buildMetadata(ctx context.Context, supplied map[string]string) map[string]string {
	metadata := make(map[string]string, len(supplied)+2)
	for key, value := range supplied {
		metadata[key] = value
	}

	metadata[workerConfigs.MetadataProvider] = string(workerConfigs.ProviderAsynq)
	if correlationID := utils.GetCorrelationID(ctx); correlationID != "" {
		metadata[workerConfigs.MetadataCorrelationID] = correlationID
	}

	return metadata
}
