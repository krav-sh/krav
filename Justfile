set shell := ['bash', '-euxo', 'pipefail', '-c']
set unstable
set positional-arguments

project := "arci"


# List available recipes
default:
  @just --list

# Format code
format: format-spelling format-config format-docs

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
fix: fix-config fix-docs

# Fix configuration files
fix-config:
  biome check --write .

# Fix documentation
fix-docs *args:
  just fix-markdown {{ args }}

# Fix Markdown files
fix-markdown *args:
  rumdl check --fix {{ if args == "" { "." } else { args } }}

# Run all linters
lint: lint-docs lint-config lint-spelling

# Lint configuration files
lint-config: lint-json lint-yaml

# Lint documentation
lint-docs *args:
  just lint-markdown {{ args }}
  just lint-prose {{ args }}

# Lint JSON/JS/TS files
lint-json:
  biome check --files-ignore-unknown=true .

# Lint Markdown files
lint-markdown *args:
  rumdl check {{ if args == "" { "." } else { args } }}

# Lint prose in Markdown files
lint-prose *args:
  vale {{ if args == "" { "README.md" } else { args } }}

# Check spelling
lint-spelling:
  codespell

# Lint YAML files
lint-yaml:
  yamllint --strict .

# Install dependencies
install:
  vale sync

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
