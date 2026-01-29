# AI Browser Assistant - Project Code Description

## Project Overview

AI Browser Assistant is an autonomous AI agent built in Go for browser automation. It uses Claude 3.5 Sonnet (via OpenAI-compatible API) to dynamically analyze web pages and execute actions without requiring hardcoded selectors, manual step definitions, or predefined workflows.

**Key Capabilities:**
- Autonomous page exploration and decision-making
- Tool calling architecture for browser operations
- Context-aware content extraction to avoid token limits
- Security layer requiring user confirmation for destructive actions
- GUI mode for debugging with visible browser sessions
- Persistent browser sessions across operations
- Works with OpenAI-compatible APIs (z.ai, OpenAI, etc.)

## Architecture

### Technology Stack

- **Language:** Go 1.24
- **Browser Automation:** Rod (Chrome DevTools Protocol wrapper)
- **AI/LLM:** OpenAI-compatible API (default: z.ai's zlm-4.5-flash model)
- **Logging:** Zap structured logging
- **Text Processing:** Langchaingo text splitters for context management
- **Configuration:** Viper for configuration management

### Directory Structure

```
ai-browser-assistant/
├── cmd/
│   └── agent/
│       └── main.go                    # Entry point, signal handling, orchestration
│
├── internal/
│   ├── browser/                       # Browser abstraction layer
│   │   ├── browser.go                 # Browser instance management, connections, pooling
│   │   ├── page.go                    # Page operations (navigate, click, input, extract)
│   │   ├── navigation.go              # Navigation flow, history, back/forward
│   │   ├── element.go                 # Element search, selection, caching
│   │   └── persistence.go             # Persistent sessions, cookies management
│   │
│   ├── agent/                         # Agent orchestration
│   │   ├── agent.go                   # Main agent logic, tool calling loop
│   │   ├── context.go                 # Conversation history, page context
│   │   ├── tools.go                   # Tool definitions, execution, security
│   │   ├── security.go                # Confirmation layer, destructive action warnings
│   │   └── executor.go                # Tool execution dispatcher
│   │
│   ├── extractor/                     # Content extraction strategies
│   │   ├── extractor.go               # Main extractor interface and strategies
│   │   ├── strategies.go              # Extraction implementations (text, links, images)
│   │   ├── context_strategy.go        # Context management for LLM
│   │   ├── token_calculator.go        # Token counting and budgeting
│   │   ├── page_summarizer.go         # Page summarization strategies
│   │   └── link_extractor.go          # Link extraction utilities
│   │
│   ├── llm/                           # LLM integration layer
│   │   ├── client.go                  # OpenAI-compatible client wrapper
│   │   ├── messages.go                # Message formatting and conversion
│   │   └── prompts.go                 # System prompts and templates
│   │
│   ├── types/                         # Shared type definitions
│   │   ├── browser.go                 # Browser-related types
│   │   ├── agent.go                   # Agent and tool types
│   │   ├── extractor.go               # Extraction-related types
│   │   └── errors.go                  # Custom error types
│   │
│   └── logger/                        # Logging infrastructure
│       └── logger.go                  # Zap logger wrapper with context
│
├── pkg/
│   └── utils/                         # Shared utilities
│       └── utils.go                   # Helper functions
│
├── configs/
│   └── config.yaml                    # Configuration file
│
├── bin/                               # Compiled binaries
├── data/                              # Data storage (user profiles, etc.)
├── go.mod                             # Go modules
├── go.sum                             # Dependency checksums
└── README.md                          # Project documentation
```

## Core Components

### 1. Browser Package (`internal/browser/`)

#### Browser Management (`browser.go`)

**Purpose:** Manages browser instances, connections, and lifecycle

**Key Functions:**
```go
type Browser struct {
    browser   *rod.Browser
    pages     []*Page
    context   context.Context
    logger    *zap.Logger
    // ... other fields
}

func NewBrowser(ctx context.Context, logger *zap.Logger) (*Browser, error)
func (b *Browser) Open(url string) (*Page, error)
func (b *Browser) OpenPage() (*Page, error)
func (b *Browser) GetCurrentPage() *Page
func (b *Browser) SetPage(page *Page) error
func (b *Browser) Close() error
func (b *Browser) CloseAllPages() error
```

**Key Behaviors:**
- Creates browser instance using Rod
- Manages page pool (max pages limit)
- Handles browser connection lifecycle
- Maintains browser context for cancellation
- Provides visible browser option for debugging

#### Page Operations (`page.go`)

**Purpose:** Abstracts Rod's page API for automation operations

**Key Functions:**
```go
type Page struct {
    page    *rod.Page
    browser *Browser
    logger  *zap.Logger
    // ... other fields
}

func (p *Page) Navigate(url string) error
func (p *Page) Click(selector string) error
func (p *Page) Input(selector, text string) error
func (p *Page) ExtractText(selector string) (string, error)
func (p *Page) WaitForStable(timeout time.Duration) error
func (p *Page) GetElements(selector string) ([]*rod.Element, error)
func (p *Page) ScrollTo(selector string) error
func (p *Page) TakeScreenshot() (string, error)
```

**Key Behaviors:**
- Navigation with stabilization wait
- Dynamic element selection (no hardcoded selectors)
- Text extraction and content retrieval
- Screenshot capture for debugging
- Error handling and retry logic

#### Navigation (`navigation.go`)

**Purpose:** Handles navigation flows and history

**Key Functions:**
```go
func (p *Page) NavigateWithRetry(url string, maxRetries int) error
func (p *Page) Back() error
func (p *Page) Forward() error
func (p *Page) Refresh() error
func (p *Page) GetCurrentURL() string
func (p *Page) GetPageTitle() string
func (p *Page) SaveState() (*PageState, error)
func (p *Page) RestoreState(state *PageState) error
```

#### Element Search (`element.go`)

**Purpose:** Advanced element search and selection strategies

**Key Functions:**
```go
func (p *Page) FindElements(selector string, limit int) ([]*rod.Element, error)
func (p *Page) FindElementByContent(content string, limit int) (*rod.Element, error)
func (p *Page) FindElementByText(text string, limit int) (*rod.Element, error)
func (p *Page) FindButtonByText(text string) (*rod.Element, error)
func (p *Page) FindLinkByText(text string) (*rod.Element, error)
func (p *Page) FindInputByPlaceholder(placeholder string) (*rod.Element, error)
func (p *Page) FindDropdownOptionByText(text string) (*rod.Element, error)
func (p *Page) CacheElement(selector string, element *rod.Element)
func (p *Page) GetCachedElement(selector string) (*rod.Element, error)
```

**Key Behaviors:**
- Multiple search strategies (CSS, content-based, text-based)
- Element caching to avoid repeated searches
- Type-specific searches (buttons, links, inputs, dropdowns)
- Content extraction with safety checks
- Multiple matches handling (returns first or limited)

#### Persistence (`persistence.go`)

**Purpose:** Session persistence across operations

**Key Functions:**
```go
func (b *Browser) SaveSession(profilePath string) error
func (b *Browser) LoadSession(profilePath string) error
func (b *Browser) ExportCookies(profilePath string) error
func (b *Browser) ImportCookies(profilePath string) error
func (p *Page) ExecuteScript(script string) (interface{}, error)
```

### 2. Agent Package (`internal/agent/`)

#### Main Agent Logic (`agent.go`)

**Purpose:** Orchestrates the AI-driven automation workflow

**Key Functions:**
```go
type Agent struct {
    browser *Browser
    llm     *LLMClient
    context *ConversationContext
    tools   []Tool
    logger  *zap.Logger
    // ... other fields
}

func NewAgent(browser *Browser, llm *LLMClient, logger *zap.Logger) *Agent
func (a *Agent) Execute(ctx context.Context, task string) (string, error)
func (a *Agent) runLoop(ctx context.Context, task string) error
func (a *Agent) handleToolCall(toolCall ToolCall) (*ToolResult, error)
```

**Workflow:**
1. Initialize conversation context with task
2. Extract initial page context
3. Loop:
   - Send conversation + context to LLM
   - Get AI response with tool calls or action
   - Execute tool calls if present
   - Update context with results
   - Repeat until task complete

#### Conversation Context (`context.go`)

**Purpose:** Manages conversation history and page context

**Key Types:**
```go
type ConversationContext struct {
    messages       []MessageParam
    pageContext    PageContext
    currentToolSet ToolSet
    sessionHistory []SessionHistory
    // ... other fields
}

type PageContext struct {
    URL        string
    Title      string
    Text       string
    Links      []Link
    Elements   []ElementInfo
    Summary    string
    TokenCount int
    // ... other fields
}

type SessionHistory struct {
    Timestamp time.Time
    Action    string
    Result    string
}
```

**Key Functions:**
```go
func (c *ConversationContext) AddMessage(msg MessageParam)
func (c *ConversationContext) GetCurrentPageContext() PageContext
func (c *ConversationContext) UpdatePageContext(page *Page, extractor Extractor) error
func (c *ConversationContext) SummarizeIfNeeded() error
func (c *ConversationContext) GetHistorySummary(maxMessages int) string
func (c *ConversationContext) Reset()
```

#### Tool Definitions (`tools.go`)

**Purpose:** Defines all available browser tools and their schemas

**Key Functions:**
```go
type Tool struct {
    Name        string
    Description string
    Schema      map[string]interface{}
    Execute     func(ctx context.Context, args map[string]interface{}) (string, error)
}

func (a *Agent) GetTools() []Tool
func (a *Agent) ExecuteTool(toolName string, args map[string]interface{}) (string, error)
```

**Available Tools:**
- `navigate`: Navigate to URL
- `click`: Click element by selector or text
- `input`: Input text into element
- `extract_text`: Extract text from element
- `extract_links`: Extract all links
- `go_back`: Go back in history
- `go_forward`: Go forward in history
- `refresh`: Refresh page
- `take_screenshot`: Take screenshot
- `execute_script`: Execute JS script
- `get_url`: Get current URL
- `search`: Search page content
- `get_text`: Get page text

#### Security Layer (`security.go`)

**Purpose:** Requires user confirmation for destructive actions

**Key Functions:**
```go
type Security struct {
    agent    *Agent
    logger   *zap.Logger
    prompts  []ConfirmationPrompt
    // ... other fields
}

func NewSecurity(agent *Agent, logger *zap.Logger) *Security
func (s *Security) ShouldRequireConfirmation(action string) bool
func (s *Security) GetConfirmation(action, details string) (bool, error)
func (s *Security) RequireConfirmation(action string) error
```

**Confirmation Prompts:**
- Navigation to new domains
- Form submissions
- Order placements
- Checkout actions
- Multiple selections

#### Tool Executor (`executor.go`)

**Purpose:** Dispatches tool execution with error handling

**Key Functions:**
```go
func (a *Agent) executeTool(ctx context.Context, toolCall ToolCall) (*ToolResult, error)
func (a *Agent) executeNavigateTool(ctx context.Context, args map[string]interface{}) (string, error)
func (a *Agent) executeClickTool(ctx context.Context, args map[string]interface{}) (string, error)
func (a *Agent) executeInputTool(ctx context.Context, args map[string]interface{}) (string, error)
func (a *Agent) executeExtractTextTool(ctx context.Context, args map[string]interface{}) (string, error)
func (a *Agent) executeGoBackTool(ctx context.Context, args map[string]interface{}) (string, error)
// ... other tool executers
```

### 3. Extractor Package (`internal/extractor/`)

#### Main Extractor (`extractor.go`)

**Purpose:** Defines extraction strategies and algorithms

**Key Types:**
```go
type Extractor interface {
    Extract(context context.Context, page *Page) (PageContext, error)
}

type PageExtractor struct {
    strategy ContextStrategy
    calculator TokenCalculator
    logger   *zap.Logger
}

type PageContext struct {
    URL        string
    Title      string
    Text       string
    Links      []Link
    Elements   []ElementInfo
    Summary    string
    TokenCount int
}
```

**Key Functions:**
```go
func NewPageExtractor(strategy ContextStrategy, calculator TokenCalculator, logger *zap.Logger) *PageExtractor
func (e *PageExtractor) Extract(context context.Context, page *Page) (PageContext, error)
func (e *PageExtractor) ExtractText(page *Page) (string, error)
func (e *PageExtractor) ExtractLinks(page *Page) ([]Link, error)
func (e *PageExtractor) ExtractElements(page *Page) ([]ElementInfo, error)
func (e *PageExtractor) SummarizePage(page *Page, maxTokens int) (string, error)
```

#### Extraction Strategies (`strategies.go`)

**Purpose:** Different content extraction approaches

**Key Types:**
```go
type ExtractionStrategy interface {
    Extract(context context.Context, page *Page) (PageContext, error)
}

type DefaultExtractor struct {
    pageExtractor *PageExtractor
    calculator    *TokenCalculator
    logger        *zap.Logger
}

type LinkExtractor struct {
    pageExtractor *PageExtractor
    logger        *zap.Logger
}
```

**Extraction Strategies:**
- `DefaultExtractor`: Full page extraction with text, links, elements
- `LinkExtractor`: Link-focused extraction
- `MinimalExtractor`: Text-only extraction for token efficiency

#### Context Strategy (`context_strategy.go`)

**Purpose:** Manages context for LLM to avoid token limits

**Key Types:**
```go
type ContextStrategy interface {
    ShouldSummarize(ctx PageContext, budget int) bool
    BuildContext(ctx PageContext) string
    UpdateContext(ctx PageContext, previousContext string) (string, int)
}

type AdaptiveContextStrategy struct {
    calculator *TokenCalculator
    logger     *zap.Logger
}

type SummarizationContextStrategy struct {
    calculator *TokenCalculator
    logger     *zap.Logger
}
```

**Context Strategies:**
- `AdaptiveContextStrategy`: Dynamically switches between full, partial, and summarized context based on token budget
- `SummarizationContextStrategy`: Always uses summarization for large pages
- `MinimalContextStrategy`: Only extracts essential context (text + links)

#### Token Calculator (`token_calculator.go`)

**Purpose:** Estimates token usage and manages budgets

**Key Functions:**
```go
type TokenCalculator struct {
    // ... fields
}

func NewTokenCalculator() *TokenCalculator
func (tc *TokenCalculator) CountTokens(text string) int
func (tc *TokenCalculator) CountContextTokens(ctx PageContext) int
func (tc *TokenCalculator) GetAvailableBudget(budget int, contextSize int) int
func (tc *TokenCalculator) ShouldReplaceContext(currentTokens int, newTokens int) bool
```

#### Page Summarizer (`page_summarizer.go`)

**Purpose:** Summarizes page content for context management

**Key Functions:**
```go
type PageSummarizer struct {
    splitter *textsplitter.RecursiveCharacterTextSplitter
    calculator *TokenCalculator
    logger    *zap.Logger
}

func NewPageSummarizer(calculator *TokenCalculator, logger *zap.Logger) *PageSummarizer
func (ps *PageSummarizer) SummarizePage(page *Page, maxTokens int) (string, error)
func (ps *PageSummarizer) SummarizeText(text string, maxTokens int) (string, error)
```

### 4. LLM Package (`internal/llm/`)

#### Client (`client.go`)

**Purpose:** OpenAI-compatible API client wrapper

**Key Types:**
```go
type LLMClient struct {
    client    *openai.Client
    baseURL   string
    model     string
    apiKey    string
    maxTokens int
    logger    *zap.Logger
}

type MessageParam struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type ToolCall struct {
    Id      string                 `json:"id"`
    Name    string                 `json:"name"`
    Arguments map[string]interface{} `json:"arguments"`
}

type LLMResponse struct {
    Content       string        `json:"content"`
    ToolCalls     []ToolCall    `json:"tool_calls"`
    FinishReason  string        `json:"finish_reason"`
}
```

**Key Functions:**
```go
func NewLLMClient(apiKey, baseURL, model string, maxTokens int, logger *zap.Logger) (*LLMClient, error)
func (lc *LLMClient) Chat(ctx context.Context, messages []MessageParam, tools []ToolParam) (*LLMResponse, error)
func (lc *LLMClient) ToolFunctionToParam(tool Tool) ToolParam
func (lc *LLMClient) ToolResultToContent(result ToolResult) string
func (lc *LLMClient) ParseToolCallResponse(content string) (string, []ToolCall)
```

#### Message Formatting (`messages.go`)

**Purpose:** Helper functions for message construction

**Key Functions:**
```go
func (lc *LLMClient) BuildSystemMessage(task string, tools []ToolParam) MessageParam
func (lc *LLMClient) BuildUserMessage(content string) MessageParam
func (lc *LLMClient) BuildToolResultMessage(toolCall ToolCall, result string) MessageParam
func (lc *LLMClient) BuildAssistantMessage(content string, toolCalls []ToolCall) MessageParam
```

#### System Prompts (`prompts.go`)

**Purpose:** LLM system prompts and instructions

**Key Functions:**
```go
func GetAgentSystemPrompt(tools []Tool) string
func GetContextExtractionPrompt() string
func GetConfirmationPrompt(action, details string) string
```

**System Prompt Sections:**
- Role definition
- Task description
- Tool descriptions and usage
- Decision-making guidelines
- Error handling requirements
- Context management instructions

### 5. Logger Package (`internal/logger/`)

**Purpose:** Structured logging with context

**Key Functions:**
```go
type Logger struct {
    *zap.Logger
    *zap.SugaredLogger
}

func NewLogger(debug bool) (*Logger, error)
func (l *Logger) DebugContext(ctx context.Context, msg string, fields ...zap.Field)
func (l *Logger) InfoContext(ctx context.Context, msg string, fields ...zap.Field)
func (l *Logger) ErrorContext(ctx context.Context, msg string, fields ...zap.Field)
func (l *Logger) FatalContext(ctx context.Context, msg string, fields ...zap.Field)
```

## Configuration

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ZAI_API_KEY` | Yes | - | API key for z.ai or OpenAI |
| `ZAI_BASE_URL` | No | https://api.z.ai/v1 | Base URL for API endpoint |
| `ZAI_MODEL` | No | zlm-4.5-flash | Model name to use |
| `USER_DATA_DIR` | No | ./user-data | Browser user data directory |
| `DEBUG` | No | false | Enable debug logging |
| `TASK` | No | - | Initial task for the agent |
| `MAX_PAGES` | No | 5 | Maximum concurrent browser pages |
| `MAX_TOKENS` | No | 4000 | Maximum context tokens per page |
| `HEADLESS` | No | true | Run browser in headless mode |
| `VISIBLE` | No | false | Show browser window for debugging |

### Config File (`configs/config.yaml`)

```yaml
api:
  key: "your-api-key"
  base_url: "https://api.z.ai/v1"
  model: "zlm-4.5-flash"
  max_tokens: 4000

browser:
  user_data_dir: "./user-data"
  headless: true
  visible: false
  max_pages: 5

logging:
  level: "info"
  format: "json"
  output: "stdout"

context:
  strategy: "adaptive"
  max_tokens: 4000
  summary_threshold: 2000

security:
  confirm_navigate: true
  confirm_submit: true
  confirm_order: true
```

## Workflow

### Agent Execution Flow

```
1. Initialize
   ├─ Load configuration
   ├─ Create logger
   ├─ Initialize LLM client
   ├─ Create browser instance
   └─ Create agent

2. Start Loop
   ├─ Extract initial page context
   ├─ Send to LLM
   ├─ Get response with tool calls
   ├─ Execute tool calls
   │   ├─ Navigate
   │   ├─ Click
   │   ├─ Input
   │   ├─ Extract
   │   └─ etc.
   ├─ Update context with results
   ├─ Check if task complete
   └─ Repeat if not complete

3. Complete
   ├─ Return final result
   ├─ Save session
   ├─ Close browser
   └─ Clean up
```

### Tool Execution Flow

```
1. Tool Call Received
   ├─ Check security requirements
   ├─ Get user confirmation if needed
   ├─ Validate arguments
   ├─ Execute operation
   │   ├─ Find element (if applicable)
   │   ├─ Perform action
   │   └─ Handle errors
   └─ Return result string
```

### Context Management Flow

```
1. Extract Context
   ├─ Get current page
   ├─ Extract text
   ├─ Extract links
   ├─ Extract elements
   ├─ Count tokens
   └─ Build context string

2. Check Token Budget
   ├─ Calculate current tokens
   ├─ Compare with budget
   ├─ If within budget:
   │   └─ Use current context
   └─ If over budget:
       ├─ Decide strategy (summarize, truncate, switch context)
       ├─ Apply strategy
       └─ Return new context

3. Update Conversation
   ├─ Add context to messages
   ├─ Keep conversation history
   └─ Remove old messages if needed
```

## Error Handling

### Error Types (`internal/types/errors.go`)

```go
type BrowserError struct {
    Operation string
    Err       error
}

type AgentError struct {
    Operation string
    Err       error
}

type ToolExecutionError struct {
    ToolName string
    Args     map[string]interface{}
    Err      error
}

type ExtractionError struct {
    Strategy string
    Err      error
}
```

### Error Wrapping Pattern

All errors are wrapped with context:

```go
if err := browser.Open(url); err != nil {
    return fmt.Errorf("failed to navigate to %s: %w", url, err)
}

if element == nil {
    return fmt.Errorf("element not found using selector %s", selector)
}
```

### Retry Logic

Browser operations implement retry logic for transient failures:

```go
func (p *Page) NavigateWithRetry(url string, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        if err := p.Navigate(url); err == nil {
            return nil
        }
        time.Sleep(2 * time.Second)
    }
    return fmt.Errorf("navigation failed after %d retries", maxRetries)
}
```

## Testing Strategy

### Interface-Based Testing

```go
// Browser interface for mocking
type Browser interface {
    Open(url string) (*Page, error)
    Close() error
    GetCurrentPage() *Page
}

// Mock implementation
type MockBrowser struct {
    ShouldFail bool
    LastURL    string
}

func (m *MockBrowser) Open(url string) (*Page, error) {
    m.LastURL = url
    if m.ShouldFail {
        return nil, errors.New("browser open failed")
    }
    return &MockPage{}, nil
}

// Test with mock
func TestAgentExecute(t *testing.T) {
    mockBrowser := &MockBrowser{}
    mockLLM := &MockLLM{}
    agent := NewAgent(mockBrowser, mockLLM)
    // Test...
}
```

### Test Coverage Areas

1. **Browser Package:**
   - Navigation operations
   - Element search strategies
   - Session persistence
   - Error handling and retries

2. **Agent Package:**
   - Tool calling workflow
   - Context management
   - Conversation flow
   - Security layer

3. **Extractor Package:**
   - Context extraction
   - Token counting
   - Summarization strategies
   - Budget management

4. **LLM Package:**
   - API communication
   - Message formatting
   - Tool call parsing
   - Response handling

## Best Practices

### Never Hardcode Selectors

```go
// ❌ BAD - Hardcoded selector
page.Click("#submit-button")

// ✅ GOOD - AI-driven selection
action := llm.DecideNextAction(context)
page.Click(action.ElementSelector)
```

### Never Hardcode URL Paths

```go
// ❌ BAD - Hardcoded paths
navigateTo("https://example.com/login")
navigateTo("https://example.com/cart")

// ✅ GOOD - Dynamic URLs
navigateTo(page.FindElementByText("login").Attribute("href"))
```

### Never Create Predefined Workflows

```go
// ❌ BAD - Predefined steps
func OrderFood(task string) {
    navigateTo("https://eda.yandex.ru")
    clickElement(".search-button")
    inputText("#search", task)
    clickElement(".add-to-cart")
}

// ✅ GOOD - AI-driven execution
context := extractor.ExtractContext(page)
action := llm.DecideNextAction(context, task)
for !action.IsComplete() {
    action.Execute()
    context = extractor.ExtractContext(page)
    action = llm.DecideNextAction(context, task)
}
```

### Always Propagate Context

```go
func (a *Agent) Execute(ctx context.Context, task string) (string, error) {
    select {
    case <-ctx.Done():
        return "", fmt.Errorf("execution canceled: %w", ctx.Err())
    default:
    }
    
    result, err := a.llm.Chat(ctx, messages, tools)
    if err != nil {
        return "", fmt.Errorf("LLM call failed: %w", err)
    }
    
    err = a.executeTool(ctx, toolCall)
    if err != nil {
        return "", fmt.Errorf("tool execution failed: %w", err)
    }
    
    return result, nil
}
```

## API Integration

### OpenAI-Compatible API

The agent uses OpenAI-compatible API format:

```go
type LLMRequest struct {
    Model    string          `json:"model"`
    Messages []MessageParam  `json:"messages"`
    Tools    []ToolParam     `json:"tools"`
    // ... other fields
}

type ToolParam struct {
    Type       string                 `json:"type"`
    Function   ToolFunctionParam      `json:"function"`
}

type ToolFunctionParam struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Parameters  map[string]interface{} `json:"parameters"`
}
```

### Supported Endpoints

- z.ai API: https://api.z.ai/v1
- OpenAI API: https://api.openai.com/v1
- Any OpenAI-compatible endpoint

## Examples

### Basic Usage

```bash
# Set API key
export ZAI_API_KEY=your-api-key

# Run with task
go run ./cmd/agent TASK="Navigate to example.com and extract its title"

# Run with debug
DEBUG=true go run ./cmd/agent TASK="Order food from Yandex.Eda"
```

### Custom Configuration

```bash
# Use custom config file
go run ./cmd/agent --config configs/config.yaml TASK="Find latest news"
```

### Multiple Tasks

```bash
# Task with complex operations
go run ./cmd/agent TASK="Apply for 3 job positions on hh.ru: Senior Go Developer, 5 years experience"
```

## Performance Considerations

### Token Management

- Adaptive context strategy balances between full context and summarization
- Token calculator estimates usage before each extraction
- Context pruning prevents token overflow

### Browser Performance

- Page pooling limits concurrent pages
- Element caching avoids repeated searches
- Navigation retry handles transient failures

### LLM Efficiency

- Tool calling reduces unnecessary API calls
- Context summarization optimizes token usage
- Message history pruning maintains relevant information

## Future Enhancements

1. **Advanced Element Selection:**
   - AI-powered element identification
   - Visual similarity matching
   - Behavior-based element detection

2. **Improved Context Management:**
   - Semantic context understanding
   - Dynamic context importance weighting
   - Multi-step context tracking

3. **Enhanced Security:**
   - Role-based permissions
   - Custom confirmation rules
   - Audit logging for all actions

4. **Multi-Agent Support:**
   - Parallel task execution
   - Agent collaboration
   - Distributed workflows

5. **Browser Intelligence:**
   - Machine learning for element detection
   - Smart navigation prediction
   - User behavior learning

## Dependencies

### Core Dependencies

```
github.com/go-rod/rod v0.116.12
github.com/sashabaranov/go-openai v1.33.2
github.com/sirupsen/zap v1.27.0
github.com/tmc/langchaingo/textsplitter
github.com/joho/godotenv
github.com/spf13/viper
```

### Version Requirements

- Go: 1.24+
- Rod: 0.116.12+
- OpenAI Client: 1.33.2+
- Zap Logger: 1.27.0+

## Security Considerations

1. **API Key Management:**
   - Never commit API keys to version control
   - Use environment variables for sensitive data
   - Implement proper key rotation

2. **Browser Security:**
   - Isolate browser sessions per user
   - Clean up cookies and session data after use
   - Use secure user data directories

3. **User Protection:**
   - Confirm destructive actions
   - Validate user input
   - Implement rate limiting

## License

[Specify license here]

## Contributing

[Contribution guidelines]
