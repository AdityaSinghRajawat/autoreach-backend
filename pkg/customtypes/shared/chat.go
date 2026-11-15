package shared

type ChatMessage struct {
	Role    string
	Content string
}

type ChatResponseSchema struct {
	Name   string
	Schema map[string]any
}

type ChatCompletionRequest struct {
	Messages       []ChatMessage
	ResponseSchema *ChatResponseSchema
}
