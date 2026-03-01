# Query patterns

## Overview

This document catalogs canonical graph traversal patterns that realize the [competency questions](competency-questions.md). Each pattern describes the traversal, its starting point, the predicates followed, and complexity characteristics.

## Traceability traversals

### Ancestry (CQ-T1, CQ-T5)

Traverse `derivesFrom` edges from a node to its sources, recursively. Answers "why does this exist?"

```
ancestry(node) =
  sources = node.derivesFrom
  for each source in sources:
    yield source
    yield* ancestry(source)
```

**Starting point**: Any NED or REQ
**Predicates**: `derivesFrom` (backward through chain)
**Terminates at**: CON nodes (concepts have no `derivesFrom`)
**Complexity**: O(n) where n is the length of the derivation chain. Typically short (2–4 hops: REQ → NED → CON).

### Descendancy (CQ-T2, CQ-T4)

Traverse `derivesFrom` edges in reverse — find all nodes that derive from a given source. Answers "what did this produce?"

```
descendancy(node) =
  for each downstream where downstream.derivesFrom contains node:
    yield downstream
    yield* descendancy(downstream)
```

**Starting point**: Any CON, NED, or REQ
**Predicates**: `derivesFrom` inverse
**Complexity**: O(n) where n is the number of nodes in the derivation subgraph. Requires scanning all nodes or maintaining a reverse index.

### Full traceability chain (CQ-T5, CQ-X1)

Combines ancestry and descendancy with related verifications, defects, and tasks. The complete context for a node.

```
full_chain(node) =
  ancestors = ancestry(node)
  descendants = descendancy(node)
  verifications = node.verifiedBy (if REQ)
  defects = all DEF where DEF.subject = node
  tasks = all TSK where TSK.module = node.module
  return {ancestors, descendants, verifications, defects, tasks}
```

**Complexity**: O(n + m) where n is the derivation subgraph size and m is the number of related entities.

## Hierarchy traversals

### Module subtree (CQ-P5, CQ-PH2)

Traverse `childOf` edges in reverse to find all descendants of a module.

```
subtree(mod) =
  for each child where child.childOf = mod:
    yield child
    yield* subtree(child)
```

**Starting point**: Any MOD
**Predicates**: `childOf` inverse
**Complexity**: O(n) where n is the number of modules in the subtree.

### Scoped subgraph (CQ-P5)

All nodes owned by a module subtree. Combines module subtree with `module` predicate.

```
scoped_subgraph(mod) =
  modules = {mod} ∪ subtree(mod)
  for each node where node.module ∈ modules:
    yield node
```

**Starting point**: Any MOD
**Predicates**: `childOf` inverse, then `module` inverse
**Complexity**: O(n) where n is the total number of nodes in the graph (requires scanning). Can be optimized with a module→nodes index.

## Dependency traversals

### Critical path (CQ-P2)

The longest path through incomplete tasks in the `dependsOn` DAG leading to a target task. This is the sequence of tasks that determines the minimum time to completion.

```
critical_path(target) =
  longest_path in dependsOn DAG from any root to target
  where all tasks on path have status ≠ complete
```

**Starting point**: A milestone TSK
**Predicates**: `dependsOn` (reverse traversal from target to roots)
**Complexity**: O(V + E) where V is the number of tasks and E is the number of dependency edges. Longest path in a DAG is computed via topological sort.

### Blocking set (CQ-P3)

All incomplete tasks that transitively block a given task.

```
blocking_set(task) =
  for each dep in task.dependsOn:
    if dep.status ≠ complete:
      yield dep
      yield* blocking_set(dep)
```

**Starting point**: Any TSK
**Predicates**: `dependsOn`
**Complexity**: O(n) where n is the size of the transitive dependency closure. Typically small relative to total task count.

### Ready set (CQ-P1)

Tasks that are not complete and have all dependencies satisfied.

```
ready_set() =
  for each task where task.status ≠ complete:
    if all dep in task.dependsOn have dep.status = complete:
      yield task
```

**Starting point**: All tasks
**Predicates**: `dependsOn`
**Complexity**: O(V + E) — scan all tasks, check each dependency.

## Coverage analysis

### Unverified requirements (CQ-C1)

Requirements with no `verifiedBy` edges.

```
unverified() =
  for each req where req.verifiedBy is empty:
    yield req
```

**Starting point**: All REQ nodes
**Predicates**: `verifiedBy`
**Complexity**: O(n) where n is the number of requirements.

### Coverage ratio (CQ-C2)

Verification coverage for a module, expressed as a ratio.

```
coverage(mod) =
  reqs = all REQ where REQ.module ∈ subtree(mod)
  verified = reqs where REQ.verifiedBy is non-empty
  passing = reqs where any VRF in REQ.verifiedBy has status = passing
  return {total: |reqs|, verified: |verified|, passing: |passing|}
```

**Starting point**: Any MOD
**Predicates**: `module` inverse, `verifiedBy`, VRF status
**Complexity**: O(n) where n is the number of requirements in the module subtree.

## Impact analysis

### Suspect link propagation (CQ-I1, CQ-I2)

When a node is modified, find all edges that become suspect.

```
suspect_impact(node) =
  if node is CON:
    yield all NED.derivesFrom edges pointing at node
  if node is NED:
    yield all REQ.derivesFrom edges pointing at node
  if node is REQ:
    yield all child REQ.derivesFrom edges pointing at node
    yield all REQ.verifiedBy edges
    yield all REQ.allocatesTo edges
```

**Starting point**: The modified node
**Predicates**: `derivesFrom` inverse, `verifiedBy`, `allocatesTo`
**Complexity**: O(d) where d is the out-degree of the modified node's inverse relationships. Non-transitive by default (see [constraints](constraints.md) C-SUSPECT2).

### What-if analysis

Simulate the impact of a proposed change without applying it.

```
what_if(node, proposed_change) =
  suspect_edges = suspect_impact(node)
  affected_nodes = targets of suspect_edges
  affected_verifications = all VRF reachable via verifiedBy from affected REQs
  return {suspect_edges, affected_nodes, affected_verifications}
```

**Complexity**: Same as suspect link propagation plus verification lookup.

## Phase gate queries

### Advancement readiness (CQ-PH1)

Check whether a module can advance to the next phase.

```
can_advance(mod, target_phase) =
  current = mod.phase
  phase_tasks = all TSK where TSK.module = mod and TSK.processPhase = current
  incomplete_tasks = phase_tasks where TSK.status ≠ complete
  blocking_defects = all DEF where DEF.module = mod
    and DEF.severity ∈ {critical, major}
    and DEF.status ∈ {open, confirmed}
  parent = mod.childOf
  parent_ok = parent is null or parent.phase ≥ target_phase
  return |incomplete_tasks| = 0 and |blocking_defects| = 0 and parent_ok
```

**Starting point**: Any MOD
**Predicates**: `module` inverse (for tasks and defects), `childOf`
**Complexity**: O(t + d) where t is the number of tasks for this module/phase and d is the number of defects for this module.

### Behind-phase modules (CQ-PH2)

Find modules whose phase is behind their parent's phase.

```
behind_phase() =
  for each mod where mod.childOf is set:
    parent = mod.childOf
    if mod.phase < parent.phase:
      yield {mod, mod.phase, parent, parent.phase}
```

**Starting point**: All MOD nodes with parents
**Predicates**: `childOf`
**Complexity**: O(n) where n is the number of modules.

## Structural integrity queries

### Orphan detection (CQ-S1)

Find nodes with no structural connections (no incoming or outgoing edges on structural predicates).

```
orphans() =
  for each node (excluding root MOD):
    has_structural_edge = any of:
      - node has derivesFrom, module, childOf, dependsOn, verifiedBy, subject, detectedBy, generates, informs
      - any other node references this node via those predicates
    if not has_structural_edge:
      yield node
```

**Complexity**: O(n + e) where n is the number of nodes and e is the number of edges.

### Cycle detection (CQ-S3, CQ-S4)

Detect cycles in DAG-constrained subgraphs.

```
has_cycle(subgraph) =
  topological_sort(subgraph) fails ⟺ cycle exists
```

**Applied to**: `derivesFrom` subgraph, `dependsOn` subgraph
**Complexity**: O(V + E) via topological sort.

### Dangling reference detection (CQ-S2)

Find predicate targets that don't correspond to existing nodes.

```
dangling_refs() =
  for each node:
    for each predicate value {"@id": target_id}:
      if target_id not in graph:
        yield {node, predicate, target_id}
```

**Complexity**: O(n × p) where n is the number of nodes and p is the average number of predicates per node.

## Baseline queries

### Semantic diff (CQ-A1)

Compare graph state between two baselines.

```
semantic_diff(baseline_a, baseline_b) =
  graph_a = reconstruct_graph(baseline_a.commitSha)
  graph_b = reconstruct_graph(baseline_b.commitSha)
  added = nodes in graph_b not in graph_a
  removed = nodes in graph_a not in graph_b
  modified = nodes in both where properties differ
  return {added, removed, modified}
```

**Starting point**: Two BSL nodes
**Complexity**: O(n) where n is the total number of nodes across both graphs. Dominated by the cost of reconstructing graph state from git commits.

### Deferred defects by target (CQ-A5)

Find deferred defects that target a specific baseline or phase.

```
deferred_for_target(target) =
  for each def where def.status = deferred:
    if def.deferralTarget matches target:
      yield def
```

**Complexity**: O(d) where d is the number of deferred defects.
