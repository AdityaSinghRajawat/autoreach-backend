package workerCustomTypes

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type ProcessResumePayload struct {
	ResumeID string `json:"resumeId" validate:"required,uuid"`
}

type MatchResumePayload struct {
	ResumeID string `json:"resumeId" validate:"required,uuid"`
}

type MatchJobPayload struct {
	JobID string `json:"jobId" validate:"required,uuid"`
}

func (p *ProcessResumePayload) Validate() error {
	if err := validations.Validate(p); err != nil {
		return fmt.Errorf("process resume payload validation failed: %w", err)
	}

	return nil
}

func (p *MatchResumePayload) Validate() error {
	if err := validations.Validate(p); err != nil {
		return fmt.Errorf("match resume payload validation failed: %w", err)
	}

	return nil
}

func (p *MatchJobPayload) Validate() error {
	if err := validations.Validate(p); err != nil {
		return fmt.Errorf("match job payload validation failed: %w", err)
	}

	return nil
}
