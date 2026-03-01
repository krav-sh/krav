# Builtins

arci ships with a collection of curated builtin policies that users can opt into. These policies encode common safety patterns, workflow guardrails, and productivity enhancements that apply broadly across projects and Claude Code. Builtins are implemented as a rules-only extension that ships with the core package, sitting at the lowest precedence level so users can easily override or disable any policy.

## Design rationale

Many arci users want sensible defaults without having to author policies from scratch. Dangerous command patterns, git hygiene practices, and tool usage guidance are universal concerns that don't vary much across projects. Rather than expecting every user to rediscover and encode these patterns, arci provides a curated set of builtin policies that capture community best practices.

At the same time, builtins must not impose opinions on users who don't want them. Some teams have legitimate reasons to run `rm -rf /` or force-push to main. Builtins are therefore opt-in at the category level and individually overridable. The precedence system ensures that any user or project policy with the same name completely replaces the builtin version.

Builtins also serve as documentation. Even users who disable certain policies benefit from seeing what patterns arci considers worth addressing. The builtin policies demonstrate expression language features, action patterns, and state management techniques that users can adapt for their own policies.

## Architecture

Builtins are implemented as a rules-only extension within the arci package itself. The `arci/builtins/` directory contains YAML policy files organized by category. At daemon startup, these policies are discovered through the standard extension mechanism and loaded at the lowest precedence level.

```
arci/
  builtins/
    extension.toml           # Extension manifest for builtin registration
    policies.d/
      safety.yaml            # Dangerous command protection
      git.yaml               # Git workflow guardrails
      tool-guidance.yaml     # Proper tool usage patterns
      code-quality.yaml      # Lint suppression warnings
```

The `extension.toml` declares the builtin extension:

```toml
[extension]
name = "arci-builtins"
version = "0.1.0"
description = "Built-in policies for common safety and workflow patterns"

[policies]
paths = ["policies.d/*.yaml"]
```

This means builtins participate in the same discovery, loading, and precedence mechanics as external extensions. The only difference is that they ship with the package rather than requiring separate installation.

## Opting into builtins

Builtins are disabled by default. Users opt in by enabling specific categories in their `policies.json`:

```json
{
  "$schema": "https://arci.dev/schemas/policies.json",
  "defaultBehavior": "all-enabled",
  "enabled": [
    "arci:block-dangerous-rm",
    "arci:block-curl-pipe-shell",
    "arci:block-chmod-dangerous",
    "arci:warn-env-file-read"
  ]
}
```

Alternatively, users can enable entire categories using a glob pattern in their `arci.yaml`:

```yaml
# <user-config-dir>/arci/arci.yaml
builtins:
  safety: true
  git: true
  tool-guidance: true
  code-quality: false  # Explicitly disabled (same as omitting)
```

Each category corresponds to a YAML file in the builtins policies directory. Enabling a category loads all policies from that file. Users can then disable or override individual policies at higher precedence levels.

This category-based approach balances convenience with control. Users don't need to understand every policy before opting in, but they can drill down and customize once they encounter specific policies they want to adjust.

Project configuration can also enable builtins:

```yaml
# .arci/arci.yaml
builtins:
  safety: true
  git: true
```

When both user and project configuration specify builtins, the union of enabled categories is loaded. A category is loaded if either configuration enables it. To prevent a category from loading when user configuration enables it, projects would need to override individual policies rather than the category.

## Policy identification

Builtin policies use namespaced names to prevent collisions and make their origin clear:

```yaml
version: 1
name: arci:block-dangerous-rm
```

The `arci:` prefix identifies these as builtin policies. Users overriding a builtin policy reference the same name in their `policies.json`:

```json
{
  "disabled": ["arci:block-dangerous-rm"]
}
```

Or provide a complete policy definition in their `policies.d/` directory with the same name, which replaces the builtin entirely due to precedence.

## Builtin categories

### Safety policies

The safety category provides protection against dangerous operations that could cause data loss or system compromise.

**arci:block-dangerous-rm** blocks recursive deletion of root or home directories.

```yaml
version: 1
name: arci:block-dangerous-rm

metadata:
  description: Block rm -rf on root, home, or other critical paths

config:
  priority: critical

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'

rules:
  - name: dangerous-rm
    validate:
      expression: '!command.matches("^rm\\s+(-[rRf]+\\s+)+(/ |~|/home|/Users|\\$HOME)")'
      message: |
        Blocked potentially destructive rm command targeting critical path.
        Command: {{ command }}

        If this operation is intentional, override arci:block-dangerous-rm
        in your project configuration.
      action: deny
```

**arci:block-curl-pipe-shell** blocks piping curl/wget output directly to shell interpreters.

```yaml
version: 1
name: arci:block-curl-pipe-shell

metadata:
  description: Block curl/wget piped to sh/bash for remote code execution

config:
  priority: critical

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'

rules:
  - name: curl-pipe-shell
    validate:
      expression: '!command.matches("(curl|wget)\\s+.*\\|\\s*(sh|bash|zsh|fish)")'
      message: |
        Blocked remote code execution pattern: piping download to shell.
        Command: {{ command }}

        Download the script first, review it, then execute:
          curl -O <url> && cat script.sh && sh script.sh
      action: deny
```

**arci:block-chmod-dangerous** blocks overly permissive chmod operations.

```yaml
version: 1
name: arci:block-chmod-dangerous

metadata:
  description: Block chmod 777 and similar overly permissive modes

config:
  priority: high

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'

rules:
  - name: chmod-777
    validate:
      expression: '!command.matches("chmod\\s+(-R\\s+)?777")'
      message: |
        Blocked chmod 777 - this grants read/write/execute to everyone.
        Consider more restrictive permissions: 755 for directories, 644 for files.
      action: deny
```

**arci:warn-env-file-read** warns when reading files that may contain secrets.

```yaml
version: 1
name: arci:warn-env-file-read

metadata:
  description: Warn when reading files that commonly contain secrets

config:
  priority: medium

match:
  events: [pre_tool_call]
  tools: [Read]

variables:
  - name: file_path
    expression: 'tool_input.file_path ?? ""'
  - name: is_sensitive
    expression: |
      file_path.contains(".env") ||
      file_path.contains("credentials") ||
      file_path.matches("secrets?\\.") ||
      file_path.matches("\\.pem$") ||
      file_path.matches("\\.key$")

rules:
  - name: sensitive-file-warning
    conditions:
      - expression: 'is_sensitive'
    validate:
      expression: 'false'
      message: |
        Reading a file that may contain secrets: {{ file_path }}
        Ensure this content is not logged, committed, or exposed.
      action: warn
```

### Git policies

The git category provides guardrails for common git workflow mistakes.

**arci:convert-force-to-lease** converts `--force` to `--force-with-lease` for safer force pushes.

```yaml
version: 1
name: arci:convert-force-to-lease

metadata:
  description: Convert git push --force to --force-with-lease

config:
  priority: high

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: is_force_push
    expression: |
      command.matches("git\\s+push\\s+.*--force(?!-with-lease)") &&
      !command.contains("--force-with-lease")

rules:
  - name: force-to-lease
    conditions:
      - expression: 'is_force_push'
    mutate:
      expression: |
        object.with({
          tool_input: object.tool_input.with({
            command: command.replace("--force", "--force-with-lease")
          })
        })
    effects:
      - type: log
        level: info
        message: "Converted --force to --force-with-lease for safer push"
        when: always

  - name: force-to-lease-warning
    conditions:
      - expression: 'is_force_push'
    effects:
      - type: notify
        title: "Force push modified"
        message: |
          Converted --force to --force-with-lease for safer push.
          --force-with-lease fails if the remote has changes you haven't seen.
        when: always
```

**arci:warn-main-branch-push** warns before pushing directly to main or master.

```yaml
version: 1
name: arci:warn-main-branch-push

metadata:
  description: Warn before pushing to main/master branch

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
  - name: is_push_to_main
    expression: |
      command.matches("git\\s+push") &&
      (current_branch == "main" || current_branch == "master")

rules:
  - name: main-branch-push-warning
    conditions:
      - expression: 'is_push_to_main'
    validate:
      expression: 'false'
      message: |
        Pushing directly to {{ current_branch }} branch.
        Consider creating a feature branch and pull request instead.
      action: warn
```

**arci:block-force-push-main** blocks force pushing to protected branches.

```yaml
version: 1
name: arci:block-force-push-main

metadata:
  description: Block force push to main/master

config:
  priority: critical

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: is_force_push_to_main
    expression: |
      command.matches("git\\s+push\\s+.*--force") &&
      command.matches("\\s+(origin\\s+)?(main|master)(\\s|$)")

rules:
  - name: force-push-main-block
    conditions:
      - expression: 'is_force_push_to_main'
    validate:
      expression: 'false'
      message: |
        Blocked force push to protected branch.
        Force pushing to main/master rewrites shared history.

        If this is intentional, override arci:block-force-push-main.
      action: deny
```

**arci:warn-uncommitted-changes** warns when certain operations run with uncommitted changes.

```yaml
version: 1
name: arci:warn-uncommitted-changes

metadata:
  description: Warn about uncommitted changes before destructive git operations

config:
  priority: medium

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: is_destructive_git
    expression: 'command.matches("git\\s+(reset|checkout|rebase|merge)")'
  - name: has_uncommitted
    expression: '$git_is_dirty()'

rules:
  - name: uncommitted-changes-warning
    conditions:
      - expression: 'is_destructive_git && has_uncommitted'
    validate:
      expression: 'false'
      message: |
        Running {{ command }} with uncommitted changes.
        Consider stashing or committing first: git stash
      action: warn
```

### Tool guidance policies

The tool guidance category helps Claude Code use appropriate tools for common operations.

**arci:redirect-bash-find** suggests using the Glob tool instead of bash find.

```yaml
version: 1
name: arci:redirect-bash-find

metadata:
  description: Suggest Glob tool instead of bash find

config:
  priority: medium

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'

rules:
  - name: find-to-glob
    conditions:
      - expression: 'command.matches("^find\\s+")'
    validate:
      expression: 'false'
      message: |
        Consider using the Glob tool instead of bash find.
        Glob provides better integration with the assistant's context.
      action: warn
```

**arci:redirect-bash-grep** suggests using the Grep tool instead of bash grep.

```yaml
version: 1
name: arci:redirect-bash-grep

metadata:
  description: Suggest Grep tool instead of bash grep/rg

config:
  priority: medium

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'

rules:
  - name: grep-to-tool
    conditions:
      - expression: 'command.matches("^(grep|rg)\\s+")'
    validate:
      expression: 'false'
      message: |
        Consider using the Grep tool instead of bash grep/rg.
        The Grep tool provides better output formatting and context.
      action: warn
```

**arci:redirect-bash-cat** suggests using the Read tool instead of bash cat.

```yaml
version: 1
name: arci:redirect-bash-cat

metadata:
  description: Suggest Read tool instead of bash cat

config:
  priority: medium

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: is_simple_cat
    expression: 'command.matches("^cat\\s+[^|]") && !command.contains("|")'

rules:
  - name: cat-to-read
    conditions:
      - expression: 'is_simple_cat'
    validate:
      expression: 'false'
      message: |
        Consider using the Read tool instead of bash cat.
        The Read tool provides line numbers and better formatting.
      action: warn
```

### Code quality policies

The code quality category encourages fixing issues rather than suppressing them.

**arci:warn-type-ignore** warns when adding type: ignore comments.

```yaml
version: 1
name: arci:warn-type-ignore

metadata:
  description: Warn when adding type ignore comments

config:
  priority: high

match:
  events: [pre_tool_call]
  tools: [Write, Edit]

variables:
  - name: content
    expression: 'tool_input.content ?? ""'

rules:
  - name: type-ignore-warning
    conditions:
      - expression: 'content.contains("# type: ignore")'
    validate:
      expression: 'false'
      message: |
        Adding type: ignore comment suppresses type checker warnings.

        Consider fixing the underlying type issue instead:
        - Add proper type annotations
        - Use typing.cast() for legitimate type narrowing
        - Fix the actual type mismatch

        Suppressions hide bugs and make refactoring harder.
      action: warn
```

**arci:warn-noqa** warns when adding noqa comments.

```yaml
version: 1
name: arci:warn-noqa

metadata:
  description: Warn when adding noqa comments

config:
  priority: medium

match:
  events: [pre_tool_call]
  tools: [Write, Edit]

variables:
  - name: content
    expression: 'tool_input.content ?? ""'

rules:
  - name: noqa-warning
    conditions:
      - expression: 'content.contains("# noqa")'
    validate:
      expression: 'false'
      message: |
        Adding noqa comment suppresses linter warnings.

        Consider fixing the underlying issue instead. If the linter
        rule is inappropriate for this codebase, disable it globally
        in your linter configuration rather than per-line.
      action: warn
```

**arci:warn-eslint-disable** warns when adding eslint-disable comments.

```yaml
version: 1
name: arci:warn-eslint-disable

metadata:
  description: Warn when adding eslint-disable comments

config:
  priority: medium

match:
  events: [pre_tool_call]
  tools: [Write, Edit]

variables:
  - name: content
    expression: 'tool_input.content ?? ""'

rules:
  - name: eslint-disable-warning
    conditions:
      - expression: 'content.contains("eslint-disable")'
    validate:
      expression: 'false'
      message: |
        Adding eslint-disable suppresses linter warnings.

        Consider fixing the underlying issue instead. If the rule
        is inappropriate for this codebase, configure it in your
        eslint config rather than disabling per-line.
      action: warn
```

## State-dependent patterns

Some builtin policies demonstrate state-dependent patterns using the state store. These policies show how to implement escalating warnings that transition from warn to deny after repeated occurrences.

**arci:escalating-warning-template** demonstrates the warn-then-block pattern. This template warns on the first two occurrences and blocks on the third.

```yaml
version: 1
name: arci:escalating-warning-template

metadata:
  description: Template for warn-then-block patterns using session state

config:
  priority: high

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'
  - name: matches_risky_pattern
    expression: 'command.matches("some-risky-pattern")'
  - name: attempt_count
    expression: '$session_get("arci:escalation:risky-pattern", 0)'
  - name: max_warnings
    expression: '2'

rules:
  - name: track-attempts
    conditions:
      - expression: 'matches_risky_pattern'
    effects:
      - type: setState
        scope: session
        key: arci:escalation:risky-pattern
        value: '{{ attempt_count + 1 }}'
        when: always

  - name: warn-on-early-attempts
    conditions:
      - expression: 'matches_risky_pattern && attempt_count < max_warnings'
    validate:
      expression: 'false'
      message: |
        Warning: risky pattern detected ({{ attempt_count + 1 }}/{{ max_warnings + 1 }}).
        This will be blocked after {{ max_warnings }} warnings in this session.
      action: warn

  - name: block-on-repeated-attempts
    conditions:
      - expression: 'matches_risky_pattern && attempt_count >= max_warnings'
    validate:
      expression: 'false'
      message: |
        Blocked: risky pattern repeated too many times.
        This pattern was warned {{ max_warnings }} times and is now blocked for this session.
      action: deny
```

This template demonstrates several key concepts. The policy uses `$session_get()` to retrieve the current attempt count with a default of 0. The `setState` effect increments the counter on every match using `when: always`. Two separate validation rules handle the warning and blocking phases, with conditions that check the attempt count. Users can copy and adapt this pattern for their own escalating policies by changing the pattern match and adjusting `max_warnings`.

## Viewing enabled builtins

The CLI provides commands to inspect which builtins are active:

```bash
# List all policies including builtins
arci hook policy list

# Show policies from the builtins extension
arci hook policy list --source builtins

# Show a specific builtin policy
arci hook policy get arci:block-dangerous-rm
```

The dashboard also displays enabled builtins alongside user and project policies, with clear visual distinction for their origin.

## Overriding builtins

Users can override builtins in three ways.

To disable a specific policy without replacing it, add it to `disabled` in `policies.json`:

```json
{
  "disabled": ["arci:warn-type-ignore"]
}
```

To completely redefine a policy, create a policy with the same name in your `policies.d/` directory. Because user and project policies have higher precedence than builtins, your policy replaces the builtin entirely:

```yaml
version: 1
name: arci:block-dangerous-rm

metadata:
  description: Custom dangerous rm policy for production paths

config:
  priority: critical

match:
  events: [pre_tool_call]
  tools: [Bash]

variables:
  - name: command
    expression: 'tool_input.command ?? ""'

rules:
  - name: production-rm-block
    conditions:
      - expression: 'command.matches("^rm\\s+") && $matches_glob(command, "*/production/*")'
    validate:
      expression: 'false'
      message: "Custom message for production paths"
      action: deny
```

To adjust a single aspect of a builtin without fully redefining it, you would need to copy the full policy and modify it, since policies are atomic units that replace entirely rather than merging.

## Future categories

Additional builtin categories may be added in future versions.

Framework-specific policies could provide convention enforcement for Python (pytest patterns, import organization), JavaScript (package.json hygiene, lockfile enforcement), and other ecosystems. These would be opt-in categories like `builtins.python: true`.

Security policies could extend the safety category with patterns for preventing secrets in commits, detecting SQL injection patterns in string concatenation, and similar concerns.

Workflow policies could provide templates for common multi-step workflows like "test before commit" or "lint on save" that users can enable and customize.

New categories follow the same opt-in model. Upgrading arci never automatically enables new builtin categories—users must explicitly opt in after reviewing what's included.
