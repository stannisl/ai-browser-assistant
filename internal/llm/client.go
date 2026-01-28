package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/stannisl/ai-browser-assistant/internal/logger"
	"github.com/stannisl/ai-browser-assistant/internal/types"
)

type Client struct {
	client *openai.Client
	model  string
	logger *logger.Logger
}

func NewClient(config *types.LLMConfig, log *logger.Logger) (*Client, error) {
	cfg := openai.DefaultConfig(config.APIKey)
	cfg.BaseURL = config.BaseURL

	client := openai.NewClientWithConfig(cfg)

	return &Client{
		client: client,
		model:  config.Model,
		logger: log,
	}, nil
}

func (c *Client) Chat(ctx context.Context, messages []openai.ChatCompletionMessage) (*openai.ChatCompletionResponse, error) {
	c.logger.Thinking()

	req := openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: messages,
		Tools:    GetTools(),
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("chat completion failed: %w", err)
	}

	return &resp, nil
}

func (c *Client) ExtractToolCall(response *openai.ChatCompletionResponse) (*types.ToolCall, bool) {
	if len(response.Choices) == 0 {
		return nil, false
	}

	choice := response.Choices[0]

	if len(choice.Message.ToolCalls) == 0 {
		return nil, false
	}

	tc := choice.Message.ToolCalls[0]

	var args map[string]interface{}
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		args = map[string]interface{}{}
	}

	return &types.ToolCall{
		ID:        tc.ID,
		ToolName:  tc.Function.Name,
		Arguments: args,
	}, true
}

func (c *Client) GetModel() string {
	return c.model
}
