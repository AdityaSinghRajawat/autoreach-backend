package openai

import (
	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type (
	embeddingsRequest struct {
		Model      string   `json:"model"`
		Input      []string `json:"input"`
		Dimensions int      `json:"dimensions"`
	}

	embeddingsResponse struct {
		Data []embeddingItem `json:"data"`
	}

	embeddingItem struct {
		Index     int       `json:"index"`
		Embedding []float32 `json:"embedding"`
	}
)

type OpenAIIntegration struct {
	httpClient *utils.HttpClient
	apiKey     string
	model      string
	dimensions int
}

func NewOpenAIIntegration() *OpenAIIntegration {
	return &OpenAIIntegration{
		httpClient: utils.NewHttpClient(config.GetOpenAIBaseURL(), config.GetOpenAITimeout()),
		apiKey:     config.GetOpenAIAPIKey(),
		model:      config.GetOpenAIEmbeddingModel(),
		dimensions: config.GetOpenAIEmbeddingDimensions(),
	}
}

func (i *OpenAIIntegration) Model() string {
	return i.model
}
