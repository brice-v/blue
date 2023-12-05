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
