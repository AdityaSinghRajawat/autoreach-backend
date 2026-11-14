package contactout

import (
	"context"
	"encoding/json"

	enrichmentTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/enrichmentManagement"
)

func (p *ContactOutProvider) FindContact(
	ctx context.Context,
	person enrichmentTypes.Person,
) (*enrichmentTypes.ContactDetails, error) {
	if !person.HasLinkedin() {
		return nil, nil
	}

	response, err := p.integration.EnrichLinkedinProfile(ctx, person.LinkedinURL)
	if err != nil {
		return nil, err
	}

	if response.Profile == nil {
		return nil, nil
	}

	raw, err := json.Marshal(response.Profile)
	if err != nil {
		raw = nil
	}

	details := &enrichmentTypes.ContactDetails{
		FullName:    response.Profile.FullName,
		Title:       response.Profile.Headline,
		LinkedinURL: response.Profile.URL,
		Raw:         raw,
	}

	switch {
	case len(response.Profile.WorkEmail) > 0:
		details.Email = response.Profile.WorkEmail[0]
		details.EmailVerified = true
	case len(response.Profile.Email) > 0:
		details.Email = response.Profile.Email[0]
	}

	if len(response.Profile.Phone) > 0 {
		details.Phone = response.Profile.Phone[0]
	}

	return details, nil
}
