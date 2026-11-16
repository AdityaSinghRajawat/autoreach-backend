package workerCustomTypes

type PollJobSourcePayload struct {
	SourceID string `json:"sourceId" validate:"required,uuid"`
}
