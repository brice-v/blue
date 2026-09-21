package vm

import (
	"strings"
	"testing"

	"blue/compiler"
	"blue/object"
)

// TestCallBuiltinConvertsPanicToError proves the builtin boundary turns a native
// panic into a blue error (so try/catch or the normal error path handles it)
// and leaves the value stack in a consistent state.
func TestCallBuiltinConvertsPanicToError(t *testing.T) {
	comp := compiler.New()
	if err := comp.Compile(parse("")); err != nil {
		t.Fatalf("compile empty program: %s", err)
	}
	vm := New(comp.Bytecode())

	builtin := &object.Builtin{
		Name: "boom",
		Fun: func(args ...object.Object) object.Object {
			panic("kaboom")
		},
	}
	// Mirror a call site: the callee sits below its arguments.
	if err := vm.push(builtin); err != nil {
		t.Fatal(err)
	}
	if err := vm.push(&object.Integer{Value: 1}); err != nil {
		t.Fatal(err)
	}

	err := vm.callBuiltin(builtin, 1)
	if err == nil {
		t.Fatal("expected the panic to surface as an error")
	}
	if !strings.Contains(err.Error(), "kaboom") {
		t.Fatalf("error does not mention the panic: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("error does not name the builtin: %s", err.Error())
	}
	if vm.sp != 0 {
		t.Fatalf("value stack not cleaned up after panic, sp=%d", vm.sp)
	}
}
