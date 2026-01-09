.PHONY: build run test clean install help install-memory install-crush install-gemini check-gemini check-crush install-all

# Binary name
BINARY_NAME=tm-tui
BINARY_PATH=./bin/$(BINARY_NAME)

# Crush repository
CRUSH_REPO=github.com/charmbracelet/crush

# Build flags
BUILD_FLAGS=-v

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	@go build $(BUILD_FLAGS) -o $(BINARY_PATH) ./cmd/tm-tui
	@go build $(BUILD_FLAGS) -o ./bin/memory ./cmd/memory

run: build ## Build and run the TUI
	@echo "Starting $(BINARY_NAME)..."
	@$(BINARY_PATH)

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@go clean

install-crush: ## Install the Crush CLI (required for task execution)
	@echo "Installing Crush CLI..."
	@go install $(CRUSH_REPO)@latest
	@echo "Crush installed successfully. Verify with: crush --help"

install: build install-crush ## Install tm-tui and dependencies to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@go install ./cmd/tm-tui

install-memory: ## Install the memory tool for LLM agents
	@echo "Installing memory tool..."
	@go build -o $(GOPATH)/bin/memory ./cmd/memory

install-all: install install-memory install-gemini ## Install all binaries (tm-tui, memory, crush, gemini)
	@echo "All tools installed successfully."

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

lint: fmt vet ## Run all linters

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

update-deps: ## Update dependencies
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

check-crush: ## Check if Crush CLI is installed
	@which crush > /dev/null 2>&1 && echo "✓ Crush is installed at: $$(which crush)" || echo "✗ Crush not found. Run 'make install-crush' to install."

install-gemini: ## Install the Gemini CLI
	@echo "Installing Gemini CLI..."
	@if command -v gemini >/dev/null 2>&1; then \
		echo "Gemini CLI already installed"; \
	else \
		go install github.com/google-gemini/gemini-cli/cmd/gemini@latest; \
		echo "Gemini CLI installed successfully"; \
	fi

check-gemini: ## Check if Gemini CLI is installed
	@echo "Checking for Gemini CLI..."
	@if command -v gemini >/dev/null 2>&1; then \
		echo "✓ Gemini CLI is installed"; \
		gemini --version; \
	else \
		echo "✗ Gemini CLI is not installed"; \
		echo "Run 'make install-gemini' to install"; \
		exit 1; \
	fi

.DEFAULT_GOAL := help
