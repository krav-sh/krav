# Documents as graph views

Contributors need traditional SE documents to derive from the same data they describe, so that the project eliminates separate synchronized artifacts rather than automating them.

## The synchronization problem

Requirements specifications, test plans, traceability matrices, and verification reports are standard systems engineering deliverables. In document-centric approaches, each is a separate artifact that a person or team maintains. The classic problem is keeping them in sync: when a requirement changes, the SRS, test plan, and traceability matrix all need updating. Automation can reduce the effort, but the fundamental problem remains that the same information exists in multiple places.

The alternative is that these documents don't exist as maintained artifacts at all. A requirements specification is a query against the graph filtered by module. A test plan is a query grouped by verification level. A traceability matrix is a projection of edges between requirement and test case nodes. The query computes the document on demand, so it cannot drift from the data it describes.

## What elimination means versus automation

This need deliberately says "eliminated rather than automated." Automating synchronization (generating documents from a database, then checking them in) still creates artifacts that can become stale if someone skips the generation step or it fails. Elimination means the document format is a view: it stays current because it reads from the source data at query time, not at generation time.
