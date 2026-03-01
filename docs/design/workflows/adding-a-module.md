# Adding a module

## What

The developer says "add an auth subsystem" or "add a caching layer" or "the parser should have a separate lexer component." A new module gets created in the hierarchy, and depending on how far the developer wants to go, it gets populated with concepts, needs, requirements, and tasks.

## Why

Projects grow. The initial module hierarchy is a starting point, not a final answer. As development reveals new subsystems, or as scope expands, the team needs to introduce new modules with proper graph connectivity, not just as orphan directories in the codebase.

## What happens in the graph

The agent creates a new MOD-* node with `childOf` pointing at its parent. If the developer has a clear picture of what the module is for, concepts get created (or existing cross-cutting concepts get linked via `informs`). If the new module picks up responsibilities from an existing module, requirements might need reparenting and needs might need revisiting.

The new module starts at the architecture phase. If the parent is further along, that's fine; with module-scoped phase gating, the new child progresses independently.

## Trigger patterns

`We need a module for X`, `add a subsystem for Y`, `the parser should be split into lexer and parser`, `create a new component for Z`.

## Graph before

An existing module hierarchy.

## Graph after

A new MOD-* node in the hierarchy, possibly with initial CON-*, NEED-*, and TASK-* nodes.

## Agent interaction layer

### Skills

The `arci:module-add` skill runs this workflow. Preprocessing loads the current module hierarchy so the agent understands where the new module fits. It also loads any allocated requirements from the parent module that might flow down to the new child, since module creation often coincides with requirement redistribution.

The skill's instructed commands create the MOD-* node, establish `childOf` edges, and optionally flow down requirements from the parent. If the new module picks up responsibilities from a sibling (a split scenario), the skill's instructions guide the agent through reparenting requirements, though the heavier restructuring scenarios are better handled by `arci:restructure`.

### Policies

The `mutation-feedback` policy fires after the agent creates the module and after any requirement reparenting, keeping the agent aware of the changing hierarchy. The `prompt-context` policy is useful when the developer references specific parent modules or requirements during the conversation about where the new module belongs.

The `graph-integrity` and `cli-auto-approve` policies apply as usual. If the parent module is baselined, the `baseline-protection` policy also comes into play, since adding a child module and flowing down requirements modifies the parent module's graph neighborhood.

### Task types

Module creation typically produces `decompose-module` tasks for the new module's architecture work. If the developer wants to immediately plan the new module's development, the agent transitions to `arci:decompose` to generate a task DAG. The new module starts at the architecture phase, so initial tasks are architecture-phase types like `decompose-module`, `define-interface`, and `decide-architecture`.

## Open questions

**What happens to existing content when splitting a module?** If the developer splits MOD-parser into MOD-lexer and MOD-parser-core, the agent must reassign requirements currently owned by MOD-parser. The agent needs to identify which requirements belong to which child and reparent them. This has traceability implications: `derivesFrom` chains stay intact but `module` ownership changes.

**Integration modules.** When two sibling modules need to interact, should the agent suggest an integration module? The design has integration modules with `integrates` edges, but creating them is a design decision the developer should make.

**Phase of the new module.** A new module starts at architecture phase, but if it's a simple utility component with a well-understood design, the developer might want to skip ahead. With module-scoped gating, the module can advance independently, but should the agent suggest starting at a later phase if the work is straightforward?
