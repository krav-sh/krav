# Permanent patterns identified in design documentation

The project shall maintain a section in the design documentation that identifies which bootstrapping conventions are permanent (survive all stages) and which are temporary (scoped to a specific stage), updated at each stage boundary.

## Rationale

Without a single reference distinguishing scaffolding from load-bearing patterns, contributors must infer permanence from context. A maintained list reduces ambiguity and provides a checkpoint for stage transitions. This requirement relies on manual discipline until Stage 2 hooks can enforce documentation freshness.

## Verification criteria

The design documentation contains a section distinguishing permanent from temporary conventions, and the list covers all patterns currently in use at the time of the most recent stage transition.
