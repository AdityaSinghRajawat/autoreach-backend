package hunter

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

// Domain Finder is free: it uses no credits and no search quota.
func (i *HunterIntegration) FindDomains(
	ctx context.Context,
	companyName string,
	limit int,
) (*DomainFinderResponse, error) {
	if i.apiKey == "" {
		return nil, fmt.Errorf("%w: HUNTER_API_KEY", config.ErrMissingCredential)
	}

	query := url.Values{"company": {companyName}, "limit": {strconv.Itoa(limit)}}
	response := DomainFinderResponse{}

	err := i.httpClient.Get(
		ctx,
		"/domain-finder?"+query.Encode(),
		map[string]string{config.GetHunterAPIKeyHeader(): i.apiKey},
		&response,
	)
	if err != nil {
		return nil, fmt.Errorf("hunter find domains: %w", err)
	}

	return &response, nil
}
