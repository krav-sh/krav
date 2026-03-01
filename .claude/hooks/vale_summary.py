#!/usr/bin/env python3
"""Format vale JSON output into a structured summary for Claude Code.

Usage:
    uvx vale --output=JSON [files...] | python3 tools/vale_summary.py [--file FILE] [--rule RULE]

Reads vale JSON from stdin and prints a structured summary with sections:
  - Total counts (by severity)
  - Findings by file (count per file)
  - Findings by rule (count per rule)
  - Spelling findings (words to add to vocabulary)
  - All findings (grouped by file, one line each)

Filters:
  --file FILE   Show only findings for FILE (substring match)
  --rule RULE   Show only findings matching RULE (substring match)
"""

import json
import sys
from collections import Counter


def main() -> None:
    file_filter = None
    rule_filter = None

    args = sys.argv[1:]
    i = 0
    while i < len(args):
        if args[i] == "--file" and i + 1 < len(args):
            file_filter = args[i + 1]
            i += 2
        elif args[i] == "--rule" and i + 1 < len(args):
            rule_filter = args[i + 1]
            i += 2
        else:
            print(f"Unknown argument: {args[i]}", file=sys.stderr)
            sys.exit(1)

    data = json.load(sys.stdin)

    # Collect all findings, applying filters
    findings: list[tuple[str, dict]] = []
    for filepath, alerts in sorted(data.items()):
        for alert in alerts:
            if file_filter and file_filter not in filepath:
                continue
            if rule_filter and rule_filter not in alert["Check"]:
                continue
            findings.append((filepath, alert))

    if not findings:
        print("No findings.")
        return

    # Severity counts
    severity_counts: Counter[str] = Counter()
    for _, alert in findings:
        severity_counts[alert["Severity"]] += 1

    # By file
    file_counts: Counter[str] = Counter()
    for filepath, _ in findings:
        file_counts[filepath] += 1

    # By rule
    rule_counts: Counter[str] = Counter()
    for _, alert in findings:
        rule_counts[alert["Check"]] += 1

    # Spelling findings (words to add to vocabulary)
    spelling_words: list[tuple[str, str]] = []
    for filepath, alert in findings:
        if alert["Check"] in ("Vale.Spelling", "Google.Spelling"):
            spelling_words.append((alert["Match"], filepath))

    # Print summary
    total = len(findings)
    severity_parts = [f"{count} {sev}" for sev, count in severity_counts.most_common()]
    print(f"## {total} findings ({', '.join(severity_parts)})")
    print()

    # By file
    print("### By file")
    for filepath, count in file_counts.most_common():
        print(f"  {count:3d}  {filepath}")
    print()

    # By rule
    print("### By rule")
    for rule, count in rule_counts.most_common():
        print(f"  {count:3d}  {rule}")
    print()

    # Spelling words
    if spelling_words:
        unique_words = sorted({w for w, _ in spelling_words}, key=str.lower)
        print(f"### Spelling ({len(unique_words)} words to add to vocabulary)")
        for word in unique_words:
            print(f"  {word}")
        print()

    # All findings, grouped by file
    print("### Findings")
    current_file = None
    for filepath, alert in findings:
        if filepath != current_file:
            current_file = filepath
            print(f"\n  {filepath}")
        line = alert["Line"]
        col_start = alert["Span"][0]
        severity = alert["Severity"][0]  # e/w/s
        check = alert["Check"]
        message = alert["Message"]
        print(f"    {line}:{col_start} {severity} {check}: {message}")


if __name__ == "__main__":
    main()
