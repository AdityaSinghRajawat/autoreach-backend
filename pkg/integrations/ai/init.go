package ai

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

	chatRequest struct {
		Model          string          `json:"model"`
		Messages       []chatMessage   `json:"messages"`
		ResponseFormat *responseFormat `json:"response_format,omitempty"`
	}

	chatMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	responseFormat struct {
		Type       string           `json:"type"`
		JSONSchema jsonSchemaFormat `json:"json_schema"`
	}

	jsonSchemaFormat struct {
		Name   string         `json:"name"`
		Strict bool           `json:"strict"`
		Schema map[string]any `json:"schema"`
	}

	chatResponse struct {
		Choices []chatChoice `json:"choices"`
	}

	chatChoice struct {
		FinishReason string `json:"finish_reason"`
		Message      struct {
			Content string  `json:"content"`
			Refusal *string `json:"refusal"`
		} `json:"message"`
	}
)

type AIIntegration struct {
	httpClient *utils.HTTPClient
	apiKey     string
	model      string
	dimensions int
	chatModel  string
}

func NewAIIntegration(baseURL string) *AIIntegration {
	return &AIIntegration{
		httpClient: utils.NewHTTPClient(baseURL, config.GetAITimeout()),
		apiKey:     config.GetAIAPIKey(),
		model:      config.GetAIEmbeddingModel(),
		dimensions: config.GetAIEmbeddingDimensions(),
		chatModel:  config.GetAIChatModel(),
	}
}

func (i *AIIntegration) Model() string {
	return i.model
}

func (i *AIIntegration) ChatModel() string {
	return i.chatModel
}

func (i *AIIntegration) authorizationHeaders() map[string]string {
	return map[string]string{config.GetAuthorizationHeader(): config.GetBearerPrefix() + i.apiKey}
}
