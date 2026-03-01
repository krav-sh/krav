# Match schema design

ARCI separates structural matching from conditional logic following patterns established by Kubernetes admission control. Unlike Kubernetes, ARCI uses a unified `Policy` type that can contain both mutations and validations.

## Schema

### Policy: matchConstraints + matchConditions

The policy declares what it's capable of handling (structural) and when it should apply (conditional).

```yaml
apiVersion: arci.dev/v1
kind: Policy
metadata:
  name: typescript-quality
spec:
  # Structural: what this policy understands
  matchConstraints:
    events:
      - pre_tool_call
      - post_tool_call

    tools:
      - Bash
      - Write
      - Edit

    paths:
      include:
        - "**/*.ts"
        - "**/*.tsx"
      exclude:
        - "**/node_modules/**"
        - "**/*.test.ts"

    branches:
      include:
        - main
        - "release/*"
      exclude:
        - "feature/*"

  # Conditional: CEL expressions that can reference params
  matchConditions:
    - name: strict-mode-enabled
      expression: "params.strictMode"

    - name: not-junior-experimenting
      expression: '!(params.userRole == "junior" && params.workflowPhase == "experimenting")'

  # Policy logic defined in mutations and validations
  # (see POLICY_SCHEMA.md for full details)
```

### Binding: matchResources only

The binding specifies deployment scope using structural selectors only. No CEL here.

```yaml
apiVersion: arci.dev/v1
kind: PolicyBinding
metadata:
  name: typescript-quality-prod
spec:
  policyRef:
    name: typescript-quality

  paramRef:
    name: strict-params

  # Structural only - narrows the policy's matchConstraints
  matchResources:
    paths:
      include:
        - "src/**"
      exclude:
        - "src/generated/**"

    branches:
      include:
        - main

  actions: [deny] # what happens when validations fail
  mutationsEnabled: true # whether to apply mutations from the policy
```

## Matching semantics

ARCI follows Kubernetes ValidatingAdmissionPolicy matching semantics closely. Understanding the AND/OR logic at each level is essential for predictable policy behavior.

### Within array fields: OR logic

When a structural field contains multiple values, a match occurs if ANY value matches (OR logic).

```yaml
# Matches if tool is Bash OR Write OR Edit
tools:
  - Bash
  - Write
  - Edit

# Matches if event is pre_tool_call OR post_tool_call
events:
  - pre_tool_call
  - post_tool_call

### Across structural fields: AND logic

All specified structural fields must match (AND logic). Omitted fields match everything.

```

```yaml
matchConstraints:
  # Must match ALL of these:
  events:
    - pre_tool_call # event must be pre_tool_call
  tools:
    - Bash # AND tool must be Bash OR Write
    - Write
  paths:
    include:
      - "src/**" # AND path must be under src/
```

### Include/exclude: Exclude takes precedence

When both `include` and `exclude` lists contain patterns, exclude patterns always take precedence. This matches Kubernetes `excludeResourceRules` behavior.

```yaml
paths:
  include:
    - "**/*.ts" # Start with all TypeScript files
  exclude:
    - "**/*.test.ts" # Remove test files (takes precedence)
    - "**/node_modules/**" # Remove node_modules
```

The evaluation order is:

1. If only `include` specified: only those patterns match
2. If only `exclude` specified: everything except those patterns matches
3. If both specified: the system evaluates include first, then exclude removes from that set

### matchConditions: AND with short circuit

All matchConditions must evaluate to true (AND logic). Evaluation short-circuits on first false.

```yaml
matchConditions:
  # ALL must be true:
  - name: strict-mode
    expression: "params.strictMode"
  - name: not-junior
    expression: 'params.userRole != "junior"'
  - name: production-branch
    expression: '$current_branch() == "main"'
```

The exact matching logic follows Kubernetes semantics:

1. If ANY matchCondition evaluates to FALSE → the system skips the policy
2. If ALL matchConditions evaluate to TRUE → the system evaluates the policy
3. If any evaluates to error (but none FALSE):
   - If failurePolicy=Fail → the hook request fails
   - If failurePolicy=Ignore → the system skips the policy

### Binding narrowing: Intersection only

A binding's `matchResources` intersects with (never widens) the policy's `matchConstraints`. The binding can only further restrict what the policy matches.

```yaml
# Policy: can handle all TypeScript files
spec:
  matchConstraints:
    paths:
      include:
        - "**/*.ts"

---
# Binding: narrows to only src/ TypeScript files
spec:
  matchResources:
    paths:
      include:
        - "src/**/*.ts"  # Valid: subset of **/*.ts

# This binding would be rejected:
spec:
  matchResources:
    paths:
      include:
        - "**/*.py"      # Invalid: outside policy's constraints
```

### Empty/omitted fields

Omitted structural fields match everything (universal match). This differs from an explicitly empty array, which matches nothing.

```yaml
# Omitted tools field = matches all tools
matchConstraints:
  events:
    - pre_tool_call
  # tools: not specified, matches all

# Explicitly empty = matches nothing (unusual but valid)
matchConstraints:
  tools: []  # Matches no tools - policy never applies
```

## Structural fields

### Events

Hook event types. Most policies only care about `pre_tool_call`.

```yaml
events:
  - pre_tool_call # Before tool executes (can block/modify)
  - post_tool_call # After tool executes (can react)
  - session_start # Session began
  - session_end # Session ended
```

### Tools

Claude Code tool names.

```yaml
tools:
  - Bash # Shell command execution
  - Write # Create/overwrite file
  - Edit # Modify existing file
  - MultiEdit # Multiple edits in one call
  - Read # Read file contents
  - Grep # Search files
  - Glob # List files by pattern
  - LS # List directory
```

### Paths

File path patterns using glob syntax. Only relevant for tools that operate on files.

```yaml
paths:
  include:
    - "src/**/*.ts"
    - "lib/**/*.ts"
  exclude:
    - "**/*.test.ts"
    - "**/*.spec.ts"
    - "**/node_modules/**"
```

See "Include/exclude: exclude takes precedence" in the Matching Semantics section for detailed evaluation logic.

### Branches

Git branch patterns. Useful for varying enforcement by branch. Same include/exclude semantics as paths.

```yaml
branches:
  include:
    - main
    - master
    - "release/*"
    - "hotfix/*"
  exclude:
    - "feature/*"
    - "experiment/*"
```

## Conditional fields (policy only)

### matchConditions

CEL expressions evaluated at runtime after structural matching succeeds. Have access to:

- `params.*` - from the binding's paramRef
- `object.*` - the tool call under evaluation
- `tool_name`, `tool_input` - current hook event
- `session.*` - session context
- `project.*` - project context
- Custom functions: `$file_exists()`, `$current_branch()`, etc.

```yaml
matchConditions:
  - name: strict-mode
    expression: "has(params.strictMode) && params.strictMode"

  - name: senior-only
    expression: 'params.userRole in ["senior", "staff", "principal"]'

  - name: not-during-review
    expression: 'params.workflowPhase != "review"'
```

All conditions must pass (AND logic) with short-circuit on first false. See the matchConditions short-circuit section for detailed evaluation semantics including error handling.

## Binding narrowing rules

A binding's matchResources can only narrow, never widen, the policy's matchConstraints.

**Valid:**

```yaml
# Policy allows **/*.ts
# Binding narrows to src/**/*.ts
matchResources:
  paths:
    include:
      - "src/**/*.ts"
```

**Invalid (config error):**

```yaml
# Policy allows **/*.ts
# Binding tries to widen to include Python
matchResources:
  paths:
    include:
      - "**/*.ts"
      - "**/*.py" # ERROR: outside policy's constraints
```

## Evaluation flow

For each (policy, binding, params) combination:

1. **Binding structural match**: Check binding's `matchResources` against current context using OR-within-arrays, AND-across-fields logic
2. **Skip if no match**: If binding doesn't match, skip this combination
3. **Policy structural match**: Check policy's `matchConstraints` against current context (should always pass if binding is properly narrowed)
4. **Skip if no match**: If policy constraints don't match, skip
5. **Resolve params**: Load parameter resource from binding's `paramRef`
6. **Evaluate matchConditions**: Run each CEL expression with params, object, and context
   - If ANY condition returns false → skip this policy (short-circuit)
   - If ALL conditions return true → proceed to policy logic
   - If any condition errors (but none false) → apply failurePolicy
7. **Collect mutations**: If `mutationsEnabled` is true, collect mutations from matching policies
8. **Apply mutations**: The system applies mutations in priority order (higher priority first), then declaration order within a policy
9. **Run validations**: Execute validation expressions against the (possibly mutated) state
10. **Apply actions**: Enforce based on binding's `actions` (deny, warn, audit)

## Defaults

When the user omits fields:

- `events`: defaults to `[pre_tool_call]`
- `tools`: defaults to all tools
- `paths`: defaults to all paths
- `branches`: defaults to all branches
- `matchConditions`: defaults to empty (always matches)

## Examples

### Policy that only makes sense for TypeScript

```yaml
apiVersion: arci.dev/v1
kind: Policy
metadata:
  name: typescript-strict
spec:
  matchConstraints:
    tools:
      - Write
      - Edit
    paths:
      include:
        - "**/*.ts"
        - "**/*.tsx"
      exclude:
        - "**/*.d.ts"
```

### Binding that narrows to src/ on main branch

```yaml
apiVersion: arci.dev/v1
kind: PolicyBinding
metadata:
  name: typescript-strict-prod
spec:
  policyRef:
    name: typescript-strict
  matchResources:
    paths:
      include:
        - "src/**"
    branches:
      include:
        - main
  actions: [deny]
```

### Policy with conditional matching via params

```yaml
apiVersion: arci.dev/v1
kind: Policy
metadata:
  name: strict-mode-checks
spec:
  matchConstraints:
    tools:
      - Bash
      - Write
      - Edit

  matchConditions:
    - name: strict-mode-on
      expression: "params.strictMode == true"

    - name: not-experimenting
      expression: 'params.workflowPhase != "experimenting"'
```

### Different params for different contexts

```yaml
# Binding for normal work
apiVersion: arci.dev/v1
kind: PolicyBinding
metadata:
  name: strict-mode-normal
spec:
  policyRef:
    name: strict-mode-checks
  paramRef:
    name: normal-params # strictMode: true
  actions: [deny]

---
# Binding for experiments directory
apiVersion: arci.dev/v1
kind: PolicyBinding
metadata:
  name: strict-mode-experiments
spec:
  policyRef:
    name: strict-mode-checks
  paramRef:
    name: relaxed-params # strictMode: false
  matchResources:
    paths:
      include:
        - "experiments/**"
  actions: [warn]
```

In the second binding, the policy's matchConditions fail because `params.strictMode` is false, so the policy does not apply to experiments/\*\*.

## Future considerations

Kubernetes ValidatingAdmissionPolicy includes additional matching properties that ARCI may adopt in future versions:

### resourceNames

Allowlist specific resource names within matched types. In ARCI's context, this could allow matching specific tool invocations by name or specific file names (not just patterns).

```yaml
# Hypothetical future syntax
matchConstraints:
  tools:
    - Bash
  toolNames: # Only these specific bash commands
    - "npm test"
    - "npm run build"
```

### Scope

Constrain to different scopes. In ARCI's context, this could distinguish session-scoped vs project-scoped operations.

```yaml
# Hypothetical future syntax
matchConstraints:
  scope: session # Only match session-scoped tool calls
```

### Label selectors

Kubernetes uses `namespaceSelector` and `objectSelector` for label-based filtering. In ARCI, this could enable matching based on project metadata or file labels.

```yaml
# Hypothetical future syntax
matchConstraints:
  projectSelector:
    matchLabels:
      team: platform
      environment: production
```

ARCI adds these extensions only if real-world usage patterns demonstrate clear need.
