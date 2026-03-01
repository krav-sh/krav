---
name: arci-verify
description: >-
  Run test cases and record verification results. Use when test cases need
  to be executed and their results recorded on TC-* nodes.
---

# Run verification

Execute test cases and record results.

## Test cases

!`MOD_ID="$1"; jq -s --arg id "$MOD_ID" '
  [.[] | select(."@type" == "TestCase" and .module."@id" == $id)] as $tcs |
  [.[] | select(."@type" == "Requirement" and .module."@id" == $id)] as $reqs |
  {
    test_cases: [$tcs[] | {id: ."@id", title: .title, method: .method, currentResult: .currentResult, acceptanceCriteria: .acceptanceCriteria, status: .status}],
    requirements: [$reqs[] | {id: ."@id", title: .title, verifiedBy: .verifiedBy}]
  }
' .arci/graph.jsonlt 2>/dev/null || echo '{"error": "Provide a MOD-* identifier."}'`

## Instructions

1. Review the test cases for this module and their acceptance criteria.
2. For test cases with method `"test"`: run the test code and record pass/fail.
3. For method `"inspection"`: read the relevant artifacts and evaluate against criteria.
4. For method `"demonstration"`: execute the workflow and observe whether criteria pass.
5. For method `"analysis"`: perform the analysis described and evaluate the result.
6. Update each TC-* node's `currentResult` to `"pass"`, `"fail"`, or `"skip"` (with rationale for skips).
7. For failures, create a DEF-* node describing what failed and why.
