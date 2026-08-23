package importguard

import (
	"os/exec"
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

func deps(t *testing.T, pkg string) []string {
	t.Helper()
	// Use the module-qualified import path so the command works from the
	// test's own directory anywhere inside the module.
	importPath := "blue/" + strings.TrimPrefix(pkg, "./")
	out, err := exec.Command("go", "list", "-deps", importPath).Output()
	if err != nil {
		t.Fatalf("go list -deps %s failed: %v\n%s", importPath, err, out)
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
