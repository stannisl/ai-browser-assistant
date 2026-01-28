package llm

import (
	"encoding/json"
	"testing"
)

func TestParseClickInput(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantID    int
		wantError bool
	}{
		{
			name:      "valid element_id",
			input:     `{"element_id": 46}`,
			wantID:    46,
			wantError: false,
		},
		{
			name:      "element_id zero",
			input:     `{"element_id": 0}`,
			wantID:    0,
			wantError: false,
		},
		{
			name:      "element_id large number",
			input:     `{"element_id": 105}`,
			wantID:    105,
			wantError: false,
		},
		{
			name:      "empty object",
			input:     `{}`,
			wantID:    0,
			wantError: false,
		},
		{
			name:      "invalid json",
			input:     `{invalid}`,
			wantID:    0,
			wantError: true,
		},
		{
			name:      "string instead of int",
			input:     `{"element_id": "46"}`,
			wantID:    0,
			wantError: true,
		},
		{
			name:      "float value - error",
			input:     `{"element_id": 46.5}`,
			wantID:    0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var params ClickInput
			err := json.Unmarshal([]byte(tt.input), &params)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if params.ElementID != tt.wantID {
				t.Errorf("got ElementID=%d, want %d", params.ElementID, tt.wantID)
			}
		})
	}
}

func TestParseTypeTextInput(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantID    int
		wantText  string
		wantError bool
	}{
		{
			name:      "valid input",
			input:     `{"element_id": 7, "text": "Hello World"}`,
			wantID:    7,
			wantText:  "Hello World",
			wantError: false,
		},
		{
			name:      "unicode text",
			input:     `{"element_id": 3, "text": "Привет мир"}`,
			wantID:    3,
			wantText:  "Привет мир",
			wantError: false,
		},
		{
			name:      "empty text",
			input:     `{"element_id": 5, "text": ""}`,
			wantID:    5,
			wantText:  "",
			wantError: false,
		},
		{
			name:      "missing text field",
			input:     `{"element_id": 5}`,
			wantID:    5,
			wantText:  "",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var params TypeTextInput
			err := json.Unmarshal([]byte(tt.input), &params)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if params.ElementID != tt.wantID {
				t.Errorf("got ElementID=%d, want %d", params.ElementID, tt.wantID)
			}
			if params.Text != tt.wantText {
				t.Errorf("got Text=%q, want %q", params.Text, tt.wantText)
			}
		})
	}
}

func TestParseNavigateInput(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantURL   string
		wantError bool
	}{
		{
			name:      "valid url",
			input:     `{"url": "https://mail.yandex.ru"}`,
			wantURL:   "https://mail.yandex.ru",
			wantError: false,
		},
		{
			name:      "url with path",
			input:     `{"url": "https://hh.ru/search/vacancy?text=go"}`,
			wantURL:   "https://hh.ru/search/vacancy?text=go",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var params NavigateInput
			err := json.Unmarshal([]byte(tt.input), &params)

			if tt.wantError && err == nil {
				t.Errorf("expected error, got nil")
				return
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if params.URL != tt.wantURL {
				t.Errorf("got URL=%q, want %q", params.URL, tt.wantURL)
			}
		})
	}
}

func TestParseScrollInput(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		wantDir   string
	}{
		{
			name:      "valid direction up",
			input:     `{"direction": "up"}`,
			wantError: false,
			wantDir:   "up",
		},
		{
			name:      "valid direction down",
			input:     `{"direction": "down"}`,
			wantError: false,
			wantDir:   "down",
		},
		{
			name:      "invalid direction",
			input:     `{"direction": "left"}`,
			wantError: false,
			wantDir:   "left",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var params ScrollInput
			err := json.Unmarshal([]byte(tt.input), &params)

			if tt.wantError && err == nil {
				t.Errorf("expected error, got nil")
				return
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if params.Direction != tt.wantDir {
				t.Errorf("got Direction=%q, want %q", params.Direction, tt.wantDir)
			}
		})
	}
}

func TestParseWaitInput(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		wantSec   int
	}{
		{
			name:      "valid seconds",
			input:     `{"seconds": 5}`,
			wantError: false,
			wantSec:   5,
		},
		{
			name:      "seconds at minimum",
			input:     `{"seconds": 1}`,
			wantError: false,
			wantSec:   1,
		},
		{
			name:      "seconds at maximum",
			input:     `{"seconds": 10}`,
			wantError: false,
			wantSec:   10,
		},
		{
			name:      "invalid seconds",
			input:     `{"seconds": 0}`,
			wantError: false,
			wantSec:   0,
		},
		{
			name:      "seconds zero",
			input:     `{"seconds": 0}`,
			wantError: false,
			wantSec:   0,
		},
		{
			name:      "missing seconds field",
			input:     `{}`,
			wantError: false,
			wantSec:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var params WaitInput
			err := json.Unmarshal([]byte(tt.input), &params)

			if tt.wantError && err == nil {
				t.Errorf("expected error, got nil")
				return
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if params.Seconds != tt.wantSec {
				t.Errorf("got Seconds=%d, want %d", params.Seconds, tt.wantSec)
			}
		})
	}
}

func TestParseAskUserInput(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		wantQ     string
	}{
		{
			name:      "valid question",
			input:     `{"question": "Do you want to proceed?"}`,
			wantError: false,
			wantQ:     "Do you want to proceed?",
		},
		{
			name:      "empty question",
			input:     `{"question": ""}`,
			wantError: false,
			wantQ:     "",
		},
		{
			name:      "question with unicode",
			input:     `{"question": "Как дела?"}`,
			wantError: false,
			wantQ:     "Как дела?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var params AskUserInput
			err := json.Unmarshal([]byte(tt.input), &params)

			if tt.wantError && err == nil {
				t.Errorf("expected error, got nil")
				return
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if params.Question != tt.wantQ {
				t.Errorf("got Question=%q, want %q", params.Question, tt.wantQ)
			}
		})
	}
}

func TestParseConfirmActionInput(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		wantDesc  string
	}{
		{
			name:      "valid description",
			input:     `{"description": "Confirm order submission"}`,
			wantError: false,
			wantDesc:  "Confirm order submission",
		},
		{
			name:      "empty description",
			input:     `{"description": ""}`,
			wantError: false,
			wantDesc:  "",
		},
		{
			name:      "description with unicode",
			input:     `{"description": "Подтвердите действие"}`,
			wantError: false,
			wantDesc:  "Подтвердите действие",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var params ConfirmActionInput
			err := json.Unmarshal([]byte(tt.input), &params)

			if tt.wantError && err == nil {
				t.Errorf("expected error, got nil")
				return
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if params.Description != tt.wantDesc {
				t.Errorf("got Description=%q, want %q", params.Description, tt.wantDesc)
			}
		})
	}
}

func TestParseReportInput(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		wantMsg   string
		wantSuccess bool
	}{
		{
			name:      "valid report",
			input:     `{"message": "Order completed", "success": true}`,
			wantError: false,
			wantMsg:   "Order completed",
			wantSuccess: true,
		},
		{
			name:      "failed report",
			input:     `{"message": "Order failed", "success": false}`,
			wantError: false,
			wantMsg:   "Order failed",
			wantSuccess: false,
		},
		{
			name:      "empty message",
			input:     `{"message": "", "success": true}`,
			wantError: false,
			wantMsg:   "",
			wantSuccess: true,
		},
		{
			name:      "unicode message",
			input:     `{"message": "Успешно!", "success": true}`,
			wantError: false,
			wantMsg:   "Успешно!",
			wantSuccess: true,
		},
		{
			name:      "missing success field",
			input:     `{"message": "Test"}`,
			wantError: false,
			wantMsg:   "Test",
			wantSuccess: false,
		},
		{
			name:      "missing message field",
			input:     `{"success": true}`,
			wantError: false,
			wantMsg:   "",
			wantSuccess: true,
		},
		{
			name:      "empty object",
			input:     `{}`,
			wantError: false,
			wantMsg:   "",
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var params ReportInput
			err := json.Unmarshal([]byte(tt.input), &params)

			if tt.wantError && err == nil {
				t.Errorf("expected error, got nil")
				return
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if params.Message != tt.wantMsg {
				t.Errorf("got Message=%q, want %q", params.Message, tt.wantMsg)
			}
			if params.Success != tt.wantSuccess {
				t.Errorf("got Success=%v, want %v", params.Success, tt.wantSuccess)
			}
		})
	}
}
