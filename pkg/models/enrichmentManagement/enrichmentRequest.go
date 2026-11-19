package enrichmentManagement

import (
	"fmt"
	"time"

	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type EnrichmentRequest struct {
	ID                string                        `db:"id"                  validate:"required,uuid"`
	JobID             string                        `db:"job_id"              validate:"required,uuid"`
	ContactID         string                        `db:"contact_id"          validate:"required,uuid"`
	Provider          string                        `db:"provider"            validate:"required,oneof=APOLLO SIGNALHIRE CONTACTOUT HUNTER"`
	Status            modelHelpers.EnrichmentStatus `db:"status"              validate:"required,oneof=PENDING COMPLETED FAILED"`
	ExternalRequestID string                        `db:"external_request_id" validate:"required,max=255"`
	CreditsUsed       int                           `db:"credits_used"        validate:"gte=0"`
	FailureReason     *string                       `db:"failure_reason"`
	CompletedAt       *time.Time                    `db:"completed_at"`
	CreatedAt         time.Time                     `db:"created_at"          validate:"required"`
	UpdatedAt         time.Time                     `db:"updated_at"          validate:"required"`
}

func NewEnrichmentRequest() *EnrichmentRequest {
	now := utils.GetCurrTimeStamp()

	return &EnrichmentRequest{
		ID:        utils.GenerateUUID(),
		Status:    modelHelpers.EnrichmentStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (r *EnrichmentRequest) UpdateTimestamp() {
	r.UpdatedAt = utils.GetCurrTimeStamp()
}

func (r *EnrichmentRequest) Validate() error {
	if err := validations.Validate(r); err != nil {
		return fmt.Errorf("enrichment request validation failed: %w", err)
	}

	return nil
}

func (r *EnrichmentRequest) MarkCompleted() {
	now := utils.GetCurrTimeStamp()
	r.Status = modelHelpers.EnrichmentStatusCompleted
	r.CompletedAt = &now
	r.FailureReason = nil
}

func (r *EnrichmentRequest) MarkFailed(reason string) {
	now := utils.GetCurrTimeStamp()
	r.Status = modelHelpers.EnrichmentStatusFailed
	r.CompletedAt = &now
	r.FailureReason = &reason
}
