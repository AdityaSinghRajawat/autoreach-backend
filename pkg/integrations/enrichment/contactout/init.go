package contactout

import (
	"bytes"
	"encoding/json"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type (
	EnrichResponse struct {
		StatusCode int      `json:"status_code"`
		Profile    *Profile `json:"profile"`
	}

	Profile struct {
		URL           string   `json:"url"`
		FullName      string   `json:"full_name"`
		Headline      string   `json:"headline"`
		Email         []string `json:"email"`
		WorkEmail     []string `json:"work_email"`
		PersonalEmail []string `json:"personal_email"`
		Phone         []string `json:"phone"`
	}
)

type ContactOutIntegration struct {
	httpClient *utils.HTTPClient
	apiKey     string
}

func NewContactOutIntegration(policy config.ProviderPolicy) *ContactOutIntegration {
	return &ContactOutIntegration{
		httpClient: utils.NewHTTPClient(policy.BaseURL, policy.Timeout),
		apiKey:     policy.APIKey,
	}
}

// ContactOut answers an unknown profile with "profile": [] instead of an object, meaning not found.
func (r *EnrichResponse) UnmarshalJSON(data []byte) error {
	payload := struct {
		StatusCode int             `json:"status_code"`
		Profile    json.RawMessage `json:"profile"`
	}{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	r.StatusCode = payload.StatusCode
	r.Profile = nil

	profile := bytes.TrimSpace(payload.Profile)
	if len(profile) == 0 || profile[0] != '{' {
		return nil
	}

	r.Profile = &Profile{}

	return json.Unmarshal(profile, r.Profile)
}
