package apollo

import (
	"context"
	"fmt"
)

func (i *ApolloIntegration) SearchOrganizations(
	ctx context.Context,
	companyName string,
	limit int,
) (*OrganizationSearchResponse, error) {
	headers, err := i.authHeaders()
	if err != nil {
		return nil, err
	}

	response := OrganizationSearchResponse{}

	err = i.httpClient.Post(ctx, "/mixed_companies/search", headers, organizationSearchRequest{
		Name:    companyName,
		Page:    1,
		PerPage: limit,
	}, &response)
	if err != nil {
		return nil, fmt.Errorf("apollo search organizations: %w", err)
	}

	return &response, nil
}
