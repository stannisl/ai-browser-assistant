package types

import (
	"testing"
	"time"
)

func TestPageElement(t *testing.T) {
	element := PageElement{
		ID:         1,
		Selector:   "#submit-button",
		Text:       "Submit",
		Tag:        "button",
		Attributes: map[string]string{"class": "btn-primary"},
		Clickable:  true,
		Visible:    true,
		Position: struct {
			X      int
			Y      int
			Width  int
			Height int
		}{
			X:      100,
			Y:      200,
			Width:  120,
			Height: 40,
		},
	}

	if element.ID != 1 {
		t.Errorf("expected ID 1, got %d", element.ID)
	}

	if element.Selector != "#submit-button" {
		t.Errorf("expected Selector '#submit-button', got '%s'", element.Selector)
	}

	if element.Text != "Submit" {
		t.Errorf("expected Text 'Submit', got '%s'", element.Text)
	}

	if element.Tag != "button" {
		t.Errorf("expected Tag 'button', got '%s'", element.Tag)
	}

	if element.Clickable != true {
		t.Error("expected Clickable true")
	}

	if element.Visible != true {
		t.Error("expected Visible true")
	}

	if element.Position.X != 100 {
		t.Errorf("expected Position X 100, got %d", element.Position.X)
	}

	if element.Position.Width != 120 {
		t.Errorf("expected Position Width 120, got %d", element.Position.Width)
	}
}

func TestPageState(t *testing.T) {
	state := PageState{
		Title:        "Test Page",
		URL:          "https://example.com/test",
		Elements:     []PageElement{},
		Scripts:      []string{"script1", "script2"},
		Forms:        []FormElement{},
		Links:        []LinkElement{},
		Timestamp:    time.Now(),
		IsLoading:    false,
		ScrollY:      100,
		Viewport: struct {
			Width  int
			Height int
		}{
			Width:  1920,
			Height: 1080,
		},
		HasModal:   false,
		InputCount: 0,
		ButtonCount: 0,
		LinkCount:   0,
		ElementCount: 0,
		Content:     "Test page content",
	}

	if state.Title != "Test Page" {
		t.Errorf("expected Title 'Test Page', got '%s'", state.Title)
	}

	if state.URL != "https://example.com/test" {
		t.Errorf("expected URL 'https://example.com/test', got '%s'", state.URL)
	}

	if state.IsLoading != false {
		t.Error("expected IsLoading false")
	}

	if state.ScrollY != 100 {
		t.Errorf("expected ScrollY 100, got %d", state.ScrollY)
	}

	if state.Viewport.Width != 1920 {
		t.Errorf("expected Viewport Width 1920, got %d", state.Viewport.Width)
	}

	if state.Viewport.Height != 1080 {
		t.Errorf("expected Viewport Height 1080, got %d", state.Viewport.Height)
	}

	if state.HasModal != false {
		t.Error("expected HasModal false")
	}

	if state.ElementCount != 0 {
		t.Errorf("expected ElementCount 0, got %d", state.ElementCount)
	}
}

func TestPageState_WithElements(t *testing.T) {
	element1 := PageElement{
		ID:         1,
		Selector:   "#btn1",
		Text:       "Button 1",
		Tag:        "button",
		Clickable:  true,
		Visible:    true,
	}

	element2 := PageElement{
		ID:         2,
		Selector:   "#btn2",
		Text:       "Button 2",
		Tag:        "button",
		Clickable:  true,
		Visible:    true,
	}

	state := PageState{
		Elements:    []PageElement{element1, element2},
		Title:       "Test",
		URL:         "https://example.com",
		ElementCount: 2,
	}

	if len(state.Elements) != 2 {
		t.Errorf("expected 2 elements, got %d", len(state.Elements))
	}

	if state.ElementCount != 2 {
		t.Errorf("expected ElementCount 2, got %d", state.ElementCount)
	}

	if state.Elements[0].ID != 1 {
		t.Errorf("expected first element ID 1, got %d", state.Elements[0].ID)
	}
}

func TestFormElement(t *testing.T) {
	form := FormElement{
		ID:         "login-form",
		Name:       "LoginForm",
		Selector:   "#login-form",
		Inputs:     []InputField{},
		SubmitBtn:  nil,
		IsComplete: false,
	}

	if form.ID != "login-form" {
		t.Errorf("expected ID 'login-form', got '%s'", form.ID)
	}

	if form.Name != "LoginForm" {
		t.Errorf("expected Name 'LoginForm', got '%s'", form.Name)
	}

	if form.IsComplete != false {
		t.Error("expected IsComplete false")
	}

	form = FormElement{
		ID:         "complete-form",
		Selector:   "#complete-form",
		SubmitBtn:  &SubmitButton{},
		IsComplete: true,
	}

	if form.IsComplete != true {
		t.Error("expected IsComplete true")
	}

	if form.SubmitBtn == nil {
		t.Error("expected SubmitBtn not to be nil")
	}
}

func TestInputField(t *testing.T) {
	field := InputField{
		ID:          "username",
		Name:        "username",
		Type:        "text",
		Selector:    "#username",
		Required:    true,
		Placeholder: "Enter username",
	}

	if field.ID != "username" {
		t.Errorf("expected ID 'username', got '%s'", field.ID)
	}

	if field.Required != true {
		t.Error("expected Required true")
	}

	if field.Placeholder != "Enter username" {
		t.Errorf("expected Placeholder 'Enter username', got '%s'", field.Placeholder)
	}
}

func TestSubmitButton(t *testing.T) {
	button := SubmitButton{
		ID:       "submit-btn",
		Name:     "submit",
		Selector: "#submit",
		Type:     "submit",
		Visible:  true,
		Enabled:  true,
	}

	if button.ID != "submit-btn" {
		t.Errorf("expected ID 'submit-btn', got '%s'", button.ID)
	}

	if button.Visible != true {
		t.Error("expected Visible true")
	}

	if button.Enabled != true {
		t.Error("expected Enabled true")
	}

	button = SubmitButton{
		ID:       "disabled-btn",
		Selector: "#disabled",
		Enabled:  false,
	}

	if button.Enabled != false {
		t.Error("expected Enabled false")
	}
}

func TestLinkElement(t *testing.T) {
	link := LinkElement{
		ID:        "home-link",
		Href:      "https://example.com",
		Text:      "Home",
		Selector:  "#home-link",
		Visible:   true,
		Clickable: true,
		Rel:       "nofollow",
		Title:     "Go to home page",
	}

	if link.ID != "home-link" {
		t.Errorf("expected ID 'home-link', got '%s'", link.ID)
	}

	if link.Href != "https://example.com" {
		t.Errorf("expected Href 'https://example.com', got '%s'", link.Href)
	}

	if link.Text != "Home" {
		t.Errorf("expected Text 'Home', got '%s'", link.Text)
	}

	if link.Rel != "nofollow" {
		t.Errorf("expected Rel 'nofollow', got '%s'", link.Rel)
	}

	if link.Title != "Go to home page" {
		t.Errorf("expected Title 'Go to home page', got '%s'", link.Title)
	}

	if link.Visible != true {
		t.Error("expected Visible true")
	}
}

func TestBrowserConfig(t *testing.T) {
	config := BrowserConfig{
		Headless:    false,
		UserDataDir: "./user-data",
		Timeout:     30 * time.Second,
		Viewport: struct {
			Width  int
			Height int
		}{
			Width:  1280,
			Height: 720,
		},
		Incognito: true,
		Debug:     true,
	}

	if config.Headless != false {
		t.Error("expected Headless false")
	}

	if config.UserDataDir != "./user-data" {
		t.Errorf("expected UserDataDir './user-data', got '%s'", config.UserDataDir)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("expected Timeout 30s, got %v", config.Timeout)
	}

	if config.Viewport.Width != 1280 {
		t.Errorf("expected Viewport Width 1280, got %d", config.Viewport.Width)
	}

	if config.Incognito != true {
		t.Error("expected Incognito true")
	}

	if config.Debug != true {
		t.Error("expected Debug true")
	}
}

func TestBrowserConfig_HeadlessMode(t *testing.T) {
	config := BrowserConfig{
		Headless:   true,
		Incognito:  false,
		Timeout:    60 * time.Second,
		Debug:      false,
	}

	if config.Headless != true {
		t.Error("expected Headless true")
	}

	if config.Debug != false {
		t.Error("expected Debug false")
	}

	if config.Incognito != false {
		t.Error("expected Incognito false")
	}
}

func TestPageState_Empty(t *testing.T) {
	state := PageState{}

	if state.Title != "" {
		t.Errorf("expected empty Title, got '%s'", state.Title)
	}

	if state.URL != "" {
		t.Errorf("expected empty URL, got '%s'", state.URL)
	}

	if len(state.Elements) != 0 {
		t.Errorf("expected empty Elements, got %d items", len(state.Elements))
	}

	if state.IsLoading != false {
		t.Error("expected IsLoading false (default)")
	}
}

func TestPageElement_ZeroID(t *testing.T) {
	element := PageElement{
		ID:     0,
		Selector: "#test",
		Text:   "Test",
	}

	if element.ID != 0 {
		t.Errorf("expected ID 0, got %d", element.ID)
	}
}

func TestFormElement_NoSubmitButton(t *testing.T) {
	form := FormElement{
		ID:         "form1",
		SubmitBtn:  nil,
		IsComplete: false,
	}

	if form.SubmitBtn != nil {
		t.Error("expected SubmitBtn to be nil")
	}
}

func TestLinkElement_EmptyHref(t *testing.T) {
	link := LinkElement{
		ID:        "broken-link",
		Href:      "",
		Text:      "Click",
		Selector:  "#broken",
		Clickable: true,
	}

	if link.Href != "" {
		t.Errorf("expected empty Href, got '%s'", link.Href)
	}
}
