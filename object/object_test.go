package object

import (
	"fmt"
	"math"
	"slices"
	"testing"

	"blue/ml"
)

func TestStringHashKey(t *testing.T) {
	hello1 := &Stringo{Value: "Hello World"}
	hello2 := &Stringo{Value: "Hello World"}
	hello1hk := HashKey{Type: STRING_OBJ, Value: HashObject(hello1)}
	hello2hk := HashKey{Type: STRING_OBJ, Value: HashObject(hello2)}
	diff1 := &Stringo{Value: "My name is johnny"}
	diff2 := &Stringo{Value: "My name is johnny"}
	diff1hk := HashKey{Type: STRING_OBJ, Value: HashObject(diff1)}
	diff2hk := HashKey{Type: STRING_OBJ, Value: HashObject(diff2)}
	if hello1hk != hello2hk {
		t.Errorf("strings with same content have different hash keys")
	}
	if diff1hk != diff2hk {
		t.Errorf("strings with same content have different hash keys")
	}
	if hello1hk == diff1hk {
		t.Errorf("strings with different content have same hash keys")
	}
}

func TestFunctionString(t *testing.T) {
	f := &Function{
		Parameters:        []string{"a", "b"},
		DefaultParameters: []Object{nil, &Null{}},
		Body:              "",
	}
	expectedInspect := "fun(a, b=null) {\n\n}"
	if f.Inspect() != expectedInspect {
		t.Fatalf("function with default parameters inspect did not match expected. got=%q, want=%q", f.Inspect(), expectedInspect)
	}
}

func TestBlueStruct(t *testing.T) {
	names := []string{"a", "b"}
	values := []Object{&Integer{Value: 123}, &Stringo{Value: "Hello World"}}
	s, err := NewBlueStruct(names, values)
	if err != nil {
		t.Fatalf("Failed to create Blue Struct: %s", err.Error())
	}
	sl, ok := s.(*BlueStruct)
	if !ok {
		t.Fatalf("sl was not a *BlueStruct. got=%T", s)
	}
	v, _ := sl.Get("a")
	i, ok := v.(*Integer)
	if !ok {
		t.Fatalf("field value for name `a` was not an Integer. got=%T", v)
	}
	if i.Value != 123 {
		t.Errorf("Integer Value was not 123, got=%d", i.Value)
	}
	v1, _ := sl.Get("b")
	s1, ok := v1.(*Stringo)
	if !ok {
		t.Fatalf("field value for name `b` was not a String. got=%T", v1)
	}
	if s1.Value != "Hello World" {
		t.Errorf("String Value was not \"Hello World\", got=%s", s1.Value)
	}
	err = sl.SetWithFieldName("a", &Stringo{Value: "abc"})
	if err != nil && err.Error() != "failed to set on struct literal: existing value type = INTEGER, new value type = STRING" {
		t.Fatalf("should receive set error got = %s", err.Error())
	}
	v2, _ := sl.Get("a")
	s2, ok := v2.(*Integer)
	if !ok {
		t.Fatalf("field value for name `a` was not a Integer. got=%T", v2)
	}
	if s2.Value != 123 {
		t.Errorf("Integer Value was not 123, got=%d", s2.Value)
	}
	err = sl.SetWithFieldName("b", &Stringo{Value: "abc"})
	if err != nil {
		t.Fatalf("set should succeed here for `b` but got error: %s", err.Error())
	}
	v3, _ := sl.Get("b")
	s3, ok := v3.(*Stringo)
	if !ok {
		t.Fatalf("field value for name `b` was not a String. got=%T", v3)
	}
	if s3.Value != "abc" {
		t.Errorf("String Value was not abc, got=%s", s3.Value)
	}
}

func TestNewBlueStruct(t *testing.T) {
	for count := range 14 {
		t.Run(fmt.Sprintf("%d-fields", count), func(t *testing.T) {
			names := make([]string, count)
			values := make([]Object, count)
			list := &List{Elements: []Object{&Integer{Value: 1}}}
			for i := range count {
				names[i] = fmt.Sprintf("f%d", i)
				values[i] = &Integer{Value: int64(i)}
			}
			if count > 0 {
				values[0] = list
			}

			obj, err := NewBlueStruct(names, values)
			if err != nil {
				t.Fatalf("NewBlueStruct() error = %v", err)
			}
			bs := obj.(*BlueStruct)
			if len(bs.Values) != count {
				t.Fatalf("len(Values) = %d, want %d", len(bs.Values), count)
			}
			if count == 0 {
				return
			}
			if bs.Values[0] != Object(list) {
				t.Fatal("struct should reference the field value instead of cloning it")
			}

			original := bs.Values[0]
			values[0] = &Integer{Value: -1}
			if bs.Values[0] != original {
				t.Fatal("struct Values must not alias the input slice")
			}
		})
	}
}

func TestNewBlueStructLengthMismatch(t *testing.T) {
	_, err := NewBlueStruct([]string{"a"}, nil)
	if err == nil {
		t.Fatal("expected error for mismatched names and values")
	}
}

func TestHashObject(t *testing.T) {
	o := &Float{Value: 0.5}
	ho := HashObject(o)
	o1 := &Integer{Value: 0}
	ho1 := HashObject(o1)
	if ho == ho1 {
		t.Errorf("These should never be equal float hash = %d, integer hash = %d", ho, ho1)
	}
}

func TestTensorHashing(t *testing.T) {
	mk := func(vals []float32) *Tensor {
		tt, err := ml.NewTensor(vals, []int{2, 2}, ml.Float32, ml.CPU)
		if err != nil {
			t.Fatalf("NewTensor() error: %v", err)
		}
		return &Tensor{T: tt}
	}

	a := mk([]float32{1, 2, 3, 4})
	b := mk([]float32{1, 2, 3, 4})
	c := mk([]float32{1, 2, 3, 5})

	if HashObject(a) != HashObject(b) {
		t.Error("equal tensors must hash equal")
	}
	if HashObject(a) == HashObject(c) {
		t.Error("tensors with different values must not hash equal")
	}
	if !IsHashable(a) {
		t.Error("tensors must be usable as Map/Set keys")
	}

	// A transposed view has logical values [1,3,2,4], so it must hash like a
	// packed tensor with those same values. This fails if the raw backing slice
	// is hashed instead of the logical elements.
	tr, err := ml.Transpose(a.T, 0, 1)
	if err != nil {
		t.Fatalf("Transpose() error: %v", err)
	}
	if HashObject(&Tensor{T: tr}) != HashObject(mk([]float32{1, 3, 2, 4})) {
		t.Error("a view must hash like its packed logical values")
	}
}

func TestFloatInspectFormatting(t *testing.T) {
	tests := []struct {
		value float64
		want  string
	}{
		{5.0, "5.0"},
		{-7.0, "-7.0"},
		{0.5, "0.5"},
		{2.25, "2.25"},
		{100000.0, "100000.0"},
		// exponent forms are left alone since they already carry no
		// ambiguity with integers
		{1e8, "1e+08"},
		{math.Inf(1), "+Inf"},
		{math.Inf(-1), "-Inf"},
		{math.NaN(), "NaN"},
	}
	for _, tt := range tests {
		f := &Float{Value: tt.value}
		if got := f.Inspect(); got != tt.want {
			t.Errorf("Float(%v).Inspect() = %q, want %q", tt.value, got, tt.want)
		}
	}
}

func intListValues(t *testing.T, o Object) []int64 {
	t.Helper()
	l, ok := o.(*List)
	if !ok {
		t.Fatalf("got %T, want *List", o)
	}
	vals := make([]int64, len(l.Elements))
	for i, e := range l.Elements {
		n, ok := e.(*Integer)
		if !ok {
			t.Fatalf("element %d is %T, want *Integer", i, e)
		}
		vals[i] = n.Value
	}
	return vals
}

func TestTensorSize(t *testing.T) {
	tests := []struct {
		name  string
		data  []float32
		shape []int
		want  []int64
	}{
		{name: "matrix", data: []float32{1, 2, 3, 4}, shape: []int{2, 2}, want: []int64{2, 2}},
		{name: "scalar", data: []float32{7}, shape: []int{}, want: []int64{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tObj := &Tensor{T: newTensorForTest(t, tt.data, tt.shape)}

			prop, err := tObj.Get("size")
			if err != nil {
				t.Fatalf("Get(size): %v", err)
			}
			fn, ok := prop.(*Builtin)
			if !ok {
				t.Fatalf("size property is %T, want *Builtin", prop)
			}
			if got := intListValues(t, fn.Fun()); !slices.Equal(got, tt.want) {
				t.Fatalf("size() = %v, want %v", got, tt.want)
			}
			if out := fn.Fun(NewInteger(0)); out.Type() != ERROR_OBJ {
				t.Fatalf("size(0) returned %T, want an error", out)
			}

			shape, err := tObj.Get("shape")
			if err != nil {
				t.Fatalf("Get(shape): %v", err)
			}
			if got := intListValues(t, shape); !slices.Equal(got, tt.want) {
				t.Fatalf("shape = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringHashStableAcrossInstances(t *testing.T) {
	want := HashObject(&Stringo{Value: "hello"})
	if got := HashObject(NewString("hello")); got != want {
		t.Fatalf("NewString hash = %d, want %d", got, want)
	}
	if got := HashObject(&Stringo{Value: "hello"}); got != want {
		t.Fatalf("second literal hash = %d, want %d", got, want)
	}
	if got := stringHash("hello"); got != want {
		t.Fatalf("stringHash = %d, want %d", got, want)
	}
}

func TestStringHashDiffers(t *testing.T) {
	seen := map[uint64]string{}
	for _, s := range []string{"", "a", "b", "ab", "ba", "hello", "world", "user_1", "user_2"} {
		h := HashObject(NewString(s))
		if other, ok := seen[h]; ok {
			t.Fatalf("hash collision between %q and %q", s, other)
		}
		seen[h] = s
	}
}

func TestStringHashDiffersFromOtherTypes(t *testing.T) {
	if HashObject(NewString("1")) == HashObject(&Integer{Value: 1}) {
		t.Fatal("string and integer hash should differ")
	}
}

func TestHashObjectBooleanNull(t *testing.T) {
	if got := HashObject(TRUE); got != 1 {
		t.Fatalf("TRUE hash = %d, want 1", got)
	}
	if got := HashObject(FALSE); got != 0 {
		t.Fatalf("FALSE hash = %d, want 0", got)
	}
	if got := HashObject(NULL); got != 2 {
		t.Fatalf("NULL hash = %d, want 2", got)
	}
}
