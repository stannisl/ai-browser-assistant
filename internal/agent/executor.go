package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/stannisl/ai-browser-assistant/internal/types"
)

func (a *Agent) executeExtractPage(ctx context.Context) (string, error) {
	state, err := a.extractor.Extract(ctx)
	if err != nil {
		return "", fmt.Errorf("extract page: %w", err)
	}
	return a.extractor.FormatForLLM(state), nil
}

func (a *Agent) executeNavigate(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		URL string `json:"url"`
	}
	
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("parse navigate: %w", err)
	}
	
	a.logger.Navigate(params.URL)
	
	if err := a.browser.Navigate(ctx, params.URL); err != nil {
		a.logger.Error("Navigation failed", err, "url", params.URL)
		return fmt.Sprintf("Error: navigation to %s failed: %v", params.URL, err), nil
	}
	
	return fmt.Sprintf("Navigated to %s. Call extract_page to see the page.", params.URL), nil
}

func (a *Agent) executeClick(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		ElementID int `json:"element_id"`
	}
	
	if err := json.Unmarshal(input, &params); err != nil {
		a.logger.Error("Failed to parse click input", err)
		return "", fmt.Errorf("parse click input: %w, raw: %s", err, string(input))
	}
	
	a.logger.Debug("Click requested", "element_id", params.ElementID)
	
	selector, err := a.extractor.GetSelector(params.ElementID)
	if err != nil {
		a.logger.Error("Element not found", err, "element_id", params.ElementID)
		return fmt.Sprintf("Error: element [%d] not found. Call extract_page to refresh elements.", params.ElementID), nil
	}
	
	a.logger.Click(params.ElementID, selector)
	
	if err := a.browser.Click(ctx, selector); err != nil {
		a.logger.Error("Click failed", err, "selector", selector)
		return fmt.Sprintf("Error: click on [%d] failed: %v. Try another element.", params.ElementID, err), nil
	}
	
	return fmt.Sprintf("Clicked element [%d]. Call extract_page to see the result.", params.ElementID), nil
}

func (a *Agent) executeTypeText(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		ElementID int    `json:"element_id"`
		Text      string `json:"text"`
	}
	
	if err := json.Unmarshal(input, &params); err != nil {
		a.logger.Error("Failed to parse type_text input", err)
		return "", fmt.Errorf("parse type_text: %w, raw: %s", err, string(input))
	}
	
	a.logger.Debug("Type text requested", "element_id", params.ElementID, "text", params.Text)
	
	if params.Text == "" {
		return "Error: text is empty. Provide text to type.", nil
	}
	
	selector, err := a.extractor.GetSelector(params.ElementID)
	if err != nil {
		a.logger.Error("Element not found for typing", err, "element_id", params.ElementID)
		return fmt.Sprintf("Error: element [%d] not found. Call extract_page first.", params.ElementID), nil
	}
	
	a.logger.Type(params.ElementID, params.Text)
	
	if err := a.browser.Type(ctx, selector, params.Text); err != nil {
		a.logger.Error("Type failed", err, "selector", selector)
		return fmt.Sprintf("Error: typing into [%d] failed: %v", params.ElementID, err), nil
	}
	
	return fmt.Sprintf("Typed '%s' into element [%d]. Now click search button or press Enter.", params.Text, params.ElementID), nil
}

func (a *Agent) executeScroll(ctx context.Context, arguments map[string]interface{}) (string, error) {
	direction, ok := arguments["direction"].(string)
	if !ok {
		return "", fmt.Errorf("invalid direction argument")
	}

	if err := a.browser.Scroll(ctx, direction); err != nil {
		return "", err
	}

	return fmt.Sprintf("Scrolled %s", direction), nil
}

func (a *Agent) executeWait(ctx context.Context, arguments map[string]interface{}) (string, error) {
	seconds, ok := arguments["seconds"].(float64)
	if !ok {
		return "", fmt.Errorf("invalid seconds argument")
	}

	waitSeconds := int(seconds)
	if waitSeconds < 1 {
		waitSeconds = 1
	}
	if waitSeconds > 10 {
		waitSeconds = 10
	}

	time.Sleep(time.Duration(waitSeconds) * time.Second)

	return fmt.Sprintf("Waited %d seconds", waitSeconds), nil
}

func (a *Agent) executeAskUser(ctx context.Context, arguments map[string]interface{}) (string, error) {
	question, ok := arguments["question"].(string)
	if !ok {
		return "", fmt.Errorf("invalid question argument")
	}

	a.logger.Ask(question)

	fmt.Printf("\n💬 %s\n", question)
	fmt.Print("Ваш ответ: ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	answer := strings.TrimSpace(scanner.Text())

	return fmt.Sprintf("User answered: %s", answer), nil
}

func (a *Agent) executeConfirmAction(ctx context.Context, arguments map[string]interface{}) (string, error) {
	description, ok := arguments["description"].(string)
	if !ok {
		return "", fmt.Errorf("invalid description argument")
	}

	a.logger.Confirm(description)

	fmt.Printf("\n🔒 [ПОДТВЕРЖДЕНИЕ ТРЕБУЕТСЯ]\n")
	fmt.Printf("Действие: %s\n", description)
	fmt.Print("Продолжить? (yes/no): ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))

	if answer == "yes" || answer == "y" || answer == "да" || answer == "д" {
		return "User confirmed the action. Proceed.", nil
	}

	return "User DENIED the action. Do NOT proceed.", types.ErrConfirmationDenied
}

func (a *Agent) executeReport(ctx context.Context, arguments map[string]interface{}) (string, error) {
	message, ok := arguments["message"].(string)
	if !ok {
		message = "Task completed"
	}

	success, ok := arguments["success"].(bool)
	if !ok {
		success = true
	}

	a.logger.Done(message, success)

	return message, nil
}
