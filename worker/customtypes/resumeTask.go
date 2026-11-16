package workerCustomTypes

type ProcessResumePayload struct {
	ResumeID string `json:"resumeId" validate:"required,uuid"`
}

type MatchResumePayload struct {
	ResumeID string `json:"resumeId" validate:"required,uuid"`
}

type MatchJobPayload struct {
	JobID string `json:"jobId" validate:"required,uuid"`
}
