# Lifecycle coordination

## Overview

Each entity type in the knowledge graph has its own lifecycle state machine. This document defines how lifecycle state changes on one entity interact with and propagate to related entities across the graph.

## Entity lifecycles

### Concept lifecycle

```
draft → exploring → crystallized → superseded
```

- **draft**: Initial capture of an idea
- **exploring**: Active investigation and elaboration
- **crystallized**: Ready for formalization into needs
- **superseded**: Replaced by a newer concept

### Need lifecycle

```
draft → validated → addressed → superseded
```

- **draft**: Initial formulation of stakeholder expectation
- **validated**: Confirmed as a real stakeholder expectation
- **addressed**: All derived requirements are satisfied
- **superseded**: Replaced or no longer relevant

### Requirement lifecycle

```
draft → approved → satisfied → superseded
```

- **draft**: Initial formulation of design obligation
- **approved**: Accepted as a binding obligation
- **satisfied**: All verifications are passing
- **superseded**: Replaced or withdrawn

### Verification lifecycle

```
planned → implementing → passing → failing → blocked
```

- **planned**: Verification defined but not yet implemented
- **implementing**: Verification under development
- **passing**: Verification executed and passing
- **failing**: Verification executed and failing
- **blocked**: Cannot execute (dependency missing, environment unavailable)

### Task lifecycle

```
pending → active → complete → blocked → cancelled
```

- **pending**: Not yet started, dependencies may be incomplete
- **active**: Currently being worked on
- **complete**: Work finished, deliverables produced
- **blocked**: Cannot proceed (dependency or external blocker)
- **cancelled**: Abandoned, will not be completed

### Defect lifecycle

```
open → confirmed → resolved → verified → closed
                ↓
             rejected
             deferred → confirmed
```

- **open**: Reported, not yet triaged
- **confirmed**: Triaged as a real problem
- **rejected**: Not a problem (with rationale)
- **deferred**: Real problem, intentionally postponed
- **resolved**: Fix complete, awaiting verification
- **verified**: Fix confirmed adequate
- **closed**: Fully resolved

### Baseline lifecycle

```
draft → proposed → approved → superseded
```

- **draft**: Being assembled
- **proposed**: Submitted for approval
- **approved**: Official reference point
- **superseded**: Replaced by a newer baseline

### Module phase lifecycle

```
architecture → design → implementation → integration → verification → validation
```

Module phase is distinct from module status. Phase tracks where the module is in the lifecycle; status tracks whether the module is active, deprecated, or archived.

## Upward propagation

State changes on leaf entities propagate upward through the derivation chain.

### Task completion drives requirement status

When all tasks for a requirement's module (at the current phase) complete, and the requirement's verifications are passing, the requirement can transition to `satisfied`:

```
TSK-* → status: complete (all tasks for this phase)
  + VRF-* → status: passing (all verifications for this requirement)
    ⇒ REQ-* can transition to: satisfied
```

This is not automatic — an explicit check or review confirms that task completion and passing verifications collectively satisfy the requirement.

### Requirement satisfaction drives need status

When all requirements derived from a need are `satisfied`, the need can transition to `addressed`:

```
REQ-* → status: satisfied (all requirements deriving from this need)
  ⇒ NED-* can transition to: addressed
```

### Verification status drives requirement assessment

When a VRF transitions to `passing`, the requirement it verifies may now be assessable for `satisfied` status. When a VRF transitions to `failing`, the requirement's `satisfied` status is called into question.

## Downward impact

Changes to upstream entities impact downstream entities through suspect link propagation.

### Concept modification

When a CON node's key properties change:
1. `derivesFrom` edges on NED nodes that derive from this CON are marked suspect
2. Reviewers examine each suspect NED to determine if the change affects the need
3. If the need is affected, the reviewer updates it and suspect propagates further downstream

### Need modification

When a NED node's `statement` or `status` changes:
1. `derivesFrom` edges on REQ nodes that derive from this NED are marked suspect
2. `verifiedBy` edges on those REQ nodes may also become suspect (if the requirement changes)

### Requirement modification

When a REQ node's `statement` changes:
1. `verifiedBy` edges are marked suspect (verification may no longer be valid)
2. `derivesFrom` edges on child REQ nodes are marked suspect (derived requirements may need updating)
3. `allocatesTo` edges are marked suspect (allocations may need re-evaluation)

## Phase-lifecycle interaction

### Module advancement criteria

A module can advance to the next phase when:

1. **All phase tasks complete**: Every TSK where `module = this MOD` and `processPhase = current phase` has `status = complete`
2. **No blocking defects**: No DEF where `module = this MOD` has `severity ∈ {critical, major}` and `status ∈ {open, confirmed}`
3. **Review dispositions acceptable**: All review tasks for the current phase are complete with acceptable dispositions (accepted, or conditionally accepted with all conditions resolved)
4. **Parent at or ahead**: If this module has a `childOf` parent, the parent's phase is ≥ the target phase

### Task completion drives phase readiness

Each task completion reduces the set of remaining work for its process phase. When the last task for a phase completes, the module becomes eligible for phase advancement (subject to defect and review checks).

### Phase regression

When a module regresses to an earlier phase:
1. A DEF node is automatically created to record why the regression occurred
2. Child modules remain at their current phase but are blocked from advancing past the new parent phase
3. Tasks for phases after the regression target may need to be re-evaluated

## Suspect link interaction

### What triggers suspect marking

| Entity modified | Edges marked suspect |
|----------------|---------------------|
| CON (key properties) | `derivesFrom` on downstream NEDs |
| NED (`statement`, `status`) | `derivesFrom` on downstream REQs |
| REQ (`statement`) | `verifiedBy` edges, `derivesFrom` on child REQs, `allocatesTo` edges |

### Suspect link review workflow

1. A modification triggers suspect marking on downstream edges
2. Suspect links appear in the suspect link view (`arci graph suspect`)
3. A reviewer examines each suspect link and takes one of:
   - **Clear**: The link is still valid despite the upstream change
   - **Update**: The downstream node needs a minor adjustment (no defect)
   - **Defect**: The upstream change reveals a real problem → create a DEF node

Suspect links do not auto-generate defects. They are a signal for human review.

### Suspect and baseline interaction

When reviewing suspect links, baselines provide a reference point. The reviewer can compare the current state against the last approved baseline to understand what changed and assess whether suspect links represent real problems.

## Defect lifecycle interaction

### Defect creation from examination

Review tasks (architecture-review, design-review, code-review) produce defects as part of their output:

```
TSK-R3V13W01 (review task)
  produces → DEF-F1L4T7W5 (detectedBy → TSK-R3V13W01)
```

### Defect remediation chain

```
DEF-F1L4T7W5 (confirmed)
  generates → TSK-F1X00001 (remediation task)
    TSK-F1X00001 → status: complete
      ⇒ DEF-F1L4T7W5 → status: resolved
        ⇒ re-examination confirms fix
          ⇒ DEF-F1L4T7W5 → status: verified → closed
```

### Defect and phase gates

Phase advancement checks defect status:
- **Blocking defects** (critical, major with status open or confirmed) prevent advancement
- **Deferred defects** do not block (the deferral is an explicit decision)
- **Resolved but unverified defects** are borderline — project policy determines whether they block

## State transition dependency matrix

This matrix shows which entity status transitions depend on the status of related entities:

| Transition | Depends on |
|-----------|------------|
| NED → addressed | All derived REQs are satisfied |
| REQ → satisfied | All VRFs are passing |
| MOD phase advance | All phase TSKs complete, no blocking DEFs, reviews acceptable, parent phase ≥ target |
| DEF → resolved | Generated TSK is complete |
| DEF → verified | Re-examination confirms fix adequate |
| BSL → approved | No blocking DEFs in scope (policy-dependent) |
