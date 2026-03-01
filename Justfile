set shell := ['bash', '-euxo', 'pipefail', '-c']
set unstable := true
set positional-arguments := true

# Go project metadata

module := "pkg.krav.sh/krav"
bin_name := "krav"
bin_dir := "bin"

# Build metadata

version := `git describe --tags 2>/dev/null || git rev-parse --short HEAD 2>/dev/null || echo "DEV"`
commit := `git rev-parse --short HEAD 2>/dev/null || echo ""`
date := `date -u +%Y-%m-%d`

# ldflags for build

ldflags := "-s -w -X " + module + "/internal/buildmeta.Version=" + version + " -X " + module + "/internal/buildmeta.Commit=" + commit + " -X " + module + "/internal/buildmeta.Date=" + date

# Add GOPATH/bin to PATH for installed tools

export PATH := `go env GOPATH` + "/bin:" + env("PATH")

# Default recipe
default: lint test

# Install all development dependencies
install-tools:
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
    go install golang.org/x/pkgsite/cmd/pkgsite@latest
    go install golang.org/x/tools/cmd/goimports@latest
    go install mvdan.cc/gofumpt@latest
    go install github.com/daixiang0/gci@latest
    vale sync
    @echo "Note: plantuml and watchman must be installed separately (e.g., brew install plantuml watchman)"

# Install the krav binary to GOPATH/bin
install:
    go install -ldflags "{{ ldflags }}" ./cmd/krav

# Build the binary
build:
    go build -trimpath -ldflags "{{ ldflags }}" -o {{ bin_dir }}/{{ bin_name }} ./cmd/krav

# Build for all platforms
build-all: (build-for "linux" "amd64") (build-for "linux" "arm64") (build-for "darwin" "amd64") (build-for "darwin" "arm64") (build-for "windows" "amd64") (build-for "windows" "arm64")

# Build for a specific OS and architecture
[private]
build-for os arch:
    #!/usr/bin/env bash
    set -euo pipefail
    ext=""
    if [[ "{{ os }}" == "windows" ]]; then
        ext=".exe"
    fi
    GOOS={{ os }} GOARCH={{ arch }} go build -trimpath -ldflags "{{ ldflags }}" -o {{ bin_dir }}/{{ bin_name }}-{{ os }}-{{ arch }}${ext} ./cmd/krav

# Run the binary
run *args:
    go run -ldflags "{{ ldflags }}" ./cmd/krav "$@"

# Run the binary with debug logging enabled
run-debug *args:
    KRAV_DEBUG=1 go run -ldflags "{{ ldflags }}" ./cmd/krav "$@"

# Clean build artifacts
clean:
    rm -rf {{ bin_dir }}

# Run all linters
lint: lint-go lint-docs lint-config lint-spelling

# Run Go linters (golangci-lint)
lint-go:
    golangci-lint run

# Lint configuration files
lint-config: lint-json lint-yaml

# Lint documentation
lint-docs *args:
    just lint-markdown {{ args }}
    just lint-prose {{ args }}
    just lint-plantuml

# Lint JSON/JS/TS files
lint-json:
    biome check --files-ignore-unknown=true .

# Lint Markdown files
lint-markdown *args:
    rumdl check {{ if args == "" { "." } else { args } }}

# Lint prose in Markdown files
lint-prose *args:
    vale {{ if args == "" { "README.md" } else { args } }}

# Lint PlantUML diagrams
lint-plantuml:
    #!/usr/bin/env bash
    set -euo pipefail
    shopt -s nullglob
    files=(docs/design/diagrams/*.puml)
    if [[ ${#files[@]} -eq 0 ]]; then
        echo "No .puml files found in docs/design/diagrams/"
        exit 0
    fi
    plantuml -checkonly "${files[@]}"

# Export PlantUML diagrams to PNG and SVG
export-plantuml:
    #!/usr/bin/env bash
    set -euxo pipefail
    mkdir -p docs/design/diagrams/png docs/design/diagrams/svg
    plantuml -tpng -o "$(pwd)/docs/design/diagrams/png" docs/design/diagrams/*.puml
    plantuml -tsvg -o "$(pwd)/docs/design/diagrams/svg" docs/design/diagrams/*.puml

# Watch PlantUML diagrams and re-export on change
watch-plantuml:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Watching docs/design/diagrams/*.puml for changes. Press Ctrl-C to stop."
    watchman-wait . -m 0 -p 'docs/design/diagrams/*.puml' | while read -r f; do
        echo "Changed: $f"
        plantuml -tpng -o "$(pwd)/docs/design/diagrams/png" "$f"
        plantuml -tsvg -o "$(pwd)/docs/design/diagrams/svg" "$f"
    done

# Check spelling
lint-spelling:
    codespell

# Lint YAML files
lint-yaml:
    yamllint --strict .

# Format code
format: fmt-go format-spelling format-config format-docs

# Format Go code (uses golangci-lint formatters)
fmt-go:
    golangci-lint fmt

# Format configuration files
format-config:
    biome format --write .

# Format documentation
format-docs *args:
    just format-markdown {{ args }}

# Format Markdown files
format-markdown *args:
    rumdl fmt {{ if args == "" { "." } else { args } }}

# Fix spelling
format-spelling *args:
    codespell -w {{ if args == "" { "." } else { args } }}

# Fix code issues
fix: fix-go fix-config fix-docs

# Fix Go linting issues
fix-go:
    go fix ./...
    golangci-lint fmt
    golangci-lint run --fix

# Fix configuration files
fix-config:
    biome check --write .

# Fix documentation
fix-docs *args:
    just fix-markdown {{ args }}

# Fix Markdown files
fix-markdown *args:
    rumdl check --fix {{ if args == "" { "." } else { args } }}

# Run tests
test *args:
    go test ./... "$@"

# Run tests with verbose output
test-verbose:
    go test -v ./...

# Run tests with coverage
test-coverage:
    go test -coverprofile=coverage.out -covermode=atomic ./...
    go tool cover -html=coverage.out -o coverage.html

# Run tests with race detector
test-race:
    go test -race ./...

# Run benchmarks
bench *args:
    go test -bench=. -benchmem ./... "$@"

# Tidy go.mod
tidy:
    go mod tidy

# Verify dependencies
verify:
    go mod verify

# Update dependencies
update:
    go get -u ./...
    go mod tidy

# Check that everything is ready for commit
check: tidy verify lint test

# Start local documentation server (pkgsite)
docs:
    pkgsite -http localhost:6060

# Print version information
version:
    @echo "Version: {{ version }}"
    @echo "Commit:  {{ commit }}"
    @echo "Date:    {{ date }}"

# Run pre-commit hooks on changed files
prek:
    prek

# Run pre-commit hooks on all files
prek-all:
    prek run --all-files

# Install pre-commit hooks
prek-install:
    prek install

# Sync Vale styles and dictionaries
vale-sync:
    vale sync

# Generate full changelog
generate-changelog:
    cog changelog | { echo "# Changelog"; cat; } | rumdl check -d MD024 --fix --stdin > CHANGELOG.md

# Preview changelog since last release
preview-changelog:
    cog changelog --at $(git describe --tags)..HEAD -t full_hash | rumdl check -d MD041 --fix --stdin

# Generate release notes
[script]
generate-release-notes version="":
    v=$([[ -n "{{ version }}" ]] && echo "v{{ version }}" || echo "..$(git rev-parse HEAD)" )
    cog changelog --at $v -t full_hash | rumdl check -d MD024,MD041 --isolated --fix --stdin
