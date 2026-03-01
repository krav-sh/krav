# Testing strategy

This document describes how to test arci at different levels: unit testing the core engine, integration testing the shells, end-to-end testing the complete system, and—importantly—how policy authors can test their configurations before deploying them.

## Testing philosophy

arci's functional core/imperative shell architecture enables a clean testing strategy. The core is pure functions operating on data structures, making it trivial to test without mocking. The shells handle I/O and are tested through integration tests that exercise real file systems, HTTP endpoints, and subprocesses.

Policy authors face a different challenge: they need to verify their configurations work before an Claude Code triggers them in production. arci provides tooling for this workflow, from quick validation to comprehensive regression testing.

## Test organization

Go's testing conventions shape how we organize tests. Test files live alongside the code they test with a `_test.go` suffix. Tests in the same package test internal behavior; tests in a `_test` package exercise the public API. This separation keeps fast unit tests close to the implementation while package-level tests verify the public interface.

Running all tests uses `go test ./...`. Running tests for a specific package uses `go test ./internal/core/...`. Running a single test by name uses `go test -run TestName ./...`. The `-v` flag shows verbose output during test runs, useful for debugging.

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/core/...

# Run a specific test
go test -run TestExpressionEvaluatesBooleanAnd ./...

# Run tests with verbose output
go test -v ./...

# Run only integration tests (using build tag)
go test -tags=integration ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Core engine tests

The core engine contains pure functions with no I/O, making tests straightforward. Each package includes `_test.go` files with unit tests using standard `func TestXxx(t *testing.T)` signatures. Go tests run concurrently by default when `t.Parallel()` is called.

### Expression evaluation

Expression tests verify the CEL expression engine and Go template system handle all supported operators and data types correctly.

### Policy compilation

Policy compilation tests verify that policies are validated and prepared for efficient evaluation. The engine compiles policies from their YAML representation into optimized structures with pre-parsed CEL expressions.

### Result aggregation

Result aggregation tests verify that validation results from multiple policies are combined correctly, and that mutations are applied in the proper order.

### Property-based testing

The `rapid` package enables property-based testing for expression edge cases and input handling.

### Custom function registration

Custom function tests verify that extension functions integrate correctly with the expression engine.

## Shell integration tests

Integration tests live in the `tests/` directory and exercise the public API of each shell crate. These tests may touch the filesystem, spawn processes, or make HTTP requests.

### CLI tests

CLI integration tests spawn the `arci` binary and verify its behavior with real configuration files.

### Daemon tests

Daemon integration tests use `net/http/httptest` to exercise the HTTP API.

### Filesystem fixtures

Go's `t.TempDir()` provides isolated directories for tests that need configuration files.

### Mocking with traits

Go interfaces and manual mocks provide test doubles for components with external dependencies.

## End-to-end tests

End-to-end tests verify the complete system from CLI input through daemon processing to final response. These tests live in `tests/e2e/` and may spawn actual processes.

## Policy testing for authors

Policy authors need confidence their configurations work before deployment. arci provides CLI commands and a dashboard interface for testing policies.

### Validation

The `config validate` command checks syntax and schema without evaluating policies.

```bash
$ arci config validate
Checking .arci/policies.d/security.yaml... OK
Checking ~/.config/arci/policies.d/defaults.yaml... OK
All configurations valid.

$ arci config validate --strict
Checking .arci/policies.d/security.yaml...
  Warning: policy 'block-rm' shadows built-in policy
  Warning: unused variable 'old_path' in policy 'file-safety'
1 warning(s), 0 error(s)
```

### Dry-run evaluation

The `policy test` command evaluates a specific policy against sample input.

```bash
$ arci hook policy test block-dangerous-commands --input '{"event_type":"pre_tool_call","tool_name":"Bash","tool_input":{"command":"rm -rf /"}}'
Policy: block-dangerous-commands
Rule: block-rm-rf
Input: {"event_type":"pre_tool_call","tool_name":"Bash","tool_input":{"command":"rm -rf /"}}
Match: tools=[Bash] ✓
Condition: !tool_input.command.matches("rm\\s+-rf\\s+/") → false
Result: deny
Message: "Dangerous recursive delete"
```

### Interactive testing

The dashboard provides a policy tester interface where authors can paste hook inputs and see which policies match, in what order, which rules fire, and what the final decision would be.

### Regression testing

For comprehensive testing, authors can maintain a directory of test cases.

```yaml
# tests/policies/dangerous-commands.yaml
name: Dangerous command blocking
policy: block-dangerous-commands

cases:
  - name: blocks rm -rf root
    input:
      event_type: pre_tool_call
      tool_name: Bash
      tool_input:
        command: "rm -rf /"
    expect:
      action: deny
      message_contains: "Dangerous recursive delete"

  - name: allows safe rm
    input:
      event_type: pre_tool_call
      tool_name: Bash
      tool_input:
        command: "rm tempfile.txt"
    expect:
      action: allow

  - name: blocks format command
    input:
      event_type: pre_tool_call
      tool_name: Bash
      tool_input:
        command: "mkfs.ext4 /dev/sda1"
    expect:
      action: deny
```

Run with `arci hook policy test --suite tests/policies/`.

## CI/CD integration

### Pre-commit hook

Add to `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: local
    hooks:
      - id: arci-validate
        name: Validate arci configuration
        entry: arci config validate
        language: system
        files: '\.arci/.*\.yaml$'
        pass_filenames: false
```

### GitHub Actions

```yaml
name: Validate arci policies

on:
  push:
    paths:
      - '.arci/**'
  pull_request:
    paths:
      - '.arci/**'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install arci
        run: go install github.com/tbhb/arci@latest

      - name: Validate configuration
        run: arci config validate --strict

      - name: Run policy tests
        run: arci hook policy test --suite tests/policies/
```

## Extension testing

Extension authors test custom functions and action handlers before publication.

## Test data and fixtures

### Generating fixtures

Capture real hook events for replay testing using the `--capture` flag.

```bash
arci hook apply --capture=captured-events.jsonl < hook-input.json
```

### Fixture organization

```text
tests/
  fixtures/
    events/
      bash-safe.json
      bash-dangerous.json
      read-allowed.json
      write-sensitive.json
    policies/
      minimal.yaml
      full-featured.yaml
      with-parameters.yaml
    expected/
      bash-safe.json
      bash-dangerous.json
```

### Anonymizing captured data

The `arci capture anonymize` command removes sensitive information from captured events.

```bash
$ arci capture anonymize captured-events.jsonl --output=anonymized.jsonl
Anonymized 47 events:
  - Replaced 12 file paths
  - Replaced 3 environment variables
  - Replaced 8 usernames
```

---

This testing strategy provides confidence at every level: fast unit tests for the core, integration tests for the shells, end-to-end tests for the complete system, and practical tooling for policy authors. Coverage targets are 90% for the core packages and 80% for shell packages, enforced in CI.
