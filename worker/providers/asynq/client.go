package asynqProvider

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hibiken/asynq"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

type Client struct {
	client    *asynq.Client
	inspector *asynq.Inspector
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

	groupKey := resolveGroupKey(params, taskConfig)
	if groupKey != "" && params.ProcessIn == 0 {
		options = append(options, asynq.Group(groupKey))

		return c.enqueue(ctx, params.TaskID, encoded, options)
	}

	return c.enqueueOnce(ctx, params.TaskID, taskConfig.QueueName, payload, encoded, options)
}

// enqueueOnce names the task after its payload, so publishing the same work again while it is still
// waiting, running or retrying is a no-op. A finished or dead task does not block a new one.
func (c *Client) enqueueOnce(
	ctx context.Context,
	taskID workerConfigs.WorkerTaskID,
	queue string,
	payload, encoded []byte,
	options []asynq.Option,
) (*workerConfigs.PublishResult, error) {
	id := fmt.Sprintf(config.GetTaskIDFormat(), taskID, sha256.Sum256(payload))
	options = append(options, asynq.TaskID(id))

	result, err := c.enqueue(ctx, taskID, encoded, options)
	if !errors.Is(err, asynq.ErrTaskIDConflict) {
		return result, err
	}

	existing, err := c.inspector.GetTaskInfo(queue, id)
	if errors.Is(err, asynq.ErrQueueNotFound) {
		return &workerConfigs.PublishResult{MessageID: id, QueueName: queue}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("publish task %s: inspect existing: %w", taskID, err)
	}

	if existing.State != asynq.TaskStateCompleted && existing.State != asynq.TaskStateArchived {
		return &workerConfigs.PublishResult{MessageID: id, QueueName: queue}, nil
	}

	if err := c.inspector.DeleteTask(queue, id); err != nil {
		return nil, fmt.Errorf("publish task %s: replace finished task: %w", taskID, err)
	}

	result, err = c.enqueue(ctx, taskID, encoded, options)
	if errors.Is(err, asynq.ErrTaskIDConflict) {
		return &workerConfigs.PublishResult{MessageID: id, QueueName: queue}, nil
	}

	return result, err
}

func (c *Client) enqueue(
	ctx context.Context,
	taskID workerConfigs.WorkerTaskID,
	encoded []byte,
	options []asynq.Option,
) (*workerConfigs.PublishResult, error) {
	info, err := c.client.EnqueueContext(ctx, asynq.NewTask(string(taskID), encoded), options...)
	if err != nil {
		return nil, fmt.Errorf("publish task %s: %w", taskID, err)
	}

	return &workerConfigs.PublishResult{MessageID: info.ID, QueueName: info.Queue}, nil
}

func (c *Client) Close() error {
	if err := errors.Join(c.client.Close(), c.inspector.Close()); err != nil {
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
