package cmd

import (
	"blue/compiler"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCutModuleMember(t *testing.T) {
	cases := []struct {
		name   string
		mod    string
		member string
		wantOk bool
	}{
		{"math.rand", "math", "rand", true},
		{"./sub/mod.b.pi", "./sub/mod.b", "pi", true},
		{"nope", "", "", false},
		{".x", "", "", false},
		{"x.", "", "", false},
	}
	for _, tc := range cases {
		mod, member, ok := cutModuleMember(tc.name)
		if ok != tc.wantOk || mod != tc.mod || member != tc.member {
			t.Errorf("cutModuleMember(%q) = (%q, %q, %v), want (%q, %q, %v)", tc.name, mod, member, ok, tc.mod, tc.member, tc.wantOk)
		}
	}
}

func TestDocStringForStdModuleMember(t *testing.T) {
	docStr := getDocStringFor("math.rand")
	if strings.TrimSpace(docStr) == "" {
		t.Fatal("no documentation produced for `math.rand`")
	}
	if !strings.Contains(docStr, "`rand` returns a random float between 0 and 1") {
		t.Errorf("documentation = %q, want the rand function's own docs", docStr)
	}
}

func TestDocStringForStdModuleMemberBuiltinAlias(t *testing.T) {
	docStr := getDocStringFor("math.acos")
	if !strings.Contains(docStr, "arccosine") {
		t.Errorf("documentation = %q, want the wrapped builtin's help", docStr)
	}
}

func TestDocStringForStdModuleMemberWrapped(t *testing.T) {
	docStr := getDocStringFor("color.style")
	if !strings.Contains(docStr, "takes a text style") || !strings.Contains(docStr, "builtin__style") {
		t.Errorf("documentation = %q, want own docs plus the wrapped builtin's", docStr)
	}
}

func TestDocStringForMemberOfLocalFile(t *testing.T) {
	dir := t.TempDir()
	fpath := filepath.Join(dir, "util.b")
	source := "## utility helpers for the demo\nfun greeting(name) {\n    ## `greeting` returns a greeting for name\n    \"hello \" + name\n}\n"
	if err := os.WriteFile(fpath, []byte(source), 0o644); err != nil {
		t.Fatalf("failed to write scratch module: %v", err)
	}

	docStr := getDocStringFor(fpath + ".greeting")
	if !strings.Contains(docStr, "`greeting` returns a greeting for name") {
		t.Errorf("documentation = %q, want the local function's docs", docStr)
	}
}

func TestDocStringForLocalModuleWithoutExtension(t *testing.T) {
	dir := t.TempDir()
	fpath := filepath.Join(dir, "mod", "util.b")
	if err := os.MkdirAll(filepath.Dir(fpath), 0o755); err != nil {
		t.Fatalf("failed to make scratch dir: %v", err)
	}
	source := "## utility helpers for the demo\nfun greeting(name) {\n    ## `greeting` returns a greeting for name\n    \"hello \" + name\n}\n"
	if err := os.WriteFile(fpath, []byte(source), 0o644); err != nil {
		t.Fatalf("failed to write scratch module: %v", err)
	}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working dir: %v", err)
	}
	if err := os.Chdir(filepath.Dir(fpath)); err != nil {
		t.Fatalf("failed to change to scratch dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	docStr := getDocStringFor("util")
	if !strings.Contains(docStr, "MODULE `util`") || !strings.Contains(docStr, "greeting") {
		t.Errorf("`blue doc util` output = %q, want the whole module with its function listed", docStr)
	}

	docMember := getDocStringFor("util.greeting")
	if !strings.Contains(docMember, "`greeting` returns a greeting for name") {
		t.Errorf("`blue doc util.greeting` output = %q, want the function's docs", docMember)
	}
}

func TestDocStringForMemberNotFound(t *testing.T) {
	for _, name := range []string{"math.PI", "nope.nope", "1.5"} {
		if docStr := getDocStringFor(name); strings.TrimSpace(docStr) != "" {
			t.Errorf("getDocStringFor(%q) = %q, want empty", name, docStr)
		}
	}
}

func TestDocStringForStdModuleMatchesHelpOutput(t *testing.T) {
	docStr := getDocStringFor("math")
	if !strings.HasPrefix(docStr, "MODULE `math`:") {
		t.Errorf("`blue doc math` output = %q, want the module's help header", docStr)
	}
	if !strings.Contains(docStr, "PUBLIC FUNCTIONS:") {
		t.Errorf("`blue doc math` output = %q, want a public functions section", docStr)
	}
	if !strings.Contains(docStr, "rand | `rand` returns a random float between 0 and 1") {
		t.Errorf("`blue doc math` output = %q, want rand's docs in the help layout", docStr)
	}
	if !strings.Contains(docStr, "\n     |  rand() -> float") {
		t.Errorf("`blue doc math` output = %q, want continuation lines to keep the | gutter", docStr)
	}
}

func TestDocStringForStdListsAllStdModules(t *testing.T) {
	docStr := getDocStringFor("std")
	if !strings.HasPrefix(docStr, "MODULE `") {
		t.Fatalf("`blue doc std` returned %q, want the std module index rather than the ml `std` builtin", strings.Split(docStr, "\n")[0])
	}
	for _, mod := range compiler.StdModuleNames() {
		if !strings.Contains(docStr, "MODULE `"+mod+"`:") {
			t.Errorf("`blue doc std` output is missing module %q", mod)
		}
	}
}

func TestHandleDocCommandUnknownNameFails(t *testing.T) {
	if err := handleDocCommand(2, []string{"doc", "no.such_thing"}); err == nil {
		t.Error("handleDocCommand succeeded on an unknown name, want an error")
	}
	if err := handleDocCommand(1, []string{"doc"}); err == nil {
		t.Error("handleDocCommand succeeded without its argument, want an error")
	}
}
