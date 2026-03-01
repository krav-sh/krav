# Exit codes

The command-line tool uses consistent exit codes across commands to signal success, failure mode, and (for the `apply` command) policy decisions.

## General exit codes

```go
const (
	// ExitSuccess indicates the command was successful.
	ExitSuccess = 0
	// ExitError indicates the command failed with a general error.
	ExitError = 1
	// ExitUsageError indicates the command failed due to invalid input.
	ExitUsageError = 2
	// ExitConfigError indicates the command failed due to invalid configuration.
	ExitConfigError = 3
)
```

- **0 (`ExitSuccess`)**: The command completed successfully.
- **1 (`ExitError`)**: The command failed with a general error.
- **2 (`ExitUsageError`)**: The command failed due to invalid input (bad flags, missing arguments).
- **3 (`ExitConfigError`)**: The command failed due to invalid configuration.

## Hook apply exit codes

The `arci hook apply` command uses a subset of exit codes with policy-specific meanings:

- **0**: The operation should proceed (possibly with output modifications).
- **2**: A deny decision from a validation rule blocks the operation.
- **128**: Catastrophic failure; something went seriously wrong during evaluation.

## See also

- [hook](commands/hook.md): the command group that uses hook apply exit codes
