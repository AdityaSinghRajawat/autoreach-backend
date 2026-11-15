package outreachManagement

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type Template struct {
	ID          string                       `db:"id"           json:"templateId"  validate:"required,uuid"`
	CandidateID string                       `db:"candidate_id" json:"candidateId" validate:"required,uuid"`
	Name        string                       `db:"name"         json:"name"        validate:"required,max=255"`
	Channel     modelHelpers.OutreachChannel `db:"channel"      json:"channel"     validate:"required,oneof=EMAIL LINKEDIN_CONNECT LINKEDIN_DM"`
	Subject     *string                      `db:"subject"      json:"subject"     validate:"omitnil,max=500"`
	Body        string                       `db:"body"         json:"body"        validate:"required,max=20000"`
	Status      modelHelpers.TemplateStatus  `db:"status"       json:"status"      validate:"required,oneof=DRAFT APPROVED ARCHIVED"`
	ApprovedAt  *time.Time                   `db:"approved_at"  json:"approvedAt"`
	CreatedAt   time.Time                    `db:"created_at"   json:"createdAt"   validate:"required"`
	UpdatedAt   time.Time                    `db:"updated_at"   json:"updatedAt"   validate:"required"`
}

func NewTemplate() *Template {
	now := utils.GetCurrTimeStamp()

	return &Template{
		ID:        uuid.NewString(),
		Status:    modelHelpers.TemplateStatusDraft,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (t *Template) UpdateTimestamp() {
	t.UpdatedAt = utils.GetCurrTimeStamp()
}

func (t *Template) Validate() error {
	if err := validations.Validate(t); err != nil {
		return fmt.Errorf("outreach template validation failed: %w", err)
	}

	return nil
}

func (t *Template) IsApproved() bool {
	return t.Status == modelHelpers.TemplateStatusApproved
}

func (t *Template) SetStatus(status modelHelpers.TemplateStatus) {
	t.Status = status
	t.ApprovedAt = nil

	if status == modelHelpers.TemplateStatusApproved {
		now := utils.GetCurrTimeStamp()
		t.ApprovedAt = &now
	}
}
