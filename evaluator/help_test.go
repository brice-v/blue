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

func TestAllStdFunctionsHaveHelpString(t *testing.T) {
	for k, v := range _std_mods {
		// k is the module name
		if k != "gg" { // TODO: Remove this and finish for all std modules
			for kk, vv := range v.Builtins.kv {
				// kk is the builtin name
				// vv is builtin function object
				if vv.HelpStr == "" {
					t.Fatalf("std mod `%s` function `%s` does not have help string", k, kk)
				}
			}
		}
	}
}
