# Agents

This directory contains design specifications for the subagents that Krav ships as part of its Claude Code plugin. Subagents provide isolated execution contexts for workflows that benefit from a fresh context window, restricted tool access, or a different evaluative posture.

Agents live in the plugin's `claude/agents/` directory. Each is a markdown file with YAML frontmatter (tool restrictions, skill preloads, hooks, model overrides, permission mode) and a markdown body that serves as the subagent's system prompt. The delegating agent's message provides the specific task, and preloaded skills inject full workflow instructions (including preprocessed graph context) at startup.

## Shipped agents

| Agent | Primary workflows | What it does |
|-------|-------------------|-------------|
| `krav-reviewer` | [Reviewing work](../reviewing-work.md) | Evaluates deliverables against requirements and produces defects |
| `krav-verifier` | [Running verification](../running-verification.md) | Executes test cases and records results |
| `krav-analyst` | [Comparing baselines](../comparing-baselines.md), [Handling suspect links](../handling-suspect-links.md), [Tracing requirements](../tracing-requirements.md) | Read-only investigation and analysis of graph state |

### `krav-reviewer`

The review agent is the canonical case for subagent isolation. The agent that wrote the code shouldn't review it: it carries coding context that biases it toward confirming its own decisions. The reviewer starts with a clean context window, loads the module's requirements and deliverable files, and evaluates independently.

Tool access allows only reads: Read, Grep, Glob, and Bash (for running `krav` CLI commands to query the graph and create defects). No Write or Edit access. The reviewer can read any file in the project but can't modify anything except the graph (via CLI commands that create defect nodes).

Preloaded skills: `krav:review`. The review skill's !`command` preprocessing injects the module's requirements and deliverable list before the reviewer sees its instructions. The skill's instructed commands tell the reviewer to run `krav defect create` for each problem found.

The reviewer's system prompt sets the posture: thorough, independent, skeptical of design choices, focused on whether deliverables satisfy requirements rather than whether the code is "good" in the abstract. The delegation message handles different review types (code review, design review, architecture review) rather than separate agents. The delegation message specifies what kind of review, which module, and what to focus on.

A SubagentStop hook (defined in the agent's frontmatter) ensures the reviewer produces required deliverables before completing: at minimum, a structured assessment of each requirement's satisfaction and any defect records.

### `krav-verifier`

The verification agent executes test cases in isolation from coding context. The agent that wrote the tests might unconsciously shape execution to match expectations. The verifier starts fresh, loads the test case specifications and their implementations, runs them, and records results.

Tool access: Read, Grep, Glob, and Bash. Bash access is broader than the reviewer's because the verifier needs to execute test suites, but Write and Edit are still excluded. The verifier records results through `krav` CLI commands, not by editing test files.

Preloaded skills: `krav:verify`. Preprocessing injects the module's test cases, their current results, and the verification method for each. The skill instructions guide the verifier through executing each test case and recording pass/fail/skip results with evidence.

The verification agent's system prompt emphasizes faithful execution: run the tests as specified, report results accurately, don't fix failing tests. If a test case specification is ambiguous, the verifier flags it as a defect rather than interpreting it charitably.

### `krav-analyst`

The analysis agent handles read-only investigation tasks that benefit from a dedicated context window: baseline comparisons, suspect link triage, and deep traceability queries. These tasks can involve loading large amounts of graph data and walking complex chains of relationships, so a fresh context window avoids competing with ongoing coding work.

Tool access: Read, Grep, Glob, and Bash (read-only `krav` CLI commands). No Write or Edit.

Preloaded skills vary by task. For baseline comparisons: `krav:diff`. For suspect link triage: `krav:suspect`. For traceability queries: `krav:trace`. The delegating agent specifies the task and the appropriate skill is preloaded based on the task type.

The analyst's system prompt emphasizes precision and completeness: walk the full chain, don't skip nodes, report what changed and what it affects. For suspect link triage, the analyst recommends a disposition (clear, create defect, update downstream) but doesn't act on it; the main agent or developer makes the final call.

## Design considerations

The three-agent model covers the isolation needs identified across the 21 workflows. Specialization beyond this (separate agents for architecture review vs. code review, or for security audit) works through delegation messages and skill preloads rather than separate agent definitions. If a review type proves different enough in tool needs or posture to warrant its own agent, the team can split it out later.

All three agents use `permissionMode: auto` to avoid interactive permission prompts during subagent execution. Their restricted tool access provides the safety boundary instead.

Subagent `hooks` in frontmatter define SubagentStop hooks (converted from Stop hooks) that enforce completion criteria. The specific criteria vary by agent: the reviewer must produce defect records, the verifier must record results for every test case, the analyst must produce a structured report.

The `agentId` field on each agent enables resumption for long-running analysis tasks. A baseline comparison across a large graph might require suspending and resuming. Whether this is practically useful depends on how often analysis tasks exceed a single context window.

## Open questions

Whether the analyst agent should be a single agent with multiple skill preloads or split into separate agents per analysis type. The current design uses one agent because the tool access and posture are identical across analysis tasks, but the skill preload mechanism means the agent's effective instructions change per invocation anyway.

Whether review subagents should have access to the coding agent's transcript for context, or whether isolation should be total. The SubagentStart hook's `additionalContext` could inject a summary of what the coding agent built without providing the full transcript, but this is a spectrum between "fully independent review" and "review informed by intent."

Whether the verifier needs Write access to generate test fixtures or test data files. The current design says no (use Bash to create temporary files), but complex verification scenarios might push against this.
