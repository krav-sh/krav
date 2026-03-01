# Knowledge graph

## Overview

The knowledge graph is the backbone of arci. It provides full traceability from stakeholder expectations through formalized requirements to verified implementations, enabling queries like "why does this requirement exist?", "what's blocking this task?", and "which requirements lack verification?".

Rather than organizing around documents (specs, plans, reports), arci uses a flat graph of typed nodes connected by semantic predicates. Modules define architectural hierarchy. Formal transformations connect concepts to needs to requirements. A task DAG expresses all work. The knowledge graph makes the relationships between these entities explicit, queryable, and enforceable.

## Foundations

The graph design draws on two primary standards:

**INCOSE Needs and Requirements Manual (NRM)**: Provides the conceptual framework for separating stakeholder expectations (needs) from design obligations (requirements), the formal transformation chain between them, and the distinction between validation ("did we capture the right thing?") and verification ("did we build it correctly?").

**ISO/IEC/IEEE 15288**: Provides the lifecycle process framework that organizes work into phases (architecture, design, implementation, integration, verification, validation) and defines the technical processes that produce and consume knowledge graph entities.

## Design principles

**Traceability**: Every requirement traces back to a stakeholder need, and every need traces back to a concept. The `derivesFrom` chain is unbroken. When something changes upstream, suspect links propagate downstream so reviewers can assess impact.

**Single source of truth**: The `graph.jsonlt` file is the canonical store for all structured metadata. No frontmatter, no duplicated data, no synchronization between separate stores. Prose lives in markdown files linked from graph nodes via the `content` property.

**Phase-gated execution**: Work is organized by lifecycle phase. Modules track their current phase; child modules cannot advance past their parent's phase. Tasks are tagged with the process phase they belong to. Phase advancement is an explicit operation with criteria (all tasks complete, no blocking defects).

**Graph-native relationships**: Relationships are first-class. They're embedded as JSON-LD properties on source nodes, carry semantic meaning (hierarchical, derivation, dependency, verification), and have structural constraints (tree, DAG, unrestricted) enforced by validation.

**Queryability**: The graph replaces document containers. "What's the plan for this module?" is a query over the task DAG. "What's blocking the release?" is a traversal. "Which requirements lack verification?" is a coverage analysis. All project questions reduce to graph queries.

## Graph design documents

These documents define the knowledge graph from an ontology engineering perspective. Read them in order:

| #   | Document                                            | What it covers                                                                              |
| --- | --------------------------------------------------- | ------------------------------------------------------------------------------------------- |
| 1   | [Competency questions](competency-questions.md)     | What must the graph answer? The requirements spec for the ontology.                         |
| 2   | [Schema](schema.md)                                 | What exists in the graph? Node types, identifiers, properties, JSON-LD vocabulary.          |
| 3   | [Predicates](predicates.md)                         | How are things connected? Every predicate with domain, range, cardinality, and constraints. |
| 4   | [Constraints](constraints.md)                       | What must always be true? Structural invariants and validation rules.                       |
| 5   | [Transformations](transformations.md)               | How do new things get created? Formal operations that produce nodes from existing ones.     |
| 6   | [Lifecycle coordination](lifecycle-coordination.md) | How do state changes propagate? Cross-entity lifecycle interactions.                        |
| 7   | [Query patterns](query-patterns.md)                 | How do we ask questions? Canonical graph traversal patterns.                                |

## Relationship to entity-specific docs

The graph design docs define the schema, relationships, and constraints that govern the knowledge graph as a whole. Entity-specific docs define the detailed semantics of each node type:

- [Concepts](../intent/concepts.md) and [Needs](../intent/needs.md) — intent capture entities
- [Modules](../requirements/modules.md) and [Requirements](../requirements/requirements.md) — requirements entities
- [Baselines](../requirements/baselines.md) — configuration management
- [Tasks](../execution/tasks.md) — execution entities
- [Verifications](../verification/verifications.md) and [Defects](../verification/defects.md) — verification entities

When these docs overlap, the entity-specific doc is authoritative for type-specific details (fields, lifecycle states, CLI commands), while the graph docs are authoritative for cross-entity concerns (predicates, constraints, transformations, query patterns).

## Relationship to implementation

These documents describe the graph *design* — what the graph contains and how it behaves. Implementation details (three-layer architecture, typed node classes, serialization dispatch, performance characteristics) are covered in entity-specific docs and the codebase itself.
