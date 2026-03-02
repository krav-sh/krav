package config

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"

	"github.com/tbhb/toolpaths-go"

	"go.krav.sh/krav/internal/fsutil"
)

// Environment variable names for configuration overrides.
const (
	// EnvConfigFile overrides all krav*.yaml configuration files when set.
	EnvConfigFile string = "KRAV_CONFIG_FILE"
	// EnvProjectDir explicitly sets the project directory, bypassing detection.
	EnvProjectDir string = "KRAV_PROJECT_DIR"
)

// File and directory naming constants used throughout the configuration system.
const (
	AppName                   = "krav"
	DottedAppName             = "." + AppName
	XDGConfigDirName          = ".config"
	ConfigFilename            = "krav.yaml"
	DottedConfigFilename      = "." + ConfigFilename
	LocalConfigFilename       = "krav.local.yaml"
	DottedLocalConfigFilename = "." + LocalConfigFilename
	ManagedDirName            = "managed"
	RecommendedDirName        = "recommended"
	RequiredDirName           = "required"
	ProjectConfigDirName      = DottedAppName
)

// ProjectMarkerNames lists directory and file names that indicate a project
// root during ancestor traversal. Includes VCS directories and krav
// configuration markers.
var ProjectMarkerNames = []string{
	".git",
	".hg",
	".svn",
	".bzr",
	XDGConfigDirName,
	ProjectConfigDirName,
	ConfigFilename,
	DottedConfigFilename,
	LocalConfigFilename,
	DottedLocalConfigFilename,
}

// ProjectXDGConfigDirName is the relative path to the XDG-style project
// configuration directory (.config/krav).
var ProjectXDGConfigDirName = filepath.Join(XDGConfigDirName, AppName)

// ProjectConfigDirNames lists possible project configuration directory
// locations in order of increasing precedence. The last entry (.krav)
// takes precedence over earlier entries (.config/krav).
var ProjectConfigDirNames = []string{
	filepath.Join(XDGConfigDirName, AppName),
	DottedAppName,
}

// ProjectConfigFileNames lists possible project configuration file locations
// in order of increasing precedence.
var ProjectConfigFileNames = []string{
	ConfigFilename,
	DottedConfigFilename,
	filepath.Join(ProjectXDGConfigDirName, ConfigFilename),
	filepath.Join(ProjectConfigDirName, ConfigFilename),
}

// LocalConfigFileNames lists possible local (gitignored) configuration file
// locations in order of increasing precedence.
var LocalConfigFileNames = []string{
	LocalConfigFilename,
	DottedLocalConfigFilename,
	filepath.Join(ProjectXDGConfigDirName, LocalConfigFilename),
	filepath.Join(ProjectConfigDirName, LocalConfigFilename),
}

// ManagedRecommendedDirName is the relative path within system config for
// enterprise defaults that can be overridden by users and projects.
var ManagedRecommendedDirName = filepath.Join(ManagedDirName, RecommendedDirName)

// ManagedRequiredDirName is the relative path within system config for
// enterprise enforcement that cannot be overridden.
var ManagedRequiredDirName = filepath.Join(ManagedDirName, RequiredDirName)

// Paths provides access to configuration file and directory locations across
// all cascade layers for the krav configuration system.
type Paths interface {
	// Managed recommended (enterprise defaults, overridable)
	ManagedRecommendedConfigFile() string

	// System (machine-wide)
	SystemConfigFile() string

	// User (personal)
	UserConfigFile() string

	// Project (committed to VCS)
	ProjectDir() (string, error)
	ProjectConfigDir() (string, []string, error)
	ProjectConfigFile() (string, []string, error)

	// Local (gitignored overrides)
	LocalConfigFile() (string, []string, error)

	// Default paths for scaffolding new configuration
	DefaultProjectConfigDir() (string, error)
	DefaultProjectConfigFile() (string, error)
	DefaultLocalConfigFile() (string, error)

	// Custom overrides from environment variables
	CustomConfigFile() string

	// Managed required (enterprise enforcement, not overridable)
	ManagedRequiredConfigFile() string

	// Path construction helpers
	JoinProjectDir(elem ...string) (string, error)
	JoinProjectConfigDir(elem ...string) (string, error)

	// Log directories
	UserLogDir() string
	ProjectLogDir(adapter, projectPath string) (string, error)
}

// DefaultPaths implements [Paths] using platform-appropriate directory
// locations via toolpaths and filesystem detection for project roots.
type DefaultPaths struct {
	dirs       toolpaths.Dirs
	cwd        string // Empty means use os.Getwd(); set for testing
	projectDir string // Override from CLI flag; takes precedence over env and detection
}

// PathOpts configures DefaultPaths behavior.
type PathOpts struct {
	ProjectDir string // Override project directory (highest precedence)
}

// NewDefaultPaths creates a [DefaultPaths] configured for the current platform.
func NewDefaultPaths() (*DefaultPaths, error) {
	return NewDefaultPathsWithOpts(PathOpts{})
}

// NewDefaultPathsWithOpts creates a [DefaultPaths] with additional configuration options.
func NewDefaultPathsWithOpts(opts PathOpts) (*DefaultPaths, error) {
	includeXDGFallbacks := true
	dirsConfig := toolpaths.Config{
		AppName:             AppName,
		IncludeXDGFallbacks: &includeXDGFallbacks,
	}
	dirs, err := toolpaths.NewWithConfig(dirsConfig)
	if err != nil {
		return nil, err
	}

	return &DefaultPaths{
		dirs:       dirs,
		projectDir: opts.ProjectDir,
	}, nil
}

func (d *DefaultPaths) getwd() (string, error) {
	if d.cwd != "" {
		return d.cwd, nil
	}
	return os.Getwd()
}

func (d *DefaultPaths) CustomConfigFile() string {
	return os.Getenv(EnvConfigFile)
}

func (d *DefaultPaths) ManagedRecommendedConfigFile() string {
	return filepath.Join(d.dirs.SystemConfigDir(), ManagedRecommendedDirName, ConfigFilename)
}

func (d *DefaultPaths) ManagedRequiredConfigFile() string {
	return filepath.Join(d.dirs.SystemConfigDir(), ManagedRequiredDirName, ConfigFilename)
}

// ProjectDir returns the project root directory by checking in precedence order:
// 1. PathOpts.ProjectDir override (from CLI flag)
// 2. KRAV_PROJECT_DIR environment variable
// 3. Detection from current working directory being inside a config directory
// 4. Ancestor traversal looking for project markers (VCS dirs, config files)
//
// Returns an empty string if no project root can be determined.
func (d *DefaultPaths) ProjectDir() (string, error) {
	// CLI flag override (highest precedence)
	if d.projectDir != "" {
		return d.projectDir, nil
	}

	// Environment variable override
	if val, ok := os.LookupEnv(EnvProjectDir); ok {
		return val, nil
	}

	cwd, err := d.getwd()
	if err != nil {
		return "", err
	}

	// Check if we're inside a config directory and short-circuit if so
	if projectDir, ok := projectDirFromConfigPath(cwd); ok {
		return projectDir, nil
	}

	if dir, _, ok := d.dirs.FindUpFunc(cwd, ProjectMarkerNames, func(path string) bool {
		if strings.HasSuffix(path, XDGConfigDirName) {
			isDir, _ := fsutil.IsAccessibleDir(filepath.Join(path, AppName))
			return isDir
		}
		return true
	}); ok {
		return dir, nil
	}

	return "", nil
}

// projectDirFromConfigPath checks if the given path is inside a krav
// config directory and returns the project root if so.
func projectDirFromConfigPath(path string) (string, bool) {
	path = filepath.Clean(path)
	parts := strings.Split(path, string(filepath.Separator))
	isAbsolute := len(parts) > 0 && parts[0] == ""

	for i := len(parts) - 1; i >= 0; i-- {
		switch {
		case parts[i] == DottedAppName:
			return joinPathParts(parts[:i], isAbsolute), true
		case parts[i] == AppName && i > 0 && parts[i-1] == XDGConfigDirName:
			return joinPathParts(parts[:i-1], isAbsolute), true
		}
	}

	return "", false
}

// joinPathParts joins path parts, handling absolute paths and empty slices.
func joinPathParts(parts []string, isAbsolute bool) string {
	if len(parts) == 0 || (len(parts) == 1 && parts[0] == "") {
		if isAbsolute {
			return string(filepath.Separator)
		}
		return "."
	}
	result := filepath.Join(parts...)
	if isAbsolute && !strings.HasPrefix(result, string(filepath.Separator)) {
		result = string(filepath.Separator) + result
	}
	return result
}

// ProjectConfigDir returns the project configuration directory and any
// shadowed alternatives. Returns (selected, shadowed, error).
func (d *DefaultPaths) ProjectConfigDir() (string, []string, error) {
	return d.matchProjectDirs(ProjectConfigDirNames)
}

// ProjectConfigFile returns the project configuration file and any
// shadowed alternatives. Returns (selected, shadowed, error).
func (d *DefaultPaths) ProjectConfigFile() (string, []string, error) {
	return d.matchProjectFiles(ProjectConfigFileNames)
}

// LocalConfigFile returns the local (gitignored) configuration file and any
// shadowed alternatives. Returns (selected, shadowed, error).
func (d *DefaultPaths) LocalConfigFile() (string, []string, error) {
	return d.matchProjectFiles(LocalConfigFileNames)
}

// matchProjectDirs finds all matching directories from names within the project
// directory, returning the highest-precedence match and any shadowed paths.
func (d *DefaultPaths) matchProjectDirs(names []string) (string, []string, error) {
	projectDir, err := d.ProjectDir()
	if err != nil {
		return "", nil, err
	}

	matches := []string{}
	for _, relName := range names {
		if isDir, _ := fsutil.IsAccessibleDir(filepath.Join(projectDir, relName)); isDir {
			matches = append(matches, filepath.Join(projectDir, relName))
		}
	}

	if len(matches) == 0 {
		return "", nil, nil
	}

	return matches[len(matches)-1], matches[:len(matches)-1], nil
}

// matchProjectFiles finds all matching files from names within the project
// directory, returning the highest-precedence match and any shadowed paths.
func (d *DefaultPaths) matchProjectFiles(names []string) (string, []string, error) {
	projectDir, err := d.ProjectDir()
	if err != nil {
		return "", nil, err
	}

	matches := []string{}
	for _, relName := range names {
		if isFile, _ := fsutil.IsAccessibleFile(filepath.Join(projectDir, relName)); isFile {
			matches = append(matches, filepath.Join(projectDir, relName))
		}
	}

	if len(matches) == 0 {
		return "", nil, nil
	}

	return matches[len(matches)-1], matches[:len(matches)-1], nil
}

// DefaultProjectConfigDir returns the default path for a new project
// configuration directory (.krav within the project root).
func (d *DefaultPaths) DefaultProjectConfigDir() (string, error) {
	return d.JoinProjectDir(ProjectConfigDirNames[len(ProjectConfigDirNames)-1])
}

// DefaultProjectConfigFile returns the default path for a new project
// configuration file (.krav/krav.yaml).
func (d *DefaultPaths) DefaultProjectConfigFile() (string, error) {
	return d.JoinProjectDir(ProjectConfigFileNames[len(ProjectConfigFileNames)-1])
}

// DefaultLocalConfigFile returns the default path for a new local
// configuration file (.krav/krav.local.yaml).
func (d *DefaultPaths) DefaultLocalConfigFile() (string, error) {
	return d.JoinProjectDir(LocalConfigFileNames[len(LocalConfigFileNames)-1])
}

// SystemConfigFile returns the system-wide configuration file path.
func (d *DefaultPaths) SystemConfigFile() string {
	return filepath.Join(d.dirs.SystemConfigDir(), ConfigFilename)
}

// UserConfigFile returns the user's personal configuration file path.
func (d *DefaultPaths) UserConfigFile() string {
	return filepath.Join(d.dirs.UserConfigDir(), ConfigFilename)
}

// JoinProjectDir joins path elements to the project directory root.
func (d *DefaultPaths) JoinProjectDir(elem ...string) (string, error) {
	projectDir, err := d.ProjectDir()
	if err != nil {
		return "", err
	}
	return fsutil.JoinBasePath(projectDir, elem...), nil
}

// JoinProjectConfigDir joins path elements to the project configuration directory.
func (d *DefaultPaths) JoinProjectConfigDir(elem ...string) (string, error) {
	configDir, _, err := d.ProjectConfigDir()
	if err != nil {
		return "", err
	}
	return fsutil.JoinBasePath(configDir, elem...), nil
}

// UserLogDir returns the user's log directory.
func (d *DefaultPaths) UserLogDir() string {
	return d.dirs.UserLogDir()
}

// ProjectLogDir returns the project-specific log directory using Hive-style
// partitioning. The returned path has the structure:
//
//	<UserLogDir>/adapter=<adapter>/project=<hash>/
func (d *DefaultPaths) ProjectLogDir(adapter, projectPath string) (string, error) {
	hash, err := ProjectPathHash(projectPath)
	if err != nil {
		return "", err
	}
	return filepath.Join(
		d.UserLogDir(),
		"adapter="+adapter,
		"project="+hash,
	), nil
}

// ProjectPathHash returns a truncated SHA256 hash of the canonical absolute
// project path. The hash is the first 12 lowercase hex characters (48 bits).
func ProjectPathHash(projectPath string) (string, error) {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return "", err
	}
	canonicalPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(canonicalPath))
	return hex.EncodeToString(hash[:])[:12], nil
}

var _ Paths = (*DefaultPaths)(nil)
