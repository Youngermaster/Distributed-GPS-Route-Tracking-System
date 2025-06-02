# Variables
CARGO := cargo
BINARY_NAME := data_ingestion_microservice

# Default target
.DEFAULT_GOAL := help

# Help command
.PHONY: help
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development commands
.PHONY: format
format: ## Format the code
	$(CARGO) fmt --quiet

.PHONY: lint
lint: ## Run clippy lints
	$(CARGO) clippy -- -D warnings

.PHONY: check
check: ## Check if the code compiles
	$(CARGO) check

.PHONY: test
test: ## Run tests
	$(CARGO) test

.PHONY: build
build: ## Build the application in debug mode
	$(CARGO) build

.PHONY: build-release
build-release: ## Build the application in release mode
	$(CARGO) build --release

.PHONY: run
run: ## Run the application
	$(CARGO) run

.PHONY: run-release
run-release: ## Run the application in release mode
	$(CARGO) run --release

.PHONY: clean
clean: ## Clean build artifacts
	$(CARGO) clean

# Docker commands
.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t $(BINARY_NAME):latest .

.PHONY: docker-run
docker-run: ## Run the application in Docker
	docker run --rm -it $(BINARY_NAME):latest

# Development environment
.PHONY: dev-start
dev-start: ## Start development services
	@echo "Starting development services..."
	cd .. && docker-compose up -d

.PHONY: dev-stop
dev-stop: ## Stop development services
	@echo "Stopping development services..."
	cd .. && docker-compose down

# Quality checks
.PHONY: check-all
check-all: format lint check test ## Run all quality checks

# Release workflow
.PHONY: release
release: check-all build-release ## Complete release workflow