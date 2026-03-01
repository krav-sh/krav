# Fix patterns reference

Common linting findings and their fixes for this project.

## Markdown structure (rumdl)

The `.rumdl.toml` file configures these rules. MD013 (line length) and MD033 (inline HTML) are off.

| Rule | Finding | Fix |
| --- | --- | --- |
| MD003 | Wrong heading style | Use ATX-style headings (`# Heading`, not underline) |
| MD004 | Wrong list marker | Use dash markers (`-`, not `*` or `+`) |
| MD005 | Inconsistent list indentation | Align nested list items consistently |
| MD009 | Trailing whitespace | Remove trailing spaces from the line |
| MD010 | Hard tabs | Replace tabs with spaces |
| MD012 | Extra consecutive blank lines | Reduce to a single blank line |
| MD018 | No space after `#` in heading | Add a space: `# Heading` not `#Heading` |
| MD019 | Extra spaces after `#` | Use exactly one space after `#` |
| MD022 | Heading not surrounded by blank lines | Add blank lines before and after headings |
| MD023 | Heading with leading spaces | Remove leading spaces before `#` |
| MD025 | Extra top-level headings | Use only one `# H1` per file |
| MD026 | Trailing punctuation in heading | Remove the trailing period, colon, or other punctuation |
| MD027 | Extra spaces after block quote marker | Use one space after `>` |
| MD028 | Blank line inside block quote | Remove the blank line or use `>` on the blank line |
| MD030 | Spaces after list markers | Use one space after `-` |
| MD031 | Fenced code block not surrounded by blank lines | Add blank lines before and after code fences |
| MD032 | Lists not surrounded by blank lines | Add blank lines before and after lists |
| MD034 | Bare URL | Wrap in angle brackets or use a markdown link |
| MD036 | Emphasis used instead of heading | Convert bold lines to proper headings |
| MD037 | Spaces inside emphasis markers | Remove inner spaces from emphasis markers |
| MD038 | Spaces inside code spans | Remove inner spaces from code span delimiters |
| MD039 | Spaces inside link text | Remove spaces from inside link text brackets |
| MD040 | Fenced code block without language | Add a language identifier after opening fence |
| MD041 | First line not a top-level heading | Start the file with `# Heading` |
| MD046 | Inconsistent code block style | Use backtick fences, not indented code blocks |
| MD047 | File not ending with newline | Add a newline at the end of the file |
| MD048 | Inconsistent code fence style | Use backtick fences (not tildes) |
| MD049 | Inconsistent emphasis style | Use consistent asterisks for emphasis |
| MD050 | Inconsistent bold style | Use consistent double asterisks |

## Prose style (vale)

### Sentence-case headings

Rule: `bunshi/Headings`

Headings must use sentence-style capitalization. Capitalize only the first word and proper nouns.

- Wrong: `## Key Design Decisions`
- Right: `## Key design decisions`

The `.vale/bunshi/Headings.yml` file lists exceptions for proper nouns and acronyms.

### Hedging and weasel words

Rules: `write-good/Weasel`, `proselint/Hedging`

Remove hedging phrases and weasel words. State facts directly.

- Wrong: "This is arguably the best approach"
- Right: "This approach reduces latency"

### Passive voice

Rules: `Google/Passive`, `write-good/Passive`

Convert passive constructions to active voice.

- Wrong: `The configuration is processed by the runtime`
- Right: `The runtime processes the configuration`

### Wordy phrases

Rule: `write-good/TooWordy`

Replace wordy phrases with concise alternatives.

- `in order to` becomes `to`
- `due to the fact that` becomes `because`
- `at this point in time` becomes `now`

### AI tell phrases

Rule: `ai-tells`

Remove formulaic phrases common in AI-generated text.

- Delete `it is important to note that` and state the point directly
- Delete `it is worth mentioning that` and state the point directly
- Replace `delve into` with `examine` or a more specific verb
- Delete `at its core` and similar filler

## Vocabulary additions

When adding words to `.vale/config/vocabularies/bunshi/accept.txt`:

- Maintain alphabetical order (case-insensitive sort)
- Use regular expression character classes for case: `[Ff]oo` accepts both "Foo" and "foo"
- Add possessive forms if needed: `Foo('s)?`
- Add plural forms if needed: `[Ff]oos?`
- Check existing entries before adding duplicates
