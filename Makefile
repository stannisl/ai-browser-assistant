.PHONY: build run test lint clean tidy fmt help

# Variables
BINARY_NAME=agent
BUILD_DIR=bin
GO=go
GOLINT=golangci-lint

# Default target
all: tidy fmt build

help:
	@echo "Available targets:"
	@echo "  make all        - Run tidy, fmt, build (default)"
	@echo "  make build      - Build the agent binary"
	@echo "  make run        - Run the agent binary"
	@echo "  make test       - Run tests"
	@echo "  make lint       - Run linter"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make tidy       - Download and tidy dependencies"
	@echo "  make fmt        - Format code"
	@echo "  make install    - Install binary to GOPATH/bin"
	@echo "  make build-all  - Build for all platforms"
	@echo "  make test-verbose - Run tests with verbose output"
	@echo "  make test-coverage - Run tests with coverage"

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/agent

run:
	@echo "Running $(BINARY_NAME)..."
	$(GO) run ./cmd/agent

test:
	@echo "Running tests..."
	$(GO) test ./... -v

test-verbose:
	@echo "Running tests with verbose output..."
	$(GO) test ./... -v

test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test ./... -coverprofile=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint:
	@echo "Running linter..."
	$(GOLINT) run

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)/*
	@rm -f coverage.out coverage.html
	@rm -rf user-data/

tidy:
	@echo "Tidying dependencies..."
	$(GO) mod tidy

fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

install:
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install ./cmd/agent

build-all:
	@echo "Building $(BINARY_NAME) for all platforms..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME)-linux ./cmd/agent
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME)-windows.exe ./cmd/agent
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin ./cmd/agent
	@echo "Builds completed in $(BUILD_DIR)/"
