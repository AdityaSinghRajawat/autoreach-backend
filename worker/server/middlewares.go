package workerServer

import (
	"context"
	"log/slog"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

func wrapWorkerMiddleware(
	logger *slog.Logger,
	handler workerConfigs.WorkerMessageHandler,
) workerConfigs.WorkerMessageHandler {
	return workerLoggerMiddleware(logger, workerMetricsMiddleware(handler))
}

func workerMetricsMiddleware(
	next workerConfigs.WorkerMessageHandler,
) workerConfigs.WorkerMessageHandler {
	return func(ctx context.Context, message workerConfigs.Message) *workerConfigs.WorkerMessageResult {
		startedAt := time.Now()
		result := next(ctx, message)

		utils.RecordWorkerTask(
			workerConfigs.GetTaskStage(message.TaskID),
			string(message.TaskID),
			taskOutcome(message, result),
			time.Since(startedAt),
		)

		return result
	}
}

// A failure is counted as a retry while attempts remain, and as failed once the task will not run again.
func taskOutcome(message workerConfigs.Message, result *workerConfigs.WorkerMessageResult) string {
	switch {
	case result == nil || result.Success:
		return config.GetTaskOutcomeSuccess()
	case result.IsRetryable && message.RetryCount < message.MaxRetry:
		return config.GetTaskOutcomeRetry()
	default:
		return config.GetTaskOutcomeFailed()
	}
}

func workerLoggerMiddleware(
	logger *slog.Logger,
	next workerConfigs.WorkerMessageHandler,
) workerConfigs.WorkerMessageHandler {
	return func(ctx context.Context, message workerConfigs.Message) *workerConfigs.WorkerMessageResult {
		startedAt := time.Now()

		correlationID := message.Metadata[workerConfigs.MetadataCorrelationID]
		if correlationID == "" {
			correlationID = utils.GenerateUUID()
		}

		ctx = utils.WithLogger(ctx, logger.With(
			slog.String("request_id", message.MessageID),
			slog.String("task_id", string(message.TaskID)),
		))
		ctx = utils.WithCorrelationID(ctx, correlationID)
		ctx = utils.WithRequestID(ctx, message.MessageID)

		result := next(ctx, message)

		utils.LogInfo(
			ctx,
			"task completed",
			"queue", message.QueueName,
			"success", result == nil || result.Success,
			"retry_count", message.RetryCount,
			"latency_ms", time.Since(startedAt).Milliseconds(),
		)

		return result
	}
}
