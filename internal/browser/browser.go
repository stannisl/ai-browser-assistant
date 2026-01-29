package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"

	"github.com/stannisl/ai-browser-assistant/internal/logger"
	"github.com/stannisl/ai-browser-assistant/internal/types"
)

var keyMapping = map[string]input.Key{
	"Enter":      input.Enter,
	"Escape":     input.Escape,
	"Tab":        input.Tab,
	"ArrowDown":  input.ArrowDown,
	"ArrowUp":    input.ArrowUp,
	"ArrowLeft":  input.ArrowLeft,
	"ArrowRight": input.ArrowRight,
	"Backspace":  input.Backspace,
	"Delete":     input.Delete,
	"Space":      input.Space,
}

type Manager struct {
	browser *rod.Browser
	page    *rod.Page
	config  *types.BrowserConfig
	log     *logger.Logger
}

func NewManager(config *types.BrowserConfig, log *logger.Logger) *Manager {
	return &Manager{
		config: config,
		log:    log,
	}
}

func (m *Manager) Launch(ctx context.Context) error {
	launcher := launcher.New().
		Headless(false).
		UserDataDir(m.config.UserDataDir).
		MustLaunch()

	if m.config.Debug {
		m.log.Debug("Browser launched in visible mode", "UserDataDir", m.config.UserDataDir)
	}

	m.browser = rod.New().
		ControlURL(launcher).
		MustConnect()

	if m.config.Debug {
		m.log.Debug("Rod browser instance created")
	}

	m.page = m.browser.MustPage("about:blank")

	if m.config.Debug {
		m.log.Debug("Browser page initialized")
	}

	return nil
}

func (m *Manager) Navigate(ctx context.Context, url string) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("navigation canceled: %w", ctx.Err())
	default:
	}

	if !isValidURL(url) {
		return fmt.Errorf("invalid URL: %s", url)
	}

	if m.config.Debug {
		m.log.Debug("Navigating to URL", "url", url)
	}

	err := m.page.Navigate(url)
	if err != nil {
		return fmt.Errorf("navigation to %s failed: %w", url, err)
	}

	err = m.page.WaitLoad()
	if err != nil {
		return fmt.Errorf("wait for page load failed: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	if m.config.Debug {
		m.log.Debug("Page loaded successfully", "url", url)
	}

	return nil
}

func (m *Manager) Click(ctx context.Context, selector string) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("click canceled: %w", ctx.Err())
	default:
	}

	if m.config.Debug {
		m.log.Debug("Clicking element", "selector", selector)
	}

	el, err := m.page.Element(selector)
	if err != nil {
		return fmt.Errorf("click on element %s failed: %w", selector, err)
	}

	err = el.Click(proto.InputMouseButtonLeft, 1)
	if err != nil {
		return fmt.Errorf("click on element %s failed: %w", selector, err)
	}

	if m.config.Debug {
		m.log.Debug("Element clicked successfully", "selector", selector)
	}

	return nil
}

func (m *Manager) ClickByID(ctx context.Context, id int) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("click canceled: %w", ctx.Err())
	default:
	}

	if m.config.Debug {
		m.log.Debug("Clicking element by ID", "id", id)
	}

	if id < 0 {
		return fmt.Errorf("invalid element ID: must be non-negative")
	}

	selector, err := m.GetSelector(id)
	if err != nil {
		return fmt.Errorf("element ID %d not found", id)
	}

	el, err := m.page.Element(selector)
	if err != nil {
		return fmt.Errorf("click on element %s failed: %w", selector, err)
	}

	if m.config.Debug {
		m.log.Debug("Clicking element", "selector", selector)
	}

	err = el.Click(proto.InputMouseButtonLeft, 1)
	if err != nil {
		m.log.Error("Click failed", err, "selector", selector)

		if m.config.Debug {
			m.log.Debug("Trying dispatchEvent fallback", "selector", selector)
		}

		_, err = el.Eval("arguments[0].click()", el)
		if err != nil {
			return fmt.Errorf("click and dispatchEvent failed for element %s: %w", selector, err)
		}

		if m.config.Debug {
			m.log.Debug("Element clicked successfully using dispatchEvent fallback", "selector", selector)
		}
	} else {
		if m.config.Debug {
			m.log.Debug("Element clicked successfully", "selector", selector)
		}
	}

	time.Sleep(100 * time.Millisecond)

	return nil
}

func (m *Manager) Type(ctx context.Context, selector, text string) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("type canceled: %w", ctx.Err())
	default:
	}

	if m.config.Debug {
		m.log.Debug("Typing into element", "selector", selector, "text", text)
	}

	el, err := m.page.Element(selector)
	if err != nil {
		return fmt.Errorf("type into element %s failed: %w", selector, err)
	}

	el.MustInput("")
	el.MustInput(text)

	if m.config.Debug {
		m.log.Debug("Text entered successfully", "selector", selector)
	}

	return nil
}

func (m *Manager) TypeByID(ctx context.Context, id int, text string) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("type canceled: %w", ctx.Err())
	default:
	}

	if m.config.Debug {
		m.log.Debug("Typing into element by ID", "id", id, "text", text)
	}

	if id < 0 {
		return fmt.Errorf("invalid element ID: must be non-negative")
	}

	if text == "" {
		return fmt.Errorf("cannot type empty text into element ID %d", id)
	}

	selector, err := m.GetSelector(id)
	if err != nil {
		return fmt.Errorf("element ID %d not found", id)
	}

	return m.Type(ctx, selector, text)
}

func (m *Manager) GetSelector(id int) (string, error) {
	return fmt.Sprintf("[data-element-id=%d]", id), nil
}

func (m *Manager) Scroll(ctx context.Context, direction string) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("scroll canceled: %w", ctx.Err())
	default:
	}

	if m.config.Debug {
		m.log.Debug("Scrolling page", "direction", direction)
	}

	if direction == "up" {
		m.page.MustElement("body").Page().Mouse.Scroll(0, -400, 1) // ScrollUp
	} else if direction == "down" {
		m.page.MustElement("body").Page().Mouse.Scroll(0, 400, 1) // ScrollDown
	} else {
		return fmt.Errorf("invalid scroll direction: %s", direction)
	}

	if m.config.Debug {
		m.log.Debug("Page scrolled successfully", "direction", direction)
	}

	return nil
}

func (m *Manager) GetPage() *rod.Page {
	return m.page
}

func (m *Manager) GetURL() string {
	return m.page.MustElement("html").MustProperty("location.href").String()
}

func (m *Manager) GetTitle() string {
	return m.page.MustElement("html").MustProperty("title").String()
}

func (m *Manager) Close() error {
	if m.page != nil {
		m.page.MustClose()
	}

	if m.browser != nil {
		m.browser.MustClose()
	}

	return nil
}

func isValidURL(url string) bool {
	return len(url) > 0 && (len(url) > 4 && url[:4] == "http")
}

func (m *Manager) PressKey(ctx context.Context, key string) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("press_key canceled: %w", ctx.Err())
	default:
	}

	inputKey, ok := keyMapping[key]
	if !ok {
		return fmt.Errorf("unsupported key: %s. Supported: Enter, Escape, Tab, ArrowDown, ArrowUp, ArrowLeft, ArrowRight, Backspace, Delete, Space", key)
	}

	if m.config.Debug {
		m.log.Debug("Pressing keyboard key", "key", key)
	}

	err := m.page.Keyboard.Press(inputKey)
	if err != nil {
		return fmt.Errorf("press key %s failed: %w", key, err)
	}

	if m.config.Debug {
		m.log.Debug("Key pressed successfully", "key", key)
	}

	return nil
}
