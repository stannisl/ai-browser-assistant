package agent

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/stannisl/ai-browser-assistant/internal/types"
)

// ExecuteTool выполняет инструмент и возвращает результат
func (a *Agent) ExecuteTool(ctx context.Context, tc *types.ToolCall) (string, error) {
	switch tc.ToolName {
	case "extract_page":
		return a.executeExtractPage(ctx)
	case "navigate":
		return a.executeNavigate(ctx, tc.Arguments)
	case "click":
		return a.executeClick(ctx, tc.Arguments)
	case "type_text":
		return a.executeTypeText(ctx, tc.Arguments)
	case "scroll":
		return a.executeScroll(ctx, tc.Arguments)
	case "wait":
		return a.executeWait(ctx, tc.Arguments)
	case "press_key":
		return a.executePressKey(ctx, tc.Arguments)
	case "ask_user":
		return a.executeAskUser(ctx, tc.Arguments)
	case "confirm_action":
		return a.executeConfirmAction(ctx, tc.Arguments)
	case "report":
		return a.executeReport(ctx, tc.Arguments)
	default:
		return fmt.Sprintf("Error: unknown tool '%s'", tc.ToolName), nil
	}
}

func (a *Agent) executeExtractPage(ctx context.Context) (string, error) {
	a.extractor.UpdatePage(a.browser.GetPage())

	state, err := a.extractor.Extract(ctx)
	if err != nil {
		return fmt.Sprintf("Error extracting page: %v", err), nil
	}
	return a.extractor.FormatForLLM(state), nil
}

func (a *Agent) executeNavigate(ctx context.Context, args map[string]interface{}) (string, error) {
	url, ok := args["url"].(string)
	if !ok || url == "" {
		return "Error: 'url' argument is required and must be a string", nil
	}

	a.logger.Navigate(url)

	if err := a.browser.Navigate(ctx, url); err != nil {
		return fmt.Sprintf("Error navigating to %s: %v", url, err), nil
	}

	return fmt.Sprintf("Navigated to %s. Call extract_page to see the page content.", url), nil
}

func (a *Agent) executeClick(ctx context.Context, args map[string]interface{}) (string, error) {
	id, err := extractElementID(args)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	a.logger.Click(id, "")

	if err := a.browser.ClickByID(ctx, id); err != nil {
		return fmt.Sprintf("Error clicking element [%d]: %v. Try extract_page to refresh elements.", id, err), nil
	}

	return fmt.Sprintf("Clicked element [%d]. Call extract_page to see the result.", id), nil
}

func (a *Agent) executeTypeText(ctx context.Context, args map[string]interface{}) (string, error) {
	id, err := extractElementID(args)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	text, ok := args["text"].(string)
	if !ok {
		return "Error: 'text' argument is required and must be a string", nil
	}

	if text == "" {
		return "Error: 'text' cannot be empty", nil
	}

	a.logger.Type(id, text)

	if err := a.browser.TypeByID(ctx, id, text); err != nil {
		return fmt.Sprintf("Error typing into element [%d]: %v. Try extract_page to refresh elements.", id, err), nil
	}

	return fmt.Sprintf("Typed '%s' into element [%d]. Call extract_page to see the result.", text, id), nil
}

func (a *Agent) executeScroll(ctx context.Context, args map[string]interface{}) (string, error) {
	direction, ok := args["direction"].(string)
	if !ok || direction == "" {
		return "Error: 'direction' argument is required (use 'up' or 'down')", nil
	}

	a.logger.Scroll(direction)

	if err := a.browser.Scroll(ctx, direction); err != nil {
		return fmt.Sprintf("Error scrolling: %v", err), nil
	}

	return fmt.Sprintf("Scrolled %s. Call extract_page to see new elements.", direction), nil
}

func (a *Agent) executeWait(ctx context.Context, args map[string]interface{}) (string, error) {
	seconds := 2.0 // default
	if s, ok := args["seconds"].(float64); ok {
		seconds = s
	}

	// Ограничиваем
	if seconds < 1 {
		seconds = 1
	}
	if seconds > 10 {
		seconds = 10
	}

	time.Sleep(time.Duration(seconds) * time.Second)

	return fmt.Sprintf("Waited %.0f seconds.", seconds), nil
}

func (a *Agent) executePressKey(ctx context.Context, args map[string]interface{}) (string, error) {
	key, ok := args["key"].(string)
	if !ok || key == "" {
		return "Error: 'key' argument is required (Enter, Escape, Tab, ArrowDown, ArrowUp)", nil
	}

	if err := a.browser.PressKey(ctx, key); err != nil {
		return fmt.Sprintf("Error pressing key '%s': %v", key, err), nil
	}

	return fmt.Sprintf("Pressed %s key. Call extract_page to see the result.", key), nil
}

func (a *Agent) executeAskUser(ctx context.Context, args map[string]interface{}) (string, error) {
	question, ok := args["question"].(string)
	if !ok || question == "" {
		return "Error: 'question' argument is required", nil
	}

	a.logger.Ask(question)

	fmt.Printf("\n💬 Agent asks: %s\n", question)
	fmt.Print("Your answer: ")

	// Используем bufio.Scanner для корректного чтения строки с пробелами
	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Sprintf("Error reading user input: %v", err), nil
	}

	answer = strings.TrimSpace(answer)

	if answer == "" {
		return "User did not provide an answer. Ask again or try a different approach.", nil
	}

	fmt.Printf("✅ Received: %s\n\n", answer)

	return fmt.Sprintf("User answered: %s", answer), nil
}

func (a *Agent) executeConfirmAction(ctx context.Context, args map[string]interface{}) (string, error) {
	description, ok := args["description"].(string)
	if !ok || description == "" {
		description = "Perform this action?"
	}

	a.logger.Confirm(description)

	fmt.Printf("\n🔒 CONFIRMATION REQUIRED\n")
	fmt.Printf("Action: %s\n", description)
	fmt.Print("Proceed? (yes/no): ")

	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return "Error reading confirmation. Action denied.", nil
	}

	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer == "yes" || answer == "y" || answer == "да" || answer == "д" {
		fmt.Println("✅ Confirmed")
		return "User confirmed. Proceed with the action.", nil
	}

	fmt.Println("❌ Denied")
	return "User DENIED the action. Do NOT proceed.", nil
}

func (a *Agent) executeReport(ctx context.Context, args map[string]interface{}) (string, error) {
	message, ok := args["message"].(string)
	if !ok {
		message = "Task completed"
	}

	success := true
	if s, ok := args["success"].(bool); ok {
		success = s
	}

	a.logger.Done(message, success)

	return message, nil
}

// extractElementID извлекает ID элемента из аргументов
func extractElementID(args map[string]interface{}) (int, error) {
	if v, ok := args["element_id"].(float64); ok {
		return int(v), nil
	}
	if v, ok := args["id"].(float64); ok {
		return int(v), nil
	}
	if v, ok := args["element_id"].(int); ok {
		return v, nil
	}
	if v, ok := args["id"].(int); ok {
		return v, nil
	}

	return 0, fmt.Errorf("'element_id' is required and must be a number")
}
