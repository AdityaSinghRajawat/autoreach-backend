package config

const (
	workerTasksMetric        = "worker_tasks_total"
	workerTaskDurationMetric = "worker_task_duration_seconds"
	httpRequestsMetric       = "http_requests_total"
	httpDurationMetric       = "http_request_duration_seconds"
	queueTasksMetric         = "queue_tasks"
	unmatchedRoute           = "unmatched"
	taskOutcomeSuccess       = "success"
	taskOutcomeRetry         = "retry"
	taskOutcomeFailed        = "failed"
	queueMetricsTimeoutSecs  = 5
)

func GetWorkerTasksMetric() string        { return workerTasksMetric }
func GetWorkerTaskDurationMetric() string { return workerTaskDurationMetric }
func GetHTTPRequestsMetric() string       { return httpRequestsMetric }
func GetHTTPDurationMetric() string       { return httpDurationMetric }
func GetQueueTasksMetric() string         { return queueTasksMetric }
func GetUnmatchedRoute() string           { return unmatchedRoute }
func GetTaskOutcomeSuccess() string       { return taskOutcomeSuccess }
func GetTaskOutcomeRetry() string         { return taskOutcomeRetry }
func GetTaskOutcomeFailed() string        { return taskOutcomeFailed }
func GetQueueMetricsTimeoutSeconds() int  { return queueMetricsTimeoutSecs }
