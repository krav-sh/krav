# Expression language

Krav uses two evaluation engines. CEL (Common Expression Language) provides type-safe evaluation for boolean conditions and data transformations. Go's `text/template` with the Sprig function library handles string interpolation and control flow. This separation matches each engine to its strength. CEL excels at evaluating expressions against structured data, while `text/template` excels at producing string output with interpolation.

## Where expressions appear

CEL expressions appear throughout the policy structure:

Policy-level `conditions` arrays contain expressions that determine whether the entire policy applies to a given event. The engine evaluates these after structural matching passes.

Rule-level `conditions` arrays contain expressions that determine whether a specific rule within the policy applies. The engine evaluates these after the rule's structural match passes.

Rule `validate.expression` fields contain the validation logic that must return true for the rule to pass. If the expression returns false, the validation fails with the specified action (deny, warn, or audit).

Rule `mutate.expression` fields contain transformation logic that receives the current event state as `object` and returns the modified state.

Variable `expression` fields compute values from parameters, the hook event, built-in functions, or other variables. The engine evaluates variables in declaration order.

Macro `expression` fields define reusable expression fragments that other CEL expressions call using the `$` prefix.

Go templates appear in action message fields like `validate.message` and effect templates like `effects[].value` where string interpolation requires runtime values.

## Evaluation modes

Krav distinguishes between two evaluation modes based on field semantics in policy definitions.

Condition and expression fields contain CEL expressions evaluated for truthiness or transformation. These expressions appear without any delimiters. The `conditions`, `validate.expression`, and `mutate.expression` fields use this mode.

Template fields use Go's `text/template` syntax for interpolation and control flow. The `{{ }}` syntax inserts values and provides control structures, and `{{/* */}}` allows comments. The Sprig function library extends the built-in template functions with string manipulation, math, and other utilities. Fields like `validate.message` and effect values use this mode.

```yaml
version: 1
name: block-dangerous-rm

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'

rules:
  - name: block-rm-rf-root
    # CEL expression for validation
    validate:
      expression: '!command.matches("^rm\\s+(-rf|-fr)\\s+/")'
      # Go template for message
      message: "Blocked dangerous command: {{ .ToolInput.Command }}"
      action: deny
```

This separation keeps conditions readable with a concise expression syntax while providing full template capabilities for constructing output messages.

## Condition expressions

Condition expressions are CEL expressions evaluated against the hook context. Policy-level conditions determine whether a policy applies. Rule-level conditions determine whether a rule applies. Validation expressions determine whether a rule passes.

### Comparison operators

Standard comparison operators work as expected. Equality uses `==` and `!=`. Ordering uses `<`, `<=`, `>`, and `>=`. CEL is strongly typed: numbers compare numerically, strings compare lexicographically, and cross-type comparisons produce type errors at check time rather than silently returning false.

```yaml
variables:
  - name: timeout
    expression: 'tool_input.timeout ?? 0'
  - name: is_bash
    expression: 'tool_name == "Bash"'
  - name: is_write
    expression: 'tool_name == "Write"'
```

### Boolean operators

The boolean operators `&&`, `||`, and `!` combine conditions. Expressions resolve left to right with standard C precedence: `!` binds tightest, then `&&`, then `||`. Parentheses override precedence when needed.

```yaml
conditions:
  - expression: 'tool_name == "Bash" && tool_input.command.contains("rm")'
  - expression: 'tool_name == "Write" || tool_name == "Edit"'
  - expression: '!(tool_input.command.startsWith("git"))'
  - expression: '(priority == "high" || priority == "critical") && enabled'
```

### Membership testing

The `in` operator tests whether a value appears in a list. CEL supports this natively.

```yaml
conditions:
  - expression: 'tool_name in ["Read", "Write", "Edit"]'
  - expression: 'tool_input.command.contains("force")'
```

### Path navigation

Dot notation navigates nested objects. Bracket notation accesses list elements by index. You can combine these to traverse complex structures.

```yaml
variables:
  - name: has_command
    expression: 'has(tool_input.command)'
  - name: first_edit_has_todo
    expression: 'tool_input.edits[0].old_string.contains("TODO")'
  - name: high_temp
    expression: 'raw.llm_request.config.temperature > 0.9'
```

### Safe navigation

CEL uses the `has()` macro to check whether a field exists before accessing it. Accessing a field that does not exist produces an error in CEL, so explicit existence checks are necessary for optional fields.

```yaml
# Check if timeout exists before comparing
conditions:
  - expression: 'has(tool_input.timeout) && tool_input.timeout > 30000'

# Check if a field is present
conditions:
  - expression: 'has(tool_input.timeout)'

# Combine existence check with value check
conditions:
  - expression: 'has(tool_input.command) && tool_input.command.startsWith("git")'
```

The `has()` macro is the idiomatic way to handle optional fields in CEL. Unlike template languages where missing properties might return a zero value or silently pass, CEL requires explicit checks for field existence.

The null-coalescing operator `??` provides defaults for optional fields:

```yaml
variables:
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: timeout
    expression: 'tool_input.timeout ?? 5000'
```

### String methods

CEL provides built-in string methods for common operations. You call these as methods on string values, and they return boolean results suitable for conditions.

The `startsWith(prefix)` method returns true if the string starts with the given prefix. The `endsWith(suffix)` method returns true if the string ends with the given suffix. The `contains(substring)` method returns true if the string contains the given substring. The `matches(regex)` method returns true if the string matches the regular expression. All these methods are case-sensitive.

```yaml
conditions:
  - expression: 'tool_input.command.startsWith("git push")'
  - expression: 'tool_input.file_path.endsWith(".py")'
  - expression: 'tool_input.command.contains("--force")'
  - expression: 'tool_input.command.matches("^git\\s+push")'
```

For case-insensitive comparison, use a custom CEL function:

```yaml
conditions:
  - expression: 'toLower(tool_input.command).contains("delete")'
```

## The jsonPointer function

While dot notation works well for typed structures, some hook inputs contain untyped JSON where the structure is not known at compile time. The `jsonPointer` function provides access to arbitrary paths using RFC 6901 JSON Pointer syntax.

JSON Pointer uses forward slashes to separate path components. Array indices are numeric. The path must be a string literal or variable containing the pointer.

```yaml
# Access a deeply nested field
conditions:
  - expression: 'jsonPointer(tool_input, "/deeply/nested/field") == "expected"'

# Access an array element
conditions:
  - expression: 'has(jsonPointer(tool_input, "/items/0/name"))'

# Access a field with special characters (escaped per RFC 6901)
conditions:
  - expression: 'has(jsonPointer(tool_input, "/config/some~1path"))'
```

The function returns a null value for missing paths rather than raising errors. This provides safe navigation behavior for untyped data.

The `jsonPointer` function is primarily useful when working with MCP tool inputs or assistant-specific raw data where the schema varies by tool or context.

## Built-in CEL functions and Go template functions

Krav extends both CEL and Go templates with domain-specific functions for common hook operations.

### CEL condition functions

CEL provides these built-in methods on strings: `startsWith()`, `endsWith()`, `contains()`, `matches()`, and `size()`. The `has()` macro checks field existence.

Krav adds custom CEL functions for domain-specific operations. You invoke these with a `$` prefix to distinguish them from CEL builtins.

Path functions include `$file_exists(path)` which returns true if the path exists as a regular file, and `$matches_glob(path, pattern)` which checks whether a path matches a glob pattern.

```yaml
conditions:
  - expression: '$file_exists(tool_input.file_path)'
  - expression: '$matches_glob(tool_input.file_path, "**/*.py")'
  - expression: '$matches_glob(tool_input.file_path, "src/**/*.go")'
```

The `$env(name)` function retrieves environment variable values. It returns an empty string if the variable is not set. The `$env(name, default)` overload returns the default value if the variable is not set.

```yaml
conditions:
  - expression: '$env("CI") != ""'
  - expression: '$env("LOG_LEVEL", "info") == "debug"'
```

Git context functions include `$current_branch()` which returns the name of the current branch (or empty string if in detached HEAD state), and `$git_is_dirty()` which indicates whether the working tree has uncommitted changes. The `$is_staged(path)` function returns true if the specified file has staged changes.

```yaml
conditions:
  - expression: '$current_branch() == "main"'
  - expression: '$current_branch().startsWith("feature/")'
  - expression: '$git_is_dirty()'
  - expression: '!$git_is_dirty()'
  - expression: '$is_staged(tool_input.file_path)'
```

State access functions include `$session_get(key, default)` for session-scoped values and `$project_get(key, default)` for project-scoped values. Both return the default if the key does not exist.

```yaml
conditions:
  - expression: '$session_get("warning_count", 0) > 3'
  - expression: '$session_get("acknowledged_risks", null) != null'
  - expression: '$project_get("last_deploy_time", null) != null'
  - expression: '$project_get("deployment_count", 0) > 10'
```

### Go template functions

Go templates have access to Sprig functions and custom Krav functions. Sprig provides `lower`, `upper`, `trim`, `replace`, `split`, `join`, `contains`, `hasPrefix`, `hasSuffix`, and many more. Custom Krav template functions include `env`, `currentBranch`, `gitIsDirty`, `sessionGet`, `projectGet`, and `regexReplace`.

```yaml
message: "Running in {{ env \"NODE_ENV\" | default \"development\" }} mode"
message: "Command: {{ .ToolInput.Command | replace \"password\" \"***\" }}"
message: "Branch: {{ currentBranch }}"
```

## Template syntax

Template fields support Go's `text/template` syntax for constructing interpolated content. Action messages, effect values, and other string outputs use this mode.

### Variable interpolation

Double braces insert values from the template context. Go templates use dot notation with exported (capitalized) field names. You can chain Sprig functions using the pipe operator.

```yaml
rules:
  - name: warn-on-command
    validate:
      expression: 'true'  # always passes, just for the message
      message: "About to execute: {{ .ToolInput.Command }}"
      action: audit

  - name: inject-context
    mutate:
      expression: |
        object.with({
          context: object.context + "\nCurrent branch is " + $current_branch()
        })
```

Missing variables resolve to zero values rather than raising errors. Krav configures templates with `missingkey=zero` for fail-safe behavior, preventing template errors from blocking evaluation.

### Conditionals

The `{{ if }}` block provides conditional content. It supports `{{ else if }}` and `{{ else }}` clauses.

```yaml
effects:
  - type: notify
    message: |
      {{ if .GitIsDirty }}
      Warning: You have uncommitted changes.
      {{ else }}
      Working tree is clean.
      {{ end }}
```

### Loops

The `{{ range }}` block iterates over sequences. The current element is available as `.` inside the block, or you can assign it to variables.

```yaml
rules:
  - name: list-edits
    validate:
      expression: 'size(tool_input.edits) <= 10'
      message: |
        Files to be modified:
        {{ range .ToolInput.Edits }}
        - {{ .FilePath }}: {{ .OldString | trunc 30 }} -> {{ .NewString | trunc 30 }}
        {{ end }}
      action: warn
```

### Comments

The `{{/* */}}` syntax creates comments that do not appear in output. Comments can span more than one line.

```yaml
effects:
  - type: notify
    message: |
      {{/* This comment won't appear in the notification */}}
      Remember to run tests before committing.
```

### Whitespace control

By default, template blocks preserve surrounding whitespace. The `{{-` and `-}}` variants strip whitespace before or after the block.

```yaml
effects:
  - type: notify
    message: |
      {{- if eq .ToolName "Bash" -}}
      Shell command detected.
      {{- end -}}
```

The `-` on the opening tag strips preceding whitespace; the `-` on the closing tag strips following whitespace. This helps produce compact output from templates with heavy structural whitespace.

## Expression context

Policies have access to a set of context variables derived from the hook event. CEL expressions access variables directly as top-level snake_case names like `tool_name` and `tool_input.command`. Go templates use dot notation with exported field names like `.ToolName` and `.ToolInput.Command`.

### Common context variables

The `event_type` variable (CEL) or `.EventType` (template) contains the canonical event type: `pre_tool_call`, `post_tool_call`, `pre_prompt`, `post_response`, `session_start`, `session_end`, or other event types defined in the hook schema.

The `session_id` variable (CEL) or `.SessionID` (template) contains the AI assistant's session identifier. This value may be empty for assistants that do not provide session IDs or for specific events where the ID remains unavailable.

The `cwd` variable (CEL) or `.Cwd` (template) contains the current working directory as an absolute path.

The `timestamp` variable (CEL) or `.Timestamp` (template) contains the event timestamp.

### Tool event context

For `pre_tool_call` and `post_tool_call` events, extra context is available.

The `tool_name` variable (CEL) or `.ToolName` (template) contains the canonical tool name: `Bash`, `Write`, `Read`, `Edit`, `Glob`, `Grep`, `Task`, `WebSearch`, `WebFetch`, or `mcp:server:tool` for MCP tools.

The `tool_input` variable (CEL) or `.ToolInput` (template) contains the tool's input parameters as an object. The structure varies by tool type. For shell commands, it includes `command` (CEL: `tool_input.command`, template: `.ToolInput.Command`), optional `timeout`, and optional `cwd`. For file operations, it includes `file_path` and relevant content or edit information.

For `post_tool_call` events, the `tool_output` variable (CEL) or `.ToolOutput` (template) contains the tool's result.

### Git context

When the current working directory is within a git repository, Krav populates git-related context.

The `git_branch` variable (CEL) or `.GitBranch` (template) contains the current branch name, or empty string in detached HEAD state.

The `git_is_dirty` variable (CEL) or `.GitIsDirty` (template) is true if the working tree contains uncommitted changes.

The `git_head_commit` variable (CEL) or `.GitHeadCommit` (template) contains the current HEAD commit hash.

The `git_is_detached` variable (CEL) or `.GitIsDetached` (template) is true if HEAD does not point to any branch.

### Raw assistant data

The `raw` variable (CEL) or `.Raw` (template) contains the unmodified hook input from the assistant, including assistant-specific fields not present in the normalized schema.

```yaml
# Access Cursor's workspace roots (CEL)
conditions:
  - expression: '"sensitive-project" in raw.workspace_roots'

# Access the assistant's native tool name (CEL)
conditions:
  - expression: 'raw.tool_name == "Bash"'
```

## Examples

The following examples show common patterns using the policy structure with conditions and templates.

### Blocking dangerous commands

```yaml
version: 1
name: block-dangerous-rm

config:
  priority: critical

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'

rules:
  - name: block-rm-rf-root
    validate:
      expression: '!command.matches("^rm\\s+(-rf|-fr)\\s+/")'
      message: "Blocked dangerous rm command: {{ .ToolInput.Command }}"
      action: deny
```

### Modifying git push commands

```yaml
version: 1
name: safe-force-push

config:
  priority: high

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: is_force_push
    expression: 'command.matches("git push.*--force") && !command.contains("--force-with-lease")'

rules:
  - name: replace-force-with-lease
    conditions:
      - expression: 'is_force_push'
    mutate:
      expression: |
        object.with({
          tool_input: object.tool_input.with({
            command: object.tool_input.command.replace("--force", "--force-with-lease")
          })
        })
    effects:
      - type: notify
        message: "Replaced --force with --force-with-lease for safety"
        when: always
```

### Rate limiting shell commands

```yaml
version: 1
name: shell-rate-limiting

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: shell_count
    expression: '$session_get("shell_count", 0)'

rules:
  - name: count-shell-commands
    effects:
      - type: setState
        scope: session
        key: shell_count
        value: '{{ add .ShellCount 1 }}'
        when: always

  - name: warn-excessive-shell
    conditions:
      - expression: 'shell_count > 20'
    validate:
      expression: 'true'  # pass but show warning
      message: |
        You have executed {{ sessionGet "shell_count" }} shell commands this session.
        Consider whether these commands could be combined or automated.
      action: warn
```

### Injecting context based on file type

```yaml
version: 1
name: coding-standards-injection

match:
  events: [pre_tool_call]
  tools: [Write, Edit]

variables:
  - name: file_path
    expression: 'tool_input.file_path ?? ""'
  - name: is_python
    expression: 'file_path.endsWith(".py")'
  - name: is_model
    expression: '$matches_glob(file_path, "**/models/*.py")'
  - name: is_api
    expression: '$matches_glob(file_path, "**/api/*.py")'
  - name: is_test
    expression: '$matches_glob(file_path, "**/test_*.py")'

rules:
  - name: python-model-reminder
    conditions:
      - expression: 'is_python && is_model && !is_test'
    mutate:
      expression: |
        object.with({
          context: object.context + "\n\nRemember: model changes require migration scripts and tests."
        })

  - name: python-api-reminder
    conditions:
      - expression: 'is_python && is_api && !is_test'
    mutate:
      expression: |
        object.with({
          context: object.context + "\n\nRemember: API changes should update documentation and tests."
        })

  - name: python-general-reminder
    conditions:
      - expression: 'is_python && !is_model && !is_api && !is_test'
    mutate:
      expression: |
        object.with({
          context: object.context + "\n\nConsider adding tests for this change."
        })
```

### Branch-specific rules

```yaml
version: 1
name: protect-main-branch

config:
  priority: critical

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: current_branch
    expression: '$current_branch()'
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: is_git_write
    expression: 'command.matches("^git\\s+(push|commit|merge)")'

conditions:
  - name: on-main-branch
    expression: 'current_branch == "main"'

rules:
  - name: block-direct-commits
    conditions:
      - expression: 'is_git_write'
    validate:
      expression: 'false'
      message: |
        Direct commits to main are not allowed.
        Current branch: {{ currentBranch }}
        Please create a feature branch and open a pull request.
      action: deny
```

### Accessing MCP tool inputs

```yaml
version: 1
name: validate-github-issues

config:
  priority: high

match:
  events: [pre_tool_call]
  tools: [mcp:github:create_issue]

variables:
  - name: title
    expression: 'tool_input.arguments.title ?? ""'
  - name: title_length
    expression: 'size(title)'

rules:
  - name: require-descriptive-title
    validate:
      expression: 'title_length >= 10'
      message: |
        Issue title "{{ .ToolInput.Arguments.Title }}" is very short.
        Consider using a more descriptive title.
      action: warn
```

### Combined validation and effects

```yaml
version: 1
name: rate-limit-dangerous-commands

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
  - name: attempts
    expression: '$session_get("dangerous_command_attempts", 0)'

macros:
  - name: is_destructive_command
    expression: |
      command.matches("rm\\s+-rf") ||
      command.contains("chmod 777")

conditions:
  - name: is-destructive
    expression: '$is_destructive_command()'

rules:
  - name: rate-limit-dangerous
    validate:
      expression: 'attempts < params.maxAttempts'
      message: "Too many dangerous command attempts ({{ .Attempts }}/{{ .Params.MaxAttempts }})"
      action: deny
    effects:
      - type: setState
        scope: session
        key: dangerous_command_attempts
        value: '{{ add .Attempts 1 }}'
        when: always
```
