# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Octopus CLI is a command-line tool that provides local API forwarding proxy service to solve Claude Code API switching problems. It allows users to configure multiple API endpoints and keys via TOML configuration files, then dynamically switch between them without restarting Claude Code or modifying environment variables.

**Technology Stack**: Go (Golang)  
**Configuration Format**: TOML

## Project Structure

```
octopus-cli/
├── cmd/octopus/          # CLI application entry point
├── internal/             # Private application code
│   ├── config/          # TOML configuration management
│   ├── proxy/           # HTTP proxy server and forwarding engine  
│   ├── api/             # Management REST API
│   ├── cli/             # CLI commands and interfaces
│   └── utils/           # Utilities (logging, validation)
├── pkg/                 # Public packages
├── configs/             # TOML configuration files
├── docs/                # Project documentation
│   ├── requirements.md  # Requirements specification
│   ├── tasks.md         # Task management and roadmap
│   └── architecture.md  # Architecture design
├── Makefile             # Build scripts
└── go.mod              # Go module definition
```

## Key Architecture Components

1. **CLI Interface**: Command-line commands for service management and configuration
2. **HTTP Proxy Server**: Receives Claude Code requests and forwards to configured APIs
3. **Configuration Manager**: Manages multiple API configurations with TOML persistence  
4. **Forward Engine**: Handles actual HTTP request forwarding and response processing
5. **Management API**: RESTful API for runtime configuration switching

## Development Commands

Once Go project is initialized:
- `go mod init octopus-cli` - Initialize Go module
- `go build ./cmd/octopus` - Build the CLI application  
- `go test ./...` - Run all tests
- `go fmt ./...` - Format code
- `make build` - Build using Makefile
- `make test` - Run tests using Makefile
- `make test-coverage` - Run tests with coverage report
- `make tdd` - Run in TDD watch mode
- `make check` - Run all quality checks (fmt, lint, vet, test)

## Development Methodology

**This project strictly follows Test-Driven Development (TDD):**

### TDD Workflow
1. **Red**: Write a failing test first
2. **Green**: Write minimal code to make test pass
3. **Refactor**: Improve code while keeping tests green

### TDD Commands
- `make tdd` - Watch mode for continuous testing
- `make test-watch` - Watch and run tests on file changes
- `make test-coverage` - Generate coverage report (target: >90%)

### Testing Standards
- Every function must have corresponding tests
- Tests must be written before implementation
- Use table-driven tests for multiple scenarios
- Mock external dependencies (APIs, file system)
- Test naming: `TestFunction_Scenario_Expected`

## CLI Commands (Planned)

- `octopus start` - Start the proxy service
- `octopus stop` - Stop the proxy service  
- `octopus status` - Show service status
- `octopus config list` - List all API configurations
- `octopus config add` - Add new API configuration
- `octopus config switch <id>` - Switch to specific API configuration
- `octopus config remove <id>` - Remove API configuration

## Development Workflow

Follow the task phases defined in `docs/tasks.md`:
1. Phase 2: Go project initialization
2. Phase 3: Core architecture implementation
3. Phase 4: Management functionality development
4. Phase 5: User experience optimization
5. Phase 6: Testing and documentation
6. Phase 7: Release preparation

## Configuration

The CLI tool uses TOML configuration files in `configs/`:
- `octopus.toml` - Main service configuration with API endpoints
- Each API config includes: ID, name, URL, API key, and active status

Example TOML configuration:
```toml
[server]
port = 8080
log_level = "info"

[[apis]]
id = "official"
name = "Anthropic Official"
url = "https://api.anthropic.com"
api_key = "sk-xxx"
is_active = true

[[apis]]
id = "proxy1"
name = "Proxy Service 1" 
url = "https://api.proxy1.com"
api_key = "pk-xxx"
is_active = false

[settings]
active_api = "official"
```

## Core Features

- Command-line interface for easy management
- Dynamic API switching without restarts
- Multiple API endpoint configuration via TOML
- Health checking for configured APIs  
- Management REST API for configuration
- Request/response logging and monitoring