# Linter output examples

Sample output from each linter to help with categorization.

## rumdl output

```text
docs/CONCEPT.md:15:1: [MD004] Unordered list style: Expected: dash; Actual: asterisk [*]
docs/CONCEPT.md:42:81: [MD009] Trailing whitespace [*]
docs/design/overview.md:1:1: [MD041] First line in a file should be a top-level heading [*]

Issues: Found 3 issues in 2/5 files (12ms)
Run `rumdl fmt` to automatically fix 3 of the 3 issues
```

Each line follows the pattern `file:line:column: [RULE] description [*]`. The `[*]` suffix means the finding supports auto-fix via `rumdl fmt`.

## Vale output

```text
 docs/CONCEPT.md
 12:1   warning  'Sentence-case headings'       bunshi.Headings
                 should use sentence-style
                 capitalization.
 45:15  error    Did you really mean            Vale.Spelling
                 'metareasoning'?
 78:5   warning  'is loaded' may be passive     write-good.Passive
                 voice. Use active voice if
                 you can.

 docs/design/overview.md
 3:20   error    AI hedge: 'it is important     ai-tells.HedgingPhrases
                 to note that'. Delete this
                 throat-clearing and state
                 your point directly.
```

Each block starts with a filename. Each finding shows `line:column`, severity (`warning` or `error`), the message, and the rule name. Use the rule name to categorize:

- `Vale.Spelling` or `Google.Spelling` = vocabulary finding
- `bunshi.Headings`, `write-good.*`, `Google.*`, `proselint.*`, `ai-tells.*` = prose style finding
