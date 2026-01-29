package types

import (
	"errors"
	"testing"
)

func TestErrorVariables(t *testing.T) {
	errorsToTest := []struct {
		name  string
		err   error
		want  string
	}{
		{"ErrElementNotFound", ErrElementNotFound, "element not found"},
		{"ErrNavigationFailed", ErrNavigationFailed, "navigation failed"},
		{"ErrContextExhausted", ErrContextExhausted, "context budget exhausted"},
		{"ErrSecurityRequired", ErrSecurityRequired, "security: action requires user confirmation"},
		{"ErrInvalidURL", ErrInvalidURL, "invalid URL"},
		{"ErrPageLoadTimeout", ErrPageLoadTimeout, "page load timeout"},
		{"ErrElementInteractionFailed", ErrElementInteractionFailed, "element interaction failed"},
		{"ErrTooManyAttempts", ErrTooManyAttempts, "too many attempts"},
		{"ErrInvalidSelector", ErrInvalidSelector, "invalid selector"},
		{"ErrNoValidElements", ErrNoValidElements, "no valid elements found"},
		{"ErrPageIncompatible", ErrPageIncompatible, "page not compatible with current operation"},
		{"ErrTimeout", ErrTimeout, "operation timeout"},
		{"ErrContextCanceled", ErrContextCanceled, "operation canceled"},
		{"ErrAuthRequired", ErrAuthRequired, "authentication required"},
		{"ErrRateLimited", ErrRateLimited, "rate limit exceeded"},
		{"ErrNetworkError", ErrNetworkError, "network error"},
		{"ErrToolExecutionFailed", ErrToolExecutionFailed, "tool execution failed"},
		{"ErrLLMResponseInvalid", ErrLLMResponseInvalid, "invalid LLM response"},
		{"ErrMaxStepsExceeded", ErrMaxStepsExceeded, "maximum steps exceeded"},
		{"ErrConfirmationDenied", ErrConfirmationDenied, "user denied action confirmation"},
	}

	for _, tt := range errorsToTest {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("expected error '%s', got '%s'", tt.want, tt.err.Error())
			}
		})
	}
}

func TestToolExecutionError(t *testing.T) {
	underlyingErr := errors.New("element not found")
	toolErr := &ToolExecutionError{
		ToolName: "click",
		Err:      underlyingErr,
	}

	if toolErr.ToolName != "click" {
		t.Errorf("expected ToolName 'click', got '%s'", toolErr.ToolName)
	}

	expected := "tool click execution failed: element not found"
	if toolErr.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, toolErr.Error())
	}

	unwrapped := errors.Unwrap(toolErr)
	if unwrapped != underlyingErr {
		t.Errorf("expected unwrapped error to be '%v', got '%v'", underlyingErr, unwrapped)
	}
}

func TestToolExecutionError_EmptyToolName(t *testing.T) {
	toolErr := &ToolExecutionError{
		ToolName: "",
		Err:      errors.New("some error"),
	}

	if toolErr.ToolName != "" {
		t.Errorf("expected empty ToolName, got '%s'", toolErr.ToolName)
	}
}

func TestToolExecutionError_Unwrap(t *testing.T) {
	underlying := errors.New("base error")
	toolErr := &ToolExecutionError{
		ToolName: "test-tool",
		Err:      underlying,
	}

	unwrapped := toolErr.Unwrap()
	if unwrapped != underlying {
		t.Errorf("expected unwrapped error to be '%v', got '%v'", underlying, unwrapped)
	}
}

func TestContextError(t *testing.T) {
	err := &ContextError{
		BudgetUsed: 8000,
		BudgetMax:  10000,
	}

	expected := "context usage: 8000/10000 tokens"
	if err.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, err.Error())
	}

	err2 := &ContextError{
		BudgetUsed: 10000,
		BudgetMax:  10000,
	}

	expected2 := "context usage: 10000/10000 tokens"
	if err2.Error() != expected2 {
		t.Errorf("expected error '%s', got '%s'", expected2, err2.Error())
	}

	err3 := &ContextError{
		BudgetUsed: 500,
		BudgetMax:  1000,
	}

	expected3 := "context usage: 500/1000 tokens"
	if err3.Error() != expected3 {
		t.Errorf("expected error '%s', got '%s'", expected3, err3.Error())
	}
}

func TestContextError_ZeroBudget(t *testing.T) {
	err := &ContextError{
		BudgetUsed: 0,
		BudgetMax:  0,
	}

	expected := "context usage: 0/0 tokens"
	if err.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestSecurityError(t *testing.T) {
	err := &SecurityError{
		Operation: "delete",
		Reason:    "irreversible action",
	}

	expected := "security: delete - irreversible action"
	if err.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, err.Error())
	}

	err2 := &SecurityError{
		Operation: "",
		Reason:    "test reason",
	}

	expected2 := "security:  - test reason"
	if err2.Error() != expected2 {
		t.Errorf("expected error '%s', got '%s'", expected2, err2.Error())
	}
}

func TestSecurityError_Unwrap(t *testing.T) {
	err := &SecurityError{
		Operation: "test",
		Reason:    "test reason",
	}

	unwrapped := err.Unwrap()
	if unwrapped == nil {
		t.Error("expected unwrapped error to be ErrSecurityRequired")
	}

	expected := "security: action requires user confirmation: test - test reason"
	if unwrapped.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, unwrapped.Error())
	}
}

func TestSecurityError_WithOperationOnly(t *testing.T) {
	err := &SecurityError{
		Operation: "test",
		Reason:    "",
	}

	expected := "security: test - "
	if err.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestErrorVariables_AllUnique(t *testing.T) {
	var errorSet = make(map[string]struct{})
	errorsToTest := []error{
		ErrElementNotFound,
		ErrNavigationFailed,
		ErrContextExhausted,
		ErrSecurityRequired,
		ErrInvalidURL,
		ErrPageLoadTimeout,
		ErrElementInteractionFailed,
		ErrTooManyAttempts,
		ErrInvalidSelector,
		ErrNoValidElements,
		ErrPageIncompatible,
		ErrTimeout,
		ErrContextCanceled,
		ErrAuthRequired,
		ErrRateLimited,
		ErrNetworkError,
		ErrToolExecutionFailed,
		ErrLLMResponseInvalid,
		ErrMaxStepsExceeded,
		ErrConfirmationDenied,
	}

	for _, err := range errorsToTest {
		errStr := err.Error()
		if _, exists := errorSet[errStr]; exists {
			t.Errorf("duplicate error found: '%s'", errStr)
		}
		errorSet[errStr] = struct{}{}
	}
}

func TestToolExecutionError_IsError(t *testing.T) {
	toolErr := &ToolExecutionError{
		ToolName: "test-tool",
		Err:      errors.New("test error"),
	}

	if !errors.Is(toolErr, toolErr.Err) {
		t.Error("expected errors.Is to return true for ToolExecutionError")
	}
}

func TestToolExecutionError_NilErr(t *testing.T) {
	toolErr := &ToolExecutionError{
		ToolName: "test-tool",
		Err:      nil,
	}

	if toolErr.Err != nil {
		t.Errorf("expected nil Err, got '%v'", toolErr.Err)
	}
}

func TestToolExecutionError_EmptyError(t *testing.T) {
	toolErr := &ToolExecutionError{
		ToolName: "test-tool",
		Err:      errors.New("execution failed"),
	}

	expected := "tool test-tool execution failed: execution failed"
	if toolErr.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, toolErr.Error())
	}
}
