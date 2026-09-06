package cmd

import (
	"blue/cmd/srcbundle"
	"os"
	"path/filepath"
	"testing"
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

func TestSrcUpToDateTracksArchive(t *testing.T) {
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
		t.Fatal("missing archive record should not be up to date")
	}
	if err := srcbundle.WriteArchive(installedArchivePath(root)); err != nil {
		t.Fatal(err)
	}
	if !srcUpToDate(root, srcDir) {
		t.Fatal("recorded archive should be up to date")
	}
	if err := os.WriteFile(installedArchivePath(root), []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	if srcUpToDate(root, srcDir) {
		t.Fatal("changed archive record should not be up to date")
	}
}

func TestRunInstallTwiceSkipsSecondExtract(t *testing.T) {
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
	if storedArchiveHash(prefix) != srcbundle.ArchiveSHA256() {
		t.Fatal("stored archive hash does not match embedded archive")
	}
}
