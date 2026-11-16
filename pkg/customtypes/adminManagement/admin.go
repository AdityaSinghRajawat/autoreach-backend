package adminManagement

import (
	"encoding/json"
	"time"

	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

type QueueSummary struct {
	Queue          string  `json:"queue"`
	Size           int     `json:"size"`
	Pending        int     `json:"pending"`
	Active         int     `json:"active"`
	Scheduled      int     `json:"scheduled"`
	Retry          int     `json:"retry"`
	Dead           int     `json:"dead"`
	ProcessedToday int     `json:"processedToday"`
	FailedToday    int     `json:"failedToday"`
	Paused         bool    `json:"paused"`
	LatencySeconds float64 `json:"latencySeconds"`
}

type DeadTask struct {
	TaskID       string          `json:"taskId"`
	TaskType     string          `json:"taskType"`
	Queue        string          `json:"queue"`
	Payload      json.RawMessage `json:"payload"`
	LastError    string          `json:"lastError"`
	LastFailedAt *time.Time      `json:"lastFailedAt"`
	Retried      int             `json:"retried"`
	MaxRetry     int             `json:"maxRetry"`
}

type EnrichmentProviderStatus struct {
	Code              string `json:"code"`
	Configured        bool   `json:"configured"`
	DailyCreditLimit  int    `json:"dailyCreditLimit"`
	CreditsUsedToday  int64  `json:"creditsUsedToday"`
	RateLimit         int    `json:"rateLimit"`
	RateWindowSeconds int    `json:"rateWindowSeconds"`
}

// EmailDeliverability covers cold emails sent in the last WindowDays days.
type EmailDeliverability struct {
	WindowDays               int      `json:"windowDays"`
	Sent                     int      `json:"sent"`
	Bounced                  int      `json:"bounced"`
	DeliverabilityPercentage *float64 `json:"deliverabilityPercentage"`
}

type ProvidersResponse struct {
	Enrichment []EnrichmentProviderStatus `json:"enrichment"`
	Email      EmailDeliverability        `json:"email"`
}

func NewQueueSummary(info workerConfigs.QueueInfo) QueueSummary {
	return QueueSummary{
		Queue:          info.Queue,
		Size:           info.Size,
		Pending:        info.Pending,
		Active:         info.Active,
		Scheduled:      info.Scheduled,
		Retry:          info.Retry,
		Dead:           info.Dead,
		ProcessedToday: info.Processed,
		FailedToday:    info.Failed,
		Paused:         info.Paused,
		LatencySeconds: info.Latency.Seconds(),
	}
}

func NewDeadTask(task workerConfigs.DeadTask) DeadTask {
	payload := json.RawMessage(task.Payload)
	if !json.Valid(payload) {
		encoded, _ := json.Marshal(string(task.Payload))
		payload = encoded
	}

	deadTask := DeadTask{
		TaskID:    task.ID,
		TaskType:  string(task.TaskID),
		Queue:     task.Queue,
		Payload:   payload,
		LastError: task.LastError,
		Retried:   task.Retried,
		MaxRetry:  task.MaxRetry,
	}

	if !task.LastFailedAt.IsZero() {
		deadTask.LastFailedAt = &task.LastFailedAt
	}

	return deadTask
}
