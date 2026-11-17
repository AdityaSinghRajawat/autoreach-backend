package apollo

import (
	"context"
	"encoding/json"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/enrichmentManagement"
)

func (p *ApolloProvider) FindContact(
	ctx context.Context,
	person enrichmentTypes.Person,
) (*enrichmentTypes.ContactDetails, error) {
	if person.Source != config.ProviderApollo || person.ProviderPersonID == "" {
		return nil, nil
	}

	response, err := p.integration.MatchPerson(ctx, person.ProviderPersonID)
	if err != nil {
		return nil, err
	}

	if response.Person == nil {
		return nil, nil
	}

	raw, err := json.Marshal(response.Person)
	if err != nil {
		raw = nil
	}

	return &enrichmentTypes.ContactDetails{
		FullName:      fullName(*response.Person),
		Title:         response.Person.Title,
		Email:         response.Person.Email,
		EmailVerified: response.Person.EmailStatus == config.GetApolloVerifiedEmailStatus(),
		LinkedinURL:   response.Person.LinkedinURL,
		Raw:           raw,
	}, nil
}
