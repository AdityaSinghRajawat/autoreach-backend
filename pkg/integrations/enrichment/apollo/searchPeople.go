package apollo

import (
	"context"
	"fmt"
)

func (i *ApolloIntegration) SearchPeople(
	ctx context.Context,
	companyDomain string,
	titles []string,
	limit int,
) (*SearchResponse, error) {
	headers, err := i.authHeaders()
	if err != nil {
		return nil, err
	}

	response := SearchResponse{}

	err = i.httpClient.Post(ctx, "/mixed_people/api_search", headers, searchRequest{
		Domains: []string{companyDomain},
		Titles:  titles,
		Page:    1,
		PerPage: limit,
	}, &response)
	if err != nil {
		return nil, fmt.Errorf("apollo search people: %w", err)
	}

	return &response, nil
}
