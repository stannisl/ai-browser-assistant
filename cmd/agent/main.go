package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/stannisl/ai-browser-assistant/internal/agent"
	"github.com/stannisl/ai-browser-assistant/internal/browser"
	"github.com/stannisl/ai-browser-assistant/internal/extractor"
	"github.com/stannisl/ai-browser-assistant/internal/llm"
	"github.com/stannisl/ai-browser-assistant/internal/logger"
	"github.com/stannisl/ai-browser-assistant/internal/types"
)

func main() {
	apiKey := flag.String("api-key", os.Getenv("ZAI_API_KEY"), "Z.AI API key")
	baseURL := flag.String("base-url", getEnvOrDefault("ZAI_BASE_URL", "https://api.z.ai/v1"), "API base URL")
	model := flag.String("model", getEnvOrDefault("ZAI_MODEL", "glm-4.5-flash"), "Model name")
	userDataDir := flag.String("user-data", getEnvOrDefault("USER_DATA_DIR", "./user-data"), "Browser session directory")
	debug := flag.Bool("debug", os.Getenv("DEBUG") == "true", "Enable debug logging")

	flag.Parse()

	if *apiKey == "" {
		fmt.Println("❌ ZAI_API_KEY не установлен")
		fmt.Println("Использование: ZAI_API_KEY=your-key go run ./cmd/agent")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log, err := logger.New(*debug)
	if err != nil {
		fmt.Printf("❌ Ошибка создания логгера: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	browserCfg := &types.BrowserConfig{
		UserDataDir: *userDataDir,
		Headless:    false,
		Timeout:     30 * time.Second,
		Debug:       *debug,
	}
	browserMgr := browser.NewManager(browserCfg, log)

	fmt.Println("🚀 Запуск браузера...")
	if err := browserMgr.Launch(ctx); err != nil {
		log.Error("Ошибка запуска браузера", err)
		os.Exit(1)
	}
	defer browserMgr.Close()

	llmCfg := &types.LLMConfig{
		APIKey:         *apiKey,
		BaseURL:        *baseURL,
		Model:          *model,
		MaxTokens:      4000,
		Temperature:    0.7,
		MaxRetries:     3,
		RequestTimeout: 60 * time.Second,
	}
	llmClient, err := llm.NewClient(llmCfg, log)
	if err != nil {
		log.Error("Ошибка создания LLM клиента", err)
		os.Exit(1)
	}

	ext := extractor.New(browserMgr.GetPage(), log)

	agentCfg := &types.AgentConfig{
		MaxRetries:           3,
		Timeout:              30 * time.Second,
		SecurityEnabled:      true,
		ConfirmationRequired: true,
		ContextBudget:        4000,
		ContextWindow:        8000,
		SummaryEnabled:       false,
		SummarizeEvery:       0,
		MaxSteps:             50,
	}
	ag := agent.New(browserMgr, ext, llmClient, log, agentCfg)

	fmt.Println()
	fmt.Println("🤖 Browser AI Agent v1.0")
	fmt.Printf("🌐 Браузер запущен (сессия: %s)\n", *userDataDir)
	fmt.Printf("🧠 Модель: %s\n", *model)
	fmt.Printf("🌐 baseURL Api модели: %s\n", *baseURL)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("🤖 Введите задачу (или 'exit'): ")

		if !scanner.Scan() {
			break
		}

		task := strings.TrimSpace(scanner.Text())
		if task == "" {
			continue
		}
		if task == "exit" || task == "quit" || task == "q" {
			break
		}

		fmt.Println()

		if err := ag.Run(ctx, task); err != nil {
			if errors.Is(err, context.Canceled) {
				fmt.Println("\n⚠️ Прервано пользователем")
				break
			}
			log.Error("Ошибка выполнения задачи", err)
		}

		fmt.Println()
	}

	fmt.Println("👋 До свидания!")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
