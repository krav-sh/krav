---
name: arci-testcase
description: >-
  Create test case specifications linked to requirements. Use when
  requirements need verification cases defined, specifying what to check
  and how, without implementing the tests themselves.
---

# Create test cases

Specify verification for requirements by creating TC-* nodes.

## Context

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
2. For each, create a TC-* node with: `module`, `method` (test, inspection, demonstration, or analysis), `level` (unit, integration, system, acceptance), `acceptanceCriteria` (explicit pass/fail), `verifies` edges to the requirements, and `status: "draft"`.
3. Add corresponding `verifiedBy` edges on the requirement nodes.
4. Choose the verification method based on the requirement: behavioral requirements → test, documentation requirements → inspection, user-facing workflows → demonstration, performance projections → analysis.
5. This skill creates specifications only. Implementation is a separate verification-phase task.
