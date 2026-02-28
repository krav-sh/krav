package settings

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"

	"github.com/tbhb/arci/internal/config"
	"github.com/tbhb/arci/internal/logging"
)

// LoadOpts configures cascade loading behavior.
type LoadOpts struct {
	Paths    config.Paths   // provides file locations for all cascade layers
	Defaults map[string]any // built-in default values
	Exclude  []Source       // sources to skip during loading
	Flags    *pflag.FlagSet // CLI flags to load
	Logger   logging.Logger // logger for warnings during loading (nil disables logging)
}

// Excludes returns true if the source should be skipped during loading.
func (opts *LoadOpts) Excludes(source Source) bool {
	return slices.Contains(opts.Exclude, source)
}

// Includes returns true if the source should be loaded.
func (opts *LoadOpts) Includes(source Source) bool {
	return !opts.Excludes(source)
}

// HasDefaults returns true if default values were provided.
func (opts *LoadOpts) HasDefaults() bool {
	return len(opts.Defaults) > 0
}

// HasFlags returns true if CLI flags were provided.
func (opts *LoadOpts) HasFlags() bool {
	return opts.Flags != nil
}

// warn logs a warning message if a logger is configured.
func (opts *LoadOpts) warn(msg string, keyvals ...any) {
	if opts.Logger != nil {
		opts.Logger.Warn(msg, keyvals...)
	}
}

// debug logs a debug message if a logger is configured.
func (opts *LoadOpts) debug(msg string, keyvals ...any) {
	if opts.Logger != nil {
		opts.Logger.Debug(msg, keyvals...)
	}
}

// excludedEnvSuffixes lists environment variable suffixes that should not be
// loaded as configuration values. These variables control loader behavior
// rather than providing config values.
var excludedEnvSuffixes = []string{
	"DEBUG",       // ARCI_DEBUG - enable debug logging
	"CONFIG_FILE", // ARCI_CONFIG_FILE - custom config override
	"PROJECT_DIR", // ARCI_PROJECT_DIR - project directory override
}

// isExcludedEnvVar returns true if the environment variable key (without prefix)
// matches one of the excluded suffixes.
func isExcludedEnvVar(key string) bool {
	upperKey := strings.ToUpper(key)
	for _, suffix := range excludedEnvSuffixes {
		if upperKey == suffix {
			return true
		}
	}
	return false
}

// LoadDefaultCascade loads configuration from all sources in precedence order.
func LoadDefaultCascade(opts LoadOpts) (*DefaultCascade, error) {
	opts.debug("loading configuration cascade")
	cfg := NewDefaultCascade()

	// Load defaults first (always)
	loadDefaults(opts, cfg)

	// Load file cascade (custom config or normal cascade)
	if err := loadFileCascade(opts, cfg); err != nil {
		return nil, err
	}

	// Environment variables (always, after files)
	loadEnv(opts, cfg, "ARCI_")

	// CLI flags (always, after env)
	if err := loadFlags(opts, cfg); err != nil {
		return nil, err
	}

	// Managed required (always last, even with custom config)
	if err := loadManagedRequired(opts, cfg); err != nil {
		return nil, err
	}

	opts.debug("configuration cascade loaded", "sources", len(cfg.layers), "keys", len(cfg.merged.Keys()))
	return cfg, nil
}

// loadFileCascade loads configuration files. If ARCI_CONFIG_FILE is set,
// it loads only that file. Otherwise, it loads the normal file cascade.
func loadFileCascade(opts LoadOpts, cfg *DefaultCascade) error {
	customConfigFile := os.Getenv(config.EnvConfigFile)
	if customConfigFile != "" {
		opts.debug("using custom config file", "path", customConfigFile)
		return loadFile(opts, cfg, SourceCustom, customConfigFile)
	}
	return loadNormalFileCascade(opts, cfg)
}

// loadNormalFileCascade loads the standard file cascade in precedence order.
func loadNormalFileCascade(opts LoadOpts, cfg *DefaultCascade) error {
	if err := loadManagedRecommended(opts, cfg); err != nil {
		return err
	}
	if err := loadSystem(opts, cfg); err != nil {
		return err
	}
	if err := loadUser(opts, cfg); err != nil {
		return err
	}
	if err := loadProject(opts, cfg); err != nil {
		return err
	}
	return loadProjectLocal(opts, cfg)
}

func loadDefaults(opts LoadOpts, cfg *DefaultCascade) {
	if opts.Excludes(SourceDefaults) {
		opts.debug("skipping source", "source", SourceDefaults, "reason", "excluded")
		return
	}

	if !opts.HasDefaults() {
		opts.debug("skipping source", "source", SourceDefaults, "reason", "no defaults provided")
		return
	}

	k := koanf.New(".")
	for key, val := range opts.Defaults {
		_ = k.Set(key, val)
	}
	cfg.setLayer(k, SourceDefaults, "", "built-in defaults")
	opts.debug("loaded source", "source", SourceDefaults, "keys", len(k.Keys()))
}

// loadFileWithPolicy loads a configuration file with either fail-open or fail-closed semantics.
func loadFileWithPolicy(opts LoadOpts, cfg *DefaultCascade, source Source, path string, failOpen bool) error {
	if opts.Excludes(source) {
		opts.debug("skipping source", "source", source, "reason", "excluded")
		return nil
	}

	if path == "" {
		opts.debug("skipping source", "source", source, "reason", "no path configured")
		return nil
	}

	// Check if file exists
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			opts.debug("skipping source", "source", source, "reason", "file not found", "path", path)
			return nil
		}
		if failOpen {
			opts.warn("cannot access config file, skipping", "path", path, "error", err)
			return nil
		}
		return fmt.Errorf("cannot access config file %s: %w", path, err)
	}

	// Load and parse
	k := koanf.New(".")
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		if failOpen {
			opts.warn("failed to parse config file, skipping", "path", path, "error", err)
			return nil
		}
		return fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	cfg.setLayer(k, source, path, "")
	opts.debug("loaded source", "source", source, "path", path, "keys", len(k.Keys()))
	return nil
}

func loadFile(opts LoadOpts, cfg *DefaultCascade, source Source, path string) error {
	return loadFileWithPolicy(opts, cfg, source, path, true)
}

func loadManagedRecommended(opts LoadOpts, cfg *DefaultCascade) error {
	if opts.Paths == nil {
		return nil
	}
	return loadFileWithPolicy(opts, cfg, SourceManagedRecommended, opts.Paths.ManagedRecommendedConfigFile(), true)
}

func loadSystem(opts LoadOpts, cfg *DefaultCascade) error {
	if opts.Paths == nil {
		return nil
	}
	return loadFileWithPolicy(opts, cfg, SourceSystem, opts.Paths.SystemConfigFile(), true)
}

func loadUser(opts LoadOpts, cfg *DefaultCascade) error {
	if opts.Paths == nil {
		return nil
	}
	return loadFileWithPolicy(opts, cfg, SourceUser, opts.Paths.UserConfigFile(), true)
}

func loadProject(opts LoadOpts, cfg *DefaultCascade) error {
	if opts.Paths == nil {
		opts.debug("skipping source", "source", SourceProject, "reason", "no paths configured")
		return nil
	}

	projectConfig, ignored, err := opts.Paths.ProjectConfigFile()
	if err != nil {
		return err
	}

	if projectConfig == "" {
		opts.debug("skipping source", "source", SourceProject, "reason", "no project directory found")
		return nil
	}

	if len(ignored) > 0 {
		opts.warn("multiple config files found, using highest precedence",
			"using", projectConfig,
			"ignoring", ignored)
	}

	return loadFileWithPolicy(opts, cfg, SourceProject, projectConfig, true)
}

func loadProjectLocal(opts LoadOpts, cfg *DefaultCascade) error {
	if opts.Paths == nil {
		opts.debug("skipping source", "source", SourceProjectLocal, "reason", "no paths configured")
		return nil
	}

	localConfig, ignored, err := opts.Paths.LocalConfigFile()
	if err != nil {
		return err
	}

	if localConfig == "" {
		opts.debug("skipping source", "source", SourceProjectLocal, "reason", "no project directory found")
		return nil
	}

	if len(ignored) > 0 {
		opts.warn("multiple local config files found, using highest precedence",
			"using", localConfig,
			"ignoring", ignored)
	}

	return loadFileWithPolicy(opts, cfg, SourceProjectLocal, localConfig, true)
}

func loadManagedRequired(opts LoadOpts, cfg *DefaultCascade) error {
	if opts.Paths == nil {
		opts.debug("skipping source", "source", SourceManagedRequired, "reason", "no paths configured")
		return nil
	}

	path := opts.Paths.ManagedRequiredConfigFile()

	// Check existence first - missing managed required is fine
	if _, err := os.Stat(path); os.IsNotExist(err) {
		opts.debug("skipping source", "source", SourceManagedRequired, "reason", "file not found", "path", path)
		return nil
	}

	// But if it exists and fails to parse, that's an error (fail closed)
	return loadFileWithPolicy(opts, cfg, SourceManagedRequired, path, false)
}

func loadEnv(opts LoadOpts, cfg *DefaultCascade, prefix string) {
	if opts.Excludes(SourceEnvironment) {
		opts.debug("skipping source", "source", SourceEnvironment, "reason", "excluded")
		return
	}

	k := koanf.New(".")
	_ = k.Load(env.Provider(".", env.Opt{
		Prefix: prefix,
		TransformFunc: func(key, val string) (string, any) {
			keyWithoutPrefix := strings.TrimPrefix(key, prefix)

			if isExcludedEnvVar(keyWithoutPrefix) {
				return "", nil
			}

			key = strings.ToLower(keyWithoutPrefix)
			key = strings.ReplaceAll(key, "__", ".")
			if strings.Contains(val, " ") {
				return key, strings.Split(val, " ")
			}
			return key, val
		},
	}), nil)

	if k.All() == nil || len(k.All()) == 0 {
		opts.debug("skipping source", "source", SourceEnvironment, "reason", "no matching env vars", "prefix", prefix)
		return
	}

	cfg.setLayer(k, SourceEnvironment, "", fmt.Sprintf("environment variables with prefix %s", prefix))
	opts.debug("loaded source", "source", SourceEnvironment, "prefix", prefix, "keys", len(k.Keys()))
}

func loadFlags(opts LoadOpts, cfg *DefaultCascade) error {
	if opts.Excludes(SourceCLI) {
		opts.debug("skipping source", "source", SourceCLI, "reason", "excluded")
		return nil
	}

	if opts.Flags == nil {
		opts.debug("skipping source", "source", SourceCLI, "reason", "no flags provided")
		return nil
	}

	k := koanf.New(".")
	if err := k.Load(posflag.Provider(opts.Flags, ".", k), nil); err != nil {
		return err
	}

	if k.All() == nil || len(k.All()) == 0 {
		opts.debug("skipping source", "source", SourceCLI, "reason", "no flag values set")
		return nil
	}

	cfg.setLayer(k, SourceCLI, "", "CLI flags")
	opts.debug("loaded source", "source", SourceCLI, "keys", len(k.Keys()))
	return nil
}

// Load loads configuration from all cascade sources and unmarshals it to a Config struct.
func Load(opts LoadOpts) (*Config, error) {
	cascade, err := LoadDefaultCascade(opts)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if unmarshalErr := cascade.Unmarshal("", &cfg); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", unmarshalErr)
	}

	return &cfg, nil
}
