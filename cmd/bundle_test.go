package cmd

import (
	"blue/consts"
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

func TestFormatVersion(t *testing.T) {
	if got := formatVersion(false); got != "blue v"+consts.ShortVersion() {
		t.Fatalf("short version mismatch: %s", got)
	}
	if got := formatVersion(true); got != "blue v"+consts.FullVersion() {
		t.Fatalf("full version mismatch: %s", got)
	}
	if consts.ShortVersion() == "" || consts.FullVersion() == "" {
		t.Fatal("versions must not be empty")
	}
}

func TestRunnerTempPathIsAbsolute(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	got := runnerTempPath("main-test")
	want := filepath.Join(dir, "main-test.bluerun-tmp")
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
	abs, err := filepath.Abs(filepath.Join(dir, "sub", "app"))
	if err != nil {
		t.Fatal(err)
	}
	if got := runnerTempPath(abs); got != abs+".bluerun-tmp" {
		t.Fatalf("absolute output changed directory: %s", got)
	}
}
