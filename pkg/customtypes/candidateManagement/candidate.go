package candidateManagement

import (
	"encoding/json"
	"strings"

	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
)

type UpdateCandidateRequest struct {
	FullName            *string         `json:"fullName"            validate:"omitnil,min=1,max=255"`
	LinkedinProfileURL  *string         `json:"linkedinProfileUrl"  validate:"omitnil,max=500,validate_linkedin_url"`
	MinMatchScore       *float64        `json:"minMatchScore"       validate:"omitnil,gte=0,lte=1"`
	AutoApplyEnabled    *bool           `json:"autoApplyEnabled"`
	AutoOutreachEnabled *bool           `json:"autoOutreachEnabled"`
	Timezone            *string         `json:"timezone"            validate:"omitnil,validate_timezone"`
	Settings            json.RawMessage `json:"settings"            validate:"omitempty,validate_json_object"`
}

func (r *UpdateCandidateRequest) IsEmpty() bool {
	return r.FullName == nil &&
		r.LinkedinProfileURL == nil &&
		r.MinMatchScore == nil &&
		r.AutoApplyEnabled == nil &&
		r.AutoOutreachEnabled == nil &&
		r.Timezone == nil &&
		len(r.Settings) == 0
}

func (r *UpdateCandidateRequest) ToCandidateModel(
	existing *candidateModel.Candidate,
) *candidateModel.Candidate {
	candidate := *existing

	if r.FullName != nil {
		candidate.FullName = strings.TrimSpace(*r.FullName)
	}

	if r.LinkedinProfileURL != nil {
		candidate.LinkedinProfileURL = nil
		if linkedinProfileURL := strings.TrimSpace(
			*r.LinkedinProfileURL,
		); linkedinProfileURL != "" {
			candidate.LinkedinProfileURL = &linkedinProfileURL
		}
	}

	if r.MinMatchScore != nil {
		candidate.MinMatchScore = *r.MinMatchScore
	}

	if r.AutoApplyEnabled != nil {
		candidate.AutoApplyEnabled = *r.AutoApplyEnabled
	}

	if r.AutoOutreachEnabled != nil {
		candidate.AutoOutreachEnabled = *r.AutoOutreachEnabled
	}

	if r.Timezone != nil {
		candidate.Timezone = *r.Timezone
	}

	if len(r.Settings) > 0 {
		candidate.Settings = r.Settings
	}

	return &candidate
}
