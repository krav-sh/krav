# Design docs refactoring: entity type naming and semantics

## Context

We've redesigned several entity types in arci's knowledge graph based on deeper alignment with traditional RE tools and standards. This task refactors all design docs under `docs/design/` to reflect these changes.

## Prefix and naming changes

The fixed 3-character prefix convention is replaced with variable-length natural abbreviations. The nanoid portion stays 8-character Crockford Base32.

Old → New prefix mappings:

| Old prefix | Old name      | New prefix | New name             | Notes                                      |
|------------|---------------|------------|----------------------|--------------------------------------------|
| CON-*      | Concept       | CON-*      | Concept              | No change                                  |
| ENT-*      | Entity        | MOD-*      | Module               | Already renamed in some docs               |
| NED-*      | Need          | NEED-*     | Need                 | Prefix change only                         |
| REQ-*      | Requirement   | REQ-*      | Requirement          | No change                                  |
| VRF-*      | Verification  | TC-*       | Test case            | Semantic change: specification only, see below |
| TSK-*      | Task          | TASK-*     | Task                 | Prefix change only                         |
| FND-*      | Finding       | DEF-*      | Defect               | Full replacement, see below                |
| (new)      |               | BSL-*      | Baseline             | New type, spec exists at docs/design/requirements/baselines.md |
| (new)      |               | TCAM-*     | Test campaign        | New type, not yet spec'd                   |

## What to change across all docs

### 1. Prefix replacements (mechanical)

In every file under `docs/design/`:

- `ENT-` → `MOD-` (if not already done)
- `NED-` → `NEED-`
- `VRF-` → `TC-`
- `TSK-` → `TASK-`
- `FND-` → `DEF-`

This applies to:
- Inline identifier examples (e.g., `NED-B7G3M9K2` → `NEED-B7G3M9K2`)
- JSON-LD examples (e.g., `"@id": "VRF-D9J5Q1R3"` → `"@id": "TC-D9J5Q1R3"`)
- Prose references (e.g., "VRF-* nodes" → "TC-* nodes")
- Tables listing entity types
- Code examples (Python, YAML, bash)

### 2. Type name replacements (mechanical)

- "Entity" / "entity" → "Module" / "module" when referring to the MOD-* type (be careful not to replace the English word "entity" when used generically)
- "Verification" / "verification" → "Test case" / "test case" when referring to the TC-* type (be careful: "verification" as a process/activity should stay as "verification" — only the entity type name changes)
- "Finding" / "finding" → "Defect" / "defect" when referring to the DEF-* type
- "Task" stays "Task" (only prefix changes)
- "Need" stays "Need" (only prefix changes)

### 3. JSON-LD @type values

- `"@type": "Entity"` → `"@type": "Module"`
- `"@type": "Verification"` → `"@type": "TestCase"`
- `"@type": "Finding"` → `"@type": "Defect"`
- `"@type": "Task"` stays (but `@id` prefixes change to `TASK-`)

### 4. Predicate renames

- `"entity":` → `"module":` (the ownership predicate on needs, requirements, tasks, etc.)
- `"regarding":` → `"subject":` (on defects, per new defect spec)
- `"addressedBy":` → `"resolvedBy":` (on defects, per new defect spec)

### 5. Python class/type renames

- `EntityNode` → `ModuleNode`
- `EntityPhase` → `ModulePhase`
- `EntityStatus` → `ModuleStatus`
- `VerificationNode` → `TestCaseNode`
- `VerificationStatus` → `TestCaseStatus`
- `VerificationLevel` → `TestCaseLevel`
- `VerificationMethod` → `TestCaseMethod` (or `TestMethod`)
- `FindingNode` → `DefectNode`
- `FindingStatus` → `DefectStatus`
- `FindingType` → remove (defects don't have finding_type)
- `TaskNode` stays but references to `TSK-` become `TASK-`
- `NeedNode` stays but references to `NED-` become `NEED-`

### 6. Module path references in Python

- `arci.core.entity` → `arci.core.module`
- `arci.core.verification` → `arci.core.testcase`
- `arci.core.finding` → `arci.core.defect`
- `arci.core.task` stays (but ID prefix references change)
- `arci.core.need` stays (but ID prefix references change)
- `arci.service.entity` → `arci.service.module`
- `arci.service.verification` → `arci.service.testcase`
- `arci.service.finding` → `arci.service.defect`
- `arci.cli._commands._entity` → `arci.cli._commands._module`
- Similar for IO layer references

### 7. CLI command name changes

- `arci entity *` → `arci module *`
- `arci verification *` → `arci tc *` (or `arci testcase *` — use `arci tc` for brevity)
- `arci finding *` → `arci defect *`
- `arci task *` stays but examples using `TSK-` become `TASK-`
- `arci need *` stays but examples using `NED-` become `NEED-`

### 8. ID format regex/validation references

The old pattern `<TYPE>-<NANOID>` where TYPE is always 3 characters changes to variable-length prefixes. Update any references to the ID format:
- Old: `<TYPE>-<NANOID>` where TYPE is a 3-character prefix
- New: `<PREFIX>-<NANOID>` where PREFIX is a variable-length uppercase alphabetic string (2-4 characters)
- Valid prefixes: CON, MOD, NEED, REQ, TC, TASK, DEF, BSL, TCAM

### 9. Entity type count

The original design references "seven entity types" in several places. The new count is nine:
- CON, MOD, NEED, REQ, TC, TASK, DEF, BSL, TCAM
- Update all references to "seven entity types" or "7 entity types"

### 10. Verification (TC-*) semantic changes

Beyond the rename, the TC-* type has semantic changes. In files that describe VRF-*/Verification:

- TC-* is now a **specification**, not a combined specification+execution entity
- Remove execution state from the TC-* node definition:
  - `status` lifecycle changes from `draft → ready → passing → failing → skipped → obsolete` to `draft → specified → implemented → executable → obsolete`
  - Add `currentResult` field (separate from `status`): `pass | fail | skip | unknown`
  - Add `lastRunAt` timestamp field
  - Remove `lastResult` object (replaced by `currentResult` + `lastRunAt`)
- Add `acceptanceCriteria` field for explicit pass/fail criteria
- The `implementation` field stays but now represents a deliverable produced by a verification-implementation task
- Add notes about specification-implementation decoupling: for complex test types (benchmarks, e2e, analyses), the implementation is a task deliverable that can itself be reviewed and have defects filed against it

### 11. Finding (FND-*) → Defect (DEF-*) semantic changes

The new defect spec already exists at `docs/design/verification/defects.md`. The old findings spec at `docs/design/verification/findings.md` should be replaced.

Key semantic changes to propagate to other docs that reference findings:
- `findingType` field is removed entirely (defects are always problems)
- `regarding` predicate → `subject` predicate
- `addressedBy` predicate → `resolvedBy` predicate  
- `generates` predicate stays
- Severity is now required (was optional on findings)
- New fields: `category`, `detectedBy`, `detectedInPhase`, `deferralTarget`, `rationale`
- Lifecycle changes from `open → acknowledged → addressed → closed | wont_fix` to `open → confirmed → resolved → verified → closed` with `rejected` and `deferred` branches
- Decisions are no longer defects — they're concepts (CON-*). Remove any references to "decision findings" and note that decisions are modeled as concepts
- Questions are no longer defects — they're task-level context. Remove references to "question findings"
- Observations are no longer defects — they're review deliverable prose. Remove references to "observation findings"
- "Blocking findings" → "blocking defects"
- Auto-generated findings from suspect links → suspect links surface in views; reviewers create defects manually when problems are real

### 12. context.jsonld updates

If any doc shows the JSON-LD context, update it:
- `"Entity": "arci:Entity"` → `"Module": "arci:Module"`
- `"Verification": "arci:Verification"` → `"TestCase": "arci:TestCase"`
- `"Finding": "arci:Finding"` → `"Defect": "arci:Defect"`
- `"entity":` predicate → `"module":` predicate
- `"regarding":` predicate → `"subject":` predicate
- `"addressedBy":` predicate → `"resolvedBy":` predicate
- Add `"Baseline": "arci:Baseline"` and `"TestCampaign": "arci:TestCampaign"`

### 13. Relationship type table updates

Several docs have a table listing all relationship types. Update:
- `entity` predicate → `module` predicate (ownership)
- `regarding` predicate → `subject` predicate (defect-to-target)
- `addressedBy` predicate → `resolvedBy` predicate (task resolves defect)
- `verifiedBy` / `verifies` predicates → `testedBy` / `tests` (connecting REQ ↔ TC) — actually keep `verifiedBy`/`verifies` since "verifies" is the correct relationship semantic even though the entity type is now "test case"

On reflection, keep the `verifiedBy`/`verifies` predicate names. "Test case X verifies requirement Y" reads correctly. The predicate describes the relationship semantic, not the entity type.

## Files to modify

Every `.md` file under `docs/design/` should be checked. Key files that will need heavy edits:

- `docs/design/data-model.md` — central reference, touches everything
- `docs/design/knowledge-graph-architecture.md` — typed nodes, serialization, all examples
- `docs/design/requirements/entities.md` — rename to `modules.md`, full rewrite of type name
- `docs/design/requirements/requirements.md` — references to entity/verification/finding
- `docs/design/verification/verifications.md` — heavy semantic changes (TC-* specification model)
- `docs/design/verification/findings.md` — replace with reference to defects.md, or delete
- `docs/design/verification/defects.md` — already written with new model, but verify consistency with other changes
- `docs/design/requirements/baselines.md` — already written, but may reference old prefixes
- `docs/design/execution/tasks.md` — TSK-* → TASK-* prefix changes
- `docs/design/intent/concepts.md` — may reference old prefixes
- `docs/design/intent/needs.md` — NED-* → NEED-* prefix changes
- `docs/design/hooks/` — examples reference old prefixes
- `docs/design/cli/` — command references
- `docs/design/glossary.md` — term definitions
- `docs/design/diagrams/*.puml` — PlantUML sources

## Files to rename

- `docs/design/requirements/entities.md` → `docs/design/requirements/modules.md`

## Files to delete

- `docs/design/verification/findings.md` (replaced by `docs/design/verification/defects.md`)

## Approach

1. Start by reading all files to understand current state
2. Do the mechanical prefix/name replacements first across all files
3. Then handle the semantic changes (TC-* specification model, DEF-* defect model)
4. Rename `entities.md` → `modules.md`
5. Delete `findings.md`
6. Update the glossary
7. Verify consistency: grep for any remaining old prefixes (NED-, VRF-, TSK-, FND-, ENT-) and old type names

## Caution

- Don't replace the English word "entity" when used generically (e.g., "entity types" as a general concept). Only replace when it refers to the specific MOD-* type.
- Don't replace "verification" when used as a process/activity name. Only replace when it refers to the TC-* entity type.
- Don't replace "finding" when used as a general English word. Only replace when it refers to the DEF-* entity type.
- Don't replace "task" when used as a general English word. Only replace when it refers to the TASK-* entity type.
- Preserve the voice and style of each document. This is a refactoring, not a rewrite.
- The implementation status tables in each doc should stay accurate — don't claim things are implemented that aren't.
