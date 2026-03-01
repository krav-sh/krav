# Type-prefix consistency

The graph engine shall enforce that each node's `@id` prefix matches its `@type` according to the node type taxonomy, rejecting nodes where the prefix and type are inconsistent.

## Context

Constraint C-ID2 in the design documentation defines the mapping between identifier prefixes and node types: CON for Concept, MOD for Module, NEED for Need, REQ for Requirement, TC for test case, TASK for Task, DEF for Defect, BSL for Baseline, STK for Stakeholder, TP for test plan, DEV for Developer, and AGT for Agent.

This convention is not just cosmetic. It makes identifiers human-readable and enables prefix-based filtering (like `grep "^.*REQ-" graph.jsonlt`). For the "single source of truth" concern, an incoherent node (where the prefix claims one type but `@type` declares another) is a data integrity violation that erodes trust in the store.

## Verification approach

For each entry in the prefix-to-type mapping, attempt to create a node with a mismatched prefix and type (such as `@id: "CON-XXXXXXXX"` with `@type: "Task"`). The engine must reject each attempt. Also verify the positive case: the engine accepts nodes with correctly matched prefix and type.
