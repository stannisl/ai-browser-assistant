package types

import "time"

type PageElement struct {
	Selector    string
	Text        string
	Tag         string
	Attributes  map[string]string
	Clickable   bool
	Visible     bool
	Position    struct {
		X      int
		Y      int
		Width  int
		Height int
	}
	DiscoveryTime time.Time
}

type PageState struct {
	Title      string
	URL        string
	Elements   []PageElement
	Scripts    []string
	Forms      []FormElement
	Links      []LinkElement
	Timestamp  time.Time
	IsLoading  bool
	ScrollY    int
	Viewport   struct {
		Width  int
		Height int
	}
}

type FormElement struct {
	ID          string
	Name        string
	Selector    string
	Inputs      []InputField
	SubmitBtn   *SubmitButton
	IsComplete  bool
}

type InputField struct {
	ID       string
	Name     string
	Type     string
	Selector string
	Required bool
	Placeholder string
}

type SubmitButton struct {
	ID       string
	Name     string
	Selector string
	Type     string
	Visible  bool
	Enabled  bool
}

type LinkElement struct {
	ID          string
	Href        string
	Text        string
	Selector    string
	Visible     bool
	Clickable   bool
	Rel         string
	Title       string
}

type BrowserConfig struct {
	Headless      bool
	UserDataDir   string
	Timeout       time.Duration
	Viewport      struct {
		Width  int
		Height int
	}
	Incognito  bool
	Debug      bool
}
