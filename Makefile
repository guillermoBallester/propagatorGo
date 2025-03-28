# Makefile
.PHONY: build run test lint clean help

# Variables
BINARY_NAME=stocksalpha
GO=go
DOCKER=docker
DOCKER_COMPOSE=docker-compose

# Default target
.DEFAULT_GOAL := help

# Build the application
build: ## Build the application
	$(GO) build -o $(BINARY_NAME) ./cmd/api

# Run the application
run: ## Run the application
	$(GO) run ./cmd/propagator

# Run tests
test: ## Run tests
	$(GO) test -v ./...

# Run linter
lint: ## Run linter
	golangci-lint run

# Clean build artifacts
clean: ## Clean build artifacts
	rm -f $(BINARY_NAME)
	rm -f coverage.txt

# Docker commands
docker-build: ## Build docker image
	$(DOCKER) build -t $(BINARY_NAME) .

docker-run: ## Run docker container
	$(DOCKER) run -p 8081:8081 $(BINARY_NAME)

docker-compose-up: ## Run with docker-compose
	$(DOCKER_COMPOSE) up -d

docker-compose-down: ## Stop docker-compose
	$(DOCKER_COMPOSE) down

# Help command
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'