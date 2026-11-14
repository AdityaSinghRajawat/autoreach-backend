package apollo

import (
	"context"
	"fmt"
)

func (i *ApolloIntegration) MatchPerson(
	ctx context.Context,
	personID string,
) (*MatchResponse, error) {
	headers, err := i.authHeaders()
	if err != nil {
		return nil, err
	}

	response := MatchResponse{}

	err = i.httpClient.Post(ctx, "/people/match", headers, matchRequest{ID: personID}, &response)
	if err != nil {
		return nil, fmt.Errorf("apollo match person: %w", err)
	}

	return &response, nil
}
