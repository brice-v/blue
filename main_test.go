package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"blue/cmd/srcbundle"
)

func TestEmbeddedSourceTreeHasWhatInstallingNeeds(t *testing.T) {
	srcbundle.SetFS(srcTree)
	if !srcbundle.Available() {
		t.Fatal("no embedded source tree")
	}
	names, err := srcbundle.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) == 0 {
		t.Fatal("embedded source tree is empty")
	}
	found := map[string]bool{}
	for _, name := range names {
		found[name] = true
	}
	for _, want := range []string{"go.mod", "go.sum", "main.go", filepath.Join("cmd", "bluerun", "main.go"), filepath.Join("lib", "core", "core.b")} {
		if !found[filepath.ToSlash(want)] {
			t.Errorf("embedded tree missing %s", want)
		}
	}
}

func TestEmbeddedSourceTreeLeavesOutJunk(t *testing.T) {
	srcbundle.SetFS(srcTree)
	names, err := srcbundle.List()
	if err != nil {
		t.Fatal(err)
	}
	excludedPrefixes := []string{"vendor/", "ignored/", "playground/", "man/", "b_test_programs/", "manual_tests/", ".github/", "tools/"}
	for _, name := range names {
		for _, prefix := range excludedPrefixes {
			if strings.HasPrefix(name, prefix) {
				t.Errorf("embedded tree should not contain %s", prefix)
			}
		}
		if strings.Contains(name, "_test") || strings.HasSuffix(name, ".out") {
			t.Errorf("embedded tree should not contain test scratch %q", name)
		}
	}
}

func TestEmbeddedSourceTreeExtractsToBuildableLayout(t *testing.T) {
	srcbundle.SetFS(srcTree)
	dest := t.TempDir()
	if err := srcbundle.Extract(dest); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"go.mod", "go.sum", filepath.Join("cmd", "bluerun", "main.go")} {
		info, err := os.Stat(filepath.Join(dest, want))
		if err != nil {
			t.Errorf("extracted tree missing %s: %v", want, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("extracted %s is empty", want)
		}
	}
	if len(srcbundle.TreeSHA256()) != 64 {
		t.Fatalf("expected a tree hash, got %q", srcbundle.TreeSHA256())
	}
}
