# Contributing to gochecks

Thank you for your interest in contributing to this project! This document outlines the process for contributing.

## Getting Started

### Prerequisites

- Go 1.24+
- Git
- Access to a SNMP-enabled device for integration testing

### Setting up your development environment

```bash
git clone https://github.com/dmabry/gochecks.git
cd gochecks
go mod download
```

## Coding Standards

### Code Style Guidelines

This project follows standard Go formatting via `gofmt`. Before committing:

1. Run all code through standard form:
   ```bash
   gofmt -s .
   ```

2. Run linting checks:
   ```bash
   golangci-lint run ./...
   # or without golangci-lint installed:
   go vet ./... && staticcheck ./...
   ```

3. Ensure tests pass:
   ```bash
   go test ./...
   ```

### Import Organization

Group imports with a blank line between stdlib and third-party packages:

```go
import (
    "fmt"
    "os"

    "github.com/example/somelib"
)
```

### Naming Conventions

- **Types**: `PascalCase` with descriptive names (`ExitCode`, `CheckResult`)
- **Variables/Parameters**: `camelCase`
- **Constants**: Group related values in a single block with iota
- **Receiver names**: Short abbreviations (e.g., `cr *CheckResult`)

### Error Handling

- Return explicit errors rather than panicking where caller may need to handle failures
- Use descriptive error messages: `fmt.Errorf("context: %w", err)`

## Pull Request Process

1. Ensure all linting and tests pass before opening a PR
2. Update documentation if the API changes
3. Provide a clear description of changes in your PR

### Opening a Pull Request

When you open a pull request, please:

- Use a descriptive title
- Include screenshots or CLI output for UI changes
- Link any related issues

## Issue Reporting

Use the [issue tracker](https://github.com/dmabry/gochecks/issues) to report bugs or suggest enhancements.

### Bug Reports

Should include:
- A clear description of the issue
- Steps to reproduce
- Expected vs actual behavior
- Go version and OS details

## Project Structure

```
gochecks/
├── cmd/              # CLI commands (one per subcommand)
│   ├── check_interfaces/
│   └── ...
├── internal/        # Internal packages
│   ├── interfaces/
│   └── snmp/
└── scripts/          # Build and utility scripts
```

### Adding New Checks

1. Create a new command in `cmd/<check_name>/`
2. Follow existing patterns for SNMP-based checks
3. Add comprehensive tests
4. Update README with usage documentation