package candidateManagement

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

const (
	defaultTimezone = "UTC"
	emptySettings   = `{}`
)

type Candidate struct {
	ID                  string          `db:"id"                    json:"candidateId"         validate:"required,uuid"`
	UserID              string          `db:"user_id"               json:"userId"              validate:"required,uuid"`
	Email               string          `db:"email"                 json:"email"               validate:"required,email,max=255"`
	FullName            string          `db:"full_name"             json:"fullName"            validate:"required,max=255"`
	LinkedinProfileURL  *string         `db:"linkedin_profile_url"  json:"linkedinProfileUrl"  validate:"omitnil,max=500,validate_linkedin_url"`
	MinMatchScore       float64         `db:"min_match_score"       json:"minMatchScore"       validate:"gte=0,lte=1"`
	AutoApplyEnabled    bool            `db:"auto_apply_enabled"    json:"autoApplyEnabled"`
	AutoOutreachEnabled bool            `db:"auto_outreach_enabled" json:"autoOutreachEnabled"`
	Timezone            string          `db:"timezone"              json:"timezone"            validate:"required,validate_timezone"`
	Settings            json.RawMessage `db:"settings"              json:"settings"            validate:"required,validate_json_object"`
	CreatedAt           time.Time       `db:"created_at"            json:"createdAt"           validate:"required"`
	UpdatedAt           time.Time       `db:"updated_at"            json:"updatedAt"           validate:"required"`
}

func NewCandidate() *Candidate {
	now := utils.GetCurrTimeStamp()

	return &Candidate{
		ID:                  uuid.NewString(),
		MinMatchScore:       config.GetDefaultMinMatchScore(),
		AutoApplyEnabled:    true,
		AutoOutreachEnabled: true,
		Timezone:            defaultTimezone,
		Settings:            json.RawMessage(emptySettings),
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

func (c *Candidate) UpdateTimestamp() {
	c.UpdatedAt = utils.GetCurrTimeStamp()
}

func (c *Candidate) Validate() error {
	if err := validations.Validate(c); err != nil {
		return fmt.Errorf("candidate validation failed: %w", err)
	}

	return nil
}
