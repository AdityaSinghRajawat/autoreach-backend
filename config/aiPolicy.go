package config

import "time"

const (
	chatRoleSystem           = "system"
	chatRoleUser             = "user"
	chatFinishReasonStop     = "stop"
	jsonSchemaResponseFormat = "json_schema"
	chatCompletionsPath      = "/chat/completions"
	embeddingsPath           = "/embeddings"
	aiEmbeddingsOperation    = "embeddings"
	aiChatOperation          = "chat"
	aiCharsPerToken          = 4
	aiRetryJitter            = 5 * time.Second
)

func GetChatRoleSystem() string           { return chatRoleSystem }
func GetChatRoleUser() string             { return chatRoleUser }
func GetChatFinishReasonStop() string     { return chatFinishReasonStop }
func GetJSONSchemaResponseFormat() string { return jsonSchemaResponseFormat }
func GetChatCompletionsPath() string      { return chatCompletionsPath }
func GetEmbeddingsPath() string           { return embeddingsPath }
func GetAIEmbeddingsOperation() string    { return aiEmbeddingsOperation }
func GetAIChatOperation() string          { return aiChatOperation }
func GetAICharsPerToken() int             { return aiCharsPerToken }
func GetAIRetryJitter() time.Duration     { return aiRetryJitter }
