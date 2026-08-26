package object

import (
	"math"
	"math/big"
	"regexp"
	"strings"
	"testing"

	"blue/code"

	"github.com/fxamacker/cbor/v2"
	"github.com/shopspring/decimal"
)

func encodeViaEncode(t *testing.T, obj Object) []byte {
	t.Helper()
	data, err := obj.Encode()
	if err != nil {
		t.Fatalf("Encode() for %T failed: %v", obj, err)
	}
	return data
}

func decodeOrFail(t *testing.T, data []byte) Object {
	t.Helper()
	obj, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}
	return obj
}

func roundTrip(t *testing.T, obj Object) Object {
	t.Helper()
	return decodeOrFail(t, encodeViaEncode(t, obj))
}

func marshalRoundTrip(t *testing.T, obj Object) Object {
	t.Helper()
	wrapper, err := marshalObject(obj)
	if err != nil {
		t.Fatalf("marshalObject(%T) failed: %v", obj, err)
	}
	data, err := cbor.Marshal(wrapper)
	if err != nil {
		t.Fatalf("cbor.Marshal(wrapper) failed: %v", err)
	}
	return decodeOrFail(t, data)
}

func TestEncodingRoundTripPrimitives(t *testing.T) {
	cases := []struct {
		name string
		obj  Object
		want string
	}{
		{"integer small", NewInteger(42), "42"},
		{"integer negative", &Integer{Value: -987654321}, "-987654321"},
		{"integer large", &Integer{Value: math.MaxInt64}, "9223372036854775807"},
		{"big integer", &BigInteger{Value: func() *big.Int {
			v, _ := new(big.Int).SetString("123456789012345678901234567890123456789", 10)
			return v
		}()}, "123456789012345678901234567890123456789"},
		{"boolean true", TRUE, "true"},
		{"boolean false", FALSE, "false"},
		{"null", NULL, "null"},
		{"uinteger", &UInteger{Value: 18446744073709551615}, "18446744073709551615"},
		{"float", &Float{Value: 3.14159}, "3.14159"},
		{"float negative", &Float{Value: -0.5}, "-0.5"},
		{"float whole", &Float{Value: 2}, "2.0"},
		{"string empty", NewString(""), ""},
		{"string ascii", NewString("hello world"), "hello world"},
		{"string unicode", NewString("héllo wörld ✓ 日本語"), "héllo wörld ✓ 日本語"},
		{"bytes", &Bytes{Value: []byte{0x00, 0xff, 0xfe, 0x42}}, "[]byte{0x0, 0xff, 0xfe, 0x42}"},
		{"regex", &Regex{Value: regexp.MustCompile(`^[a-z]+(\d{2,})?$`)}, "/^[a-z]+(\\d{2,})?$/"},
		{"exec string", &ExecString{Value: "println(42)"}, "println(42)"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := roundTrip(t, c.obj)
			if got.Type() != c.obj.Type() {
				t.Errorf("Type() = %s, want %s", got.Type(), c.obj.Type())
			}
			if got.Inspect() != c.want {
				t.Errorf("Inspect() = %q, want %q", got.Inspect(), c.want)
			}
		})
	}
}

func TestEncodingFloatSpecialValues(t *testing.T) {
	cases := []struct {
		name string
		val  float64
		want string
	}{
		{"pos inf", math.Inf(1), "+Inf"},
		{"neg inf", math.Inf(-1), "-Inf"},
		{"nan", math.NaN(), "NaN"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := roundTrip(t, &Float{Value: c.val})
			if got.Inspect() != c.want {
				t.Errorf("Inspect() = %q, want %q", got.Inspect(), c.want)
			}
		})
	}
}

func TestEncodingBigFloat(t *testing.T) {
	bf := BigFloat{Value: decimal.RequireFromString("123456789.987654321")}
	got := roundTrip(t, bf)
	if got.Type() != BIG_FLOAT_OBJ {
		t.Fatalf("Type() = %s, want %s", got.Type(), BIG_FLOAT_OBJ)
	}
	gotBf := got.(*BigFloat)
	if gotBf.Value.String() != "123456789.987654321" {
		t.Errorf("value = %q, want %q (round trip lost precision?)", gotBf.Value.String(), "123456789.987654321")
	}
}

func TestEncodingList(t *testing.T) {
	inner := &Map{Pairs: NewPairsMapWithSize(1)}
	hk := HashKey{Type: STRING_OBJ, Value: HashObject(NewString("a"))}
	inner.Pairs.Set(hk, MapPair{Key: NewString("a"), Value: TRUE})

	lst := &List{Elements: []Object{
		NewInteger(1),
		NewString("two"),
		&List{Elements: []Object{&Float{Value: 3.5}, NULL}},
		inner,
	}}
	got := roundTrip(t, lst)
	gotList := got.(*List)
	if len(gotList.Elements) != 4 {
		t.Fatalf("len(Elements) = %d, want 4", len(gotList.Elements))
	}
	if gotList.Elements[0].(*Integer).Value != 1 {
		t.Errorf("Elements[0] = %s", gotList.Elements[0].Inspect())
	}
	if gotList.Elements[1].Inspect() != "two" {
		t.Errorf("Elements[1] = %s", gotList.Elements[1].Inspect())
	}
	nested := gotList.Elements[2].(*List)
	if len(nested.Elements) != 2 || nested.Elements[0].Inspect() != "3.5" || nested.Elements[1].Type() != NULL_OBJ {
		t.Errorf("nested list = %s", nested.Inspect())
	}
	m := gotList.Elements[3].(*Map)
	pair, ok := m.Pairs.Get(HashKey{Type: STRING_OBJ, Value: HashObject(NewString("a"))})
	if !ok || pair.Key.Inspect() != "a" || pair.Value.Inspect() != "true" {
		t.Errorf("decoded map = %s", m.Inspect())
	}
}

func TestEncodingEmptyContainers(t *testing.T) {
	cases := []Object{
		&List{Elements: []Object{}},
		&Set{Elements: NewSetElements()},
		&Map{Pairs: NewPairsMap()},
	}
	for _, obj := range cases {
		got := roundTrip(t, obj)
		if got.Type() != obj.Type() {
			t.Errorf("%T: Type() = %s, want %s", obj, got.Type(), obj.Type())
		}
		switch g := got.(type) {
		case *List:
			if len(g.Elements) != 0 {
				t.Errorf("list not empty: %s", g.Inspect())
			}
		case *Set:
			if g.Elements.Len() != 0 {
				t.Errorf("set not empty: %s", g.Inspect())
			}
		case *Map:
			if g.Pairs.Len() != 0 {
				t.Errorf("map not empty: %s", g.Inspect())
			}
		}
	}
}

func makeTestSet(elems ...Object) *Set {
	s := &Set{Elements: NewSetElementsWithSize(len(elems))}
	for _, e := range elems {
		s.Elements.Set(HashObject(e), SetPair{Value: e, Present: struct{}{}})
	}
	return s
}

func TestEncodingSet(t *testing.T) {
	set := makeTestSet(NewInteger(1), NewInteger(2), NewString("x"))
	set.Elements.Set(HashObject(TRUE), SetPair{Value: TRUE, Present: struct{}{}})

	got := roundTrip(t, set).(*Set)
	if got.Elements.Len() != 4 {
		t.Fatalf("Len() = %d, want 4 (%s)", got.Elements.Len(), got.Inspect())
	}
	for _, want := range []Object{NewInteger(1), NewInteger(2)} {
		found := false
		for _, k := range got.Elements.Keys {
			pair, _ := got.Elements.Get(k)
			if pair.Value.Inspect() == want.Inspect() {
				found = true
			}
		}
		if !found {
			t.Errorf("set missing element %s (%s)", want.Inspect(), got.Inspect())
		}
	}
}

func makeTestMap(keys []Object, vals []Object) *Map {
	m := &Map{Pairs: NewPairsMapWithSize(len(keys))}
	for i, k := range keys {
		hk := HashKey{Type: k.Type(), Value: HashObject(k)}
		m.Pairs.Set(hk, MapPair{Key: k, Value: vals[i]})
	}
	return m
}

func TestEncodingMap(t *testing.T) {
	m := makeTestMap(
		[]Object{NewString("name"), NewInteger(7)},
		[]Object{NewString("blue"), &List{Elements: []Object{NewInteger(1), NewInteger(2)}}},
	)
	got := roundTrip(t, m).(*Map)
	if got.Pairs.Len() != 2 {
		t.Fatalf("Len() = %d, want 2 (%s)", got.Pairs.Len(), got.Inspect())
	}
	strPair, ok := got.Pairs.Get(HashKey{Type: STRING_OBJ, Value: HashObject(NewString("name"))})
	if !ok || strPair.Value.Inspect() != "blue" {
		t.Errorf("map[name] = %v", strPair)
	}
	intPair, ok := got.Pairs.Get(HashKey{Type: INTEGER_OBJ, Value: HashObject(NewInteger(7))})
	if !ok || intPair.Value.Inspect() != "[1, 2]" {
		t.Errorf("map[7] = %v", intPair)
	}
}

func TestEncodingFunctionBecomesStringFunction(t *testing.T) {
	fn := &Function{
		Parameters:        []string{"a", "b"},
		DefaultParameters: []Object{nil, NewInteger(5)},
		Body:              "a + b",
	}
	want := fn.Inspect()
	got := roundTrip(t, fn)
	sf, ok := got.(*StringFunction)
	if !ok {
		t.Fatalf("decoded type = %T, want *StringFunction", got)
	}
	if sf.Value != want {
		t.Errorf("Value = %q, want %q", sf.Value, want)
	}
	if sf.Inspect() != want {
		t.Errorf("Inspect() = %q, want %q", sf.Inspect(), want)
	}
}

func makeTestCompiledFunction() *CompiledFunction {
	return &CompiledFunction{
		Instructions:             code.Make(code.OpConstant, 1),
		NumLocals:                2,
		NumParameters:            1,
		Parameters:               []string{"x"},
		ParameterHasDefault:      []bool{false},
		NumDefaultParams:         0,
		DisplayString:            "fn x()",
		HelpStr:                  "help text",
		SpecialFunctionParameters: map[NameIndexKey]map[NameIndexKey]Object{
			{Name: "defaults", Index: 1}: {
				{Name: "dv", Index: 0}: NewString("defval"),
			},
		},
	}
}

func TestEncodingCompiledFunction(t *testing.T) {
	cf := makeTestCompiledFunction()
	got := roundTrip(t, cf).(*CompiledFunction)

	if string(got.Instructions) != string(cf.Instructions) {
		t.Errorf("Instructions = %v, want %v", got.Instructions, cf.Instructions)
	}
	if got.NumLocals != cf.NumLocals || got.NumParameters != cf.NumParameters {
		t.Errorf("NumLocals/NumParameters = %d/%d, want %d/%d", got.NumLocals, got.NumParameters, cf.NumLocals, cf.NumParameters)
	}
	if len(got.Parameters) != 1 || got.Parameters[0] != "x" {
		t.Errorf("Parameters = %v", got.Parameters)
	}
	if len(got.ParameterHasDefault) != 1 || got.ParameterHasDefault[0] != false {
		t.Errorf("ParameterHasDefault = %v", got.ParameterHasDefault)
	}
	if got.DisplayString != cf.DisplayString || got.HelpStr != cf.HelpStr {
		t.Errorf("DisplayString/HelpStr = %q/%q, want %q/%q", got.DisplayString, got.HelpStr, cf.DisplayString, cf.HelpStr)
	}
	group, ok := got.SpecialFunctionParameters[NameIndexKey{Name: "defaults", Index: 1}]
	if !ok {
		t.Fatalf("SFP missing group defaults:1 (%#v)", got.SpecialFunctionParameters)
	}
	val, ok := group[NameIndexKey{Name: "dv", Index: 0}]
	if !ok || val.Inspect() != "defval" {
		t.Errorf("SFP inner entry = %v", val)
	}
}

func TestEncodingBlueStruct(t *testing.T) {
	bs := &BlueStruct{
		Fields: []string{"name", "age"},
		Values: []Object{NewString("brice"), NewInteger(30)},
	}
	got := roundTrip(t, bs).(*BlueStruct)
	if len(got.Fields) != 2 || got.Fields[0] != "name" || got.Fields[1] != "age" {
		t.Errorf("Fields = %v", got.Fields)
	}
	if got.Values[0].Inspect() != "brice" || got.Values[1].Inspect() != "30" {
		t.Errorf("Values = [%s, %s]", got.Values[0].Inspect(), got.Values[1].Inspect())
	}
}

func TestEncodingStructFieldsGoObj(t *testing.T) {
	fields := NewGoObj([]string{"x", "y"})
	got := marshalRoundTrip(t, fields)
	gotFields, ok := got.(*GoObj[[]string])
	if !ok {
		t.Fatalf("decoded type = %T, want *GoObj[[]string]", got)
	}
	if len(gotFields.Value) != 2 || gotFields.Value[0] != "x" || gotFields.Value[1] != "y" {
		t.Errorf("Value = %#v", gotFields.Value)
	}
}

func TestEncodingDefaultArgs(t *testing.T) {
	da := &DefaultArgs{Value: map[string]Object{"x": NewInteger(7)}}
	got := roundTrip(t, da).(*DefaultArgs)
	v, ok := got.Value["x"]
	if !ok || v.Inspect() != "7" {
		t.Errorf("Value = %#v", got.Value)
	}
}

func TestEncodingSentinelObjects(t *testing.T) {
	if got := roundTrip(t, VM_IGNORE); got != VM_IGNORE {
		t.Errorf("Ignore decoded to different object: %T", got)
	}
	if got := marshalRoundTrip(t, BREAK); got != BREAK {
		t.Errorf("Break decoded to different object: %T", got)
	}
	if got := marshalRoundTrip(t, CONTINUE); got != CONTINUE {
		t.Errorf("Continue decoded to different object: %T", got)
	}

	for _, obj := range []Object{BREAK, CONTINUE} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%T.Encode() should panic", obj)
				}
			}()
			_, _ = obj.Encode()
		}()
	}
}

func TestEncodingModuleConstant(t *testing.T) {
	mod := &Module{Name: "mymod", HelpStr: "does things"}

	if _, err := mod.Encode(); err == nil {
		t.Error("Module.Encode() should refuse live modules")
	}

	got := marshalRoundTrip(t, mod).(*Module)
	if got.Name != "mymod" || got.HelpStr != "does things" {
		t.Errorf("Name/HelpStr = %q/%q", got.Name, got.HelpStr)
	}
	if got.Env != nil {
		t.Error("decoded module Env should be nil")
	}
	if got.Inspect() != "Module 'mymod'" {
		t.Errorf("Inspect() = %q", got.Inspect())
	}
}

func TestEncodingRejections(t *testing.T) {
	t.Run("marshal nil", func(t *testing.T) {
		if _, err := marshalObject(nil); err == nil {
			t.Error("marshalObject(nil) should error")
		}
	})
	t.Run("live GoObj rejected by marshalObjectDepth", func(t *testing.T) {
		if _, err := marshalObject(NewGoObj(42)); err == nil {
			t.Error("expected live GoObj to be unserializable")
		} else if !strings.Contains(err.Error(), "live Go state") {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("GoObj Encode refuses", func(t *testing.T) {
		if _, err := NewGoObj(42).Encode(); err == nil {
			t.Error("GoObj.Encode() should error")
		}
	})
	t.Run("Closure unsupported", func(t *testing.T) {
		if _, err := marshalObject(&Closure{Fun: makeTestCompiledFunction()}); err == nil {
			t.Error("closure should not serialize")
		} else if !strings.Contains(err.Error(), "TODO") {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("Closure Encode refuses", func(t *testing.T) {
		if _, err := (&Closure{}).Encode(); err == nil {
			t.Error("Closure.Encode() should error")
		}
	})
	t.Run("encode depth limit", func(t *testing.T) {
		var cur Object = NewInteger(1)
		for i := 0; i < maxSerializeDepth+10; i++ {
			cur = &List{Elements: []Object{cur}}
		}
		if _, err := marshalObject(cur); err != errTooDeep {
			t.Errorf("err = %v, want errTooDeep", err)
		}
	})
}

func TestEncodingPanicOnlyTypes(t *testing.T) {
	cases := []Object{
		&Error{Message: "boom"},
		&ListCompLiteral{},
		&MapCompLiteral{},
		&SetCompLiteral{},
		&Builtin{Name: "b"},
		&BuiltinObj{},
		&Process{Ch: make(chan Object)},
		&StringFunction{Value: "s"},
		&ReturnValue{Value: NewInteger(1)},
		&GoObjectGob{T: "t"},
	}
	for _, obj := range cases {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%T.Encode() should panic", obj)
				}
			}()
			_, _ = obj.Encode()
		}()
	}
}

func TestDecodeTypeMismatchErrors(t *testing.T) {
	wrongShapeString := cborOf("a plain string")
	wrongShapeInt := cborOf(int64(7))

	cases := []struct {
		name string
		typ  iType
		data []byte
	}{
		{"integer", i_INTEGER_OBJ, wrongShapeInt},
		{"big integer", i_BIG_INTEGER_OBJ, wrongShapeInt},
		{"float", i_FLOAT_OBJ, wrongShapeInt},
		{"big float", i_BIG_FLOAT_OBJ, wrongShapeInt},
		{"boolean", i_BOOLEAN_OBJ, wrongShapeInt},
		{"uinteger", i_UINTEGER_OBJ, wrongShapeInt},
		{"null", i_NULL_OBJ, wrongShapeInt},
		{"string", i_STRING_OBJ, wrongShapeInt},
		{"regex", i_REGEX_OBJ, wrongShapeInt},
		{"bytes", i_BYTES_OBJ, wrongShapeInt},
		{"list", i_LIST_OBJ, wrongShapeString},
		{"set", i_SET_OBJ, wrongShapeString},
		{"map", i_MAP_OBJ, wrongShapeString},
		{"function", i_FUNCTION_OBJ, wrongShapeInt},
		{"go obj", i_GO_OBJ, wrongShapeString},
		{"struct fields", i_STRUCT_FIELDS_OBJ, wrongShapeInt},
		{"compiled function", i_COMPILED_FUNCTION_OBJ, wrongShapeString},
		{"module", i_MODULE_OBJ, wrongShapeInt},
		{"blue struct", i_BLUE_STRUCT_OBJ, wrongShapeString},
		{"exec string", i_EXEC_STRING_OBJ, wrongShapeInt},
		{"default args", i_DEFAULT_ARGS_OBJ, wrongShapeString},
		{"closure unsupported", i_CLOSURE_OBJ, wrongShapeString},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := decodeFromType(c.typ, c.data, 0); err == nil {
				t.Errorf("decodeFromType(%s, wrong shape) should fail", c.name)
			}
		})
	}
}

func cborOf(v any) []byte {
	b, err := cbor.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func TestDecodingErrors(t *testing.T) {
	t.Run("garbage bytes", func(t *testing.T) {
		if _, err := Decode([]byte("this is not cbor")); err == nil {
			t.Error("Decode of garbage should fail")
		}
	})
	t.Run("unknown type tag", func(t *testing.T) {
		data, err := cbor.Marshal(ObjectWrapper{Type: iType(200)})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := Decode(data); err == nil || !strings.Contains(err.Error(), "handle 200") {
			t.Errorf("err = %v, want unknown-type error", err)
		}
	})
	t.Run("decode depth limit", func(t *testing.T) {
		data, err := cbor.Marshal(NewInteger(1))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := decodeFromType(i_INTEGER_OBJ, data, maxSerializeDepth+1); err != errTooDeep {
			t.Errorf("err = %v, want errTooDeep", err)
		}
	})
}

func TestDecodeGoObjectGobForm(t *testing.T) {
	sw, err := marshalObject(NewString("someGoType"))
	if err != nil {
		t.Fatal(err)
	}
	bw, err := marshalObject(&Bytes{Value: []byte{0x01, 0x02}})
	if err != nil {
		t.Fatal(err)
	}
	innerData, err := cbor.Marshal([]ObjectWrapper{sw, bw})
	if err != nil {
		t.Fatal(err)
	}
	full, err := cbor.Marshal(ObjectWrapper{Type: i_GO_OBJ, Data: innerData})
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(full)
	if err != nil {
		t.Fatal(err)
	}
	gob, ok := got.(*GoObjectGob)
	if !ok {
		t.Fatalf("type = %T, want *GoObjectGob", got)
	}
	if gob.T != "someGoType" || string(gob.Value) != "\x01\x02" {
		t.Errorf("T/Value = %q/%v", gob.T, gob.Value)
	}
}
