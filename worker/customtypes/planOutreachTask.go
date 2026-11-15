package workerCustomTypes

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type PlanOutreachPayload struct {
	CandidateID string `json:"candidateId" validate:"required,uuid"`
	JobID       string `json:"jobId"       validate:"required,uuid"`
}

func (p *PlanOutreachPayload) Validate() error {
	if err := validations.Validate(p); err != nil {
		return fmt.Errorf("plan outreach payload validation failed: %w", err)
	}

	return nil
}
