# Starlark scripting

arci uses CEL (Common Expression Language) for all policy expressions: conditions, validations, mutations, and variables. For situations requiring more complex logic than CEL provides, arci supports Starlark as an embedded scripting language. Starlark serves two roles: implementing complex macros that CEL expressions can call, and running as script effects for sophisticated side actions.

## Why Starlark

Starlark complements CEL's declarative expressions with imperative scripting capabilities. While CEL excels at succinct boolean checks and simple transformations, some logic benefits from multi-step computation, loops, or accumulating intermediate results. Starlark handles these cases without requiring external shell commands or compromising the security model.

Starlark provides Python-like syntax that most developers already know. There is no new language to learn, no unfamiliar operators or sigils. Anyone who has written a Python function can write a Starlark script. This dramatically lowers the barrier to writing custom hook logic compared to domain-specific languages with novel syntax.

The language is deterministic and hermetic by default. Starlark programs cannot access the filesystem, network, or environment unless the host application explicitly provides those capabilities. There is no `import` statement, no `open()` call, no `os` module. A misconfigured shell script might delete files or exfiltrate data; a Starlark script can only return values or call APIs that arci explicitly exposes. This hermetic property is not bolted on through sandboxing but is fundamental to the language design.

Starlark runs natively in Go via `go.starlark.net` without FFI or external interpreters. There is no Python runtime to initialize, no subprocess to spawn. Scripts execute in the same process as arci with minimal overhead. This matters for hook evaluation latency, where every millisecond of delay is felt by the AI assistant and its user.

The `go.starlark.net` implementation provides resource limits through step counting and memory bounds. Scripts that loop too long or allocate unbounded memory are terminated before they can cause problems. Combined with timeouts, this ensures that a poorly written script cannot hang the hook evaluation pipeline.

## Starlark macros

Macros provide reusable logic that CEL expressions can call. While simple macros are defined directly as CEL expressions, complex macros that require loops, multiple steps, or sophisticated string manipulation can be implemented in Starlark.

### Macro definition

Starlark macros are defined in the policy's `macros` section with `language: starlark`:

```yaml
version: 1
name: security-checks

macros:
  - name: is_destructive_command
    language: starlark
    source: |
      def is_destructive_command(command):
          patterns = [
              "rm -rf",
              "chmod 777",
              "> /dev/sd",
              "mkfs.",
              "dd if=",
          ]
          for pattern in patterns:
              if pattern in command:
                  return True
          return False

  - name: count_sensitive_paths
    language: starlark
    source: |
      def count_sensitive_paths(paths):
          sensitive = [".env", "secret", "credential", "password", ".pem", ".key"]
          count = 0
          for path in paths:
              for s in sensitive:
                  if s in path.lower():
                      count += 1
                      break
          return count
```

### Calling macros from CEL

CEL expressions invoke Starlark macros using the `$` prefix, just like CEL macros:

```yaml
rules:
  - name: block-destructive-commands
    match:
      tools: [Bash]
    validate:
      expression: '!$is_destructive_command(tool_input.command)'
      message: "Destructive command pattern detected"
      action: deny

  - name: warn-sensitive-files
    match:
      tools: [Read, Write, Edit]
    variables:
      - name: sensitive_count
        expression: '$count_sensitive_paths([tool_input.file_path])'
    validate:
      expression: 'sensitive_count == 0'
      message: "Operation involves {{ sensitive_count }} sensitive path(s)"
      action: warn
```

### Macro context

Starlark macros receive their arguments as function parameters. They do not have direct access to the evaluation context (tool_input, params, etc.). All required data must be passed explicitly as arguments.

This design is intentional. Macros are pure functions that transform inputs to outputs. They cannot access or modify external state. This makes macros predictable, testable, and safe to call from any context.

```yaml
# Correct: pass required data as arguments
macros:
  - name: analyze_command
    language: starlark
    source: |
      def analyze_command(command, blocked_patterns):
          for pattern in blocked_patterns:
              if pattern in command:
                  return {"blocked": True, "pattern": pattern}
          return {"blocked": False, "pattern": None}

variables:
  - name: analysis
    expression: '$analyze_command(tool_input.command, params.blockedPatterns)'
```

### Return values

Starlark macros return values that CEL can use in expressions. Supported return types include booleans, strings, integers, floats, lists, and dicts. The return value is automatically converted to CEL-compatible types.

```yaml
macros:
  - name: extract_file_metadata
    language: starlark
    source: |
      def extract_file_metadata(path):
          parts = path.split("/")
          filename = parts[-1] if parts else ""
          ext_parts = filename.rsplit(".", 1)
          return {
              "directory": "/".join(parts[:-1]),
              "filename": filename,
              "extension": ext_parts[1] if len(ext_parts) > 1 else "",
              "depth": len(parts) - 1,
          }

variables:
  - name: metadata
    expression: '$extract_file_metadata(tool_input.file_path)'

rules:
  - name: limit-nesting-depth
    validate:
      expression: 'metadata.depth <= 10'
      message: "File path too deeply nested ({{ metadata.depth }} levels)"
      action: warn
```

## Script effects

For side actions that require complex logic beyond what the built-in effect types provide, policies can define script effects. Unlike macros, script effects have access to state functions and run after the admission decision is determined.

### Script effect syntax

Script effects appear in a rule's `effects` list with `type: script`:

```yaml
version: 1
name: advanced-tracking

variables:
  - name: command
    expression: 'tool_input.command ?? ""'

match:
  tools: [Bash]

rules:
  - name: track-command-patterns
    effects:
      - type: script
        language: starlark
        source: |
          # Track command usage patterns per branch
          branch = current_branch()
          key = "commands_" + branch

          history = session_get(key)
          if history == None:
              history = []

          # Keep last 50 commands
          history = history[-49:] + [command]
          session_set(key, history)

          # Check for repetitive patterns
          if len(history) >= 10:
              recent = history[-10:]
              unique = len(set(recent))
              if unique <= 2:
                  log("warning", "Repetitive command pattern detected: %d unique commands in last 10" % unique)
        timeout_ms: 5000
        when: always
```

### Script effect context

Script effects have access to several context variables and functions. The hook context provides information about the current evaluation:

```python
# The canonical tool name (for tool events)
tool_name       # "Bash", "Write", "Read", etc.

# The tool's input parameters as a dict
tool_input      # {"command": "ls -la", "timeout": 5000}

# The canonical event type
event_type      # "pre_tool_call", "post_tool_call", etc.

# The current session identifier (may be None)
session_id      # "sess_abc123" or None

# The current working directory
cwd             # "/home/user/project"

# Policy-level variables (read-only)
vars            # {"command": "ls -la", "is_blocked": false}

# Resolved parameters (read-only)
params          # {"blockedCommands": ["rm -rf"], "maxAttempts": 3}
```

State functions allow scripts to read and write persistent state:

```python
# Session state (scoped to this conversation)
session_get("key")              # returns value or None if not set
session_set("key", value)       # stores value for this session

# Project state (persists across sessions)
project_get("key")              # returns value or None if not set
project_set("key", value)       # stores value for this project
```

Git context functions provide repository information:

```python
current_branch()                # "main", "feature/foo", etc.
git_is_dirty()                  # True if uncommitted changes exist
is_staged(path)                 # True if file is staged for commit
```

Logging functions allow scripts to emit messages:

```python
log("info", "Processing command: %s" % command)
log("warning", "Approaching rate limit")
log("error", "Failed to parse input")
```

### Script effect constraints

Script effects are fire-and-forget. They cannot influence the admission decision (allow, warn, deny) because they run after that decision is made. They cannot return values that policies use. Their purpose is purely side effects: updating state, logging, sending notifications via state that other systems watch.

Script effects should not perform long-running operations. The timeout_ms setting enforces a maximum execution time. Effects that exceed their timeout are terminated, and evaluation continues normally. The default timeout is 5000 milliseconds.

```yaml
effects:
  - type: script
    language: starlark
    source: |
      # This runs AFTER the allow/deny decision
      # It cannot change whether the tool call proceeds
      count = session_get("tool_calls")
      session_set("tool_calls", (count or 0) + 1)
    timeout_ms: 1000
    when: always
```

### When conditions

Script effects support the `when` condition to control when they execute:

```yaml
effects:
  - type: script
    source: |
      # Only runs when the rule's validation passed
      session_set("successful_writes", (session_get("successful_writes") or 0) + 1)
    when: on_pass

  - type: script
    source: |
      # Only runs when the rule's validation failed
      log("warning", "Write blocked: %s" % tool_input.file_path)
    when: on_fail

  - type: script
    source: |
      # Runs regardless of validation outcome
      project_set("last_write_attempt", timestamp())
    when: always  # default
```

## Starlark language basics

Starlark uses Python syntax with a restricted feature set. Developers familiar with Python will feel immediately at home. This section covers the essentials; the [Starlark specification](https://github.com/bazelbuild/starlark/blob/master/spec.md) provides comprehensive coverage.

Variables use simple assignment and are dynamically typed:

```python
name = "test.py"
count = 42
ratio = 3.14
enabled = True
items = ["a", "b", "c"]
config = {"key": "value", "nested": {"inner": 1}}
```

Dicts use standard Python `{}` syntax with string keys and values of any type.

Control flow uses Python-style indentation and keywords:

```python
if count > 10:
    level = "high"
elif count > 5:
    level = "medium"
else:
    level = "low"

for item in items:
    process(item)
```

The `go.starlark.net` implementation supports `while` loops and recursion with configurable limits, though many Starlark implementations intentionally restrict these to guarantee termination. arci enables both but enforces step limits to prevent runaway execution.

```python
while count > 0:
    count -= 1
```

Functions are defined with `def`:

```python
def is_sensitive_path(path):
    return ".env" in path or "secret" in path or "credential" in path

result = is_sensitive_path(tool_input["path"])
```

String methods provide text manipulation without regex complexity:

```python
path = "/home/user/project/.env"
".env" in path              # True
path.startswith("/home")    # True
path.endswith(".yaml")      # False
path.split("/")             # ["", "home", "user", "project", ".env"]
path.upper()                # "/HOME/USER/PROJECT/.ENV"
len(path)                   # 24
path.find(".env")           # 19 (-1 if not found)
```

Lists support indexing, slicing, and comprehensions:

```python
files = ["main.rs", "lib.rs", "test.rs"]
len(files)                              # 3
files[0]                                # "main.rs"
files.append("mod.rs")                  # mutates the list
test_files = [f for f in files if "test" in f]  # ["test.rs"]
files[1:]                               # ["lib.rs", "test.rs", "mod.rs"]
```

## Resource limits

Scripts execute with strict resource constraints to prevent runaway execution from blocking hook evaluation.

The `timeout_ms` setting controls maximum wall-clock execution time. When a script exceeds its timeout, execution is terminated and the script fails open: macros return a default value (null/false), and effects are skipped. Evaluation continues as if the script completed normally. The default timeout is 5000 milliseconds.

Step counting limits CPU usage independently of wall-clock time. The `go.starlark.net` `Thread` object supports a step counter that increments with each Starlark operation. A tight loop that would run forever is terminated after the step limit is reached, even if wall-clock time has not expired. The default step limit is 1,000,000 steps.

Memory limits prevent scripts from allocating unbounded data structures. The default memory limit is 10 megabytes. Scripts that exceed this limit are terminated.

When any limit is exceeded, script execution stops immediately. For macros, a default value is returned to the calling CEL expression. For effects, the effect is skipped. An error is logged with details about which limit was exceeded. Evaluation continues with remaining rules and effects. The policy evaluation succeeds, preserving fail-open semantics.

These limits can be configured globally or per-script:

```yaml
macros:
  - name: complex_analysis
    language: starlark
    source: |
      def complex_analysis(data):
          # Complex computation that needs more resources
          ...
    timeout_ms: 10000
    max_steps: 5000000
    max_memory_mb: 50

effects:
  - type: script
    source: |
      # Complex effect that needs more time
      ...
    timeout_ms: 10000
    max_steps: 5000000
    max_memory_mb: 50
```

## Error handling

Script errors follow arci' fail-open philosophy. A broken script never blocks the AI assistant; macros return default values and effects are skipped.

When a script fails, arci logs the error with context for debugging:

```
[WARN] Starlark macro 'is_destructive_command' failed:
  Error: name 'patterns' is not defined (line 3, column 15)
  Returning: false (default for boolean context)
  Context: tool_name=Bash, policy=security-checks

[WARN] Starlark effect failed in policy 'tracking', rule 'count-commands':
  Error: 'NoneType' object has no attribute 'append' (line 5, column 12)
  Effect skipped, continuing evaluation
  Context: tool_name=Bash, session=sess_abc123
```

Common error categories include the following.

Syntax errors occur when the script contains invalid Starlark code. These are caught during policy loading, before any evaluation begins. Missing colons, incorrect indentation, and unmatched parentheses all fall in this category.

Runtime errors occur during execution, such as accessing undefined variables, calling methods on wrong types, indexing beyond list bounds, or dividing by zero. The error message includes the line number and column.

Type errors occur when operations are applied to incompatible types, like adding a string to an integer without explicit conversion. Starlark is stricter than Python here and does not perform implicit type coercion.

Resource exhaustion occurs when scripts exceed their timeout, step count, or memory limit.

To debug script failures, start by checking the arci log for the specific error message and line number. Then test the script in isolation using `arci script test` with sample context. Adding intermediate `print()` calls can help trace execution, as output appears in logs. When the error is not obvious, simplify the script to isolate the failing logic.

The `arci script test` command provides an interactive environment for script development:

```bash
# Test a macro with sample input
arci script test --macro is_destructive_command \
  --args '["rm -rf /"]'

# Test an effect with sample context
arci script test --effect .arci/scripts/tracking.star \
  --tool-name Bash \
  --tool-input '{"command": "ls -la"}' \
  --session-id sess_test123

# Test inline script
arci script test --source 'session_get("count")' --session-id abc123
```

This command executes the script with the provided context and displays the return value, any logged output, and resource usage statistics.

## Practical examples

This section demonstrates common patterns for Starlark macros and effects.

### Complex pattern matching macro

A macro that performs sophisticated command analysis beyond simple substring matching:

```yaml
macros:
  - name: analyze_git_command
    language: starlark
    source: |
      def analyze_git_command(command):
          """Analyze a git command for risk factors."""
          if not command.startswith("git "):
              return {"is_git": False, "risk": "none", "factors": []}

          factors = []
          risk = "low"

          # Check for force operations
          if "--force" in command or "-f" in command.split():
              factors.append("force_flag")
              risk = "high"

          # Check for destructive operations
          destructive = ["reset --hard", "clean -f", "push --force", "branch -D"]
          for pattern in destructive:
              if pattern in command:
                  factors.append("destructive_" + pattern.replace(" ", "_").replace("-", ""))
                  risk = "critical"

          # Check for operations on protected branches
          protected = ["main", "master", "release", "production"]
          for branch in protected:
              if branch in command:
                  factors.append("protected_branch_" + branch)
                  if risk == "low":
                      risk = "medium"

          return {
              "is_git": True,
              "risk": risk,
              "factors": factors,
              "command": command,
          }

rules:
  - name: block-critical-git
    match:
      tools: [Bash]
    variables:
      - name: git_analysis
        expression: '$analyze_git_command(tool_input.command)'
    conditions:
      - expression: 'git_analysis.is_git'
    validate:
      expression: 'git_analysis.risk != "critical"'
      message: "Critical git operation blocked: {{ git_analysis.factors }}"
      action: deny
```

### State-tracking effect

An effect that maintains complex session state:

```yaml
rules:
  - name: track-file-operations
    match:
      tools: [Write, Edit, Read]
    effects:
      - type: script
        language: starlark
        source: |
          # Track file operations with timestamps and patterns
          history_key = "file_ops_history"
          history = session_get(history_key) or []

          entry = {
              "tool": tool_name,
              "path": tool_input.get("file_path", ""),
              "timestamp": timestamp(),
          }

          # Keep last 100 operations
          history = history[-99:] + [entry]
          session_set(history_key, history)

          # Analyze patterns
          recent = history[-20:] if len(history) >= 20 else history
          paths = [e["path"] for e in recent]
          unique_paths = len(set(paths))

          # Warn if touching too many files rapidly
          if len(recent) >= 20 and unique_paths > 15:
              log("warning", "High file churn: %d unique files in last 20 operations" % unique_paths)

          # Track per-file access counts
          counts_key = "file_access_counts"
          counts = project_get(counts_key) or {}
          path = tool_input.get("file_path", "")
          counts[path] = counts.get(path, 0) + 1
          project_set(counts_key, counts)
        when: always
```

### Workflow enforcement macro

A macro that checks workflow state for conditional validation:

```yaml
parameters:
  - name: workflow
    from: specs
    defaults:
      phase: "development"
      strict: false

macros:
  - name: check_workflow_rules
    language: starlark
    source: |
      def check_workflow_rules(phase, strict, tool, path):
          """Check if operation is allowed in current workflow phase."""

          # Define phase restrictions
          restrictions = {
              "review": {
                  "blocked_tools": ["Write", "Edit"],
                  "allowed_paths": ["docs/", "README"],
                  "message": "Code changes blocked during review phase",
              },
              "testing": {
                  "blocked_tools": [],
                  "allowed_paths": ["tests/", "test_", "_test."],
                  "message": "Only test files can be modified during testing phase",
              },
              "release": {
                  "blocked_tools": ["Bash"],
                  "allowed_paths": ["CHANGELOG", "version"],
                  "message": "Shell commands blocked during release phase",
              },
          }

          if phase not in restrictions:
              return {"allowed": True, "message": ""}

          rules = restrictions[phase]

          # Check blocked tools
          if tool in rules["blocked_tools"]:
              return {"allowed": not strict, "message": rules["message"]}

          # Check path restrictions (if any allowed_paths defined)
          if rules["allowed_paths"] and path:
              allowed = False
              for pattern in rules["allowed_paths"]:
                  if pattern in path:
                      allowed = True
                      break
              if not allowed:
                  return {"allowed": not strict, "message": rules["message"]}

          return {"allowed": True, "message": ""}

rules:
  - name: enforce-workflow
    variables:
      - name: workflow_check
        expression: |
          $check_workflow_rules(
            params.workflow.phase,
            params.workflow.strict,
            tool_name,
            tool_input.file_path ?? ""
          )
    validate:
      expression: 'workflow_check.allowed'
      message: "{{ workflow_check.message }}"
      action: '{{ params.workflow.strict ? "deny" : "warn" }}'
```
