# Single authoritative data store

Contributors need all project metadata to live in a single authoritative location, so that the project's architecture has one clear source of truth rather than scattered artifacts.

## Why this matters

When project metadata spans document frontmatter, issue trackers, separate databases, and ad-hoc files, no single source is authoritative. Contributors waste time reconciling contradictions between sources, and tooling must integrate with multiple backends to assemble a complete picture. A single authoritative store eliminates this class of problem entirely.

The cost of scattering is particularly acute for traceability. If requirements live in one system, test cases in another, and defects in a third, answering "is this requirement verified?" requires joining across all three. A single store makes this a local query.

## Stakeholder perspective

The arci contributor (STK-C2B6R1GY) cares about clean architecture and clear module boundaries. A single data store is the foundation of both: the architecture has one well-defined persistence layer, and module boundaries take the form of relationships in the same graph rather than conventions scattered across file hierarchies, issue labels, and configuration files.
