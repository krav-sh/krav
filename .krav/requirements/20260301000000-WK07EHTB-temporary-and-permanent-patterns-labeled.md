# Temporary and permanent patterns labeled

The project shall mark every Stage 0 artifact (including shell-based skills, CLAUDE.md enforcement rules, and manual graph-editing conventions) as either temporary (with the stage at which its replacement takes effect and the mechanism that replaces it) or permanent (indicating it survives all stages).

## Rationale

Labeling creates useful documentation at the point of use. A contributor reading a skill file immediately sees whether it is scaffolding or load-bearing, and if temporary, what replaces it. Without explicit permanent markers, a missing temporary annotation is ambiguous (it could mean permanent or forgotten). This also produces a natural checklist for stage transition verification.

## Verification criteria

Every shell-based skill in .claude/skills/, every development discipline rule in CLAUDE.md, and every codified graph-editing convention includes an annotation classifying it as temporary (with replacement stage and mechanism) or permanent.
