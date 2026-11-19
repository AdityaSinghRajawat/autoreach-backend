package hunter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// Hunter charges one credit for every email it returns.
func (i *HunterIntegration) SearchDomain(
	ctx context.Context,
	companyDomain string,
	jobTitles []string,
	limit int,
) (*SearchResponse, error) {
	if i.apiKey == "" {
		return nil, fmt.Errorf("%w: HUNTER_API_KEY", config.ErrMissingCredential)
	}

	query := url.Values{
		"domain":         {companyDomain},
		"limit":          {strconv.Itoa(limit)},
		"type":           {"personal"},
		"required_field": {"full_name,position"},
	}
	if len(jobTitles) > 0 {
		query.Set("job_titles", strings.Join(jobTitles, ","))
	}

	response := SearchResponse{}

	err := i.httpClient.Get(
		ctx,
		"/domain-search?"+query.Encode(),
		map[string]string{config.GetHunterAPIKeyHeader(): i.apiKey},
		&response,
	)

	// 451 means the people asked not to be processed: nobody is returned.
	var apiErr *utils.APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusUnavailableForLegalReasons {
		return &SearchResponse{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("hunter search domain: %w", err)
	}

	return &response, nil
}
