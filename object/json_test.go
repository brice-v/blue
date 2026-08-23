package object

import (
	"testing"
)

// TestFromJsonParity checks the dedicated JSON converter (object/json.go,
// used by from_json in every build) against a set of representative
// documents. The historical parser-based implementation lives in
// object/astjson and is compared in that package's tests; here we pin the
// exact expected Inspect() output.
func TestFromJsonString(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{`{}`, "{}"},
		{`[]`, "[]"},
		{`null`, "null"},
		{`true`, "true"},
		{`false`, "false"},
		{`123`, "123"},
		{`-456`, "-456"},
		{`1.5`, "1.5"},
		{`"hello world"`, "hello world"},
		{`[1, 2, 3]`, "[1, 2, 3]"},
		{`{"a": 1, "b": "two"}`, `{a: 1, b: two}`},
		{`{"nested": {"deep": [true, null, -2.5]}}`, `{nested: {deep: [true, null, -2.5]}}`},
		// Big numbers
		{`123456789012345678901234567890`, "123456789012345678901234567890"},
		{`-987654321098765432109876543210`, "-987654321098765432109876543210"},
	}
	for _, c := range cases {
		got := FromJsonString(c.input)
		if got.Type() == ERROR_OBJ {
			t.Errorf("from_json(%q) returned error: %s", c.input, got.Inspect())
			continue
		}
		if got.Inspect() != c.want {
			t.Errorf("from_json(%q) = %q, want %q", c.input, got.Inspect(), c.want)
		}
	}

	errorCases := []string{
		`{invalid}`,
		`{"a": 1,`,
		`[1, 2`,
		`1 2`,
		`nul`,
	}
	for _, input := range errorCases {
		got := FromJsonString(input)
		if got.Type() != ERROR_OBJ {
			t.Errorf("from_json(%q) should be an error, got %s (%q)", input, got.Type(), got.Inspect())
		}
	}
}
