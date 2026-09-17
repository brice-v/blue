package object

import (
	"math"
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
	if IsHashable(a) {
		t.Error("tensors must not be usable as Map/Set keys")
	}

	// A transposed view has logical values [1,3,2,4], so it must hash like a
	// packed tensor with those same values. This fails if the raw backing slice
	// is hashed instead of the logical elements.
	tr, err := ml.DefaultBackend.Transpose(a.T, 0, 1)
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
