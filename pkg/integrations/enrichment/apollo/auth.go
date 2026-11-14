package apollo

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

func (i *ApolloIntegration) authHeaders() (map[string]string, error) {
	if i.apiKey == "" {
		return nil, fmt.Errorf("%w: APOLLO_API_KEY", config.ErrMissingCredential)
	}

	return map[string]string{
		config.GetApolloAPIKeyHeader(): i.apiKey,
		config.GetCacheControlHeader(): config.GetNoCacheValue(),
	}, nil
}
