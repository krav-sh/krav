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
2. If the defect lacks a prose file at `.arci/defects/{timestamp}-{NANOID}-{slug}.md`, create one during triage. Include the full problem description, evidence, reproduction context, and the reasoning behind the disposition decision.
3. Determine disposition: confirm (it's a real problem), defer (real but not now, set `deferralTarget`), or reject (not a problem, add `rationale`).
4. For confirmed defects, create a remediation TASK-* node and link it with `generates`. Create a prose file for the task at `.arci/tasks/{timestamp}-{NANOID}-{slug}.md` with the remediation approach and constraints.
5. After the fix lands, update the defect: `status` → `"resolved"`, add `resolutionNotes`.
6. After verification of the fix, update to `"verified"`, then `"closed"`.

Prioritize by severity. Critical and major defects block module advancement.

## Graph-editing conventions

| Pattern | Classification | Stage | Replacement |
|---------|---------------|-------|-------------|
| Open defects by module context query | Temporary | 1 | `arci defect list` CLI command |
