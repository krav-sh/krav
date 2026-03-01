# Canonical vocabulary context

Each node's `@context` shall resolve to the arci vocabulary (namespace `https://arci.dev/schema#`), mapping all compact property names used in the node to their canonical IRIs in the arci, Dublin Core, PROV-O, or OSLC namespaces. The context shall be resolvable without network access.

## Context

CON-A8HG1CNG identifies formal vocabulary alignment as a primary reason for choosing JSON-LD over plain JSON. The arci schema maps its types and predicates to established vocabularies: `title` maps to `dcterms:title`, `attributedTo` maps to `prov:wasAttributedTo`, `implements` maps to `oslc_cm:implementsRequirement`. This alignment is not decorative. It positions arci's data to interoperate with OSLC-compatible tools and lets standard RDF tooling query it directly.

This requirement constrains the `@context` to resolve to the canonical arci vocabulary rather than permitting arbitrary IRI mappings. A node with `{"@context": {"title": "http://example.com/foo"}}` would satisfy a weaker "maps to IRIs" requirement but would violate the architectural intent of vocabulary alignment.

The offline-resolvable constraint ensures the graph remains self-describing even without network access. The `@context` value may be a relative path to a bundled context file (such as `context.jsonld`) or an inline context object, but it must not require HTTP resolution at read time.

## Verification approach

For each node in the graph, resolve its `@context` and verify that: (1) every compact property name used in the node maps to an IRI in the `https://arci.dev/schema#`, Dublin Core, PROV-O, or OSLC namespaces, and (2) the resolution succeeds with no network I/O (mock or block network access during the test).
