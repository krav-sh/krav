# Ad-hoc queryable structured data

The developer needs to ask ad-hoc questions of the project's structured data (filtering by module, grouping by verification level, projecting traceability edges) so that information retrieval is a query, not a document search.

## Beyond predefined reports

Predefined reports answer the questions someone anticipated. Ad-hoc queries answer the questions that arise during actual work: "which requirements in this module lack test coverage?" or "what tasks block the next phase gate?" or "which defects did someone file against nodes that changed since the last baseline?" These questions arise in the moment and no one can anticipate them all in advance.

The graph's structure makes these queries straightforward. Typed nodes and semantic predicates mean that filtering, grouping, and traversal operations have well-defined meanings. A query for "REQ-* nodes where module is MOD-QRN0SQCF and verifiedBy is empty" produces an unambiguous result because `module`, `verifiedBy`, and the node types all belong to the schema.

## Value for AI agents

Ad-hoc queryability becomes particularly valuable when AI agents interact with the project. An agent composing a `jq` query against the graph can answer traceability and status questions programmatically, without parsing prose documents or searching file hierarchies. The structured data becomes a reliable API for agent-driven workflows, enabling skills and subagents to make informed decisions based on current project state.
