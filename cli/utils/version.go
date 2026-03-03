package utils

import "runtime/debug"

// The `Version` variable contains the application version. It's read from Go's build metadata at runtime:
// - Release builds: semver from git tag (e.g., "v1.0.0")
// - Development builds: "(devel)" with helpful installation message
var Version = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version == "" || info.Main.Version == "(devel)" {
			return "-devel - install with: go install ytsruh.com/envoy@latest"
		}
		return info.Main.Version
	}
	return "unknown"
}()
