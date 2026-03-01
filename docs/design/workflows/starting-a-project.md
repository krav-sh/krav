# Starting a new project

## What

The developer tells Claude Code they want to build something: "building a CLI tool for managing dotfiles" or "set up a new library for date parsing." The graph is empty or the `.arci/` directory doesn't exist yet. By the end of this workflow, the graph has a root module, an initial module hierarchy, founding concepts capturing the core design thinking, and stakeholder needs derived from those concepts. Depending on how far the developer wants to go in a single session, it might also produce initial requirements and an architecture-phase task DAG.

## Why

This is the on-ramp. If bootstrapping is tedious, people won't use ARCI. If it's too automated, the graph starts with generic placeholder content that nobody trusts. The workflow needs to feel like a productive design conversation that happens to leave structured artifacts behind, not like filling out a form.

## What happens in the graph

The workflow starts with `arci init` (or whatever the bootstrapping command is; this isn't designed yet). That creates the `.arci/` directory, `graph.jsonlt`, and the root module.

From there, the agent needs to extract concepts from the developer's description of what they're building. "Building a CLI for dotfiles" implies architectural concepts (how the CLI structures itself), operational concepts (how users interact with it), and possibly integration concepts (how it interacts with git, cloud sync, etc.). Each becomes a CON-* node.

The agent then establishes the module hierarchy. For a CLI tool, this might be root then command parsing, config management, file operations, output formatting. Each becomes a MOD-* node with `childOf` edges.

Bootstrapping also creates the project's initial stakeholders as STK-* nodes. The agent identifies who cares about the system from the developer's description and asks about any parties it can't infer. For a CLI tool, this might produce stakeholders like "CLI end user", "package ecosystem" (if it gets distributed), and "contributor" (if it's open source).

The founding concepts get formalized into needs. "Users need to sync dotfiles across machines" becomes NEED-* with `derivesFrom` pointing at the relevant concept and `stakeholder` referencing the appropriate STK-* nodes. Needs get MoSCoW priorities.

If the developer wants to go further, needs get derived into requirements, and architecture-phase tasks get created for each module.

## Trigger patterns

`Build a new X`, `start a new project`, `set up a library for`, `I want to create`, `initialize ARCI`, `bootstrap this project`.

## Graph before

Empty, or `.arci/` doesn't exist.

## Graph after

At minimum: one root MOD-*, 2-6 child MOD-* nodes, 2-4 STK-* nodes, 2-5 CON-* nodes, 3-10 NEED-* nodes. The exact counts depend on how far the conversation goes.

## Agent interaction layer

### Skills

The `arci:init` skill builds this workflow. Because the graph is empty at invocation time, there's no preprocessing to do: the !`command` directives that other skills use to inject graph context have nothing to load. Instead, the skill body focuses on interactive instructions: how to identify stakeholders from the developer's project description and create STK-* nodes, how to extract concepts, when to ask about architectural boundaries, and how to structure the initial module hierarchy.

If the developer wants to go further in the same session (formalizing concepts, deriving requirements, creating architecture-phase tasks), the `arci:init` skill doesn't handle that directly. The agent transitions to `arci:formalize`, `arci:derive`, or `arci:decompose` as needed. The `arci:feature` skill's orchestration model is one possible approach for chaining these, though bootstrapping is different enough from feature development that `arci:init` manages its own flow.

### Policies

The `session-context` policy fires at session start but has minimal effect during bootstrapping since the graph is empty or nonexistent. After `arci init` creates the `.arci/` directory and root module, subsequent graph mutations start producing useful context through `mutation-feedback`. That policy fires after each `arci` CLI command that creates modules, concepts, or needs, injecting updated graph state so the agent stays aware of what it has built so far without re-querying.

The `graph-integrity` policy matters from the first mutation. Even during bootstrapping, the agent must create nodes through the CLI rather than writing `.arci/` files directly. The `cli-auto-approve` policy keeps this frictionless by auto-approving `arci` commands so the agent doesn't stall on permission prompts while creating the initial module hierarchy.

### Task types

Bootstrapping doesn't inherently create tasks. The workflow's primary outputs are stakeholders, modules, concepts, and needs. If the developer pushes the session far enough to start decomposing work, `decompose-module` tasks appear for fleshing out the child module hierarchy, and other architecture-phase tasks like `decide-architecture` or `spike` might follow. But the typical bootstrapping session stops at needs, with task creation happening in a subsequent session via `arci:decompose`.

## Open questions

**How interactive should bootstrapping be?** Should the agent ask the developer a series of questions (`who are the stakeholders?`, `what are the main subsystems?`) or should it propose a structure based on the initial description and let the developer adjust? The former is thorough but slow. The latter is faster but might miss things.

**Where does `arci init` end and agent-driven exploration begin?** The CLI command creates the directory structure and root module. Everything after that is agent behavior. Should `arci init` have an interactive mode that walks through module creation, or is that purely the agent's job?

**How much should the agent infer vs. ask?** If the developer says "build a REST API," the agent can infer a lot about the module hierarchy (routes, middleware, data layer, auth). Should it propose that structure or ask about it? The answer probably depends on how conventional the project is.

**What's the minimum viable bootstrap?** Can a developer just create a root module and start working, adding structure incrementally? Or does ARCI expect a certain minimum graph to function well (needs must exist before the agent can derive requirements, etc.)? The transformation preconditions in the design suggest the latter, but forcing all of that upfront might stop adoption.

**Template-driven vs. conversational.** Should there be project-type templates (`CLI tool`, `library`, `web app`) that pre-populate a starting structure? Or is the conversational approach more flexible? Both have precedent: `cargo init` gives you a structure, but it's a fixed one.
