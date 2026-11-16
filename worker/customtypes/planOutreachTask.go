package workerCustomTypes

type PlanOutreachPayload struct {
	CandidateID string `json:"candidateId" validate:"required,uuid"`
	JobID       string `json:"jobId"       validate:"required,uuid"`
}
