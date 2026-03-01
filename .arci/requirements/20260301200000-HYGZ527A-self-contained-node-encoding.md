# Self-contained node encoding

The graph engine shall encode each graph node as exactly one newline-delimited record in the JSONLT file, where each record is valid JSON and valid JSON-LD compact form including a `@context` reference.

## Context

The JSONLT format is central to arci's persistence strategy. Any tool can parse, validate, and process each line independently as a self-contained document. This property enables line-oriented tooling (`grep`, `jq`, `sed`) during Stage 0 when the graph engine doesn't yet exist, and it enables streaming processing in later stages.

The "exactly one line" constraint means literal newlines must not appear in the serialized JSON. JSON string values may contain `\n` escape sequences, but the serialized record itself occupies a single line ending with a newline character. This is the standard JSON Lines convention.

The `@context` requirement ensures every line is interpretable as JSON-LD without relying on out-of-band knowledge. A tool reading line 47 of graph.jsonlt can resolve the compact property names to their full IRIs using only information present in that line (via the `@context` reference).

## Verification approach

Parse every line of `graph.jsonlt` independently. Each line must satisfy three checks: (1) valid JSON per RFC 8259, (2) valid JSON-LD compact form per the JSON-LD 1.1 specification, and (3) contains an `@context` key. Lines that fail any check are violations.
