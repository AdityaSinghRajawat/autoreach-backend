package outreachManagement

import (
	"fmt"
	"time"

	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type Campaign struct {
	ID               string                      `db:"id"                 json:"campaignId"       validate:"required,uuid"`
	CandidateID      string                      `db:"candidate_id"       json:"candidateId"      validate:"required,uuid"`
	ChannelAccountID *string                     `db:"channel_account_id" json:"channelAccountId" validate:"omitnil,uuid"`
	Name             string                      `db:"name"               json:"name"             validate:"required,max=255"`
	Status           modelHelpers.CampaignStatus `db:"status"             json:"status"           validate:"required,oneof=ACTIVE PAUSED ARCHIVED"`
	CreatedAt        time.Time                   `db:"created_at"         json:"createdAt"        validate:"required"`
	UpdatedAt        time.Time                   `db:"updated_at"         json:"updatedAt"        validate:"required"`
}

type CampaignStep struct {
	CampaignID string `db:"campaign_id" json:"-"          validate:"required,uuid"`
	StepNumber int    `db:"step_number" json:"stepNumber" validate:"gt=0"`
	TemplateID string `db:"template_id" json:"templateId" validate:"required,uuid"`
	DelayDays  int    `db:"delay_days"  json:"delayDays"  validate:"gte=0,lte=60"`
}

func NewCampaign() *Campaign {
	now := utils.GetCurrTimeStamp()

	return &Campaign{
		ID:        utils.GenerateUUID(),
		Status:    modelHelpers.CampaignStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (c *Campaign) UpdateTimestamp() {
	c.UpdatedAt = utils.GetCurrTimeStamp()
}

func (c *Campaign) Validate() error {
	if err := validations.Validate(c); err != nil {
		return fmt.Errorf("outreach campaign validation failed: %w", err)
	}

	return nil
}

func (s *CampaignStep) Validate() error {
	if err := validations.Validate(s); err != nil {
		return fmt.Errorf("campaign step validation failed: %w", err)
	}

	return nil
}
