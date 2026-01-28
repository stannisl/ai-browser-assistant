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

func (a *Agent) executeExtractPage(ctx context.Context) (string, error) {
	state, err := a.extractor.Extract(ctx)
	if err != nil {
		return "", fmt.Errorf("extract page: %w", err)
	}
	return a.extractor.FormatForLLM(state), nil
}

func (a *Agent) executeNavigate(ctx context.Context, arguments map[string]interface{}) (string, error) {
	url, ok := arguments["url"].(string)
	if !ok {
		return "", fmt.Errorf("invalid URL argument")
	}

	if err := a.browser.Navigate(ctx, url); err != nil {
		return "", err
	}

	return fmt.Sprintf("Navigated to %s", url), nil
}

func (a *Agent) executeClick(ctx context.Context, arguments map[string]interface{}) (string, error) {
	id, ok := arguments["element_id"].(float64)
	if !ok {
		return "", fmt.Errorf("invalid element_id argument")
	}

	selector, err := a.extractor.GetSelector(int(id))
	if err != nil {
		return "", fmt.Errorf("element [%d] not found", int(id))
	}

	a.logger.Click(int(id), selector)

	if err := a.browser.Click(ctx, selector); err != nil {
		return "", err
	}

	return fmt.Sprintf("Clicked element [%d]", int(id)), nil
}

func (a *Agent) executeTypeText(ctx context.Context, arguments map[string]interface{}) (string, error) {
	id, ok := arguments["element_id"].(float64)
	if !ok {
		return "", fmt.Errorf("invalid element_id argument")
	}

	text, ok := arguments["text"].(string)
	if !ok {
		return "", fmt.Errorf("invalid text argument")
	}

	selector, err := a.extractor.GetSelector(int(id))
	if err != nil {
		return "", fmt.Errorf("element [%d] not found", int(id))
	}

	a.logger.Type(int(id), text)

	if err := a.browser.Type(ctx, selector, text); err != nil {
		return "", err
	}

	return fmt.Sprintf("Typed '%s' into element [%d]", text, int(id)), nil
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
