package workerCustomTypes

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type SendLinkedinDMPayload struct {
	MessageID string `json:"messageId" validate:"required,uuid"`
}

func (p *SendLinkedinDMPayload) Validate() error {
	if err := validations.Validate(p); err != nil {
		return fmt.Errorf("send linkedin dm payload validation failed: %w", err)
	}

	return nil
}
