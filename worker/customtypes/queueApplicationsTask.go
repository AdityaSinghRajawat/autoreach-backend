package workerCustomTypes

type QueueApplicationsPayload struct {
	CandidateID string `json:"candidateId" validate:"required,uuid"`
}
