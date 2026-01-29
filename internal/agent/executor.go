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

func (a *Agent) ExecuteExtractPage(ctx context.Context, arguments map[string]interface{}) (string, error) {
	state, err := a.extractor.Extract(ctx)
	if err != nil {
		return "", fmt.Errorf("extract page: %w", err)
	}
	return a.extractor.FormatForLLM(state), nil
}

func (a *Agent) ExecuteNavigate(ctx context.Context, arguments map[string]interface{}) (string, error) {
	url, ok := arguments["url"].(string)
	if !ok || url == "" {
		return "", fmt.Errorf("invalid URL argument: URL is required and must be a string")
	}

	a.logger.Navigate(url)

	if err := a.browser.Navigate(ctx, url); err != nil {
		a.logger.Error("Navigation failed", err, "url", url)
		return fmt.Sprintf("Error: navigation to %s failed: %v", url, err), nil
	}

	return fmt.Sprintf("Navigated to %s. Call extract_page to see the page.", url), nil
}

func (a *Agent) ExecuteClick(ctx context.Context, arguments map[string]interface{}) (string, error) {
	elementID, ok := arguments["element_id"]
	if !ok {
		return "", fmt.Errorf("invalid element_id argument: element_id is required")
	}

	id, ok := elementID.(float64)
	if !ok {
		return "", fmt.Errorf("invalid element_id argument: must be a number")
	}

	if id < 0 {
		return "", fmt.Errorf("invalid element_id argument: must be non-negative")
	}

	a.logger.Debug("Click requested", "element_id", int(id))

	selector, err := a.extractor.GetSelector(int(id))
	if err != nil {
		a.logger.Error("Element not found", err, "element_id", int(id))
		return fmt.Sprintf("Error: element [%d] not found. Call extract_page to refresh elements.", int(id)), nil
	}

	a.logger.Click(int(id), selector)

	if err := a.browser.Click(ctx, selector); err != nil {
		a.logger.Error("Click failed", err, "selector", selector)
		return fmt.Sprintf("Error: click on [%d] failed: %v. Try another element.", int(id), err), nil
	}

	return fmt.Sprintf("Clicked element [%d]. Call extract_page to see the result.", int(id)), nil
}

func (a *Agent) ExecuteTypeText(ctx context.Context, arguments map[string]interface{}) (string, error) {
	elementID, ok := arguments["element_id"]
	if !ok {
		return "", fmt.Errorf("invalid element_id argument: element_id is required")
	}

	text, ok := arguments["text"]
	if !ok {
		return "", fmt.Errorf("invalid text argument: text is required")
	}

	id, ok := elementID.(float64)
	if !ok {
		return "", fmt.Errorf("invalid element_id argument: must be a number")
	}

	textStr, ok := text.(string)
	if !ok {
		return "", fmt.Errorf("invalid text argument: must be a string")
	}

	if textStr == "" {
		return "Error: text is empty. Provide text to type.", nil
	}

	a.logger.Debug("Type text requested", "element_id", int(id), "text", textStr)

	selector, err := a.extractor.GetSelector(int(id))
	if err != nil {
		a.logger.Error("Element not found for typing", err, "element_id", int(id))
		return fmt.Sprintf("Error: element [%d] not found. Call extract_page first.", int(id)), nil
	}

	a.logger.Type(int(id), textStr)

	if err := a.browser.Type(ctx, selector, textStr); err != nil {
		a.logger.Error("Type failed", err, "selector", selector)
		return fmt.Sprintf("Error: typing into [%d] failed: %v", int(id), err), nil
	}

	return fmt.Sprintf("Typed '%s' into element [%d]. Now click search button or press Enter.", textStr, int(id)), nil
}

func (a *Agent) ExecuteScroll(ctx context.Context, arguments map[string]interface{}) (string, error) {
	direction, ok := arguments["direction"]
	if !ok {
		return "", fmt.Errorf("invalid direction argument: direction is required")
	}

	directionStr, ok := direction.(string)
	if !ok {
		return "", fmt.Errorf("invalid direction argument: must be a string")
	}

	validDirections := map[string]bool{
		"up":    true,
		"down":  true,
		"left":  true,
		"right": true,
	}

	if !validDirections[directionStr] {
		return "", fmt.Errorf("invalid direction argument: must be one of %v", validDirections)
	}

	if err := a.browser.Scroll(ctx, directionStr); err != nil {
		return "", err
	}

	return fmt.Sprintf("Scrolled %s", directionStr), nil
}

func (a *Agent) ExecuteWait(ctx context.Context, arguments map[string]interface{}) (string, error) {
	seconds, ok := arguments["seconds"]
	if !ok {
		return "", fmt.Errorf("invalid seconds argument: seconds is required")
	}

	val, ok := seconds.(float64)
	if !ok {
		return "", fmt.Errorf("invalid seconds argument: must be a number")
	}

	secondsInt := int(val)
	if secondsInt < 1 {
		return "", fmt.Errorf("invalid seconds argument: must be at least 1")
	}
	if secondsInt > 10 {
		return "", fmt.Errorf("invalid seconds argument: must be at most 10")
	}

	time.Sleep(time.Duration(secondsInt) * time.Second)

	return fmt.Sprintf("Waited %d seconds", secondsInt), nil
}

func (a *Agent) ExecuteAskUser(ctx context.Context, arguments map[string]interface{}) (string, error) {
	question, ok := arguments["question"]
	if !ok {
		return "", fmt.Errorf("invalid question argument: question is required")
	}

	questionStr, ok := question.(string)
	if !ok {
		return "", fmt.Errorf("invalid question argument: must be a string")
	}

	if questionStr == "" {
		return "", fmt.Errorf("invalid question argument: question cannot be empty")
	}

	a.logger.Ask(questionStr)

	fmt.Printf("\n💬 %s\n", questionStr)
	fmt.Print("Ваш ответ: ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	answer := strings.TrimSpace(scanner.Text())

	return fmt.Sprintf("User answered: %s", answer), nil
}

func (a *Agent) ExecuteConfirmAction(ctx context.Context, arguments map[string]interface{}) (string, error) {
	description, ok := arguments["description"]
	if !ok {
		return "", fmt.Errorf("invalid description argument: description is required")
	}

	descriptionStr, ok := description.(string)
	if !ok {
		return "", fmt.Errorf("invalid description argument: must be a string")
	}

	if descriptionStr == "" {
		return "", fmt.Errorf("invalid description argument: description cannot be empty")
	}

	a.logger.Confirm(descriptionStr)

	fmt.Printf("\n🔒 [ПОДТВЕРЖДЕНИЕ ТРЕБУЕТСЯ]\n")
	fmt.Printf("Действие: %s\n", descriptionStr)
	fmt.Print("Продолжить? (yes/no): ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))

	if answer == "yes" || answer == "y" || answer == "да" || answer == "д" {
		return "User confirmed the action. Proceed.", nil
	}

	return "User DENIED the action. Do NOT proceed.", types.ErrConfirmationDenied
}

func (a *Agent) ExecuteReport(ctx context.Context, arguments map[string]interface{}) (string, error) {
	message, ok := arguments["message"]
	if !ok {
		message = "Task completed"
	}

	messageStr, ok := message.(string)
	if !ok {
		messageStr = "Task completed"
	}

	success, ok := arguments["success"]
	if !ok {
		success = true
	}

	successBool, ok := success.(bool)
	if !ok {
		successBool = true
	}

	a.logger.Done(messageStr, successBool)

	return messageStr, nil
}
