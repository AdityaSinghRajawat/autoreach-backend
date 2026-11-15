package mocks

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
)

const ChatModel = "fake-chat-model"

type ChatCompleter struct {
	mutex    sync.Mutex
	reply    string
	requests []sharedTypes.ChatCompletionRequest
}

var _ sharedTypes.ChatCompleter = (*ChatCompleter)(nil)

func (c *ChatCompleter) ChatModel() string { return ChatModel }

func (c *ChatCompleter) CreateChatCompletion(
	_ context.Context,
	request sharedTypes.ChatCompletionRequest,
) (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.requests = append(c.requests, request)
	if c.reply == "" {
		return "", config.ErrMissingCredential
	}

	return c.reply, nil
}

func (c *ChatCompleter) ReplyWith(value any) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	encoded, _ := json.Marshal(value)
	c.reply = string(encoded)
}

func (c *ChatCompleter) CallCount() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return len(c.requests)
}

func (c *ChatCompleter) Reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.reply = ""
	c.requests = nil
}
