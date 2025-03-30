package object

import (
	"blue/ast"
	"blue/token"
	"testing"
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
	a := &ast.Identifier{
		Token: token.Token{},
		Value: "a",
	}
	b := &ast.Identifier{
		Token: token.Token{},
		Value: "b",
	}
	f := &Function{
		Parameters:        []*ast.Identifier{a, b},
		DefaultParameters: []Object{nil, &Null{}},
		Body:              &ast.BlockStatement{},
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
