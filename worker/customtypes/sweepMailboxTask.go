package workerCustomTypes

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type SweepMailboxPayload struct {
	AccountID string `json:"accountId" validate:"required,uuid"`
}

func (p *SweepMailboxPayload) Validate() error {
	if err := validations.Validate(p); err != nil {
		return fmt.Errorf("sweep mailbox payload validation failed: %w", err)
	}

	return nil
}
