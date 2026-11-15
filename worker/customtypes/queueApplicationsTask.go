package workerCustomTypes

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type QueueApplicationsPayload struct {
	CandidateID string `json:"candidateId" validate:"required,uuid"`
}

func (p *QueueApplicationsPayload) Validate() error {
	if err := validations.Validate(p); err != nil {
		return fmt.Errorf("queue applications payload validation failed: %w", err)
	}

	return nil
}
