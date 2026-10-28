package b_program_test

import (
	"runtime"
	"runtime/debug"
	"strings"
)

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

func exeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}
