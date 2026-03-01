// Package logging provides a structured logging interface for krav.
// It wraps charmbracelet/log to provide colorful console output, logfmt,
// and JSON formats. All loggers must be explicitly instantiated; there is
// no global logger.
package logging

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/mattn/go-isatty"
)

// Format specifies the output format for log entries.
type Format string

const (
	// FormatAuto selects the format based on whether output is a TTY.
	// It uses console format for TTY output, logfmt otherwise.
	FormatAuto Format = "auto"
	// FormatConsole produces colorful human-readable output suitable for
	// terminal display. Colors are disabled when output is not a TTY.
	FormatConsole Format = "console"
	// FormatLogfmt produces logfmt-style output (key=value pairs).
	FormatLogfmt Format = "logfmt"
	// FormatJSON produces JSON-formatted output suitable for log aggregation.
	FormatJSON Format = "json"
)

// ParseFormat converts a string to a Format value. It accepts "auto", "console",
// "logfmt", and "json" (case-insensitive). Empty string defaults to "auto".
// Unknown values return an error.
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(s) {
	case "auto", "":
		return FormatAuto, nil
	case "console", "text":
		return FormatConsole, nil
	case "logfmt":
		return FormatLogfmt, nil
	case "json":
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("unknown log format: %q", s)
	}
}

// String returns the string representation of the format.
func (f Format) String() string {
	return string(f)
}

// Level specifies the minimum severity for log output.
type Level int

// Level constants match charmbracelet/log's internal values for direct conversion.
const (
	LevelDebug Level = -4
	LevelInfo  Level = 0
	LevelWarn  Level = 4
	LevelError Level = 8
	LevelFatal Level = 12
)

// ParseLevel converts a string to a Level value. It accepts "debug", "info",
// "warn", "error", and "fatal" (case-insensitive). Unknown values return an error.
func ParseLevel(s string) (Level, error) {
	switch strings.ToLower(s) {
	case "debug":
		return LevelDebug, nil
	case "info", "":
		return LevelInfo, nil
	case "warn", "warning":
		return LevelWarn, nil
	case "error":
		return LevelError, nil
	case "fatal":
		return LevelFatal, nil
	default:
		return 0, fmt.Errorf("unknown log level: %q", s)
	}
}

// String returns the string representation of the level.
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	case LevelFatal:
		return "fatal"
	default:
		return fmt.Sprintf("Level(%d)", l)
	}
}

// Logger defines the interface for structured logging operations.
type Logger interface {
	Debug(msg string, keyvals ...any)
	Debugf(format string, args ...any)
	Info(msg string, keyvals ...any)
	Infof(format string, args ...any)
	Warn(msg string, keyvals ...any)
	Warnf(format string, args ...any)
	Error(msg string, keyvals ...any)
	Errorf(format string, args ...any)
	Fatal(msg string, keyvals ...any)
	Fatalf(format string, args ...any)
	With(keyvals ...any) Logger
	WithPrefix(prefix string) Logger
}

// LoggerOpts configures DefaultLogger behavior.
type LoggerOpts struct {
	Output          io.Writer
	Format          Format
	Level           Level
	Prefix          string
	ReportTimestamp bool
	ReportCaller    bool
}

// DefaultLogger implements Logger using charmbracelet/log.
type DefaultLogger struct {
	logger *log.Logger
}

// NewDefaultLogger creates a Logger with sensible defaults.
func NewDefaultLogger() *DefaultLogger {
	return NewDefaultLoggerWithOpts(LoggerOpts{
		ReportTimestamp: true,
	})
}

// NewDefaultLoggerWithOpts creates a Logger with the specified options.
func NewDefaultLoggerWithOpts(opts LoggerOpts) *DefaultLogger {
	output := opts.Output
	if output == nil {
		output = os.Stderr
	}

	logOpts := log.Options{
		Level:           log.Level(opts.Level),
		Prefix:          opts.Prefix,
		ReportTimestamp: opts.ReportTimestamp,
		ReportCaller:    opts.ReportCaller,
	}

	logOpts.Formatter = resolveFormatter(opts.Format, output)

	return &DefaultLogger{
		logger: log.NewWithOptions(output, logOpts),
	}
}

// resolveFormatter returns the appropriate log.Formatter for the given format.
func resolveFormatter(format Format, output io.Writer) log.Formatter {
	switch format {
	case FormatJSON:
		return log.JSONFormatter
	case FormatLogfmt:
		return log.LogfmtFormatter
	case FormatConsole:
		return log.TextFormatter
	case FormatAuto:
		if isTTY(output) {
			return log.TextFormatter
		}
		return log.LogfmtFormatter
	default:
		if isTTY(output) {
			return log.TextFormatter
		}
		return log.LogfmtFormatter
	}
}

// isTTY returns true if the output is a terminal.
func isTTY(output io.Writer) bool {
	if f, ok := output.(*os.File); ok {
		return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
	}
	return false
}

func (l *DefaultLogger) Debug(msg string, keyvals ...any) {
	l.logger.Debug(msg, keyvals...)
}

func (l *DefaultLogger) Debugf(format string, args ...any) {
	l.logger.Debugf(format, args...)
}

func (l *DefaultLogger) Info(msg string, keyvals ...any) {
	l.logger.Info(msg, keyvals...)
}

func (l *DefaultLogger) Infof(format string, args ...any) {
	l.logger.Infof(format, args...)
}

func (l *DefaultLogger) Warn(msg string, keyvals ...any) {
	l.logger.Warn(msg, keyvals...)
}

func (l *DefaultLogger) Warnf(format string, args ...any) {
	l.logger.Warnf(format, args...)
}

func (l *DefaultLogger) Error(msg string, keyvals ...any) {
	l.logger.Error(msg, keyvals...)
}

func (l *DefaultLogger) Errorf(format string, args ...any) {
	l.logger.Errorf(format, args...)
}

func (l *DefaultLogger) Fatal(msg string, keyvals ...any) {
	l.logger.Fatal(msg, keyvals...)
}

func (l *DefaultLogger) Fatalf(format string, args ...any) {
	l.logger.Fatalf(format, args...)
}

func (l *DefaultLogger) With(keyvals ...any) Logger {
	return &DefaultLogger{
		logger: l.logger.With(keyvals...),
	}
}

func (l *DefaultLogger) WithPrefix(prefix string) Logger {
	return &DefaultLogger{
		logger: l.logger.WithPrefix(prefix),
	}
}

var _ Logger = (*DefaultLogger)(nil)
