package importguard

import (
	"os/exec"
	"runtime/debug"
	"strings"
	"testing"
)

// This test guards the minimal VM build (see blue-binary-plan.md, Phase 4):
// the runtime packages must form a closed set with NO imports of the
// lexing/parsing/compiling toolchain, so `cmd/bluerun` (built with
// -tags minivm) can execute precompiled .bbc images without embedding the
// whole toolchain. If this test fails, a heavy import crept back in.
//
// Two tiers:
//   - strict: code, token, consts, util, binc, object, vm, blueutil may not
//     import lexer/parser/compiler/ast/repl/cmd at all
//   - runner: package runner additionally formats error traces via
//     lexer.GetErrorLineMessage (pure data embeds), so only
//     parser/compiler/ast/repl/cmd are forbidden for it

var forbiddenForAll = []string{
	"blue/parser",
	"blue/compiler",
	"blue/ast",
	"blue/repl",
	"blue/cmd",
}

var forbiddenExtraStrict = []string{"blue/lexer"}

var strictPackages = []string{
	"./code", "./token", "./consts", "./util", "./binc", "./object", "./vm", "./blueutil",
}

var runnerPackages = []string{"./runner"}

// runningBuildTags returns the -tags the test binary itself was built with
// so the child go command evaluates the same build configuration. Without
// this, listing e.g. blue/object under -tags static would still pick up the
// non-static fyne/raylib imports and fail when cgo is disabled.
func runningBuildTags() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, setting := range info.Settings {
		if setting.Key == "-tags" {
			return setting.Value
		}
	}
	return ""
}

func deps(t *testing.T, pkg string) []string {
	t.Helper()
	// Use the module-qualified import path so the command works from the
	// test's own directory anywhere inside the module.
	importPath := "blue/" + strings.TrimPrefix(pkg, "./")
	args := []string{"list", "-deps"}
	if tags := runningBuildTags(); tags != "" {
		args = append(args, "-tags="+tags)
	}
	args = append(args, importPath)
	out, err := exec.Command("go", args...).Output()
	if err != nil {
		t.Fatalf("go %s failed: %v\n%s", strings.Join(args, " "), err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	deps := make([]string, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "blue/") && l != "blue" {
			deps = append(deps, l)
		}
	}
	return deps
}

func assertNotImporting(t *testing.T, pkg string, deps []string, forbidden []string) {
	t.Helper()
	for _, d := range deps {
		for _, f := range forbidden {
			if d == f || strings.HasPrefix(d, f+"/") {
				t.Errorf("%s imports %s, which is forbidden for the minimal VM build (see tools/importguard)", pkg, d)
			}
		}
	}
}

func TestMinimalVMBuildHasNoToolchainImports(t *testing.T) {
	for _, pkg := range strictPackages {
		t.Run(pkg, func(t *testing.T) {
			assertNotImporting(t, pkg, deps(t, pkg), forbiddenForAll)
			assertNotImporting(t, pkg, deps(t, pkg), forbiddenExtraStrict)
		})
	}
	for _, pkg := range runnerPackages {
		t.Run(pkg, func(t *testing.T) {
			assertNotImporting(t, pkg, deps(t, pkg), forbiddenForAll)
		})
	}
}
