# Dashboard Features

The dashboard provides five views for monitoring and debugging hook activity. Each view is accessible from the main navigation and serves a distinct diagnostic purpose.

## Live event stream

The primary dashboard view shows a live stream of hook events. As hooks fire, events appear in real-time via WebSocket. Each event shows:

- Timestamp
- Event type (pre_tool_use, session_start, etc.)
- Project and session
- Which rules matched
- What actions executed
- Outcome (allowed, denied, error)

Events can be filtered by event type, project, session, or rule ID. Filtering happens client-side for immediate feedback.

Clicking an event expands it to show full details including the complete hook input, evaluation context, each rule that was considered and whether it matched, and action outputs.

This view is essential for understanding what arci is doing and debugging unexpected behavior. See [real-time.md](real-time.md) for WebSocket and htmx integration details.

## Rule match statistics

The statistics view shows aggregate information about rule matching. For each rule, it displays:

- Total number of times the rule was evaluated
- How many times the condition matched
- How many times each action type executed
- Timing percentiles for evaluation

Rules can be sorted by match count, name, or priority. Filtering narrows the view to specific rules.

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

The tester shows which rules would match, what actions would execute, and what the output would be — all without affecting the actual hook system. This is a dry-run: no state is modified and no actions are executed.

This is invaluable for developing and testing rules before deploying them.
