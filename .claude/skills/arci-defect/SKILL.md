---
name: arci-defect
description: >-
  Triage and fix defects. Use when reviewing open defects, deciding their
  disposition (confirm, defer, reject), creating remediation tasks, or
  verifying fixes.
stage-classification: temporary
replacement-stage: 1
replacement: "`arci defect` CLI subcommands (create, confirm, defer, reject)"
---

# Triage and fix defects

Walk through open defects and drive them toward resolution.

## Open defects

!`MOD_ID="$1"; jq -s --arg id "$MOD_ID" '
  [.[] | select(."@type" == "Defect" and .module."@id" == $id and (.status == "open" or .status == "confirmed"))] |
  map({id: ."@id", title: .title, severity: .severity, category: .category, status: .status, subject: .subject."@id"})
' .arci/graph.jsonlt 2>/dev/null || echo '[]'`

## Instructions

For each open defect:

1. Read the defect statement and the subject node it references.
2. Determine disposition: confirm (it's a real problem), defer (real but not now, set `deferralTarget`), or reject (not a problem, add `rationale`).
3. For confirmed defects, create a remediation TASK-* node and link it with `generates`.
4. After the fix lands, update the defect: `status` → `"resolved"`, add `resolutionNotes`.
5. After verification of the fix, update to `"verified"`, then `"closed"`.

Prioritize by severity. Critical and major defects block module advancement.
