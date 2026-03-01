---
name: arci-testcase
description: >-
  Create test case specifications linked to requirements. Use when
  requirements need verification cases defined, specifying what to check
  and how, without implementing the tests themselves.
stage-classification: temporary
replacement-stage: 3
replacement: "Production arci-testcase skill backed by `arci tc create` CLI command"
---

# Create test cases

Specify verification for requirements by creating TC-* nodes.

## Candidate picker

If the developer did not provide a MOD-* identifier, list modules that have requirements lacking test cases and ask the developer to pick one:

!`jq -s '
  [.[] | select(."@type" == "Requirement" and (.verifiedBy == null or (.verifiedBy | length) == 0))] as $uncovered |
  [$uncovered[].module."@id"] | unique as $mod_ids |
  [.[] | select(."@id" == ($mod_ids[] // empty)) | {id: ."@id", title: .title, uncovered_count: ([$uncovered[] | select(.module."@id" == ."@id")] | length)}]
' .arci/graph.jsonlt 2>/dev/null || echo '[]'`

Present the candidates using the AskUserQuestion tool so the developer can select one. If the list is empty, all requirements have test cases (or there are no requirements yet).

## Context

After identifying the MOD-* identifier, load its context:

!`MOD_ID="$1"; jq -s --arg id "$MOD_ID" '
  (.[] | select(."@id" == $id)) as $mod |
  [.[] | select(."@type" == "Requirement" and .module."@id" == $id)] as $reqs |
  [.[] | select(."@type" == "TestCase" and .module."@id" == $id)] as $tcs |
  {
    module: {id: $mod."@id", title: $mod.title},
    requirements: [$reqs[] | {id: ."@id", title: .title, statement: .statement, verifiedBy: .verifiedBy}],
    existing_test_cases: [$tcs[] | {id: ."@id", title: .title, method: .method, status: .status}]
  }
' .arci/graph.jsonlt 2>/dev/null || echo '{"error": "Provide a MOD-* identifier."}'`

## Instructions

1. Identify requirements that lack test cases (no `verifiedBy` edges).
2. For each, draft a TC-* node with: `module`, `method` (test, inspection, demonstration, or analysis), `level` (unit, integration, system, acceptance), `acceptanceCriteria` (explicit pass/fail), `verifies` edges to the requirements, and `status: "draft"`.
3. Choose the verification method based on the requirement: behavioral requirements → test, documentation requirements → inspection, user-facing workflows → demonstration, performance projections → analysis.
4. This skill creates specifications only. Implementation is a separate verification-phase task.
5. Before writing to the graph, run the review loop (see below).
6. Incorporate review feedback, then present the final test cases to the developer for approval.
7. Write approved test cases to `graph.jsonlt` and add corresponding `verifiedBy` edges on the requirement nodes.

## Review loop

After drafting the test cases but before writing them to the graph, use the Agent tool to review them. Pass the agent the drafted test cases, the requirements they verify, and the existing test cases in the module. Instruct the review agent:

"Review these drafted test cases against the requirements they verify. Check each of the following and report only problems found:

- Does each test case fully cover the requirement's statement, or are there clauses in the requirement that no test case checks?
- Is the verification method (test, inspection, demonstration, analysis) appropriate for what the requirement actually asks?
- Are the acceptance criteria concrete enough to produce an unambiguous pass/fail, or could reasonable people disagree on the outcome?
- Do any test cases go beyond the requirement and test things not actually required? If so, flag the requirement as potentially underspecified.
- Are there requirements in the module that still lack any test case coverage after this batch?

Do not create any graph nodes, tasks, or defects. Return only your critique."

## Graph-editing conventions

| Pattern | Classification | Stage | Replacement |
|---------|---------------|-------|-------------|
| Candidate picker: modules with uncovered requirements | Temporary | 3 | `arci req list --uncovered` CLI command |
| Module requirements and test case inventory query | Temporary | 3 | `arci tc create` CLI command |
