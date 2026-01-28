package types

import "fmt"

var (
	ErrElementNotFound          = fmt.Errorf("element not found")
	ErrNavigationFailed         = fmt.Errorf("navigation failed")
	ErrContextExhausted         = fmt.Errorf("context budget exhausted")
	ErrSecurityRequired         = fmt.Errorf("security: action requires user confirmation")
	ErrInvalidURL               = fmt.Errorf("invalid URL")
	ErrPageLoadTimeout          = fmt.Errorf("page load timeout")
	ErrElementInteractionFailed = fmt.Errorf("element interaction failed")
	ErrTooManyAttempts          = fmt.Errorf("too many attempts")
	ErrInvalidSelector          = fmt.Errorf("invalid selector")
	ErrNoValidElements          = fmt.Errorf("no valid elements found")
	ErrPageIncompatible         = fmt.Errorf("page not compatible with current operation")
	ErrTimeout                  = fmt.Errorf("operation timeout")
	ErrContextCanceled          = fmt.Errorf("operation canceled")
	ErrAuthRequired             = fmt.Errorf("authentication required")
	ErrRateLimited              = fmt.Errorf("rate limit exceeded")
	ErrNetworkError             = fmt.Errorf("network error")
	ErrToolExecutionFailed      = fmt.Errorf("tool execution failed")
	ErrLLMResponseInvalid       = fmt.Errorf("invalid LLM response")
	ErrMaxStepsExceeded         = fmt.Errorf("maximum steps exceeded")
	ErrConfirmationDenied       = fmt.Errorf("user denied action confirmation")
)

type ToolExecutionError struct {
	ToolName string
	Err      error
}

func (e *ToolExecutionError) Error() string {
	return fmt.Sprintf("tool %s execution failed: %w", e.ToolName, e.Error)
}

func (e *ToolExecutionError) Unwrap() error {
	return e.Err
}

type ContextError struct {
	BudgetUsed int
	BudgetMax  int
}

func (e *ContextError) Error() string {
	return fmt.Sprintf("context usage: %d/%d tokens", e.BudgetUsed, e.BudgetMax)
}

type SecurityError struct {
	Operation string
	Reason    string
}

func (e *SecurityError) Unwrap() error {
	return fmt.Errorf("%w: %s - %s", ErrSecurityRequired, e.Operation, e.Reason)
}

func (e *SecurityError) Error() string {
	return fmt.Sprintf("security: %s - %s", e.Operation, e.Reason)
}
