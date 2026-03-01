# Hook schema

This document defines arci's normalized hook schema for Claude Code integration. The schema maps Claude Code's hook events to canonical event types used in policy evaluation.

For how policies use this schema, see [Policy model](policy-model.md). For how evaluation works, see [Execution model](execution-model.md).

## Design principles

The schema follows three principles. First, use action-oriented event names and canonical tool identifiers that align with Claude Code's conventions. Second, provide silent skip for unavailable context so that policies requiring session scope are silently skipped when session ID is unavailable, rather than failing evaluation. Third, normalize common fields while providing escape hatches so that Claude Code-specific fields remain accessible through the `raw` field for policies that need them.

## Event types

arci defines canonical event types that map to Claude Code's native hook event names.

### Core lifecycle events

The `pre_tool_call` event fires after Claude decides to invoke a tool but before execution begins. Policies can block the operation, auto-approve, mutate the input, or inject warnings. This is the primary enforcement point for security and safety policies. Claude Code event: `PreToolUse`.

The `post_tool_call` event fires immediately after a tool completes successfully. Policies can validate output, inject feedback to the agent, log for auditing, or trigger follow-up actions. This event cannot block since the operation has already completed. Claude Code event: `PostToolUse`.

The `pre_prompt` event fires when a user submits a prompt, before Claude begins processing. Policies can inject additional context, block the prompt entirely, or log for analysis. Claude Code event: `UserPromptSubmit`.

The `post_response` event fires when Claude finishes responding. Policies can prevent stopping and provide instructions to continue, log the completion, or trigger cleanup actions. Claude Code event: `Stop`.

### Session lifecycle events

The `session_start` event fires when a new session begins or an existing session resumes. Policies can inject initial context, load state, set up environment variables, or perform session initialization. The `source` field indicates whether this is a new startup, a resumed session, or a cleared context. Claude Code event: `SessionStart`.

The `session_end` event fires when a session ends. Policies can perform cleanup, log session summaries, or persist state. The `reason` field indicates why the session ended. Claude Code event: `SessionEnd`.

### Other events

The `permission_request` event fires when Claude displays a permission dialog to the user. Policies can auto-approve or auto-deny on behalf of the user. Claude Code event: `PermissionRequest`.

The `notification` event fires when Claude sends a notification. Policies can filter, log, or suppress notifications based on type and content. Claude Code event: `Notification`.

The `pre_compact` event fires before context compaction occurs. Policies can inject content that should survive compaction, adjust behavior based on whether compaction is manual or automatic, or log compaction events. Claude Code event: `PreCompact`.

The `subagent_stop` event fires when a subagent completes its task. It shares the same schema as `post_response` but adds `subagent_id` to identify which subagent finished. Claude Code event: `SubagentStop`.

## Tool names

arci uses Claude Code's native tool names as canonical identifiers.

### Canonical tool names

The `Bash` tool executes shell commands.

The `Write` tool creates or overwrites files.

The `Read` tool reads file contents.

The `Edit` tool modifies existing files with targeted edits. `MultiEdit` performs multiple edits in one call.

The `Glob` tool searches for files by pattern.

The `Grep` tool searches file contents.

The `Task` tool spawns subagents.

The `WebSearch` tool performs web searches.

The `WebFetch` tool fetches web content.

The `mcp:` prefix identifies MCP server tools. The canonical format is `mcp:server:tool`.

```yaml
# Matches shell commands
conditions:
  - expression: 'tool_name == "Bash"'

# Matches all MCP tools
conditions:
  - expression: 'tool_name.startsWith("mcp:")'
```

## Input schema

Policies evaluate against a normalized input context. Common fields are present in all events; event-specific fields are available only for relevant event types.

### Common fields

The `event_type` field contains the canonical event type (e.g., `pre_tool_call`, `session_start`).

The `session_id` field contains Claude Code's session identifier (a UUID that persists across the conversation).

The `cwd` field contains the current working directory as an absolute path.

The `timestamp` field contains the event timestamp.

### Tool event fields

For `pre_tool_call` and `post_tool_call` events:

The `tool_name` field contains the tool name (e.g., `Bash`, `Write`).

The `tool_input` field contains tool parameters as an object. Contents vary by tool. See tool input schemas below.

For `post_tool_call` only:

The `tool_output` field contains the tool's result. For shell commands, this includes `exit_code`, `stdout`, and `stderr`. For file operations, this may include success indicators and content.

### Prompt event fields

For `pre_prompt`:

The `prompt` field contains the user's submitted text.

For `post_response`:

The `response` field contains Claude's response text (where available).

### Session event fields

For `session_start`:

The `source` field indicates the session source: `startup` for new sessions, `resume` for continued sessions, or `clear` for cleared context.

For `session_end`:

The `reason` field indicates why the session ended.

### Subagent event fields

For `subagent_stop`:

The `subagent_id` field contains the identifier for the subagent that completed.

The `response` field contains the subagent's response text (where available).

### Tool input schemas

The `tool_input` object structure varies by tool type.

For `Bash`:

- `command` (string): The shell command to execute
- `timeout` (integer, optional): Timeout in milliseconds
- `cwd` (string, optional): Working directory for command

For `Write`:

- `file_path` (string): Absolute path to the file
- `content` (string): File contents to write

For `Read`:

- `file_path` (string): Absolute path to the file
- `start_line` (integer, optional): Starting line number
- `end_line` (integer, optional): Ending line number

For `Edit`:

- `file_path` (string): Absolute path to the file
- `edits` (array): List of edits, each with `old_string` and `new_string`

For `Glob`:

- `pattern` (string): Glob pattern to match files (e.g., `**/*.py`)
- `path` (string, optional): Base directory to search from; defaults to `cwd`

For `Grep`:

- `pattern` (string): Regular expression pattern to search for
- `path` (string, optional): File or directory to search; defaults to `cwd`
- `include` (string, optional): Glob pattern to filter files (e.g., `*.py`)

For `Task`:

- `prompt` (string): The task description for the subagent
- `subagent_type` (string, optional): Type of subagent to spawn
- `context` (string, optional): Additional context for the subagent

For `WebSearch`:

- `query` (string): The search query
- `num_results` (integer, optional): Maximum number of results

For `WebFetch`:

- `url` (string): The URL to fetch
- `prompt` (string, optional): Instructions for processing the content

For `mcp:` tools:

- `server` (string): MCP server name
- `tool` (string): Tool name within the server
- `arguments` (object): Tool-specific parameters

### Accessing raw Claude Code data

The `raw` field contains the unmodified hook input from Claude Code. This provides access to fields not present in the normalized schema, such as `permission_mode`, `transcript_path`, and `tool_use_id`.

```yaml
# Access Claude Code's permission mode
conditions:
  - expression: 'raw.permission_mode == "plan"'

# Access the tool use ID
variables:
  - name: tool_id
    expression: 'raw.tool_use_id'
```

## Output schema

arci produces a response that is translated to Claude Code's expected output format.

### Response structure

```json
{
  "action": "allow",
  "messages": {
    "user": ["Warning: large file detected"],
    "assistant": ["User's security policy requires confirmation for files over 50KB"]
  },
  "mutations": {
    "tool_input": {
      "command": "rm -i file.txt"
    }
  },
  "continue": true
}
```

### Action values

The `action` field determines the outcome:

- `allow`: Permit the operation to proceed
- `warn`: Permit the operation but surface warnings
- `deny`: Block the operation

### Messages

The `messages` object contains arrays of messages for different audiences:

- `user`: Messages displayed to the human operator
- `assistant`: Messages injected into Claude's context

Messages accumulate from all policies that match and produce output.

### Mutations

The `mutations` object contains field modifications to apply to the hook event. Mutations are merged from all matching policies according to priority and cascade order. The mutated event is what Claude receives if the action is `allow` or `warn`.

### Control fields

The `continue` field (boolean, default true) determines whether Claude should continue processing. Setting `continue: false` halts the assistant loop entirely.

### Output format translation

arci translates the abstract output to Claude Code's expected format:

| Action | Claude Code Behavior |
|--------|---------------------|
| Allow  | Exit 0, JSON response |
| Deny   | Exit 2, stderr message |
| Warn   | Exit 0, JSON with message |

## Edge cases

### Empty or null fields

Optional or unavailable fields should be checked with CEL's `has()` macro or the null-coalescing operator before use.

```yaml
# Check for field existence
conditions:
  - expression: 'has(tool_input.timeout) && tool_input.timeout > 30000'

# Use null-coalescing for defaults
variables:
  - name: timeout
    expression: 'tool_input.timeout ?? 5000'
```

### MCP tool matching

MCP tools use hierarchical names. The canonical format is `mcp:server:tool`. Policies can match at any level:

```yaml
# Match all MCP tools
conditions:
  - expression: 'tool_name.startsWith("mcp:")'

# Match all tools from a specific server
conditions:
  - expression: 'tool_name.startsWith("mcp:github:")'

# Match a specific MCP tool
conditions:
  - expression: 'tool_name == "mcp:github:create_issue"'
```

### Directory reads

Some events may fire `Read` for directory paths when performing recursive reads. Policies should handle this case if they need to distinguish files from directories:

```yaml
conditions:
  - expression: 'tool_name == "Read" && !$is_directory(tool_input.file_path)'
```

### Timeout handling

Hook timeout behavior is configured in Claude Code's settings (default 30 seconds per hook). arci enforces its own timeout on policy evaluation. When evaluation times out, the policy is skipped with fail-open semantics.

### Regex patterns

CEL provides regex support through the `.matches()` method (RE2 syntax). Note that backslashes must be escaped in YAML strings.

```yaml
# Match rm with -rf flags
conditions:
  - expression: 'tool_input.command.matches("rm\\s+-rf")'

# Case-insensitive match
conditions:
  - expression: 'tool_input.path.matches("(?i)\\.py$")'
```

For simple substring checks, prefer `.contains()` which is more readable.
