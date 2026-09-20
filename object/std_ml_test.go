package object

import (
	"strings"
	"testing"
)

// TestMlRegistry checks the ml builtin table is well formed: every entry has a
// name, a help string, and a callable, and no name is registered twice. A
// duplicate name would shadow a builtin at import time, and a missing help
// string leaves `help(...)` empty for that builtin.
func TestMlRegistry(t *testing.T) {
	seen := make(map[string]bool)
	for _, b := range MlBuiltins {
		if b.Name == "" {
			t.Fatalf("ml builtin with empty name")
		}
		if b.HelpStr == "" {
			t.Errorf("ml builtin %q has no help string", b.Name)
		}
		if seen[b.Name] {
			t.Errorf("duplicate ml builtin name %q", b.Name)
		}
		seen[b.Name] = true
		if b.Fun == nil {
			t.Errorf("ml builtin %q has nil Fun", b.Name)
		}
	}
}

// TestMlHelpFormat checks every ml help string carries the four labelled
// sections the help printer renders. A malformed help string (for example one
// built from a bare string instead of helpStrArgs) would drop a section and
// make `help(...)` output inconsistent with every other builtin.
func TestMlHelpFormat(t *testing.T) {
	for _, b := range MlBuiltins {
		for _, want := range []string{"Signature:", "Error(s):", "Example(s):"} {
			if !strings.Contains(b.HelpStr, want) {
				t.Errorf("ml builtin %q help missing %q:\n%s", b.Name, want, b.HelpStr)
			}
		}
	}
}
