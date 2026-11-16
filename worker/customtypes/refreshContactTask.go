package workerCustomTypes

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type RefreshContactPayload struct {
	ContactID string `json:"contactId" validate:"required,uuid"`
}

func (p *RefreshContactPayload) Validate() error {
	if err := validations.Validate(p); err != nil {
		return fmt.Errorf("refresh contact payload validation failed: %w", err)
	}

	return nil
}
