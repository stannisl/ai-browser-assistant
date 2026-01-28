package extractor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/stannisl/ai-browser-assistant/internal/logger"
	"github.com/stannisl/ai-browser-assistant/internal/types"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

type Extractor struct {
	page      *rod.Page
	selectors sync.Map
	counter   int
	logger    *logger.Logger
}

func New(page *rod.Page, log *logger.Logger) *Extractor {
	return &Extractor{
		page:   page,
		logger: log,
	}
}

func (e *Extractor) Extract(ctx context.Context) (*types.PageState, error) {
	e.counter = 0
	e.selectors = sync.Map{}

	var url string
	if info, err := e.page.Info(); err == nil && info != nil {
		url = info.URL
	} else {
		url = e.page.MustElement("html").MustProperty("location.href").String()
	}
	title := e.page.MustElement("html").MustProperty("title").String()

	jsCode := `() => {
    const results = [];
    const selectors = [
        'button', 'a[href]', 'input', 'select', 'textarea',
        '[role="button"]', '[onclick]', '[role="link"]',
        '[role="menuitem"]', '[role="tab"]', '[type="submit"]'
    ];
    
    const seen = new Set();
    
    selectors.forEach(sel => {
        document.querySelectorAll(sel).forEach(el => {
            if (seen.has(el)) return;
            seen.add(el);
            
            const rect = el.getBoundingClientRect();
            const style = window.getComputedStyle(el);
            
            const isVisible = rect.width > 0 && 
                             rect.height > 0 && 
                             style.display !== 'none' &&
                             style.visibility !== 'hidden' &&
                             parseFloat(style.opacity) > 0;
            
            if (!isVisible) return;
            
            let selector = '';
            if (el.id) {
                selector = '#' + el.id;
            } else {
                const path = [];
                let current = el;
                while (current && current !== document.body) {
                    let index = 1;
                    let sibling = current.previousElementSibling;
                    while (sibling) {
                        if (sibling.tagName === current.tagName) index++;
                        sibling = sibling.previousElementSibling;
                    }
                    path.unshift(current.tagName.toLowerCase() + ':nth-of-type(' + index + ')');
                    current = current.parentElement;
                }
                selector = 'body > ' + path.join(' > ');
            }
            
            results.push({
                type: el.tagName.toLowerCase(),
                text: (el.innerText || el.value || '').slice(0, 100).trim(),
                placeholder: el.placeholder || '',
                ariaLabel: el.getAttribute('aria-label') || '',
                href: el.href || '',
                inputType: el.type || '',
                selector: selector
            });
        });
    });
    
    const hasModal = !!(
        document.querySelector('[role="dialog"]') ||
        document.querySelector('[role="alertdialog"]') ||
        document.querySelector('.modal.show') ||
        document.querySelector('[class*="modal"][class*="open"]')
    );
    
    return { elements: results, hasModal: hasModal };
}`

	res := e.page.MustEval(jsCode)

	var result struct {
		Elements []struct {
			Type        string `json:"type"`
			Text        string `json:"text"`
			Placeholder string `json:"placeholder"`
			AriaLabel   string `json:"ariaLabel"`
			Href        string `json:"href"`
			InputType   string `json:"inputType"`
			Selector    string `json:"selector"`
		} `json:"elements"`
		HasModal bool `json:"hasModal"`
	}

	err := json.Unmarshal([]byte(res.JSON("", "")), &result)
	if err != nil {
		return nil, fmt.Errorf("JSON unmarshal failed: %w", err)
	}

	pageState := &types.PageState{
		Title:        title,
		URL:          url,
		Elements:     []types.PageElement{},
		ElementCount: 0,
		InputCount:   0,
		ButtonCount:  0,
		LinkCount:    0,
		Timestamp:    time.Now(),
		IsLoading:    false,
		ScrollY:      0,
		Viewport: struct {
			Width  int
			Height int
		}{
			Width:  1920,
			Height: 1080,
		},
		HasModal: result.HasModal,
	}

	const maxElements = 50
	var elements []types.PageElement
	elementsRaw := result.Elements

	for i, elem := range elementsRaw {
		if i >= maxElements {
			break
		}
		selectorID := e.counter
		e.selectors.Store(selectorID, elem.Selector)

		pageElement := types.PageElement{
			ID:         i,
			Selector:   elem.Selector,
			Text:       elem.Text,
			Tag:        elem.Type,
			Attributes: map[string]string{
				"placeholder": elem.Placeholder,
				"aria-label":  elem.AriaLabel,
				"type":        elem.InputType,
				"href":        elem.Href,
			},
			Clickable:     true,
			Visible:       true,
			DiscoveryTime: time.Now(),
		}

		if elem.Type == "a" && len(elem.Href) > 0 {
			pageElement.Attributes["href"] = elem.Href
		}

		pageState.ElementCount++
		if elem.Type == "input" || elem.Type == "textarea" || elem.Type == "select" {
			pageState.InputCount++
		}
		if elem.Type == "button" || strings.Contains(elem.AriaLabel, "submit") {
			pageState.ButtonCount++
		}
		if elem.Type == "a" {
			pageState.LinkCount++
		}

		elements = append(elements, pageElement)
		e.counter++
	}

	if len(elementsRaw) > maxElements {
		pageState.Content = fmt.Sprintf("Showing %d of %d elements", len(elements), len(elementsRaw))
	} else {
		pageState.Content = fmt.Sprintf("Showing %d elements", pageState.ElementCount)
	}

	pageState.Elements = elements

	pageState.Scripts = []string{}
	pageState.Forms = []types.FormElement{}
	pageState.Links = []types.LinkElement{}

	if e.logger != nil {
		e.logger.Extract(url, pageState.ElementCount)
	}

	return pageState, nil
}

func (e *Extractor) GetSelector(id int) (string, error) {
	value, ok := e.selectors.Load(id)
	if !ok {
		return "", types.ErrElementNotFound
	}
	return value.(string), nil
}

func (e *Extractor) FormatForLLM(state *types.PageState) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Page: %s\n", state.Title))
	builder.WriteString(fmt.Sprintf("URL: %s\n\n", state.URL))

	if state.HasModal {
		builder.WriteString("[!!! MODAL WINDOW ACTIVE !!!]\n\n")
	}

	var inputs, buttons, links, other []types.PageElement

	for _, el := range state.Elements {
		switch {
		case el.Tag == "input" || el.Tag == "textarea" || el.Tag == "select":
			state.InputCount++
			inputs = append(inputs, el)
		case el.Tag == "button" || strings.Contains(el.Attributes["aria-label"], "submit"):
			state.ButtonCount++
			buttons = append(buttons, el)
		case el.Tag == "a":
			state.LinkCount++
			links = append(links, el)
		default:
			other = append(other, el)
		}
	}

	builder.WriteString("[Input Fields]\n")
	for _, el := range inputs {
		builder.WriteString(e.formatElement(el) + "\n")
	}

	builder.WriteString("\n[Buttons]\n")
	for _, el := range buttons {
		builder.WriteString(e.formatElement(el) + "\n")
	}

	builder.WriteString("\n[Links - first 20]\n")
	for i, el := range links {
		if i >= 20 {
			break
		}
		builder.WriteString(e.formatElement(el) + "\n")
	}

	builder.WriteString(fmt.Sprintf("\n[Page Info]\nTotal: %d | Shown: %d\n", state.ElementCount, len(inputs)+len(buttons)+min(len(links), 20)))

	return builder.String()
}

func (e *Extractor) formatElement(el types.PageElement) string {
	line := fmt.Sprintf("[%d] %s", el.ID, el.Tag)

	if len(el.Text) > 0 {
		line += fmt.Sprintf(" \"%s\"", truncate(el.Text, 40))
	}
	if len(el.Attributes["placeholder"]) > 0 {
		line += fmt.Sprintf(" placeholder=\"%s\"", truncate(el.Attributes["placeholder"], 25))
	}
	if el.Tag == "a" && strings.HasPrefix(el.Attributes["href"], "http") {
		line += fmt.Sprintf(" → %s", truncate(el.Attributes["href"], 50))
	}

	return line
}

func boolToString(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
