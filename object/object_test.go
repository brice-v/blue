package object

import (
	"blue/ast"
	"blue/token"
	"testing"
)

func TestStringHashKey(t *testing.T) {
	hello1 := &Stringo{Value: "Hello World"}
	hello2 := &Stringo{Value: "Hello World"}
	diff1 := &Stringo{Value: "My name is johnny"}
	diff2 := &Stringo{Value: "My name is johnny"}
	if hello1.HashKey() != hello2.HashKey() {
		t.Errorf("strings with same content have different hash keys")
	}
	if diff1.HashKey() != diff2.HashKey() {
		t.Errorf("strings with same content have different hash keys")
	}
	if hello1.HashKey() == diff1.HashKey() {
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
