---
name: fix-docs
description: Fix documentation linting issues by running markdown (rumdl) and prose (vale) linters in a loop, fixing findings until all checks pass. Use when docs fail linting, after editing Markdown files, or when asked to fix docs, lint docs, fix prose, fix markdown, or clean up documentation.
---

# Fix documentation linting issues

Run markdown and prose linters, fix all findings, and repeat until every check passes with zero warnings and errors.

## Instructions

### Step 1: Determine target files

Parse `$ARGUMENTS` for file paths. If the user provided no arguments, target all Markdown files in the repository.

Store the file list for later steps. If the user provided specific files, verify each file exists. If a file does not exist, report the missing file and stop.

### Step 2: Run linters and collect findings

Run both linters against the target files. Capture the full output from each command.

**Markdown linting (rumdl)**:

- With files: `just lint-markdown <file1> <file2> ...`
- Without files: `just lint-markdown`

**Prose linting (vale)**:

- With files: `just lint-prose <file1> <file2> ...`
- Without files: `just lint-prose`

If both linters return zero findings, report success and stop.

See [examples/linter-output.md](examples/linter-output.md) for sample output from each linter.

### Step 3: Categorize each finding

Read the linter output and categorize every finding into one of these types:

1. **Markdown structure** (rumdl): heading style, list markers, code fences, blank lines, trailing whitespace, and other Markdown formatting rules. See [references/fix-patterns.md](references/fix-patterns.md) for common patterns.
2. **Vocabulary** (vale `Vale.Spelling` or `Google.Spelling`): a word flagged as misspelled that is a legitimate project term, proper noun, or technical term.
3. **Prose style** (vale, all other rules): wording, grammar, punctuation, hedging, passive voice, and other prose quality findings.

### Step 4: Fix findings

Work through the categorized findings and apply fixes:

**Markdown structure fixes** - run `uvx rumdl fmt <files>` first to auto-fix trivial issues (marked with `[*]` in the output). Then manually edit any remaining findings to match the project conventions:

- ATX-style headings (`#` prefix, not underline style)
- Dash list markers (`-`, not `*` or `+`)
- Backtick code fences (not tildes)

**Vocabulary fixes** - add the flagged word to `.vale/config/vocabularies/bunshi/accept.txt` in alphabetical order. Use regular expression character classes for case variation (`[Ff]oo` accepts both "Foo" and "foo"). Do not rewrite prose to avoid a legitimate technical term.

**Prose style fixes** - rewrite the sentence or phrase to address the rule while preserving the original meaning. Common rewrites:

- Remove hedging phrases and state the point directly
- Convert passive voice to active voice
- Remove weasel words and filler phrases
- Fix capitalization in headings (sentence case required)

If a finding looks like a false positive that you cannot fix without degrading the prose, skip it and report it to the user at the end.

### Step 5: Re-run linters

After applying all fixes, re-run both linters using the same commands from Step 2.

- If zero findings remain, report success and stop.
- If findings remain, return to Step 3 and fix the new findings.
- If the same finding persists after two consecutive fix attempts, skip it and report it to the user as unresolvable.

Limit the loop to 5 iterations. If findings remain after 5 iterations, report the remaining findings to the user and stop.

### Step 6: Report results

Summarize the changes:

- Number of files checked
- Total findings fixed
- Vocabulary terms added (list each term)
- Skipped findings (list each with the reason)
- Final linter status (pass or remaining finding count)

## Error handling

- If `just lint-markdown` or `just lint-prose` fails with a non-linting error (tool not installed, configuration error, or similar), report the error to the user and stop. Do not attempt to fix infrastructure issues.
- If you cannot read or write a file, report the error and continue with remaining files.
- If vale styles have not synced (missing style packages), tell the user to run `just vale-sync` and stop.
