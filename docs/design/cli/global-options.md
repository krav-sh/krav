# Global options

Six options apply to all commands. The root Cobra command registers them as persistent flags, and all subcommands inherit them.

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

`--project <path>`: sets the project root explicitly instead of detecting it from the current directory.

`--scope <scope>`: specifies which configuration layer to target for commands that change files. Valid values are:

- `worktree_assistant`, `worktree`
- `local_assistant`, `local`
- `project_assistant`, `project`
- `user_assistant`, `user`
- `site_assistant`, `site`

The `cli`, `env`, and `default*` scopes only support reading. You cannot target them for modifications.

`--socket <path>`: specifies the Unix socket path for daemon communication.

`--url <url>`: specifies the HTTP URL for daemon communication.

`--verbose` (`-v`): increases output verbosity. You can repeat it for more detail (`-v`, `-vv`, `-vvv`).

`--quiet` (`-q`): suppresses non-essential output. You can repeat it for less output (`-q`, `-qq`, `-qqq`). Mutually exclusive with `--verbose`.

## See also

- [environment-variables](environment-variables.md): environment variable equivalents for common options
