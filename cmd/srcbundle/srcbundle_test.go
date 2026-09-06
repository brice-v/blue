package srcbundle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArchiveAvailable(t *testing.T) {
	if !Available() {
		t.Fatal("embedded source archive missing: run ./make_src_bundle.sh first")
	}
}

func TestArchiveSHA256Stable(t *testing.T) {
	first := ArchiveSHA256()
	if len(first) != 64 {
		t.Fatalf("expected 64 hex chars, got %q", first)
	}
	if second := ArchiveSHA256(); second != first {
		t.Fatal("archive hash is not stable across calls")
	}
}

func TestArchiveExcludesTests(t *testing.T) {
	names, err := List()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) == 0 {
		t.Fatal("archive is empty")
	}
	for _, n := range names {
		if strings.HasSuffix(n, "_test.go") {
			t.Errorf("archive embeds test file %s", n)
		}
		if strings.HasPrefix(n, "b_test_programs/") || strings.HasPrefix(n, "manual_tests/") {
			t.Errorf("archive embeds test directory file %s", n)
		}
		if strings.HasPrefix(n, "vendor/") || strings.HasPrefix(n, "ignored/") || strings.HasPrefix(n, "playground/") {
			t.Errorf("archive embeds excluded dir file %s", n)
		}
		if strings.HasPrefix(n, "man/") || strings.HasPrefix(n, ".github/") {
			t.Errorf("archive embeds excluded dir file %s", n)
		}
	}
	found := map[string]bool{}
	for _, n := range names {
		found[n] = true
	}
	for _, want := range []string{"go.mod", "go.sum", "cmd/bluerun/main.go", "lib/core/core.b"} {
		if !found[want] {
			t.Errorf("archive missing required file %s", want)
		}
	}
	for _, excluded := range []string{"README.md", "blue-TODO.txt", "b.txt", "hf.md", "scratchfile.b", "gen-man.sh", "make_release", "make_src_bundle.sh", "benchmark-things.sh", "parser/parser_illegal_tok_test_killed_by_oom_in_parser.b"} {
		if found[excluded] {
			t.Errorf("archive embeds excluded file %s", excluded)
		}
	}
}

func TestExtractWritesSourceTree(t *testing.T) {
	dest := t.TempDir()
	if err := Extract(dest); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"go.mod", filepath.Join("cmd", "bluerun", "main.go")} {
		data, err := os.ReadFile(filepath.Join(dest, want))
		if err != nil || len(data) == 0 {
			t.Errorf("extracted %s missing or empty: %v", want, err)
		}
	}
}
