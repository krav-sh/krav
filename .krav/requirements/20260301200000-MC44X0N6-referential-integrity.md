# Referential integrity on mutation and load

The graph engine shall validate that every `@id` reference in a node's relationship properties resolves to an existing node, rejecting mutations that would create dangling references and reporting violations detected when loading a persisted graph.

## Context

Referential integrity is the difference between a data store that is authoritative and one that merely contains data. If a requirement's `derivesFrom` points to a need that doesn't exist, the traceability chain breaks and the store cannot answer "why does this requirement exist?" NEED-7DT7XGE6's rationale identifies that question as the cost of scattered metadata.

Constraint C-REF1 in the design documentation requires that every `{"@id": "TARGET-ID"}` in a relationship property corresponds to an existing node. This requirement makes the graph engine responsible for enforcing that constraint in two distinct contexts.

The mutation path is the primary enforcement point: when creating or updating a node, the engine validates all relationship properties before committing the change. The load path is the secondary enforcement point: when reading a persisted JSONLT file, the engine detects and reports any dangling references that exist in the file (whether from manual editing, corruption, or version conflicts).

The distinction matters because the mutation path can reject the operation and preserve consistency, while the load path must report violations without necessarily refusing to load the graph (a graph with dangling references has degraded integrity but may still be partially useful for diagnosis).

## Verification approach

Test the mutation path by attempting to create a node with a relationship property referencing a non-existent `@id`; the engine must reject the operation with an error identifying the dangling reference. Test the load path by constructing a JSONLT file containing a node with a dangling reference and loading it; the engine must report the violation (the specific handling, whether warning or error, is a design decision, but detection and reporting are mandatory).
