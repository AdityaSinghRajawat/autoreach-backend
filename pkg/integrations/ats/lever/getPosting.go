package lever

import (
	"context"
	"fmt"
)

func (i *LeverIntegration) GetPosting(ctx context.Context, site, postingID string) (string, error) {
	response := postingResponse{}
	if err := i.httpClient.Get(ctx, postingPath(site, postingID), nil, &response); err != nil {
		return "", fmt.Errorf("get lever posting: %w", err)
	}

	return response.ID, nil
}
