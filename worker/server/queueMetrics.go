package workerServer

import (
	"context"
	"errors"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

// queueMetricsCollector reads queue depth from the queue backend on every scrape.
type queueMetricsCollector struct {
	inspector workerConfigs.IWorkerInspector
	tasks     *prometheus.Desc
}

func registerQueueMetrics(inspector workerConfigs.IWorkerInspector) {
	collector := &queueMetricsCollector{
		inspector: inspector,
		tasks: prometheus.NewDesc(
			prometheus.BuildFQName(config.GetMetricsNamespace(), "", config.GetQueueTasksMetric()),
			"Tasks in each queue, by state.",
			[]string{"queue", "state"},
			nil,
		),
	}

	alreadyRegistered := prometheus.AlreadyRegisteredError{}
	if err := prometheus.Register(collector); err != nil && !errors.As(err, &alreadyRegistered) {
		utils.LogErrorWithoutCtx(err)
	}
}

func (c *queueMetricsCollector) Describe(descriptions chan<- *prometheus.Desc) {
	descriptions <- c.tasks
}

func (c *queueMetricsCollector) Collect(metrics chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(config.GetQueueMetricsTimeoutSeconds())*time.Second,
	)
	defer cancel()

	queues, err := c.inspector.ListQueues(ctx)
	if err != nil {
		utils.LogErrorWithoutCtx(err)

		return
	}

	for _, queue := range queues {
		for state, count := range map[string]int{
			"pending":   queue.Pending,
			"active":    queue.Active,
			"scheduled": queue.Scheduled,
			"retry":     queue.Retry,
			"dead":      queue.Dead,
		} {
			metrics <- prometheus.MustNewConstMetric(
				c.tasks,
				prometheus.GaugeValue,
				float64(count),
				queue.Queue,
				state,
			)
		}
	}
}
