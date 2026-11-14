package workerCustomTypes

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type EnrichContactsPayload struct {
	JobID       string   `json:"jobId"       validate:"required,uuid"`
	TargetRoles []string `json:"targetRoles" validate:"omitempty,dive,oneof=ENGINEERING_MANAGER HIRING_MANAGER RECRUITER"`
}

func (p *EnrichContactsPayload) Validate() error {
	if err := validations.Validate(p); err != nil {
		return fmt.Errorf("enrich contacts payload validation failed: %w", err)
	}

	return nil
}
