package vm

import (
	"testing"

	"blue/compiler"
	"blue/lexer"
	"blue/object"
	"blue/parser"
)

func compileClosure(t *testing.T, input string) (*object.Closure, *VM) {
	t.Helper()
	l := lexer.New(input, "<test>")
	p := parser.New(l)
	program := p.ParseProgram()
	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	vm := New(comp.Bytecode())
	if err := vm.Run(); err != nil {
		t.Fatalf("vm error: %s", err)
	}
	cl, ok := vm.LastPoppedStackElem().(*object.Closure)
	if !ok {
		t.Fatalf("expected closure, got %T", vm.LastPoppedStackElem())
	}
	return cl, vm
}

func TestCloneHandlerClosureIndependence(t *testing.T) {
	cl, _ := compileClosure(t, `fun(a, b) { a }`)
	key := object.NameIndexKey{Name: "query_params", Index: 0}
	cl.Fun.SpecialFunctionParameters = map[object.NameIndexKey]map[object.NameIndexKey]object.Object{
		key: {object.NameIndexKey{Name: "x", Index: 0}: &object.Stringo{Value: "orig"}},
	}

	c1 := cloneHandlerClosure(cl)
	if c1 == cl || c1.Fun == cl.Fun {
		t.Fatal("clone did not produce a new closure and compiled function")
	}
	if &c1.Fun.Instructions[0] != &cl.Fun.Instructions[0] {
		t.Fatal("bytecode is immutable and should be shared with the original")
	}
	inner := c1.Fun.SpecialFunctionParameters[key]
	inner[object.NameIndexKey{Name: "x", Index: 0}] = &object.Stringo{Value: "changed"}
	if cl.Fun.SpecialFunctionParameters[key][object.NameIndexKey{Name: "x", Index: 0}].(*object.Stringo).Value != "orig" {
		t.Fatal("mutating clone special params affected the original")
	}
}
