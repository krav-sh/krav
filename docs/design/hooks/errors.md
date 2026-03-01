# Hook error troubleshooting

This document covers troubleshooting workflows, diagnostic commands, and debug logging for the hook subsystem.

## Troubleshooting workflows

Common troubleshooting scenarios and how to approach them.

### A rule is not matching

Start with `arci hook policy test <rule-selector> --input @sample.json` to see if the rule matches against known input. If it does not match, the test command shows which part of the condition evaluated to false.

Confirm that the configuration enables the rule with `arci hook policy explain <rule-selector>`. The output shows enabled status and source file.

Verify the event type filter. A rule with `events: [PostToolUse]` does not match `PreToolUse` hooks.

Check priority and terminal rules. A higher-priority terminal rule may stop evaluation before your rule runs. Use `arci hook logs --rule <rule-selector>` to see if the engine evaluates the rule at all.

Enable debug logging with `ARCI_LOG_LEVEL=debug` to see expression evaluation details.

### A rule is matching when it should not

Use `arci hook policy test <rule-selector> --input @sample.json` with input that should not match. The test output shows the evaluation trace.

Check for overly broad conditions. A condition like `tool.name =~ /rm/` matches `transform` as well as `rm`.

Review rule precedence. Lower-precedence rules may override your rule's decision.

Use the dashboard rule tester for interactive exploration of complex conditions.

### An action is not executing

Check that the action type is compatible with the hook type. Some actions only make sense for certain events.

Review timeout configuration. Shell commands have a default timeout, and the runtime may stop long-running commands before completion.

Check action handler output with debug logging. ARCI logs invalid output from an action handler as a warning.

For shell actions, verify the command path is correct and executable. The shell action runs in the project directory by default.

## Diagnostic commands

### ARCI hook policy explain

The explain command shows everything about a rule:

```text
$ arci hook policy explain block-rm-rf

Rule: block-rm-rf
Source: ~/.config/arci/rules.yaml:15
Priority: critical
Enabled: true
Events: PreToolUse

Condition:
  tool.name == "bash" && input.command =~ /\brm\b.*-rf/

Result: block
Message: "Recursive force delete is not allowed. Please confirm this action."

Actions:
  - log: { level: "warn", message: "Attempted rm -rf" }

Match history (last 7 days):
  - 2024-01-14 15:30:22: matched (blocked)
  - 2024-01-12 09:15:01: matched (blocked)
```

## Debug logging

Enable detailed logging for hook evaluation:

```bash
# All debug output
ARCI_LOG_LEVEL=debug arci run --event PreToolUse
```

The `ARCI_LOG_LEVEL` environment variable sets the global log level. For more on diagnostic tracing, see [CLI logging](../cli/logging.md).

## See also

- [CLI errors](../cli/errors.md): CLI error presentation and health checks
- [Server errors](../server/errors.md): server troubleshooting and recovery
- [Hook event logging](logging.md): the `arci hook apply` output contract and event log schema
