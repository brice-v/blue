package evaluator

import "testing"

func TestAllBuiltinsHaveHelpString(t *testing.T) {
	for k, v := range builtins.Kv {
		if v.HelpStr == "" {
			t.Fatalf("builtin `%s` does not have help string", k)
		}
	}
}

func TestAllStringBuiltinsHaveHelpString(t *testing.T) {
	for k, v := range stringbuiltins.Kv {
		if v.HelpStr == "" {
			t.Fatalf("string builtin `%s` does not have help string", k)
		}
	}
}

func TestAllStdFunctionsHaveHelpString(t *testing.T) {
	for k, v := range _std_mods {
		// k is the module name
		for kk, vv := range v.Builtins.Kv {
			// kk is the builtin name
			// vv is builtin function object
			if vv.HelpStr == "" {
				t.Fatalf("std mod `%s` function `%s` does not have help string", k, kk)
			}
		}
	}
}
