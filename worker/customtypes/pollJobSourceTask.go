package workerCustomTypes

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type PollJobSourcePayload struct {
	SourceID string `json:"sourceId" validate:"required,uuid"`
}

func (p *PollJobSourcePayload) Validate() error {
	if err := validations.Validate(p); err != nil {
		return fmt.Errorf("poll job source payload validation failed: %w", err)
	}

	return nil
}
