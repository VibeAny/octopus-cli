# Contributing to Octopus CLI

Thank you for your interest in contributing to Octopus CLI! This document provides guidelines and information for contributors.

## Code of Conduct

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## Development Philosophy

This project follows **Test-Driven Development (TDD)**:

1. **Red**: Write a failing test first
2. **Green**: Write minimal code to make the test pass  
3. **Refactor**: Improve the code while keeping all tests green

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- Make

### Setting Up Development Environment

```bash
# Fork and clone the repository
git clone https://github.com/your-username/octopus-cli.git
cd octopus-cli

# Install dependencies
go mod download

# Run tests to ensure everything works
make test
```

## Development Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
```

### 2. Follow TDD Process

For every new feature or bug fix:

1. **Write a failing test first**
   ```bash
   # Create or modify test files
   # Run tests to see them fail
   go test ./...
   ```

2. **Write minimal code to pass the test**
   ```bash
   # Implement the minimal functionality
   # Run tests to see them pass
   go test ./...
   ```

3. **Refactor and improve**
   ```bash
   # Improve code quality
   # Ensure tests still pass
   go test ./...
   ```

### 3. Testing Guidelines

#### Test Naming Convention
```go
func TestFunctionName_Scenario_ExpectedBehavior(t *testing.T) {
    // Test implementation
}
```

#### Test Structure
```go
func TestConfigManager_LoadConfig_ReturnsValidConfig(t *testing.T) {
    // Arrange
    expected := &Config{...}
    
    // Act
    result, err := configManager.LoadConfig()
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

#### Table-Driven Tests
```go
func TestAPIConfig_Validate(t *testing.T) {
    tests := []struct {
        name        string
        config      APIConfig
        expectError bool
    }{
        {
            name:        "valid config",
            config:      APIConfig{ID: "test", URL: "https://api.test.com"},
            expectError: false,
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 4. Code Quality Standards

#### Required Checks
- [ ] All tests pass (`make test`)
- [ ] Test coverage > 90% (`make test-coverage`)
- [ ] Code is formatted (`make fmt`)
- [ ] No lint errors (`make lint`)
- [ ] Documentation is updated

#### Code Style
- Follow standard Go conventions
- Use `gofmt` and `goimports`
- Write clear, self-documenting code
- Add comments for exported functions
- Keep functions small and focused

### 5. Commit Guidelines

Use [Conventional Commits](https://www.conventionalcommits.org/) format:

```
type(scope): description

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or modifying tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(config): add TOML configuration support
fix(proxy): handle connection timeout errors
test(cli): add tests for config commands
docs: update installation instructions
```

### 6. Pull Request Process

1. **Ensure your branch is up to date**
   ```bash
   git checkout main
   git pull origin main
   git checkout feature/your-feature-name
   git rebase main
   ```

2. **Run all checks**
   ```bash
   make check  # This should run tests, lint, format checks
   ```

3. **Create a Pull Request**
   - Provide a clear title and description
   - Link to related issues
   - Include screenshots if applicable
   - Request review from maintainers

4. **Address Review Feedback**
   - Make requested changes
   - Keep the same commit structure
   - Update tests as needed

## Project Structure

```
octopus-cli/
├── cmd/octopus/           # CLI application entry point
├── internal/              # Private application code
│   ├── config/           # Configuration management
│   ├── proxy/            # HTTP proxy server
│   ├── cli/              # CLI command handlers
│   └── utils/            # Utility functions
├── pkg/                  # Public packages (if any)
├── configs/              # Configuration file templates
├── docs/                 # Documentation
├── scripts/              # Build and utility scripts
└── test/                 # Test utilities and fixtures
```

## Testing Strategy

### Test Pyramid
- **Unit Tests (70%)**: Test individual functions and methods
- **Integration Tests (20%)**: Test component interactions
- **End-to-End Tests (10%)**: Test complete user scenarios

### Mock Strategy
- Use interfaces for external dependencies
- Mock HTTP clients for API calls
- Mock file system operations
- Use testify/mock for complex mocking scenarios

### Test Categories

#### Unit Tests
```bash
# Run unit tests
go test ./internal/...

# Run with coverage
go test -cover ./internal/...
```

#### Integration Tests
```bash
# Run integration tests
go test -tags=integration ./test/integration/...
```

#### CLI Tests
```bash
# Run CLI command tests
go test ./cmd/...
```

## Documentation

### Required Documentation Updates

When contributing:
- Update README.md if adding new features
- Add godoc comments for exported functions
- Update architecture docs for significant changes
- Add examples for new CLI commands

### Documentation Style
- Use clear, concise language
- Provide code examples
- Include expected outputs
- Document error conditions

## Release Process

Releases are managed by maintainers:

1. Version is tagged using semantic versioning (v1.2.3)
2. GitHub Actions builds binaries for multiple platforms
3. Release notes are automatically generated
4. Binaries are published to GitHub Releases

## Getting Help

- 📖 Read the [documentation](docs/)
- 🐛 Search [existing issues](https://github.com/your-username/octopus-cli/issues)
- 💬 Start a [discussion](https://github.com/your-username/octopus-cli/discussions)
- 📧 Contact maintainers

## Recognition

Contributors are recognized in:
- README.md contributors section
- Release notes for their contributions
- GitHub contributor statistics

Thank you for contributing to Octopus CLI! 🐙