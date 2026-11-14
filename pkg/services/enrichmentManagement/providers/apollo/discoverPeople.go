package apollo

import (
	"context"
	"strings"

	enrichmentTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/enrichmentManagement"
	apolloIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/enrichment/apollo"
)

func (p *ApolloProvider) DiscoverPeople(
	ctx context.Context,
	query enrichmentTypes.DiscoveryQuery,
) ([]enrichmentTypes.Person, error) {
	response, err := p.integration.SearchPeople(ctx, query.CompanyDomain, query.Titles, query.Limit)
	if err != nil {
		return nil, err
	}

	people := make([]enrichmentTypes.Person, 0, len(response.People))

	for _, person := range response.People {
		if person.ID == "" || person.Title == "" {
			continue
		}

		people = append(people, toPerson(person))
	}

	return people, nil
}

func toPerson(person apolloIntegration.Person) enrichmentTypes.Person {
	companyName := ""
	if person.Organization != nil {
		companyName = person.Organization.Name
	}

	return enrichmentTypes.Person{
		ProviderPersonID: person.ID,
		FullName:         fullName(person),
		Title:            person.Title,
		LinkedinURL:      person.LinkedinURL,
		CompanyName:      companyName,
	}
}

func fullName(person apolloIntegration.Person) string {
	if strings.TrimSpace(person.Name) != "" {
		return strings.TrimSpace(person.Name)
	}

	return strings.TrimSpace(person.FirstName + " " + person.LastName)
}
