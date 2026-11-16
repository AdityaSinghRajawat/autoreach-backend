package asynqProvider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hibiken/asynq"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

type Inspector struct {
	inspector *asynq.Inspector
}

func (s *Strategy) InitializeInspector(_ context.Context) (workerConfigs.IWorkerInspector, error) {
	return &Inspector{inspector: asynq.NewInspector(redisOptions())}, nil
}

// ListQueues reports every configured queue, including ones that have not received a task yet.
func (i *Inspector) ListQueues(_ context.Context) ([]workerConfigs.QueueInfo, error) {
	existing, err := i.existingQueues()
	if err != nil {
		return nil, err
	}

	definitions := workerConfigs.GetQueueDefinitions()
	queues := make([]workerConfigs.QueueInfo, 0, len(definitions))

	for _, definition := range definitions {
		if !existing[definition.Name] {
			queues = append(queues, workerConfigs.QueueInfo{Queue: definition.Name})

			continue
		}

		info, err := i.inspector.GetQueueInfo(definition.Name)
		if err != nil {
			return nil, fmt.Errorf("get queue info %s: %w", definition.Name, err)
		}

		queues = append(queues, workerConfigs.QueueInfo{
			Queue:     info.Queue,
			Size:      info.Size,
			Pending:   info.Pending,
			Active:    info.Active,
			Scheduled: info.Scheduled,
			Retry:     info.Retry,
			Dead:      info.Archived,
			Processed: info.Processed,
			Failed:    info.Failed,
			Paused:    info.Paused,
			Latency:   info.Latency,
		})
	}

	return queues, nil
}

func (i *Inspector) ListDeadTasks(
	_ context.Context,
	queue string,
	page, perPage int,
) ([]workerConfigs.DeadTask, int, error) {
	existing, err := i.existingQueues()
	if err != nil {
		return nil, 0, err
	}

	if !existing[queue] {
		return []workerConfigs.DeadTask{}, 0, nil
	}

	info, err := i.inspector.GetQueueInfo(queue)
	if err != nil {
		return nil, 0, fmt.Errorf("get queue info %s: %w", queue, err)
	}

	tasks, err := i.inspector.ListArchivedTasks(queue, asynq.Page(page), asynq.PageSize(perPage))
	if err != nil {
		return nil, 0, fmt.Errorf("list dead tasks %s: %w", queue, err)
	}

	deadTasks := make([]workerConfigs.DeadTask, 0, len(tasks))

	for _, task := range tasks {
		payload := task.Payload
		if decoded, err := decodeEnvelope(task.Payload); err == nil {
			payload = decoded.Payload
		}

		deadTasks = append(deadTasks, workerConfigs.DeadTask{
			ID:           task.ID,
			TaskID:       workerConfigs.WorkerTaskID(task.Type),
			Queue:        task.Queue,
			Payload:      payload,
			LastError:    task.LastErr,
			LastFailedAt: task.LastFailedAt,
			Retried:      task.Retried,
			MaxRetry:     task.MaxRetry,
		})
	}

	return deadTasks, info.Archived, nil
}

func (i *Inspector) RetryDeadTask(_ context.Context, queue, id string) error {
	info, err := i.inspector.GetTaskInfo(queue, id)
	if errors.Is(err, asynq.ErrQueueNotFound) || errors.Is(err, asynq.ErrTaskNotFound) ||
		(err == nil && info.State != asynq.TaskStateArchived) {
		return fmt.Errorf("retry dead task %s: %w", id, config.ErrDeadTaskNotFound)
	}

	if err != nil {
		return fmt.Errorf("retry dead task %s: %w", id, err)
	}

	if err := i.inspector.RunTask(queue, id); err != nil {
		return fmt.Errorf("retry dead task %s: %w", id, err)
	}

	return nil
}

// asynq creates a queue when its first task arrives and reports unknown queues only as an internal
// error, so queues are checked against its list first.
func (i *Inspector) existingQueues() (map[string]bool, error) {
	names, err := i.inspector.Queues()
	if err != nil {
		return nil, fmt.Errorf("list queues: %w", err)
	}

	existing := make(map[string]bool, len(names))
	for _, name := range names {
		existing[name] = true
	}

	return existing, nil
}

func (i *Inspector) Close() error {
	return i.inspector.Close()
}
