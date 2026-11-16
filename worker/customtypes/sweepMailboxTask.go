package workerCustomTypes

type SweepMailboxPayload struct {
	AccountID string `json:"accountId" validate:"required,uuid"`
}
