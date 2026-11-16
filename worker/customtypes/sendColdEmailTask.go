package workerCustomTypes

type SendColdEmailPayload struct {
	MessageID string `json:"messageId" validate:"required,uuid"`
}
