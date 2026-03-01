# Graph overview

The Krav knowledge graph stores typed nodes connected by semantic predicates. It is the source of truth for what the project must build and why. The runtime engine is DuckDB with the DuckPGQ extension, providing both standard SQL for filtering and aggregation and SQL/PGQ (SQL:2023 standard) for property graph pattern matching. On-disk serialization uses per-table NDJSON files under `.krav/graph/` for git-friendly version control.

## Reading order

Start with the schema for node types and the storage model, then predicates for relationships, then constraints for structural invariants. Query patterns document how to interrogate the graph using SQL and SQL/PGQ.

1. [Schema](schema.md): node types, identifier scheme, properties, storage model, NDJSON representation
2. [Predicates](predicates.md): 15 semantic relationships with domain/range matrix, directionality, suspect propagation
3. [Constraints](constraints.md): structural invariants (trees, DAGs, referential integrity, phase gates)
4. [Query patterns](query-patterns.md): canonical SQL and SQL/PGQ queries for traceability, coverage, impact analysis
5. [Vocabulary alignment](vocabulary-alignment.md): design-time alignment with Dublin Core, PROV-O, and OSLC ontologies

## Node type documentation

Each node type has a detailed specification:

- [Concepts](nodes/concepts.md) (CON-\*): exploration, design decisions, crystallized thinking
- [Modules](nodes/modules.md) (MOD-\*): architectural containers with hierarchy and phase tracking
- [Needs](nodes/needs.md) (NEED-\*): stakeholder expectations
- [Requirements](nodes/requirements.md) (REQ-\*): design obligations with verification criteria
- [Test cases](nodes/test-cases.md) (TC-\*): verification specifications
- [Tasks](nodes/tasks.md) (TASK-\*): atomic work units in a dependency DAG
- [Defects](nodes/defects.md) (DEF-\*): identified problems with disposition and resolution
- [Baselines](nodes/baselines.md) (BSL-\*): named graph state snapshots anchored to commits
- [Stakeholders](nodes/stakeholders.md) (STK-\*): named parties with concerns
- [Developers](nodes/developers.md) (DEV-\*): human actors with persistent identity
- [Agents](nodes/agents.md) (AGT-\*): Claude Code sessions and subagents

## Related documentation

- [Transformations](transformations.md): formal transformation chain (concept → need → requirement → task)
- [Lifecycle coordination](lifecycle-coordination.md): phase-gated execution and suspect link propagation
- [Competency questions](competency-questions.md): the questions the graph must answer
