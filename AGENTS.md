# AGENTS.md - Guidelines for AI Coding Assistants

This document provides build, lint, and test commands plus code style guidelines for working in this repository.

## Build / Lint / Test Commands

### Core Testing
```bash
go test ./...                           # Run all tests
go test -run TestName                 # Run single test by name
go test -v ./...                      # Verbose test output
```

### Code Quality Checks
```bash
go build -v ./...                       # Build with verbose output
go vet ./...                            # Static analysis checks
gofmt -l .                              # Check formatting (fails if reformat needed)
go mod verify                           # Verify module dependencies
go mod tidy                             # Tidy go.mod/go.sum files
```

### CI/CD Pipeline
The GitHub workflows in `.github/workflows/` run:
1. `test.yml` (push/PR to main, manual): `go mod verify`, `go vet ./...`, `go test -race ./...`
2. `lint.yml` (push/PR to main, manual): `golangci-lint run ./...` (golangci-lint-action@v7, v2.14.0; config in `.golangci.yml`)
3. `build.yml` (PR to main, manual): `go mod tidy`, `go mod download`, `scripts/build-multiplatform.sh test`
4. `vuln.yml` (push/PR to main, weekly schedule, manual): `govulncheck ./...` against the Go vulnerability database
5. `release.yml` (tag push `v*`): runs tests, builds multi-platform binaries (`scripts/build-multiplatform.sh`) and RPM/DEB/APK packages (`scripts/build-packages.sh` via nfpm), uploads artifacts to a GitHub Release

All workflows use actions/checkout@v4, actions/setup-go@v5, and `go-version-file: 'go.mod'`.

## Code Style Guidelines

### Formatting & Imports
- Run code through `gofmt` before committing (use standard Go formatting)
- Group imports with a blank line between stdlib and third-party packages:
  ```go
  import (
      "fmt"
      "os"

      "github.com/dmabry/gomonitor"
  )
  ```

### Naming Conventions
- Types: `PascalCase` with descriptive names (`ExitCode`, `CheckResult`, `PerformanceMetric`)
- Method receivers: short abbreviations (e.g., `cr *CheckResult`, `ec ExitCode`)
- Variables/parameters: camelCase for local variables, consider using full words
- Constants: group related values in a single `const()` block with iota

### Type Definitions & Documentation
- Add type comments explaining the purpose of public types and their usage patterns
- Document constants describing what each exit code represents (OK=0, Warning=1, Critical=2, Unknown=3)
- Comment complex logic or non-obvious behavior in methods

### Error Handling Patterns
- Return explicit errors rather than panicking where caller may need to handle failures
- Use descriptive error messages that help with debugging: `fmt.Errorf("context: %v", err)`
- In main/CLI tools, wrap unknown states into the `Unknown` exit code when appropriate

### Performance Data & Monitoring
- Performance metrics follow the Nagios plugin specification: `'label'=value[UOM];warn;crit;min;max`, aligned with Icinga 2 output conventions (integral values print without trailing decimals; unset thresholds are omitted).
- Threshold fields (`Warn`, `Crit`, `Min`, `Max`) on `PerformanceMetric` are optional `*float64` (gomonitor v1.4.0+). Use Go 1.26 `new(expr)` at call sites (e.g., `Warn: new(float64(warnIn))`) rather than pointer helpers.
- `gomonitor.CheckResult` maintains metric insertion order internally; do not build parallel ordering slices in callers.

## Project Structure Overview

This is a collection of Nagios-compatible monitoring checks built on the `github.com/dmabry/gomonitor` library:
- **Core types** (from gomonitor): `ExitCode` (OK/Warning/Critical/Unknown), `CheckResult`, `PerformanceMetric`
- **Key methods** (from gomonitor): `NewCheckResult()`, `SetResult()`, performance data methods (`Add*`, `Update*`, `Delete*`), output methods (`FormatResult()`, `SendResult()`)
- **This repo**: `cmd/` check binaries (`check_interfaces`, `check_interface_usage`, `check_sysdescr`, `device_inventory`), `internal/snmp` client with retry/backoff, `internal/interfaces` OID definitions, multi-platform build scripts under `scripts/`
- Tests follow the table-driven test pattern

## Working with this Codebase

1. Read existing source to understand current patterns before modifying
2. Run `go vet ./...` and `gofmt -l .` after making changes
3. Add tests for new functionality following the table-driven style in `*_test.go`
4. Update documentation (README.md) when API changes affect usage examples

## Version & Compatibility

- Go version: latest stable release (keep go.mod updated)
- CI uses Ubuntu latest with actions/checkout@v3, actions/setup-go@v4