package unit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/ai"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
)

func chatServer(
	t *testing.T,
	finishReason, content string,
	received *map[string]any,
) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			assert.Equal(t, "/chat/completions", request.URL.Path)
			assert.Equal(t, "Bearer "+config.GetAIAPIKey(), request.Header.Get("Authorization"))
			require.NoError(t, json.NewDecoder(request.Body).Decode(received))

			writer.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"choices": []map[string]any{{
					"finish_reason": finishReason,
					"message":       map[string]any{"content": content, "refusal": nil},
				}},
			})
		}),
	)
	t.Cleanup(server.Close)

	return server
}

func TestChatCompletionSendsMessagesWithAStrictSchemaAndReturnsTheReply(t *testing.T) {
	setup.Config(t)

	received := map[string]any{}
	server := chatServer(t, "stop", `{"answer":42}`, &received)

	reply, err := ai.NewAIIntegration(server.URL).CreateChatCompletion(t.Context(),
		sharedTypes.ChatCompletionRequest{
			Messages: []sharedTypes.ChatMessage{
				{Role: "system", Content: "Answer in JSON."},
				{Role: "user", Content: "question"},
			},
			ResponseSchema: &sharedTypes.ChatResponseSchema{
				Name:   "answer",
				Schema: map[string]any{"type": "object", "additionalProperties": false},
			},
		})
	require.NoError(t, err)
	assert.JSONEq(t, `{"answer":42}`, reply)

	assert.Equal(t, config.GetAIChatModel(), received["model"])
	format := received["response_format"].(map[string]any)
	assert.Equal(t, "json_schema", format["type"])
	schema := format["json_schema"].(map[string]any)
	assert.Equal(t, "answer", schema["name"])
	assert.Equal(t, true, schema["strict"])

	messages := received["messages"].([]any)
	require.Len(t, messages, 2)
	assert.Equal(t, map[string]any{"role": "user", "content": "question"}, messages[1])
}

func TestChatCompletionWithoutASchemaOmitsTheResponseFormat(t *testing.T) {
	setup.Config(t)

	received := map[string]any{}
	server := chatServer(t, "stop", "plain text", &received)

	reply, err := ai.NewAIIntegration(server.URL).CreateChatCompletion(t.Context(),
		sharedTypes.ChatCompletionRequest{
			Messages: []sharedTypes.ChatMessage{{Role: "user", Content: "hi"}},
		})
	require.NoError(t, err)
	assert.Equal(t, "plain text", reply)
	assert.NotContains(t, received, "response_format")
}

func TestChatCompletionRejectsTruncatedReplies(t *testing.T) {
	setup.Config(t)

	received := map[string]any{}
	server := chatServer(t, "length", `{"answer":`, &received)

	_, err := ai.NewAIIntegration(server.URL).CreateChatCompletion(t.Context(),
		sharedTypes.ChatCompletionRequest{
			Messages: []sharedTypes.ChatMessage{{Role: "user", Content: "hi"}},
		})
	assert.ErrorIs(t, err, config.ErrModelReplyIncomplete)
}
