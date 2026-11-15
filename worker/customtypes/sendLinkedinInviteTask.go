package workerCustomTypes

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type SendLinkedinInvitePayload struct {
	MessageID string `json:"messageId" validate:"required,uuid"`
}

func (p *SendLinkedinInvitePayload) Validate() error {
	if err := validations.Validate(p); err != nil {
		return fmt.Errorf("send linkedin invite payload validation failed: %w", err)
	}

	return nil
}
