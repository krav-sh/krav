# Allocating requirements to child modules

## What

The developer says "flow down the performance requirement to the subsystems" or "allocate the latency budget across modules." A parent module's requirement gets distributed to child modules, creating derived requirements with budgets or partitions. This is how system-level obligations become component-level obligations.

## Why

Flow-down is essential for any project with more than one level of decomposition. A system-level requirement like "respond within 100 ms" is meaningless to a component developer unless it's broken down: "the parser gets 50 ms, the renderer gets 30 ms, 20 ms for overhead." Without flow-down, component teams work against vague expectations rather than concrete budgets.

## What happens in the graph

The source requirement gets `allocatesTo` edges pointing at child modules, each with optional allocation metadata (budget, partition). The agent creates derived requirements on each child module with `derivesFrom` edges pointing back at the parent requirement.

The allocation might be additive (latency budgets that sum to the parent budget), partitioned (each child handles a subset of the parent's scope), or replicated (each child must independently satisfy the full requirement).

## Trigger patterns

`Flow down REQ-X to child modules`, `allocate the latency budget`, `distribute this requirement`, `what's each module's share of X?`

## Graph before

A REQ-* node on a parent module that has child modules.

## Graph after

`allocatesTo` edges on the parent requirement, new REQ-* nodes on child modules with `derivesFrom` edges to the parent.

## Agent interaction layer

### Skills

The `krav:allocate` skill runs this workflow. Preprocessing loads the parent module's requirements and child module structure, giving the agent a complete picture of what it needs to distribute and where the allocations can go. This is essential context because allocation decisions depend on understanding each child module's responsibilities and boundaries.

The skill's instructed commands create `allocatesTo` edges on the parent requirement and derived REQ-* nodes on child modules. The skill instructions guide the agent through determining the allocation type (additive, partitioned, replicated) and ensuring budgets are consistent with the parent requirement's bounds.

### Policies

The `mutation-feedback` policy fires after the agent creates each allocation edge and derived requirement, injecting the running allocation state. For additive budgets, this means the agent sees how much of the parent's budget it has allocated so far and how much remains. This is the main mechanism for catching over-allocation before it becomes a problem.

The `prompt-context`, `graph-integrity`, and `cli-auto-approve` policies operate as in other graph-building workflows.

### Task types

Allocation doesn't create tasks. It creates derived requirements on child modules, which then feed into `krav:decompose` for task creation. The allocation step is strictly about distributing obligations, not planning work.

## Open questions

**How does the agent determine allocation strategy?** Additive budgets (50 ms + 30 ms + 20 ms = 100 ms) work for latency and resource consumption. But "the system shall support offline mode" doesn't partition numerically. The agent needs to understand allocation types and ask the right questions for each.

**What if the budgets don't add up?** If the parent requires 100 ms and the developer allocates 60 ms + 50 ms to children, the agent should flag this. But should it refuse the allocation, warn, or just record it? Some over-allocation is intentional (margin) and some is a mistake.

**How deep does flow-down go?** Can a flowed-down requirement on a subsystem be further flowed down to its components? The design supports this (requirements can derive from other requirements), but multi-level flow-down gets complex. The agent needs to handle it but probably shouldn't propose it unprompted.
