package workerCustomTypes

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type PrepareApplicationPayload struct {
	ApplicationID string `json:"applicationId" validate:"required,uuid"`
}

func (p *PrepareApplicationPayload) Validate() error {
	if err := validations.Validate(p); err != nil {
		return fmt.Errorf("prepare application payload validation failed: %w", err)
	}

	return nil
}
