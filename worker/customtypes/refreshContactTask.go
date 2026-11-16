package workerCustomTypes

type RefreshContactPayload struct {
	ContactID string `json:"contactId" validate:"required,uuid"`
}
