package evaluator

import "testing"

func TestAllBuiltinsHaveHelpString(t *testing.T) {
	for k, v := range builtins.kv {
		if v.HelpStr == "" {
			t.Fatalf("`%s` does not have help string", k)
		}
	}
}
