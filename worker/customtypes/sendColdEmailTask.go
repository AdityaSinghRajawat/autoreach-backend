package workerCustomTypes

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type SendColdEmailPayload struct {
	MessageID string `json:"messageId" validate:"required,uuid"`
}

func (p *SendColdEmailPayload) Validate() error {
	if err := validations.Validate(p); err != nil {
		return fmt.Errorf("send cold email payload validation failed: %w", err)
	}

	return nil
}
