package workerCustomTypes

type PrepareApplicationPayload struct {
	ApplicationID string `json:"applicationId" validate:"required,uuid"`
}
