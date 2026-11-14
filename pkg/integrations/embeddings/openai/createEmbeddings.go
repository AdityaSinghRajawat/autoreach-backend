package openai

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

func (i *OpenAIIntegration) CreateEmbeddings(
	ctx context.Context,
	inputs []string,
) ([][]float32, error) {
	if i.apiKey == "" {
		return nil, fmt.Errorf("%w: OPENAI_API_KEY", config.ErrMissingCredential)
	}

	if len(inputs) == 0 {
		return nil, nil
	}

	response := embeddingsResponse{}

	err := i.httpClient.Post(
		ctx,
		"/embeddings",
		map[string]string{config.GetAuthorizationHeader(): config.GetBearerPrefix() + i.apiKey},
		embeddingsRequest{Model: i.model, Input: inputs, Dimensions: i.dimensions},
		&response,
	)
	if err != nil {
		return nil, fmt.Errorf("create embeddings: %w", err)
	}

	if len(response.Data) != len(inputs) {
		return nil, fmt.Errorf(
			"create embeddings: expected %d vectors, received %d",
			len(inputs),
			len(response.Data),
		)
	}

	embeddings := make([][]float32, len(inputs))
	for _, item := range response.Data {
		if item.Index < 0 || item.Index >= len(inputs) || len(item.Embedding) != i.dimensions {
			return nil, fmt.Errorf("create embeddings: malformed vector at index %d", item.Index)
		}

		embeddings[item.Index] = item.Embedding
	}

	for index, embedding := range embeddings {
		if embedding == nil {
			return nil, fmt.Errorf("create embeddings: missing vector at index %d", index)
		}
	}

	return embeddings, nil
}
