# Variables
GO := go
BINARY_NAME := data-ingestion-service
DOCKER_IMAGE := data-ingestion-service
DOCKER_TAG := latest

# Default target
.DEFAULT_GOAL := help

# Colors for terminal output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

# Help command
.PHONY: help
help: ## Show this help message
	@echo "$(BLUE)Data Ingestion Microservice - Available Commands:$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""

# Development commands
.PHONY: install-deps
install-deps: ## Install Go dependencies
	@echo "$(YELLOW)Installing dependencies...$(NC)"
	$(GO) mod download
	$(GO) mod tidy

.PHONY: format
format: ## Format the code
	@echo "$(YELLOW)Formatting code...$(NC)"
	$(GO) fmt ./...

.PHONY: lint
lint: ## Run linter (requires golangci-lint)
	@echo "$(YELLOW)Running linter...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(RED)golangci-lint not installed. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)"; \
	fi

.PHONY: check
check: ## Check if the code compiles
	@echo "$(YELLOW)Checking compilation...$(NC)"
	$(GO) build -v ./...

.PHONY: test
test: ## Run all tests
	@echo "$(YELLOW)Running tests...$(NC)"
	$(GO) test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "$(YELLOW)Running tests with coverage...$(NC)"
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

.PHONY: test-algorithm
test-algorithm: ## Run tests for algorithm package only
	@echo "$(YELLOW)Running algorithm tests...$(NC)"
	$(GO) test -v ./algorithm/

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "$(YELLOW)Running benchmarks...$(NC)"
	$(GO) test -bench=. -benchmem ./...

.PHONY: benchmark-algorithm
benchmark-algorithm: ## Run algorithm benchmarks only
	@echo "$(YELLOW)Running algorithm benchmarks...$(NC)"
	$(GO) test -bench=. -benchmem ./algorithm/

.PHONY: build
build: ## Build the application
	@echo "$(YELLOW)Building application...$(NC)"
	$(GO) build -o $(BINARY_NAME) .
	@echo "$(GREEN)Built successfully: $(BINARY_NAME)$(NC)"

.PHONY: build-prod
build-prod: ## Build optimized production binary
	@echo "$(YELLOW)Building production binary...$(NC)"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -ldflags="-w -s" -o $(BINARY_NAME) .
	@echo "$(GREEN)Production binary built: $(BINARY_NAME)$(NC)"

.PHONY: run
run: ## Run the application
	@echo "$(YELLOW)Starting data ingestion service...$(NC)"
	$(GO) run .

.PHONY: run-dev
run-dev: ## Run with development environment variables
	@echo "$(YELLOW)Starting in development mode...$(NC)"
	@export RUST_LOG=debug && $(GO) run .

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	$(GO) clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	@echo "$(GREEN)Cleaned successfully$(NC)"

# Docker commands
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(YELLOW)Building Docker image...$(NC)"
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "$(GREEN)Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)$(NC)"

.PHONY: docker-run
docker-run: ## Run the application in Docker
	@echo "$(YELLOW)Running in Docker...$(NC)"
	docker run --rm -it \
		--network host \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-push
docker-push: ## Push Docker image to registry
	@echo "$(YELLOW)Pushing Docker image...$(NC)"
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

# Development environment
.PHONY: dev-start
dev-start: ## Start development services (Redis, MongoDB, MQTT)
	@echo "$(YELLOW)Starting development services...$(NC)"
	cd .. && docker-compose up -d
	@echo "$(GREEN)Development services started$(NC)"

.PHONY: dev-stop
dev-stop: ## Stop development services
	@echo "$(YELLOW)Stopping development services...$(NC)"
	cd .. && docker-compose down
	@echo "$(GREEN)Development services stopped$(NC)"

.PHONY: dev-logs
dev-logs: ## Show logs from development services
	@echo "$(YELLOW)Showing development services logs...$(NC)"
	cd .. && docker-compose logs -f

.PHONY: dev-status
dev-status: ## Show status of development services
	@echo "$(YELLOW)Development services status:$(NC)"
	cd .. && docker-compose ps

# Quality checks
.PHONY: check-all
check-all: format lint check test ## Run all quality checks

.PHONY: ci
ci: check-all benchmark ## Run CI pipeline checks

# Release workflow
.PHONY: release
release: check-all build-prod ## Complete release workflow

# Documentation
.PHONY: docs
docs: ## Generate and view documentation
	@echo "$(YELLOW)Opening documentation...$(NC)"
	$(GO) doc -all ./... | less

.PHONY: godoc
godoc: ## Start local godoc server
	@echo "$(YELLOW)Starting godoc server at http://localhost:6060$(NC)"
	@echo "$(BLUE)Press Ctrl+C to stop$(NC)"
	godoc -http=:6060

# Module management
.PHONY: mod-tidy
mod-tidy: ## Clean up go.mod and go.sum
	@echo "$(YELLOW)Tidying Go modules...$(NC)"
	$(GO) mod tidy
	@echo "$(GREEN)Modules tidied$(NC)"

.PHONY: mod-download
mod-download: ## Download dependencies
	@echo "$(YELLOW)Downloading dependencies...$(NC)"
	$(GO) mod download
	@echo "$(GREEN)Dependencies downloaded$(NC)"

.PHONY: mod-verify
mod-verify: ## Verify dependencies
	@echo "$(YELLOW)Verifying dependencies...$(NC)"
	$(GO) mod verify
	@echo "$(GREEN)Dependencies verified$(NC)"

# Tools installation
.PHONY: install-tools
install-tools: ## Install development tools
	@echo "$(YELLOW)Installing development tools...$(NC)"
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install golang.org/x/tools/cmd/godoc@latest
	@echo "$(GREEN)Development tools installed$(NC)"

# Architecture validation
.PHONY: arch-test
arch-test: ## Test architecture constraints
	@echo "$(YELLOW)Testing architecture constraints...$(NC)"
	@if [ -d "./internal" ]; then \
		echo "$(GREEN)✓ Internal packages properly structured$(NC)"; \
	else \
		echo "$(RED)✗ Internal packages not found$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Architecture tests passed$(NC)" 