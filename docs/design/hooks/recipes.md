# Recipes and examples

This document collects practical examples of arci policies for common use cases. Each recipe includes complete policy definitions, explanations of the approach, and notes on customization.

## Safety policies

These policies help prevent dangerous operations and provide guardrails for AI assistant behavior.

### Block dangerous rm commands

This policy blocks recursive deletion targeting root or other critical system paths.

```yaml
version: 1
name: block-rm-rf-root
metadata:
  description: Block rm -rf commands targeting root filesystem
  labels:
    category: safety
    severity: critical

config:
  priority: critical

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'

rules:
  - name: block-root-deletion
    validate:
      expression: '!command.matches("^rm\\s+(-[rRf]+\\s+)*/($|\\s)")'
      message: "Blocked rm command targeting root filesystem"
      action: deny
```

The regex pattern matches `rm` followed by any combination of `-r`, `-R`, and `-f` flags, then a path starting with `/`. Critical priority ensures this policy evaluates before lower-priority policies.

To allow deletion of specific paths like `/tmp`, use a more permissive pattern:

```yaml
version: 1
name: block-rm-rf-root-except-tmp
metadata:
  description: Block rm -rf commands targeting root filesystem (except /tmp)
  labels:
    category: safety
    severity: critical

config:
  priority: critical

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'

rules:
  - name: block-root-deletion
    validate:
      expression: '!command.matches("^rm\\s+(-[rRf]+\\s+)*/(?!tmp)")'
      message: "Blocked rm command targeting root filesystem (except /tmp)"
      action: deny
```

### Prevent force push to protected branches

This policy blocks force pushes to main, master, and other protected branches.

```yaml
version: 1
name: block-force-push-protected
metadata:
  description: Block force push to protected branches
  labels:
    category: safety
    severity: critical

config:
  priority: critical

match:
  events: [pre_tool_call]
  tools: [Bash]

parameters:
  - name: protectedBranches
    value: ["main", "master", "release", "production"]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: current_branch
    expression: '$current_branch()'
  - name: is_force_push
    expression: 'command.matches("git\\s+push.*--force")'
  - name: on_protected
    expression: 'current_branch in params.protectedBranches'

rules:
  - name: block-force-push
    conditions:
      - expression: 'is_force_push'
    validate:
      expression: '!on_protected'
      message: "Force push to {{ current_branch }} is not allowed"
      action: deny
```

The protected branches are configurable through the `protectedBranches` parameter, making it easy to customize for different workflows.

### Block curl/wget piped to shell

This policy blocks the dangerous pattern of piping remote scripts directly to a shell.

```yaml
version: 1
name: block-pipe-to-shell
metadata:
  description: Block piping remote content directly to shell
  labels:
    category: safety
    severity: critical

config:
  priority: critical

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'

rules:
  - name: block-pipe-to-shell
    validate:
      expression: '!command.matches("(curl|wget).*\\|.*(sh|bash)")'
      message: |
        Blocked: piping remote content directly to shell is dangerous.
        Download the script first, review it, then execute.
      action: deny
```

### Require confirmation for destructive database operations

This policy uses state tracking to warn first, then block repeated attempts. It demonstrates combining validation with effects for stateful behavior.

```yaml
version: 1
name: database-destructive-ops
metadata:
  description: Warn then block repeated destructive database operations
  labels:
    category: safety
    severity: high

config:
  priority: high

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: is_destructive
    expression: 'command.matches("(DROP|TRUNCATE|DELETE\\s+FROM)")'
  - name: attempts
    expression: '$session_get("destructive_db_ops", 0)'

conditions:
  - name: is-destructive-db-op
    expression: 'is_destructive'

rules:
  - name: track-and-limit
    validate:
      expression: 'attempts < 2'
      message: "Blocked: too many destructive database operations this session ({{ attempts }}/2)"
      action: deny
    effects:
      - type: setState
        scope: session
        key: destructive_db_ops
        value: '{{ attempts + 1 }}'
        when: always

  - name: warn-first-attempts
    conditions:
      - expression: 'attempts < 2'
    effects:
      - type: notify
        title: "Destructive database operation"
        message: "Warning: attempt {{ attempts + 1 }} of 2 allowed"
        when: on_pass
```

## Git workflow policies

Policies for enforcing git practices and preventing common mistakes.

### Convert --force to --force-with-lease

This policy automatically converts dangerous `--force` pushes to the safer `--force-with-lease` variant using a mutation rule.

```yaml
version: 1
name: safe-force-push
metadata:
  description: Convert --force to --force-with-lease for safer pushing
  labels:
    category: git
    severity: high

config:
  priority: high

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: is_force_push
    expression: 'command.matches("git\\s+push.*--force(?!-with-lease)")'

conditions:
  - name: is-unsafe-force-push
    expression: 'is_force_push'

rules:
  - name: convert-to-force-with-lease
    mutate:
      expression: |
        object.with({
          tool_input: object.tool_input.with({
            command: tool_input.command.replace("--force", "--force-with-lease")
          })
        })
    effects:
      - type: notify
        title: "Force push converted"
        message: |
          Converted --force to --force-with-lease for safety.
          This will fail if the remote has changes you haven't fetched.
        when: always
```

The negative lookahead `(?!-with-lease)` in the condition ensures we don't match commands that already use the safer option.

### Warn on commits to main/master

This policy warns when attempting to commit directly to protected branches.

```yaml
version: 1
name: warn-commit-main
metadata:
  description: Warn when committing directly to main or master
  labels:
    category: git
    severity: high

config:
  priority: high

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: current_branch
    expression: '$current_branch()'
  - name: is_commit
    expression: 'command.matches("^git\\s+commit")'
  - name: on_protected
    expression: 'current_branch.matches("^(main|master)$")'

conditions:
  - name: is-commit-to-protected
    expression: 'is_commit && on_protected'

rules:
  - name: warn-direct-commit
    validate:
      expression: 'false'
      message: |
        You are committing directly to {{ current_branch }}.
        Consider creating a feature branch instead.
      action: warn
```

### Block push without tests passing

This policy uses a Starlark macro to detect the project type and verify tests pass before allowing a push.

```yaml
version: 1
name: require-tests-before-push
metadata:
  description: Run tests before allowing git push
  labels:
    category: git
    severity: high

config:
  priority: high

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: is_push
    expression: 'command.matches("^git\\s+push")'
  - name: tests_passed
    expression: '$session_get("tests_passed", false)'

conditions:
  - name: is-git-push
    expression: 'is_push'

macros:
  - name: detect_test_command
    language: starlark
    source: |
      def detect_test_command():
          if file_exists("package.json"):
              return "npm test"
          elif file_exists("pyproject.toml"):
              return "pytest"
          elif file_exists("go.mod"):
              return "go test ./..."
          elif file_exists("Cargo.toml"):
              return "cargo test"
          return None

rules:
  - name: require-tests
    validate:
      expression: 'tests_passed || $detect_test_command() == null'
      message: "Push blocked: run tests first and ensure they pass"
      action: deny
```

For projects that want to run tests automatically, a separate policy can handle test execution and state tracking.

### Ensure commits are signed

This policy adds the GPG signing flag to commits that don't already include it.

```yaml
version: 1
name: require-signed-commits
metadata:
  description: Ensure all commits are GPG signed
  labels:
    category: git
    severity: high

config:
  priority: high

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: is_commit
    expression: 'command.matches("^git\\s+commit")'
  - name: is_signed
    expression: 'command.matches("-S|--gpg-sign")'

conditions:
  - name: is-unsigned-commit
    expression: 'is_commit && !is_signed'

rules:
  - name: add-signing-flag
    mutate:
      expression: |
        object.with({
          tool_input: object.tool_input.with({
            command: tool_input.command + " -S"
          })
        })
    effects:
      - type: notify
        title: "Commit signing"
        message: "Added -S flag to sign commit with GPG"
        when: always
```

## Python project policies

Policies tailored for Python development workflows.

### Inject project conventions on session start

This policy injects project-specific guidance when a session starts in a Python project.

```yaml
version: 1
name: python-project-context
metadata:
  description: Inject Python project conventions at session start
  labels:
    category: python
    severity: medium

config:
  priority: medium

match:
  events: [session_start]

conditions:
  - name: is-python-project
    expression: '$file_exists("pyproject.toml")'

rules:
  - name: inject-conventions
    mutate:
      expression: |
        object.with({
          context: (object.context ?? "") + "\n\n" +
            "This is a Python project using:\n" +
            "- pytest for testing (run with `pytest`)\n" +
            "- ruff for linting (run with `ruff check .`)\n" +
            "- mypy for type checking (run with `mypy .`)\n" +
            "Always run tests before committing."
        })
```

The `$file_exists()` function checks for `pyproject.toml` relative to the current working directory. Customize the injected content to match your project's actual tooling.

### Check import sorting after file writes

This policy checks for unsorted imports after Python files are written.

```yaml
version: 1
name: check-import-sorting
metadata:
  description: Warn about unsorted imports in Python files
  labels:
    category: python
    severity: medium

config:
  priority: medium

match:
  events: [post_tool_call]
  tools: [Write, Edit]

variables:
  - name: file_path
    expression: 'tool_input.file_path ?? ""'
  - name: is_python
    expression: 'file_path.endsWith(".py")'

conditions:
  - name: is-python-file
    expression: 'is_python'

rules:
  - name: check-imports
    effects:
      - type: log
        level: info
        message: "Python file written: {{ file_path }}. Consider running `ruff check --select I {{ file_path }}` to check imports."
        when: always
```

For automated checking, use a Starlark script effect that can invoke external tools:

```yaml
version: 1
name: check-import-sorting-auto
metadata:
  description: Automatically check import sorting in Python files
  labels:
    category: python
    severity: medium

config:
  priority: medium

match:
  events: [post_tool_call]
  tools: [Write, Edit]

variables:
  - name: file_path
    expression: 'tool_input.file_path ?? ""'

conditions:
  - name: is-python-file
    expression: 'file_path.endsWith(".py")'

rules:
  - name: check-imports
    effects:
      - type: script
        language: starlark
        source: |
          # Note: shell() is only available when explicitly enabled
          # This example shows the pattern for when it's available
          path = vars["file_path"]
          log("info", "Checking imports in " + path)
        when: always
```

### Validate pyproject.toml changes

This policy validates pyproject.toml syntax after modifications.

```yaml
version: 1
name: validate-pyproject
metadata:
  description: Validate pyproject.toml syntax after modifications
  labels:
    category: python
    severity: medium

config:
  priority: medium

match:
  events: [post_tool_call]
  tools: [Write, Edit]

variables:
  - name: file_path
    expression: 'tool_input.file_path ?? ""'

conditions:
  - name: is-pyproject
    expression: 'file_path.endsWith("pyproject.toml")'

rules:
  - name: validate-syntax
    effects:
      - type: log
        level: info
        message: "pyproject.toml modified. Validate syntax with: python -c \"import tomllib; tomllib.load(open('{{ file_path }}', 'rb'))\""
        when: always
```

## JavaScript/TypeScript project policies

Policies for Node.js and frontend development.

### Inject project conventions

```yaml
version: 1
name: js-project-context
metadata:
  description: Inject JavaScript/TypeScript project conventions at session start
  labels:
    category: javascript
    severity: medium

config:
  priority: medium

match:
  events: [session_start]

conditions:
  - name: is-js-project
    expression: '$file_exists("package.json")'

rules:
  - name: inject-conventions
    mutate:
      expression: |
        object.with({
          context: (object.context ?? "") + "\n\n" +
            "This is a JavaScript/TypeScript project. Please:\n" +
            "- Use npm/yarn/pnpm as appropriate for this project\n" +
            "- Follow the existing code style\n" +
            "- Run tests with `npm test` before committing"
        })
```

### Prevent npm install without lockfile update

This policy warns when `npm install` might not update the lockfile.

```yaml
version: 1
name: npm-install-lockfile
metadata:
  description: Warn about npm install without explicit lockfile handling
  labels:
    category: javascript
    severity: medium

config:
  priority: medium

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: is_npm_install
    expression: 'command.matches("npm\\s+install")'
  - name: has_lockfile_flag
    expression: 'command.matches("--package-lock|--save|--save-dev")'

conditions:
  - name: is-npm-install-without-flag
    expression: 'is_npm_install && !has_lockfile_flag'

rules:
  - name: warn-lockfile
    validate:
      expression: 'false'
      message: |
        Running npm install without explicit lockfile handling.
        Consider using `npm ci` for reproducible installs or
        `npm install --package-lock` to ensure lockfile is updated.
      action: warn
```

### Warn on package.json dependency changes

This policy alerts when dependencies are modified.

```yaml
version: 1
name: dependency-change-warning
metadata:
  description: Warn when package.json dependencies are modified
  labels:
    category: javascript
    severity: medium

config:
  priority: medium

match:
  events: [post_tool_call]
  tools: [Write, Edit]

variables:
  - name: file_path
    expression: 'tool_input.file_path ?? ""'

conditions:
  - name: is-package-json
    expression: 'file_path.endsWith("package.json")'

rules:
  - name: remind-install
    effects:
      - type: notify
        title: "package.json modified"
        message: "Remember to run `npm install` and commit the lockfile if dependencies changed"
        when: always
```

## State-dependent policies

Policies that use the state store for counting, tracking, and escalating.

### Warn once, then block

This pattern warns the user on first occurrence, then blocks on repeated violations within the same session.

```yaml
version: 1
name: escalating-rm-warning
metadata:
  description: Warn then block repeated rm -rf commands
  labels:
    category: safety
    severity: high

config:
  priority: high

match:
  events: [pre_tool_call]
  tools: [Bash]

parameters:
  - name: maxAttempts
    value: 3

variables:
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: is_rm_rf
    expression: 'command.matches("^rm\\s+-rf")'
  - name: attempts
    expression: '$session_get("rm_rf_count", 0)'

conditions:
  - name: is-rm-rf-command
    expression: 'is_rm_rf'

rules:
  - name: track-and-limit
    validate:
      expression: 'attempts < params.maxAttempts'
      message: |
        Blocked: too many recursive rm commands this session.
        This pattern was warned {{ params.maxAttempts }} times and is now blocked.
      action: deny
    effects:
      - type: setState
        scope: session
        key: rm_rf_count
        value: '{{ attempts + 1 }}'
        when: always

  - name: warn-approaching-limit
    conditions:
      - expression: 'attempts < params.maxAttempts'
    effects:
      - type: notify
        title: "Recursive rm detected"
        message: "Warning: attempt {{ attempts + 1 }} of {{ params.maxAttempts }}. This will be blocked after {{ params.maxAttempts }} occurrences."
        when: on_pass
```

The `setState` effect always runs, tracking the count. The validation blocks after the limit is reached. The notification effect only runs when the validation passes.

### Rate limiting operations

This policy limits expensive operations like API calls to a maximum rate using a Starlark script effect.

```yaml
version: 1
name: rate-limit-api-calls
metadata:
  description: Rate limit API calls to external services
  labels:
    category: safety
    severity: high

config:
  priority: high

match:
  events: [pre_tool_call]
  tools: [Bash]

parameters:
  - name: maxCallsPerMinute
    value: 10
  - name: apiDomain
    value: "api.example.com"

variables:
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: is_api_call
    expression: 'command.matches("curl.*" + params.apiDomain)'

conditions:
  - name: is-api-call
    expression: 'is_api_call'

macros:
  - name: get_rate_limit_key
    language: starlark
    source: |
      def get_rate_limit_key():
          # Use minute-granularity key for rate limiting
          import time
          minute = int(time.time() / 60)
          return "api_calls_" + str(minute)

rules:
  - name: check-rate-limit
    variables:
      - name: rate_key
        expression: '$get_rate_limit_key()'
      - name: current_count
        expression: '$session_get(rate_key, 0)'
    validate:
      expression: 'current_count < params.maxCallsPerMinute'
      message: "Rate limit exceeded: max {{ params.maxCallsPerMinute }} API calls per minute"
      action: deny
    effects:
      - type: setState
        scope: session
        key: '{{ rate_key }}'
        value: '{{ current_count + 1 }}'
        when: on_pass
```

### Session-scoped context accumulation

This policy builds up context about what the assistant has done during a session using a Starlark script effect.

```yaml
version: 1
name: track-file-modifications
metadata:
  description: Track modified files and warn about high churn
  labels:
    category: tracking
    severity: low

config:
  priority: low

match:
  events: [post_tool_call]
  tools: [Write, Edit]

variables:
  - name: file_path
    expression: 'tool_input.file_path ?? ""'

rules:
  - name: track-modifications
    effects:
      - type: script
        language: starlark
        source: |
          files = session_get("modified_files") or []
          path = vars["file_path"]

          if path not in files:
              files.append(path)
              session_set("modified_files", files)

          # Warn if too many files modified
          if len(files) > 20:
              log("warning", "You've modified " + str(len(files)) + " files this session. Consider committing your changes.")
        when: always
```

## Integration policies

Policies that integrate with external services.

### Slack notification on sensitive file access

This policy logs when sensitive files are accessed. For actual Slack notifications, use a native extension with an HTTP client or configure a webhook integration.

```yaml
version: 1
name: slack-notify-sensitive-access
metadata:
  description: Log access to sensitive files for notification
  labels:
    category: security
    severity: low

config:
  priority: low

match:
  events: [post_tool_call]
  tools: [Read]

variables:
  - name: file_path
    expression: 'tool_input.file_path ?? ""'
  - name: is_sensitive
    expression: 'file_path.matches("\\.(env|pem|key)$")'

conditions:
  - name: is-sensitive-file
    expression: 'is_sensitive'

rules:
  - name: log-sensitive-access
    effects:
      - type: log
        level: warning
        message: "Sensitive file accessed: {{ file_path }}"
        when: always
      - type: setState
        scope: session
        key: sensitive_files_accessed
        value: '{{ ($session_get("sensitive_files_accessed", []) + [file_path]) | unique }}'
        when: always
```

For actual Slack notifications, use a native extension that provides a `slack:send_message` effect type:

```yaml
version: 1
name: slack-notify-sensitive-access-native
metadata:
  description: Send Slack notification when sensitive files are accessed
  labels:
    category: security
    severity: low

config:
  priority: low

match:
  events: [post_tool_call]
  tools: [Read]

variables:
  - name: file_path
    expression: 'tool_input.file_path ?? ""'
  - name: is_sensitive
    expression: 'file_path.matches("\\.(env|pem|key)$")'

conditions:
  - name: is-sensitive-file
    expression: 'is_sensitive'

rules:
  - name: notify-slack
    effects:
      # Requires arci-slack native extension
      - type: slack:send_message
        channel: "#security-alerts"
        message: "Sensitive file accessed: {{ file_path }}"
        when: always
```

### Jira ticket validation

This policy checks that commit messages reference valid Jira tickets using a Starlark macro.

```yaml
version: 1
name: validate-jira-ticket
metadata:
  description: Ensure commit messages reference Jira tickets
  labels:
    category: workflow
    severity: medium

config:
  priority: medium

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: is_commit
    expression: 'command.matches("^git\\s+commit")'
  - name: current_branch
    expression: '$current_branch()'

conditions:
  - name: is-git-commit
    expression: 'is_commit'

macros:
  - name: has_jira_ticket
    language: starlark
    source: |
      def has_jira_ticket(text):
          # Pattern: PROJECT-123
          import re
          pattern = r"[A-Z]+-\d+"
          return bool(re.search(pattern, text))

rules:
  - name: require-ticket-reference
    validate:
      expression: '$has_jira_ticket(command) || $has_jira_ticket(current_branch)'
      message: "Commit message should reference a Jira ticket (e.g., PROJ-123)"
      action: warn
```

### PagerDuty on-call checking

This policy blocks production changes when the user isn't on-call. It uses a native extension macro that can make HTTP calls to PagerDuty.

```yaml
version: 1
name: require-oncall-for-prod
metadata:
  description: Block production deployments unless user is on-call
  labels:
    category: safety
    severity: critical

config:
  priority: critical

match:
  events: [pre_tool_call]
  tools: [Bash]

parameters:
  - name: userEmail
    from:
      env: USER_EMAIL
    default: ""

variables:
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: is_prod_deploy
    expression: 'command.matches("deploy.*production")'

conditions:
  - name: is-production-deployment
    expression: 'is_prod_deploy'

rules:
  - name: check-oncall-status
    # $pagerduty.is_oncall() requires native extension with HTTP capability
    validate:
      expression: '$pagerduty.is_oncall(params.userEmail)'
      message: "Production deployments require being on-call. Check PagerDuty for current on-call schedule."
      action: deny
```

For environments without the PagerDuty extension, use a simpler approach with session-based acknowledgment:

```yaml
version: 1
name: require-oncall-acknowledgment
metadata:
  description: Require acknowledgment before production deployments
  labels:
    category: safety
    severity: critical

config:
  priority: critical

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: is_prod_deploy
    expression: 'command.matches("deploy.*production")'
  - name: acknowledged
    expression: '$session_get("oncall_acknowledged", false)'

conditions:
  - name: is-production-deployment
    expression: 'is_prod_deploy'

rules:
  - name: require-acknowledgment
    validate:
      expression: 'acknowledged'
      message: |
        Production deployments require on-call acknowledgment.
        Set session state 'oncall_acknowledged' to true to proceed.
      action: deny
```

## Code pattern detection with GritQL

GritQL enables powerful structural code matching beyond regex patterns. Unlike regex which operates on text, GritQL understands syntax trees and can match code patterns regardless of formatting or whitespace. arci integrates GritQL through CEL functions that can be used in conditions and validate expressions.

### Detect unsafe eval usage

```yaml
version: 1
name: block-eval-usage
metadata:
  description: Block use of eval() with non-literal arguments
  labels:
    category: security
    severity: high

config:
  priority: high

match:
  events: [pre_tool_call]
  tools: [Write]
  paths:
    include: ["**/*.js", "**/*.ts"]

variables:
  - name: content
    expression: 'tool_input.content ?? ""'

rules:
  - name: block-eval
    conditions:
      - expression: 'content.contains("eval")'  # fast pre-filter
    validate:
      expression: '!$gritql_matching(content, "`eval($expr)`")'
      message: |
        Use of eval() is not allowed.

        eval() executes arbitrary code and poses security risks.
        Consider using safer alternatives like JSON.parse() for data
        or Function constructor for controlled dynamic code.
      action: deny
```

### Require error handling in async functions

```yaml
version: 1
name: require-async-error-handling
metadata:
  description: Warn about async functions without error handling
  labels:
    category: quality
    severity: medium

config:
  priority: medium

match:
  events: [pre_tool_call]
  tools: [Write, Edit]
  paths:
    include: ["**/*.js", "**/*.ts", "**/*.jsx", "**/*.tsx"]

variables:
  - name: content
    expression: 'tool_input.content ?? ""'
  - name: has_await
    expression: '$gritql_matching(content, "`await $expr`")'
  - name: has_try_catch
    expression: '$gritql_matching(content, "`try { $...body } catch ($e) { $...handler }`")'

rules:
  - name: check-error-handling
    conditions:
      - expression: 'has_await'
    validate:
      expression: 'has_try_catch'
      message: |
        Async code should include error handling.
        Found await expressions without surrounding try/catch.
      action: warn
```

### Detect SQL injection vulnerabilities

```yaml
version: 1
name: detect-sql-injection
metadata:
  description: Block SQL queries using string concatenation
  labels:
    category: security
    severity: critical

config:
  priority: critical

match:
  events: [pre_tool_call]
  tools: [Write]
  paths:
    include: ["**/*.py"]

variables:
  - name: content
    expression: 'tool_input.content ?? ""'
  - name: has_execute
    expression: 'content.contains("execute")'

rules:
  - name: block-sql-concat
    conditions:
      - expression: 'has_execute'  # fast pre-filter
    validate:
      expression: |
        !$gritql_matching(content, "`cursor.execute($query)` where { $query <: `$left + $right` }", "python")
      message: |
        Blocked SQL query using string concatenation.

        String concatenation in SQL queries enables SQL injection.
        Use parameterized queries instead:
          cursor.execute("SELECT * FROM users WHERE id = ?", (user_id,))
      action: deny

  - name: block-fstring-sql
    conditions:
      - expression: 'has_execute'
    validate:
      expression: '!$gritql_matching(content, "`cursor.execute(f\"$_\")`", "python")'
      message: |
        Blocked SQL query using f-string formatting.

        F-strings in SQL queries enable SQL injection.
        Use parameterized queries instead.
      action: deny
```

For comprehensive GritQL documentation, pattern syntax, and more examples, see the [GritQL design document](gritql.md).

## Per-project configuration

Examples of project-specific policy sets.

### Monorepo with multiple conventions

In a monorepo, you might want different policies for different subdirectories. Use path matching to scope policies appropriately:

```yaml
version: 1
name: frontend-conventions
metadata:
  description: Inject React best practices for frontend package
  labels:
    category: conventions
    package: frontend

config:
  priority: medium

match:
  events: [session_start]
  paths:
    include: ["packages/frontend/**"]

rules:
  - name: inject-frontend-context
    mutate:
      expression: |
        object.with({
          context: (object.context ?? "") + "\n\n" +
            "This is the frontend package. Use React best practices:\n" +
            "- Prefer functional components with hooks\n" +
            "- Use TypeScript for type safety\n" +
            "- Follow existing component patterns"
        })
```

```yaml
version: 1
name: backend-conventions
metadata:
  description: Inject Django conventions for backend package
  labels:
    category: conventions
    package: backend

config:
  priority: medium

match:
  events: [session_start]
  paths:
    include: ["packages/backend/**"]

rules:
  - name: inject-backend-context
    mutate:
      expression: |
        object.with({
          context: (object.context ?? "") + "\n\n" +
            "This is the backend package. Follow Django conventions:\n" +
            "- Use Django REST framework for APIs\n" +
            "- Follow Django model patterns\n" +
            "- Write tests for all new views"
        })
```

### Open source project with contributor guidelines

```yaml
version: 1
name: oss-contributor-context
metadata:
  description: Inject open source contribution guidelines
  labels:
    category: conventions
    type: open-source

config:
  priority: medium

match:
  events: [session_start]

conditions:
  - name: has-contributing
    expression: '$file_exists("CONTRIBUTING.md")'

rules:
  - name: inject-oss-context
    mutate:
      expression: |
        object.with({
          context: (object.context ?? "") + "\n\n" +
            "This is an open source project. Please:\n" +
            "- Follow the contribution guidelines in CONTRIBUTING.md\n" +
            "- Ensure all tests pass before submitting\n" +
            "- Sign commits with DCO (use -s flag)\n" +
            "- Keep PRs focused and well-documented"
        })
```

### Internal tool with compliance requirements

For compliance audit logging, use a script effect that logs tool usage:

```yaml
version: 1
name: compliance-audit-logging
metadata:
  description: Log all tool usage for compliance auditing
  labels:
    category: compliance
    severity: low

config:
  priority: low

match:
  events: [pre_tool_call, post_tool_call]

rules:
  - name: log-tool-usage
    effects:
      - type: script
        language: starlark
        source: |
          # Build audit log entry
          log_entry = {
              "timestamp": timestamp(),
              "event": event_type,
              "tool": tool_name,
              "session": session_id or "unknown",
              "cwd": cwd,
          }

          # Store in project state for later retrieval
          audit_log = project_get("audit_log") or []
          audit_log.append(log_entry)

          # Keep last 1000 entries
          if len(audit_log) > 1000:
              audit_log = audit_log[-1000:]

          project_set("audit_log", audit_log)

          log("info", "Audit: " + event_type + " " + tool_name)
        when: always
```

## Extension examples

Examples of custom extensions that package and distribute policies.

### Simple policies-only extension

A policies-only extension contains YAML policy files with no custom code. This is the simplest and most auditable extension type.

```
my-company-rules/
├── extension.toml
└── policies/
    └── company-standards.yaml
```

```toml
# my-company-rules/extension.toml
[extension]
name = "my-company-rules"
version = "1.0.0"
description = "Standard policies for My Company projects"
type = "policies"

[extension.arci]
min_version = "0.1.0"
```

```yaml
# my-company-rules/policies/company-standards.yaml
version: 1
name: my-company-rules:header-check
metadata:
  description: Ensure files include company copyright header
  labels:
    category: standards
    owner: engineering

config:
  priority: medium

match:
  events: [post_tool_call]
  tools: [Write]
  paths:
    include: ["**/*.py", "**/*.js", "**/*.ts"]
    exclude: ["**/node_modules/**", "**/.git/**"]

variables:
  - name: content
    expression: 'tool_input.content ?? ""'
  - name: has_copyright
    expression: 'content.contains("Copyright My Company")'

rules:
  - name: check-copyright-header
    validate:
      expression: 'has_copyright'
      message: "File should include company copyright header"
      action: warn
```

### Extension with custom Starlark macros

Full extensions can provide custom expression macros via Starlark scripts. These macros run in a sandboxed environment and become available in policy expressions using a namespaced syntax.

```
my-extension/
├── extension.toml
├── policies/
│   └── business-hours.yaml
└── scripts/
    └── macros.star
```

```toml
# my-extension/extension.toml
[extension]
name = "my-extension"
version = "1.0.0"
description = "Custom policies with helper macros"
type = "full"

[extension.arci]
min_version = "0.1.0"

[extension.scripts]
macros = ["scripts/macros.star"]
```

```python
# my-extension/scripts/macros.star

def is_business_hours():
    """Check if current time is within business hours (9-5 M-F)."""
    now = timestamp_now()
    weekday = now.weekday()  # 0 = Monday, 6 = Sunday
    hour = now.hour()
    return weekday < 5 and hour >= 9 and hour < 17

def team_owns_path(path, team):
    """Check if a team owns a given path based on CODEOWNERS patterns."""
    codeowners = read_file(".github/CODEOWNERS")
    if codeowners == None:
        return False

    for line in codeowners.split("\n"):
        if line.startswith("#") or line.strip() == "":
            continue
        parts = line.split()
        if len(parts) >= 2:
            pattern = parts[0]
            owners = parts[1:]
            if matches_glob(path, pattern) and ("@" + team) in owners:
                return True
    return False
```

Custom macros are namespaced by extension name in expressions. The macros above become `$my_extension.is_business_hours()` and `$my_extension.team_owns_path()`:

```yaml
# my-extension/policies/business-hours.yaml
version: 1
name: my-extension:after-hours-warning
metadata:
  description: Remind users to take breaks when working outside business hours

config:
  priority: low

match:
  events: [session_start]

conditions:
  - name: outside-business-hours
    expression: '!$my_extension.is_business_hours()'

rules:
  - name: after-hours-notice
    effects:
      - type: notify
        title: "After hours"
        message: "Working outside business hours. Remember to take breaks!"
        when: always
```

### Extension with native effect handlers

Custom effect handlers require native extensions because they need to perform I/O operations like HTTP requests. Native extensions are Go plugins or gRPC services that implement the Extension interface.

```
linear-extension/
├── extension.toml
├── go.mod
├── main.go
└── policies/
    └── linear-defaults.yaml
```

```toml
# linear-extension/extension.toml
[extension]
name = "arci-linear"
version = "1.0.0"
description = "Linear integration for arci"
type = "native"

[extension.arci]
min_version = "0.1.0"

[extension.native]
binary = "arci-ext-linear"
```

```go
// linear-extension/main.go
package main

import (
 "bytes"
 "encoding/json"
 "fmt"
 "net/http"

 "github.com/tbhb/arci/pkg/extension"
)

type LinearIssueHandler struct{}

func (h *LinearIssueHandler) EffectType() string {
 return "linear:create_issue"
}

func (h *LinearIssueHandler) Execute(params extension.EffectParams, ctx *extension.Context) error {
 query := `mutation CreateIssue($title: String!, $teamId: String!) {
  issueCreate(input: { title: $title, teamId: $teamId }) {
   issue { id identifier url }
  }
 }`

 body, _ := json.Marshal(map[string]any{
  "query": query,
  "variables": map[string]string{
   "title":  params.GetString("title"),
   "teamId": params.GetString("team_id"),
  },
 })

 req, _ := http.NewRequest("POST", "https://api.linear.app/graphql", bytes.NewReader(body))
 req.Header.Set("Authorization", params.GetString("api_key"))
 req.Header.Set("Content-Type", "application/json")

 resp, err := http.DefaultClient.Do(req)
 if err != nil {
  return fmt.Errorf("linear API request failed: %w", err)
 }
 defer resp.Body.Close()

 return nil
}

type LinearExtension struct{}

func (e *LinearExtension) Name() string { return "linear" }
func (e *LinearExtension) EffectHandlers() []extension.EffectHandler {
 return []extension.EffectHandler{&LinearIssueHandler{}}
}

var Extension LinearExtension
```

Native extensions require explicit trust when installed. Once trusted, the custom effect type becomes available in policies:

```yaml
# Using the Linear extension
version: 1
name: create-tracking-issue
metadata:
  description: Create Linear tracking issues for AI sessions

config:
  priority: low

match:
  events: [session_start]

parameters:
  - name: createIssues
    from:
      env: CREATE_TRACKING_ISSUES
    default: ""
  - name: linearApiKey
    from:
      env: LINEAR_API_KEY
    default: ""
  - name: linearTeamId
    from:
      env: LINEAR_TEAM_ID
    default: ""

conditions:
  - name: issues-enabled
    expression: 'params.createIssues != ""'

rules:
  - name: create-linear-issue
    effects:
      # Requires arci-linear native extension
      - type: linear:create_issue
        api_key: '{{ params.linearApiKey }}'
        title: 'AI Session: {{ session_id }}'
        team_id: '{{ params.linearTeamId }}'
        when: always
```

For comprehensive extension documentation, see the [Extensions design document](../extensions.md).

## Claude Code-specific policies

Policies that use Claude Code-specific features like permission_request events.

### Auto-approve read-only operations

```yaml
version: 1
name: auto-approve-read-only
metadata:
  description: Auto-approve read-only tool operations
  labels:
    category: permissions

config:
  priority: high

match:
  events: [permission_request]
  tools: [Read, Glob, Grep]

rules:
  - name: allow-reads
    mutate:
      expression: |
        object.with({
          action: "allow"
        })
```

## Anti-patterns

Examples of what not to do and why.

### Overly broad conditions

```yaml
# BAD: Matches too many commands
version: 1
name: too-broad
match:
  tools: [Bash]
rules:
  - name: block-rm
    validate:
      expression: '!tool_input.command.contains("rm")'
      message: "No rm allowed"
      action: deny
```

This blocks legitimate commands like `npm run build` or `karma test`. Be specific:

```yaml
# GOOD: Matches only dangerous rm patterns
version: 1
name: specific-rm-block
match:
  events: [pre_tool_call]
  tools: [Bash]
variables:
  - name: command
    expression: 'tool_input.command ?? ""'
rules:
  - name: block-dangerous-rm
    validate:
      expression: '!command.matches("^rm\\s+(-[rRf]+\\s+)*/(?!tmp)")'
      message: "Recursive deletion of root paths is blocked"
      action: deny
```

### Priority misuse

```yaml
# BAD: Logging at critical priority
version: 1
name: log-everything
config:
  priority: critical  # Wrong! This isn't security-critical
match:
  events: [pre_tool_call]
rules:
  - name: log
    effects:
      - type: log
        message: "Tool used"
```

Reserve critical priority for security rules. Use low priority for logging:

```yaml
# GOOD: Appropriate priority for logging
version: 1
name: log-everything
config:
  priority: low
match:
  events: [pre_tool_call]
rules:
  - name: log
    effects:
      - type: log
        level: info
        message: "Tool used: {{ tool_name }}"
```

### State store abuse

```yaml
# BAD: Using state for things that don't need persistence
version: 1
name: unnecessary-state
match:
  tools: [Bash]
rules:
  - name: track-last-tool
    effects:
      - type: setState
        scope: session
        key: last_tool
        value: "Bash"
```

Only use state when you need to track things across policy evaluations for rate limiting, escalation, or context accumulation. Simple logging doesn't need state.

### Inline Starlark overuse

```yaml
# BAD: Using Starlark script where declarative expressions suffice
version: 1
name: starlark-overkill
match:
  tools: [Bash]
macros:
  - name: is_rm_rf
    language: starlark
    source: |
      def is_rm_rf(command):
          return "rm -rf" in command
rules:
  - name: check-rm
    validate:
      expression: '!$is_rm_rf(tool_input.command)'
      message: "blocked"
      action: deny
```

Use CEL expressions directly when they're sufficient:

```yaml
# GOOD: Declarative CEL expression is clearer and faster
version: 1
name: declarative-check
match:
  tools: [Bash]
variables:
  - name: command
    expression: 'tool_input.command ?? ""'
rules:
  - name: check-rm
    validate:
      expression: '!command.contains("rm -rf")'
      message: "Recursive deletion blocked"
      action: deny
```

Use Starlark macros when logic genuinely benefits from loops, complex string manipulation, or multi-step computation:

```yaml
# BETTER: Starlark for genuinely complex logic
version: 1
name: complex-validation
match:
  tools: [Bash]
variables:
  - name: command
    expression: 'tool_input.command ?? ""'
macros:
  - name: analyze_rm_command
    language: starlark
    source: |
      def analyze_rm_command(cmd):
          if not cmd.startswith("rm "):
              return {"dangerous": False}

          parts = cmd.split()
          has_recursive = any("-r" in p or "-R" in p for p in parts)
          targets_root = any(
              p.startswith("/") and not p.startswith("/tmp")
              for p in parts if not p.startswith("-")
          )

          return {
              "dangerous": has_recursive and targets_root,
              "reason": "recursive deletion of root paths"
          }
rules:
  - name: check-dangerous-rm
    variables:
      - name: analysis
        expression: '$analyze_rm_command(command)'
    validate:
      expression: '!analysis.dangerous'
      message: "Blocked: {{ analysis.reason }}"
      action: deny
```

### Missing structural matching

```yaml
# BAD: Using only conditions without structural matching
version: 1
name: no-structural-match
match:
  events: [pre_tool_call]
  # Missing tools: field means this evaluates for ALL tools
conditions:
  - expression: 'tool_name == "Bash" && command.contains("rm")'
rules:
  - name: check-rm
    validate:
      # ...
```

Use structural matching via the `match` block for efficient filtering:

```yaml
# GOOD: Structural matching filters policies efficiently
version: 1
name: with-structural-match
match:
  events: [pre_tool_call]
  tools: [Bash]  # Only evaluate for Bash tool
variables:
  - name: command
    expression: 'tool_input.command ?? ""'
conditions:
  - expression: 'command.contains("rm")'  # Further filter with CEL
rules:
  - name: check-rm
    validate:
      # ...
```

Structural matching using `match.tools`, `match.events`, and `match.paths` is indexed for fast lookup. CEL conditions in `conditions` require expression evaluation. Put as much filtering as possible in `match` for best performance.

---

This document will grow as more patterns emerge from real-world usage. Contributions of useful recipes are welcome.
