package cmd

import (
	"context"
	"os"

	"github.com/tbhb/arci/internal/cmd/factory"
	"github.com/tbhb/arci/internal/cmd/root"
	"github.com/tbhb/arci/internal/logging"
)

type ExitCode int

const (
	ExitCodeSuccess ExitCode = 0
	ExitCodeError   ExitCode = 1
	ExitCodeCancel  ExitCode = 2

	EnvDebug = "ARCI_DEBUG"
)

func Main() ExitCode {
	ctx := context.Background()

	debug := os.Getenv(EnvDebug) == "1"
	logLevel := logging.LevelError
	if debug {
		logLevel = logging.LevelDebug
	}

	f := factory.NewFactory(factory.Opts{
		LogLevel: logLevel,
	})

	// Force config loading to trigger debug logging
	if debug {
		_, _ = f.Config()
	}

	rootCmd, err := root.NewCmdRoot(f)
	if err != nil {
		return ExitCodeError
	}

	if err = rootCmd.ExecuteContext(ctx); err != nil {
		return ExitCodeError
	}
	return ExitCodeSuccess
}
