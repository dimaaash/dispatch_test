.PHONY: test coverage lint build clean benchmark help

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the project
	go build -v ./...

test: ## Run all tests
	go test -v ./...

coverage: ## Run tests with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

clean: ## Clean build artifacts
	go clean
	rm -f coverage.out coverage.html

deps: ## Download dependencies
	go mod download
	go mod verify

tidy: ## Tidy up dependencies
	go mod tidy


.DEFAULT_GOAL := 
