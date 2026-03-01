# Claude Code integration

Claude Code is Anthropic's AI coding assistant that runs in the terminal. Its hooks system provides the foundation for ARCI's policy evaluation engine.

## Hooks overview

Claude Code provides a full lifecycle hooks system that allows intercepting and controlling tool execution, permission requests, session events, and more. Settings files in JSON configure hooks as shell commands that receive JSON input via stdin and communicate decisions via exit codes and stdout.

## Hook events

Claude Code supports the following hook events:

PreToolUse fires after Claude creates tool parameters but before the tool executes. Common matchers include Bash for shell commands, Read/Write/Edit for file operations, and Task for subagent invocations. This hook can block tool execution, change tool input, or auto-approve operations.

PostToolUse fires immediately after a tool completes successfully. It receives the tool response along with the input, allowing result checking and injection of feedback to Claude.

PostToolUseFailure fires when a tool execution fails. It includes the error message and an `is_interrupt` flag indicating whether user interruption caused the failure. Hooks can use this event for error monitoring and recovery logic.

PermissionRequest fires when Claude Code shows the user a permission dialog. Hooks can automatically approve or deny permissions on behalf of the user.

UserPromptSubmit fires when the user submits a prompt before Claude processes it. Hooks can check prompts, block certain requests, or inject extra context.

Notification fires when Claude Code sends notifications, supporting matchers to filter by notification type such as permission_prompt or idle_prompt.

SessionStart fires when a new session begins or an existing session resumes. Useful for loading development context, installing dependencies, or setting environment variables. Supports persisting environment variables via CLAUDE_ENV_FILE. The input includes the `model` field indicating which Claude model the session uses.

SessionEnd fires when a session ends, enabling cleanup tasks and logging. The `reason` field indicates why the session ended, such as `prompt_input_exit`.

Stop fires when Claude finishes responding. Hooks can prevent Claude from stopping and instruct it to continue working. Includes a `stop_hook_active` flag.

SubagentStart fires when Claude Code launches a subagent. It includes `agent_id` (a short hex identifier) and `agent_type` (such as `arci:code-explorer` or `Explore`). Hooks can track and audit subagent activity through this event.

SubagentStop fires when a subagent completes its task. It works like Stop but scoped to subagents. Includes `agent_id`, `agent_transcript_path` for the subagent's transcript, and `stop_hook_active` flag.

PreCompact fires before context compaction, with matchers for manual versus auto triggers.

## Configuration hierarchy

Claude Code uses a three-level configuration model with project-level settings taking precedence:

User-level configuration at `~/.claude/settings.json` provides personal defaults that apply across all projects. This is where global safety rules or personal preferences belong.

Project-level configuration at `.claude/settings.json` contains team-shared settings that should be version controlled. Project rules override user rules.

Local project configuration at `.claude/settings.local.json` provides personal settings that developers should not commit to version control.

Claude Code does not support system-level configuration for enterprise deployments.

For ARCI integration, configuration lives at `~/.claude/arci/config.yaml` for user-level rules, `.claude/arci/config.yaml` for project rules, and `.claude/arci/config.local.yaml` for personal project settings.

## Configuration

Settings.json files at user level (~/.claude/settings.json), project level (.claude/settings.json), or local project level (.claude/settings.local.json) define hooks.

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "/path/to/validator.sh",
            "timeout": 30
          }
        ]
      }
    ]
  }
}
```

Matchers are case-sensitive strings or regular expression patterns that filter which tools trigger the hook. Use * or empty string to match all tools.

## Input schema

All hooks receive JSON on stdin. The schema includes common fields present in all events plus event-specific fields.

### Common input fields

Every hook event receives these base fields:

| Field             | Type   | Description                                                  |
|-------------------|--------|--------------------------------------------------------------|
| `session_id`      | string | UUID identifying the conversation session                    |
| `transcript_path` | string | Path to the conversation JSONL transcript file               |
| `cwd`             | string | Current working directory                                    |
| `permission_mode` | string | One of `default`, `plan`, `acceptEdits`, `bypassPermissions` |
| `hook_event_name` | string | The event type (such as `PreToolUse`, `PostToolUse`)         |

The `permission_mode` field is particularly valuable for ARCI rules. It indicates the current permission context and lets rules behave differently in plan mode (where Claude is just proposing actions) versus normal execution mode.

### `PreToolUse`

PreToolUse fires before tool execution and includes:

| Field         | Type   | Description                                                     |
|---------------|--------|-----------------------------------------------------------------|
| `tool_name`   | string | Tool identifier (such as `Bash`, `Write`, `Edit`, `Read`, `Task`) |
| `tool_input`  | object | Tool-specific parameters (schema varies by tool)                |
| `tool_use_id` | string | Unique identifier for this tool invocation                      |

Example PreToolUse input for a Bash command:

```json
{
  "session_id": "eb5b0174-0555-4601-804e-672d68069c89",
  "transcript_path": "/Users/.../.claude/projects/.../eb5b0174-0555-4601-804e-672d68069c89.jsonl",
  "cwd": "/Users/user/project",
  "permission_mode": "default",
  "hook_event_name": "PreToolUse",
  "tool_name": "Bash",
  "tool_input": {
    "command": "npm test",
    "timeout": 300000
  },
  "tool_use_id": "toolu_01ABC123..."
}
```

Example PreToolUse input for a Write operation:

```json
{
  "session_id": "eb5b0174-0555-4601-804e-672d68069c89",
  "transcript_path": "/Users/.../.claude/projects/.../eb5b0174-0555-4601-804e-672d68069c89.jsonl",
  "cwd": "/Users/user/project",
  "permission_mode": "default",
  "hook_event_name": "PreToolUse",
  "tool_name": "Write",
  "tool_input": {
    "file_path": "/path/to/file.txt",
    "content": "file content"
  },
  "tool_use_id": "toolu_01ABC123..."
}
```

### `PostToolUse`

PostToolUse fires after successful tool execution and adds the tool response:

| Field           | Type   | Description                                   |
|-----------------|--------|-----------------------------------------------|
| `tool_name`     | string | Tool identifier                               |
| `tool_input`    | object | The original tool parameters                  |
| `tool_response` | object | Tool execution result (schema varies by tool) |
| `tool_use_id`   | string | Unique identifier for this tool invocation    |

Example PostToolUse input:

```json
{
  "session_id": "eb5b0174-0555-4601-804e-672d68069c89",
  "transcript_path": "/Users/.../.claude/projects/.../eb5b0174-0555-4601-804e-672d68069c89.jsonl",
  "cwd": "/Users/user/project",
  "permission_mode": "default",
  "hook_event_name": "PostToolUse",
  "tool_name": "Write",
  "tool_input": {
    "file_path": "/path/to/file.txt",
    "content": "file content"
  },
  "tool_response": {
    "filePath": "/path/to/file.txt",
    "success": true
  },
  "tool_use_id": "toolu_01ABC123..."
}
```

### Tool input schemas

The `tool_input` structure varies by tool. Below are the common tools; Claude Code supports many more tools with varying schemas.

**Bash tool:**

```json
{
  "command": "string - the shell command to execute",
  "description": "string - optional description of what the command does",
  "timeout": "number - optional timeout in milliseconds",
  "run_in_background": "boolean - optional, run async"
}
```

**Write tool:**

```json
{
  "file_path": "string - absolute path to file",
  "content": "string - complete file content"
}
```

**Edit tool:**

```json
{
  "file_path": "string - absolute path to file",
  "old_string": "string - text to find",
  "new_string": "string - replacement text",
  "replace_all": "boolean - optional, replace all occurrences"
}
```

**MultiEdit tool:**

```json
{
  "file_path": "string - absolute path to file",
  "edits": [
    {
      "old_string": "string - text to find",
      "new_string": "string - replacement text"
    }
  ]
}
```

**Read tool:**

```json
{
  "file_path": "string - absolute path to file",
  "offset": "number - optional line offset to start reading from",
  "limit": "number - optional number of lines to read"
}
```

**Glob tool:**

```json
{
  "pattern": "string - glob pattern to match files",
  "path": "string - optional directory to search in"
}
```

**Grep tool:**

```json
{
  "pattern": "string - regex pattern to search for",
  "path": "string - optional file or directory to search in",
  "output_mode": "string - 'content', 'files_with_matches', or 'count'",
  "-C": "number - optional context lines around matches",
  "-i": "boolean - optional case insensitive search"
}
```

**Task tool (subagent):**

```json
{
  "prompt": "string - the task for the agent to perform",
  "description": "string - short description of the task",
  "subagent_type": "string - the type of specialized agent to use"
}
```

**TaskCreate tool:**

```json
{
  "subject": "string - brief title for the task",
  "description": "string - detailed description of what needs to be done",
  "activeForm": "string - present continuous form shown in spinner"
}
```

**TaskUpdate tool:**

```json
{
  "taskId": "string - the ID of the task to update",
  "status": "string - 'pending', 'in_progress', or 'completed'",
  "subject": "string - optional new subject",
  "description": "string - optional new description"
}
```

**WebFetch tool:**

```json
{
  "url": "string - the URL to fetch content from",
  "prompt": "string - what information to extract from the page"
}
```

**WebSearch tool:**

```json
{
  "query": "string - the search query"
}
```

**AskUserQuestion tool:**

```json
{
  "questions": [
    {
      "question": "string - the question to ask",
      "header": "string - short label for the question",
      "multiSelect": "boolean - allow multiple selections",
      "options": [
        {
          "label": "string - display text for the option",
          "description": "string - explanation of the option"
        }
      ]
    }
  ]
}
```

### `UserPromptSubmit`

| Field    | Type   | Description                      |
|----------|--------|----------------------------------|
| `prompt` | string | The user's submitted prompt text |

### `Notification`

| Field               | Type   | Description                                  |
|---------------------|--------|----------------------------------------------|
| `message`           | string | The notification message                     |
| `notification_type` | string | Type like `permission_prompt`, `idle_prompt` |

### `PostToolUseFailure`

PostToolUseFailure fires when a tool execution fails and includes:

| Field         | Type    | Description                                                |
|---------------|---------|-----------------------------------------------------------|
| `tool_name`   | string  | Tool identifier                                           |
| `tool_input`  | object  | The original tool parameters                              |
| `error`       | string  | The error message describing what went wrong              |
| `is_interrupt`| boolean | True if user interruption caused the failure              |
| `tool_use_id` | string  | Unique identifier for this tool invocation                |

Example PostToolUseFailure input:

```json
{
  "session_id": "14946806-2900-40de-9807-94d621687af7",
  "transcript_path": "/Users/.../.claude/projects/.../14946806-2900-40de-9807-94d621687af7.jsonl",
  "cwd": "/Users/user/project",
  "permission_mode": "acceptEdits",
  "hook_event_name": "PostToolUseFailure",
  "tool_name": "Read",
  "tool_input": {
    "file_path": "/path/to/large/file.rs"
  },
  "error": "File content exceeds maximum allowed tokens.",
  "is_interrupt": false,
  "tool_use_id": "toolu_01ABC123..."
}
```

### `SubagentStart`

SubagentStart fires when Claude Code launches a subagent and includes:

| Field        | Type   | Description                                              |
|--------------|--------|----------------------------------------------------------|
| `agent_id`   | string | Short hex identifier for this subagent instance          |
| `agent_type` | string | The subagent kind (such as `Explore`, `arci:code-explorer`) |

Example SubagentStart input:

```json
{
  "session_id": "14946806-2900-40de-9807-94d621687af7",
  "transcript_path": "/Users/.../.claude/projects/.../14946806-2900-40de-9807-94d621687af7.jsonl",
  "cwd": "/Users/user/project",
  "hook_event_name": "SubagentStart",
  "agent_id": "a90c92f",
  "agent_type": "arci:code-explorer"
}
```

### `Stop`

Stop fires when Claude finishes responding and includes:

| Field             | Type    | Description                                        |
|-------------------|---------|---------------------------------------------------|
| `stop_hook_active`| boolean | Whether a stop hook is currently active            |

### `SubagentStop`

SubagentStop fires when a subagent completes and includes:

| Field                  | Type    | Description                                        |
|------------------------|---------|---------------------------------------------------|
| `agent_id`             | string  | Short hex identifier for this subagent instance    |
| `agent_transcript_path`| string  | Path to the subagent's transcript JSONL file       |
| `stop_hook_active`     | boolean | Whether a stop hook is currently active            |

Example SubagentStop input:

```json
{
  "session_id": "14946806-2900-40de-9807-94d621687af7",
  "transcript_path": "/Users/.../.claude/projects/.../14946806-2900-40de-9807-94d621687af7.jsonl",
  "cwd": "/Users/user/project",
  "permission_mode": "acceptEdits",
  "hook_event_name": "SubagentStop",
  "agent_id": "a5d1bfb",
  "agent_transcript_path": "/Users/.../.claude/projects/.../subagents/agent-a5d1bfb.jsonl",
  "stop_hook_active": false
}
```

### `SessionStart`

| Field    | Type   | Description                                              |
|----------|--------|----------------------------------------------------------|
| `source` | string | One of `startup`, `resume`, `clear`                      |
| `model`  | string | The Claude model for the session (such as `claude-opus-4-5-20251101`) |

Example SessionStart input:

```json
{
  "session_id": "a7257b2d-d264-4667-a76e-184377cbdeab",
  "transcript_path": "/Users/.../.claude/projects/.../a7257b2d-d264-4667-a76e-184377cbdeab.jsonl",
  "cwd": "/Users/user/project",
  "hook_event_name": "SessionStart",
  "source": "startup",
  "model": "claude-opus-4-5-20251101"
}
```

### `SessionEnd`

| Field    | Type   | Description                                              |
|----------|--------|----------------------------------------------------------|
| `reason` | string | Why the session ended (such as `prompt_input_exit`)      |

Example SessionEnd input:

```json
{
  "session_id": "d940cbb1-9fb0-4c1d-85bc-332952f2ca21",
  "transcript_path": "/Users/.../.claude/projects/.../d940cbb1-9fb0-4c1d-85bc-332952f2ca21.jsonl",
  "cwd": "/Users/user/project",
  "hook_event_name": "SessionEnd",
  "reason": "prompt_input_exit"
}
```

### `PreCompact`

| Field     | Type   | Description             |
|-----------|--------|-------------------------|
| `trigger` | string | One of `manual`, `auto` |

### `PermissionRequest`

| Field             | Type   | Description                  |
|-------------------|--------|------------------------------|
| `tool_name`       | string | Tool requesting permission   |
| `tool_input`      | object | Tool parameters              |
| `permission_type` | string | The requested permission kind |

## Environment variables

Claude Code injects environment variables that hook scripts can access:

`CLAUDECODE` equals "1" when running inside Claude Code. Scripts can use it to distinguish Claude Code invocations from other contexts.

`CLAUDE_PROJECT_DIR` contains the absolute path to the project root directory where Claude Code started. This is the most commonly used variable, enabling portable scripts that reference project files regardless of the hook's current working directory. Available for all hook events.

`CLAUDE_SESSION_ID` contains the current session UUID. While this is also available in the JSON stdin input, having it as an environment variable enables simpler scripts that don't need to parse JSON. Available for all hook events.

`CLAUDE_TRANSCRIPT_DIR` contains the directory path where Claude Code stores transcript files. This is the parent directory containing all session transcripts for the current project.

`CLAUDE_TRANSCRIPT_PATH` contains the full path to the current session's transcript file (a JSONL file). It mirrors `transcript_path` from the JSON input as an environment variable.

`CLAUDE_CODE_ENTRYPOINT` indicates how Claude Code started, with values like `cli` for command-line invocation.

`CLAUDE_CODE_SSE_PORT` contains the port number for Server-Sent Events communication between Claude Code components.

`CLAUDE_BASH_MAINTAIN_PROJECT_WORKING_DIR` when set to "1," indicates that bash commands should maintain the project working directory rather than changing to a different directory.

`CLAUDE_CODE_REMOTE` indicates whether the hook runs in a remote (web) environment. When set to "true," the hook executes in Claude Code's web interface. When not set or empty, the hook runs in the local command-line environment. Hooks can use this to adjust behavior based on execution context.

`CLAUDE_PLUGIN_ROOT` is available only for hooks defined within a plugin's `hooks/hooks.json` file. It contains the absolute path to the plugin directory, so plugin hooks can reference bundled scripts and resources portably. Plugin hooks typically use `${CLAUDE_PLUGIN_ROOT}/scripts/myscript.sh` patterns in the command field.

`CLAUDE_ENV_FILE` is available during SessionStart hooks. The hook can write environment variable definitions to this path, and Claude Code loads them into the session environment.

For ARCI integration, the expression language exposes these variables. Rules can access them via `{{ env("CLAUDE_PROJECT_DIR") }}` or through normalized functions like `{{ project_dir }}`.

## Session identifiers

Claude Code provides `session_id` as a common field in all hook events via the JSON stdin payload. The session ID is a UUID that persists across the entire conversation, enabling reliable session-scoped state tracking.

Every hook event includes `session_id` in the JSON input, covering SessionStart, SessionEnd, PreToolUse, PostToolUse, UserPromptSubmit, Stop, Notification, PermissionRequest, PreCompact, and SubagentStop. This consistent availability makes Claude Code fully compatible with ARCI's state store feature.

Example JSON input showing session_id:

```json
{
  "session_id": "eb5b0174-0555-4601-804e-672d68069c89",
  "transcript_path": "/Users/.../.claude/projects/.../eb5b0174-0555-4601-804e-672d68069c89.jsonl",
  "cwd": "/Users/user/project",
  "permission_mode": "default",
  "hook_event_name": "PreToolUse",
  "tool_name": "Bash",
  "tool_input": { "command": "npm test" }
}
```

The session ID enables patterns like "warn on first occurrence, block on third" where ARCI tracks state across many hook invocations within the same conversation. Rules can use `session_get`, `session_set`, and session-scoped counters with full confidence that the session ID is always available.

The `transcript_path` field provides another correlation mechanism. The path includes the session UUID and supports audit logging or conversation history access.

## Output schema

Hooks communicate decisions through a combination of exit codes and JSON output on stdout.

### Exit codes

| Exit Code | Meaning              | Behavior                                                      |
|-----------|----------------------|---------------------------------------------------------------|
| `0`       | Success              | Action proceeds; stdout parsed as JSON for structured control |
| `2`       | Blocking error       | Action blocked; stderr shown to Claude as feedback            |
| `128`     | Catastrophic failure | Reserved for fatal errors                                     |
| Other     | Non-blocking warning | Action proceeds; stderr shown to user in verbose mode         |

JSON output is only processed when exit code is 0. For exit code 2, Claude Code uses stderr content directly as the block reason.

### Common JSON output fields

When exiting with code 0, hooks can return JSON with these common fields:

| Field            | Type    | Description                                                 |
|------------------|---------|-------------------------------------------------------------|
| `continue`       | boolean | If false, halt Claude's processing entirely (default: true) |
| `stopReason`     | string  | Message shown to user when continue is false                |
| `suppressOutput` | boolean | If true, hide hook output from transcript mode              |
| `systemMessage`  | string  | Warning message displayed to the user                       |

The `continue: false` behavior differs from permission denial. Denying permission blocks only the specific tool call and provides feedback to Claude. Setting `continue: false` stops Claude entirely.

### `PreToolUse`

PreToolUse hooks can control tool execution through `hookSpecificOutput`:

```json
{
  "hookSpecificOutput": {
    "permissionDecision": "allow | deny | ask",
    "permissionDecisionReason": "string - explanation for decision",
    "updatedInput": { /* modified tool_input object */ }
  },
  "systemMessage": "optional warning to user"
}
```

| Field                      | Values  | Description                                               |
|----------------------------|---------|-----------------------------------------------------------|
| `permissionDecision`       | `allow` | Auto-approve the tool call without user confirmation      |
|                            | `deny`  | Block the tool call; reason sent to Claude                |
|                            | `ask`   | Defer to user confirmation (default behavior)             |
| `permissionDecisionReason` | string  | Explanation shown to Claude (on deny) or user             |
| `updatedInput`             | object  | Modified tool input; Claude uses this instead of original |

The `updatedInput` feature (added in v2.0.10) enables transparent sandboxing, automatic security enforcement, and convention adherence. Claude sees the modified input as if it were the original.

### `PermissionRequest`

PermissionRequest hooks control automatic permission decisions:

```json
{
  "hookSpecificOutput": {
    "decision": "approve | deny",
    "reason": "string - explanation for decision"
  }
}
```

| Field      | Values    | Description                         |
|------------|-----------|-------------------------------------|
| `decision` | `approve` | Auto-approve the permission request |
|            | `deny`    | Auto-deny the permission request    |
| `reason`   | string    | Explanation shown to user           |

### `PostToolUse`

PostToolUse hooks can inject context or block further processing:

```json
{
  "hookSpecificOutput": {
    "decision": "block",
    "reason": "string - feedback to Claude",
    "additionalContext": "string - context added to conversation"
  }
}
```

| Field               | Description                                        |
|---------------------|----------------------------------------------------|
| `decision`          | Set to `block` to stop Claude and provide feedback |
| `reason`            | Required when decision is `block`                  |
| `additionalContext` | Text appended to Claude's context window           |

### `UserPromptSubmit`

```json
{
  "hookSpecificOutput": {
    "decision": "block",
    "additionalContext": "string - context injected before processing"
  },
  "continue": false  // prevents prompt from being processed
}
```

### `Stop` and `SubagentStop`

```json
{
  "hookSpecificOutput": {
    "decision": "block",
    "reason": "string - instructions for Claude to continue"
  }
}
```

Setting `decision: "block"` prevents Claude from stopping and the reason tells Claude how to proceed.

### `SessionStart`

```json
{
  "hookSpecificOutput": {
    "additionalContext": "string - initial context for the session"
  }
}
```

SessionStart is unique in that hooks can also write to `CLAUDE_ENV_FILE` to set environment variables for the session.

## Modification capabilities

### Tool input modification (`PreToolUse`)

Claude Code fully supports tool input modification. The PreToolUse output schema includes an `updatedInput` field for transparent tool input modification. When a hook returns `updatedInput`, Claude uses the modified parameters instead of the original, unaware of the modification. Supported patterns include automatic sandboxing, security enforcement, and convention adherence.

Claude Code v2.0.10 added this capability.

### Prompt modification (`UserPromptSubmit`)

Claude Code does not support direct prompt modification. The UserPromptSubmit hook can inject extra context via the `additionalContext` field, which Claude Code adds before the conversation before Claude processes it. The hook cannot change or replace the user's original prompt text. The hook can also block prompts entirely with `continue: false`.

## ARCI integration

ARCI integrates directly with Claude Code's hooks system. The JSON-over-stdin contract, exit code semantics, and matcher syntax provide a stable foundation for policy evaluation.

ARCI contributes configuration sources at `~/.claude/arci/config.yaml` for user-level rules, `.claude/arci/config.yaml` for project rules, and `.claude/arci/config.local.yaml` for personal project settings.

ARCI parses Claude Code's camelCase hook input (hookEventType, toolName, etc.) and normalizes to snake_case internal representations. It formats output according to Claude Code's expected schemas and maps evaluation results to appropriate exit codes (0 for success, 2 for block, 128 for catastrophic failure).

## Considerations

Claude Code's hooks are well-documented and stable, having been battle-tested in production. The system supports both simple exit-code-based decisions and complex JSON output for fine-grained control.

The prompt-based hooks feature (type: "prompt") uses an LLM for context-aware decisions, which is orthogonal to ARCI's rules-based approach but could be complementary.

Claude Code merges plugin hooks with user and project hooks, which aligns well with ARCI's layered configuration model.

The CLAUDE_PROJECT_DIR environment variable enables portable scripts that work regardless of working directory, which ARCI can use.

## Hook installation for ARCI

ARCI integration with Claude Code requires adding hook entries to settings.json files. The recommended approach is to ship a Claude Code plugin that users install via `/plugin install`.

### Manual installation

To manually configure ARCI, add entries to `~/.claude/settings.json` (user-level) or `.claude/settings.json` (project-level):

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "arci hook apply",
            "timeout": 5000
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "arci hook apply",
            "timeout": 5000
          }
        ]
      }
    ]
  }
}
```

### Plugin installation

Once published, users would install ARCI via the plugin system:

```text
/plugin marketplace add tbhb/arci
/plugin install arci
```

The plugin would include a `hooks/hooks.json` file with pre-configured hook entries using `${CLAUDE_PLUGIN_ROOT}` for portable paths:

```json
{
  "PreToolUse": [
    {
      "matcher": "*",
      "hooks": [
        {
          "type": "command",
          "command": "arci hook apply --plugin-root ${CLAUDE_PLUGIN_ROOT}",
          "timeout": 5000
        }
      ]
    }
  ]
}
```

### Enterprise deployment

Enterprise administrators can use `allowManagedHooksOnly` to restrict hooks to managed sources, ensuring only approved plugins and hooks run. ARCI's fail-open semantics align well with this model since configuration errors or server unavailability won't block developer workflows.

## Plugin mechanism

Claude Code's plugin system (public beta since October 2025) provides full extensibility:

Plugin structure follows a standardized directory layout with `.claude-plugin/plugin.json` manifest, optional `commands/`, `agents/`, `skills/`, `hooks/`, and `.mcp.json` files.

Installation methods include `/plugin install` from marketplaces, npm packages, GitHub repositories, or local directories.

Hook bundling places hook configuration in `hooks/hooks.json` within the plugin. The `${CLAUDE_PLUGIN_ROOT}` environment variable enables portable script paths.

Marketplaces like the official `anthropics/claude-code` collection and community marketplaces provide discoverability and one-command installation.

This plugin system makes ARCI distribution straightforward. Users install once and the hooks are automatically configured, with updates handled through the plugin update mechanism.

## References

Official documentation: <https://code.claude.com/docs/en/hooks>
Plugins documentation: <https://code.claude.com/docs/en/plugins>
Plugin announcement: <https://www.anthropic.com/news/claude-code-plugins>
