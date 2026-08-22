package object

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func coreBuiltinFn(t *testing.T, name string) BuiltinFunction {
	t.Helper()
	for _, b := range Builtins {
		if b.Name == name {
			if b.Fun == nil {
				t.Fatalf("builtin %q has nil Fun", name)
			}
			return b.Fun
		}
	}
	t.Fatalf("builtin %q not found", name)
	return nil
}

func intObjs(vs ...int64) []Object {
	objs := make([]Object, len(vs))
	for i, v := range vs {
		objs[i] = &Integer{Value: v}
	}
	return objs
}

func strObjs(vs ...string) []Object {
	objs := make([]Object, len(vs))
	for i, v := range vs {
		objs[i] = &Stringo{Value: v}
	}
	return objs
}

func newTestMap(kvs ...Object) *Map {
	m := &Map{Pairs: NewPairsMap()}
	for i := 0; i < len(kvs); i += 2 {
		hk := HashKey{Type: kvs[i].Type(), Value: HashObject(kvs[i])}
		m.Pairs.Set(hk, MapPair{Key: kvs[i], Value: kvs[i+1]})
	}
	return m
}

func newTestSet(elems ...Object) *Set {
	s := &Set{Elements: NewSetElementsWithSize(len(elems))}
	for _, e := range elems {
		s.Elements.Set(HashObject(e), SetPair{Value: e, Present: struct{}{}})
	}
	return s
}

type builtinTestCase struct {
	name string
	args []Object
	want string
	err  string
}

func runBuiltinTests(t *testing.T, builtinName string, tests []builtinTestCase) {
	t.Helper()
	runBuiltinTestsFor(t, Builtins, builtinName, tests)
}

func runBuiltinTestsFor(t *testing.T, list []*Builtin, builtinName string, tests []builtinTestCase) {
	t.Helper()
	var fn BuiltinFunction
	for _, b := range list {
		if b.Name == builtinName {
			fn = b.Fun
			break
		}
	}
	if fn == nil {
		t.Fatalf("builtin %q not found or has nil Fun", builtinName)
	}
	for _, tt := range tests {
		got := fn(tt.args...)
		if tt.err != "" {
			errObj, ok := got.(*Error)
			if !ok {
				t.Errorf("%s(%s): expected Error containing %q, got %T %s", builtinName, inspectArgs(tt.args), tt.err, got, got.Inspect())
				continue
			}
			if !strings.Contains(errObj.Message, tt.err) {
				t.Errorf("%s(%s): error message %q does not contain %q", builtinName, inspectArgs(tt.args), errObj.Message, tt.err)
			}
			continue
		}
		if isError(got) {
			t.Errorf("%s(%s): unexpected error: %s", builtinName, inspectArgs(tt.args), got.Inspect())
			continue
		}
		if got.Inspect() != tt.want {
			t.Errorf("%s(%s) = %s, want %s", builtinName, inspectArgs(tt.args), got.Inspect(), tt.want)
		}
	}
}

func inspectArgs(args []Object) string {
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = a.Inspect()
	}
	return strings.Join(parts, ", ")
}

func TestBuiltinRegistry(t *testing.T) {
	seen := make(map[string]bool)
	for _, b := range Builtins {
		if b.Name == "" {
			t.Fatal("found builtin with empty Name")
		}
		if seen[b.Name] {
			t.Errorf("duplicate builtin name %q", b.Name)
		}
		seen[b.Name] = true
		if b.HelpStr == "" {
			t.Errorf("builtin %q has empty HelpStr", b.Name)
		}
	}
	vmImplemented := map[string]bool{
		"println": true, "print": true, "str": true, "to_num": true,
		"_sort": true, "_sorted": true, "all": true, "any": true,
		"map": true, "filter": true, "load": true,
	}
	for _, b := range Builtins {
		if b.Fun == nil && !vmImplemented[b.Name] {
			t.Errorf("builtin %q has nil Fun but is not known to be VM-implemented", b.Name)
		}
	}
}

func TestGetBuiltin(t *testing.T) {
	tests := []builtinTestCase{
		{name: "string index", args: []Object{&Stringo{Value: "abc"}, &Integer{Value: 1}}, want: "b"},
		{name: "string index with index", args: []Object{&Stringo{Value: "abc"}, &Integer{Value: 2}, TRUE}, want: "[2, c]"},
		{name: "list index", args: []Object{&List{Elements: intObjs(10, 20)}, &Integer{Value: 0}}, want: "10"},
		{name: "list index with index", args: []Object{&List{Elements: intObjs(10, 20)}, &Integer{Value: 1}, TRUE}, want: "[1, 20]"},
		{name: "set index", args: []Object{newTestSet(&Integer{Value: 7}, &Integer{Value: 8}), &Integer{Value: 1}}, want: "8"},
		{name: "map index", args: []Object{newTestMap(&Stringo{Value: "a"}, &Integer{Value: 1}, &Stringo{Value: "b"}, &Integer{Value: 2}), &Integer{Value: 1}}, want: "2"},
		{name: "map index with index", args: []Object{newTestMap(&Stringo{Value: "a"}, &Integer{Value: 1}, &Stringo{Value: "b"}, &Integer{Value: 2}), &Integer{Value: 0}, TRUE}, want: "[a, 1]"},
		{name: "string out of bounds", args: []Object{&Stringo{Value: "abc"}, &Integer{Value: 5}}, err: "index 5 out of bounds"},
		{name: "negative index", args: []Object{&Stringo{Value: "abc"}, &Integer{Value: -1}}, err: "index -1 out of bounds"},
		{name: "list out of bounds", args: []Object{&List{Elements: intObjs(1)}, &Integer{Value: 3}}, err: "index 3 out of bounds"},
		{name: "wrong arg count", args: []Object{&Stringo{Value: "abc"}}, err: "InvalidArgCountError"},
		{name: "bad container type", args: []Object{&Integer{Value: 1}, &Integer{Value: 0}}, err: "PositionalTypeError"},
		{name: "bad index type", args: []Object{&Stringo{Value: "abc"}, &Stringo{Value: "x"}}, err: "PositionalTypeError"},
		{name: "bad with_index type", args: []Object{&Stringo{Value: "abc"}, &Integer{Value: 0}, &Integer{Value: 1}}, err: "PositionalTypeError"},
	}
	runBuiltinTests(t, "_get_", tests)
}

func TestKeysValuesDel(t *testing.T) {
	runBuiltinTests(t, "keys", []builtinTestCase{
		{name: "keys of map", args: []Object{newTestMap(&Stringo{Value: "a"}, &Integer{Value: 1}, &Stringo{Value: "b"}, &Integer{Value: 2})}, want: "[a, b]"},
		{name: "keys of empty map", args: []Object{newTestMap()}, want: "[]"},
		{name: "keys wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
		{name: "keys wrong type", args: []Object{&List{Elements: intObjs(1)}}, err: "PositionalTypeError"},
	})
	runBuiltinTests(t, "values", []builtinTestCase{
		{name: "values of map", args: []Object{newTestMap(&Stringo{Value: "a"}, &Integer{Value: 1}, &Stringo{Value: "b"}, &Integer{Value: 2})}, want: "[1, 2]"},
		{name: "values wrong type", args: []Object{&Integer{Value: 1}}, err: "PositionalTypeError"},
	})

	del := coreBuiltinFn(t, "del")
	l := &List{Elements: intObjs(1, 2, 3)}
	if res := del(l, &Integer{Value: 1}); isError(res) {
		t.Fatalf("del on list errored: %s", res.Inspect())
	}
	if got := l.Inspect(); got != "[1, 3]" {
		t.Errorf("del(list, 1) left list as %s, want [1, 3]", got)
	}
	if res := del(l, &Integer{Value: 10}); !isError(res) {
		t.Errorf("del(list, out-of-range index) should error, got %T %s", res, res.Inspect())
	}
	m := newTestMap(&Stringo{Value: "a"}, &Integer{Value: 1}, &Stringo{Value: "b"}, &Integer{Value: 2})
	del(m, &Stringo{Value: "a"})
	if got := m.Inspect(); got != "{b: 2}" {
		t.Errorf("del(map, 'a') left map as %s, want {b: 2}", got)
	}
	s := newTestSet(intObjs(1, 2, 3)...)
	del(s, &Integer{Value: 2})
	if got := s.Inspect(); got != "{1, 3}" {
		t.Errorf("del(set, 2) left set as %s, want {1, 3}", got)
	}
	runBuiltinTests(t, "del", []builtinTestCase{
		{name: "wrong arg count", args: []Object{&List{Elements: intObjs(1)}}, err: "InvalidArgCountError"},
		{name: "non-collection", args: []Object{&Integer{Value: 1}, &Integer{Value: 0}}, err: "PositionalTypeError"},
		{name: "index not integer", args: []Object{&List{Elements: intObjs(1)}, &Stringo{Value: "x"}}, err: "PositionalTypeError"},
	})
}

func TestLenNew(t *testing.T) {
	runBuiltinTests(t, "len", []builtinTestCase{
		{name: "string", args: []Object{&Stringo{Value: "hello"}}, want: "5"},
		{name: "unicode string counts runes", args: []Object{&Stringo{Value: "héllo"}}, want: "5"},
		{name: "empty string", args: []Object{&Stringo{Value: ""}}, want: "0"},
		{name: "list", args: []Object{&List{Elements: intObjs(1, 2, 3)}}, want: "3"},
		{name: "map", args: []Object{newTestMap(&Stringo{Value: "a"}, &Integer{Value: 1})}, want: "1"},
		{name: "set", args: []Object{newTestSet(intObjs(1, 2)...)}, want: "2"},
		{name: "bytes", args: []Object{&Bytes{Value: []byte("abcd")}}, want: "4"},
		{name: "unsupported type", args: []Object{&Integer{Value: 1}}, err: "PositionalTypeError"},
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
	runBuiltinTests(t, "new", []builtinTestCase{
		{name: "clone list", args: []Object{&List{Elements: intObjs(1, 2)}}, want: "[1, 2]"},
		{name: "clone map", args: []Object{newTestMap(&Stringo{Value: "a"}, &Integer{Value: 1})}, want: "{a: 1}"},
		{name: "clone set", args: []Object{newTestSet(intObjs(1)...)}, want: "{1}"},
		{name: "unsupported type", args: []Object{&Integer{Value: 1}}, err: "PositionalTypeError"},
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
	newFn := coreBuiltinFn(t, "new")
	orig := &List{Elements: intObjs(1)}
	cloned := newFn(orig).(*List)
	cloned.Elements[0] = &Integer{Value: 99}
	if orig.Elements[0].(*Integer).Value != 1 {
		t.Error("new() did not return an independent clone")
	}
}

func TestListMutatingBuiltins(t *testing.T) {
	push := coreBuiltinFn(t, "push")
	l := &List{Elements: intObjs(1, 2)}
	if res := push(l, &Integer{Value: 3}, &Integer{Value: 4}); res.Inspect() != "4" {
		t.Errorf("push returned %s, want 4 (new length)", res.Inspect())
	}
	if l.Inspect() != "[1, 2, 3, 4]" {
		t.Errorf("push left list as %s", l.Inspect())
	}

	pop := coreBuiltinFn(t, "pop")
	if res := pop(l); res.Inspect() != "4" {
		t.Errorf("pop returned %s, want 4", res.Inspect())
	}
	if pop(&List{Elements: []Object{}}) != NULL {
		t.Error("pop on empty list should return NULL")
	}

	unshift := coreBuiltinFn(t, "unshift")
	l2 := &List{Elements: intObjs(2, 3)}
	if res := unshift(l2, &Integer{Value: 1}); res.Inspect() != "3" {
		t.Errorf("unshift returned %s, want 3", res.Inspect())
	}
	if l2.Inspect() != "[1, 2, 3]" {
		t.Errorf("unshift left list as %s", l2.Inspect())
	}

	shift := coreBuiltinFn(t, "shift")
	if res := shift(l2); res.Inspect() != "1" {
		t.Errorf("shift returned %s, want 1", res.Inspect())
	}
	if shift(&List{Elements: []Object{}}) != NULL {
		t.Error("shift on empty list should return NULL")
	}
	single := &List{Elements: intObjs(9)}
	if res := shift(single); res.Inspect() != "9" {
		t.Errorf("shift single returned %s, want 9", res.Inspect())
	}
	if single.Inspect() != "[]" {
		t.Errorf("shift on single-element list left %s, want []", single.Inspect())
	}

	runBuiltinTests(t, "append", []builtinTestCase{
		{name: "append one", args: []Object{&List{Elements: intObjs(1)}, &Integer{Value: 2}}, want: "[1, 2]"},
		{name: "append many", args: []Object{&List{Elements: intObjs(1)}, &Integer{Value: 2}, &Stringo{Value: "x"}}, want: "[1, 2, x]"},
		{name: "does not mutate input", args: []Object{&List{Elements: intObjs(1)}, &Integer{Value: 2}}, want: "[1, 2]"},
		{name: "wrong arg count", args: []Object{&List{Elements: intObjs(1)}}, err: "InvalidArgCountError"},
		{name: "wrong type", args: []Object{&Stringo{Value: "ab"}, &Integer{Value: 1}}, err: "PositionalTypeError"},
	})
	runBuiltinTests(t, "prepend", []builtinTestCase{
		{name: "prepend one", args: []Object{&List{Elements: intObjs(2)}, &Integer{Value: 1}}, want: "[1, 2]"},
		{name: "prepend many", args: []Object{&List{Elements: intObjs(3)}, &Integer{Value: 1}, &Integer{Value: 2}}, want: "[1, 2, 3]"},
		{name: "wrong arg count", args: []Object{&List{Elements: intObjs(1)}}, err: "InvalidArgCountError"},
		{name: "wrong type", args: []Object{&Boolean{Value: true}, &Integer{Value: 1}}, err: "PositionalTypeError"},
	})
	runBuiltinTests(t, "concat", []builtinTestCase{
		{name: "two lists", args: []Object{&List{Elements: intObjs(1, 2)}, &List{Elements: intObjs(3)}}, want: "[1, 2, 3]"},
		{name: "three lists", args: []Object{&List{Elements: intObjs(1)}, &List{Elements: []Object{}}, &List{Elements: strObjs("a")}}, want: "[1, a]"},
		{name: "too few args", args: []Object{&List{Elements: intObjs(1)}}, err: "InvalidArgCountError"},
		{name: "non-list element", args: []Object{&List{Elements: intObjs(1)}, &Integer{Value: 2}}, err: "PositionalTypeError"},
	})
	runBuiltinTests(t, "reverse", []builtinTestCase{
		{name: "reverse list", args: []Object{&List{Elements: intObjs(1, 2, 3)}}, want: "[3, 2, 1]"},
		{name: "reverse string", args: []Object{&Stringo{Value: "abc"}}, want: "cba"},
		{name: "wrong type", args: []Object{&Integer{Value: 1}}, err: "PositionalTypeError"},
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
	for _, name := range []string{"push", "pop", "unshift", "shift"} {
		fn := coreBuiltinFn(t, name)
		if res := fn(&Stringo{Value: "ab"}, &Integer{Value: 1}); !isError(res) {
			t.Errorf("%s on non-list should error, got %s", name, res.Inspect())
		}
	}
}

func TestSetToListZip(t *testing.T) {
	runBuiltinTests(t, "set", []builtinTestCase{
		{name: "dedupe list", args: []Object{&List{Elements: intObjs(1, 2, 2, 3)}}, want: "{1, 2, 3}"},
		{name: "empty set", args: []Object{}, want: "{}"},
		{name: "wrong type", args: []Object{&Integer{Value: 1}}, err: "PositionalTypeError"},
		{name: "wrong arg count", args: []Object{&Integer{Value: 1}, &Integer{Value: 2}}, err: "InvalidArgCountError"},
	})
	runBuiltinTests(t, "to_list", []builtinTestCase{
		{name: "set to list preserves order", args: []Object{newTestSet(intObjs(1, 2, 3)...)}, want: "[1, 2, 3]"},
		{name: "empty set", args: []Object{newTestSet()}, want: "[]"},
		{name: "wrong type", args: []Object{&List{Elements: intObjs(1)}}, err: "PositionalTypeError"},
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
	runBuiltinTests(t, "zip", []builtinTestCase{
		{name: "two equal lists", args: []Object{&List{Elements: []Object{&List{Elements: intObjs(1, 2)}, &List{Elements: intObjs(3, 4)}}}}, want: "[[1, 3], [2, 4]]"},
		{name: "uneven lengths uses min", args: []Object{&List{Elements: []Object{&List{Elements: intObjs(1, 2, 3)}, &List{Elements: intObjs(4, 5)}}}}, want: "[[1, 4], [2, 5]]"},
		{name: "single list", args: []Object{&List{Elements: []Object{&List{Elements: intObjs(1, 2)}}}}, want: "[[1], [2]]"},
		{name: "element not list", args: []Object{&List{Elements: []Object{&Integer{Value: 1}}}}, err: "`zip` error"},
		{name: "wrong arg type", args: []Object{&Integer{Value: 1}}, err: "PositionalTypeError"},
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
}

func TestConversionBuiltins(t *testing.T) {
	runBuiltinTests(t, "int", []builtinTestCase{
		{name: "from float truncates", args: []Object{&Float{Value: 3.9}}, want: "3"},
		{name: "from negative float truncates", args: []Object{&Float{Value: -3.9}}, want: "-3"},
		{name: "from uint", args: []Object{&UInteger{Value: 42}}, want: "42"},
		{name: "from bigint", args: []Object{mustBigint(t, "123456789")}, want: "123456789"},
		{name: "from bigfloat", args: []Object{mustBigfloat(t, "9.5")}, want: "9"},
		{name: "identity", args: []Object{&Integer{Value: 7}}, want: "7"},
		{name: "from string", args: []Object{&Stringo{Value: "123"}}, want: "123"},
		{name: "invalid string", args: []Object{&Stringo{Value: "12x"}}, err: "`int` error"},
		{name: "unsupported type", args: []Object{TRUE}, err: "`int` error: unsupported type"},
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
	runBuiltinTests(t, "float", []builtinTestCase{
		{name: "from int", args: []Object{&Integer{Value: 3}}, want: "3"},
		{name: "from uint", args: []Object{&UInteger{Value: 2}}, want: "2"},
		{name: "identity", args: []Object{&Float{Value: 1.5}}, want: "1.5"},
		{name: "from string", args: []Object{&Stringo{Value: "2.5"}}, want: "2.5"},
		{name: "invalid string", args: []Object{&Stringo{Value: "2.5x"}}, err: "`float` error"},
		{name: "unsupported type", args: []Object{TRUE}, err: "`float` error: unsupported type"},
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
	runBuiltinTests(t, "bigint", []builtinTestCase{
		{name: "from int", args: []Object{&Integer{Value: 5}}, want: "5"},
		{name: "from float truncates", args: []Object{&Float{Value: 5.9}}, want: "5"},
		{name: "from uint", args: []Object{&UInteger{Value: 18446744073709551615}}, want: "18446744073709551615"},
		{name: "identity", args: []Object{mustBigint(t, "999999999999999999999")}, want: "999999999999999999999"},
		{name: "from bigfloat", args: []Object{mustBigfloat(t, "7.2")}, want: "7"},
		{name: "from string", args: []Object{&Stringo{Value: "123456789012345678901234567890"}}, want: "123456789012345678901234567890"},
		{name: "invalid string", args: []Object{&Stringo{Value: "not a number"}}, err: "`bigint` error"},
		{name: "unsupported type", args: []Object{TRUE}, err: "`bigint` error: unsupported type"},
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
	runBuiltinTests(t, "bigfloat", []builtinTestCase{
		{name: "from int", args: []Object{&Integer{Value: 5}}, want: "5"},
		{name: "from float", args: []Object{&Float{Value: 1.25}}, want: "1.25"},
		{name: "identity", args: []Object{mustBigfloat(t, "3.14")}, want: "3.14"},
		{name: "from bigint", args: []Object{mustBigint(t, "1000000000000000000001")}, want: "1000000000000000000001"},
		{name: "from string", args: []Object{&Stringo{Value: "2.75"}}, want: "2.75"},
		{name: "invalid string", args: []Object{&Stringo{Value: "xyz"}}, err: "`bigfloat` error"},
		{name: "unsupported type", args: []Object{TRUE}, err: "`bigfloat` error: unsupported type"},
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
	runBuiltinTests(t, "uint", []builtinTestCase{
		{name: "from int", args: []Object{&Integer{Value: 9}}, want: "9"},
		{name: "from float truncates", args: []Object{&Float{Value: 9.9}}, want: "9"},
		{name: "identity", args: []Object{&UInteger{Value: 11}}, want: "11"},
		{name: "from string", args: []Object{&Stringo{Value: "77"}}, want: "77"},
		{name: "negative string invalid", args: []Object{&Stringo{Value: "-1"}}, err: "`uint` error"},
		{name: "unsupported type", args: []Object{TRUE}, err: "`uint` error: unsupported type"},
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
	runBuiltinTests(t, "_to_bytes", []builtinTestCase{
		{name: "ascii", args: []Object{&Stringo{Value: "AB"}}, want: "[]byte{0x41, 0x42}"},
		{name: "empty string", args: []Object{&Stringo{Value: ""}}, want: "[]byte{}"},
		{name: "wrong type", args: []Object{&Integer{Value: 1}}, err: "PositionalTypeError"},
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
}

func mustBigint(t *testing.T, s string) Object {
	t.Helper()
	res := coreBuiltinFn(t, "bigint")(&Stringo{Value: s})
	if isError(res) {
		t.Fatalf("bigint(%s) failed: %s", s, res.Inspect())
	}
	return res
}

func mustBigfloat(t *testing.T, s string) Object {
	t.Helper()
	res := coreBuiltinFn(t, "bigfloat")(&Stringo{Value: s})
	if isError(res) {
		t.Fatalf("bigfloat(%s) failed: %s", s, res.Inspect())
	}
	return res
}

func TestIntFromBigNumber(t *testing.T) {
	bi := mustBigint(t, "123456789012345678")
	res := coreBuiltinFn(t, "int")(bi)
	i, ok := res.(*Integer)
	if !ok {
		t.Fatalf("int(bigint) returned %T", res)
	}
	if i.Value != 123456789012345678 {
		t.Errorf("int(bigint) = %d, want 123456789012345678", i.Value)
	}
}

func TestStringBuiltins(t *testing.T) {
	runBuiltinTests(t, "split", []builtinTestCase{
		{name: "default separator is space", args: []Object{&Stringo{Value: "a b c"}}, want: "[a, b, c]"},
		{name: "explicit separator", args: []Object{&Stringo{Value: "a,b,c"}, &Stringo{Value: ","}}, want: "[a, b, c]"},
		{name: "regex separator", args: []Object{&Stringo{Value: "a1b22c"}, &Regex{Value: regexp.MustCompile("[0-9]+")}}, want: "[a, b, c]"},
		{name: "no match returns whole string", args: []Object{&Stringo{Value: "abc"}, &Stringo{Value: ","}}, want: "[abc]"},
		{name: "wrong type", args: []Object{&Integer{Value: 1}}, err: "PositionalTypeError"},
		{name: "bad second arg type", args: []Object{&Stringo{Value: "abc"}, &Integer{Value: 1}}, err: "PositionalTypeError"},
		{name: "too many args", args: []Object{&Stringo{Value: "a"}, &Stringo{Value: "b"}, &Stringo{Value: "c"}}, err: "InvalidArgCountError"},
	})
	runBuiltinTests(t, "_replace", []builtinTestCase{
		{name: "replace all", args: []Object{&Stringo{Value: "Hello"}, &Stringo{Value: "l"}, &Stringo{Value: "X"}}, want: "HeXXo"},
		{name: "no occurrences", args: []Object{&Stringo{Value: "Hello"}, &Stringo{Value: "z"}, &Stringo{Value: "X"}}, want: "Hello"},
		{name: "wrong arg count", args: []Object{&Stringo{Value: "a"}, &Stringo{Value: "b"}}, err: "InvalidArgCountError"},
		{name: "wrong type", args: []Object{&Stringo{Value: "a"}, &Integer{Value: 1}, &Stringo{Value: "c"}}, err: "PositionalTypeError"},
	})
	runBuiltinTests(t, "_replace_regex", []builtinTestCase{
		{name: "pattern from string", args: []Object{&Stringo{Value: "a1b2"}, &Stringo{Value: "[0-9]"}, &Stringo{Value: "-"}}, want: "a-b-"},
		{name: "regex object", args: []Object{&Stringo{Value: "a1b2"}, &Regex{Value: regexp.MustCompile("[0-9]")}, &Stringo{Value: "-"}}, want: "a-b-"},
		{name: "invalid pattern", args: []Object{&Stringo{Value: "a"}, &Stringo{Value: "("}, &Stringo{Value: "-"}}, err: "`replace_regex` error"},
		{name: "wrong arg count", args: []Object{&Stringo{Value: "a"}}, err: "InvalidArgCountError"},
		{name: "bad replacer type", args: []Object{&Stringo{Value: "a"}, &Integer{Value: 1}, &Stringo{Value: "-"}}, err: "PositionalTypeError"},
	})
	runBuiltinTests(t, "strip", []builtinTestCase{
		{name: "whitespace both sides", args: []Object{&Stringo{Value: "  hi  "}}, want: "hi"},
		{name: "custom cutset", args: []Object{&Stringo{Value: "xxhixx"}, &Stringo{Value: "x"}}, want: "hi"},
		{name: "wrong type", args: []Object{&Integer{Value: 1}}, err: "PositionalTypeError"},
		{name: "too many args", args: []Object{&Stringo{Value: "a"}, &Stringo{Value: "b"}, &Stringo{Value: "c"}}, err: "InvalidArgCountError"},
	})
	runBuiltinTests(t, "lstrip", []builtinTestCase{
		{name: "whitespace left", args: []Object{&Stringo{Value: "  hi  "}}, want: "hi  "},
		{name: "cutset", args: []Object{&Stringo{Value: "aabbaa"}, &Stringo{Value: "a"}}, want: "bbaa"},
		{name: "wrong type", args: []Object{&Integer{Value: 1}}, err: "PositionalTypeError"},
	})
	runBuiltinTests(t, "rstrip", []builtinTestCase{
		{name: "whitespace right", args: []Object{&Stringo{Value: "  hi  "}}, want: "  hi"},
		{name: "cutset", args: []Object{&Stringo{Value: "aabbaa"}, &Stringo{Value: "a"}}, want: "aabb"},
		{name: "wrong type", args: []Object{&Integer{Value: 1}}, err: "PositionalTypeError"},
	})
	runBuiltinTests(t, "to_upper", []builtinTestCase{
		{name: "upper", args: []Object{&Stringo{Value: "Hello"}}, want: "HELLO"},
		{name: "already upper", args: []Object{&Stringo{Value: "ABC"}}, want: "ABC"},
		{name: "wrong type", args: []Object{&Integer{Value: 1}}, err: "PositionalTypeError"},
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
	runBuiltinTests(t, "to_lower", []builtinTestCase{
		{name: "lower", args: []Object{&Stringo{Value: "HeLLo"}}, want: "hello"},
		{name: "wrong type", args: []Object{&Float{Value: 1}}, err: "PositionalTypeError"},
	})
	runBuiltinTests(t, "join", []builtinTestCase{
		{name: "join strings", args: []Object{&List{Elements: strObjs("a", "b", "c")}, &Stringo{Value: "-"}}, want: "a-b-c"},
		{name: "join empty list", args: []Object{&List{Elements: []Object{}}, &Stringo{Value: "-"}}, want: ""},
		{name: "element not string", args: []Object{&List{Elements: intObjs(1)}, &Stringo{Value: "-"}}, err: "was not a STRING in `join`"},
		{name: "wrong arg count", args: []Object{&List{Elements: []Object{}}}, err: "InvalidArgCountError"},
		{name: "wrong joiner type", args: []Object{&List{Elements: []Object{}}, &Integer{Value: 1}}, err: "PositionalTypeError"},
	})
	runBuiltinTests(t, "_substr", []builtinTestCase{
		{name: "middle", args: []Object{&Stringo{Value: "Hello"}, &Integer{Value: 1}, &Integer{Value: 3}}, want: "el"},
		{name: "end -1 means rest", args: []Object{&Stringo{Value: "Hello"}, &Integer{Value: 2}, &Integer{Value: -1}}, want: "llo"},
		{name: "wrong arg count", args: []Object{&Stringo{Value: "Hi"}}, err: "InvalidArgCountError"},
		{name: "start not int", args: []Object{&Stringo{Value: "Hi"}, &Stringo{Value: "0"}, &Integer{Value: 1}}, err: "PositionalTypeError"},
	})
	runBuiltinTests(t, "index_of", []builtinTestCase{
		{name: "found", args: []Object{&Stringo{Value: "Hello"}, &Stringo{Value: "ell"}}, want: "1"},
		{name: "not found", args: []Object{&Stringo{Value: "Hello"}, &Stringo{Value: "z"}}, want: "-1"},
		{name: "wrong type", args: []Object{&Stringo{Value: "Hello"}, &Integer{Value: 1}}, err: "PositionalTypeError"},
		{name: "wrong arg count", args: []Object{&Stringo{Value: "Hello"}}, err: "InvalidArgCountError"},
	})
	runBuiltinTests(t, "_center", []builtinTestCase{
		{name: "center pads", args: []Object{&Stringo{Value: "Hi"}, &Integer{Value: 6}, &Stringo{Value: " "}}, want: "  Hi  "},
		{name: "wrong pad type", args: []Object{&Stringo{Value: "Hi"}, &Integer{Value: 6}, &Integer{Value: 1}}, err: "PositionalTypeError"},
		{name: "wrong arg count", args: []Object{&Stringo{Value: "Hi"}}, err: "InvalidArgCountError"},
	})
	runBuiltinTests(t, "_ljust", []builtinTestCase{
		{name: "left justify", args: []Object{&Stringo{Value: "Hi"}, &Integer{Value: 5}, &Stringo{Value: "."}}, want: "Hi..."},
		{name: "wrong length type", args: []Object{&Stringo{Value: "Hi"}, &Stringo{Value: "5"}, &Stringo{Value: "."}}, err: "PositionalTypeError"},
	})
	runBuiltinTests(t, "_rjust", []builtinTestCase{
		{name: "right justify", args: []Object{&Stringo{Value: "Hi"}, &Integer{Value: 5}, &Stringo{Value: "."}}, want: "...Hi"},
		{name: "wrong pad type", args: []Object{&Stringo{Value: "Hi"}, &Integer{Value: 5}, &Integer{Value: 0}}, err: "PositionalTypeError"},
	})
	runBuiltinTests(t, "to_title", []builtinTestCase{
		{name: "title case", args: []Object{&Stringo{Value: "hello world"}}, want: "Hello World"},
		{name: "wrong type", args: []Object{&Integer{Value: 1}}, err: "PositionalTypeError"},
	})
	runBuiltinTests(t, "to_kebab", []builtinTestCase{
		{name: "kebab case", args: []Object{&Stringo{Value: "hello world"}}, want: "hello-world"},
		{name: "wrong type", args: []Object{&Integer{Value: 1}}, err: "PositionalTypeError"},
	})
	runBuiltinTests(t, "to_camel", []builtinTestCase{
		{name: "camel case", args: []Object{&Stringo{Value: "hello world"}}, want: "helloWorld"},
		{name: "wrong type", args: []Object{&Integer{Value: 1}}, err: "PositionalTypeError"},
	})
	runBuiltinTests(t, "to_snake", []builtinTestCase{
		{name: "snake case", args: []Object{&Stringo{Value: "hello world"}}, want: "hello_world"},
		{name: "wrong type", args: []Object{&Integer{Value: 1}}, err: "PositionalTypeError"},
	})
	runBuiltinTests(t, "startswith", []builtinTestCase{
		{name: "true", args: []Object{&Stringo{Value: "Hello"}, &Stringo{Value: "Hel"}}, want: "true"},
		{name: "false", args: []Object{&Stringo{Value: "Hello"}, &Stringo{Value: "ell"}}, want: "false"},
		{name: "prefix not string", args: []Object{&Stringo{Value: "Hello"}, &Integer{Value: 1}}, err: "PositionalTypeError"},
		{name: "wrong arg count", args: []Object{&Stringo{Value: "Hello"}}, err: "InvalidArgCountError"},
	})
	runBuiltinTests(t, "endswith", []builtinTestCase{
		{name: "true", args: []Object{&Stringo{Value: "Hello"}, &Stringo{Value: "llo"}}, want: "true"},
		{name: "false", args: []Object{&Stringo{Value: "Hello"}, &Stringo{Value: "ell"}}, want: "false"},
		{name: "suffix not string", args: []Object{&Stringo{Value: "Hello"}, &Integer{Value: 1}}, err: "PositionalTypeError"},
	})
	runBuiltinTests(t, "contains", []builtinTestCase{
		{name: "true", args: []Object{&Stringo{Value: "hello world"}, &Stringo{Value: "lo w"}}, want: "true"},
		{name: "false", args: []Object{&Stringo{Value: "hello"}, &Stringo{Value: "z"}}, want: "false"},
		{name: "needle not string", args: []Object{&Stringo{Value: "hello"}, &Integer{Value: 1}}, err: "PositionalTypeError"},
		{name: "wrong arg count", args: []Object{&Stringo{Value: "hello"}}, err: "InvalidArgCountError"},
	})
	runBuiltinTests(t, "matches", []builtinTestCase{
		{name: "string and regex object", args: []Object{&Stringo{Value: "hello"}, &Regex{Value: regexp.MustCompile("^he")}}, want: "true"},
		{name: "regex object and string", args: []Object{&Regex{Value: regexp.MustCompile("^he")}, &Stringo{Value: "hello"}}, want: "true"},
		{name: "both strings", args: []Object{&Stringo{Value: "hello"}, &Stringo{Value: "^ll"}}, want: "false"},
		{name: "invalid pattern", args: []Object{&Stringo{Value: "hello"}, &Stringo{Value: "("}}, err: "`matches` error"},
		{name: "second arg wrong type", args: []Object{&Regex{Value: regexp.MustCompile("^he")}, &Integer{Value: 1}}, err: "PositionalTypeError"},
		{name: "first arg wrong type", args: []Object{&Integer{Value: 1}, &Stringo{Value: "a"}}, err: "PositionalTypeError"},
		{name: "wrong arg count", args: []Object{&Stringo{Value: "hello"}}, err: "InvalidArgCountError"},
	})
}

func TestJsonBuiltins(t *testing.T) {
	runBuiltinTests(t, "is_valid_json", []builtinTestCase{
		{name: "valid object", args: []Object{&Stringo{Value: `{"a": 1}`}}, want: "true"},
		{name: "valid array", args: []Object{&Stringo{Value: `[1, 2, 3]`}}, want: "true"},
		{name: "valid scalar", args: []Object{&Stringo{Value: `42`}}, want: "true"},
		{name: "invalid", args: []Object{&Stringo{Value: `{a: 1}`}}, want: "false"},
		{name: "wrong type", args: []Object{&Integer{Value: 1}}, err: "PositionalTypeError"},
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
	runBuiltinTests(t, "from_json", []builtinTestCase{
		{name: "object", args: []Object{&Stringo{Value: `{"a": 1}`}}, want: "{a: 1}"},
		{name: "array", args: []Object{&Stringo{Value: `[1, 2]`}}, want: "[1, 2]"},
		{name: "string value", args: []Object{&Stringo{Value: `"hi"`}}, want: "hi"},
		{name: "null", args: []Object{&Stringo{Value: `null`}}, want: "null"},
		{name: "invalid json", args: []Object{&Stringo{Value: `{a: }`}}, err: "invalid json"},
		{name: "wrong type", args: []Object{&Integer{Value: 1}}, err: "PositionalTypeError"},
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
	toJson := coreBuiltinFn(t, "to_json")
	roundTripInput := newTestMap(
		&Stringo{Value: "name"}, &Stringo{Value: "blue"},
		&Stringo{Value: "n"}, &Integer{Value: 5},
	)
	encoded := toJson(roundTripInput)
	s, ok := encoded.(*Stringo)
	if !ok {
		t.Fatalf("to_json returned %T", encoded)
	}
	decoded := coreBuiltinFn(t, "from_json")(s)
	dm, ok := decoded.(*Map)
	if !ok {
		t.Fatalf("from_json(to_json(m)) returned %T", decoded)
	}
	if dm.Pairs.Len() != 2 {
		t.Errorf("round trip map has %d pairs, want 2", dm.Pairs.Len())
	}
	nameHk := HashKey{Type: STRING_OBJ, Value: HashObject(&Stringo{Value: "name"})}
	pair, ok := dm.Pairs.Get(nameHk)
	if !ok || pair.Value.Inspect() != "blue" {
		t.Errorf("round trip lost name field: ok=%v pair=%v", ok, pair)
	}
	nHk := HashKey{Type: STRING_OBJ, Value: HashObject(&Stringo{Value: "n"})}
	pair, ok = dm.Pairs.Get(nHk)
	if !ok || pair.Value.Inspect() != "5" {
		t.Errorf("round trip lost n field: ok=%v pair=%v", ok, pair)
	}
	runBuiltinTests(t, "to_json", []builtinTestCase{
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
}

func TestMetaBuiltins(t *testing.T) {
	runBuiltinTests(t, "type", []builtinTestCase{
		{name: "integer", args: []Object{&Integer{Value: 1}}, want: "INTEGER"},
		{name: "string", args: []Object{&Stringo{Value: "s"}}, want: "STRING"},
		{name: "list", args: []Object{&List{Elements: []Object{}}}, want: "LIST"},
		{name: "map", args: []Object{newTestMap()}, want: "MAP"},
		{name: "set", args: []Object{newTestSet()}, want: "SET"},
		{name: "bool", args: []Object{TRUE}, want: "BOOLEAN"},
		{name: "float", args: []Object{&Float{Value: 1}}, want: "FLOAT"},
		{name: "null", args: []Object{NULL}, want: "NULL"},
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
	runBuiltinTests(t, "raw_type", []builtinTestCase{
		{name: "integer", args: []Object{&Integer{Value: 1}}, want: "*object.Integer"},
		{name: "string", args: []Object{&Stringo{Value: "s"}}, want: "*object.Stringo"},
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
	runBuiltinTests(t, "is_callable", []builtinTestCase{
		{name: "builtin is callable", args: []Object{&Builtin{Name: "len"}}, want: "true"},
		{name: "integer not callable", args: []Object{&Integer{Value: 1}}, want: "false"},
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
	hashFn := coreBuiltinFn(t, "__hash")
	a := hashFn(&Stringo{Value: "dup"})
	b := hashFn(&Stringo{Value: "dup"})
	if a.Inspect() != b.Inspect() {
		t.Errorf("__hash inconsistent for equal objects: %s vs %s", a.Inspect(), b.Inspect())
	}
	c := hashFn(&Stringo{Value: "other"})
	if a.Inspect() == c.Inspect() {
		t.Error("__hash equal for different objects")
	}
	runBuiltinTests(t, "__hash", []builtinTestCase{
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
	errFn := coreBuiltinFn(t, "error")
	errRes := errFn(&Stringo{Value: "boom"})
	if !isError(errRes) || errRes.(*Error).Message != "boom" {
		t.Errorf("error('boom') should return Error with message 'boom', got %v", errRes)
	}
	runBuiltinTests(t, "error", []builtinTestCase{
		{name: "wrong type", args: []Object{&Integer{Value: 1}}, err: "PositionalTypeError"},
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
	assertFn := coreBuiltinFn(t, "assert")
	if res := assertFn(TRUE); res != TRUE {
		t.Errorf("assert(true) = %s, want true", res.Inspect())
	}
	if res := assertFn(FALSE); !isError(res) {
		t.Errorf("assert(false) should be an error, got %s", res.Inspect())
	}
	if res := assertFn(FALSE, &Stringo{Value: "custom msg"}); !isError(res) || !strings.Contains(res.Inspect(), "custom msg") {
		t.Errorf("assert(false, msg) should contain the message, got %s", res.Inspect())
	}
	runBuiltinTests(t, "assert", []builtinTestCase{
		{name: "condition not bool", args: []Object{&Integer{Value: 1}}, err: "PositionalTypeError"},
		{name: "message not string", args: []Object{FALSE, &Integer{Value: 1}}, err: "PositionalTypeError"},
		{name: "wrong arg count", args: []Object{TRUE, TRUE, TRUE}, err: "InvalidArgCountError"},
	})
	helpFn := coreBuiltinFn(t, "help")
	res := helpFn(&Builtin{Name: "len", HelpStr: "len returns the length"})
	hs, ok := res.(*Stringo)
	if !ok || !strings.Contains(hs.Value, "len") {
		t.Errorf("help(len) should return its help string, got %#v", res)
	}
	runBuiltinTests(t, "help", []builtinTestCase{
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
}

func TestFileSystemBuiltins(t *testing.T) {
	cwdFn := coreBuiltinFn(t, "cwd")
	res := cwdFn()
	dir, ok := res.(*Stringo)
	if !ok || dir.Value == "" {
		t.Fatalf("cwd() returned %T %v", res, res)
	}
	runBuiltinTests(t, "cwd", []builtinTestCase{
		{name: "wrong arg count", args: []Object{&Stringo{Value: "x"}}, err: "InvalidArgCountError"},
	})

	tmp := t.TempDir()
	filePath := tmp + "/blue_test_file.txt"
	if err := os.WriteFile(filePath, []byte("data"), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	runBuiltinTests(t, "is_file", []builtinTestCase{
		{name: "existing file", args: []Object{&Stringo{Value: filePath}}, want: "true"},
		{name: "directory is not file", args: []Object{&Stringo{Value: tmp}}, want: "false"},
		{name: "missing path", args: []Object{&Stringo{Value: filePath + ".missing"}}, want: "false"},
		{name: "wrong type", args: []Object{&Integer{Value: 1}}, err: "PositionalTypeError"},
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
	runBuiltinTests(t, "is_dir", []builtinTestCase{
		{name: "existing dir", args: []Object{&Stringo{Value: tmp}}, want: "true"},
		{name: "file is not dir", args: []Object{&Stringo{Value: filePath}}, want: "false"},
		{name: "missing path", args: []Object{&Stringo{Value: filePath + ".missing"}}, want: "false"},
		{name: "wrong type", args: []Object{&Integer{Value: 1}}, err: "PositionalTypeError"},
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
	runBuiltinTests(t, "abs_path", []builtinTestCase{
		{name: "relative becomes absolute", args: []Object{&Stringo{Value: "some_file.txt"}}, want: dir.Value + "/some_file.txt"},
		{name: "wrong type", args: []Object{&Integer{Value: 1}}, err: "PositionalTypeError"},
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
}

func TestFmtAndReBuiltins(t *testing.T) {
	runBuiltinTests(t, "fmt", []builtinTestCase{
		{name: "binary format", args: []Object{&Integer{Value: 3}, &Stringo{Value: "%04b"}}, want: "0011"},
		{name: "string format", args: []Object{&Stringo{Value: "x"}, &Stringo{Value: "<%s>"}}, want: "<x>"},
		{name: "format not string", args: []Object{&Integer{Value: 3}, &Integer{Value: 1}}, err: "PositionalTypeError"},
		{name: "wrong arg count", args: []Object{&Integer{Value: 3}}, err: "InvalidArgCountError"},
	})
	runBuiltinTests(t, "re", []builtinTestCase{
		{name: "compile", args: []Object{&Stringo{Value: "a+"}}, want: "/a+/"},
		{name: "invalid pattern", args: []Object{&Stringo{Value: "("}}, err: "`re` error"},
		{name: "wrong type", args: []Object{&Integer{Value: 1}}, err: "PositionalTypeError"},
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
	saveFn := coreBuiltinFn(t, "save")
	enc := saveFn(&Integer{Value: 1234})
	bs, ok := enc.(*Bytes)
	if !ok || len(bs.Value) == 0 {
		t.Fatalf("save(integer) returned %T", enc)
	}
	runBuiltinTests(t, "save", []builtinTestCase{
		{name: "wrong arg count", args: []Object{}, err: "InvalidArgCountError"},
	})
}
