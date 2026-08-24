package b_program_test

import (
	"runtime"
	"runtime/debug"
	"strings"
)

// runningBuildTags returns the -tags the test binary itself was built with
// (e.g. "static") so the blue and bluerun binaries built for the tests match
// the flavor under test instead of defaulting to the non-static build.
func runningBuildTags() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, setting := range info.Settings {
		if setting.Key == "-tags" {
			return strings.Join(strings.Fields(setting.Value), ",")
		}
	}
	return ""
}

func withRunningTags(args []string) []string {
	if tags := runningBuildTags(); tags != "" {
		args = append(args, "-tags", tags)
	}
	return args
}

// exeSuffix returns the filename extension required for executables on the
// host OS. Windows is special: os/exec resolves even absolute paths through
// PATHEXT (Go 1.25+), so spawning an extensionless binary fails with
// "executable file not found" and a nil ProcessState (exit code -1, no
// output) instead of running it.
func exeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}
