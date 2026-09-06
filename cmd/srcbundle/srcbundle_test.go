package srcbundle

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func testTree() fstest.MapFS {
	return fstest.MapFS{
		"go.mod":                  {Data: []byte("module blue\n")},
		"go.sum":                  {Data: []byte("sums\n")},
		"main.go":                 {Data: []byte("package main\n")},
		"cmd/bluerun/main.go":     {Data: []byte("//go:build minivm\n")},
		"lib/core/core.b":         {Data: []byte("// core\n")},
		"vm/vm.go":                {Data: []byte("package vm\n")},
		"vm/vm_test.go":           {Data: []byte("package vm\n")},
		"bluec/testdata/fuzz/abc": {Data: []byte("seed\n")},
		".gitignore":              {Data: []byte("ignored\n")},
	}
}

func TestNotAvailableWithoutTree(t *testing.T) {
	SetFS(nil)
	if Available() {
		t.Fatal("available without a source tree")
	}
	if got := TreeSHA256(); got != "" {
		t.Fatalf("expected empty hash, got %q", got)
	}
	if _, err := List(); err == nil {
		t.Fatal("expected error listing an absent tree")
	}
	if err := Extract(t.TempDir()); err == nil {
		t.Fatal("expected error extracting an absent tree")
	}
}

func TestListSkipsTestScratch(t *testing.T) {
	SetFS(testTree())
	names, err := List()
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	for _, name := range names {
		if filepath.IsAbs(name) {
			t.Errorf("list contains absolute path %q", name)
		}
		found[name] = true
	}
	for _, want := range []string{"go.mod", "go.sum", "main.go", "cmd/bluerun/main.go", "lib/core/core.b", "vm/vm.go"} {
		if !found[want] {
			t.Errorf("list missing %q", want)
		}
	}
	for _, unwanted := range []string{"vm/vm_test.go", "bluec/testdata/fuzz/abc"} {
		if found[unwanted] {
			t.Errorf("list should skip test scratch %q", unwanted)
		}
	}
}

func TestExtractWritesSourceTree(t *testing.T) {
	SetFS(testTree())
	dest := t.TempDir()
	if err := Extract(dest); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"go.mod", filepath.Join("cmd", "bluerun", "main.go"), filepath.Join("lib", "core", "core.b")} {
		data, err := os.ReadFile(filepath.Join(dest, want))
		if err != nil || len(data) == 0 {
			t.Errorf("extracted %s missing or empty: %v", want, err)
		}
	}
	for _, unwanted := range []string{filepath.Join("vm", "vm_test.go"), filepath.Join("bluec", "testdata")} {
		if _, err := os.Stat(filepath.Join(dest, unwanted)); err == nil {
			t.Errorf("extract wrote test scratch %s", unwanted)
		}
	}
	if _, err := os.Stat(filepath.Join(dest, ".gitignore")); err != nil {
		t.Errorf(".gitignore in tree should be extracted when present: %v", err)
	}
}

func TestTreeSHA256IsStableAndContentSensitive(t *testing.T) {
	SetFS(testTree())
	first := TreeSHA256()
	if len(first) != 64 {
		t.Fatalf("expected 64 hex chars, got %q", first)
	}
	if second := TreeSHA256(); second != first {
		t.Fatal("hash is not stable across calls")
	}
	SetFS(fstest.MapFS{"go.mod": {Data: []byte("module other\n")}})
	if changed := TreeSHA256(); changed == first {
		t.Fatal("hash did not change with the tree contents")
	}
}

func TestIsTestScratch(t *testing.T) {
	cases := map[string]bool{
		"vm_test.go":                     true,
		"testdata":                       true,
		"parser_illegal_tok_test_blow.b": true,
		"vm.go":                          false,
		"core.b":                         false,
		"latest.b":                       false,
	}
	for name, want := range cases {
		if got := isTestScratch(name); got != want {
			t.Errorf("isTestScratch(%q) = %v, want %v", name, got, want)
		}
	}
}
