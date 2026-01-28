package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"

	"github.com/stannisl/ai-browser-assistant/internal/browser"
	"github.com/stannisl/ai-browser-assistant/internal/extractor"
	"github.com/stannisl/ai-browser-assistant/internal/llm"
	"github.com/stannisl/ai-browser-assistant/internal/logger"
	"github.com/stannisl/ai-browser-assistant/internal/types"
)

type Agent struct {
	browser   *browser.Manager
	extractor *extractor.Extractor
	llm       *llm.Client
	logger    *logger.Logger
	config    *types.AgentConfig

	messages        []openai.ChatCompletionMessage
	step            int
	lastToolName    string
	sameToolCount   int
}

func New(
	browser *browser.Manager,
	ext *extractor.Extractor,
	llmClient *llm.Client,
	log *logger.Logger,
	config *types.AgentConfig,
) *Agent {
	return &Agent{
		browser:   browser,
		extractor: ext,
		llm:       llmClient,
		logger:    log,
		config:    config,
	}
}

func (a *Agent) Run(ctx context.Context, task string) error {
	a.step = 0
	a.messages = []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: llm.SystemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: task,
		},
	}
	a.lastToolName = ""
	a.sameToolCount = 0

	for a.step < a.config.MaxSteps {
		select {
		case <-ctx.Done():
			return types.ErrContextCanceled
		default:
		}

		a.step++
		a.logger.Step(a.step, a.config.MaxSteps)

		response, err := a.llm.Chat(ctx, a.messages)
		if err != nil {
			return fmt.Errorf("llm chat: %w", err)
		}

		toolCall, hasToolCall := a.llm.ExtractToolCall(response)

		if !hasToolCall {
			if len(response.Choices) > 0 {
				a.messages = append(a.messages, response.Choices[0].Message)
			}
			continue
		}

		a.logger.Tool(toolCall.ToolName)

		if toolCall.ToolName == a.lastToolName {
			a.sameToolCount++
			if a.sameToolCount >= 3 {
				a.logger.Warn("Possible loop detected", "tool", toolCall.ToolName, "count", a.sameToolCount)
			}
		} else {
			a.lastToolName = toolCall.ToolName
			a.sameToolCount = 1
		}

		a.messages = append(a.messages, openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleAssistant,
			ToolCalls:  response.Choices[0].Message.ToolCalls,
		})

		result, err := a.executeTool(ctx, toolCall)

		toolResultContent := result
		if err != nil {
			toolResultContent = fmt.Sprintf("Error: %v", err)
		}

		a.messages = append(a.messages, openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			ToolCallID: toolCall.ID,
			Content:    toolResultContent,
		})

		if toolCall.ToolName == "report" {
			a.logger.Done(result, err == nil)
			return nil
		}
	}

	return types.ErrMaxStepsExceeded
}

func (a *Agent) executeTool(ctx context.Context, tc *types.ToolCall) (string, error) {
	return a.executeToolImpl(ctx, tc)
}

func (a *Agent) executeToolImpl(ctx context.Context, tc *types.ToolCall) (string, error) {
	switch tc.ToolName {
	case "extract_page":
		return a.extractPage(ctx)
	case "navigate":
		url, ok := tc.Arguments["url"].(string)
		if !ok {
			return "", fmt.Errorf("invalid URL argument")
		}
		err := a.browser.Navigate(ctx, url)
		return "", err
	case "click":
		id, ok := tc.Arguments["id"].(float64)
		if !ok {
			return "", fmt.Errorf("invalid id argument")
		}
		err := a.browser.Click(ctx, fmt.Sprintf("[%d]", int(id)))
		return "", err
	case "type_text":
		id, ok := tc.Arguments["id"].(float64)
		if !ok {
			return "", fmt.Errorf("invalid id argument")
		}
		text, ok := tc.Arguments["text"].(string)
		if !ok {
			return "", fmt.Errorf("invalid text argument")
		}
		err := a.browser.Type(ctx, fmt.Sprintf("[%d]", int(id)), text)
		return "", err
	case "scroll":
		direction, ok := tc.Arguments["direction"].(string)
		if !ok {
			return "", fmt.Errorf("invalid direction argument")
		}
		err := a.browser.Scroll(ctx, direction)
		return "", err
	case "wait":
		seconds, ok := tc.Arguments["seconds"].(float64)
		if !ok {
			return "", fmt.Errorf("invalid seconds argument")
		}
		time.Sleep(time.Duration(seconds) * time.Second)
		return fmt.Sprintf("Waited %d seconds", int(seconds)), nil
	case "ask_user":
		question, ok := tc.Arguments["question"].(string)
		if !ok {
			return "", fmt.Errorf("invalid question argument")
		}
		return fmt.Sprintf("Question: %s", question), nil
	case "confirm_action":
		return "Confirmation required for this action", nil
	case "report":
		result, ok := tc.Arguments["result"].(string)
		if !ok {
			result = "Task completed"
		}
		return result, nil
	default:
		return "", fmt.Errorf("unknown tool: %s", tc.ToolName)
	}
}

func (a *Agent) extractPage(ctx context.Context) (string, error) {
	pageState, err := a.extractor.Extract(ctx)
	if err != nil {
		return "", fmt.Errorf("extraction failed: %w", err)
	}

	format := a.extractor.FormatForLLM(pageState)
	return format, nil
}
