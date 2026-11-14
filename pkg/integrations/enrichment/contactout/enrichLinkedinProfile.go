package contactout

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (i *ContactOutIntegration) EnrichLinkedinProfile(
	ctx context.Context,
	linkedinURL string,
) (*EnrichResponse, error) {
	if i.apiKey == "" {
		return nil, fmt.Errorf("%w: CONTACTOUT_API_KEY", config.ErrMissingCredential)
	}

	response := EnrichResponse{}

	err := i.httpClient.Get(
		ctx,
		"/linkedin/enrich?profile="+url.QueryEscape(linkedinURL),
		map[string]string{config.GetContactOutTokenHeader(): i.apiKey},
		&response,
	)

	var apiErr *utils.APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
		return &EnrichResponse{StatusCode: http.StatusNotFound}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("contactout enrich linkedin profile: %w", err)
	}

	return &response, nil
}
