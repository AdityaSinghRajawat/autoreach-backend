package workerCustomTypes

type SendLinkedinInvitePayload struct {
	MessageID string `json:"messageId" validate:"required,uuid"`
}
