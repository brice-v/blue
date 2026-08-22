package object

import (
	"regexp"
	"testing"
	"time"
)

func timeBuiltinFn(t *testing.T, name string) BuiltinFunction {
	t.Helper()
	for _, b := range TimeBuiltins {
		if b.Name == name {
			if b.Fun == nil {
				t.Fatalf("time builtin %q has nil Fun", name)
			}
			return b.Fun
		}
	}
	t.Fatalf("time builtin %q not found", name)
	return nil
}

func TestTimeRegistry(t *testing.T) {
	seen := make(map[string]bool)
	for _, b := range TimeBuiltins {
		if b.Name == "" || b.HelpStr == "" {
			t.Fatalf("time builtin %q missing Name or HelpStr", b.Name)
		}
		if seen[b.Name] {
			t.Errorf("duplicate time builtin name %q", b.Name)
		}
		seen[b.Name] = true
		if b.Fun == nil {
			t.Errorf("time builtin %q has nil Fun", b.Name)
		}
	}
}

func TestSleepBuiltin(t *testing.T) {
	sleepFn := timeBuiltinFn(t, "_sleep")
	start := time.Now()
	res := sleepFn(in(1))
	if res != NULL {
		t.Errorf("sleep(1) = %s, want NULL", res.Inspect())
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("sleep(1) took %v, should be ~1ms", elapsed)
	}
	runBuiltinTestsFor(t, TimeBuiltins, "_sleep", []builtinTestCase{
		{name: "negative", args: []Object{in(-1)}, err: "must be > 0"},
		{name: "not an int", args: []Object{fl(1)}, err: "PositionalTypeError"},
		{name: "no args", args: []Object{}, err: "InvalidArgCountError"},
		{name: "too many args", args: []Object{in(1), in(2)}, err: "InvalidArgCountError"},
	})
}

func TestNowBuiltin(t *testing.T) {
	nowFn := timeBuiltinFn(t, "_now")
	before := time.Now().UnixMilli() - 5000
	res := nowFn()
	after := time.Now().UnixMilli() + 5000
	n, ok := res.(*Integer)
	if !ok {
		t.Fatalf("_now returned %T, want *Integer", res)
	}
	if n.Value < before || n.Value > after {
		t.Errorf("_now() = %d, want within [%d, %d]", n.Value, before, after)
	}
	runBuiltinTestsFor(t, TimeBuiltins, "_now", []builtinTestCase{
		{name: "takes no args", args: []Object{in(1)}, err: "InvalidArgCountError"},
	})
}

func TestParseBuiltin(t *testing.T) {
	parseFn := timeBuiltinFn(t, "_parse")
	tests := []string{
		"2024-01-02T03:04:05Z",
		"2024-01-02T03:04:05+00:00",
	}
	for _, s := range tests {
		ref, refErr := time.Parse(time.RFC3339, "2024-01-02T03:04:05Z")
		if refErr != nil {
			t.Fatal(refErr)
		}
		want := ref.UnixMilli()
		res := parseFn(&Stringo{Value: s})
		got, ok := res.(*Integer)
		if !ok {
			t.Errorf("parse(%q) returned %T, want *Integer", s, res)
			continue
		}
		if got.Value != want {
			t.Errorf("parse(%q) = %d, want %d (RFC3339 reference)", s, got.Value, want)
		}
	}
	badRes := parseFn(&Stringo{Value: "this is not a date"})
	if _, ok := badRes.(*Integer); !ok {
		t.Errorf("parse(garbage) returned %T, want *Integer (no panic)", badRes)
	}
	runBuiltinTestsFor(t, TimeBuiltins, "_parse", []builtinTestCase{
		{name: "wrong type", args: []Object{in(123)}, err: "PositionalTypeError"},
		{name: "no args", args: []Object{}, err: "InvalidArgCountError"},
		{name: "too many args", args: []Object{&Stringo{Value: "a"}, &Stringo{Value: "b"}}, err: "InvalidArgCountError"},
	})
}

func TestToStrBuiltin(t *testing.T) {
	toStrFn := timeBuiltinFn(t, "_to_str")
	ts := int64(1703479130144)

	localRes := toStrFn(in(ts), NULL)
	s, ok := localRes.(*Stringo)
	if !ok {
		t.Fatalf("to_str(ts, null) returned %T, want *Stringo", localRes)
	}
	wantLocal := time.UnixMilli(ts).Format("2006-01-02 15:04:05.999")
	if s.Value != wantLocal {
		t.Errorf("to_str(ts, null) = %q, want %q", s.Value, wantLocal)
	}

	utcRes := toStrFn(in(ts), &Stringo{Value: "UTC"})
	u := utcRes.(*Stringo)
	wantUTC := time.UnixMilli(ts).UTC().Format("2006-01-02 15:04:05.999")
	if u.Value != wantUTC {
		t.Errorf("to_str(ts, 'UTC') = %q, want %q", u.Value, wantUTC)
	}

	pattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}(\.\d{1,3})?$`)
	if !pattern.MatchString(s.Value) || !pattern.MatchString(u.Value) {
		t.Errorf("formatted output does not match datetime pattern: %q / %q", s.Value, u.Value)
	}

	roundTrip := timeBuiltinFn(t, "_parse")(&Stringo{Value: u.Value})
	back, ok := roundTrip.(*Integer)
	if !ok {
		t.Fatalf("parse(to_str(ts,'UTC')) returned %T", roundTrip)
	}
	if back.Value < ts-60000 || back.Value > ts+60000 {
		t.Errorf("round trip through UTC string drifted too far: got %d, started at %d", back.Value, ts)
	}

	runBuiltinTestsFor(t, TimeBuiltins, "_to_str", []builtinTestCase{
		{name: "timestamp not int", args: []Object{&Float{Value: 1.5}, NULL}, err: "PositionalTypeError"},
		{name: "tz not string or null", args: []Object{in(ts), in(0)}, err: "PositionalTypeError"},
		{name: "no args", args: []Object{}, err: "InvalidArgCountError"},
		{name: "one arg", args: []Object{in(ts)}, err: "InvalidArgCountError"},
	})
}
