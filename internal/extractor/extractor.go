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

func (e *Extractor) UpdatePage(page *rod.Page) {
	e.page = page
}

func (e *Extractor) Extract(ctx context.Context) (*types.PageState, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	info, err := e.page.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get page info: %w", err)
	}

	// JavaScript для извлечения элементов И контента
	jsCode := `() => {
		window._ai_elements = [];
		
		const results = [];
		
		// === ИНТЕРАКТИВНЫЕ ЭЛЕМЕНТЫ ===
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
			'[role="listitem"]',
			'[role="row"]',
			'[onclick]',
			'[tabindex]:not([tabindex="-1"])',
			'label[for]',
			'summary',
			'tr',
			'li'
		];
		
		const seen = new Set();
		
		selectors.forEach(sel => {
			try {
				document.querySelectorAll(sel).forEach(el => {
					if (seen.has(el)) return;
					seen.add(el);
					
					const rect = el.getBoundingClientRect();
					const style = window.getComputedStyle(el);
					
					const isVisible = 
						rect.width > 0 && 
						rect.height > 0 && 
						style.display !== 'none' && 
						style.visibility !== 'hidden' && 
						parseFloat(style.opacity) > 0;
					
					if (!isVisible) return;
					
					const id = window._ai_elements.length;
					window._ai_elements.push(el);
					
					let text = '';
					if (el.tagName.toLowerCase() === 'input' || el.tagName.toLowerCase() === 'textarea') {
						text = el.value || el.placeholder || '';
					} else {
						text = el.innerText || el.textContent || el.getAttribute('aria-label') || '';
					}
					text = text.trim().replace(/\\s+/g, ' ').substring(0, 150);
					
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
			} catch (e) {}
		});
		
		// === ТЕКСТОВЫЙ КОНТЕНТ СТРАНИЦЫ ===
		let pageContent = [];
		
		// Извлекаем текст из основных контейнеров
		const contentSelectors = [
			'main',
			'article', 
			'[role="main"]',
			'.content',
			'.mail-list',
			'.inbox',
			'.messages',
			'.letter-list',
			'table tbody',
			'ul',
			'ol'
		];
		
		// Ищем списки писем (типичные паттерны почтовых сервисов)
		const mailPatterns = [
			// Mail.ru
			'.letter-list .letter-list-item',
			'.dataset__items .dataset__item',
			'.llc',  // letter list container
			// Gmail
			'.zA',
			'[role="row"]',
			// Yandex
			'.mail-MessageSnippet',
			'.ns-view-messages'
		];
		
		let mailItems = [];
		for (const pattern of mailPatterns) {
			try {
				const items = document.querySelectorAll(pattern);
				if (items.length > 0) {
					items.forEach((item, idx) => {
						if (idx < 15) { // Берём первые 15 писем
							const text = item.innerText || item.textContent || '';
							const cleanText = text.trim().replace(/\\s+/g, ' ').substring(0, 300);
							if (cleanText.length > 10) {
								mailItems.push({
									index: idx + 1,
									content: cleanText
								});
							}
						}
					});
					if (mailItems.length > 0) break;
				}
			} catch(e) {}
		}
		
		// Если не нашли по паттернам, берём общий контент
		if (mailItems.length === 0) {
			try {
				// Пробуем найти любые повторяющиеся элементы (строки списка)
				const rows = document.querySelectorAll('div[class*="item"], div[class*="row"], div[class*="message"], div[class*="letter"], li');
				const uniqueTexts = new Set();
				rows.forEach((row, idx) => {
					if (idx < 20) {
						const text = (row.innerText || '').trim().replace(/\\s+/g, ' ');
						if (text.length > 20 && text.length < 500 && !uniqueTexts.has(text)) {
							uniqueTexts.add(text);
							pageContent.push(text.substring(0, 300));
						}
					}
				});
			} catch(e) {}
		}
		
		// Определяем модальные окна
		let hasModal = false;
		const modalSelectors = [
			'[role="dialog"]',
			'[role="alertdialog"]',
			'[aria-modal="true"]',
			'.modal:not([style*="display: none"])',
			'.popup:not([style*="display: none"])'
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
			totalElements: window._ai_elements.length,
			mailItems: mailItems,
			pageContent: pageContent.slice(0, 15)
		};
	}`

	res, err := e.page.Eval(jsCode)
	if err != nil {
		return nil, fmt.Errorf("JS extraction failed: %w", err)
	}

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
		MailItems     []struct {
			Index   int    `json:"index"`
			Content string `json:"content"`
		} `json:"mailItems"`
		PageContent []string `json:"pageContent"`
	}

	jsonStr := res.Value.JSON("", "")
	if err := json.Unmarshal([]byte(jsonStr), &jsResult); err != nil {
		return nil, fmt.Errorf("failed to parse JS result: %w", err)
	}

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

	// Сохраняем контент
	var contentParts []string
	for _, item := range jsResult.MailItems {
		contentParts = append(contentParts, fmt.Sprintf("%d. %s", item.Index, item.Content))
	}
	if len(contentParts) == 0 {
		contentParts = jsResult.PageContent
	}
	pageState.Content = strings.Join(contentParts, "\n")

	if e.logger != nil {
		e.logger.Debug("Extracted elements", "count", len(pageState.Elements), "hasModal", pageState.HasModal, "contentItems", len(contentParts))
	}

	return pageState, nil
}

func (e *Extractor) FormatForLLM(state *types.PageState) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("## Page: %s\n", state.Title))
	b.WriteString(fmt.Sprintf("## URL: %s\n\n", state.URL))

	if state.HasModal {
		b.WriteString("⚠️ **MODAL/POPUP DETECTED** - Close it first with Escape or find close button.\n\n")
	}

	// === КОНТЕНТ СТРАНИЦЫ (важно для почты!) ===
	if state.Content != "" {
		b.WriteString("### Page Content (emails, messages, items):\n")
		b.WriteString("```\n")
		b.WriteString(state.Content)
		b.WriteString("\n```\n\n")
	}

	// Группируем элементы
	var inputs, buttons, links, listItems []types.PageElement

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
		case "tr", "li", "div":
			if el.Attributes["role"] == "row" || el.Attributes["role"] == "listitem" {
				if el.Text != "" && len(el.Text) > 10 {
					listItems = append(listItems, el)
				}
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
		limit := 15
		for i, el := range buttons {
			if i >= limit {
				b.WriteString(fmt.Sprintf("... and %d more buttons\n", len(buttons)-limit))
				break
			}
			b.WriteString(e.formatElement(el) + "\n")
		}
		b.WriteString("\n")
	}

	// List items (для почты важно!)
	if len(listItems) > 0 {
		b.WriteString("### List Items (emails/messages)\n")
		limit := 15
		for i, el := range listItems {
			if i >= limit {
				break
			}
			b.WriteString(e.formatElement(el) + "\n")
		}
		b.WriteString("\n")
	}

	// Links
	if len(links) > 0 {
		b.WriteString("### Links\n")
		limit := 20
		for i, el := range links {
			if i >= limit {
				b.WriteString(fmt.Sprintf("... and %d more links\n", len(links)-limit))
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

	parts = append(parts, fmt.Sprintf("[%d] %s", el.ID, el.Tag))

	if el.Text != "" {
		text := el.Text
		if len(text) > 80 {
			text = text[:80] + "..."
		}
		parts = append(parts, fmt.Sprintf("%q", text))
	}

	if ph, ok := el.Attributes["placeholder"]; ok && ph != "" {
		if len(ph) > 30 {
			ph = ph[:30] + "..."
		}
		parts = append(parts, fmt.Sprintf("placeholder=%q", ph))
	}

	if t, ok := el.Attributes["type"]; ok && t != "" && t != "text" {
		parts = append(parts, fmt.Sprintf("type=%s", t))
	}

	if role, ok := el.Attributes["role"]; ok && role != "" {
		parts = append(parts, fmt.Sprintf("role=%s", role))
	}

	return strings.Join(parts, " ")
}
