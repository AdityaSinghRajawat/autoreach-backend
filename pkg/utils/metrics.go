package utils

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

var (
	workerTasksI = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: config.GetMetricsNamespace(),
		Name:      config.GetWorkerTasksMetric(),
		Help:      "Worker tasks handled, by pipeline stage, task and outcome (success, retry, failed).",
	}, []string{"stage", "task", "outcome"})

	workerTaskDurationI = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: config.GetMetricsNamespace(),
		Name:      config.GetWorkerTaskDurationMetric(),
		Help:      "Time spent handling a worker task, by pipeline stage and task.",
		Buckets:   prometheus.ExponentialBuckets(0.01, 4, 8),
	}, []string{"stage", "task"})

	httpRequestsI = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: config.GetMetricsNamespace(),
		Name:      config.GetHTTPRequestsMetric(),
		Help:      "HTTP requests served, by method, route template and status.",
	}, []string{"method", "route", "status"})

	httpDurationI = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: config.GetMetricsNamespace(),
		Name:      config.GetHTTPDurationMetric(),
		Help:      "HTTP request latency, by method and route template.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "route"})
)

func RecordWorkerTask(stage, task, outcome string, duration time.Duration) {
	workerTasksI.WithLabelValues(stage, task, outcome).Inc()
	workerTaskDurationI.WithLabelValues(stage, task).Observe(duration.Seconds())
}

func RecordHTTPRequest(method, route string, status int, duration time.Duration) {
	httpRequestsI.WithLabelValues(method, route, strconv.Itoa(status)).Inc()
	httpDurationI.WithLabelValues(method, route).Observe(duration.Seconds())
}
