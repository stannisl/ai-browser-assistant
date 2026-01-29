package types

import (
	"testing"
	"time"
)

func TestToolCall(t *testing.T) {
	completedAt := time.Now().Add(1 * time.Second)

	toolCall := ToolCall{
		ID:          "call-1",
		ToolName:    "navigate",
		Arguments:   map[string]interface{}{"url": "https://example.com"},
		Result:      "navigated",
		Error:       nil,
		ExecuteTime: 1 * time.Second,
		CreatedAt:   time.Now(),
		CompletedAt: &completedAt,
		ToolCalls:   nil,
	}

	if toolCall.ID != "call-1" {
		t.Errorf("expected ID 'call-1', got '%s'", toolCall.ID)
	}

	if toolCall.ToolName != "navigate" {
		t.Errorf("expected ToolName 'navigate', got '%s'", toolCall.ToolName)
	}

	if toolCall.Arguments["url"] != "https://example.com" {
		t.Errorf("expected URL 'https://example.com', got '%v'", toolCall.Arguments["url"])
	}

	if toolCall.ExecuteTime != 1*time.Second {
		t.Errorf("expected ExecuteTime 1s, got %v", toolCall.ExecuteTime)
	}

	if toolCall.CompletedAt.IsZero() {
		t.Error("expected CompletedAt to be set")
	}
}

func TestAgentConfig(t *testing.T) {
	config := AgentConfig{
		MaxRetries:           3,
		Timeout:              30 * time.Second,
		SecurityEnabled:      true,
		ConfirmationRequired: true,
		ContextBudget:        10000,
		ContextWindow:        20000,
		SummaryEnabled:       true,
		SummarizeEvery:       5 * time.Minute,
		MaxSteps:             100,
	}

	if config.MaxRetries != 3 {
		t.Errorf("expected MaxRetries 3, got %d", config.MaxRetries)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("expected Timeout 30s, got %v", config.Timeout)
	}

	if config.SecurityEnabled != true {
		t.Error("expected SecurityEnabled true")
	}

	if config.ConfirmationRequired != true {
		t.Error("expected ConfirmationRequired true")
	}

	if config.ContextBudget != 10000 {
		t.Errorf("expected ContextBudget 10000, got %d", config.ContextBudget)
	}

	if config.MaxSteps != 100 {
		t.Errorf("expected MaxSteps 100, got %d", config.MaxSteps)
	}
}

func TestLLMConfig(t *testing.T) {
	config := LLMConfig{
		APIKey:         "test-key-123",
		BaseURL:        "https://api.test.com/v1",
		Model:          "test-model",
		MaxTokens:      4096,
		Temperature:    0.7,
		MaxRetries:     2,
		RequestTimeout: 30 * time.Second,
	}

	if config.APIKey != "test-key-123" {
		t.Errorf("expected APIKey 'test-key-123', got '%s'", config.APIKey)
	}

	if config.BaseURL != "https://api.test.com/v1" {
		t.Errorf("expected BaseURL 'https://api.test.com/v1', got '%s'", config.BaseURL)
	}

	if config.Model != "test-model" {
		t.Errorf("expected Model 'test-model', got '%s'", config.Model)
	}

	if config.MaxTokens != 4096 {
		t.Errorf("expected MaxTokens 4096, got %d", config.MaxTokens)
	}

	if config.Temperature != 0.7 {
		t.Errorf("expected Temperature 0.7, got %f", config.Temperature)
	}

	if config.RequestTimeout != 30*time.Second {
		t.Errorf("expected RequestTimeout 30s, got %v", config.RequestTimeout)
	}

	if config.MaxRetries != 2 {
		t.Errorf("expected MaxRetries 2, got %d", config.MaxRetries)
	}
}

func TestToolDefinition(t *testing.T) {
	def := ToolDefinition{
		Name:        "test-tool",
		Description: "A test tool",
		Parameters: map[string]interface{}{
			"type":     "object",
			"required": []string{"param1"},
			"properties": map[string]interface{}{
				"param1": map[string]interface{}{"type": "string"},
				"param2": map[string]interface{}{"type": "integer"},
			},
		},
	}

	if def.Name != "test-tool" {
		t.Errorf("expected Name 'test-tool', got '%s'", def.Name)
	}

	if def.Description != "A test tool" {
		t.Errorf("expected Description 'A test tool', got '%s'", def.Description)
	}

	if def.Parameters["type"] != "object" {
		t.Errorf("expected Parameters type 'object', got '%v'", def.Parameters["type"])
	}
}

func TestToolParam(t *testing.T) {
	param := ToolParam{
		Type:       "string",
		Required:   true,
		Properties: map[string]interface{}{"format": "text"},
	}

	if param.Type != "string" {
		t.Errorf("expected Type 'string', got '%s'", param.Type)
	}

	if param.Required != true {
		t.Error("expected Required true")
	}

	if param.Properties["format"] != "text" {
		t.Errorf("expected Properties format 'text', got '%v'", param.Properties["format"])
	}
}

func TestLLMResponse(t *testing.T) {
	response := LLMResponse{
		Content:      "Test response content",
		ToolCalls:    nil,
		UsedTokens:   150,
		Model:        "test-model",
		FinishReason: "stop",
	}

	if response.Content != "Test response content" {
		t.Errorf("expected Content 'Test response content', got '%s'", response.Content)
	}

	if response.UsedTokens != 150 {
		t.Errorf("expected UsedTokens 150, got %d", response.UsedTokens)
	}

	if response.Model != "test-model" {
		t.Errorf("expected Model 'test-model', got '%s'", response.Model)
	}

	if response.FinishReason != "stop" {
		t.Errorf("expected FinishReason 'stop', got '%s'", response.FinishReason)
	}

	if response.ToolCalls != nil {
		t.Error("expected ToolCalls to be nil")
	}
}

func TestMessageParam(t *testing.T) {
	message := MessageParam{
		Role:    "user",
		Content: "Test message content",
	}

	if message.Role != "user" {
		t.Errorf("expected Role 'user', got '%s'", message.Role)
	}

	if message.Content != "Test message content" {
		t.Errorf("expected Content 'Test message content', got '%s'", message.Content)
	}

	message = MessageParam{
		Role:    "assistant",
		Content: "Assistant response",
	}

	if message.Role != "assistant" {
		t.Errorf("expected Role 'assistant', got '%s'", message.Role)
	}
}

func TestToolCall_Empty(t *testing.T) {
	toolCall := ToolCall{}

	if toolCall.ID != "" {
		t.Errorf("expected empty ID, got '%s'", toolCall.ID)
	}

	if toolCall.ToolName != "" {
		t.Errorf("expected empty ToolName, got '%s'", toolCall.ToolName)
	}

	if len(toolCall.Arguments) != 0 {
		t.Errorf("expected empty Arguments, got %d items", len(toolCall.Arguments))
	}
}

func TestToolCall_WithNestedToolCalls(t *testing.T) {
	nestedCall := ToolCall{
		ID:        "nested-1",
		ToolName:  "click",
		Arguments: map[string]interface{}{"id": 1},
	}

	toolCall := ToolCall{
		ID:        "parent-1",
		ToolName:  "navigate",
		ToolCalls: []ToolCall{nestedCall},
	}

	if len(toolCall.ToolCalls) != 1 {
		t.Errorf("expected 1 nested call, got %d", len(toolCall.ToolCalls))
	}

	if toolCall.ToolCalls[0].ToolName != "click" {
		t.Errorf("expected nested call ToolName 'click', got '%s'", toolCall.ToolCalls[0].ToolName)
	}
}

func TestAgentConfig_Defaults(t *testing.T) {
	config := AgentConfig{}

	if config.MaxRetries != 0 {
		t.Errorf("expected MaxRetries 0 (default), got %d", config.MaxRetries)
	}

	if config.Timeout != 0 {
		t.Errorf("expected Timeout 0 (default), got %v", config.Timeout)
	}

	if config.SecurityEnabled != false {
		t.Error("expected SecurityEnabled false (default)")
	}

	if config.MaxSteps != 0 {
		t.Errorf("expected MaxSteps 0 (default), got %d", config.MaxSteps)
	}
}

func TestLLMConfig_Defaults(t *testing.T) {
	config := LLMConfig{}

	if config.APIKey != "" {
		t.Errorf("expected empty APIKey, got '%s'", config.APIKey)
	}

	if config.MaxTokens != 0 {
		t.Errorf("expected MaxTokens 0 (default), got %d", config.MaxTokens)
	}

	if config.Temperature != 0 {
		t.Errorf("expected Temperature 0 (default), got %f", config.Temperature)
	}

	if config.RequestTimeout != 0 {
		t.Errorf("expected RequestTimeout 0 (default), got %v", config.RequestTimeout)
	}
}
