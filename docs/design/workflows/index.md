# Workflows

## Overview

This directory documents the human-initiated workflows that ARCI supports. Each workflow describes what the developer asks for, what happens in the knowledge graph, what skills and subagents participate, and what CLI commands or MCP tools execute the underlying operations.

These workflow documents serve as the design specification for the agent interaction layer, the skills, subagents, hooks, and context injection that sit between the developer's intent and the CLI/graph operations. The CLI commands and graph ontology are well-specified elsewhere in the design docs. What's missing is the connective tissue: how does a developer saying "build the parser" turn into a sequence of graph operations that produce concepts, needs, requirements, tasks, and deliverables?

## Workflow categories

Workflows fall into three broad categories based on what they do to the knowledge graph.

**Graph-building workflows** create new nodes and edges. These are the most complex because they involve formal transformations, creative decisions about decomposition, and potentially long chains of operations. Starting a project, formalizing concepts, deriving requirements, decomposing work, and creating test cases all belong here.

**Graph-changing workflows** transition existing nodes through their lifecycles, record results, and resolve problems. Working on tasks, running verification, triaging defects, advancing phases, and handling suspect links are in this category. These workflows read the graph to determine what's possible, then make targeted mutations.

**Graph-reading workflows** query the graph without changing it. Status checks, traceability queries, baseline diffs, and coverage reports are pure reads. These are the simplest from an agent perspective but still require skills to synthesize useful answers from raw graph data.

## Ceremony spectrum

Not every interaction with ARCI needs the full INCOSE transformation chain. The workflows span a spectrum from high-ceremony (starting a new project, building a feature end-to-end) to zero-ceremony (checking status, fixing a trivial bug). The agent layer must support this full range without forcing formality where it adds no value or permitting informality where traceability matters.

Where to draw that line is an open design question that cuts across all workflows. The "coding without ceremony" workflow addresses it most directly, but it affects everything.

## Reading order

The workflows appear roughly in the order they'd occur in a greenfield project, followed by maintenance workflows, then query workflows. But they're all standalone documents; read whichever is relevant.

### Graph-building workflows

| # | Workflow | What it does |
|---|----------|-------------|
| 1 | [Starting a new project](starting-a-project.md) | Bootstrap graph with root module, hierarchy, initial concepts and needs |
| 2 | [Exploring a design question](exploring-a-design-question.md) | Create and develop concepts without immediate formalization |
| 3 | [Formalizing concepts into needs](formalizing-concepts.md) | Extract stakeholder expectations from crystallized concepts |
| 4 | [Deriving requirements from needs](deriving-requirements.md) | Transform validated needs into verifiable design obligations |
| 5 | [Allocating requirements to child modules](allocating-requirements.md) | Flow down parent requirements with budgets to child modules |
| 6 | [Decomposing work into tasks](decomposing-work.md) | Generate a task DAG from requirements and module scope |
| 7 | [Creating test cases](creating-test-cases.md) | Specify and write verification for requirements |
| 8 | [Adding a module](adding-a-module.md) | Introduce a new subsystem or component to an existing project |

### Graph-changing workflows

| # | Workflow | What it does |
|---|----------|-------------|
| 9 | [Working on a task](working-on-a-task.md) | Execute a specific task in a focused session |
| 10 | [Building a feature end-to-end](building-a-feature.md) | Composite workflow from concept through verification |
| 11 | [Implementing without ceremony](implementing-without-ceremony.md) | Write code with minimal or no graph involvement |
| 12 | [Reviewing work](reviewing-work.md) | Examine deliverables against requirements, produce defects |
| 13 | [Running verification](running-verification.md) | Execute test cases and record results |
| 14 | [Triaging and fixing defects](triaging-defects.md) | Disposition, remediate, and verify defects |
| 15 | [Handling suspect links](handling-suspect-links.md) | Triage downstream impacts of upstream changes |
| 16 | [Advancing a module's phase](advancing-a-phase.md) | Check criteria and move a module to its next phase |
| 17 | [Creating a baseline](creating-a-baseline.md) | Capture graph state at a decision point |
| 18 | [Restructuring modules](restructuring-modules.md) | Reparent, split, or merge modules |

### Graph-reading workflows

| # | Workflow | What it does |
|---|----------|-------------|
| 19 | [Checking project status](checking-status.md) | Synthesize current state from graph queries |
| 20 | [Tracing requirements](tracing-requirements.md) | Walk the derivation chain to explain why things exist |
| 21 | [Comparing baselines](comparing-baselines.md) | Semantic diff between graph states |

## Agent interaction layer

The agent interaction layer sits between the developer's intent and ARCI's CLI/graph operations. It consists of composable mechanisms: skills drive workflows, subagents provide isolated execution, hooks enforce policies and inject ambient context, the state store gives hooks memory across invocations, and context injection (both preprocessed and instructed) feeds live graph data into the agent's context.

### Skills

Skills are Claude Code's extension mechanism for teaching the agent new capabilities. Each skill is a directory with a `SKILL.md` file (YAML frontmatter plus markdown instructions) and optional supporting files (reference docs, scripts, templates). Skills follow the [Agent Skills](https://agentskills.io) open standard. Claude Code extends the standard with invocation control, subagent execution, and context injection.

Skills can live in four places: enterprise (managed settings, organization-wide), personal (`~/.claude/skills/`), project (`.claude/skills/`), or bundled with a plugin. When skills share a name across levels, higher-priority locations win: enterprise > personal > project. Plugin skills use namespacing (`plugin-name:skill-name`) so they can't conflict.

Claude Code loads skill descriptions into context so the agent knows what's available, but only loads full skill content when invoked. This two-tier model keeps context usage proportional to what's actually needed. Claude invokes skills automatically when relevant to the conversation, or the developer can invoke them directly with `/skill-name`. Two frontmatter fields control this: `disable-model-invocation: true` prevents Claude from loading the skill automatically (developer-only invocation, useful for workflows with side effects like deploy or commit), and `user-invocable: false` hides the skill from the `/` menu (Claude-only invocation, useful for background knowledge that isn't a useful user action).

Key frontmatter fields for ARCI skills: `allowed-tools` restricts which tools Claude can use when the skill is active (a status-checking skill that only needs Read, Grep, Glob, and Bash declares that explicitly). `model` overrides the model while the skill is active. `context: fork` runs the skill in a forked subagent context with its own isolated conversation, using the `agent` field to specify which subagent type executes it (built-in like `Explore` or `Plan`, or a custom agent from `.claude/agents/`). `hooks` defines lifecycle-scoped hooks in the skill's frontmatter. `$ARGUMENTS` in skill content gets replaced with arguments passed at invocation (positional access via `$ARGUMENTS[N]` or `$N` shorthand).

Skills support `!`command`` syntax for preprocessing: shell commands that run before the skill content reaches Claude, with their output replacing the placeholder. This is how ARCI skills inject live graph data. A task-execution skill includes `!`arci taskcontext $ARGUMENTS`` so Claude receives the full task context (requirements, deliverables, dependencies) as part of the skill prompt without having to run any commands itself. See [Dynamic context in skills](#dynamic-context-in-skills) for the full picture.

ARCI's skills encode the graph-aware workflows documented in this directory. A formalization skill reads the concept's prose, walks through stakeholder classes, and calls `arci needcreate` for each extracted expectation. A review skill loads module requirements, reads deliverable files, and calls `arci defect create` for problems found. The skill instructions are where INCOSE method meets Claude Code tool calls.

### Subagents

Subagents provide specialized agents with their own context window, system prompt, tool access, and optionally their own model. Developers define them as markdown files in `.claude/agents/` (project) or `~/.claude/agents/` (personal), or bundle them with the ARCI plugin in `claude/agents/`.

Subagents are the right tool when a workflow benefits from a fresh context window, focused instructions, or restricted tool access. Code review is the canonical example: the reviewing agent shouldn't carry the coding context that would bias it toward confirming its own work. A review subagent gets a clean context, loads the relevant requirements and code, and evaluates independently.

Key subagent configuration fields for ARCI's purposes: `tools` restricts what the subagent can do (a review subagent gets Read, Grep, Glob, Bash but not Write or Edit). `model` controls the model (inherit from the main conversation, or use a specific model alias). `skills` preloads specified skills into the subagent's context at startup, and unlike the normal two-tier loading model where only skill descriptions are in context, preloaded skills inject their full content. This means an ARCI review subagent with `skills: [arci-review]` starts with the complete review workflow instructions already loaded, including any !`command` preprocessing output. The subagent's markdown body provides the system prompt, the delegating agent's message provides the task, and preloaded skills plus CLAUDE.md provide reference material. `permissionMode` controls whether the subagent asks for permission on tool calls. `hooks` defines lifecycle-scoped hooks in the subagent's frontmatter (the system automatically converts Stop hooks to SubagentStop).

Subagents can resume via their `agentId`, which enables multi-session workflows. A long-running analysis subagent can pause and resume with full context from its previous conversation. This partially addresses the session continuity problem for subagent-driven workflows.

Claude Code's built-in subagents (Explore for read-only codebase search, Plan for plan-mode research, and the general-purpose agent for complex multi-step tasks) are available alongside ARCI's custom subagents. ARCI's subagents should complement rather than duplicate these.

### Hooks

Hooks intercept Claude Code lifecycle events and can inject context, block operations, modify tool inputs, or enforce completion criteria. They're the mechanism for pushing graph-awareness into the agent's workflow without the agent having to explicitly ask for it. ARCI registers hooks via its Claude Code plugin, and individual skills and subagents can define their own hooks in YAML frontmatter (scoping them to the component's lifetime).

Four hook handler types are available. Command hooks (`type: "command"`) run a shell script that receives JSON on stdin. HTTP hooks (`type: "http"`) POST the JSON to a URL. Prompt hooks (`type: "prompt"`) send the context to a fast LLM for single-turn evaluation. Agent hooks (`type: "agent"`) spawn a subagent with tool access (Read, Grep, Glob, Bash) that can investigate the codebase across up to 50 turns before returning a decision. Command hooks can also run asynchronously (`async: true`), executing in the background without blocking the agent.

Not every event supports every hook type. PreToolUse, PostToolUse, PostToolUseFailure, PermissionRequest, Stop, SubagentStop, TaskCompleted, and UserPromptSubmit support all four types. The remaining events (SessionStart, SessionEnd, SubagentStart, PreCompact, Notification, TeammateIdle, ConfigChange, WorktreeCreate, WorktreeRemove) only support command hooks.

A key property of hook output: when a hook blocks or denies an action, the output text (stderr for exit code 2, `reason` or `permissionDecisionReason` in JSON, `reason` from prompt/agent hooks) gets fed back to Claude as context. This makes denials a steering mechanism, not just a gate. A PreToolUse hook that denies a write to a baselined module doesn't just say "denied"; it can say "this module has a baseline; create a defect against REQ-42 instead of editing the file directly, or run `arci moduleunlock MOD-X` if you have justification." A Stop hook that blocks completion can say "you haven't recorded deliverables yet; run `arci taskcomplete TASK-123 --deliverables src/parser.ts src/parser.test.ts` before stopping." A TaskCompleted hook can say "3 requirements in this module still lack test cases; run `arci reqcoverage --module MOD-X` to see what's missing." The denial output is where ARCI teaches the agent how to recover from a blocked action, including specific CLI commands to run next.

The hook events relevant to ARCI workflows, grouped by what they enable:

**Context injection points.** These events support injecting text into the agent's context via `additionalContext` in the JSON output or plain text on stdout.

SessionStart fires when a session begins (matchers: `startup`, `resume`, `clear`, `compact`). Command hooks only. This is the primary entry point for loading graph state: the current task context, the ready task list, recent defects, suspect links pending review. The `source` matcher distinguishes first-start from resume from post-compaction, allowing different context loads for each. SessionStart also receives the `model` identifier and, if launched with `claude --agent <name>`, an `agent_type` field. It can persist environment variables via `CLAUDE_ENV_FILE` for the session's duration.

UserPromptSubmit fires when the developer submits a prompt, before Claude processes it. The hook receives the prompt text and can inject `additionalContext` that Claude sees alongside the prompt. If the developer mentions a module name, the hook injects that module's status and domain context. The hook can also block prompts entirely via `decision: "block"`.

SubagentStart fires when the system spawns a subagent. Command hooks only. Matches on `agent_type` (built-in agents like `Bash`, `Explore`, `Plan`, or custom agent names from `.claude/agents/`). The hook receives `agent_id` and `agent_type` and can inject `additionalContext` into the subagent's context. SubagentStart cannot block subagent creation. For ARCI, this is how graph context reaches subagents: when a review subagent launches, the SubagentStart hook injects the target module's requirements, domain context, and relevant deliverables.

PreCompact fires before context compaction (matchers: `manual`, `auto`). This is where ARCI injects a structured summary of the current graph state that survives compaction. Without this, compaction might lose the graph context that SessionStart originally injected.

**Task and completion control.** These events let ARCI enforce completion criteria and prevent premature stopping.

TaskCompleted fires when an agent marks a task as completed, either by explicitly calling TaskUpdate or when an agent team teammate finishes its turn with in-progress tasks. Supports all four hook types. For command hooks, exit code 2 blocks completion and feeds stderr back to the model as feedback. For prompt and agent hooks, returning `{"ok": false, "reason": "..."}` has the same effect. This is directly useful for ARCI: a TaskCompleted hook can check whether the agent has recorded deliverables on the ARCI TASK-* node, whether required graph mutations have occurred, and whether verification results exist. If not, the hook blocks completion and tells the agent what's missing. TaskCompleted receives `task_id`, `task_subject`, and optionally `task_description`, `teammate_name`, and `team_name`. No matchers; fires on every occurrence.

Stop fires when the main agent finishes responding. Returning `decision: "block"` with a `reason` prevents the agent from stopping and tells it how to proceed. The `stop_hook_active` flag prevents infinite loops. The `last_assistant_message` field contains the text of Claude's final response, so the hook can evaluate completion without parsing the transcript. For ARCI, a Stop hook can check whether the agent was working on a task and whether it recorded deliverables and updated task status before stopping.

SubagentStop fires when a subagent finishes. Same blocking capability as Stop, scoped to a specific subagent. Matches on `agent_type`. The hook receives `agent_transcript_path` (the subagent's own transcript) and `last_assistant_message`. ARCI uses this to ensure review subagents produce all required deliverables (review report, defect records) before completing.

TeammateIdle fires when an agent team teammate is about to go idle after finishing its turn. Command hooks only. Exit code 2 prevents idling and sends stderr as feedback to the teammate, making it continue working. The hook receives `teammate_name` and `team_name`. No matchers; fires on every occurrence. For ARCI, this enforces quality gates in team scenarios: a teammate working on a module can't go idle until its tasks are properly closed and graph state reflects the work.

**Operation control.** These events let ARCI gate, modify, or audit tool calls.

PreToolUse fires before any tool executes. The hook can allow (auto-approve, bypassing the permission dialog), deny (block the tool call with a reason fed back to Claude), or ask (defer to the user). It can also modify tool inputs via `updatedInput` and inject `additionalContext`. Matchers filter by tool name: `Bash`, `Write`, `Edit`, `Read`, `Agent` (for subagent invocations), `mcp__*` (for MCP tools), etc.

For ARCI, PreToolUse is the enforcement layer. A policy that denies writes to files under `.arci/modules/` for baselined modules. A policy that denies `arci moduleadvance` when blocking defects exist (defense in depth alongside the CLI's own checks). A policy that auto-approves `arci` CLI commands to reduce permission friction. Input mutation could rewrite Bash commands to inject flags (adding `--module` context automatically).

PostToolUse fires after a tool completes successfully. The hook receives both the `tool_input` and `tool_response`. It can inject `additionalContext` or return `decision: "block"` to give Claude corrective feedback. For ARCI, PostToolUse on Bash commands matching `arci task*` or `arci defect*` can inject updated graph state after mutations: "task completed, 3 dependent tasks are now ready" or "defect created, module has 2 blocking defects preventing advancement." Async PostToolUse hooks can run verification (like `npm test`) in the background after file writes without blocking the agent.

PostToolUseFailure fires when a tool call fails. It receives `error` and `is_interrupt` fields alongside the tool input. For ARCI, this catches failed CLI commands: if `arci taskcomplete` fails because preconditions aren't met, the hook can inject context explaining what's wrong.

PermissionRequest fires when the user sees a permission dialog, allowing auto-approve or auto-deny decisions before the user has to respond. Can also apply `updatedPermissions` to set persistent "always allow" rules.

ConfigChange fires when a configuration file changes during a session (matchers: `user_settings`, `project_settings`, `local_settings`, `policy_settings`, `skills`). Command hooks only. Can block non-policy config changes via exit code 2 or `decision: "block"`. Hooks still fire for `policy_settings` sources but the system ignores blocking decisions, ensuring enterprise-managed settings always take effect. Relevant for ARCI if hook policies or skill definitions change mid-session.

WorktreeCreate and WorktreeRemove fire when git worktrees get created or removed (for `--worktree` sessions or subagents with `isolation: "worktree"`). Command hooks only. WorktreeCreate replaces the default git behavior: the hook must print the absolute path to the created worktree on stdout. WorktreeRemove handles cleanup. These are relevant if ARCI uses worktree isolation for subagent tasks that need a clean working tree, though the default git behavior may suffice initially.

**Hook handler types and their use for ARCI.**

Command hooks (`type: "command"`) are the workhorse. Most ARCI hooks are command hooks running ARCI CLI commands to check graph state, validate preconditions, or inject context. They're fast and deterministic.

HTTP hooks (`type: "http"`) POST the event JSON to a URL. Useful if ARCI's daemon exposes an HTTP endpoint for hook evaluation, avoiding the overhead of spawning a CLI process per hook invocation. Headers support environment variable interpolation via `allowedEnvVars`.

Prompt hooks (`type: "prompt"`) send context to a fast LLM for single-turn evaluation. The LLM returns `{"ok": true/false, "reason": "..."}`. Useful for Stop and SubagentStop when deterministic rules can't judge whether the agent's work is actually complete. A prompt-based Stop hook could evaluate: `Given this transcript, has the agent recorded task deliverables and updated the graph?`

Agent hooks (`type: "agent"`) spawn a subagent with tool access (Read, Grep, Glob, Bash) that can investigate the codebase across up to 50 turns before returning a decision. Same response schema as prompt hooks (`{"ok": true/false, "reason": "..."}`). More capable than prompt hooks at the cost of latency (default timeout: 60 s vs 30 s for prompts). A Stop agent hook could read the task's deliverables array, check that referenced files exist, and verify test results before allowing completion. A TaskCompleted agent hook could verify that ARCI graph mutations actually happened by inspecting the `.arci/` directory.

**Hooks in skill and agent frontmatter.** Skills and subagents can define hooks in their YAML frontmatter, scoped to the component's lifetime. A `once: true` field on skill hooks causes the hook to run only once per session then get removed. This is useful for one-time setup: a skill's SessionStart hook that loads domain context the first time the skill activates. Subagent Stop hooks defined in frontmatter are automatically converted to SubagentStop events.

### State store

Hook invocations are stateless by default: each time a hook fires, it evaluates the event in isolation. The state store (documented in [state-store.md](../state-store.md)) gives hooks persistent memory across invocations, scoped to either a session or a project.

Session-scoped state ties to a single Claude Code session and resets when a new session starts. It tracks things like how many times the agent has attempted a particular action, what context the system has already injected, and which tasks the agent has worked on so far. A resumed session (`--resume`, `--continue`) continues the same session and retains its state, but starting a new session or clearing with `/clear` begins fresh. Project-scoped state persists across sessions on the local machine: cumulative statistics, configuration flags, historical records. Neither scope goes into the repo. The project database lives at `.arci/state.db` (which is gitignored), and user-level state lives at `~/.config/arci/state.db`. State store data is local to the developer's machine; it doesn't travel with the project and isn't shared across collaborators. The store is a SQLite-backed key-value store with metadata (timestamps, author) on each entry.

Policies read state through `$session_get(key)` and `$project_get(key)` in CEL expressions, and write state through `setState` effects that run after the admission decision. This separation matters: reads happen during evaluation so they can influence the decision, writes happen afterward so they reflect what actually occurred.

The canonical use for ARCI is escalating enforcement within a session. A hook that warns on the first attempt to write to a baselined file, warns more firmly on the second, and blocks on the third needs to track the attempt count somewhere. Without session-scoped state, every attempt would look like the first. The state store also enables smarter context injection within a session: a SessionStart hook firing after compaction (`source: compact`) can check session state to see what context the system already injected and avoid redundant loads, or a PostToolUse hook can track which graph mutations have happened and tailor its `additionalContext`. Project-scoped state serves different purposes: tracking which baselines the developer created across sessions, recording historical verification pass rates per module, or flagging modules that have had repeated defect regressions.

The state store also supports Starlark script effects for logic that declarative rules can't express. Scripts run in a sandbox with access to `session_get`, `session_set`, `project_get`, and `project_set` but no filesystem or network access. They execute after the admission decision, so they can't influence it, but they can record derived state for subsequent hook invocations.

### Dynamic context in skills

Skills support two mechanisms for injecting live data into Claude's context: preprocessing substitution and instructed commands.

**Preprocessing with !`command`.** The !`command` syntax in skill content runs shell commands before the skill reaches Claude. The command output replaces the placeholder, so Claude receives actual data, not the command itself. This is not something Claude decides to execute; it's automatic substitution that happens at skill load time.

For ARCI skills, this is powerful. A task-execution skill can include a preprocessing directive like `!arci task context $ARGUMENTS` to inject the full task context (requirements, deliverables, dependent tasks, module domain context) before Claude sees the skill. A review skill can include `!arci reqlist --module $ARGUMENTS --format agent` to inject the module's requirements as structured data. A status skill can include `!arci graphsummary` to inject the current graph state. Claude receives the rendered context as part of the skill prompt without having to run any commands itself.

**Instructed commands.** Skills, subagents, and hook output can all include instructions that tell Claude to run `arci` CLI commands during execution. "Run `arci reqcoverage --module MOD-X` to check verification coverage" in a skill's instructions, "check for blocking defects with `arci defectlist --blocking --module MOD-X`" in a SubagentStart hook's `additionalContext`, or "run `arci taskready` to see what's unblocked" in a Stop hook's `reason` are all instances of the same pattern: the system tells the agent what to do next by pointing it at a CLI command. This is how the agent interacts with the graph during execution. Preprocessing can't replace it because `!`command`` runs once at skill load time and can't respond to what the agent discovers mid-workflow.

The two mechanisms compose well. Use !`command` for context that the skill always needs up front (task details, module requirements, domain context). Use instructed commands for interactive graph queries during execution: conditional checks, iterative workflows, and responses to intermediate results. A review skill preprocesses the module's requirements list but instructs the agent to run `arci defect create` for each problem found, because the number and content of defects aren't knowable at load time.

Compared to hook-injected context, skill-based context (both preprocessed and instructed) is more precise. A skill knows exactly what information the workflow needs. A SessionStart hook has to guess what context matters for the entire session. Hook output can bridge this gap by including instructed commands in `additionalContext` or `reason` fields, giving the agent specific next steps based on what just happened in the graph.

### Mechanism selection

These mechanisms serve different purposes and compose together.

Skills are the primary workflow driver. They encode the step-by-step process for each workflow, including what CLI commands to run and what context to load. Every workflow in this directory should have a corresponding skill (or set of skills) that the agent invokes.

Subagents handle workflows that benefit from isolation. Reviews, security audits, and focused research tasks get subagents. The subagent's `skills` field auto-loads the relevant ARCI skills. The subagent's system prompt sets the posture (critical reviewer, thorough analyst).

Hooks handle enforcement, ambient context, and continuation control. They're the safety net that ensures the agent doesn't skip graph operations, the ambient layer that keeps graph state visible, and the gate that prevents premature completion. The state store gives hooks memory across invocations, enabling escalating enforcement and context injection decisions that depend on what's already happened in the session.

Skills with preprocessing handle precision loading. When a specific workflow step needs specific graph data, the skill tells the agent to fetch it.

A typical workflow involves all of these: the agent recognizes a request and invokes a skill. The skill's `!`command`` preprocessing injects graph context before Claude sees the instructions. The skill's body tells the agent to run CLI commands during execution. If the workflow needs isolation, the skill delegates to a subagent with its own preloaded skills. Hooks inject ambient graph state at session start and after mutations, using the state store to track what they already injected and what changed. TaskCompleted and Stop hooks enforce completion criteria, with denial output steering the agent toward the right next step.

## Open design questions

A number of questions remain open beyond the preceding mechanism architecture.

**Ceremony calibration.** How does the agent decide whether a request needs the full transformation chain or just a quick task? This is the hardest question. Possible signals: whether the work touches existing requirements, whether it affects baselined content, the developer's explicit intent ("just fix it" vs. "design this properly"), and hook policies that enforce minimum ceremony for certain modules or phases.

**Session continuity.** Tasks execute in atomic sessions. How does state carry between sessions? Multiple mechanisms contribute partial answers. Subagent resumption (via `agentId`) handles subagent continuity. The task's prose content file can hold progress notes. The graph records what's complete and what's not. SessionStart hooks with `source: resume` can inject continuation context. Project-scoped state in the state store persists structured data across sessions (historical statistics, module flags, baseline records), though session-scoped state resets on each new session. But the specific in-progress state (which file the agent was editing, what approach it was taking) may need a dedicated persistence mechanism beyond what the graph, task files, and state store provide.

**MCP vs. CLI.** Should graph queries be available as MCP tools, or is Bash + CLI sufficient? MCP gives structured tool responses the agent can reason over directly. CLI via Bash requires parsing text output. For write operations the CLI is fine (the agent formats commands and reads exit codes). For read-heavy workflows (status checks, tracing, coverage) where the agent needs to make decisions based on graph data, MCP might give better ergonomics. The CLI's `--format agent` output mode (designed for machine consumption) partially addresses this. This might be a phased decision: start with CLI, add MCP if the CLI-via-Bash friction proves real.

**Hook context budget.** SessionStart hooks inject context that persists for the session. If ARCI injects a large graph summary, it consumes context window that the agent needs for actual work. The injection needs to be concise and relevant. PreCompact hooks can re-inject a condensed version before compaction, but the initial budget question remains: how much graph state is worth the context cost?

**Stop and TaskCompleted hook strategy.** Three handler types are available for completion enforcement: command hooks (fast, deterministic, can check whether `arci taskcomplete` ran), prompt hooks (single-turn LLM evaluation of whether the agent finished the work, 30 s default timeout), and agent hooks (multi-turn investigation that can read files and inspect the `.arci/` directory, 50-turn limit, 60 s default timeout). The right choice depends on how reliably the agent records deliverables. If the agent almost always finishes properly, a command hook checking graph state suffices. If it frequently stops with incomplete work, a prompt or agent hook that evaluates the transcript and codebase is worth the latency. TaskCompleted hooks are a better enforcement point than Stop hooks for ARCI's purposes, since they fire when a task closes rather than whenever Claude stops talking.

**Skill composition.** Complex workflows like "build a feature end-to-end" invoke multiple sub-workflows, each with its own skill. How do skills compose? Does the top-level skill reference sub-skills by name? Does it delegate to subagents that have those skills loaded? Does the agent just follow one skill's instructions, which tell it to `now follow the formalization workflow`? The progressive disclosure model in Claude Code skills (where supporting files load on demand) helps, but the orchestration pattern needs to be explicit.
