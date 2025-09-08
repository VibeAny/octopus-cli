# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Octopus CLI is a command-line tool that provides local API forwarding proxy service to solve Claude Code API switching problems. It allows users to configure multiple API endpoints and keys via TOML configuration files, then dynamically switch between them without restarting Claude Code or modifying environment variables.

**Technology Stack**: Go (Golang)  
**Configuration Format**: TOML

## Project Structure

```
octopus-cli/
├── cmd/                  # CLI application entry point
├── internal/             # Private application code
│   ├── config/          # TOML configuration management
│   ├── proxy/           # HTTP proxy server and forwarding engine  
│   ├── process/         # Process management and daemon control
│   ├── state/           # Application state management
│   └── utils/           # Utilities (logging, validation, formatter)
├── pkg/                 # Public packages
├── configs/             # TOML configuration files
├── docs/                # Project documentation
│   ├── requirements.md  # Requirements specification
│   ├── tasks.md         # Task management and roadmap
│   └── architecture.md  # Architecture design
├── test/                # Testing documentation and reports
│   ├── functional/      # Functional test specifications
│   ├── reports/         # Test execution reports
│   ├── scripts/         # Test execution scripts
│   └── README.md        # Testing guide
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
- `go build ./cmd` - Build the CLI application  
- `go test ./...` - Run all unit tests
- `./test/scripts/run-functional-tests.sh` - Run functional tests
- `go fmt ./...` - Format code
- `make build` - Build using Makefile
- `make build-all` - Build for all platforms (8 platforms supported)
- `make test` - Run tests using Makefile
- `make test-coverage` - Run tests with coverage report
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
- `octopus config show <id>` - Show configuration details
- `octopus health` - Check API endpoints health
- `octopus logs` - View service logs
- `octopus version` - Show version information
- `octopus update` - Check and update to latest version (Phase 8)

## Development Workflow

Follow the task phases defined in `docs/tasks.md`:
1. ✅ Phase 2: Go project initialization
2. ✅ Phase 3: Core architecture implementation
3. ✅ Phase 4: Management functionality development
4. ✅ Phase 5: CLI command implementation
5. ✅ Phase 6: User experience optimization (colorized output, multi-platform builds)
6. ✅ Phase 7: Health checking and monitoring (basic functionality)
7. ⏳ Phase 8: Auto-update and CI/CD system (new requirement)
8. ⏳ Phase 9: Testing and documentation
9. ⏳ Phase 10: Release preparation

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
- Colorized CLI output with beautiful table formatting
- Multi-platform binary builds (Windows, macOS, Linux - all architectures)
- Version-controlled release naming convention
- Automatic update system with GitHub Releases integration (Phase 8)
- CI/CD pipeline with GitHub Actions for automated builds and releases