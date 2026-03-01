# Checking project status

## What

The developer says `where are we?`, `what's the status of the parser?`, `what's blocking the release?`, `what's ready to work on?`, or `give me an overview`. The agent queries the graph and synthesizes a status report. Nothing changes in the graph; this is pure read.

## Why

Status is the most common question in any project. The graph contains all the information needed to answer it, but the raw data (node counts, phase values, task states, defect counts) needs synthesis into something a human can act on. "The parser has 3 of 7 coding tasks complete, 2 minor defects open, and 85% verification coverage" is more useful than a dump of task records.

## What happens in the graph

Nothing. The agent runs graph queries and presents results.

The queries depend on the scope of the question. The project-wide query hits everything: module phases, task completion rates, defect status, coverage percentages. A module-scoped query focuses on that subtree. A goal-directed query traverses the dependency graph backward from the release milestone.

Key CLI commands involved: `arci moduletree` (hierarchy overview), `arci taskready` (available work), `arci taskblocking TASK-X` (blockers for a specific goal), `arci reqcoverage` (verification coverage), `arci defect open` (unresolved defects), `arci modulephase` (phase status).

## Trigger patterns

`Where are we?`, `project status`, `what's the status of X?`, `what should I work on?`, `what's blocking Y?`, `how's coverage?`, `any open defects?`

## Graph before

Whatever the current state is.

## Graph after

Unchanged.

## Agent interaction layer

### Skills

The `arci:status` skill builds this workflow. Preprocessing loads the graph summary via `arci moduletree` and other read-only queries, giving the agent a precomputed overview of module phases, task progress, defect counts, verification coverage, and suspect links pending review. The skill body is shorter than graph-building skills: it tells the agent how to synthesize raw graph data into a narrative status report rather than guiding a complex transformation.

The skill's instructed commands let the agent drill into specific areas during the conversation: `arci taskready` for available work, `arci defect open` for unresolved problems, `arci reqcoverage` for verification gaps. These on-demand queries complement the preprocessed overview when the developer asks follow-up questions about specific modules or blockers.

### Policies

The `session-context` policy does most of the heavy lifting for status-oriented sessions. It injects the current task context, ready task list, recent defects, and suspect links at session start, which overlaps with what the status skill preprocesses. The difference is that `session-context` fires automatically on every session, while the status skill fires deliberately when the developer asks for a status report.

The `prompt-context` policy fires when the developer mentions specific modules or tasks in follow-up questions, injecting their current state. No enforcement or mutation policies are relevant since this workflow only reads the graph.

### Task types

This workflow doesn't create or execute tasks. It reports on them.

## Open questions

**How should the agent present status?** A single session response? A structured report? The answer depends on scope. A module-level status check is a paragraph. A project-wide status overview might need a structured format with sections per module. Should the agent produce a document as a deliverable, or just respond conversationally?

**Derived metrics.** Some useful status indicators require computation: percentage of requirements verified, critical path length to a milestone, defect density per module. Should the CLI compute these (and make them available as commands) or should the agent compute them from raw graph data? The former is more reliable. The latter is more flexible.

**Status over time.** "How have things changed since last week?" requires comparing current state to historical state. Baselines provide point-in-time snapshots, but they're created intentionally, not continuously. Is there a lightweight mechanism for tracking progress over time beyond baseline diffs?

**Proactive status.** Should ARCI surface status information without the developer asking? A SessionStart hook could inject "since your last session: 2 tasks completed, 1 defect opened, parser advanced to coding" into the agent's context. This keeps the developer oriented between sessions, but it consumes context window for information the developer might not care about.
