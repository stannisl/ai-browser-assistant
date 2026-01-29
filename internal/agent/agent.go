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
	browser      *browser.Manager
	extractor    *extractor.Extractor
	llm          *llm.Client
	logger       *logger.Logger
	config       *types.AgentConfig

	messages       []openai.ChatCompletionMessage
	step           int
	lastToolName   string
	sameToolCount  int
	toolCallCounter int
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
	a.toolCallCounter = 0
	a.lastToolName = ""

	for a.step < a.config.MaxSteps {
		select {
		case <-ctx.Done():
			return types.ErrContextCanceled
		default:
			time.Sleep(200 * time.Millisecond)
		}

		a.step++
		a.logger.Step(a.step, a.config.MaxSteps)

		a.trimHistory()

		response, err := a.llm.Chat(ctx, a.messages)
		if err != nil {
			return fmt.Errorf("llm chat: %w", err)
		}

		toolCall, hasToolCall := a.llm.ExtractToolCall(response)

		if !hasToolCall {
			if len(response.Choices) > 0 {
				msg := response.Choices[0].Message
				a.messages = append(a.messages, msg)
			}
			continue
		}

		a.logger.Tool(toolCall.ToolName)
		a.toolCallCounter++

		if a.toolCallCounter > 10 {
			a.logger.Warn("Possible infinite loop detected", "steps", a.toolCallCounter)
			return types.ErrMaxStepsExceeded
		}

		msg := response.Choices[0].Message

		if len(msg.ToolCalls) > 0 {
			msg.ToolCalls = nil
			a.messages = append(a.messages, msg)
		} else {
			a.messages = append(a.messages, msg)
		}

		var firstErr error
		var lastResult string

		for _, tc := range toolCall.ToolCalls {
			result, err := a.ExecuteTool(ctx, &tc)
			lastResult = result

			toolResultContent := result
			if err != nil && firstErr == nil {
				firstErr = err
				toolResultContent = fmt.Sprintf("Error: %v", err)
			}

			a.messages = append(a.messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				ToolCallID: tc.ID,
				Content:    toolResultContent,
			})

			if tc.ToolName == "report" {
				a.logger.Done(result, err == nil)
				return nil
			}
		}

		if firstErr != nil {
			a.logger.Done(lastResult, false)
			return firstErr
		}
	}

	return types.ErrMaxStepsExceeded
}

func (a *Agent) trimHistory() {
	const maxMessages = 20

	if len(a.messages) <= maxMessages {
		return
	}

	preserved := a.messages[:2]
	recent := a.messages[len(a.messages)-(maxMessages-2):]
	a.messages = append(preserved, recent...)

	a.logger.Debug("History trimmed", "from", len(a.messages)+len(recent), "to", len(a.messages))
}

func (a *Agent) ExecuteTool(ctx context.Context, tc *types.ToolCall) (string, error) {
	if len(tc.ToolCalls) > 0 {
		var results []string
		var firstErr error
		
		for _, tcc := range tc.ToolCalls {
			result, err := a.executeToolInternal(ctx, &tcc)
			results = append(results, result)
			if err != nil && firstErr == nil {
				firstErr = err
			}
		}
		
		if firstErr != nil {
			return results[0], firstErr
		}
		return results[0], nil
	}
	
	return a.executeToolInternal(ctx, tc)
}

func (a *Agent) executeToolInternal(ctx context.Context, tc *types.ToolCall) (string, error) {
	switch tc.ToolName {
	case "extract_page":
		return a.ExecuteExtractPage(ctx, tc.Arguments)
	case "navigate":
		return a.ExecuteNavigate(ctx, tc.Arguments)
	case "click":
		return a.ExecuteClick(ctx, tc.Arguments)
	case "type_text":
		return a.ExecuteTypeText(ctx, tc.Arguments)
	case "scroll":
		return a.ExecuteScroll(ctx, tc.Arguments)
	case "wait":
		return a.ExecuteWait(ctx, tc.Arguments)
	case "ask_user":
		return a.ExecuteAskUser(ctx, tc.Arguments)
	case "confirm_action":
		return a.ExecuteConfirmAction(ctx, tc.Arguments)
	case "report":
		return a.ExecuteReport(ctx, tc.Arguments)
	default:
		return "", fmt.Errorf("unknown tool: %s", tc.ToolName)
	}
}
