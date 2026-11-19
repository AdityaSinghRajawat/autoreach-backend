package hunter

import (
	"context"
	"net/url"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/enrichmentManagement"
	hunterIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/enrichment/hunter"
)

// Hunter returns each person with their email, so no separate lookup is needed.
func (p *HunterProvider) DiscoverPeople(
	ctx context.Context,
	query enrichmentTypes.DiscoveryQuery,
) ([]enrichmentTypes.Person, error) {
	response, err := p.integration.SearchDomain(ctx, query.CompanyDomain, query.Titles, query.Limit)
	if err != nil {
		return nil, err
	}

	people := make([]enrichmentTypes.Person, 0, len(response.Data.Emails))

	for _, email := range response.Data.Emails {
		firstName := strings.TrimSpace(email.FirstName)
		lastName := strings.TrimSpace(email.LastName)
		title := strings.TrimSpace(email.Position)

		if email.Value == "" || firstName == "" || lastName == "" || title == "" {
			continue
		}

		fullName := firstName + " " + lastName

		address := strings.ToLower(strings.TrimSpace(email.Value))

		people = append(people, enrichmentTypes.Person{
			Source:           config.ProviderHunter,
			ProviderPersonID: address,
			FullName:         fullName,
			Title:            title,
			LinkedinURL:      linkedinURL(email.LinkedIn),
			CompanyName:      query.CompanyName,
			Email:            address,
			EmailVerified:    isVerified(email),
		})
	}

	return people, nil
}

func isVerified(email hunterIntegration.Email) bool {
	return email.Verification != nil &&
		email.Verification.Status == config.GetHunterVerifiedEmailStatus()
}

// Hunter gives either a full URL or just the profile handle; anything else is dropped.
func linkedinURL(raw string) string {
	trimmed := strings.TrimSpace(raw)

	switch {
	case trimmed == "":
		return ""
	case strings.Contains(strings.ToLower(trimmed), config.GetLinkedinHost()):
		return enrichmentTypes.NormalizeLinkedinURL(trimmed)
	}

	handle := strings.Trim(trimmed, "/")
	if url.PathEscape(handle) != handle {
		return ""
	}

	return config.GetLinkedinProfilePath() + handle
}
