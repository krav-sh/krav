# Policies

This directory contains design specifications for the hook policies that Krav ships as part of its Claude Code plugin. These policies intercept Claude Code lifecycle events to inject graph context, enforce process integrity, and steer the agent toward correct behavior.

Krav registers policies through its plugin hook configuration and evaluates them via Krav's policy engine (for declarative YAML/CEL policies) or as command, prompt, or agent hooks that call Krav CLI commands (for context injection and multi-turn evaluation). Both kinds ship with Krav and the docs here describe them.

The policies fall into groups by function: context injection (getting graph state into the agent's context), enforcement (preventing incorrect behavior), and workflow support (reducing friction for correct behavior).

## Context injection policies

These policies push graph state into the agent's context at the right moments. Without them, the agent would have to explicitly query the graph before every action. Context injection makes graph awareness ambient.

| Policy                                      | Hook event       | What it does                                                                                                                                                                                                                                                                                                                                                                                    |
| ------------------------------------------- | ---------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| session-context       | SessionStart     | Injects graph state at session start: current task context, ready task list, recent defects, suspect links pending review. Varies by `source` matcher (`startup` gets full context, `resume` gets continuation context, `compact` gets condensed summary). Uses session-scoped state store to track what it already injected.                                                                       |
| prompt-context         | UserPromptSubmit | Parses the developer's prompt for module names, task IDs, and other graph references, then injects relevant context as `additionalContext`. If the developer mentions MOD-3, the hook injects MOD-3's current phase, blocking defects, and active tasks.                                                                                                                                        |
| subagent-context     | SubagentStart    | Injects graph context when review, verification, or analysis subagents launch. Matches on `agent_type` for Krav agents. Injects the target module's requirements, domain context, and relevant deliverables into `additionalContext`. Complements the subagent's preloaded skills with live graph state.                                                                                        |
| compaction-context | PreCompact       | Injects a condensed graph state summary before context compaction. Ensures the post-compaction context retains awareness of the current task, module phase, blocking defects, and pending suspect links. Without this, compaction could lose the graph context that session-context originally injected.                                                                                        |
| mutation-feedback   | PostToolUse      | Fires after `krav` CLI commands that mutate the graph (task updates, defect creation, status transitions, deliverable recording). Injects updated graph state as `additionalContext`: "deliverable recorded, 2 of 4 requirements now have deliverables" or "defect created, module has 2 blocking defects preventing phase advancement." Matches on Bash tool calls containing `krav` commands. |

## Enforcement policies

These policies prevent the agent from skipping process steps, modifying protected content, or completing work prematurely. When they block an action, the denial output includes specific CLI commands the agent should run instead.

| Policy                                                | Hook event    | What it does                                                                                                                                                                                                                                                                                                                                          |
| ----------------------------------------------------- | ------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| baseline-protection         | PreToolUse    | Denies Write and Edit tool calls targeting files that belong to baselined modules. The denial message tells the agent to create a defect or run `krav moduleunlock` with justification. Uses escalating enforcement via session-scoped state: warns on first attempt, blocks on subsequent attempts.                                                  |
| task-completion-gate       | TaskCompleted | Checks whether the agent has recorded deliverables on the TASK-\* node, made required graph mutations, and recorded verification results before allowing task completion. Blocks with a specific list of what's missing. The system can run this as a command hook calling `krav taskvalidate` or as an agent hook for multi-turn investigation of complex tasks. |
| session-completion-gate | Stop          | Checks whether the agent was working on a task and left it in an incomplete state. If the current task has unrecorded deliverables or status hasn't changed, the hook blocks the stop and tells the agent what to finish. Uses `last_assistant_message` to avoid re-parsing the transcript. The `stop_hook_active` flag prevents infinite loops.          |
| review-completion-gate   | SubagentStop  | Ensures review and verification subagents produce required deliverables before completing. For reviewers: a structured assessment of each requirement's satisfaction and any defect records. For verifiers: recorded results for every test case in scope. Matches on `agent_type` for Krav subagents.                                                |
| phase-gate-defense           | PreToolUse    | Denies `krav moduleadvance` Bash commands when phase advancement preconditions aren't met. Defense in depth alongside the CLI's own precondition checks. The denial message reports what's blocking: open defects, incomplete tasks, insufficient verification coverage.                                                                              |
| graph-integrity                 | PreToolUse    | Prevents direct modification of `.krav/` files via Write or Edit tools. All graph mutations must go through the `krav` CLI to maintain referential integrity, provenance tracking, and suspect link propagation. The denial message tells the agent which CLI command to use instead.                                                                 |
|                                                       |               |                                                                                                                                                                                                                                                                                                                                                       |

## Workflow support policies

These policies reduce friction for correct behavior without enforcing hard rules.

| Policy | Hook event | What it does |
|--------|-----------|-------------|
| CLI-auto-approve | PermissionRequest | Auto-approves `krav` CLI commands to eliminate permission prompts for graph operations. The agent shouldn't have to ask permission to record a deliverable or check task status. Matches Bash tool calls where the command starts with `krav`. |
| failure-context | PostToolUseFailure | Catches failed `krav` CLI commands and injects context explaining what went wrong. If `krav taskcomplete` fails because preconditions aren't met, the hook injects the specific precondition violations rather than letting the agent guess from a generic error message. |
| teammate-quality-gate | TeammateIdle | Prevents agent team teammates from going idle until their assigned tasks are properly closed and graph state reflects the work. Sends stderr feedback telling the teammate what's still open. Only relevant for multi-agent team scenarios. |

## Policy composition

Multiple policies fire on the same events and need to compose correctly. PreToolUse is the busiest event: baseline-protection, phase-gate-defense, graph-integrity, and CLI-auto-approve all match on it. Krav's deny-wins aggregation handles the precedence: if any policy denies, the action fails regardless of other policies allowing it. CLI-auto-approve only fires on PermissionRequest (a different event), so it doesn't conflict with PreToolUse denials.

PostToolUse carries both mutation-feedback (always fires after Krav commands) and any async hooks running verification in the background. These compose naturally since mutation-feedback injects context and async hooks run without blocking.

The Stop and TaskCompleted events both serve completion enforcement but fire at different times. TaskCompleted fires when a task closes; Stop fires whenever Claude stops talking. session-completion-gate is the catch-all for when the agent stops without properly closing its task, while task-completion-gate is the precise check for task closure quality.

## Open questions

The right handler type for task-completion-gate and session-completion-gate is an open design question. Command hooks are fast and deterministic (check whether `krav taskvalidate` passes) but can only verify structural completion (deliverables recorded, status updated). Prompt hooks can evaluate whether the work is actually done. Agent hooks can investigate the codebase (do the deliverable files exist? do the tests pass?). The right choice depends on how reliably the agent records deliverables, which the team won't know until coding begins.

How aggressively prompt-context should parse developer messages for graph references is a design question. Conservative matching (only explicit IDs like TASK-42 or MOD-3) is safe but misses natural language references like "the parser module." Fuzzy matching against module and task names is more helpful but risks injecting irrelevant context. This might be a configurable sensitivity level.

Whether baseline-protection should use escalating enforcement (warn, warn, block) or immediate blocking is a ceremony calibration question. Escalating enforcement is friendlier but allows one accidental write to go through before blocking. For baselined content, even one unauthorized write could be a problem. The policy might need a configuration option.

Whether mutation-feedback should fire on every Krav CLI command or only on mutations is a performance question. Querying graph state after every `krav status` call adds latency for no benefit. The policy should match on mutation commands (create, update, delete, advance, complete) and skip read-only commands (status, list, trace, coverage).
