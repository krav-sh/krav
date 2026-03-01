# Stage transition verifies temporary pattern removal

The project shall not declare a stage transition complete until its designated successor replaces every temporary pattern scoped to the departing stage, or a DEF-* defect node records a deferral rationale.

## Rationale

Without a gate, temporary patterns silently become permanent. The DEF-* deferral escape hatch avoids demanding perfection while ensuring deliberate disposition. This exercises the same defect workflow required by REQ-TF2VHMMY.

## Verification criteria

At each stage transition, demonstrate that every temporary pattern marked for the departing stage either has a working replacement or a DEF-* node in graph.jsonlt recording the deferral rationale.
