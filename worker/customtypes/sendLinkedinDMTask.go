package workerCustomTypes

type SendLinkedinDMPayload struct {
	MessageID string `json:"messageId" validate:"required,uuid"`
}
