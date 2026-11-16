package outreachManagement

import "strings"

// ContactLookupParams finds the records held about a person, for example before erasing them.
type ContactLookupParams struct {
	Email       string `form:"email"       validate:"omitempty,email,max=255"`
	LinkedinURL string `form:"linkedinUrl" validate:"omitempty,validate_linkedin_url"`
}

func (p *ContactLookupParams) IsEmpty() bool {
	return strings.TrimSpace(p.Email) == "" && strings.TrimSpace(p.LinkedinURL) == ""
}

func (p *ContactLookupParams) NormalizedEmail() string {
	return strings.ToLower(strings.TrimSpace(p.Email))
}

func (p *ContactLookupParams) NormalizedLinkedinURL() string {
	return strings.TrimSpace(p.LinkedinURL)
}
