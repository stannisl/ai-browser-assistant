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
		Title:     title,
		URL:       url,
		Elements:  []types.PageElement{},
		Timestamp: time.Now(),
		IsLoading: false,
		ScrollY:   0,
		Viewport: struct {
			Width  int
			Height int
		}{
			Width:  1920,
			Height: 1080,
		},
		HasModal: result.HasModal,
	}

	for _, elem := range result.Elements {
		selectorID := e.counter
		e.selectors.Store(selectorID, elem.Selector)

		pageElement := types.PageElement{
			Selector: elem.Selector,
			Text:     elem.Text,
			Tag:      elem.Type,
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

		pageState.Elements = append(pageState.Elements, pageElement)
		e.counter++
	}

	pageState.Scripts = []string{}
	pageState.Forms = []types.FormElement{}
	pageState.Links = []types.LinkElement{}

	if e.logger != nil {
		e.logger.Extract(url, len(pageState.Elements))
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

	builder.WriteString("[Interactive Elements]\n")
	for i, elem := range state.Elements {
		line := fmt.Sprintf("[%d] %s", i, elem.Tag)

		if len(elem.Text) > 0 {
			line += fmt.Sprintf(" \"%s\"", elem.Text)
		}
		if len(elem.Attributes["placeholder"]) > 0 {
			line += fmt.Sprintf(" placeholder=\"%s\"", elem.Attributes["placeholder"])
		}
		if len(elem.Attributes["aria-label"]) > 0 && len(elem.Text) == 0 {
			line += fmt.Sprintf(" aria-label=\"%s\"", elem.Attributes["aria-label"])
		}
		if len(elem.Attributes["type"]) > 0 && elem.Tag == "input" {
			line += fmt.Sprintf(" type=\"%s\"", elem.Attributes["type"])
		}
		if elem.Tag == "a" && len(elem.Attributes["href"]) > 0 {
			line += fmt.Sprintf(" → %s", elem.Attributes["href"])
		}

		builder.WriteString(line + "\n")
	}

	builder.WriteString(fmt.Sprintf("\n[Page Info]\nElements: %d\n", len(state.Elements)))

	return builder.String()
}

func boolToString(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
