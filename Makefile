# Makefile for Octopus CLI

.PHONY: help build test test-coverage clean fmt lint check install dev

# Variables
BINARY_NAME=octopus
VERSION?=dev
LDFLAGS=-ldflags "-X main.version=${VERSION}"
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Default target
help: ## Show this help message
	@echo "Octopus CLI Development Commands"
	@echo "================================"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development commands
dev: ## Run in development mode
	go run ./cmd

build: ## Build the binary
	@echo "Building ${BINARY_NAME}..."
	go build ${LDFLAGS} -o ${BINARY_NAME} ./cmd

build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-linux-amd64 ./cmd
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-darwin-amd64 ./cmd
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-darwin-arm64 ./cmd
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o dist/${BINARY_NAME}-windows-amd64.exe ./cmd

install: build ## Install the binary to $GOPATH/bin
	@echo "Installing ${BINARY_NAME}..."
	go install ${LDFLAGS} ./cmd

# Testing commands
test: ## Run all tests
	@echo "Running tests..."
	go test -v ./...

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	go test -v -short ./...

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	go test -v -tags=integration ./test/integration/...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -coverprofile=${COVERAGE_FILE} ./...
	go tool cover -html=${COVERAGE_FILE} -o ${COVERAGE_HTML}
	@echo "Coverage report generated: ${COVERAGE_HTML}"

test-coverage-func: ## Show test coverage by function
	@echo "Test coverage by function..."
	go test -coverprofile=${COVERAGE_FILE} ./... > /dev/null
	go tool cover -func=${COVERAGE_FILE}

# Code quality commands
fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

check: fmt lint vet test ## Run all quality checks

# Dependency management
mod-tidy: ## Tidy go modules
	@echo "Tidying modules..."
	go mod tidy

mod-download: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download

mod-verify: ## Verify dependencies
	@echo "Verifying dependencies..."
	go mod verify

# Clean up
clean: ## Clean build artifacts
	@echo "Cleaning up..."
	rm -f ${BINARY_NAME}
	rm -f ${COVERAGE_FILE}
	rm -f ${COVERAGE_HTML}
	rm -rf dist/

clean-all: clean ## Clean everything including vendor
	rm -rf vendor/

# Docker commands (optional)
docker-build: ## Build Docker image
	docker build -t octopus-cli:${VERSION} .

docker-run: ## Run in Docker container
	docker run --rm -it octopus-cli:${VERSION}

# Release commands
release-dry: ## Dry run release
	goreleaser release --snapshot --rm-dist --skip-publish

release: ## Create a release
	goreleaser release --rm-dist

# Development utilities
watch: ## Watch for changes and rebuild
	@echo "Watching for changes..."
	air

tools: ## Install development tools
	@echo "Installing development tools..."
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/cosmtrek/air@latest
	go install github.com/goreleaser/goreleaser@latest

init: tools mod-download ## Initialize development environment
	@echo "Development environment initialized!"

# TDD specific commands
tdd: ## Run tests in watch mode (TDD)
	@echo "Running TDD mode - watching for changes..."
	air -c .air-test.toml

test-watch: ## Watch and run tests on file changes
	@echo "Watching tests..."
	find . -name "*.go" | entr -r make test

# Benchmarking
bench: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Security
security: ## Run security checks
	@echo "Running security checks..."
	gosec ./...

# Generate
generate: ## Run go generate
	@echo "Running go generate..."
	go generate ./...