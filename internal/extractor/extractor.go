package extractor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/stannisl/ai-browser-assistant/internal/logger"
	"github.com/stannisl/ai-browser-assistant/internal/types"
)

type Extractor struct {
	page   *rod.Page
	logger *logger.Logger
}

func New(page *rod.Page, log *logger.Logger) *Extractor {
	return &Extractor{
		page:   page,
		logger: log,
	}
}

func (e *Extractor) Extract(ctx context.Context) (*types.PageState, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Получаем информацию о странице
	info, err := e.page.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get page info: %w", err)
	}

	// JavaScript для извлечения элементов
	jsCode := `() => {
		// Очищаем и инициализируем глобальный массив элементов
		window._ai_elements = [];
		
		const results = [];
		
		// Селекторы для поиска интерактивных элементов
		const selectors = [
			'a[href]',
			'button',
			'input:not([type="hidden"])',
			'textarea',
			'select',
			'[role="button"]',
			'[role="link"]',
			'[role="menuitem"]',
			'[role="tab"]',
			'[role="checkbox"]',
			'[role="radio"]',
			'[onclick]',
			'[tabindex]:not([tabindex="-1"])',
			'label[for]',
			'summary'
		];
		
		const seen = new Set();
		
		selectors.forEach(sel => {
			try {
				document.querySelectorAll(sel).forEach(el => {
					if (seen.has(el)) return;
					seen.add(el);
					
					// Проверка видимости
					const rect = el.getBoundingClientRect();
					const style = window.getComputedStyle(el);
					
					const isVisible = 
						rect.width > 0 && 
						rect.height > 0 && 
						style.display !== 'none' && 
						style.visibility !== 'hidden' && 
						parseFloat(style.opacity) > 0;
					
					if (!isVisible) return;
					
					// Сохраняем элемент в глобальный массив
					const id = window._ai_elements.length;
					window._ai_elements.push(el);
					
					// Получаем текст элемента
					let text = '';
					if (el.tagName.toLowerCase() === 'input' || el.tagName.toLowerCase() === 'textarea') {
						text = el.value || el.placeholder || '';
					} else {
						text = el.innerText || el.textContent || el.getAttribute('aria-label') || '';
					}
					text = text.trim().substring(0, 100);
					
					// Получаем тип для input
					let inputType = '';
					if (el.tagName.toLowerCase() === 'input') {
						inputType = el.type || 'text';
					}
					
					results.push({
						id: id,
						tag: el.tagName.toLowerCase(),
						text: text,
						type: inputType,
						href: el.href || '',
						placeholder: el.placeholder || '',
						ariaLabel: el.getAttribute('aria-label') || '',
						name: el.name || '',
						role: el.getAttribute('role') || ''
					});
				});
			} catch (e) {
				// Игнорируем ошибки для отдельных селекторов
			}
		});
		
		// Определяем наличие модального окна
		let hasModal = false;
		const modalSelectors = [
			'[role="dialog"]',
			'[role="alertdialog"]',
			'[aria-modal="true"]',
			'.modal:not([style*="display: none"])',
			'.popup:not([style*="display: none"])',
			'[class*="modal"]:not([style*="display: none"])'
		];
		
		for (const sel of modalSelectors) {
			try {
				const modal = document.querySelector(sel);
				if (modal) {
					const style = window.getComputedStyle(modal);
					if (style.display !== 'none' && style.visibility !== 'hidden') {
						hasModal = true;
						break;
					}
				}
			} catch (e) {}
		}
		
		return {
			elements: results,
			hasModal: hasModal,
			totalElements: window._ai_elements.length
		};
	}`

	res, err := e.page.Eval(jsCode)
	if err != nil {
		return nil, fmt.Errorf("JS extraction failed: %w", err)
	}

	// Парсим результат
	var jsResult struct {
		Elements []struct {
			ID          int    `json:"id"`
			Tag         string `json:"tag"`
			Text        string `json:"text"`
			Type        string `json:"type"`
			Href        string `json:"href"`
			Placeholder string `json:"placeholder"`
			AriaLabel   string `json:"ariaLabel"`
			Name        string `json:"name"`
			Role        string `json:"role"`
		} `json:"elements"`
		HasModal      bool `json:"hasModal"`
		TotalElements int  `json:"totalElements"`
	}

	jsonStr := res.Value.JSON("", "")
	if err := json.Unmarshal([]byte(jsonStr), &jsResult); err != nil {
		return nil, fmt.Errorf("failed to parse JS result: %w", err)
	}

	// Формируем PageState
	pageState := &types.PageState{
		Title:        info.Title,
		URL:          info.URL,
		HasModal:     jsResult.HasModal,
		ElementCount: jsResult.TotalElements,
		Timestamp:    time.Now(),
	}

	// Конвертируем элементы
	for _, elem := range jsResult.Elements {
		attrs := map[string]string{}
		if elem.Href != "" {
			attrs["href"] = elem.Href
		}
		if elem.Placeholder != "" {
			attrs["placeholder"] = elem.Placeholder
		}
		if elem.AriaLabel != "" {
			attrs["aria-label"] = elem.AriaLabel
		}
		if elem.Name != "" {
			attrs["name"] = elem.Name
		}
		if elem.Type != "" {
			attrs["type"] = elem.Type
		}
		if elem.Role != "" {
			attrs["role"] = elem.Role
		}

		pe := types.PageElement{
			ID:         elem.ID,
			Tag:        elem.Tag,
			Text:       elem.Text,
			Attributes: attrs,
			Visible:    true,
		}
		pageState.Elements = append(pageState.Elements, pe)
	}

	if e.logger != nil {
		e.logger.Debug("Extracted elements", "count", len(pageState.Elements), "hasModal", pageState.HasModal)
	}

	return pageState, nil
}

func (e *Extractor) FormatForLLM(state *types.PageState) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("## Page: %s\n", state.Title))
	b.WriteString(fmt.Sprintf("## URL: %s\n\n", state.URL))

	if state.HasModal {
		b.WriteString("⚠️ **MODAL/POPUP DETECTED** - Focus on modal elements first. Close with Escape or find close button.\n\n")
	}

	// Группируем элементы по типу
	var inputs, buttons, links, other []types.PageElement

	for _, el := range state.Elements {
		switch el.Tag {
		case "input", "textarea", "select":
			inputs = append(inputs, el)
		case "button":
			buttons = append(buttons, el)
		case "a":
			if el.Text != "" || el.Attributes["aria-label"] != "" {
				links = append(links, el)
			}
		default:
			if el.Text != "" {
				other = append(other, el)
			}
		}
	}

	// Inputs
	if len(inputs) > 0 {
		b.WriteString("### Input Fields\n")
		for _, el := range inputs {
			b.WriteString(e.formatElement(el) + "\n")
		}
		b.WriteString("\n")
	}

	// Buttons
	if len(buttons) > 0 {
		b.WriteString("### Buttons\n")
		for _, el := range buttons {
			b.WriteString(e.formatElement(el) + "\n")
		}
		b.WriteString("\n")
	}

	// Links (ограничиваем количество)
	if len(links) > 0 {
		b.WriteString("### Links\n")
		limit := 30
		for i, el := range links {
			if i >= limit {
				b.WriteString(fmt.Sprintf("... and %d more links\n", len(links)-limit))
				break
			}
			b.WriteString(e.formatElement(el) + "\n")
		}
		b.WriteString("\n")
	}

	// Other interactive elements
	if len(other) > 0 {
		b.WriteString("### Other Interactive Elements\n")
		limit := 15
		for i, el := range other {
			if i >= limit {
				break
			}
			b.WriteString(e.formatElement(el) + "\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(fmt.Sprintf("Total interactive elements: %d\n", state.ElementCount))

	return b.String()
}

func (e *Extractor) formatElement(el types.PageElement) string {
	var parts []string

	// ID и тег
	parts = append(parts, fmt.Sprintf("[%d] %s", el.ID, el.Tag))

	// Текст
	if el.Text != "" {
		text := el.Text
		if len(text) > 50 {
			text = text[:50] + "..."
		}
		parts = append(parts, fmt.Sprintf("%q", text))
	}

	// Placeholder для input
	if ph, ok := el.Attributes["placeholder"]; ok && ph != "" {
		if len(ph) > 30 {
			ph = ph[:30] + "..."
		}
		parts = append(parts, fmt.Sprintf("placeholder=%q", ph))
	}

	// Тип для input
	if t, ok := el.Attributes["type"]; ok && t != "" && t != "text" {
		parts = append(parts, fmt.Sprintf("type=%s", t))
	}

	// Role если есть
	if role, ok := el.Attributes["role"]; ok && role != "" {
		parts = append(parts, fmt.Sprintf("role=%s", role))
	}

	return strings.Join(parts, " ")
}

// Вспомогательные функции
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (e *Extractor) UpdatePage(page *rod.Page) {
	e.page = page
}
