# Dashboard features

The dashboard provides five views for monitoring and debugging hook activity. Each view is accessible from the main navigation and serves a distinct diagnostic purpose.

## Live event stream

The primary dashboard view shows a live stream of hook events. As hooks fire, events appear in real-time via WebSocket. Each event shows:

- Timestamp
- Event type (pre_tool_use, session_start, etc.)
- Project and session
- Which rules matched
- What actions executed
- Outcome (allowed, denied, error)

Users can filter events by event type, project, session, or rule ID. Filtering happens client-side for immediate feedback.

Clicking an event expands it to show full details including the complete hook input, evaluation context, each rule the evaluator considered and whether it matched, and action outputs.

This view is essential for understanding what ARCI is doing and debugging unexpected behavior. See [real-time.md](real-time.md) for WebSocket and htmx integration details.

## Rule match statistics

The statistics view shows aggregate information about rule matching. For each rule, it displays:

- Total number of times the evaluator ran the rule
- How many times the condition matched
- How many times each action type executed
- Timing percentiles for evaluation

Users can sort rules by match count, name, or priority. Filtering narrows the view to specific rules.

This view helps identify which rules are active and which might never be triggering.

## Configuration status

The configuration view shows the current configuration state:

- All discovered configuration sources with their paths, whether the file exists, and when it was last modified
- Validation status, highlighting any errors in configuration files
- The merged rule list with each rule's source file

If there are configuration errors, they're prominently displayed with details about the file, line number, and error message.

This view helps diagnose configuration problems and understand where rules are coming from.

## State store browser

The state view provides access to the state store. It shows all entries for the selected project and session. Each entry displays:

- Key
- Current value
- Creation timestamp and author
- Last update timestamp and author

Users can filter entries by key prefix. For debugging, users can also set or delete entries directly from the dashboard.

This view helps understand what state rules have accumulated and debug state-dependent rules.

## Rule tester

The tester view allows dry-running rules against sample input. The workflow:

1. Select a hook event type
2. Provide JSON input (or use a template)
3. Run the evaluation

The tester shows which rules would match, what actions would execute, and what the output would be, all without affecting the actual hook system. This is a dry-run: the system modifies no state and executes no actions.

This is invaluable for developing and testing rules before deploying them.

## Diagnostics

The configuration panel shows each configuration source with its load status (loaded, error, not found). Errors display inline with the source path and expand to show full error details.

Rule validation status appears as badges: green checkmark for valid rules, yellow warning for rules with non-fatal issues, red X for rules that failed to compile. Hovering shows the specific issue.

The rule tester provides step-by-step evaluation traces, showing each rule's condition, whether it matched, and why. This is invaluable for debugging "why didn't the rule fire?" questions.

## Log integration

The dashboard reads hook event logs (JSONL files) for display and filtering. The Hive-partitioned file structure supports efficient tailing for live updates and scoping queries by project without scanning irrelevant files.

For complex queries, the dashboard can use DuckDB's WASM build for in-browser analytics, or query via the daemon's API when advanced aggregation requires server-side processing.
