package llm

import (
	"github.com/sashabaranov/go-openai"
)

func GetTools() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "extract_page",
				Description: "Get current page state with all interactive elements. ALWAYS call this first and after any action to see the result.",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
					"required":   []string{},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "navigate",
				Description: "Navigate to a URL",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"url": map[string]interface{}{
							"type":        "string",
							"description": "The URL to navigate to",
						},
					},
					"required": []string{"url"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "click",
				Description: "Click on an element by its ID from [Interactive Elements] list",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"element_id": map[string]interface{}{
							"type":        "integer",
							"description": "The ID of element to click, e.g. 5 for [5]",
						},
					},
					"required": []string{"element_id"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "type_text",
				Description: "Type text into an input field by its ID",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"element_id": map[string]interface{}{
							"type":        "integer",
							"description": "The ID of element to type into",
						},
						"text": map[string]interface{}{
							"type":        "string",
							"description": "The text to type",
						},
					},
					"required": []string{"element_id", "text"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "scroll",
				Description: "Scroll the page in the specified direction",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"direction": map[string]interface{}{
							"type":        "string",
							"description": "Direction to scroll: 'up' or 'down'",
						},
					},
					"required": []string{"direction"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "wait",
				Description: "Wait for the page to stabilize",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"seconds": map[string]interface{}{
							"type":        "integer",
							"description": "Number of seconds to wait (1-10)",
							"minimum":     1,
							"maximum":     10,
						},
					},
					"required": []string{"seconds"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "ask_user",
				Description: "Ask the user a question and wait for response",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"question": map[string]interface{}{
							"type":        "string",
							"description": "The question to ask the user",
						},
					},
					"required": []string{"question"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "confirm_action",
				Description: "Confirm a potentially destructive action with the user",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Description of the action to confirm",
						},
					},
					"required": []string{"description"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "report",
				Description: "Report the completion of a task or operation",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"message": map[string]interface{}{
							"type":        "string",
							"description": "The message to report",
						},
						"success": map[string]interface{}{
							"type":        "boolean",
							"description": "Whether the operation was successful",
						},
					},
					"required": []string{"message", "success"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "press_key",
				Description: "Press a keyboard key. Use for: Enter (submit forms), Escape (close modals), Tab (next field)",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"key": map[string]interface{}{
							"type":        "string",
							"description": "Key to press: Enter, Escape, Tab, ArrowDown, ArrowUp",
						},
					},
					"required": []string{"key"},
				},
			},
		},
	}
}

type NavigateInput struct {
	URL string `json:"url"`
}

type ClickInput struct {
	ElementID int `json:"element_id"`
}

type TypeTextInput struct {
	ElementID int    `json:"element_id"`
	Text      string `json:"text"`
}

type ScrollInput struct {
	Direction string `json:"direction"`
}

type WaitInput struct {
	Seconds int `json:"seconds"`
}

type AskUserInput struct {
	Question string `json:"question"`
}

type ConfirmActionInput struct {
	Description string `json:"description"`
}

type ReportInput struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type PressKeyInput struct {
	Key string `json:"key"`
}
