package workerCustomTypes

type EmbedJobPayload struct {
	JobID string `json:"jobId" validate:"required,uuid"`
}

type EmbedJobBatchPayload struct {
	JobIDs []string `json:"jobIds" validate:"required,min=1,dive,uuid"`
}
