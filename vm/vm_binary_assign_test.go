package vm

import (
	"strings"
	"testing"

	"blue/compiler"
)

func runInspect(t *testing.T, input string) string {
	t.Helper()
	program := parse(input)
	comp := compiler.NewFromCore()
	if err := comp.Compile(program); err != nil {
		t.Fatalf("compiler error for %q: %s", input, err)
	}
	vm := New(comp.Bytecode())
	if err := vm.Run(); err != nil {
		t.Fatalf("vm error for %q: %s", input, err)
	}
	return vm.LastPoppedStackElem().Inspect()
}

func runExpectVmError(t *testing.T, input, wantSubstring string) {
	t.Helper()
	program := parse(input)
	comp := compiler.NewFromCore()
	if err := comp.Compile(program); err != nil {
		t.Fatalf("compiler error for %q: %s", input, err)
	}
	vm := New(comp.Bytecode())
	err := vm.Run()
	if err == nil {
		t.Fatalf("expected VM error for %q, got none (result=%s)", input, vm.LastPoppedStackElem().Inspect())
	}
	if !strings.Contains(err.Error(), wantSubstring) {
		t.Fatalf("error %q does not contain %q", err.Error(), wantSubstring)
	}
}

func TestBinarySetOperations(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`{1, 2} + 3`, "{1, 2, 3}"},
		{`3 + {1, 2}`, "{1, 2, 3}"},
		{`val s = {1, 2}; val t = s + 3; s`, "{1, 2}"},
		{`val s = {1, 2}; val t = s + 3; t`, "{1, 2, 3}"},
		{`val s = {1, 2}; val u = s + "dup"; val v = s + "dup"; v`, "{1, 2, dup}"},
		{`{1, 2} | {3}`, "{1, 2, 3}"},
		{`{3} | {1, 2}`, "{1, 2, 3}"},
		{`{1, 2, 3} & {2, 3, 4}`, "{2, 3}"},
		{`{1, 2} & {3, 4}`, "{}"},
		{`{1, 2} ^ {2, 3}`, "{1, 3}"},
		{`{1, 2, 3} >= {2, 3}`, "true"},
		{`{1, 2, 3} >= {9}`, "false"},
		{`{1, 2, 3} - {2}`, "{1, 3}"},
		{`{1, 2} == {2, 1}`, "true"},
		{`{1, 2} == {1}`, "false"},
		{`{1, 2} != {1}`, "true"},
		{`{1, 2} != {2, 1}`, "false"},
	}
	for _, tt := range tests {
		if got := runInspect(t, tt.input); got != tt.want {
			t.Errorf("%s = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestBinarySetUnknownOperatorErrors(t *testing.T) {
	runExpectVmError(t, `{1, 2} % {3}`, "unknown operator")
}

func TestBinaryBytesOperations(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`"abc".to_bytes() & "abd".to_bytes()`, "[]byte{0x61, 0x62, 0x60}"},
		{`"abc".to_bytes() | "abd".to_bytes()`, "[]byte{0x61, 0x62, 0x67}"},
		{`"abc".to_bytes() ^ "abd".to_bytes()`, "[]byte{0x0, 0x0, 0x7}"},
		{`"ab".to_bytes() == "ab".to_bytes()`, "true"},
		{`"ab".to_bytes() == "ba".to_bytes()`, "false"},
		{`"ab".to_bytes() != "ba".to_bytes()`, "true"},
	}
	for _, tt := range tests {
		if got := runInspect(t, tt.input); got != tt.want {
			t.Errorf("%s = %q, want %q", tt.input, got, tt.want)
		}
	}

	for _, input := range []string{
		`"a".to_bytes() & "ab".to_bytes()`,
		`"a".to_bytes() | "ab".to_bytes()`,
		`"a".to_bytes() ^ "ab".to_bytes()`,
	} {
		runExpectVmError(t, input, "length of left and right bytes must match")
	}
}

func TestBinaryBigFloatOperations(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`bigfloat("0.1") + bigfloat("0.2")`, "0.3"},
		{`bigfloat(2) - bigfloat("1.5")`, "0.5"},
		{`bigfloat("1.5") * bigfloat(4)`, "6"},
		{`bigfloat("1.5") / bigfloat(2)`, "0.75"},
		{`bigfloat("1.5") ** 2`, "2.25"},
		{`bigfloat("7.5") // 2`, "3"},
		{`bigfloat("-7.5") // 2`, "-4"},
		{`bigfloat("7.5") % 2`, "1.5"},
		{`bigfloat("-7.5") % 2`, "0.5"},
		{`bigfloat("-7.5") % -2`, "-1.5"},
		{`bigfloat("1.5") > 1`, "true"},
		{`bigfloat("0.5") >= bigfloat("0.5")`, "true"},
		{`bigfloat("1.5") == bigfloat("1.5")`, "true"},
		{`bigfloat("1.5") != 2`, "true"},
		{`bigint("99999999999999999999") + bigfloat("0.5")`, "99999999999999999999.5"},
	}
	for _, tt := range tests {
		if got := runInspect(t, tt.input); got != tt.want {
			t.Errorf("%s = %q, want %q", tt.input, got, tt.want)
		}
	}

	runExpectVmError(t, `bigfloat(1) // 0`, "Floor Division by zero is not allowed")
	runExpectVmError(t, `bigfloat(1) % 0`, "Modulus by zero is not allowed")
}

func TestBinaryNullOperations(t *testing.T) {
	tests := []vmTestCase{
		{"null == null", true},
		{"null != null", false},
	}
	runVmTests(t, tests)
}

func TestListRepeatAndInNotIn(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`[1, 2] * 2`, "[1, 2, 1, 2]"},
		{`[1] * 0`, "[]"},
		{`"ab" in ["ab", "cd"]`, "true"},
		{`"zz" in ["ab"]`, "false"},
		{`1 notin [2, 3]`, "true"},
		{`2 notin [1, 2]`, "false"},
		{`1 in {1, 2}`, "true"},
		{`9 in {1, 2}`, "false"},
		{`1 notin {2}`, "true"},
		{`"k" in {"k": 1}`, "true"},
		{`"z" in {"k": 1}`, "false"},
		{`"z" notin {"k": 1}`, "true"},
	}
	for _, tt := range tests {
		if got := runInspect(t, tt.input); got != tt.want {
			t.Errorf("%s = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestShiftIntoContainers(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`var l = [2, 3]; 1 >> l; l`, "[1, 2, 3]"},
		{`var l = [2]; l << 3; l`, "[2, 3]"},
		{`var s = {1}; 2 >> s; s`, "{1, 2}"},
		{`var s = {1}; s << 1; s`, "{1}"},
	}
	for _, tt := range tests {
		if got := runInspect(t, tt.input); got != tt.want {
			t.Errorf("%s = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCompoundIndexAssignments(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`var m = {"a": 1}; m["a"] += 5; m["a"]`, "6"},
		{`var l = [10, 20]; l[1] *= 3; l[1]`, "60"},
		{`var l = [1, 2]; l[0] -= 5; l`, "[-4, 2]"},
		{`var st = @{name: "b"}; st.name += "!"; st.name`, "b!"},
		{`var s = "abc"; s[0] = "X"; s`, "Xbc"},
		{`var s = "abc"; s[1] = "Z"; s`, "aZc"},
	}
	for _, tt := range tests {
		if got := runInspect(t, tt.input); got != tt.want {
			t.Errorf("%s = %q, want %q", tt.input, got, tt.want)
		}
	}

	runExpectVmError(t, `var s = "abc"; s[0] = "XY"; s`, "must be 1 character long")
	runExpectVmError(t, `var s = "abc"; s[0] = 5; s`, "cannot assign INTEGER to STRING")
}

func TestIndexSetErrors(t *testing.T) {
	runExpectVmError(t, `var l = [1]; l[5] = 2`, "index out of bounds: 5")
	runExpectVmError(t, `var n = 5; n[0] = 1`, "is not indexable")
}
