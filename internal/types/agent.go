package types

import "time"

type ToolCall struct {
	ID        string
	ToolName  string
	Arguments map[string]interface{}
	Result    interface{}
	Error     error
	ExecuteTime time.Duration
	CreatedAt time.Time
	CompletedAt *time.Time
}

type AgentConfig struct {
	MaxRetries      int
	Timeout         time.Duration
	SecurityEnabled bool
	ConfirmationRequired bool
	ContextBudget    int
	ContextWindow    int
	SummaryEnabled   bool
	SummarizeEvery   time.Duration
	MaxSteps         int
}

type LLMConfig struct {
	APIKey       string
	BaseURL      string
	Model        string
	MaxTokens    int
	Temperature  float64
	MaxRetries   int
	RequestTimeout time.Duration
}

type ToolDefinition struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
}

type ToolParam struct {
	Type     string
	Required bool
	Properties map[string]interface{}
}

type LLMResponse struct {
	Content    string
	ToolCalls  []ToolCall
	UsedTokens int
	Model      string
	FinishReason string
}

type MessageParam struct {
	Role    string
	Content string
}
