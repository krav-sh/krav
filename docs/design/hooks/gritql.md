# Structural code analysis with GritQL

GritQL is a query language for structural code search and transformation. Unlike regular expressions, which operate on text, GritQL understands syntax trees and can match code patterns regardless of formatting, whitespace, or superficial differences. ARCI integrates GritQL as an optional feature for policies that need to analyze code structure rather than raw text.

The key insight of GritQL is that any valid code snippet in backticks is a valid pattern. Policy authors don't need to understand abstract syntax trees or parser internals. Writing `` `console.log($msg)` `` matches any console.log call and captures the argument as `$msg`. This makes structural matching accessible without requiring deep language knowledge.

ARCI delegates to the grit command-line tool rather than embedding TreeSitter grammars directly. This keeps the ARCI package lightweight and ensures users get the latest GritQL features and language support without ARCI releases. If the grit command-line tool is not installed, GritQL support logs a warning and the engine skips it entirely, maintaining fail-open semantics.

## Use cases

GritQL is valuable when regular expression matching is fragile or insufficient. Common scenarios include detecting specific API usage patterns (finding all calls to a deprecated function regardless of how the code formats arguments), matching code constructs that span many lines (function definitions with specific signatures), and identifying unsafe patterns that require understanding code structure (SQL injection via string concatenation rather than parameterized queries).

For simple substring or pattern matching, CEL's `.contains()` and `.matches()` methods remain appropriate. GritQL adds value when you need to distinguish between syntactically different uses of the same text. A regular expression matching `console.log` also matches comments containing the text, variable names like `console_log_enabled`, or strings containing `"console.log"`. GritQL matches only actual console.log function calls.

## Integration with the policy model

GritQL integrates with the policy model through CEL functions that work in conditions, variables, and check expressions. No separate "GritQL action type" exists. GritQL functions plug into the standard policy evaluation pipeline.

The primary functions are:

`$gritql_matching(content, pattern)` returns true if the content contains at least one match for the pattern. Use this for boolean conditions.

`$gritql_matches(content, pattern)` returns detailed match information including count and matched text. Use this when you need to inspect the matches or count occurrences.

Both functions accept an optional third parameter for language specification when auto-detection is insufficient.

### Using GritQL in check expressions

The most common use of GritQL is in check expressions, where you test whether code matches a problematic pattern:

```yaml
version: 1
name: no-console-log

match:
  events: [pre_tool_call]
  tools: [Write]
  paths:
    include: ["src/**/*.js", "src/**/*.ts"]
    exclude: ["**/*.test.*"]

rules:
  - name: block-console-log
    validate:
      expression: '!$gritql_matching(tool_input.content, "`console.log($...args)`")'
      message: "console.log statements should be removed before production"
      action: warn
```

The check expression returns true when the code is acceptable (no matches) and false when it contains the pattern. When the check fails, the specified action and message apply.

For blocking dangerous patterns, use `action: deny`:

```yaml
version: 1
name: no-eval

match:
  events: [pre_tool_call]
  tools: [Write]

variables:
  - name: is_javascript
    expression: 'tool_input.file_path.endsWith(".js") || tool_input.file_path.endsWith(".ts")'

rules:
  - name: block-eval
    conditions:
      - expression: 'is_javascript'
    validate:
      expression: '!$gritql_matching(tool_input.content, "`eval($code)`")'
      message: |
        Blocked use of eval().

        eval() executes arbitrary code and poses security risks.
        Consider using safer alternatives like JSON.parse() for data
        or Function constructor for controlled dynamic code.
      action: deny
```

### Using GritQL in conditions

GritQL functions can also appear in conditions to determine whether a rule applies. This is useful when you need structural matching as part of deciding relevance rather than validation:

```yaml
version: 1
name: async-error-handling

match:
  events: [pre_tool_call]
  tools: [Write, Edit]

rules:
  - name: require-try-catch
    conditions:
      - expression: '$gritql_matching(tool_input.content, "`await $expr`")'
    validate:
      expression: '$gritql_matching(tool_input.content, "`try { $...body } catch ($e) { $...handler }`")'
      message: |
        Async code should include error handling.
        Found await expressions without surrounding try/catch.
      action: warn
```

This rule only applies to code containing await expressions (checked in conditions), then validates that such code also contains try/catch blocks.

## Pattern syntax

GritQL patterns use a concise syntax that closely resembles the target language. The official GritQL documentation at grit.io provides full coverage; this section highlights the most useful features for policy authors.

### Backtick patterns

The simplest patterns are code snippets in backticks. The pattern `` `function $name() { }` `` matches any function declaration and captures its name. The pattern `` `import $module from "$path"` `` matches ES6 imports and captures both the imported binding and the module path.

Backtick patterns match structurally, ignoring whitespace and formatting differences. A pattern written on one line matches code formatted across many lines. This is one of GritQL's key advantages over regular expressions.

### Metavariables

Metavariables capture matched content for use in conditions or message templates. Single-word metavariables like `$name` match a single AST node. Multi-capture metavariables like `$...args` match zero or more nodes, useful for function arguments or list elements.

The pattern `` `function $name($...params) { $...body }` `` matches any function definition, capturing its name, parameter list, and body separately.

### Where clauses

Where clauses add constraints to patterns. The syntax `where { $var <: "pattern" }` requires the captured variable to match a further pattern.

The pattern `` `fetch($url, $options)` where { $url <: `"http://$_"` } `` matches fetch calls where the URL is a string literal starting with `http://` rather than `https://`.

### Common patterns

Matching function calls: `` `someFunction($...args)` ``

Matching method calls: `` `$receiver.methodName($...args)` ``

Matching imports: `` `import { $...imports } from "$module"` ``

Matching class definitions: `` `class $name extends $parent { $...body }` ``

Matching variable declarations: `` `const $name = $value` `` or `` `let $name = $value` ``

## GritQL functions

ARCI provides two CEL functions for GritQL pattern matching.

### $gritql_matching

The `$gritql_matching(content, pattern)` function returns true if the content matches the pattern, false otherwise. Use this for boolean conditions and check expressions.

```yaml
version: 1
name: require-error-handling

match:
  events: [pre_tool_call]
  tools: [Write]
  paths:
    include: ["**/*.ts", "**/*.js"]

rules:
  - name: async-needs-catch
    conditions:
      - expression: '$gritql_matching(tool_input.content, "`await $expr`")'
    validate:
      expression: '$gritql_matching(tool_input.content, "`try { $...body } catch ($e) { $...handler }`")'
      message: |
        Async code should include error handling.
        Found await expressions without surrounding try/catch.
      action: warn
```

### $gritql_matches

The `$gritql_matches(content, pattern)` function returns detailed match information. Use this when you need to inspect the results, count occurrences, or extract captured values.

The function returns an object with:

- `count` (number): Number of matches found
- `matches` (array): Array of match details, each with `text` and captured metavariables
- `matched` (boolean): Convenience field, true if count > 0

```yaml
version: 1
name: limit-todos

match:
  events: [pre_tool_call]
  tools: [Write]

variables:
  - name: todo_matches
    expression: '$gritql_matches(tool_input.content, "`// TODO: $msg`")'

rules:
  - name: too-many-todos
    validate:
      expression: 'todo_matches.count <= 5'
      message: "File contains {{ todo_matches.count }} TODO comments; limit is 5"
      action: warn
```

### Language specification

Both functions accept an optional third parameter for explicit language specification:

```yaml
variables:
  - name: has_eval
    expression: '$gritql_matching(tool_input.content, "`eval($code)`", "javascript")'
```

GritQL infers language from file extensions in most cases, but explicit specification is useful when analyzing code snippets without file context or when the extension is ambiguous.

## Performance considerations

GritQL parses source files to build syntax trees before pattern matching. For small files this overhead is negligible, but for large files or high-frequency policies, consider scoping GritQL checks appropriately.

Pre-filter with fast checks (file extension, simple string contains) before invoking GritQL:

```yaml
version: 1
name: react-safety

match:
  events: [pre_tool_call]
  tools: [Write]
  paths:
    include: ["**/*.tsx", "**/*.jsx"]

variables:
  - name: content
    expression: 'tool_input.content'

rules:
  - name: dangerous-html-check
    conditions:
      - expression: 'content.contains("dangerouslySetInnerHTML")'  # fast pre-filter
    validate:
      expression: '!$gritql_matching(content, "`dangerouslySetInnerHTML={{ $html }}`")'
      message: "Use of dangerouslySetInnerHTML detected"
      action: warn
```

This policy first applies structural matching via the `paths` field, then uses a fast string contains check in conditions, and only invokes GritQL parsing when those checks pass.

## Fail-open behavior

GritQL integration follows ARCI's fail-open semantics. If the grit command-line tool is not installed, GritQL functions return false (for `$gritql_matching`) or an empty result (for `$gritql_matches`) and log a warning. Policy evaluation continues, and missing tooling does not block operations.

Specific fail-open scenarios:

When the grit command-line tool is not found in PATH, all GritQL operations log a warning like "grit not found; skipping GritQL pattern matching" and return no-match results. The warning appears once per server startup, not on every evaluation.

When a GritQL pattern has syntax errors, the function logs the error and returns a no-match result. The policy continues evaluating with GritQL checks effectively skipped.

When the grit command-line tool times out (default 5 seconds), the function returns a no-match result. The timeout is configurable via runtime settings.

When file content is not valid source code for the detected language, GritQL parsing may fail. The function logs the error and returns a no-match result.

To verify grit availability:

```bash
# Check if grit is installed
grit --version

# Install via homebrew
brew install grit

# Or via npm
npm install -g @getgrit/cli
```

The `arci doctor` command includes a check for grit availability and reports whether GritQL features work.

## Example policies

These examples show common uses of GritQL in ARCI policies.

### Warn about console.log in production code

```yaml
version: 1
name: no-console-log

match:
  events: [pre_tool_call]
  tools: [Write]
  paths:
    include: ["src/**/*.js", "src/**/*.ts", "src/**/*.jsx", "src/**/*.tsx"]
    exclude: ["**/*.test.*", "**/__tests__/**"]

rules:
  - name: warn-console-log
    validate:
      expression: '!$gritql_matching(tool_input.content, "`console.log($...args)`")'
      message: |
        Found console.log in production code.

        Remove debug logging before committing, or use a proper
        logging framework that can be disabled in production.
      action: warn
```

### Track to-do comments

```yaml
version: 1
name: track-todos

match:
  events: [post_tool_call]
  tools: [Write]

variables:
  - name: todo_count
    expression: '$gritql_matches(tool_output.content ?? "", "`// TODO: $msg`").count'

rules:
  - name: log-todos
    conditions:
      - expression: 'todo_count > 0'
    effects:
      - type: log
        level: info
        message: "File contains {{ todo_count }} TODO comments"
```

### Warn about deprecated API methods

```yaml
version: 1
name: deprecation-warnings

config:
  priority: high

match:
  events: [pre_tool_call]
  tools: [Write]
  paths:
    include: ["**/*.py"]

rules:
  - name: old-method-deprecated
    validate:
      expression: '!$gritql_matching(tool_input.content, "`$obj.old_method($...args)`", "python")'
      message: |
        old_method() is deprecated.

        Use new_method() instead. See migration guide at:
        https://docs.example.com/migration
      action: warn
```

### Block SQL injection patterns

```yaml
version: 1
name: sql-injection-prevention

config:
  priority: critical

match:
  events: [pre_tool_call]
  tools: [Write]
  paths:
    include: ["**/*.py"]

variables:
  - name: content
    expression: 'tool_input.content'

rules:
  - name: block-sql-string-concat
    conditions:
      - expression: 'content.contains("cursor.execute")'  # fast pre-filter
    validate:
      expression: |
        !$gritql_matching(content, "`cursor.execute($query)` where { $query <: `$left + $right` }", "python")
      message: |
        Blocked SQL query using string concatenation.

        String concatenation in SQL queries enables SQL injection.
        Use parameterized queries instead:
          cursor.execute("SELECT * FROM users WHERE id = ?", (user_id,))
      action: deny
```

### Check React hook dependencies

```yaml
version: 1
name: react-hook-hints

config:
  priority: low

match:
  events: [pre_tool_call]
  tools: [Write]
  paths:
    include: ["**/*.jsx", "**/*.tsx"]

rules:
  - name: empty-deps-warning
    validate:
      expression: '!$gritql_matching(tool_input.content, "`useEffect($callback, [])`")'
      message: |
        useEffect with empty dependency array detected.

        Verify this is intentional. If the callback references any
        props or state, they should be in the dependency array.
      action: warn
```

## Relationship to other tools

GritQL complements rather than replaces other pattern matching capabilities in ARCI.

CEL's `.contains()` and `.matches()` methods remain the primary tools for simple text matching. They're faster, require no external dependencies, and handle most validation needs.

Shell effects can invoke any external tool, including dedicated linters or analyzers. Use these for complex analysis that GritQL does not cover or when you need to use existing tool configurations.

GritQL fills the gap between simple regular expressions and full static analysis. When you need to understand code structure but do not need a complete type system or data flow analysis, GritQL provides the right level of power without the complexity of embedding a full analyzer.

For teams already using GritQL for code transformation (via grit's rewrite capabilities), the integration allows reusing patterns between ARCI policies and standalone grit workflows. You can promote a pattern developed for ARCI to a grit transformation, or vice versa.
