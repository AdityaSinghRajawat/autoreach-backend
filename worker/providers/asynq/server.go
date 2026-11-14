package asynqProvider

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hibiken/asynq"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

type Server struct {
	aggregatorRegistry workerConfigs.WorkerAggregatorRegistry
}

func newServer(aggregatorRegistry workerConfigs.WorkerAggregatorRegistry) *Server {
	return &Server{aggregatorRegistry: aggregatorRegistry}
}

func (s *Server) Start(
	ctx context.Context,
	handlerRegistry workerConfigs.WorkerHandlerRegistry,
) error {
	server := asynq.NewServer(redisOptions(), asynq.Config{
		Concurrency:      config.GetWorkerConcurrency(),
		Queues:           workerConfigs.GetQueuePriorities(),
		ShutdownTimeout:  config.GetWorkerShutdownTimeout(),
		RetryDelayFunc:   retryDelay,
		IsFailure:        isFailure,
		ErrorHandler:     asynq.ErrorHandlerFunc(logTaskError),
		GroupAggregator:  asynq.GroupAggregatorFunc(s.aggregate),
		GroupMaxSize:     config.GetEmbedGroupMaxSize(),
		GroupMaxDelay:    config.GetEmbedGroupMaxDelay(),
		GroupGracePeriod: config.GetEmbedGroupGracePeriod(),
		LogLevel:         asynq.WarnLevel,
	})

	mux := asynq.NewServeMux()
	for taskID, handler := range handlerRegistry {
		mux.HandleFunc(string(taskID), adaptHandler(taskID, handler))
	}

	if err := server.Start(mux); err != nil {
		return fmt.Errorf("start asynq server: %w", err)
	}

	utils.LogInfo(
		ctx,
		"worker server consuming",
		"queues",
		workerConfigs.GetQueuePriorities(),
		"concurrency",
		config.GetWorkerConcurrency(),
	)

	<-ctx.Done()
	server.Shutdown()

	return nil
}

func adaptHandler(
	taskID workerConfigs.WorkerTaskID,
	handler workerConfigs.WorkerMessageHandler,
) asynq.HandlerFunc {
	return func(ctx context.Context, task *asynq.Task) error {
		decoded, err := decodeEnvelope(task.Payload())
		if err != nil {
			return fmt.Errorf("%w: %w", err, asynq.SkipRetry)
		}

		messageID, _ := asynq.GetTaskID(ctx)
		queueName, _ := asynq.GetQueueName(ctx)
		retryCount, _ := asynq.GetRetryCount(ctx)
		maxRetry, _ := asynq.GetMaxRetry(ctx)

		result := handler(ctx, workerConfigs.Message{
			TaskID:     taskID,
			MessageID:  messageID,
			Payload:    decoded.Payload,
			Metadata:   decoded.Metadata,
			QueueName:  queueName,
			RetryCount: retryCount,
			MaxRetry:   maxRetry,
		})

		return toAsynqError(result)
	}
}

func toAsynqError(result *workerConfigs.WorkerMessageResult) error {
	if result == nil || result.Success {
		return nil
	}

	err := result.Err
	if err == nil {
		err = fmt.Errorf("%w: %s", config.ErrTaskFailed, result.ErrCode)
	}

	if !result.IsRetryable {
		return fmt.Errorf("%w: %w", err, asynq.SkipRetry)
	}

	return err
}

func (s *Server) aggregate(groupKey string, tasks []*asynq.Task) *asynq.Task {
	aggregator, found := s.aggregatorRegistry[groupKey]
	if !found {
		utils.LogErrorWithoutCtx(
			fmt.Errorf("aggregate group %s: no aggregator registered", groupKey),
		)

		return asynq.NewTask(string(workerConfigs.TaskUnroutable), nil)
	}

	messages := make([]workerConfigs.Message, 0, len(tasks))
	for _, task := range tasks {
		decoded, err := decodeEnvelope(task.Payload())
		if err != nil {
			continue
		}

		messages = append(messages, workerConfigs.Message{
			TaskID:   workerConfigs.WorkerTaskID(task.Type()),
			Payload:  decoded.Payload,
			Metadata: decoded.Metadata,
			GroupKey: groupKey,
		})
	}

	aggregated, err := aggregator(groupKey, messages)
	if err != nil {
		utils.LogErrorWithoutCtx(fmt.Errorf("aggregate group %s: %w", groupKey, err))

		return asynq.NewTask(string(workerConfigs.TaskUnroutable), nil)
	}

	encoded, err := encodeEnvelope(aggregated.Payload, aggregated.Metadata)
	if err != nil {
		utils.LogErrorWithoutCtx(err)

		return asynq.NewTask(string(workerConfigs.TaskUnroutable), nil)
	}

	options := []asynq.Option{}
	if taskConfig, found := workerConfigs.GetTaskConfig(aggregated.TaskID); found {
		options = append(options, asynq.MaxRetry(taskConfig.MaxRetry))
		if taskConfig.Timeout > 0 {
			options = append(options, asynq.Timeout(taskConfig.Timeout))
		}
	}

	return asynq.NewTask(string(aggregated.TaskID), encoded, options...)
}

func retryDelay(retryCount int, _ error, task *asynq.Task) time.Duration {
	return workerConfigs.RetryDelay(workerConfigs.WorkerTaskID(task.Type()), retryCount)
}

func isFailure(err error) bool {
	return err != nil
}

func logTaskError(ctx context.Context, task *asynq.Task, err error) {
	retryCount, _ := asynq.GetRetryCount(ctx)
	maxRetry, _ := asynq.GetMaxRetry(ctx)
	utils.LogError(
		ctx,
		err,
		"task_id", task.Type(),
		"retry_count", retryCount,
		"max_retry", maxRetry,
		"will_retry", !errors.Is(err, asynq.SkipRetry) && retryCount < maxRetry,
	)
}
