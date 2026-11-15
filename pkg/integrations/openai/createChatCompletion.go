package openai

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
)

func (i *OpenAIIntegration) CreateChatCompletion(
	ctx context.Context,
	request sharedTypes.ChatCompletionRequest,
) (string, error) {
	if i.apiKey == "" {
		return "", fmt.Errorf("%w: OPENAI_API_KEY", config.ErrMissingCredential)
	}

	response := chatResponse{}

	err := i.httpClient.Post(
		ctx,
		config.GetChatCompletionsPath(),
		i.authorizationHeaders(),
		i.toChatRequest(request),
		&response,
	)
	if err != nil {
		return "", fmt.Errorf("create chat completion: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("create chat completion: %w", config.ErrModelReplyMissing)
	}

	choice := response.Choices[0]
	if choice.Message.Refusal != nil || choice.FinishReason != config.GetChatFinishReasonStop() {
		return "", fmt.Errorf(
			"create chat completion: %w: finish reason %q",
			config.ErrModelReplyIncomplete,
			choice.FinishReason,
		)
	}

	return choice.Message.Content, nil
}

func (i *OpenAIIntegration) toChatRequest(request sharedTypes.ChatCompletionRequest) chatRequest {
	wire := chatRequest{Model: i.chatModel, Messages: make([]chatMessage, 0, len(request.Messages))}

	for _, message := range request.Messages {
		wire.Messages = append(
			wire.Messages,
			chatMessage{Role: message.Role, Content: message.Content},
		)
	}

	if request.ResponseSchema != nil {
		wire.ResponseFormat = &responseFormat{
			Type: config.GetJSONSchemaResponseFormat(),
			JSONSchema: jsonSchemaFormat{
				Name:   request.ResponseSchema.Name,
				Strict: true,
				Schema: request.ResponseSchema.Schema,
			},
		}
	}

	return wire
}
