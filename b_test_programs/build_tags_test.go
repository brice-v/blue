package b_program_test

import (
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
