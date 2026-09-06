package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempProgram(t *testing.T, body string) string {
	t.Helper()
	fpath := filepath.Join(t.TempDir(), "prog.b")
	if err := os.WriteFile(fpath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return fpath
}

func TestParseBundleArgsDefaultsToGoBuild(t *testing.T) {
	fpath := writeTempProgram(t, "println(1)\n")
	opts, err := parseBundleArgs(5, []string{"bundle", "-o", "out", fpath, "--all-parser-errors"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.usePrebuilt {
		t.Fatal("default should build the template, not use a prebuilt one")
	}
	if !opts.allErrors || opts.outPath != "out" || opts.fpath != fpath {
		t.Fatalf("unexpected parsed options: %+v", opts)
	}
}

func TestParseBundleArgsBluerunFlag(t *testing.T) {
	fpath := writeTempProgram(t, "println(1)\n")
	opts, err := parseBundleArgs(4, []string{"bundle", "--bluerun", "-o", "out"})
	if err == nil {
		t.Fatalf("expected error for missing source file, got %+v", opts)
	}
	opts, err = parseBundleArgs(5, []string{"bundle", "--bluerun", "-o", "out", fpath})
	if err != nil {
		t.Fatal(err)
	}
	if !opts.usePrebuilt {
		t.Fatal("--bluerun should select the prebuilt template")
	}
}

func TestParseBundleArgsGoBuildAlias(t *testing.T) {
	fpath := writeTempProgram(t, "println(1)\n")
	opts, err := parseBundleArgs(5, []string{"bundle", "--go-build", "-o", "out", fpath})
	if err != nil {
		t.Fatal(err)
	}
	if opts.usePrebuilt {
		t.Fatal("--go-build alias should keep the default build behavior")
	}
}

func TestParseBundleArgsConflictingFlags(t *testing.T) {
	fpath := writeTempProgram(t, "println(1)\n")
	if _, err := parseBundleArgs(6, []string{"bundle", "--bluerun", "--go-build", "-o", "out", fpath}); err == nil {
		t.Fatal("expected error when --bluerun and --go-build are combined")
	}
}

func TestParseBundleArgsRequiresOutput(t *testing.T) {
	fpath := writeTempProgram(t, "println(1)\n")
	if _, err := parseBundleArgs(2, []string{"bundle", fpath}); err == nil {
		t.Fatal("expected error when -o is missing")
	}
}
