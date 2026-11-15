package config

const (
	chatRoleSystem           = "system"
	chatRoleUser             = "user"
	chatFinishReasonStop     = "stop"
	jsonSchemaResponseFormat = "json_schema"
	chatCompletionsPath      = "/chat/completions"
	embeddingsPath           = "/embeddings"
)

func GetChatRoleSystem() string           { return chatRoleSystem }
func GetChatRoleUser() string             { return chatRoleUser }
func GetChatFinishReasonStop() string     { return chatFinishReasonStop }
func GetJSONSchemaResponseFormat() string { return jsonSchemaResponseFormat }
func GetChatCompletionsPath() string      { return chatCompletionsPath }
func GetEmbeddingsPath() string           { return embeddingsPath }
