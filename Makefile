.PHONY: build run test clean install help install-memory install-crush install-task-master install-gemini check-gemini check-crush check-task-master check-deps check-project-setup init-crush-config install-all test-init

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

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v ./... -coverprofile=coverage.out -covermode=atomic
	@echo "\nCoverage Summary:"
	@go tool cover -func=coverage.out | tail -1
	@echo "\nGenerate HTML coverage report with: make coverage-html"

coverage-html: test-coverage ## Generate HTML coverage report
	@echo "Generating HTML coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	@go test -v -short ./...

test-integration: ## Run integration tests only
	@echo "Running integration tests..."
	@go test -v -run Integration ./...

test-ci: ## Run tests for CI/CD pipeline
	@echo "Running CI tests with coverage..."
	@go test -v ./... -coverprofile=coverage.out -covermode=atomic -race
	@echo "\nCoverage Summary:"
	@go tool cover -func=coverage.out | tail -1

test-suite: ## Run test suite with coverage verification (target: >80%)
	@echo "Running test suite with coverage..."
	@go test -v ./internal/executor -coverprofile=suite_coverage.out -covermode=atomic
	@echo "\nExecutor Package Coverage:"
	@go tool cover -func=suite_coverage.out | tail -1
	@echo "\nRunning full test suite..."
	@go test -v ./... -coverprofile=coverage.out -covermode=atomic
	@echo "\nOverall Coverage:"
	@go tool cover -func=coverage.out | tail -1
	@echo "\nVerifying coverage threshold (>80%)..."
	@./scripts/check-coverage.sh coverage.out 80 || echo "⚠ Coverage below 80% threshold"

test-init: build ## Run startup initialization tests
	@echo "Running startup initialization tests..."
	@./scripts/test-startup-init.sh

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@go clean

check-project-setup: ## Verify .crush.json exists in project
	@echo "Checking for .crush.json..."
	@if [ -f .crush.json ]; then \
		echo "✓ .crush.json found"; \
		exit 0; \
	else \
		echo "✗ .crush.json not found"; \
		echo "Run 'make init-crush-config' to create it"; \
		exit 1; \
	fi

init-crush-config: build ## Initialize .crush.json if missing
	@echo "Initializing Crush configuration..."
	@if [ -f .crush.json ]; then \
		echo "✓ .crush.json already exists (skipping)"; \
	else \
		echo "Creating default .crush.json..."; \
		$(BINARY_PATH) --help > /dev/null 2>&1 || true; \
		if [ -f .crush.json ]; then \
			echo "✓ .crush.json created successfully"; \
		else \
			echo "⚠ Warning: Could not create .crush.json automatically"; \
		fi; \
	fi

install-task-master: ## Install Task Master AI CLI via npm
	@echo "Installing Task Master AI CLI..."
	@if ! command -v npm >/dev/null 2>&1; then \
		echo "✗ npm not found. Please install Node.js first."; \
		echo "  Visit: https://nodejs.org/"; \
		exit 1; \
	fi
	@if command -v task-master >/dev/null 2>&1; then \
		echo "✓ Task Master CLI already installed at: $$(which task-master)"; \
	else \
		npm install -g task-master-ai; \
		if command -v task-master >/dev/null 2>&1; then \
			echo "✓ Task Master CLI installed successfully at: $$(which task-master)"; \
		else \
			echo "⚠ Warning: Installation completed but task-master not found in PATH"; \
			echo "  You may need to configure your PATH or use npx task-master"; \
		fi; \
	fi

install-crush: ## Install the Crush CLI (required for task execution)
	@echo "Installing Crush CLI..."
	@go install $(CRUSH_REPO)@latest
	@echo "Crush installed successfully. Verify with: crush --help"

install: build init-crush-config install-crush ## Install tm-tui and dependencies to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@go install ./cmd/tm-tui

install-memory: ## Install the memory tool for LLM agents
	@echo "Installing memory tool..."
	@go build -o $(GOPATH)/bin/memory ./cmd/memory

install-all: install install-memory install-task-master install-gemini ## Install all binaries (tm-tui, memory, task-master, crush, gemini)
	@echo "All tools installed successfully."

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

check-task-master: ## Check if Task Master CLI is installed
	@echo "Checking for Task Master CLI..."
	@if command -v task-master >/dev/null 2>&1; then \
		echo "✓ Task Master CLI is installed at: $$(which task-master)"; \
		task-master --version 2>/dev/null || echo "  (version command not available)"; \
	else \
		echo "✗ Task Master CLI is not installed"; \
		echo "Run 'make install-task-master' to install"; \
		exit 1; \
	fi

check-deps: ## Verify runtime dependencies
	@echo "Checking runtime dependencies..."
	@command -v task-master >/dev/null 2>&1 || \
		(echo "⚠ Warning: task-master not found. The TUI requires Task Master AI."; \
		 echo "  Install: make install-task-master")
	@echo "✓ Dependency check complete"

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
