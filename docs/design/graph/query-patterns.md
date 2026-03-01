# Query patterns

## Overview

This document catalogs canonical graph traversal patterns that realize the [competency questions](competency-questions.md). Patterns use SQL and SQL/PGQ syntax against the DuckDB + DuckPGQ runtime. Each pattern describes the intent, the query, and complexity characteristics. Standard SQL handles filtering, aggregation, and set operations. SQL/PGQ handles graph traversals, variable-length paths, and pattern matching.

## Traceability traversals

### Ancestry (CQ-T1, CQ-T5)

Traverse `derives_from` edges from a node to its sources, recursively. Answers "why does this exist?"

```sql
FROM GRAPH_TABLE (knowledge_graph
  MATCH (downstream)-[d:derives_from]->{1,5}(upstream)
  WHERE downstream.id = 'REQ-C2H6N4P8'
  COLUMNS (downstream.id AS start_id, upstream.id AS ancestor_id, upstream.type AS ancestor_type)
);
```

The `{1,5}` quantifier follows `derives_from` edges one to five hops. Derivation chains are typically short (2-4 hops: REQ → NEED → CON).

**Starting point**: any NEED or REQ node.
**Predicates**: `derives_from` (backward through chain).
**Terminates at**: CON nodes (concepts have no outgoing `derives_from` edges).
**Complexity**: O(n) where n is the length of the derivation chain.

### Descendancy (CQ-T2, CQ-T4)

Traverse `derives_from` edges in reverse to find all nodes that derive from a given source. Answers "what did this produce?"

```sql
FROM GRAPH_TABLE (knowledge_graph
  MATCH (downstream)-[d:derives_from]->{1,5}(upstream)
  WHERE upstream.id = 'CON-K7M3NP2Q'
  COLUMNS (downstream.id AS descendant_id, downstream.type AS descendant_type)
);
```

**Starting point**: any CON, NEED, or REQ node.
**Predicates**: `derives_from` inverse.
**Complexity**: O(n) where n is the number of nodes in the derivation subgraph.

### Full traceability chain (CQ-T5, CQ-X1)

Combines ancestry and descendancy with related verifications, defects, and tasks. The complete context for a node. This pattern combines SQL/PGQ for the derivation chain with standard SQL joins for related entities.

```sql
-- Ancestors via SQL/PGQ
WITH ancestors AS (
  FROM GRAPH_TABLE (knowledge_graph
    MATCH (downstream)-[d:derives_from]->{1,5}(upstream)
    WHERE downstream.id = ?
    COLUMNS (upstream.id AS id, upstream.type AS type)
  )
),
-- Descendants via SQL/PGQ
descendants AS (
  FROM GRAPH_TABLE (knowledge_graph
    MATCH (downstream)-[d:derives_from]->{1,5}(upstream)
    WHERE upstream.id = ?
    COLUMNS (downstream.id AS id, downstream.type AS type)
  )
),
-- Related verifications via standard SQL
verifications AS (
  SELECT tc.id, tc.title, tc.current_result
  FROM verified_by vb
  JOIN test_cases tc ON tc.id = vb.dst
  WHERE vb.src = ?
),
-- Related defects via standard SQL
related_defects AS (
  SELECT d.id, d.title, d.severity, d.status
  FROM subject s
  JOIN defects d ON d.id = s.src
  WHERE s.dst = ?
)
SELECT * FROM ancestors
UNION ALL SELECT * FROM descendants;
```

**Complexity**: O(n + m) where n is the derivation subgraph size and m is the number of related entities.

## Hierarchy traversals

### Module subtree (CQ-P5, CQ-PH2)

Traverse `child_of` edges in reverse to find all descendants of a module.

```sql
FROM GRAPH_TABLE (knowledge_graph
  MATCH (child)-[c:child_of]->{1,10}(parent)
  WHERE parent.id = 'MOD-OAPSROOT'
  COLUMNS (child.id AS module_id, child.title AS module_title, child.phase AS phase)
);
```

**Starting point**: any MOD node.
**Predicates**: `child_of` inverse.
**Complexity**: O(n) where n is the number of modules in the subtree.

### Scoped subgraph (CQ-P5)

All nodes owned by a module subtree. Combines module subtree traversal with `module` edge table joins.

```sql
WITH subtree AS (
  SELECT 'MOD-OAPSROOT' AS id
  UNION ALL
  FROM GRAPH_TABLE (knowledge_graph
    MATCH (child)-[c:child_of]->{1,10}(parent)
    WHERE parent.id = 'MOD-OAPSROOT'
    COLUMNS (child.id AS id)
  )
)
SELECT m.dst AS node_id
FROM module m
WHERE m.dst IN (SELECT id FROM subtree);
```

**Starting point**: any MOD node.
**Predicates**: `child_of` inverse, then `module` inverse.
**Complexity**: O(n) where n is the total number of nodes in the graph. The module edge table index speeds this up.

## Dependency traversals

### Critical path (CQ-P2)

The longest path through incomplete tasks in the `depends_on` DAG leading to a target task. This pattern identifies the sequence of tasks that determines the minimum time to completion.

```sql
FROM GRAPH_TABLE (knowledge_graph
  MATCH p = ANY LONGEST (blocker:tasks)-[d:depends_on]->{1,50}(target:tasks)
  WHERE target.id = ?
    AND blocker.status != 'complete'
  COLUMNS (blocker.id AS task_id, blocker.title AS task_title, blocker.status AS status)
);
```

**Starting point**: a milestone TASK node.
**Predicates**: `depends_on` (reverse traversal from target to roots).
**Complexity**: O(V + E) where V is the number of tasks and E is the number of dependency edges.

### Blocking set (CQ-P3)

All incomplete tasks that transitively block a given task. SQL/PGQ variable-length paths handle the transitive closure.

```sql
FROM GRAPH_TABLE (knowledge_graph
  MATCH (blocker:tasks)-[d:depends_on]->{1,50}(target:tasks)
  WHERE target.id = ?
    AND blocker.status != 'complete'
  COLUMNS (blocker.id AS task_id, blocker.title AS task_title, blocker.status AS status)
);
```

**Starting point**: any TASK node.
**Predicates**: `depends_on`.
**Complexity**: O(n) where n is the size of the transitive dependency closure.

### Ready set (CQ-P1)

Tasks that are not complete and have all dependencies satisfied. Standard SQL with `LEFT JOIN` and `NOT EXISTS` expresses this more naturally than SQL/PGQ.

```sql
SELECT t.id, t.title, t.process_phase
FROM tasks t
WHERE t.status NOT IN ('complete', 'cancelled')
  AND NOT EXISTS (
    SELECT 1
    FROM depends_on dep
    JOIN tasks blocker ON blocker.id = dep.dst
    WHERE dep.src = t.id
      AND blocker.status != 'complete'
  );
```

**Starting point**: all tasks.
**Predicates**: `depends_on`.
**Complexity**: O(V + E); scan all tasks, check each dependency.

## Coverage analysis

### Unverified requirements (CQ-C1)

Requirements with no `verified_by` edges. Standard SQL with `LEFT JOIN`.

```sql
SELECT r.id, r.title, r.status
FROM requirements r
LEFT JOIN verified_by vb ON vb.src = r.id
WHERE vb.src IS NULL;
```

**Starting point**: all REQ nodes.
**Predicates**: `verified_by`.
**Complexity**: O(n) where n is the number of requirements.

### Coverage ratio (CQ-C2)

Verification coverage for a module, expressed as a ratio. Combines module subtree traversal with aggregation.

```sql
WITH subtree AS (
  SELECT 'MOD-OAPSROOT' AS id
  UNION ALL
  FROM GRAPH_TABLE (knowledge_graph
    MATCH (child)-[c:child_of]->{1,10}(parent)
    WHERE parent.id = 'MOD-OAPSROOT'
    COLUMNS (child.id AS id)
  )
)
SELECT
  count(DISTINCT r.id) AS total_requirements,
  count(DISTINCT CASE WHEN vb.src IS NOT NULL THEN r.id END) AS verified,
  count(DISTINCT CASE WHEN tc.current_result = 'pass' THEN r.id END) AS passing
FROM requirements r
JOIN module m ON m.src = r.id AND m.dst IN (SELECT id FROM subtree)
LEFT JOIN verified_by vb ON vb.src = r.id
LEFT JOIN test_cases tc ON tc.id = vb.dst;
```

**Starting point**: any MOD node.
**Predicates**: `module` inverse, `verified_by`, test case status.
**Complexity**: O(n) where n is the number of requirements in the module subtree.

## Impact analysis

### Suspect link propagation (CQ-I1, CQ-I2)

When someone modifies a node, find all edges that become suspect. SQL/PGQ pattern matching identifies the affected relationships.

```sql
-- For a modified concept: find needs that derive from it
FROM GRAPH_TABLE (knowledge_graph
  MATCH (n:needs)-[d:derives_from]->(c:concepts)
  WHERE c.id = ?
  COLUMNS (n.id AS affected_id, 'derives_from' AS edge_type)
);

-- For a modified requirement: find child requirements, verifications, and allocations
SELECT df.src AS affected_id, 'derives_from' AS edge_type
FROM derives_from df WHERE df.dst = ?
UNION ALL
SELECT vb.src AS affected_id, 'verified_by' AS edge_type
FROM verified_by vb WHERE vb.src = ?
UNION ALL
SELECT at.src AS affected_id, 'allocates_to' AS edge_type
FROM allocates_to at WHERE at.src = ?;
```

**Starting point**: the modified node.
**Predicates**: `derives_from` inverse, `verified_by`, `allocates_to`.
**Complexity**: O(d) where d is the out-degree of the modified node's inverse relationships. Non-transitive by default (see [constraints](constraints.md) C-SUSPECT2).

### What-if analysis

Simulate the impact of a proposed change without applying it. Combines suspect link propagation with verification lookup.

```sql
WITH suspect_edges AS (
  -- Same suspect link propagation query as above
  SELECT df.src AS node_id, 'derives_from' AS edge_type FROM derives_from df WHERE df.dst = ?
  UNION ALL
  SELECT vb.dst AS node_id, 'verified_by' AS edge_type FROM verified_by vb WHERE vb.src = ?
  UNION ALL
  SELECT at.dst AS node_id, 'allocates_to' AS edge_type FROM allocates_to at WHERE at.src = ?
)
SELECT se.node_id, se.edge_type, tc.id AS affected_test, tc.current_result
FROM suspect_edges se
LEFT JOIN verified_by vb ON vb.src = se.node_id
LEFT JOIN test_cases tc ON tc.id = vb.dst;
```

**Complexity**: same as suspect link propagation plus verification lookup.

## Phase gate queries

### Advancement readiness (CQ-PH1)

Check whether a module can advance to the next phase. Standard SQL aggregation.

```sql
SELECT
  m.id AS module_id,
  m.phase AS current_phase,
  (SELECT count(*) FROM tasks t
   JOIN module mo ON mo.src = t.id AND mo.dst = m.id
   WHERE t.process_phase = m.phase AND t.status != 'complete') AS incomplete_tasks,
  (SELECT count(*) FROM defects d
   JOIN module mo ON mo.src = d.id AND mo.dst = m.id
   WHERE d.severity IN ('critical', 'major')
   AND d.status IN ('open', 'confirmed')) AS blocking_defects
FROM modules m
WHERE m.id = ?;
```

A module can advance when `incomplete_tasks = 0` and `blocking_defects = 0`.

**Starting point**: any MOD node.
**Predicates**: `module` inverse (for tasks and defects).
**Complexity**: O(t + d) where t is the number of tasks for this module/phase and d is the number of defects for this module.

### Phase comparison across siblings (CQ-PH2)

Find modules at an earlier phase than their siblings.

```sql
WITH sibling_phases AS (
  SELECT
    co.dst AS parent_id,
    m.id AS module_id,
    m.phase,
    max(m.phase) OVER (PARTITION BY co.dst) AS max_sibling_phase
  FROM modules m
  JOIN child_of co ON co.src = m.id
)
SELECT module_id, phase, parent_id, max_sibling_phase
FROM sibling_phases
WHERE phase < max_sibling_phase;
```

**Starting point**: all MOD nodes with children.
**Predicates**: `child_of`.
**Complexity**: O(n) where n is the number of modules.

## Structural integrity queries

### Orphan detection (CQ-S1)

Find nodes with no structural connections. Standard SQL with `LEFT JOIN` across all edge tables.

```sql
-- Example for requirements: find those with no incoming or outgoing edges
SELECT r.id, r.title
FROM requirements r
LEFT JOIN module m ON m.src = r.id
LEFT JOIN derives_from df_out ON df_out.src = r.id
LEFT JOIN derives_from df_in ON df_in.dst = r.id
LEFT JOIN verified_by vb ON vb.src = r.id
LEFT JOIN allocates_to at ON at.src = r.id
WHERE m.src IS NULL
  AND df_out.src IS NULL
  AND df_in.dst IS NULL
  AND vb.src IS NULL
  AND at.src IS NULL;
```

Run similar queries per vertex table. Exclude the root MOD (which has no `child_of` edge by definition).

**Complexity**: O(n + e) where n is the number of nodes and e is the number of edges.

### Cycle detection (CQ-S3, CQ-S4)

Detect cycles in DAG-constrained subgraphs. SQL/PGQ path queries with self-referencing patterns detect cycles directly.

```sql
-- Detect cycles in derives_from
FROM GRAPH_TABLE (knowledge_graph
  MATCH (a)-[d:derives_from]->{1,50}(b)
  WHERE a.id = b.id
  COLUMNS (a.id AS cyclic_node)
);
```

**Applied to**: `derives_from` subgraph, `depends_on` subgraph.
**Complexity**: O(V + E) for each subgraph.

### Dangling reference detection (CQ-S2)

Find edge table entries that reference nonexistent nodes. Standard SQL with `LEFT JOIN`.

```sql
-- Check derives_from for dangling references
SELECT df.src, df.dst, 'derives_from' AS edge_type
FROM derives_from df
LEFT JOIN needs n ON n.id = df.src
LEFT JOIN requirements r ON r.id = df.src
LEFT JOIN concepts c ON c.id = df.dst
LEFT JOIN needs n2 ON n2.id = df.dst
LEFT JOIN requirements r2 ON r2.id = df.dst
WHERE (n.id IS NULL AND r.id IS NULL)
   OR (c.id IS NULL AND n2.id IS NULL AND r2.id IS NULL);
```

**Complexity**: O(e) where e is the number of edges. Run per edge table.

## Baseline queries

### Semantic diff (CQ-A1)

Compare graph state between two baselines. ARCI reconstructs historical graph state by checking out the baseline's commit and reading the NDJSON files from that commit. DuckDB's `read_json` function loads historical NDJSON directly for comparison.

```sql
-- Load graph state from two commits into temporary tables, then diff
WITH current AS (
  SELECT * FROM requirements
),
baseline AS (
  SELECT * FROM read_json('.arci/graph/requirements.ndjson')  -- at baseline commit
)
SELECT 'added' AS change, c.id, c.title FROM current c
  WHERE c.id NOT IN (SELECT id FROM baseline)
UNION ALL
SELECT 'removed' AS change, b.id, b.title FROM baseline b
  WHERE b.id NOT IN (SELECT id FROM current)
UNION ALL
SELECT 'modified' AS change, c.id, c.title FROM current c
  JOIN baseline b ON b.id = c.id
  WHERE c.status != b.status OR c.title != b.title;
```

**Starting point**: two BSL nodes.
**Complexity**: O(n) where n is the total number of nodes across both graphs. The cost of checking out and reading NDJSON from the baseline commit dominates.

### Deferred defects by target (CQ-A5)

Find deferred defects that target a specific baseline or phase. Standard SQL.

```sql
SELECT d.id, d.title, d.severity, d.deferral_target
FROM defects d
WHERE d.status = 'deferred'
  AND d.deferral_target = ?;
```

**Complexity**: O(d) where d is the number of deferred defects.
