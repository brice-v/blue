package consts

import (
	"os"
	"runtime/debug"

	"github.com/gookit/color"
)

// VERSION is the version number of the blang repl and language
// it will be incremented as seen fit
var VERSION = func() string {
	version := "0.1.10"
	hash := ""
	os := ""
	arch := ""
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				// This is the short hash
				hash = setting.Value[:7]
			}
			if setting.Key == "GOOS" {
				os = setting.Value
			}
			if setting.Key == "GOARCH" {
				arch = setting.Value
			}
		}
	}
	if hash == "" {
		return version
	}
	return version + "-" + hash + "-" + os + "/" + arch
}()

const PARSER_ERROR_PREFIX = "ParserError: "
const PROCESS_ERROR_PREFIX = "ProcessError: "
const EVAL_ERROR_PREFIX = "EvaluatorError: "

const CORE_FILE_PATH = "<embed: core/core.b>"

const BLUE_INSTALL_PATH = "BLUE_INSTALL_PATH"
const BLUE_NO_COLOR = "BLUE_NO_COLOR"

const EMBED_FILES_PREFIX = "embed_files/"

var ErrorPrinter = color.New(color.FgRed, color.Bold).Printf
var InfoPrinter = color.New(color.FgBlue, color.Bold).Printf

func DisableColorIfNoColorEnvVarSet() {
	if os.Getenv(BLUE_NO_COLOR) != "" || os.Getenv("NO_COLOR") != "" {
		color.Disable()
	}
}
