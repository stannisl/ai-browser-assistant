# Comprehensive Testing Plan for AI Browser Assistant

## Overview
This document outlines the complete testing strategy for the ai-browser-assistant project, including unit tests, mocks, integration tests, coverage matrix, and testing priorities.

## Current Test Coverage Status
- **internal/llm/tools_test.go**: 6 test functions, ~476 lines (pre-existing)
- All other packages: NO tests written yet

---

## Package-by-Package Test Analysis

### 1. internal/llm/ Package

**Dependencies:**
- `github.com/sashabaranov/go-openai` (external API - needs mocking)
- `github.com/stannisl/ai-browser-assistant/internal/logger` (local, no mocking needed)

**Files to Test:**
- `client.go` (GOOGLE DEPENDENCY)
- `prompts.go` (no dependencies)
- `tools.go` (no dependencies)
- `tools_test.go` (exists, already has input parsing tests)

**Mock Requirements:**
- Mock `openai.Client` from `github.com/sashabaranov/go-openai`
- Mock `*openai.ChatCompletionResponse`
- Mock `*openai.ChatCompletionRequest`

**Test Structure:**
```
internal/llm/
├── client_test.go          # NEW - Mock tests for Client
├── prompts_test.go          # NEW - System prompt tests
├── tools.go
└── tools_test.go (exists)
```

**Test Coverage:**

#### client_test.go (NEW)
- `TestNewClient`: 
  - Happy path
  - BaseURL configuration
  - Debug mode
- `TestChat`:
  - Happy path with tool calls
  - Happy path without tool calls
  - Context cancellation
  - API error handling
  - Response parsing
- `TestExtractToolCall`:
  - Tool call present
  - Tool call absent
  - Multiple tool calls
  - Invalid JSON in arguments
  - Empty response
- `TestGetModel`: Simple getter test

#### prompts_test.go (NEW)
- `TestSystemPromptContent`: Verify prompt structure and constraints
- `TestGetTools`: Verify all 9 tools are defined correctly
- Tool-specific parameter validation tests

---

### 2. internal/logger/ Package

**Dependencies:**
- `github.com/uber.org/zap` (external library, no mocking needed)
- `github.com/fatih/color` (external library, no mocking needed)

**Files to Test:**
- `logger.go`

**Mock Requirements:**
- None (pure Go library with no external dependencies that need mocking)

**Test Structure:**
```
internal/logger/
└── logger_test.go          # NEW
```

**Test Coverage:**

#### logger_test.go (NEW)
- `TestNewLogger`:
  - Production mode
  - Debug mode
  - Invalid configuration
- `TestLoggerMethods`:
  - Info, Debug, Error, Warn methods
  - All logging methods return successfully
- `TestLoggerMethodsWithValues`:
  - Key-value logging
  - Edge cases with nil values
- `TestLoggerClosing`:
  - Close method succeeds
  - Multiple closes don't panic
- `TestTruncateFunction`:
  - Short strings
  - Exact length strings
  - Long strings with truncation
- `TestColorHandling`:
  - Terminal detection
  - Non-terminal output

---

### 3. internal/types/ Package

**Dependencies:**
- None (pure data structures)

**Files to Test:**
- `agent.go` (types only)
- `browser.go` (types only)
- `errors.go` (error types)

**Test Structure:**
```
internal/types/
├── agent_test.go           # NEW
├── browser_test.go         # NEW
├── errors_test.go          # NEW
└── (existing files)
```

**Test Coverage:**

#### agent_test.go (NEW)
- `TestToolCall`: Field initialization
- `TestAgentConfig`: All fields validation
- `TestLLMConfig`: All fields validation
- `TestToolDefinition`: Structure
- `TestToolParam`: Structure
- `TestLLMResponse`: Structure
- `TestMessageParam`: Structure

#### browser_test.go (NEW)
- `TestPageElement`: Structure and methods
- `TestPageState`: Structure and methods
- `TestFormElement`: Structure and methods
- `TestInputField`: Structure and methods
- `TestSubmitButton`: Structure and methods
- `TestLinkElement`: Structure and methods
- `TestBrowserConfig`: Structure

#### errors_test.go (NEW)
- `TestErrorVariables`: All predefined errors
- `TestToolExecutionError`: Error wrapping
- `TestContextError`: Error wrapping
- `TestSecurityError`: Error wrapping and unwrapping

---

### 4. internal/browser/ Package

**CRITICAL DEPENDENCY**: `github.com/go-rod/rod` (real browser - MUST BE MOCKED)

**Dependencies:**
- `github.com/go-rod/rod` (external browser automation - needs mocking)
- `github.com/go-rod/rod/lib/launcher` (external - needs mocking)
- `github.com/go-rod/rod/lib/proto` (external - needs mocking)
- `github.com/stannisl/ai-browser-assistant/internal/logger` (local)
- `github.com/stannisl/ai-browser-assistant/internal/types` (local)

**Files to Test:**
- `browser.go`

**Mock Requirements:**
- Mock `*rod.Browser`
- Mock `*rod.Page`
- Mock `*rod.Page.Info()` method
- Mock `rod.Launcher` methods
- Mock `*proto.InputMouseButtonLeft`

**Test Structure:**
```
internal/browser/
└── browser_test.go          # NEW
```

**Test Coverage:**

#### browser_test.go (NEW)
- `TestNewManager`: Constructor
- `TestManager_Launch`:
  - Happy path
  - Debug mode
  - UserDataDir configuration
- `TestManager_Navigate`:
  - Happy path with valid URL
  - Invalid URL
  - Context cancellation
  - Page load timeout
  - Navigation error
- `TestManager_Click`:
  - Happy path
  - Element not found
  - Click error
  - Context cancellation
- `TestManager_Type`:
  - Happy path
  - Element not found
  - Type error
  - Empty text
- `TestManager_Scroll`:
  - Up direction
  - Down direction
  - Invalid direction
  - Context cancellation
- `TestManager_GetPage`: Getter
- `TestManager_GetURL`: Getter
- `TestManager_GetTitle`: Getter
- `TestManager_Close`: Cleanup
- `TestIsValidURL`: Helper function

**Mock Strategy:**
- Use `rod`'s internal testing or create interface-based mocks
- Since rod is Cgo-bound, we may need to:
  - Create interfaces for `Browser`, `Page`, `Element`
  - Use `rod/lib/launcher/mock` if available
  - OR use integration tests for browser operations (lower priority)

---

### 5. internal/extractor/ Package

**CRITICAL DEPENDENCY**: `github.com/go-rod/rod.Page` (real browser - MUST BE MOCKED)

**Dependencies:**
- `github.com/go-rod/rod` (external browser - needs mocking)
- `github.com/stannisl/ai-browser-assistant/internal/logger` (local)
- `github.com/stannisl/ai-browser-assistant/internal/types` (local)

**Files to Test:**
- `extractor.go`

**Mock Requirements:**
- Mock `*rod.Page`
- Mock `page.MustEval()` method
- Mock `page.MustElement()` method
- Mock `page.MustProperty()` method
- Mock `page.Info()` method

**Test Structure:**
```
internal/extractor/
└── extractor_test.go        # NEW
```

**Test Coverage:**

#### extractor_test.go (NEW)
- `TestExtractor_New`: Constructor
- `TestExtractor_Extract`:
  - Happy path with multiple elements
  - Happy path with no elements
  - Empty page
  - Has modal detection
  - Element truncation (max 50)
  - Context cancellation
  - JSON parsing error
  - Page error
- `TestExtractor_GetSelector`:
  - Existing ID
  - Non-existent ID
  - Concurrent access
- `TestExtractor_FormatForLLM`:
  - Multiple page elements
  - Single element
  - Has modal
  - Input fields only
  - Buttons only
  - Links only
  - Complex page structure
- `TestFormatElement`: Helper function
- `TestTruncateHelper`: Helper function
- `TestMinHelper`: Helper function

**Mock Strategy:**
- Create interface for `Page`
- Mock `page.Eval()` for JS execution
- Mock page property getters
- Test JS-generated JSON structures

---

### 6. internal/agent/ Package

**CRITICAL DEPENDENCIES**: 
- `internal/browser` (must be mocked)
- `internal/extractor` (must be mocked)
- `internal/llm` (must be mocked)
- `internal/logger` (local)
- `internal/types` (local)

**Files to Test:**
- `agent.go` (main logic, NO EXISTING TESTS)
- `executor.go` (all Execute* methods, NO EXISTING TESTS)

**Mock Requirements:**
- Mock `browser.Manager` interface
- Mock `extractor.Extractor` interface
- Mock `llm.Client` interface
- Mock `logger.Logger` interface

**Test Structure:**
```
internal/agent/
├── agent_test.go            # NEW
└── executor_test.go          # NEW
```

**Test Coverage:**

#### agent_test.go (NEW)
- `TestNewAgent`: Constructor
- `TestRun`:
  - Happy path: Tool calls only
  - Happy path: Single step completion with report
  - Context cancellation (SIGINT/SIGTERM)
  - Max steps exceeded
  - LLM error handling
  - Tool execution error (stop at first error)
  - Multiple tool calls in one response
  - History trimming (20 messages)
  - Infinite loop detection (>10 calls)
- `TestTrimHistory`:
  - Under limit
  - Exactly at limit
  - Over limit
  - Preserve first 2 messages
- `TestExecuteTool`:
  - Single tool call
  - Multiple tool calls
  - Tool execution error propagation
- `TestExecuteToolInternal`:
  - All 9 tools (extract_page, navigate, click, type_text, scroll, wait, ask_user, confirm_action, report)
  - Unknown tool error
  - Tool-specific argument validation
- `TestToolArgumentValidation`:
  - Missing required arguments
  - Invalid argument types
  - Invalid values (negative ID, etc.)
  - Empty strings
  - Out of range values

#### executor_test.go (NEW)
- `TestExecuteExtractPage`:
  - Happy path
  - Extractor error
- `TestExecuteNavigate`:
  - Happy path
  - Missing URL
  - Invalid URL
  - Browser navigation error
  - Context cancellation
- `TestExecuteClick`:
  - Happy path
  - Missing element_id
  - Non-numeric element_id
  - Negative element_id
  - Element not found
  - Browser click error
  - Context cancellation
- `TestExecuteTypeText`:
  - Happy path
  - Missing element_id
  - Missing text
  - Non-numeric element_id
  - Non-string text
  - Empty text
  - Element not found
  - Browser type error
  - Context cancellation
- `TestExecuteScroll`:
  - Happy path (up)
  - Happy path (down)
  - Missing direction
  - Invalid direction (left, right, invalid string)
  - Browser scroll error
  - Context cancellation
- `TestExecuteWait`:
  - Happy path (1s)
  - Happy path (10s)
  - Missing seconds
  - Non-numeric seconds
  - Zero seconds
  - Negative seconds
  - >10 seconds
  - Context cancellation
- `TestExecuteAskUser`:
  - Happy path
  - Missing question
  - Non-string question
  - Empty question
  - Context cancellation
- `TestExecuteConfirmAction`:
  - Happy path (yes)
  - Happy path (no)
  - Missing description
  - Non-string description
  - Empty description
  - Context cancellation
  - User denies (returns ErrConfirmationDenied)
- `TestExecuteReport`:
  - Happy path with message
  - Missing message (default "Task completed")
  - Non-string message (default)
  - Boolean success field
  - Context cancellation

**Mock Strategy:**
- Define interfaces for all external dependencies
- Create test mocks for each interface
- Test error paths and edge cases extensively
- Verify context propagation

---

### 7. cmd/agent/ Package

**Dependencies:**
- All internal packages (no mocking needed - integration test)

**Files to Test:**
- `main.go`

**Test Structure:**
```
cmd/agent/
└── main_test.go             # INTEGRATION TEST (optional)
```

**Test Coverage:**

#### main_test.go (NEW - Integration Test)
- `TestMain`: 
  - Application startup (with mocks)
  - Environment variable handling
  - Flag parsing
  - Signal handling (SIGINT, SIGTERM)
  - Graceful shutdown
- `TestGetEnvOrDefault`:
  - Environment variable set
  - Environment variable not set
  - Empty environment variable
  - Default value

**Note**: This is an integration test that should use mocks for all internal dependencies to avoid requiring a real browser and API key.

---

## File Structure Summary

```
ai-browser-assistant/
├── internal/
│   ├── agent/
│   │   ├── agent.go
│   │   ├── executor.go
│   │   ├── agent_test.go           [NEW - comprehensive tests]
│   │   └── executor_test.go         [NEW - comprehensive tests]
│   │
│   ├── browser/
│   │   ├── browser.go
│   │   └── browser_test.go          [NEW - mock tests]
│   │
│   ├── extractor/
│   │   ├── extractor.go
│   │   └── extractor_test.go        [NEW - mock tests]
│   │
│   ├── llm/
│   │   ├── client.go
│   │   ├── prompts.go
│   │   ├── tools.go
│   │   ├── tools_test.go (exists)
│   │   ├── client_test.go           [NEW - mock tests]
│   │   └── prompts_test.go          [NEW]
│   │
│   ├── logger/
│   │   ├── logger.go
│   │   └── logger_test.go           [NEW]
│   │
│   └── types/
│       ├── agent.go
│       ├── browser.go
│       ├── errors.go
│       ├── agent_test.go            [NEW]
│       ├── browser_test.go          [NEW]
│       └── errors_test.go           [NEW]
│
└── cmd/
    └── agent/
        ├── main.go
        └── main_test.go             [NEW - integration test]
```

---

## Mock Interface Definitions

### browser.MockBrowser (for internal/browser)
```go
type Browser interface {
    Navigate(ctx context.Context, url string) error
    Click(ctx context.Context, selector string) error
    Type(ctx context.Context, selector, text string) error
    Scroll(ctx context.Context, direction string) error
    GetPage() interface{}
    GetURL() string
    GetTitle() string
}

type Page interface {
    Navigate(url string) error
    Element(selector string) (interface{}, error)
    MustElement(selector string) interface{}
    MustInput(text string)
    MustEval(jsCode string) interface{}
    MustProperty(name string) string
    Page() interface{}
    Mouse interface{}
}
```

### extractor.MockExtractor (for internal/extractor)
```go
type Extractor interface {
    Extract(ctx context.Context) (*types.PageState, error)
    GetSelector(id int) (string, error)
    FormatForLLM(state *types.PageState) string
}

type PageState struct {
    Title, URL string
    Elements   []PageElement
    HasModal   bool
    // ... other fields
}
```

### llm.MockClient (for internal/llm)
```go
type Client interface {
    Chat(ctx context.Context, messages []openai.ChatCompletionMessage) (*openai.ChatCompletionResponse, error)
    ExtractToolCall(response *openai.ChatCompletionResponse) (*types.ToolCall, bool)
    GetModel() string
}
```

### agent.MockBrowser (for internal/agent)
- Use same interface as browser.MockBrowser
- Add Execute methods if needed

### agent.MockExtractor (for internal/agent)
- Use same interface as extractor.MockExtractor

### agent.MockLLM (for internal/agent)
- Use same interface as llm.MockClient
- Add ToolCall mock response methods

### logger.MockLogger (for internal/agent, extractor, browser)
```go
type Logger interface {
    Info(msg string, keysAndValues ...interface{})
    Debug(msg string, keysAndValues ...interface{})
    Error(msg string, keysAndValues ...interface{})
    Warn(msg string, keysAndValues ...interface{})
    Step(current, max int)
    Tool(name string)
    Thinking()
    Extract(url string, count int)
    Navigate(url string)
    Click(id int, text string)
    Type(id int, text string)
    Scroll(direction string)
    Ask(question string)
    Confirm(description string)
    Done(message string, success bool)
    Close()
}
```

---

## Coverage Matrix

### Priority 1: Critical Path Testing (Must Test First)

| Package | File | Function | Test Cases | Priority |
|---------|------|----------|------------|----------|
| llm | client.go | NewClient | 3 | High |
| llm | client.go | Chat | 5 | High |
| llm | client.go | ExtractToolCall | 5 | High |
| agent | agent.go | Run (happy path) | 5 | High |
| agent | agent.go | Run (error paths) | 4 | High |
| agent | agent.go | TrimHistory | 4 | High |
| agent | executor.go | ExecuteNavigate | 6 | High |
| agent | executor.go | ExecuteClick | 7 | High |
| agent | executor.go | ExecuteTypeText | 7 | High |
| agent | executor.go | ExecuteScroll | 5 | High |
| agent | executor.go | ExecuteWait | 6 | High |
| agent | executor.go | ExecuteAskUser | 5 | High |
| agent | executor.go | ExecuteConfirmAction | 7 | High |
| agent | executor.go | ExecuteReport | 5 | High |
| browser | browser.go | Navigate | 5 | High |
| browser | browser.go | Click | 5 | High |
| extractor | extractor.go | Extract | 8 | High |
| extractor | extractor.go | GetSelector | 3 | High |

### Priority 2: Essential Testing (Test Second)

| Package | File | Function | Test Cases | Priority |
|---------|------|----------|------------|----------|
| llm | tools.go | GetTools | 1 | Medium |
| logger | logger.go | NewLogger | 3 | Medium |
| types | agent.go | All types | 7 | Medium |
| types | browser.go | All types | 7 | Medium |
| types | errors.go | Error types | 4 | Medium |
| browser | browser.go | Type, Scroll, Close | 4 | Medium |
| extractor | extractor.go | FormatForLLM | 7 | Medium |

### Priority 3: Nice-to-Have (Test Later)

| Package | File | Function | Test Cases | Priority |
|---------|------|----------|------------|----------|
| cmd | main.go | Signal handling | 3 | Low |
| logger | logger.go | Color handling | 2 | Low |

---

## Integration Testing Strategy

### High-Level Integration Tests (Manual Testing)

1. **End-to-End Task Execution**
   - Navigate to a real website
   - Perform a simple task (search, click, submit)
   - Verify final state
   - Test with real browser and LLM API

2. **Security Confirmations**
   - Test confirm_action with dangerous operations
   - Verify user can deny actions

3. **Context Propagation**
   - Multi-step tasks
   - Verify conversation history is maintained

4. **Error Recovery**
   - Network failures
   - Element not found
   - Page load timeouts

### Note
- Full integration tests require real browser and API key
- Should be separate from unit tests (e.g., in `tests/` directory)
- Not covered in this plan (out of scope)

---

## Testing Commands

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run with verbose output
go test ./... -v

# Run specific package
go test ./internal/agent -v

# Run specific test
go test ./internal/agent -run TestRun -v

# Run with race detector
go test ./... -race

# Generate coverage report
go test ./... -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out

# Run benchmarks
go test ./... -bench=. -benchmem

# List all tests
go test ./... -list=.
```

---

## Expected Test Failures (TDD Phase)

### Initial State (Before Implementation)

These tests will fail until the corresponding implementation is added:

1. **internal/browser/browser_test.go**
   - All tests expect mock Browser/Page interfaces
   - Need to implement mocks first

2. **internal/extractor/extractor_test.go**
   - All tests expect mock Page
   - Need to implement mocks first

3. **internal/agent/agent_test.go**
   - All tests expect mocked dependencies
   - Need to create mock interfaces and implementations

4. **internal/agent/executor_test.go**
   - All tests expect mocked dependencies
   - Need to create mock interfaces and implementations

5. **internal/llm/client_test.go**
   - All tests expect mock openai.Client
   - Need to create mock for openai library

### Implementation Order

1. Define interfaces for all external dependencies
2. Implement mock implementations
3. Write tests (red phase - should fail)
4. Implement functionality (green phase)
5. Write more tests
6. Refactor
7. Repeat until all tests pass

---

## Code Quality Requirements

### Effective Go Compliance

1. **Error Handling**
   - Always wrap errors with context
   - Test both success and error paths
   - Use meaningful error messages

2. **Context Propagation**
   - All methods accept context.Context
   - Test context cancellation
   - Test timeout scenarios

3. **Interface Design**
   - Small, focused interfaces
   - Test via interfaces, not implementations
   - Define interfaces in types package

4. **Testing**
   - Table-driven tests
   - Clear test names
   - Arrange-Act-Assert pattern
   - Test both happy and error paths

5. **Documentation**
   - Document all exported functions
   - Include usage examples
   - Document error conditions

---

## Success Criteria

### Minimum Coverage Requirements

- **Unit Tests**: 80% code coverage minimum
- **Critical Paths**: 100% coverage for all agent execution paths
- **Error Paths**: 100% coverage for all error conditions
- **Edge Cases**: 90% coverage for edge cases

### Quality Metrics

- Zero tests failing in CI
- Test execution time < 5 minutes
- No race conditions
- Zero memory leaks in tests

---

## Risk Assessment

### High Risk Areas

1. **Browser Mocking**
   - Risk: Rod is Cgo-bound, difficult to mock completely
   - Mitigation: Create comprehensive interfaces, test with integration tests

2. **LLM API Mocking**
   - Risk: OpenAI API responses can vary
   - Mitigation: Use deterministic test responses, separate production mocking

3. **Stateful Operations**
   - Risk: Agent state management can be complex
   - Mitigation: Isolate state changes, test with fresh instances

### Medium Risk Areas

1. **Context Propagation**
   - Risk: Context can be lost in complex flows
   - Mitigation: Explicit context checks, test cancellation propagation

2. **Error Recovery**
   - Risk: Recovery logic can be incomplete
   - Mitigation: Comprehensive error path testing

---

## Summary

This testing plan provides a comprehensive strategy for achieving full test coverage of the ai-browser-assistant project. The plan prioritizes critical paths, ensures all external dependencies are properly mocked, and follows TDD principles. The estimated test count is approximately **150+ tests** across **11 new test files**.

### Key Highlights:

- **150+ tests** to write
- **11 new test files**
- **3 priority levels** (Critical, Essential, Nice-to-Have)
- **Mock interfaces** defined for all external dependencies
- **TDD approach** with expected failures
- **80%+ coverage** target
- **Context and error handling** emphasized throughout

The plan follows Effective Go guidelines and ensures the codebase is well-tested, maintainable, and reliable.
