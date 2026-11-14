package signalhire

import (
	"context"
	"fmt"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

func (i *SignalHireIntegration) RequestPerson(
	ctx context.Context,
	identifier, callbackURL string,
) (string, error) {
	if i.apiKey == "" {
		return "", fmt.Errorf("%w: SIGNALHIRE_API_KEY", config.ErrMissingCredential)
	}

	response := SearchResponse{}

	err := i.httpClient.Post(
		ctx,
		"/candidate/search",
		map[string]string{config.GetSignalHireAPIKeyHeader(): i.apiKey},
		searchRequest{Items: []string{identifier}, CallbackURL: callbackURL},
		&response,
	)
	if err != nil {
		return "", fmt.Errorf("signalhire request person: %w", err)
	}

	requestID := strings.Trim(strings.TrimSpace(string(response.RequestID)), `"`)
	if requestID == "" || requestID == "null" {
		return "", fmt.Errorf("signalhire request person: %w", config.ErrMalformedProviderPayload)
	}

	return requestID, nil
}
