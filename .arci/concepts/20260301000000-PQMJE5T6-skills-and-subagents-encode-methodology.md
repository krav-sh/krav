# Skills and subagents encode development method

arci's development method isn't just enforced by hooks or documented in CLAUDE.md. Skills teach the agent step-by-step workflow instructions, and subagents provide isolated execution contexts. Together they form the agent interaction layer: the bridge between the knowledge graph and what Claude Code actually does.

## Skills

Skills are markdown files with YAML frontmatter that encode specific development workflows. Each skill maps to an arci workflow: formalization, derivation, decomposition, task execution, review, verification, and so on.

Skills have two mechanisms for incorporating graph context:

**Preprocessing commands** (`!`command``) run at skill load time and inject their output into the skill content before Claude sees it. The task execution skill preprocesses `arci taskcontext TASK-42` to inject the task's requirements, deliverables, dependencies, and module domain context. The agent receives the rendered skill with full task context already embedded.

**Instructed commands** are command-line invocations that the skill instructions tell the agent to run during execution. "Run `arci task update TASK-42 --add-deliverable src/parser.go`" is an instructed command. The agent executes these as part of following the workflow.

## Subagents

Subagents provide isolated execution contexts for workflows that need a fresh context window, restricted tool access, or a different posture. The primary use case is review: a review subagent starts with the review skill loaded (full content, including preprocessed requirements data) and operates with read-only tool access. It hasn't seen the coding context, so it reviews without bias.

Subagents load specified skills at startup (full content, not just descriptions) and can define lifecycle-scoped hooks in their frontmatter. They're defined as markdown files in `.claude/agents/` and bundled with the arci plugin.

## Why encode the method, not just enforce it

Hooks tell the agent "no." Skills tell the agent "here's how." Both are necessary. A hook that denies a write to a baselined file is useful, but without a skill that teaches the agent the correct alternative workflow (create a defect, unlock the module, make changes, re-baseline), the agent has nowhere to go.

The combination creates a teaching system: hooks catch violations, and skills provide the positive guidance that prevents violations in the first place. Over time, as the agent follows skills successfully, the hooks become a safety net rather than the primary guidance mechanism.
