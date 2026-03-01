# Knowledge graph as single source of truth

The central architectural decision in arci is that all structured metadata about a project lives in the knowledge graph, not scattered across document frontmatter, separate databases, issue trackers, or ad-hoc files. The graph stores typed nodes (concepts, modules, needs, requirements, test cases, tasks, defects, baselines) connected by semantic predicates, and views on the graph replace traditional documents.

## What this means in practice

A requirements specification is a query: "all REQ-* nodes where module is MOD-parser, sorted by priority." A test plan is a query: "all TC-* nodes grouped by verification level." A traceability matrix is a bipartite projection of REQ ↔ TC edges. None of these are documents that someone maintains separately from the data they describe.

The graph uses JSON-LD compact form serialized as JSON Lines (JSONLT). Each line in `graph.jsonlt` is a self-contained JSON-LD document representing one node. The JSON-LD context maps compact JSON keys to full RDF IRIs, giving the graph formal semantics while keeping the on-disk format human-editable.

## Why JSON-LD and not just JSON

JSON-LD buys two things. First, formal vocabulary alignment: arci's node types and predicates map to established ontologies (Dublin Core for metadata, PROV-O for provenance, OSLC for requirements management) via `rdfs:subClassOf` and `rdfs:subPropertyOf` declarations. This isn't academic; arci's graph has well-defined semantics that tools can reason about.

Second, the `@context` mechanism separates compact keys from their full meaning. The graph file uses short keys like `title`, `derivesFrom`, and `module`, but these expand to `dcterms:title`, `arci:derivesFrom`, and `arci:module` when processed as RDF. This keeps the file readable while maintaining formal rigor.

## What this replaces

Traditional document-centric requirements management maintains separate SRS documents, test plans, traceability matrices, and V&V plans as independent artifacts. Keeping these in sync is the classic document management problem. arci eliminates it by making all these "documents" views on the same underlying data.
