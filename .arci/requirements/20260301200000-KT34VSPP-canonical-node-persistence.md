# Canonical node persistence

The graph engine shall persist every graph node to a single JSONLT store such that no node type requires a separate persistence mechanism and no structured metadata about a node exists outside the store.

## Context

This requirement addresses the core concern of NEED-7DT7XGE6: that project metadata should live in a single authoritative location. The originating concept (CON-A8HG1CNG) establishes that all structured metadata lives in the knowledge graph rather than in document frontmatter, separate databases, or ad-hoc files. This requirement constrains the graph engine to deliver on that architectural commitment.

The requirement is deliberately framed around the behavioral property (all structured metadata in one store) rather than prescribing a specific file format or serialization strategy. The JSONLT format is an architectural choice; the obligation is that whatever persistence mechanism the engine uses, it must be singular and complete. No node type gets special treatment with a sidecar database, and no structured field lives only in a prose file's frontmatter.

## Verification approach

Create nodes of all 12 types defined in the ontology (Concept, Module, Need, Requirement, test case, Task, Defect, Baseline, Stakeholder, test plan, Developer, Agent). For each, verify it can be fully reconstructed from the JSONLT store alone. Confirm that retrieving any node's structured properties needs no external files, caches, or databases.

## Relationship to sibling requirements

REQ-HYGZ527A (self-contained node encoding) constrains the encoding of individual nodes within the store. REQ-S3HV9ZRA (canonical vocabulary context) constrains the vocabulary mapping. This requirement constrains the store's completeness and exclusivity: there is one store, and it contains everything.
