package cmd

import (
	"blue/cmd/srcbundle"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestDefaultDirs(t *testing.T) {
	if got := DefaultInstallRoot(); got == "" {
		t.Fatal("DefaultInstallRoot is empty")
	}
	if got := DefaultSrcDir(); got == "" {
		t.Fatal("DefaultSrcDir is empty")
	}
	if got := DefaultBinDir(); got == "" {
		t.Fatal("DefaultBinDir is empty")
	}
	found := false
	for _, c := range defaultSourceCandidates() {
		if c == DefaultSrcDir() {
			found = true
		}
	}
	if !found {
		t.Fatal("default candidates do not include DefaultSrcDir")
	}
}

func TestRepoRootIsSourceDir(t *testing.T) {
	ok, err := isBlueSourceDir("..")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("repo root not recognized as source dir")
	}
}

func TestFindBlueSourceDirUsesEnv(t *testing.T) {
	t.Setenv("BLUE_INSTALL_PATH", "..")
	dir, ok := findBlueSourceDir()
	if !ok {
		t.Fatal("findBlueSourceDir missed BLUE_INSTALL_PATH")
	}
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err != nil {
		t.Fatalf("candidate %s has no go.mod: %v", dir, err)
	}
}

func TestRunInstallBinOnly(t *testing.T) {
	prefix := t.TempDir()
	if err := RunInstall(InstallOptions{Prefix: prefix, NoSrc: true}); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(prefix, "bin"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("bin dir empty after install")
	}
}

func TestRunInstallNoop(t *testing.T) {
	prefix := t.TempDir()
	if err := RunInstall(InstallOptions{Prefix: prefix, NoSrc: true, NoBin: true}); err != nil {
		t.Fatal(err)
	}
}

func TestInstallExecutableKeepsBaseName(t *testing.T) {
	dest, err := installExecutable(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(dest) != filepath.Base(exe) {
		t.Fatalf("installed as %s, want %s", filepath.Base(dest), filepath.Base(exe))
	}
}

// useTestTree gives srcbundle an in-memory source tree to install from, as the
// module root's main package does for the real binary.
func useTestTree(t *testing.T) {
	t.Helper()
	srcbundle.SetFS(fstest.MapFS{
		"go.mod":              {Data: []byte("module blue\n\ngo 1.25.0\n")},
		"main.go":             {Data: []byte("package main\n")},
		"cmd/bluerun/main.go": {Data: []byte("//go:build minivm\n")},
		"lib/core/core.b":     {Data: []byte("// core\n")},
		"vm/vm.go":            {Data: []byte("package vm\n")},
		"vm/vm_test.go":       {Data: []byte("package vm\n")},
	})
}

func TestSrcUpToDateTracksTreeHash(t *testing.T) {
	useTestTree(t)
	root := t.TempDir()
	srcDir := filepath.Join(root, "src")
	if srcUpToDate(root, srcDir) {
		t.Fatal("empty root should not be up to date")
	}
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := srcbundle.Extract(srcDir); err != nil {
		t.Fatal(err)
	}
	if srcUpToDate(root, srcDir) {
		t.Fatal("missing tree record should not be up to date")
	}
	if err := writeTreeHash(root); err != nil {
		t.Fatal(err)
	}
	if !srcUpToDate(root, srcDir) {
		t.Fatal("recorded tree should be up to date")
	}
	if err := os.WriteFile(treeHashPath(root), []byte("stale\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if srcUpToDate(root, srcDir) {
		t.Fatal("changed tree record should not be up to date")
	}
}

func TestRunInstallTwiceSkipsSecondExtract(t *testing.T) {
	useTestTree(t)
	prefix := t.TempDir()
	if err := RunInstall(InstallOptions{Prefix: prefix}); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(prefix, "src", "go.mod")
	info, err := os.Stat(marker)
	if err != nil {
		t.Fatal(err)
	}
	if err := RunInstall(InstallOptions{Prefix: prefix}); err != nil {
		t.Fatal(err)
	}
	again, err := os.Stat(marker)
	if err != nil {
		t.Fatal(err)
	}
	if !again.ModTime().Equal(info.ModTime()) {
		t.Fatal("second install rewrote an up to date source tree")
	}
	if got := storedTreeHash(prefix); got != srcbundle.TreeSHA256() {
		t.Fatalf("stored tree hash %q does not match the source tree %q", got, srcbundle.TreeSHA256())
	}
	if _, err := os.Stat(filepath.Join(prefix, "src", "vm", "vm_test.go")); err == nil {
		t.Fatal("installed tree should not contain test files")
	}
}
