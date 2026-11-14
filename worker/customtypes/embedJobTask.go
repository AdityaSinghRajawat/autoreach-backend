package workerCustomTypes

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type EmbedJobPayload struct {
	JobID string `json:"jobId" validate:"required,uuid"`
}

type EmbedJobBatchPayload struct {
	JobIDs []string `json:"jobIds" validate:"required,min=1,dive,uuid"`
}

func (p *EmbedJobPayload) Validate() error {
	if err := validations.Validate(p); err != nil {
		return fmt.Errorf("embed job payload validation failed: %w", err)
	}

	return nil
}

func (p *EmbedJobBatchPayload) Validate() error {
	if err := validations.Validate(p); err != nil {
		return fmt.Errorf("embed job batch payload validation failed: %w", err)
	}

	return nil
}
