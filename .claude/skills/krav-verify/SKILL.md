---
name: krav-verify
description: >-
  Run test cases and record verification results. Use when test cases need
  to be executed and their results recorded on TC-* nodes.
stage-classification: temporary
replacement-stage: 1
replacement: "`krav tc record` CLI command"
---

# Run verification

Execute test cases and record results.

## Candidate picker

If the developer did not provide a MOD-* identifier, list modules that have executable test cases and ask them to pick one:

!`jq -s '
  [.[] | select(."@type" == "TestCase" and .status == "executable")] as $executable |
  [$executable[].module."@id"] | unique as $mod_ids |
  [.[] | select(."@id" == ($mod_ids[] // empty)) | {id: ."@id", title: .title, executable_count: ([$executable[] | select(.module."@id" == ."@id")] | length)}]
' .krav/graph.jsonlt 2>/dev/null || echo '[]'`

Present the candidates using the AskUserQuestion tool so the developer can select one. If the list is empty, no modules have executable test cases.

## Context

Once you have a MOD-* identifier, load test cases and their linked requirements:

!`MOD_ID="$1"; jq -s --arg id "$MOD_ID" '
  [.[] | select(."@type" == "TestCase" and .module."@id" == $id)] as $tcs |
  [.[] | select(."@type" == "Requirement" and .module."@id" == $id)] as $reqs |
  {
    test_cases: [$tcs[] | {id: ."@id", title: .title, method: .method, level: .level, status: .status, currentResult: .currentResult, acceptanceCriteria: .acceptanceCriteria, checklist: .checklist, procedure: .procedure, approach: .approach, verifies: .verifies}],
    requirements: [$reqs[] | {id: ."@id", title: .title, statement: .statement, verifiedBy: .verifiedBy}]
  }
' .krav/graph.jsonlt 2>/dev/null || echo '{"error": "Provide a MOD-* identifier."}'`

## Instructions

1. Review the test cases for this module. Focus on executable test cases (status `"executable"`). Skip test cases in draft or specified status since they aren't ready to run.
2. For each executable test case, execute according to its method:
   - `"test"`: Run the test code referenced in the test case's source field and record pass/fail.
   - `"inspection"`: Read the relevant artifacts and evaluate each checklist item against the acceptance criteria.
   - `"demonstration"`: Execute the workflow described in the procedure and observe whether criteria pass.
   - `"analysis"`: Perform the analysis described in the approach and evaluate the result.
3. For inspection test cases with checklists, evaluate each item individually and report which passed and which failed.
4. Before recording results to the graph, run the review loop (see below).
5. After incorporating review feedback, update each TC-* node: set `currentResult` to `"pass"`, `"fail"`, or `"skip"` (with rationale for skips), and set `lastRunAt` to the current timestamp.
6. For failures, create a DEF-* defect node describing what failed and why, with `subject` referencing the failed test case. Create a prose file at `.krav/defects/{timestamp}-{NANOID}-{slug}.md` with the full failure evidence, the expected vs. observed outcome, and any relevant context from the verification execution.

## Review loop

After evaluating the test cases but before recording results to the graph, use the Agent tool to review the judgments. Pass the agent the test cases, their acceptance criteria, the evaluation findings, and the proposed pass/fail results. Instruct the review agent:

"Review these verification judgments against the test case acceptance criteria. Check each of the following and report only problems found:

- For each pass judgment, is the evidence actually sufficient to satisfy the acceptance criteria? Flag any passes that seem generous or where the evidence is ambiguous.
- For each fail judgment, does the failure identification hold up? Could a different reading of the test case lead to a pass?
- For inspection checklists, did the evaluator cover all items or skip some without justification?
- Did anyone skip test cases without adequate rationale?
- Do the proposed defect descriptions for failures accurately capture what went wrong?

Do not create any graph nodes, tasks, or defects. Return only your critique."

## Graph-editing conventions

| Pattern | Classification | Stage | Replacement |
|---------|---------------|-------|-------------|
| Candidate picker: modules with executable test cases | Temporary | 1 | `krav tc list --status executable` CLI command |
| Module test case and requirement inventory query | Temporary | 1 | `krav tc record` CLI command |
