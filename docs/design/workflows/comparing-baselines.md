# Comparing baselines

## What

The developer says `what changed since the architecture baseline?`, `diff against the last release`, `compare BSL-X to BSL-Y`, or `how has the graph evolved?` The agent reconstructs historical graph states from the baseline commits and produces a semantic diff showing what changed between them.

## Why

Baseline diffs are the project's changelog at the specification level. Git diffs show file changes. Baseline diffs show specification changes: requirements the team added, modified, or removed; test coverage gained or lost; defects opened and closed; modules advancing through phases. This is higher-fidelity than a commit log because it's structured: you can ask "how many requirements got added?" rather than grepping through prose diffs.

Baseline diffs also support audit and compliance. For projects that need to demonstrate change control (regulated industries, enterprise environments), baseline comparison provides evidence of what changed, when, and why.

## What happens in the graph

Nothing in the current graph. The agent reconstructs historical graph state from the baseline's `commitSha` (by checking out that commit and reading the graph from it) and compares it to either the current state or another baseline's state.

The diff is semantic, not textual. Rather than showing line changes in `graph.jsonlt`, it shows: nodes added, nodes modified (with field-level changes), nodes removed, relationships added/removed, phase changes, coverage delta (verification coverage at baseline A vs. baseline B), defect status changes.

Key CLI command: `arci baseline diff BSL-A BSL-B`.

## Trigger patterns

`What changed since the last baseline?`, `diff BSL-X and BSL-Y`, `how has the project evolved?`, `show me the changelog since release X`.

## Graph before

Two baselines exist (or one baseline and the current state).

## Graph after

Unchanged.

## Agent interaction layer

### Skills

The `arci:diff` skill builds this workflow. Preprocessing loads the two baselines' metadata (commit anchors, module scope, captured statistics) so the agent can present what it compares before running the actual diff. The skill's instructed command is `arci baseline diff BSL-A BSL-B`, which handles the heavy lifting of graph reconstruction and semantic comparison. The skill body tells the agent how to present the structured diff output: summarize the highlights, group changes by category, and offer to drill into specific areas.

The skill runs inside the `arci-analyst` subagent for larger comparisons, since reconstructing and comparing two full graph states can consume substantial context. For small comparisons (a module-scoped baseline diff with a handful of changes), the main agent can handle it directly.

### Agents

The `arci-analyst` subagent is the preferred execution context for non-trivial baseline comparisons. A release baseline diff across a large project might involve hundreds of node changes, which a dedicated context window handles better than an ongoing coding session. The analyst's system prompt emphasizes precision and completeness: walk the full diff, don't skip changes, report what changed and what it affects.

The `agentId` field on the analyst enables resumption for large comparisons that might exceed a single context window, though whether this is practically necessary depends on project size.

### Policies

The `subagent-context` policy fires on SubagentStart when the analyst launches for a baseline comparison, injecting the baseline metadata and module scope as `additionalContext`. The `prompt-context` policy fires when the developer mentions specific baselines or modules in follow-up questions.

No enforcement or mutation policies apply since baseline comparison only reads data. The `review-completion-gate` policy (which fires on SubagentStop for ARCI subagents) ensures the analyst produces a structured comparison report before completing.

### Task types

This workflow doesn't create or execute tasks. It produces analysis that informs decisions about what work the project needs, particularly in identifying regressions or scope changes between releases.

## Open questions

**Diff presentation.** Semantic diffs can be large. A release baseline diff for a project with hundreds of nodes might have dozens of changes. How should the agent present this? A summary with detail views? A structured report document? Grouped by change type (added, modified, removed) or by node type (requirements, tasks, defects)?

**Graph reconstruction performance.** Reconstructing historical graph state means checking out the baseline's commit and loading the graph from it. For large projects, this might be slow. Should the CLI cache reconstructed states? Should baselines store a snapshot of the graph alongside the commit reference?

**Three-way diffs.** Comparing a feature branch's graph state against the main branch and the common ancestor could show what a branch adds relative to the baseline. This is useful for review ("what requirements did this branch introduce?") but the current design does not cover git-aware graph reconstruction for this scenario.

**Granularity of "modified."** If a requirement's description changed but its statement didn't, does that count as a real modification in the diff? Probably not for most purposes. The diff should distinguish between substantive changes (statement, status, verification criteria) and cosmetic changes (description, tags) and default to showing only substantive ones.

**Diff as a deliverable.** Should baseline diffs produce a document? For audit purposes, a structured PDF or markdown report of what changed between releases would be valuable. This bridges the "comparing baselines" workflow with the "creating a baseline" workflow, where the diff becomes part of the release documentation.
