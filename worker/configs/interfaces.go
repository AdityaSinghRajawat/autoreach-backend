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

// QueueInfo is a snapshot of one queue; Processed and Failed count today's tasks.
type QueueInfo struct {
	Queue     string
	Size      int
	Pending   int
	Active    int
	Scheduled int
	Retry     int
	Dead      int
	Processed int
	Failed    int
	Paused    bool
	Latency   time.Duration
}

// DeadTask is a task that used up its retries or failed permanently.
type DeadTask struct {
	ID           string
	TaskID       WorkerTaskID
	Queue        string
	Payload      []byte
	LastError    string
	LastFailedAt time.Time
	Retried      int
	MaxRetry     int
}

type IWorkerInspector interface {
	ListQueues(ctx context.Context) ([]QueueInfo, error)
	ListDeadTasks(ctx context.Context, queue string, page, perPage int) ([]DeadTask, int, error)
	RetryDeadTask(ctx context.Context, queue, id string) error
	Close() error
}

type IWorkerStrategy interface {
	InitializeInspector(ctx context.Context) (IWorkerInspector, error)
	InitializeClient(ctx context.Context) (IWorkerClient, error)
	InitializeServer(
		ctx context.Context,
		aggregatorRegistry WorkerAggregatorRegistry,
	) (IWorkerServer, error)
}
