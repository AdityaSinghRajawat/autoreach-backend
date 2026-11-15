package helpers

type ApplicationStatus string

const (
	ApplicationStatusQueued         ApplicationStatus = "QUEUED"
	ApplicationStatusInProgress     ApplicationStatus = "IN_PROGRESS"
	ApplicationStatusSubmitted      ApplicationStatus = "SUBMITTED"
	ApplicationStatusActionRequired ApplicationStatus = "ACTION_REQUIRED"
	ApplicationStatusFailed         ApplicationStatus = "FAILED"
	ApplicationStatusInterviewing   ApplicationStatus = "INTERVIEWING"
	ApplicationStatusRejected       ApplicationStatus = "REJECTED"
)
