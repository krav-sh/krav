// Package buildmeta contains build-time information injected via ldflags.
package buildmeta

import (
	"runtime/debug"
)

// These variables are set at build time via -ldflags.
var (
	// Version is the semantic version of the build.
	// Set via ldflags or detected from debug.BuildInfo.
	Version = "DEV"

	// Commit is the git commit SHA of the build.
	// Set via ldflags.
	Commit = ""

	// Date is the build date in YYYY-MM-DD format.
	// Set via ldflags.
	Date = ""
)

func init() {
	if Version == "DEV" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" {
			Version = info.Main.Version
		}
	}
}
