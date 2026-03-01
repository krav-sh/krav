# Global options

Several options apply to all commands. These are registered as persistent flags on the root Cobra command and inherited by all subcommands.

## Persistent flags

```go
// Global flags are registered on the root command and inherited by all subcommands.
func addGlobalFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().CountP("quiet", "q", "Decrease output verbosity (can be repeated: -q, -qq, -qqq)")
	cmd.PersistentFlags().CountP("verbose", "v", "Increase output verbosity (can be repeated: -v, -vv, -vvv)")
	cmd.MarkFlagsMutuallyExclusive("quiet", "verbose")
}
```

## Flag reference

`--project <path>` — Sets the project root explicitly instead of detecting it from the current directory.

`--scope <scope>` — Specifies which configuration layer to target for commands that modify files. Valid values are:

- `worktree_assistant`, `worktree`
- `local_assistant`, `local`
- `project_assistant`, `project`
- `user_assistant`, `user`
- `site_assistant`, `site`

Note that `cli`, `env`, and `default*` scopes are read-only and cannot be targeted for modifications.

`--socket <path>` — Specifies the Unix socket path for daemon communication.

`--url <url>` — Specifies the HTTP URL for daemon communication.

`--verbose` (`-v`) — Increases output verbosity. Can be repeated for more detail (`-v`, `-vv`, `-vvv`).

`--quiet` (`-q`) — Suppresses non-essential output. Can be repeated for less output (`-q`, `-qq`, `-qqq`). Mutually exclusive with `--verbose`.

## See also

- [environment-variables](environment-variables.md) — environment variable equivalents for common options
