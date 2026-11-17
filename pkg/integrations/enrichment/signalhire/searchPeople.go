package signalhire

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

// Searches count against SignalHire's daily search quota, not its contact credits.
func (i *SignalHireIntegration) SearchPeople(
	ctx context.Context,
	companyQuery, titleQuery string,
	size int,
) (*SearchByQueryResponse, error) {
	if i.apiKey == "" {
		return nil, fmt.Errorf("%w: SIGNALHIRE_API_KEY", config.ErrMissingCredential)
	}

	response := SearchByQueryResponse{}

	err := i.httpClient.Post(
		ctx,
		"/candidate/searchByQuery",
		map[string]string{config.GetSignalHireAPIKeyHeader(): i.apiKey},
		searchByQueryRequest{CurrentCompany: companyQuery, CurrentTitle: titleQuery, Size: size},
		&response,
	)
	if err != nil {
		return nil, fmt.Errorf("signalhire search people: %w", err)
	}

	return &response, nil
}
