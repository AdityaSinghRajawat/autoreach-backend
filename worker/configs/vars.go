package workerConfigs

import (
	"fmt"
	"slices"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

const (
	ProviderAsynq ProviderName = "asynq"

	TaskEmbedJob      WorkerTaskID = "task:embed_job"
	TaskEmbedJobBatch WorkerTaskID = "task:embed_job_batch"
	TaskProcessResume WorkerTaskID = "task:process_resume"
	TaskMatchResume   WorkerTaskID = "task:match_resume"
	TaskMatchJob      WorkerTaskID = "task:match_job"

	QueueEmbedding = "embedding_queue"
	QueueMatching  = "matching_queue"

	GroupEmbedJob = "embed_job"

	MetadataCorrelationID = "correlationId"
	MetadataProvider      = "provider"

	PublishStatusEnqueued = "ENQUEUED"
	PublishStatusDeferred = "DEFERRED"
)

type TaskConfig struct {
	QueueName    string
	MaxRetry     int
	RetryBackoff []time.Duration
	Timeout      time.Duration
	Retention    time.Duration
	AggregatedBy string
}

type QueueDefinition struct {
	Name     string
	Priority int
}

var (
	embeddingBackoffI = []time.Duration{
		5 * time.Second,
		15 * time.Second,
		45 * time.Second,
		2 * time.Minute,
		5 * time.Minute,
	}
	matchingBackoffI = []time.Duration{5 * time.Second, 20 * time.Second, time.Minute}
)

var taskConfigsI = map[WorkerTaskID]TaskConfig{
	TaskEmbedJob: {
		QueueName:    QueueEmbedding,
		MaxRetry:     5,
		RetryBackoff: embeddingBackoffI,
		Timeout:      60 * time.Second,
		AggregatedBy: GroupEmbedJob,
	},
	TaskEmbedJobBatch: {
		QueueName:    QueueEmbedding,
		MaxRetry:     5,
		RetryBackoff: embeddingBackoffI,
		Timeout:      60 * time.Second,
	},
	TaskProcessResume: {
		QueueName:    QueueEmbedding,
		MaxRetry:     5,
		RetryBackoff: embeddingBackoffI,
		Timeout:      2 * time.Minute,
	},
	TaskMatchResume: {
		QueueName:    QueueMatching,
		MaxRetry:     3,
		RetryBackoff: matchingBackoffI,
		Timeout:      60 * time.Second,
	},
	TaskMatchJob: {
		QueueName:    QueueMatching,
		MaxRetry:     3,
		RetryBackoff: matchingBackoffI,
		Timeout:      30 * time.Second,
	},
}

var queueDefinitionsI = []QueueDefinition{
	{Name: QueueEmbedding, Priority: 3},
	{Name: QueueMatching, Priority: 2},
}

func GetTaskConfig(taskID WorkerTaskID) (TaskConfig, bool) {
	taskConfig, found := taskConfigsI[taskID]

	return taskConfig, found
}

func GetTaskIDs() []WorkerTaskID {
	taskIDs := make([]WorkerTaskID, 0, len(taskConfigsI))
	for taskID := range taskConfigsI {
		taskIDs = append(taskIDs, taskID)
	}

	slices.Sort(taskIDs)

	return taskIDs
}

func GetQueueDefinitions() []QueueDefinition {
	return append([]QueueDefinition(nil), queueDefinitionsI...)
}

func GetQueuePriorities() map[string]int {
	priorities := make(map[string]int, len(queueDefinitionsI))
	for _, definition := range queueDefinitionsI {
		priorities[definition.Name] = definition.Priority
	}

	return priorities
}

func RetryDelay(taskID WorkerTaskID, retryCount int) time.Duration {
	taskConfig, found := taskConfigsI[taskID]
	if !found || len(taskConfig.RetryBackoff) == 0 {
		return time.Duration(retryCount+1) * 10 * time.Second
	}

	if retryCount >= len(taskConfig.RetryBackoff) {
		return taskConfig.RetryBackoff[len(taskConfig.RetryBackoff)-1]
	}

	if retryCount < 0 {
		return taskConfig.RetryBackoff[0]
	}

	return taskConfig.RetryBackoff[retryCount]
}

func ValidateTopology(
	handlerRegistry WorkerHandlerRegistry,
	aggregatorRegistry WorkerAggregatorRegistry,
) error {
	queues := GetQueuePriorities()
	for taskID, taskConfig := range taskConfigsI {
		if _, found := queues[taskConfig.QueueName]; !found {
			return fmt.Errorf(
				"%w: task %s uses undeclared queue %s",
				config.ErrInvalidTopology,
				taskID,
				taskConfig.QueueName,
			)
		}

		if _, found := handlerRegistry[taskID]; !found {
			return fmt.Errorf(
				"%w: task %s has a config but no handler",
				config.ErrInvalidTopology,
				taskID,
			)
		}

		if taskConfig.AggregatedBy == "" {
			continue
		}

		if _, found := aggregatorRegistry[taskConfig.AggregatedBy]; !found {
			return fmt.Errorf(
				"%w: task %s is aggregated by %s but no aggregator is registered",
				config.ErrInvalidTopology,
				taskID,
				taskConfig.AggregatedBy,
			)
		}
	}

	for taskID := range handlerRegistry {
		if _, found := taskConfigsI[taskID]; !found {
			return fmt.Errorf(
				"%w: task %s has a handler but no config",
				config.ErrInvalidTopology,
				taskID,
			)
		}
	}

	return nil
}
