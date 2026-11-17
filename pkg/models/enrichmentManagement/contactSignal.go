package enrichmentManagement

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type ContactSignal struct {
	ID          string                         `db:"id" json:"signalId" validate:"required,uuid"`
	ContactID   string                         `db:"contact_id" json:"contactId" validate:"required,uuid"`
	CandidateID string                         `db:"candidate_id" json:"candidateId" validate:"required,uuid"`
	SignalType  modelHelpers.ContactSignalType `db:"signal_type" json:"signalType" validate:"required,oneof=BOUNCED REPLIED LINKEDIN_RESPONDED"`
	Address     string                         `db:"address" json:"address" validate:"omitempty,email,max=255"`
	OccurredAt  time.Time                      `db:"occurred_at" json:"occurredAt" validate:"required"`
	AppliedAt   *time.Time                     `db:"applied_at" json:"appliedAt"`
	CreatedAt   time.Time                      `db:"created_at" json:"createdAt" validate:"required"`
	UpdatedAt   time.Time                      `db:"updated_at" json:"updatedAt" validate:"required"`
}

func NewContactSignal() *ContactSignal {
	now := utils.GetCurrTimeStamp()

	return &ContactSignal{ID: uuid.NewString(), OccurredAt: now, CreatedAt: now, UpdatedAt: now}
}

func (s *ContactSignal) UpdateTimestamp() {
	s.UpdatedAt = utils.GetCurrTimeStamp()
}

func (s *ContactSignal) Validate() error {
	if err := validations.Validate(s); err != nil {
		return fmt.Errorf("contact signal validation failed: %w", err)
	}

	return nil
}
