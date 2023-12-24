package evaluator

import "testing"

func TestAllBuiltinsHaveHelpString(t *testing.T) {
	for k, v := range builtins.kv {
		if v.HelpStr == "" {
			t.Fatalf("builtin `%s` does not have help string", k)
		}
	}
}

func TestAllStringBuiltinsHaveHelpString(t *testing.T) {
	for k, v := range stringbuiltins.kv {
		if v.HelpStr == "" {
			t.Fatalf("string builtin `%s` does not have help string", k)
		}
	}
}
