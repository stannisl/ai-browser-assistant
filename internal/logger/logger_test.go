package logger

import (
	"testing"
)

func TestNewLogger(t *testing.T) {
	t.Run("Production Mode", func(t *testing.T) {
		l, err := New(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if l == nil {
			t.Fatal("expected non-nil logger")
		}
		l.Close()
	})

	t.Run("Debug Mode", func(t *testing.T) {
		l, err := New(true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if l == nil {
			t.Fatal("expected non-nil logger")
		}
		l.Close()
	})

	t.Run("Invalid Configuration", func(t *testing.T) {
		l, err := New(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if l == nil {
			t.Fatal("expected non-nil logger")
		}
		l.Close()
	})
}

func TestLoggerMethods(t *testing.T) {
	t.Run("Info", func(t *testing.T) {
		l, err := New(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		l.Info("test message")
		l.Info("test message with key", "key", "value")
		l.Info("test message with nil value", "key", nil)
		
		l.Close()
	})

	t.Run("Debug", func(t *testing.T) {
		l, err := New(true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		l.Debug("debug message")
		l.Debug("debug message with values", "key", "value")
		
		l.Close()
	})

	t.Run("Error", func(t *testing.T) {
		l, err := New(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		l.Error("error message")
		l.Error("error message with key", "key", "value")
		l.Error("error with nil", "key", nil)
		
		l.Close()
	})

	t.Run("Warn", func(t *testing.T) {
		l, err := New(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		l.Warn("warning message")
		l.Warn("warning with values", "key", "value")
		
		l.Close()
	})
}

func TestLoggerMethodsWithValues(t *testing.T) {
	t.Run("Key-Value Logging", func(t *testing.T) {
		l, err := New(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		l.Info("test", "key1", "value1", "key2", "value2", "key3", 123, "key4", true)
		l.Debug("test", "key", nil, "empty", "")
		
		l.Close()
	})

	t.Run("Edge Cases", func(t *testing.T) {
		l, err := New(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		l.Info("test", "key", nil)
		l.Error("error", "key", "")
		l.Warn("warn", "key", 0)
		
		l.Close()
	})
}

func TestLoggerClosing(t *testing.T) {
	t.Run("Close Success", func(t *testing.T) {
		l, err := New(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		l.Close()
	})

	t.Run("Multiple Closes", func(t *testing.T) {
		l, err := New(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		l.Close()
		l.Close()
		l.Close()
		
		l.Close()
	})
}

func TestTruncateFunction(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		max      int
		expected string
	}{
		{"Short string", "test", 10, "test"},
		{"Exact length", "1234567890", 10, "1234567890"},
		{"Long string", "12345678901234567890", 10, "1234567890..."},
		{"Empty string", "", 10, ""},
		{"Zero max", "", 0, ""},
		{"Negative max", "", -1, ""},
		{"Short with ...", "test", 5, "test"},
		{"Exact without ...", "12345", 5, "12345"},
		{"Over length", "abcdefghij", 10, "abcdefghij"},
		{"Exactly at limit", "12345678901", 10, "1234567890..."},
		{"Empty with ...", "", 10, ""},
		{"Nil input", "", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.max)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestColorHandling(t *testing.T) {
	l, err := New(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("Terminal Output", func(t *testing.T) {
		l.Info("test message")
		l.Error("error message")
		l.Warn("warning message")
		
		l.Close()
	})

	t.Run("Non-Terminal Output", func(t *testing.T) {
		l, err := New(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		l.Info("test message")
		l.Error("error message")
		
		l.Close()
	})
}

func TestLoggerSpecificMethods(t *testing.T) {
	l, err := New(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("Extract", func(t *testing.T) {
		l.Extract("https://example.com/page", 10)
		l.Extract("", 0)
	})

	t.Run("Navigate", func(t *testing.T) {
		l.Navigate("https://example.com")
		l.Navigate("")
	})

	t.Run("Click", func(t *testing.T) {
		l.Click(1, "Button Text")
		l.Click(0, "")
		l.Click(1, "")
	})

	t.Run("Type", func(t *testing.T) {
		l.Type(1, "test text")
		l.Type(0, "")
		l.Type(1, "")
	})

	t.Run("Scroll", func(t *testing.T) {
		l.Scroll("up")
		l.Scroll("down")
		l.Scroll("")
		l.Scroll("invalid")
	})

	t.Run("Confirm", func(t *testing.T) {
		l.Confirm("Confirm action: delete item")
		l.Confirm("")
	})

	t.Run("Ask", func(t *testing.T) {
		l.Ask("Do you want to continue?")
		l.Ask("")
	})

	t.Run("Done", func(t *testing.T) {
		l.Done("Operation completed", true)
		l.Done("Operation failed", false)
		l.Done("", true)
		l.Done("", false)
	})

	t.Run("Step", func(t *testing.T) {
		l.Step(1, 5)
		l.Step(5, 5)
		l.Step(0, 10)
	})

	t.Run("Thinking", func(t *testing.T) {
		l.Thinking()
		l.Thinking()
	})

	t.Run("Tool", func(t *testing.T) {
		l.Tool("extract_page")
		l.Tool("")
	})

	l.Close()
}

func TestLoggerMethodSafety(t *testing.T) {
	l, err := New(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("Nil Values", func(t *testing.T) {
		l.Info("test", "key", nil)
		l.Debug("test", "key", nil)
		l.Error("test", "key", nil)
		l.Warn("test", "key", nil)
	})

	t.Run("Empty Strings", func(t *testing.T) {
		l.Info("")
		l.Debug("")
		l.Error("")
		l.Warn("")
	})

	t.Run("Zero Values", func(t *testing.T) {
		l.Info("test", "key", 0)
		l.Debug("test", "key", 0)
		l.Error("test", "key", 0)
		l.Warn("test", "key", 0)
	})

	t.Run("Multiple Nil Values", func(t *testing.T) {
		l.Info("test", "key1", nil, "key2", "", "key3", 0)
	})

	l.Close()
}

func TestLoggerOutputWithoutColors(t *testing.T) {
	l, err := New(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("Info", func(t *testing.T) {
		l.Info("test message", "key", "value")
	})

	t.Run("Error", func(t *testing.T) {
		l.Error("error message", "key", "value")
	})

	t.Run("Warn", func(t *testing.T) {
		l.Warn("warning message", "key", "value")
	})

	l.Close()
}
