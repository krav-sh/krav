# Node identity uniqueness

The graph engine shall enforce that no two nodes share the same `@id` value, rejecting any mutation that would create a duplicate.

## Context

If two nodes share an `@id`, the graph has no unambiguous authority for what that identifier refers to. This directly undermines the "single source of truth" guarantee that NEED-7DT7XGE6 establishes. The design documentation lists uniqueness as constraint C-ID3, but this requirement makes it a formal obligation on the graph engine rather than just a documented constraint.

Uniqueness enforcement is the graph engine's responsibility, not a convention for contributors to follow manually. When the engine rejects a duplicate, it prevents the inconsistency at the point of creation rather than discovering it later through validation.

## Verification approach

Attempt to create a node whose `@id` matches an existing node in the graph. The engine must reject the operation with an error identifying the conflict. Verify that the existing node is not modified or corrupted by the failed operation.
