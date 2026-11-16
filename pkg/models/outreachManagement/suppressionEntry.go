package outreachManagement

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type SuppressionEntry struct {
	ID          string                         `db:"id" json:"suppressionId" validate:"required,uuid"`
	CandidateID *string                        `db:"candidate_id" json:"candidateId" validate:"omitnil,uuid"`
	EntryType   modelHelpers.SuppressionType   `db:"entry_type" json:"entryType" validate:"required,oneof=EMAIL DOMAIN"`
	Value       string                         `db:"value" json:"value" validate:"required,max=255"`
	Reason      modelHelpers.SuppressionReason `db:"reason" json:"reason" validate:"required,oneof=BOUNCED UNSUBSCRIBED COMPLAINT MANUAL"`
	CreatedAt   time.Time                      `db:"created_at" json:"createdAt" validate:"required"`
}

func NewSuppressionEntry() *SuppressionEntry {
	return &SuppressionEntry{ID: uuid.NewString(), CreatedAt: utils.GetCurrTimeStamp()}
}

func (s *SuppressionEntry) Validate() error {
	if err := validations.Validate(s); err != nil {
		return fmt.Errorf("suppression entry validation failed: %w", err)
	}

	return nil
}
