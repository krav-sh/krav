# Reviewing work

## What

The developer says "review the parser" or "review what the agent built for this module" or "do a code review of the last set of changes." The agent examines deliverables against their source requirements, checks for problems, and produces a structured review report and zero or more defects.

## Why

Reviews are how the graph's verification model connects to actual code quality. A review task doesn't just check that code compiles; it evaluates whether the code satisfies the requirements it targets. The defects it produces carry types and categories, feeding into process analytics. A pattern of "ambiguous" defects in requirements tells you the derivation step needs work. A pattern of "incorrect" defects tells you the coding step does.

Reviews are also the primary mechanism for gatekeeping phase advancement. A module can't advance from coding to integration if a review task has an unacceptable disposition or if blocking defects remain open.

## What happens in the graph

The review starts as a TASK-* node with `taskType: code-review` (or `security-audit`, `architecture-review`, `design-review` depending on the target). The agent loads the task's context, which includes the module's requirements, the deliverables from predecessor tasks, and any existing defects.

The agent reads the code (or design doc, or API spec, whatever the review target is) and evaluates it against each relevant requirement. For each problem found, it creates a DEF-* node with `subject` pointing at the defective item (requirement, task, or module), `category` describing the defect type (missing, incorrect, ambiguous, inconsistent, non-verifiable, non-traceable, incomplete, superfluous, non-conformant, regression), and `severity` (critical, major, minor, editorial).

The review task's deliverables include a review report document and the list of defects created. The task disposition gets recorded: accepted (no blocking defects), conditionally accepted (minor defects that don't block advancement), or not accepted (blocking defects found).

## Trigger patterns

`Review the parser`, `do a code review`, `check the code against requirements`, `review TASK-X deliverables`, `architecture review for module Y`.

## Graph before

A module with completed coding (or design, or architecture) tasks and their deliverables. Requirements for the module exist and stakeholders have approved them.

## Graph after

A completed review TASK-* with disposition. Zero or more DEF-* nodes linked to the review task via `detectedBy` and to defective items via `subject`. A review report document as a deliverable.

## Agent interaction layer

### Skills

The `arci:review` skill builds this workflow. It runs inside the `arci-reviewer` subagent rather than the main agent context, so its preprocessing and instructions target an evaluator starting with a clean context window. Preprocessing loads the module's requirements and the deliverable file list from predecessor tasks via `!`command`` directives, giving the reviewer everything it needs to assess whether deliverables satisfy their requirements.

The skill's instructed commands create DEF-* nodes via `arci defect create` for each problem found, and produce a structured assessment covering each requirement's satisfaction status. The skill body sets the evaluative posture: check deliverables against requirements, not against the reviewer's aesthetic preferences.

### Agents

The `arci-reviewer` subagent is the canonical execution context for this workflow. It starts with a fresh context window, eliminating the bias that comes from reviewing your own code. The reviewer has read-only tool access (Read, Grep, Glob, and Bash for `arci` CLI commands) so it can read any project file and create defect records but can't modify source code.

The delegation message from the main agent handles different review types (code review, design review, architecture review, security audit), not separate subagents. The delegation message specifies the review scope, target module, and focus areas. The reviewer's system prompt provides the critical evaluation posture that stays consistent across review types.

### Policies

The `subagent-context` policy fires on SubagentStart when the reviewer launches, injecting the target module's requirements, domain context, and deliverable files as `additionalContext`. This complements the skill's preprocessing by adding live graph state that may have changed since the skill loaded.

The `review-completion-gate` policy fires on SubagentStop and is the workflow's primary enforcement mechanism. It ensures the reviewer produces required deliverables before completing: a structured assessment of each requirement's satisfaction and any defect records. Without this gate, a reviewer could stop early without recording findings.

The `graph-integrity` and `cli-auto-approve` policies apply within the subagent's execution context, ensuring defect creation goes through the CLI without permission prompts.

### Task types

This workflow executes `review-code`, `review-design`, `review-architecture`, and `audit-security` task types. The deliverable expectations are identical across review types (review report, defect records, per-requirement assessment), with the review scope and evaluative focus varying by type. The `arci:advance` skill also creates review tasks as preconditions for phase advancement, so review tasks may originate from decomposition or from advancement checks.

## Open questions

**Should reviews run in a subagent?** Reviews need a fresh, focused perspective. The main agent that wrote the code would review its own work, which defeats the purpose. A review subagent with a system prompt emphasizing critical evaluation (rather than the coding-oriented main agent) might produce better results. But Claude Code's subagent model has constraints: the subagent has a limited context window and can't access the main conversation's history.

**How does the agent access the code to review?** For code reviews, the agent needs to read the files that changed. The deliverables array on predecessor tasks includes file paths. But should the agent read every changed file, or only files relevant to the requirements it's checking against? For large coding efforts, reading everything might exceed context limits.

**Severity calibration.** How does the agent decide between critical, major, minor, and editorial? The design defines these by impact on the system, but in practice the distinction is subjective. "This function doesn't handle null input": is that critical (crashes the system), major (data corruption), or minor (user sees an ugly error)? The answer depends on context the agent might not have.

**Self-review vs. independent review.** If the agent writes code and then reviews it in the same session, the review is performative: the agent won't find problems it didn't notice while coding. The design says tasks execute in "atomic sessions," which suggests the review should be a separate session. But that requires the review subagent to understand code it didn't write.

**What constitutes a review report?** The design mentions review reports as deliverables but doesn't specify format or content. Should it be a structured document (sections per requirement, pass/fail judgments, evidence)? A narrative assessment? A checklist? This needs specification.

**Human review integration.** Sometimes the developer reviews code themselves and wants to record findings in the graph. The agent should support both agent-driven and human-driven reviews, with the human review case being the developer dictating findings that the agent records as DEF-* nodes.
