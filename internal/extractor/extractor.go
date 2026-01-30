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
	// ОБНОВЛЕННАЯ ВЕРСИЯ
	jsCode := `() => {
		window._ai_elements = [];
		const results = [];
		
		// === 1. ИНТЕРАКТИВНЫЕ ЭЛЕМЕНТЫ ===
		const selectors = [
			'a[href]',
			'button',
			'input', // Убрали not(hidden), проверим видимость в коде
			'textarea',
			'select',
			'[role="button"]',
			'[role="checkbox"]',
			'[role="link"]',
			'[role="menuitem"]',
			'[role="tab"]',
			'[onclick]',
			'[title]', 
			'[data-title-shortcut]',
			// Специфичные классы для почтовиков (чекбоксы и кнопки)
			'.mail-MessageSnippet-Checkbox', // Yandex checkbox
			'.checkbox__box',                // General UI checkbox
			'.checkbox__control',
			'[class*="button"]',
			'[class*="btn"]',
			'label'
		];
		
		const seen = new Set();
		
		selectors.forEach(sel => {
			try {
				document.querySelectorAll(sel).forEach(el => {
					if (el === document.body || el === document.documentElement) return;
					if (seen.has(el)) return;
					
					const rect = el.getBoundingClientRect();
					const style = window.getComputedStyle(el);
					
					// ХИТРАЯ ПРОВЕРКА ВИДИМОСТИ
					// Некоторые чекбоксы (input) имеют opacity 0, но лежат поверх видимого элемента
					let isVisible = 
						rect.width > 0 && 
						rect.height > 0 && 
						style.display !== 'none' && 
						style.visibility !== 'hidden';

					// Если это не input, требуем непрозрачность
					if (el.tagName.toLowerCase() !== 'input' && !el.classList.contains('checkbox__control')) {
						if (parseFloat(style.opacity) < 0.1) isVisible = false;
					}
					
					if (!isVisible) return;

					seen.add(el);
					const id = window._ai_elements.length;
					window._ai_elements.push(el);
					
					// === ИЗВЛЕЧЕНИЕ ТЕКСТА ===
					let text = '';
					
					// 1. Значения полей
					if (el.tagName.toLowerCase() === 'input' || el.tagName.toLowerCase() === 'textarea') {
						text = el.value || el.placeholder || '';
					} 
					// 2. Текст внутри
					else {
						text = el.innerText || el.textContent || '';
					}

					// 3. Атрибуты (title, aria)
					if (!text.trim()) {
						text = el.getAttribute('title') || 
							   el.getAttribute('aria-label') || 
							   el.getAttribute('data-title-shortcut') || '';
					}

					// 4. ЕСЛИ ЭТО ЧЕКБОКС БЕЗ ТЕКСТА (ВАЖНО!)
					// Пытаемся найти тему письма рядом, чтобы ЛЛМ поняла "Чекбокс для письма X"
					const isCheckbox = el.getAttribute('role') === 'checkbox' || 
									   el.classList.contains('mail-MessageSnippet-Checkbox') ||
									   el.type === 'checkbox';
					
					if (isCheckbox && !text) {
						// Ищем родительскую строку таблицы/списка
						const row = el.closest('[role="row"], .mail-MessageSnippet, .letter-list-item-content');
						if (row) {
							// Ищем тему или отправителя в этой строке
							const subject = row.querySelector('.mail-MessageSnippet-Item_subject, .ll-sj, .bog');
							if (subject) text = "Выбрать: " + subject.innerText;
							else text = "Чекбокс выбора";
						} else {
							text = "Чекбокс";
						}
					}
					
					text = text.trim().replace(/\s+/g, ' ').substring(0, 150);
					
					// Фильтрация мусора (пустые span/div без роли)
					const tag = el.tagName.toLowerCase();
					if (!text && !['input', 'select', 'textarea', 'button'].includes(tag) && !isCheckbox) {
						// Если нет текста и это не кнопка/инпут - пропускаем, если нет вложенного SVG с title
						const svgTitle = el.querySelector('svg title');
						if (svgTitle) text = svgTitle.textContent.trim();
						else return; 
					}
					
					// Определяем тип для ЛЛМ
					let role = el.getAttribute('role') || '';
					if (isCheckbox) role = 'checkbox';
					
					results.push({
						id: id,
						tag: tag,
						text: text,
						type: el.type || '',
						title: el.getAttribute('title') || '',
						role: role,
						// Маркер, что это похоже на чекбокс
						isCheckbox: isCheckbox
					});
				});
			} catch (e) {}
		});
		
		// === ТЕКСТОВЫЙ КОНТЕНТ (Остался прежним) ===
		let pageContent = [];
		let mailItems = [];
		
		// Сбор контента (упрощено для примера, используй логику из предыдущего ответа для контента)
		const mailPatterns = ['.mail-MessageSnippet', '.letter-list-item', '.zA', '[role="row"]'];
		for (const pattern of mailPatterns) {
			const items = document.querySelectorAll(pattern);
			if (items.length > 0) {
				items.forEach((item, idx) => {
					if (idx < 15) {
						const t = item.innerText.replace(/\s+/g, ' ').substring(0, 200);
						if(t.length > 10) mailItems.push({index: idx+1, content: t});
					}
				});
				if(mailItems.length) break;
			}
		}
		
		// Fallback content
		if (!mailItems.length) {
			document.querySelectorAll('h1, h2, .letter-body, .article').forEach(el => {
				const t = el.innerText.replace(/\s+/g, ' ');
				if(t.length > 20) pageContent.push(t.substring(0,500));
			});
		}

		let hasModal = !!document.querySelector('[role="dialog"], .modal');
		
		return {
			elements: results,
			hasModal: hasModal,
			totalElements: window._ai_elements.length,
			mailItems: mailItems,
			pageContent: pageContent
		};
	}`

	res, err := e.page.Eval(jsCode)
	if err != nil {
		return nil, fmt.Errorf("JS extraction failed: %w", err)
	}

	// Структура результата JS
	var jsResult struct {
		Elements []struct {
			ID         int    `json:"id"`
			Tag        string `json:"tag"`
			Text       string `json:"text"`
			Type       string `json:"type"`
			Href       string `json:"href"`
			Title      string `json:"title"` // Добавили Title
			Role       string `json:"role"`
			IsButton   bool   `json:"isButton"`
			IsCheckbox bool   `json:"isCheckbox"`
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
		if elem.Title != "" {
			attrs["title"] = elem.Title
		}
		if elem.Type != "" {
			attrs["type"] = elem.Type
		}
		if elem.Role != "" {
			attrs["role"] = elem.Role
		}

		// Улучшаем отображение тега для ЛЛМ
		tag := elem.Tag
		if elem.IsButton || elem.Role == "button" || strings.Contains(strings.ToLower(elem.Text), "удалить") {
			tag = "button" // Подменяем tag для ЛЛМ, чтобы он понимал, что это кнопка
		}

		pe := types.PageElement{
			ID:         elem.ID,
			Tag:        tag,
			Text:       elem.Text,
			Attributes: attrs,
			Visible:    true,
		}
		pageState.Elements = append(pageState.Elements, pe)
	}

	// Сохраняем контент
	var contentParts []string
	// Сначала текст открытого письма
	if len(jsResult.PageContent) > 0 {
		contentParts = append(contentParts, "--- OPENED CONTENT ---")
		contentParts = append(contentParts, jsResult.PageContent...)
	}
	// Потом список писем
	if len(jsResult.MailItems) > 0 {
		contentParts = append(contentParts, "--- LIST ITEMS ---")
		for _, item := range jsResult.MailItems {
			contentParts = append(contentParts, fmt.Sprintf("%d. %s", item.Index, item.Content))
		}
	}

	pageState.Content = strings.Join(contentParts, "\n")

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
		b.WriteString("⚠️ **MODAL/POPUP DETECTED** - Close it first with Escape or find close button.\n\n")
	}

	// === КОНТЕНТ СТРАНИЦЫ ===
	// Важно показывать контент ПЕРЕД элементами, чтобы ЛЛМ понимала контекст
	if state.Content != "" {
		b.WriteString("### Page Content (emails, messages, text):\n")
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
			if el.Text != "" {
				links = append(links, el)
			}
		default:
			if val, ok := el.Attributes["title"]; ok && val != "" {
				buttons = append(buttons, el)

			} else if val, ok := el.Attributes["role"]; ok && val == "checkbox" {
				buttons = append(buttons, el)
			} else {
				// Остальное
				if el.Text != "" && len(el.Text) > 20 {
					listItems = append(listItems, el)
				}
			}
		}
	}

	// Buttons (Самое важное для действий)
	if len(buttons) > 0 {
		b.WriteString("### Clickable Elements (Buttons/Tools)\n")
		for _, el := range buttons {
			b.WriteString(e.formatElement(el) + "\n")
		}
		b.WriteString("\n")
	}

	// Inputs
	if len(inputs) > 0 {
		b.WriteString("### Input Fields\n")
		for _, el := range inputs {
			b.WriteString(e.formatElement(el) + "\n")
		}
		b.WriteString("\n")
	}

	// Links
	if len(links) > 0 {
		b.WriteString("### Navigation Links\n")
		// Показываем меньше ссылок, чтобы не забивать контекст, если есть кнопки
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

	b.WriteString(fmt.Sprintf("Total interactive elements: %d\n", state.ElementCount))

	return b.String()
}

func (e *Extractor) formatElement(el types.PageElement) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("[%d]", el.ID))

	// Если это "фейковая" кнопка (span/div), пишем button для понятности агенту
	if el.Tag == "button" {
		parts = append(parts, "button")
	} else {
		parts = append(parts, el.Tag)
	}

	if el.Attributes["role"] == "checkbox" {
		parts = append(parts, "[CHECKBOX]")
	}

	// Главное - текст
	if el.Text != "" {
		text := el.Text
		if len(text) > 50 {
			text = text[:50] + "..."
		}
		parts = append(parts, fmt.Sprintf("%q", text))
	}

	// Если есть title, обязательно показываем (там "Удалить", "Ответить")
	if title, ok := el.Attributes["title"]; ok && title != "" && title != el.Text {
		parts = append(parts, fmt.Sprintf("title=%q", title))
	}

	if ph, ok := el.Attributes["placeholder"]; ok && ph != "" {
		parts = append(parts, fmt.Sprintf("placeholder=%q", ph))
	}

	return strings.Join(parts, " ")
}
