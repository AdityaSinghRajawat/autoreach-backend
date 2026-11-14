package workerServer

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

func wrapWorkerMiddleware(
	logger *slog.Logger,
	handler workerConfigs.WorkerMessageHandler,
) workerConfigs.WorkerMessageHandler {
	return workerLoggerMiddleware(logger, handler)
}

func workerLoggerMiddleware(
	logger *slog.Logger,
	next workerConfigs.WorkerMessageHandler,
) workerConfigs.WorkerMessageHandler {
	return func(ctx context.Context, message workerConfigs.Message) *workerConfigs.WorkerMessageResult {
		startedAt := time.Now()

		correlationID := message.Metadata[workerConfigs.MetadataCorrelationID]
		if correlationID == "" {
			correlationID = uuid.NewString()
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
