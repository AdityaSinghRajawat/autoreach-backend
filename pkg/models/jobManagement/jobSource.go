package jobManagement

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type JobSource struct {
	ID            string                       `db:"id"             json:"sourceId"      validate:"required,uuid"`
	ATSType       string                       `db:"ats_type"       json:"atsType"       validate:"required,oneof=GREENHOUSE LEVER ASHBY SMARTRECRUITERS"`
	BoardToken    string                       `db:"board_token"    json:"boardToken"    validate:"required,max=100,validate_board_token"`
	Region        modelHelpers.JobSourceRegion `db:"region"         json:"region"        validate:"required,oneof=GLOBAL EU"`
	CompanyName   string                       `db:"company_name"   json:"companyName"   validate:"required,max=255"`
	CompanyDomain string                       `db:"company_domain" json:"companyDomain" validate:"required,fqdn,max=255"`
	IsActive      bool                         `db:"is_active"      json:"isActive"`
	LastPolledAt  *time.Time                   `db:"last_polled_at" json:"lastPolledAt"`
	LastJobCount  *int                         `db:"last_job_count" json:"lastJobCount"`
	LastError     *string                      `db:"last_error"     json:"lastError"`
	CreatedAt     time.Time                    `db:"created_at"     json:"createdAt"     validate:"required"`
	UpdatedAt     time.Time                    `db:"updated_at"     json:"updatedAt"     validate:"required"`
}

func NewJobSource() *JobSource {
	now := utils.GetCurrTimeStamp()

	return &JobSource{
		ID:        uuid.NewString(),
		Region:    modelHelpers.JobSourceRegionGlobal,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (s *JobSource) UpdateTimestamp() {
	s.UpdatedAt = utils.GetCurrTimeStamp()
}

func (s *JobSource) Validate() error {
	if err := validations.Validate(s); err != nil {
		return fmt.Errorf("job source validation failed: %w", err)
	}

	return nil
}

func (s *JobSource) RecordPoll(jobCount int, pollErr error) {
	now := utils.GetCurrTimeStamp()
	s.LastPolledAt = &now

	if pollErr != nil {
		reason := pollErr.Error()
		if limit := config.GetMaxJobSourceErrorLength(); len(reason) > limit {
			reason = reason[:limit]
		}

		s.LastError = &reason

		return
	}

	s.LastJobCount = &jobCount
	s.LastError = nil
}
