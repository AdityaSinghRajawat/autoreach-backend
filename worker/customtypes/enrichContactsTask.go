package workerCustomTypes

type EnrichContactsPayload struct {
	JobID       string   `json:"jobId"       validate:"required,uuid"`
	TargetRoles []string `json:"targetRoles" validate:"omitempty,dive,oneof=ENGINEERING_MANAGER HIRING_MANAGER RECRUITER"`
}
