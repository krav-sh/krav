# Creating test cases

## What

The developer says `write tests for these requirements` or `test cases are needed for the parser` or `check verification coverage`. The agent reads requirements that lack verification, creates TC-* nodes specifying what to verify and how, then writes the actual test code as a deliverable.

## Why

Test cases close the traceability loop. Without them, requirements are claims without evidence. The decoupled model (specification as TC-* nodes, code as task deliverables) means Krav can track coverage before any tests exist. You know which requirements have planned verification even if the test code doesn't exist yet, and the agent can plan the test code as downstream work.

## What happens in the graph

The agent runs a coverage analysis to identify requirements without test cases (or with insufficient coverage). For each gap, it creates a TC-* node specifying the verification method, acceptance criteria, and which requirements it verifies via `verifies` edges.

The test case starts in `draft`, moves to `specified` when the agent defines acceptance criteria, then to `implemented` when test code exists, and to `executable` when the test can actually run.

Writing the test code is a separate task (a verification task). The task's deliverable includes the test path, which gets recorded in the test case's `implementation` field.

## Trigger patterns

`Write tests for X`, `check coverage`, `which requirements don't have tests?`, `create test cases for the parser`, `verification is needed for these requirements`.

## Graph before

REQ-* nodes that lack `verifiedBy` edges (or have incomplete coverage).

## Graph after

TC-* nodes with `verifies` edges to requirements, acceptance criteria, verification methods, and level designations. Possibly TASK-* nodes for writing the test code.

## Agent interaction layer

### Skills

The `krav:testcase` skill builds this workflow. Preprocessing loads the module's requirements that lack `verifiedBy` edges (or have incomplete coverage), along with any existing TC-* nodes so the agent can see what's already specified. The coverage gap analysis happens at preprocessing time, so the agent starts with a clear picture of what needs test cases rather than having to compute it.

The skill's instructed commands create TC-* nodes via the CLI, set `verifies` edges to requirements, and define acceptance criteria and verification methods. The skill instructions emphasize the separation between specification (this workflow) and the coding step (a subsequent `implement-tests` task).

### Policies

The `mutation-feedback` policy fires after the agent creates each TC-* node, injecting updated coverage metrics: how many requirements now have test cases, what the verification coverage looks like as a whole. This running coverage view helps the agent decide when it has specified enough test cases.

The `prompt-context`, `graph-integrity`, and `cli-auto-approve` policies apply as in other graph-building workflows.

### Task types

This workflow produces test case specifications, not test code. The actual test writing comes from `implement-tests` tasks, which the `krav:decompose` skill creates as downstream work. Once test code exists, `execute-tests` tasks (run inside the `krav-verifier` subagent) exercise them. The test-case workflow may also trigger decomposition if the developer wants to immediately plan the work for newly created test cases.

## Open questions

**Specification-first or code-first?** Should the agent create TC-* specification nodes and then write the test code? Or write the test code and create the TC-* nodes to match? The design says specification-first, but in practice developers often write tests and then want them tracked in the graph retroactively.

**How does the agent pick verification methods?** The four INCOSE methods (test, inspection, demonstration, analysis) each suit different requirement types. The agent needs to match method to requirement. Behavioral requirements lead to test. Documentation requirements lead to inspection. UX requirements lead to demonstration. Performance projections lead to analysis. The project should codify this mapping somewhere the agent can reference.

**Test level determination.** Should a test case be unit, integration, system, or acceptance level? The module hierarchy gives hints (component module leads to unit, subsystem leads to integration, root leads to system), but it's not automatic. A component requirement about API contracts might need an integration-level test.

**How does existing test code get linked?** If the project already has tests that Krav did not create, how does the agent discover and link them to requirements? This is the "brownfield" problem: the graph is empty but the codebase has tests.

**Acceptance criteria quality.** The agent needs to write acceptance criteria that are specific enough to be useful ("p99 latency < 50 ms across 1000 iterations") not vague ("should be fast"). What makes this hard is that good acceptance criteria require domain understanding. The agent gets this from the requirement's `verificationCriteria` field, but that field might itself be vague.
