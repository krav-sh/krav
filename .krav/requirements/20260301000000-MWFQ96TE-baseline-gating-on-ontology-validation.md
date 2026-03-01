# Baseline gating on ontology validation

The project shall not create the first ontology baseline until the graph contains at least one instance of every node type and at least one relationship uses every predicate.

## Rationale

Exercising every node type and predicate under manual editing conditions reveals schema problems (missing fields, awkward constraints, unclear semantics) that would be expensive to fix after tooling codifies the schema. Baselining before full coverage risks locking in untested elements.

## Verification criteria

A jq query over graph.jsonlt at baseline time confirms all 12 node types (Concept, Module, Need, Requirement, test case, Task, Defect, Baseline, Stakeholder, test plan, Developer, Agent) have at least one instance and all 15 predicates appear in at least one relationship.
