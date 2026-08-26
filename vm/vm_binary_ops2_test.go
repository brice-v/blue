package vm

import (
	"os"
	"testing"
)

func TestIntegerBitwiseAndArithmetic(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`7 & 3`, "3"},
		{`7 | 8`, "15"},
		{`7 ^ 3`, "4"},
		{`1 << 4`, "16"},
		{`16 >> 2`, "4"},
		{`2 ** 10`, "1024"},
		{`var a = 9223372036854775807; var b = a + 1; b`, "-9223372036854775808"},
		{`-9223372036854775808 - 1`, "-9223372036854775809"},
		{`3037000500 * 2`, "6074001000"},
		{`7 // 2`, "3"},
		{`-7 // 2`, "-4"},
		{`7 % 3`, "1"},
		{`-7 % 3`, "2"},
		{`7 % -3`, "-2"},
		{`5 / 7`, "0"},
		{`1..4`, "[1, 2, 3, 4]"},
		{`4..1`, "[4, 3, 2, 1]"},
		{`1..<4`, "[1, 2, 3]"},
		{`2..2`, "[2]"},
		{`"a".."e"`, "[a, b, c, d, e]"},
		{`"a"..<"e"`, "[a, b, c, d]"},
		{`"e".."a"`, "[e, d, c, b, a]"},
		{`"c".."c"`, "[c]"},
		{`"c"..<"c"`, "[]"},
		{`5 > 3`, "true"},
		{`5 >= 5`, "true"},
		{`5 == 5`, "true"},
		{`5 != 5`, "false"},
	}
	for _, tt := range tests {
		if got := runInspect(t, tt.input); got != tt.want {
			t.Errorf("%s = %q, want %q", tt.input, got, tt.want)
		}
	}

	runExpectVmError(t, `1 / 0`, "division by zero is not allowed")
	runExpectVmError(t, `1 // 0`, "floor division by zero is not allowed")
	runExpectVmError(t, `1 % 0`, "Modulus by zero is not allowed")
}

func TestStringBinaryOperations(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`"foo" + "bar"`, "foobar"},
		{`"abc" == "abc"`, "true"},
		{`"abc" != "abc"`, "false"},
		{`"b" > "a"`, "true"},
		{`"b" >= "b"`, "true"},
		{`"ell" in "hello"`, "true"},
		{`"x" in "hello"`, "false"},
		{`"x" notin "hello"`, "true"},
		{`"ab" * 3`, "ababab"},
		{`3 * "ab"`, "ababab"},
	}
	for _, tt := range tests {
		if got := runInspect(t, tt.input); got != tt.want {
			t.Errorf("%s = %q, want %q", tt.input, got, tt.want)
		}
	}

	runExpectVmError(t, `"a".."abc"`, "operator .. expects right string to be 1 rune")
	runExpectVmError(t, `"abc".."a"`, "operator .. expects left string to be 1 rune")
}

func TestBigIntegerMixedOperations(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`bigint("99999999999999999999") + 1`, "100000000000000000000"},
		{`bigint("99999999999999999999") - 99999999999999999999`, "0"},
		{`bigint("1000000000") * bigint("1000000000")`, "1000000000000000000"},
		{`bigint("10000000000000000000") // 3`, "3333333333333333333"},
		{`bigint("10000000000000000000") % 7`, "3"},
		{`bigint("2") ** 64`, "18446744073709551616"},
		{`bigint("5") > 3`, "true"},
		{`bigint("5") == 5`, "true"},
	}
	for _, tt := range tests {
		if got := runInspect(t, tt.input); got != tt.want {
			t.Errorf("%s = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMapBinaryOperations(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`{"a": 1} == {"a": 1}`, "true"},
		{`{"a": 1} == {"a": 2}`, "false"},
		{`{"a": 1} != {"b": 1}`, "true"},
	}
	for _, tt := range tests {
		if got := runInspect(t, tt.input); got != tt.want {
			t.Errorf("%s = %q, want %q", tt.input, got, tt.want)
		}
	}

	runExpectVmError(t, `{"a": 1} + {"b": 2}`, "unknown operator: MAP OpAdd MAP")

	tests2 := []struct {
		input string
		want  string
	}{
		{`var a = {"__add": fun(o) { 100 }}; var b = {"__add": fun(o) { 5 }}; a + b`, "100"},
		{`var a = {"__eq": fun(o) { true }}; var b = {"__eq": fun(o) { false }}; a == b`, "true"},
	}
	for _, tt := range tests2 {
		if got := runInspect(t, tt.input); got != tt.want {
			t.Errorf("%s = %q, want %q", tt.input, got, tt.want)
		}
	}

	runExpectVmError(t, `var m = {"__add": fun(o) { 1 }}; m + 21`, "type mismatch")
}

func TestNullCoalescingAndBooleanOps(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`null || 5`, "5"},
		{`null || "fallback"`, "fallback"},
		{`true && false`, "false"},
		{`true || false`, "true"},
	}
	for _, tt := range tests {
		if got := runInspect(t, tt.input); got != tt.want {
			t.Errorf("%s = %q, want %q", tt.input, got, tt.want)
		}
	}

	runExpectVmError(t, `1 && 2`, "unknown operator: INTEGER OpAnd INTEGER")
	runExpectVmError(t, `[1] && [2]`, "unknown operator: LIST OpAnd LIST")
}

func TestTypeMismatchErrors(t *testing.T) {
	runExpectVmError(t, `"a" > 1`, "type mismatch: STRING OpGreaterThan INTEGER")
	runExpectVmError(t, `{1} & [2]`, "type mismatch")
}

func TestENVIndexAssignment(t *testing.T) {
	t.Setenv("BLUE_TEST_VAR", "orig")

	if got := runInspect(t, `ENV["BLUE_TEST_VAR"] = "updated"; ENV["BLUE_TEST_VAR"]`); got != "updated" {
		t.Errorf("ENV set failed: got %q", got)
	}
	if os.Getenv("BLUE_TEST_VAR") != "updated" {
		t.Errorf("os env not updated: %q", os.Getenv("BLUE_TEST_VAR"))
	}

	if got := runInspect(t, `ENV["BLUE_TEST_VAR"] = null; ENV["BLUE_TEST_VAR"]`); got != "null" {
		t.Errorf("ENV unset failed: got %q", got)
	}
	if os.Getenv("BLUE_TEST_VAR") != "" {
		t.Errorf("os env not unset: %q", os.Getenv("BLUE_TEST_VAR"))
	}

	runExpectVmError(t, `ENV[5] = "x"`, "ENV requires string key")
	runExpectVmError(t, `ENV["X"] = 5`, "ENV requires string value or null")
}
