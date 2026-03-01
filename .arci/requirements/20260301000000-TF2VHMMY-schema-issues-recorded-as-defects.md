# Schema issues recorded as defects

The project shall record every ontology schema problem discovered during bootstrapping as a DEF-* defect node with category 'incorrect' or 'inconsistent' and subject referencing the affected schema element's documentation.

## Rationale

Ad-hoc fixes without formal tracking lose the learning. Defect nodes create an auditable record of ontology evolution and ensure schema problems receive deliberate disposition rather than silent workarounds.

## Verification criteria

Every schema change commit during bootstrapping has a corresponding DEF-* node in graph.jsonlt whose subject traces to the relevant schema documentation.
