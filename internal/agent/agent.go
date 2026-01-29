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

	messages      []openai.ChatCompletionMessage
	step          int
	lastToolName  string
	lastToolArgs  string
	sameToolCount int
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
	a.lastToolArgs = ""
	a.sameToolCount = 0

	for a.step < a.config.MaxSteps {
		select {
		case <-ctx.Done():
			return types.ErrContextCanceled
		default:
		}

		a.step++
		a.logger.Step(a.step, a.config.MaxSteps)

		// Обрезаем историю если слишком длинная
		a.trimHistory()

		// Запрос к LLM
		response, err := a.llm.Chat(ctx, a.messages)
		if err != nil {
			return fmt.Errorf("llm chat: %w", err)
		}

		// Извлекаем tool call
		toolCall, hasToolCall := a.llm.ExtractToolCall(response)

		if !hasToolCall {
			// LLM ответил текстом без tool call
			if len(response.Choices) > 0 {
				msg := response.Choices[0].Message
				a.messages = append(a.messages, msg)

				// Просим продолжить
				a.messages = append(a.messages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: "Continue. Use extract_page to see the page, or report if done.",
				})
			}
			continue
		}

		a.logger.Tool(toolCall.ToolName)

		// Проверка на loop
		if a.detectLoop(toolCall) {
			a.messages = append(a.messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: "You seem stuck repeating the same action. Try extract_page to refresh, or try a different approach.",
			})
		}

		// Добавляем assistant message с tool call
		assistantMsg := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: response.Choices[0].Message.Content,
			ToolCalls: []openai.ToolCall{
				{
					ID:   toolCall.ID,
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      toolCall.ToolName,
						Arguments: response.Choices[0].Message.ToolCalls[0].Function.Arguments,
					},
				},
			},
		}
		a.messages = append(a.messages, assistantMsg)

		// Выполняем tool
		result, err := a.ExecuteTool(ctx, toolCall)

		toolResultContent := result
		if err != nil {
			toolResultContent = fmt.Sprintf("Error: %v", err)
		}

		// Добавляем результат tool
		a.messages = append(a.messages, openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			ToolCallID: toolCall.ID,
			Content:    toolResultContent,
		})

		// Если report — завершаем
		if toolCall.ToolName == "report" {
			return nil
		}

		// Небольшая пауза между шагами
		time.Sleep(200 * time.Millisecond)
	}

	return types.ErrMaxStepsExceeded
}

func (a *Agent) detectLoop(tc *types.ToolCall) bool {
	argsStr := fmt.Sprintf("%v", tc.Arguments)

	if tc.ToolName == a.lastToolName && argsStr == a.lastToolArgs {
		a.sameToolCount++
	} else {
		a.lastToolName = tc.ToolName
		a.lastToolArgs = argsStr
		a.sameToolCount = 1
	}

	return a.sameToolCount >= 3
}

func (a *Agent) trimHistory() {
	const maxMessages = 20

	if len(a.messages) <= maxMessages {
		return
	}

	// Сохраняем system + первый user (индексы 0, 1)
	preserved := a.messages[:2]

	// Берём последние сообщения
	tailSize := maxMessages - 2
	tail := a.messages[len(a.messages)-tailSize:]

	// Проверяем что tail не начинается с Tool message (нужен parent Assistant)
	for len(tail) > 0 && tail[0].Role == openai.ChatMessageRoleTool {
		// Пропускаем orphan Tool messages
		tail = tail[1:]
	}

	a.messages = append(preserved, tail...)
	a.logger.Debug("History trimmed", "new_len", len(a.messages))
}
