package shared

import "runtime/debug"

var Version = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version == "" || info.Main.Version == "(devel)" {
			return "-devel - install with: go install ytsruh.com/envoy@latest"
		}
		return info.Main.Version
	}
	return "unknown"
}()
