# Task and decomposition templates

## Overview

Templates provide reusable patterns for common workflows. Rather than manually constructing task DAGs for every feature, fix, or review, templates encode proven patterns that users instantiate with project-specific context.

Two template types serve different scopes:

- **Task templates**: single-task patterns (a code review, an architecture exploration, a test build)
- **Decomposition templates**: Multi-task DAG patterns (full feature lifecycle, RFC process, quick fix)

Templates are extendable. ARCI ships with built-in templates for common patterns, while projects define their own in `.arci/templates/`. Project templates override built-ins with the same id.

## Template resolution

Templates resolve in order:

1. `.arci/templates/tasks/` and `.arci/templates/decompositions/` (project-specific)
2. Built-in templates bundled with ARCI

A project template with `id: full-feature` shadows the built-in `full-feature` template.

## Task templates

Task templates define single-task patterns. They specify the phase, task type, expected deliverables, and can include instructions that interpolate context.

```yaml
# .arci/templates/tasks/quick-review.yaml
id: quick-review
name: Quick code review
description: Single-pass review of implementation against requirements
phase: verification
task_type: code-review

parameters:
  - name: focus
    description: What aspect to focus on
    default: correctness
    enum: [correctness, security, performance]

title: "Review {{ module.title }}"

instructions: |
  Review the implementation of {{ module.id }} ({{ module.title }}).

  Focus: {{ params.focus }}

  ## Requirements to verify

  {% for req in module.requirements %}
  - {{ req.id }}: {{ req.statement }}
  {% endfor %}

  ## Review checklist

  - [ ] Implementation satisfies all requirements
  - [ ] Error handling is appropriate
  - [ ] Code follows project conventions
  {% if params.focus == "security" %}
  - [ ] Input validation is present
  - [ ] No sensitive data exposure
  {% endif %}

deliverables:
  - kind: findings
  - kind: document
    type: review-report
    path: "modules/{{ module.id }}/reviews/{{ task.id }}.md"
```

### Task template fields

| Field        | Required | Description                                                                                 |
|--------------|----------|---------------------------------------------------------------------------------------------|
| id           | yes      | Unique identifier for the template                                                          |
| name         | yes      | Human-readable name                                                                         |
| description  | no       | Longer description of what this template does                                               |
| phase        | yes      | Process phase (architecture, design, `implementation`, integration, verification, validation) |
| task_type    | yes      | Task type within the phase                                                                  |
| parameters   | no       | Declared parameters with defaults and validation                                            |
| title        | no       | Task title template (interpolated)                                                          |
| instructions | no       | Task instructions template (interpolated)                                                   |
| deliverables | no       | Expected deliverable structure                                                              |

## Decomposition templates

Decomposition templates define task DAG patterns. They compose multiple tasks (potentially referencing task templates) with dependency relationships.

```yaml
# .arci/templates/decompositions/full-feature.yaml
id: full-feature
name: Full phased feature
description: Complete feature lifecycle from architecture through validation

parameters:
  - name: skip_architecture
    description: Skip architecture phase for well-understood features
    default: false
  - name: reviewer
    description: Who should review
  - name: parallel_testing
    description: Run tests in parallel with review
    default: true

tasks:
  - id: arch
    template: architecture-exploration
    phase: architecture
    skip_if: "{{ params.skip_architecture }}"
    title: "Architecture exploration for {{ module.title }}"

  - id: design
    template: api-design
    phase: design
    depends_on: [arch]

  - id: impl
    template: feature-implementation
    phase: implementation
    depends_on: [design]
    params:
      priority: "{{ params.priority | default: 'normal' }}"

  - id: test
    template: test-implementation
    phase: verification
    depends_on: [impl]

  - id: review
    template: quick-review
    phase: verification
    depends_on: [impl]
    parallel_with: "{{ params.parallel_testing ? ['test'] : [] }}"
    params:
      focus: "{{ params.review_focus | default: 'correctness' }}"
      reviewer: "{{ params.reviewer }}"

  - id: validate
    template: stakeholder-demo
    phase: validation
    depends_on: [test, review]
```

### Decomposition template fields

| Field       | Required | Description                            |
|-------------|----------|----------------------------------------|
| id          | yes      | Unique identifier                      |
| name        | yes      | Human-readable name                    |
| description | no       | What this decomposition pattern is for |
| parameters  | no       | Declared parameters                    |
| tasks       | yes      | Array of task definitions              |

### Task definition fields (within decomposition)

| Field         | Required           | Description                                                      |
|---------------|--------------------|------------------------------------------------------------------|
| id            | yes                | Local identifier within this decomposition                       |
| template      | no                 | Task template to use (mutually exclusive with inline definition) |
| phase         | yes                | Process phase                                                    |
| task_type     | yes if no template | Task type                                                        |
| depends_on    | no                 | Array of local task ids this depends on                          |
| parallel_with | no                 | Array of local task ids this can run parallel to                 |
| skip_if       | no                 | Condition to skip this task                                      |
| title         | no                 | Override template title                                          |
| instructions  | no                 | Override or extend template instructions                         |
| params        | no                 | Parameters to pass to the task template                          |

## Context object

Templates have access to a rich context object during interpolation:

```yaml
module:                          # The target module
  id: MOD-A4F8R2X1
  title: Parser
  description: "..."
  phase: design
  parent: MOD-OAPSROOT

  needs:                         # Needs owned by this module
    - id: NEED-B7G3M9K2
      statement: "..."
      status: validated

  requirements:                  # Requirements owned by this module
    - id: REQ-C2H6N4P8
      statement: "..."
      status: approved
      priority: must

  children:                      # Child modules
    - id: MOD-X1Y2Z3A4
      title: Lexer

  findings:                      # Open findings for this module
    - id: DEF-F1L4T7W5
      type: issue
      status: open

task:                            # Current task (when in task template context)
  id: TASK-E3K8S6V2
  phase: implementation

graph:                           # Graph query helpers
  ancestors: [...]               # Full ancestor chain
  descendants: [...]             # Full descendant tree
  parent: {...}                  # Parent module
  siblings: [...]                # Sibling modules

project:                         # Project-level info
  root_module: MOD-OAPSROOT
  name: "arci"

params:                          # Caller-provided parameters
  skip_architecture: false
  reviewer: "@tony"
  priority: "high"
  custom_param: "any value"
```

### Graph query helpers

The `graph` object provides query helpers for templates:

```yaml
# Get ancestors of a specific module
{% for ancestor in graph.ancestors(module.id) %}
  {{ ancestor.title }}
{% endfor %}

# Get all requirements that flow down to this module
{% for req in graph.allocated_requirements(module.id) %}
  {{ req.statement }}
{% endfor %}

# Find the derivation chain for a requirement
{% for node in graph.path(some_concept_id, some_req_id) %}
  {{ node.id }} ->
{% endfor %}
```

## Parameter handling

### Declared parameters

Templates can declare parameters for documentation and optional validation:

```yaml
parameters:
  - name: reviewer
    description: Who should review this work
    required: true

  - name: priority
    description: Priority level
    default: normal
    enum: [low, normal, high, critical]

  - name: latency_budget
    description: Maximum allowed latency in ms
    type: number

  - name: tags
    description: Additional tags
    type: array
```

Declared parameters appear in `arci template show <id>` output and the system validates them at invocation time.

### Arbitrary parameters

Beyond declared parameters, callers can pass any parameter. Undeclared parameters are available in `params.*` without validation:

```bash
arci module decompose MOD-A4F8R2X1 \
  --template full-feature \
  --param reviewer=@tony \
  --param priority=high \
  --param custom_deadline="2026-01-15" \
  --param notify_slack_channel="#releases"
```

All of these are available as `params.reviewer`, `params.priority`, `params.custom_deadline`, `params.notify_slack_channel`.

### Parameter files

For complex invocations, parameters can come from a file:

```bash
arci module decompose MOD-A4F8R2X1 \
  --template full-feature \
  --params-file ./params/parser-feature.yaml
```

```yaml
# ./params/parser-feature.yaml
reviewer: "@tony"
priority: high
skip_architecture: true
stakeholders:
  - "@alice"
  - "@bob"
custom:
  deadline: "2026-01-15"
  tracking_issue: "https://github.com/org/repo/issues/42"
```

## Templating syntax

Templates use Jinja2 syntax for interpolation, including its inheritance system for composing and extending templates.

### Template inheritance

Base templates define a structure with blocks that child templates can override:

```yaml
# .arci/templates/decompositions/_base-feature.yaml
id: _base-feature
abstract: true

tasks:
  - id: arch
    template: architecture-exploration
    phase: architecture
    {% block arch_config %}{% endblock %}

  - id: design
    template: api-design
    phase: design
    depends_on: [arch]
    {% block design_config %}{% endblock %}

  - id: impl
    template: feature-implementation
    phase: implementation
    depends_on: [design]
    {% block impl_config %}{% endblock %}

  {% block verification_tasks %}
  - id: test
    template: test-implementation
    phase: verification
    depends_on: [impl]

  - id: review
    template: quick-review
    phase: verification
    depends_on: [impl]
  {% endblock %}

  {% block validation_tasks %}
  - id: validate
    template: stakeholder-demo
    phase: validation
    depends_on: [test, review]
  {% endblock %}
```

Child templates extend the base and override specific blocks:

```yaml
# .arci/templates/decompositions/security-feature.yaml
{% extends "_base-feature.yaml" %}

id: security-feature
name: Security-critical feature
description: Feature lifecycle with enhanced security review

{% block verification_tasks %}
  - id: test
    template: test-implementation
    phase: verification
    depends_on: [impl]

  - id: security-review
    template: security-audit
    phase: verification
    depends_on: [impl]
    params:
      depth: thorough

  - id: code-review
    template: thorough-review
    phase: verification
    depends_on: [impl]
    params:
      focus: security
{% endblock %}
```

```yaml
# .arci/templates/decompositions/internal-feature.yaml
{% extends "_base-feature.yaml" %}

id: internal-feature
name: Internal tooling feature
description: Lighter-weight process for internal tools

{% block validation_tasks %}
  {# Skip formal validation for internal tooling #}
{% endblock %}
```

### Block operations

Beyond full replacement, blocks support `super()` to include parent content:

```yaml
{% block verification_tasks %}
{{ super() }}
  - id: compliance-check
    template: compliance-check
    phase: verification
    depends_on: [impl]
{% endblock %}
```

This adds a compliance check while preserving all parent verification tasks.

### Variable interpolation

```yaml
title: "Implement {{ module.title }}"
```

### Conditionals

```yaml
instructions: |
  {% if params.priority == "critical" %}
  ⚠️ This is critical priority. Escalate blockers immediately.
  {% endif %}
```

### Loops

```yaml
instructions: |
  ## Requirements

  {% for req in module.requirements %}
  - {{ req.id }}: {{ req.statement }}
  {% endfor %}
```

### Filters

```yaml
path: "src/{{ module.title | slugify }}/index.ts"
reviewer: "{{ params.reviewer | default: 'unassigned' }}"
```

### Expressions in skip_if

```yaml
skip_if: "{{ params.skip_architecture }}"
skip_if: "{{ module.phase != 'architecture' }}"
skip_if: "{{ module.requirements | length == 0 }}"
```

## Built-in templates

ARCI ships with built-in templates for common patterns:

### Task templates

| Id                       | Phase          | Description                                 |
|--------------------------|----------------|---------------------------------------------|
| architecture-exploration | architecture   | Explore architectural options for an module |
| module-decomposition     | architecture   | Identify child modules and boundaries      |
| interface-identification | architecture   | Define interaction points between modules  |
| api-design               | design         | Design API contracts and data types         |
| data-model               | design         | Design data schemas and relationships       |
| `feature-implementation`   | `implementation` | Build a feature                         |
| `test-implementation`      | verification   | Write tests for requirements                |
| quick-review             | verification   | Single-pass code review                     |
| thorough-review          | verification   | Multi-aspect deep review                    |
| stakeholder-demo         | validation     | Demo to stakeholders for sign-off           |

### Decomposition templates

| Id           | Description                                                                            |
|--------------|----------------------------------------------------------------------------------------|
| full-feature | Complete lifecycle: architecture, design, coding, verification, validation |
| quick-fix    | Fast path: coding, then quick review                                               |
| rfc          | RFC process: exploration → formalization → stakeholder review                          |
| tech-debt    | Technical debt: analysis, design, coding, verification                      |
| spike        | Time-boxed exploration with findings                                                   |

## CLI commands

```bash
# List all available templates
arci template list
arci template list --type task
arci template list --type decomposition

# Show template details
arci template show full-feature
arci template show quick-review --format yaml

# Create task from template
arci task create --template quick-review --module MOD-A4F8R2X1
arci task create --template quick-review --module MOD-A4F8R2X1 --param focus=security

# Decompose using template
arci module decompose MOD-A4F8R2X1 --template full-feature
arci module decompose MOD-A4F8R2X1 --template full-feature --param skip_architecture=true
arci module decompose MOD-A4F8R2X1 --template quick-fix --param finding=DEF-X1Y2Z3A4

# Validate a template
arci template validate .arci/templates/tasks/my-custom-review.yaml

# Preview decomposition without creating tasks
arci module decompose MOD-A4F8R2X1 --template full-feature --dry-run
```

## Project template structure

```text
.arci/
  templates/
    tasks/
      custom-review.yaml
      security-audit.yaml
    decompositions/
      our-release-process.yaml
      incident-response.yaml
```

## Implementation status

| Component | Status | Notes |
|-----------|--------|-------|
| Template system | ○ Not yet | Template loading, resolution, and rendering not implemented |
| Task templates | ○ Not yet | No template storage or CLI commands |
| Decomposition templates | ○ Not yet | No decomposition template support |
| Built-in templates | ○ Not yet | None of the listed built-in templates exist |
| CLI commands | ○ Not yet | No `arci template` command group |

This document describes the target architecture for the templating system. The team plans to build this but has not yet started.

## Summary

The templating system provides reusable patterns for task and decomposition workflows:

- **Task templates** define single-task patterns with phase, type, instructions, and deliverables
- **Decomposition templates** define multi-task DAG patterns with dependencies and conditional logic
- **Template inheritance** via Jinja2 extends/block allows composing workflows without duplication
- **Rich context** gives templates access to module data, graph queries, and project info
- **Arbitrary parameters** allow callers to pass any data for interpolation
- **Extendable resolution** lets project templates override built-ins

Templates encode proven patterns while remaining flexible through inheritance, parameterization, and context interpolation.
