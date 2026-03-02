package factory

import (
	"os"
	"sync"

	"go.krav.sh/krav/internal/cmdutil"
	"go.krav.sh/krav/internal/config"
	"go.krav.sh/krav/internal/config/settings"
	"go.krav.sh/krav/internal/logging"
)

// Opts configures factory creation.
type Opts struct {
	LogFormat logging.Format
	LogLevel  logging.Level
	LogFile   string
}

// NewFactory creates a new cmdutil.Factory with the given options.
func NewFactory(opts Opts) *cmdutil.Factory {
	f := &cmdutil.Factory{
		Logger: loggerFunc(opts),
		Paths:  pathsFunc(),
	}

	f.Config = configFunc(f)

	return f
}

func loggerFunc(opts Opts) func() logging.Logger {
	var logger logging.Logger
	var loggerOnce sync.Once

	return func() logging.Logger {
		loggerOnce.Do(func() {
			logger = createLogger(opts.LogFormat, opts.LogLevel, opts.LogFile)
		})

		return logger
	}
}

func createLogger(format logging.Format, level logging.Level, file string) logging.Logger {
	if file != "" {
		//nolint:gosec // log file permissions are intentionally more permissive
		f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			panic(err)
		}
		return logging.NewDefaultLoggerWithOpts(logging.LoggerOpts{
			Output:          f,
			Format:          format,
			Level:           level,
			ReportTimestamp: true,
		})
	}

	return logging.NewDefaultLoggerWithOpts(logging.LoggerOpts{
		Output:          os.Stderr,
		Format:          format,
		Level:           level,
		ReportTimestamp: true,
	})
}

func pathsFunc() func() (config.Paths, error) {
	var paths config.Paths
	var pathsOnce sync.Once
	var err error

	return func() (config.Paths, error) {
		pathsOnce.Do(func() {
			paths, err = config.NewDefaultPaths()
		})

		if err != nil {
			return nil, err
		}

		return paths, nil
	}
}

func configFunc(f *cmdutil.Factory) func() (*settings.Config, error) {
	var cfg *settings.Config
	var cfgOnce sync.Once
	paths, err := f.Paths()
	if err != nil {
		return func() (*settings.Config, error) {
			return nil, err
		}
	}

	return func() (*settings.Config, error) {
		cfgOnce.Do(func() {
			cfg, err = settings.Load(settings.LoadOpts{
				Paths:  paths,
				Logger: f.Logger(),
			})
		})

		if err != nil {
			return nil, err
		}

		return cfg, nil
	}
}
