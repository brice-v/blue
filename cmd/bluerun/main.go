//go:build minivm

// bluerun is the minimal blue runner: it embeds no lexer, parser or
// compiler and only executes precompiled program images (see package bluec).
//
// Usage:
//
//	bluerun app.bluec [args...]   run a sidecar binary image
//	bluerun [args...]           run an image APPENDED to this executable
//	                            (produced by: blue pack -o myapp main.b)
package main

import (
	"fmt"
	"os"
	"strings"

	"blue/bluec"
	"blue/consts"
	"blue/object"
	"blue/runner"
)

func main() {
	consts.DisableColorIfNoColorEnvVarSet()

	// If this executable carries an APPENDED payload (it was produced by
	// `blue pack`), always run that payload and forward every argument to
	// the program. Otherwise behave as a plain image runner where the
	// first argument is the path of a sidecar .bluec image.
	if exeBytes, err := readOwnExecutable(); err == nil {
		if payload, ok := bluec.FindAppendedPayload(exeBytes); ok {
			bc, derr := bluec.Decode(payload, true)
			if derr != nil {
				printDecodeError(derr)
				os.Exit(1)
			}
			object.SetProgramArgs(os.Args)
			os.Exit(runner.RunBytecode(bc, false, false))
		}
	}

	args := os.Args[1:]
	if len(args) == 0 {
		consts.ErrorPrinter("no compiled blue program found.\nusage:\n  %s app.bluec        run a sidecar image\n  pack one into an executable with: blue pack -o <name> main.b\n", os.Args[0])
		os.Exit(1)
	}

	programPath := args[0]
	data, err := os.ReadFile(programPath)
	if err != nil {
		consts.ErrorPrinter("error trying to read `%s`. error: %s\n", programPath, err.Error())
		os.Exit(1)
	}
	if !bluec.SniffMagic(data) && !strings.HasSuffix(strings.ToLower(programPath), ".bluec") {
		consts.ErrorPrinter("`%s` is not a compiled blue image (missing BLUEBC magic). compile it first with:\n  blue compile -o %s.bluec %s\n", programPath, programPath, programPath)
		os.Exit(1)
	}
	bc, err := bluec.Decode(data, true)
	if err != nil {
		printDecodeError(err)
		os.Exit(1)
	}
	// Forward everything after the program path so ARGV looks like
	// [runner-name, forwarded...], matching `blue prog.b args...`.
	object.SetProgramArgs(append([]string{os.Args[0]}, args[1:]...))
	os.Exit(runner.RunBytecode(bc, false, false))
}

func readOwnExecutable() ([]byte, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	return os.ReadFile(exePath)
}

// printDecodeError renders loader failures with actionable context.
func printDecodeError(err error) {
	switch {
	case err == bluec.ErrBadMagic:
		consts.ErrorPrinter("%s: not a blue binary container (BLUEBC)\n", consts.VM_ERROR_PREFIX)
	case err == bluec.ErrTruncated:
		consts.ErrorPrinter("%s: container is truncated or corrupted (payload size mismatch)\n", consts.VM_ERROR_PREFIX)
	default:
		consts.ErrorPrinter("%s%s\n", consts.VM_ERROR_PREFIX, err.Error())
	}
	fmt.Fprintln(os.Stderr, "recompile this program against the running build:")
	fmt.Fprintln(os.Stderr, "  blue compile -o out.bluec main.b")
}
