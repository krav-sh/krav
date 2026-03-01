# Predicate domain/range enforcement

The graph engine shall validate that each relationship property appears only on source node types and targets only node types permitted by the predicate's domain/range definition, rejecting violations.

## Context

Constraint C-REF2 in the design documentation requires that each predicate respect its defined domain and range. A `childOf` property is only valid on Module nodes targeting other Module nodes. A `derivesFrom` on a Requirement must target a Need or another Requirement. These constraints are not arbitrary; they encode the ontology's semantic structure.

Domain/range violations are a more subtle form of data integrity problem than dangling references. A dangling reference fails visibly (the target doesn't exist), but a domain/range violation can silently corrupt the graph's semantics: a `verifiedBy` edge from a Module to a test case would pass referential integrity (both nodes exist) but would violate the ontology (modules aren't verified by test cases; requirements are).

For the "single authoritative data store" concern, domain/range enforcement ensures the store's semantics are trustworthy, not just its references. Without it, queries that traverse relationship edges may return nonsensical results, undermining the graph's role as the source of truth.

## Verification approach

For each predicate in the domain/range matrix (defined in the predicates design doc), attempt to create a node that uses the predicate with an invalid source type or an invalid target type. The engine must reject each attempt. Test representative cases: `childOf` on a non-Module source, `derivesFrom` targeting an invalid type, `verifiedBy` on a Module instead of a Requirement.
