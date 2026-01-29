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
	l, err := launcher.New().
		Headless(m.config.Headless).
		UserDataDir(m.config.UserDataDir).Launch()
	if err != nil {
		return fmt.Errorf("creating launcher failed: %w", err)
	}

	if m.config.Debug {
		m.log.Debug("Browser launched", "headless", m.config.Headless, "userDataDir", m.config.UserDataDir)
	}

	m.browser = rod.New().
		ControlURL(l).
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

	url = normalizeURL(url)

	if m.config.Debug {
		m.log.Debug("Navigating to URL", "url", url)
	}

	err := m.page.Navigate(url)
	if err != nil {
		return fmt.Errorf("navigation to %s failed: %w", url, err)
	}

	_ = m.page.WaitLoad()
	time.Sleep(1 * time.Second)

	if m.config.Debug {
		m.log.Debug("Page loaded successfully", "url", url)
	}

	return nil
}

func (m *Manager) ClickByID(ctx context.Context, id int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if id < 0 {
		return fmt.Errorf("invalid element ID: %d", id)
	}

	if m.config.Debug {
		m.log.Debug("Clicking element by ID", "id", id)
	}

	// Получаем количество страниц до клика
	pagesBefore := len(m.browser.MustPages())

	// Клик через JS, который предотвращает открытие новых вкладок
	_, err := m.page.Eval(`(id) => {
		if (!window._ai_elements) {
			throw new Error("No elements extracted. Call extract_page first.");
		}
		
		const el = window._ai_elements[id];
		if (!el) {
			throw new Error("Element ID " + id + " not found. Total elements: " + window._ai_elements.length);
		}
		
		if (!document.contains(el)) {
			throw new Error("Element ID " + id + " is no longer in the DOM. Call extract_page again.");
		}
		
		// Скроллим к элементу
		el.scrollIntoView({block: "center", inline: "center"});
		
		// Удаляем target="_blank" чтобы не открывать новую вкладку
		if (el.tagName.toLowerCase() === 'a') {
			el.removeAttribute('target');
			// Также удаляем rel="noopener" который может мешать
			el.removeAttribute('rel');
		}
		
		// Небольшая пауза после скролла
		return new Promise(resolve => {
			setTimeout(() => {
				el.click();
				resolve(true);
			}, 100);
		});
	}`, id)

	if err != nil {
		return fmt.Errorf("click element [%d]: %w", id, err)
	}

	// Ждём реакции страницы
	time.Sleep(300 * time.Millisecond)

	// Проверяем, не открылась ли новая вкладка
	pagesAfter := m.browser.MustPages()
	if len(pagesAfter) > pagesBefore {
		// Переключаемся на новую вкладку
		m.page = pagesAfter[len(pagesAfter)-1]

		// Закрываем старые вкладки (кроме текущей)
		for i, p := range pagesAfter {
			if i < len(pagesAfter)-1 {
				p.Close()
			}
		}

		// Ждём загрузки новой страницы
		_ = m.page.WaitLoad()
		time.Sleep(500 * time.Millisecond)

		if m.config.Debug {
			m.log.Debug("Switched to new tab", "id", id)
		}
	}

	if m.config.Debug {
		m.log.Debug("Element clicked successfully", "id", id)
	}

	return nil
}

func (m *Manager) TypeByID(ctx context.Context, id int, text string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if id < 0 {
		return fmt.Errorf("invalid element ID: %d", id)
	}

	if text == "" {
		return fmt.Errorf("text cannot be empty")
	}

	if m.config.Debug {
		m.log.Debug("Typing into element by ID", "id", id, "text", text)
	}

	_, err := m.page.Eval(`(args) => {
		if (!window._ai_elements) {
			throw new Error("No elements extracted. Call extract_page first.");
		}
		
		const el = window._ai_elements[args.id];
		if (!el) {
			throw new Error("Element ID " + args.id + " not found. Total elements: " + window._ai_elements.length);
		}
		
		if (!document.contains(el)) {
			throw new Error("Element ID " + args.id + " is no longer in the DOM. Call extract_page again.");
		}
		
		el.scrollIntoView({block: "center"});
		el.focus();
		el.value = '';
		el.value = args.text;
		el.dispatchEvent(new Event('input', { bubbles: true }));
		el.dispatchEvent(new Event('change', { bubbles: true }));
		el.dispatchEvent(new KeyboardEvent('keyup', { bubbles: true }));
		return true;
	}`, map[string]interface{}{"id": id, "text": text})

	if err != nil {
		return fmt.Errorf("type into element [%d]: %w", id, err)
	}

	if m.config.Debug {
		m.log.Debug("Text entered successfully", "id", id)
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
		m.log.Debug("Clicking element by selector", "selector", selector)
	}

	el, err := m.page.Element(selector)
	if err != nil {
		return fmt.Errorf("element %s not found: %w", selector, err)
	}

	err = el.Click(proto.InputMouseButtonLeft, 1)
	if err != nil {
		return fmt.Errorf("click on %s failed: %w", selector, err)
	}

	if m.config.Debug {
		m.log.Debug("Element clicked successfully", "selector", selector)
	}

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
		return fmt.Errorf("element %s not found: %w", selector, err)
	}

	_ = el.SelectAllText()
	el.MustInput("")
	el.MustInput(text)

	if m.config.Debug {
		m.log.Debug("Text entered successfully", "selector", selector)
	}

	return nil
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

	var scrollScript string
	switch direction {
	case "up":
		scrollScript = `window.scrollBy(0, -400)`
	case "down":
		scrollScript = `window.scrollBy(0, 400)`
	default:
		return fmt.Errorf("invalid scroll direction: %s (use 'up' or 'down')", direction)
	}

	_, err := m.page.Eval(scrollScript)
	if err != nil {
		return fmt.Errorf("scroll failed: %w", err)
	}

	time.Sleep(300 * time.Millisecond)

	if m.config.Debug {
		m.log.Debug("Page scrolled successfully", "direction", direction)
	}

	return nil
}

func (m *Manager) PressKey(ctx context.Context, key string) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("press_key canceled: %w", ctx.Err())
	default:
	}

	inputKey, ok := keyMapping[key]
	if !ok {
		keys := make([]string, 0, len(keyMapping))
		for k := range keyMapping {
			keys = append(keys, k)
		}
		return fmt.Errorf("unsupported key: %s. Supported: %v", key, keys)
	}

	if m.config.Debug {
		m.log.Debug("Pressing keyboard key", "key", key)
	}

	err := m.page.Keyboard.Press(inputKey)
	if err != nil {
		return fmt.Errorf("press key %s failed: %w", key, err)
	}

	time.Sleep(100 * time.Millisecond)

	if m.config.Debug {
		m.log.Debug("Key pressed successfully", "key", key)
	}

	return nil
}

func (m *Manager) GetPage() *rod.Page {
	return m.page
}

func (m *Manager) GetURL() string {
	info, err := m.page.Info()
	if err != nil {
		return ""
	}
	return info.URL
}

func (m *Manager) GetTitle() string {
	info, err := m.page.Info()
	if err != nil {
		return ""
	}
	return info.Title
}

func (m *Manager) Close() error {
	if m.browser != nil {
		m.browser.Close()
	}
	return nil
}

func normalizeURL(url string) string {
	if len(url) == 0 {
		return url
	}

	for len(url) > 0 && url[0] == ' ' {
		url = url[1:]
	}
	for len(url) > 0 && url[len(url)-1] == ' ' {
		url = url[:len(url)-1]
	}

	if len(url) > 0 && !hasProtocol(url) {
		url = "https://" + url
	}

	return url
}

func hasProtocol(url string) bool {
	if len(url) >= 7 && url[:7] == "http://" {
		return true
	}
	if len(url) >= 8 && url[:8] == "https://" {
		return true
	}
	return false
}
