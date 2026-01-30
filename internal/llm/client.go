package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/stannisl/ai-browser-assistant/internal/logger"
	"github.com/stannisl/ai-browser-assistant/internal/types"
)

type Client struct {
	client     *openai.Client
	model      string
	logger     *logger.Logger
	maxRetries int
}

func NewClient(config *types.LLMConfig, log *logger.Logger) (*Client, error) {
	cfg := openai.DefaultConfig(config.APIKey)
	cfg.BaseURL = config.BaseURL

	client := openai.NewClientWithConfig(cfg)

	maxRetries := config.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	return &Client{
		client:     client,
		model:      config.Model,
		logger:     log,
		maxRetries: maxRetries,
	}, nil
}

func (c *Client) Chat(ctx context.Context, messages []openai.ChatCompletionMessage) (*openai.ChatCompletionResponse, error) {
	c.logger.Thinking()

	req := openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: messages,
		Tools:    GetTools(),
	}

	var resp openai.ChatCompletionResponse
	var lastErr error

	for attempt := 1; attempt <= c.maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		c.logger.Debug("request", "req", req.Messages)
		var err error
		resp, err = c.client.CreateChatCompletion(ctx, req)

		c.logger.Debug("choices by ai", "choices", resp.Choices)

		if err == nil {
			return &resp, nil
		}

		lastErr = err
		c.logger.Warn("LLM request failed, retrying...",
			"attempt", attempt,
			"max_retries", c.maxRetries,
			"error", err.Error())

		// Экспоненциальная задержка перед retry
		if attempt < c.maxRetries {
			delay := time.Duration(attempt) * 2 * time.Second
			c.logger.Debug("Waiting before retry", "delay", delay)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	return nil, fmt.Errorf("chat completion failed after %d retries: %w", c.maxRetries, lastErr)
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
