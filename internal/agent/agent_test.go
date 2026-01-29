package agent

import (
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

func TestValidateMessagePairs(t *testing.T) {
	tests := []struct {
		name      string
		messages  []openai.ChatCompletionMessage
		wantCount int
	}{
		{
			name: "all messages valid",
			messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: "system"},
				{Role: openai.ChatMessageRoleUser, Content: "user"},
				{
					Role: openai.ChatMessageRoleAssistant,
					ToolCalls: []openai.ToolCall{
						{ID: "call_1", Type: "function", Function: openai.FunctionCall{Name: "test", Arguments: "{}"}},
					},
				},
				{
					Role:       openai.ChatMessageRoleTool,
					ToolCallID: "call_1",
					Content:    "result",
				},
			},
			wantCount: 4,
		},
		{
			name: "orphaned tool message without parent",
			messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: "system"},
				{
					Role: openai.ChatMessageRoleAssistant,
					ToolCalls: []openai.ToolCall{
						{ID: "call_1", Type: "function", Function: openai.FunctionCall{Name: "test", Arguments: "{}"}},
					},
				},
				{
					Role:       openai.ChatMessageRoleTool,
					ToolCallID: "call_2",
					Content:    "orphan result",
				},
			},
			wantCount: 2,
		},
		{
			name: "multiple orphaned tool messages",
			messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: "system"},
				{
					Role: openai.ChatMessageRoleAssistant,
					ToolCalls: []openai.ToolCall{
						{ID: "call_1", Type: "function", Function: openai.FunctionCall{Name: "test", Arguments: "{}"}},
					},
				},
				{
					Role:       openai.ChatMessageRoleTool,
					ToolCallID: "call_1",
					Content:    "valid result",
				},
				{
					Role: openai.ChatMessageRoleAssistant,
					ToolCalls: []openai.ToolCall{
						{ID: "call_3", Type: "function", Function: openai.FunctionCall{Name: "test2", Arguments: "{}"}},
					},
				},
				{
					Role:       openai.ChatMessageRoleTool,
					ToolCallID: "call_2",
					Content:    "orphan result",
				},
				{
					Role:       openai.ChatMessageRoleTool,
					ToolCallID: "call_4",
					Content:    "orphan result 2",
				},
			},
			wantCount: 4,
		},
		{
			name: "mixed messages with some invalid",
			messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: "system"},
				{Role: openai.ChatMessageRoleUser, Content: "user"},
				{Role: openai.ChatMessageRoleAssistant, Content: "assistant text"},
				{
					Role:       openai.ChatMessageRoleTool,
					ToolCallID: "call_1",
					Content:    "result",
				},
				{
					Role:       openai.ChatMessageRoleTool,
					ToolCallID: "call_2",
					Content:    "orphan result",
				},
			},
			wantCount: 3,
		},
		{
			name:      "empty messages",
			messages:  []openai.ChatCompletionMessage{},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{}
			agent.messages = tt.messages
			result := agent.validateMessagePairs()

			assert.Equal(t, tt.wantCount, len(result), "message count mismatch")
		})
	}
}

func TestValidateMessagePairsForSlice(t *testing.T) {
	tests := []struct {
		name      string
		messages  []openai.ChatCompletionMessage
		wantCount int
	}{
		{
			name: "all messages valid",
			messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleAssistant,
					ToolCalls: []openai.ToolCall{
						{ID: "call_1", Type: "function", Function: openai.FunctionCall{Name: "test", Arguments: "{}"}},
					},
				},
				{
					Role:       openai.ChatMessageRoleTool,
					ToolCallID: "call_1",
					Content:    "result",
				},
			},
			wantCount: 2,
		},
		{
			name: "orphaned tool message without parent",
			messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleAssistant,
					ToolCalls: []openai.ToolCall{
						{ID: "call_1", Type: "function", Function: openai.FunctionCall{Name: "test", Arguments: "{}"}},
					},
				},
				{
					Role:       openai.ChatMessageRoleTool,
					ToolCallID: "call_2",
					Content:    "orphan result",
				},
			},
			wantCount: 1,
		},
		{
			name:      "empty slice",
			messages:  []openai.ChatCompletionMessage{},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{}
			result := agent.validateMessagePairsForSlice(tt.messages)

			assert.Equal(t, tt.wantCount, len(result), "message count mismatch")
		})
	}
}
