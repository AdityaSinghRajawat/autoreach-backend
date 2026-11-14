package workerConfigs

import (
	"context"
	"time"
)

type WorkerTaskID string

type ProviderName string

type Message struct {
	TaskID     WorkerTaskID
	MessageID  string
	Payload    []byte
	Metadata   map[string]string
	QueueName  string
	GroupKey   string
	RetryCount int
	MaxRetry   int
}

type PublishParams struct {
	TaskID    WorkerTaskID
	Payload   any
	GroupKey  string
	ProcessIn time.Duration
	Metadata  map[string]string
}

type PublishResult struct {
	MessageID string
	QueueName string
}

type WorkerMessageHandler func(ctx context.Context, message Message) *WorkerMessageResult

type WorkerHandlerRegistry map[WorkerTaskID]WorkerMessageHandler

type WorkerAggregator func(groupKey string, messages []Message) (Message, error)

type WorkerAggregatorRegistry map[string]WorkerAggregator

type IWorkerClient interface {
	Publish(ctx context.Context, params PublishParams) (*PublishResult, error)
	Close() error
}

type IWorkerServer interface {
	Start(ctx context.Context, handlerRegistry WorkerHandlerRegistry) error
}

type IWorkerStrategy interface {
	InitializeClient(ctx context.Context) (IWorkerClient, error)
	InitializeServer(
		ctx context.Context,
		aggregatorRegistry WorkerAggregatorRegistry,
	) (IWorkerServer, error)
}
