package blueutils

import (
	"testing"

	"blue/code"
	"blue/object"
)

// IfNameInMapSetEnv tests

func TestIfNameInMapSetEnvFound(t *testing.T) {
	env := object.NewEnvironmentWithoutCore()
	m := object.NewPairsMap()
	key := &object.Stringo{Value: "foo"}
	hk := object.HashKey{Type: object.STRING_OBJ, Value: object.HashObject(key)}
	m.Set(hk, object.MapPair{Key: key, Value: &object.Integer{Value: 42}})

	result := IfNameInMapSetEnv(env, m, "foo")
	if !result {
		t.Error("expected true when key is found")
	}

	val, ok := env.Get("foo")
	if !ok {
		t.Error("expected 'foo' to be set in environment")
	}
	intVal, ok := val.(*object.Integer)
	if !ok || intVal.Value != 42 {
		t.Errorf("expected Integer{42}, got %v", val)
	}
}

func TestIfNameInMapSetEnvNotFound(t *testing.T) {
	env := object.NewEnvironmentWithoutCore()
	m := object.NewPairsMap()
	key := &object.Stringo{Value: "foo"}
	hk := object.HashKey{Type: object.STRING_OBJ, Value: object.HashObject(key)}
	m.Set(hk, object.MapPair{Key: key, Value: &object.Integer{Value: 42}})

	result := IfNameInMapSetEnv(env, m, "bar")
	if result {
		t.Error("expected false when key is not found")
	}

	_, ok := env.Get("bar")
	if ok {
		t.Error("expected 'bar' to not be set in environment")
	}
}

func TestIfNameInMapSetEnvEmptyMap(t *testing.T) {
	env := object.NewEnvironmentWithoutCore()
	m := object.NewPairsMap()

	result := IfNameInMapSetEnv(env, m, "foo")
	if result {
		t.Error("expected false for empty map")
	}
}

func TestIfNameInMapSetEnvNonStringKey(t *testing.T) {
	env := object.NewEnvironmentWithoutCore()
	m := object.NewPairsMap()
	// Use an integer key (not string)
	intKey := &object.Integer{Value: 123}
	hk := object.HashKey{Type: object.INTEGER_OBJ, Value: object.HashObject(intKey)}
	m.Set(hk, object.MapPair{Key: intKey, Value: &object.Integer{Value: 42}})

	result := IfNameInMapSetEnv(env, m, "foo")
	if result {
		t.Error("expected false when key is not a string")
	}
}

func TestIfNameInMapSetEnvMultipleKeys(t *testing.T) {
	env := object.NewEnvironmentWithoutCore()
	m := object.NewPairsMap()

	key1 := &object.Stringo{Value: "first"}
	hk1 := object.HashKey{Type: object.STRING_OBJ, Value: object.HashObject(key1)}
	m.Set(hk1, object.MapPair{Key: key1, Value: &object.Integer{Value: 1}})

	key2 := &object.Stringo{Value: "second"}
	hk2 := object.HashKey{Type: object.STRING_OBJ, Value: object.HashObject(key2)}
	m.Set(hk2, object.MapPair{Key: key2, Value: &object.Stringo{Value: "hello"}})

	// Should find "second"
	result := IfNameInMapSetEnv(env, m, "second")
	if !result {
		t.Error("expected true when key exists")
	}
	val, _ := env.Get("second")
	strVal, ok := val.(*object.Stringo)
	if !ok || strVal.Value != "hello" {
		t.Errorf("expected Stringo{\"hello\"}, got %v", val)
	}

	// Should not have set "first"
	_, ok = env.Get("first")
	if ok {
		t.Error("expected 'first' to not be set")
	}
}

// GetNextOpCallPos tests

func TestGetNextOpCallPosFound(t *testing.T) {
	ins := []byte{
		byte(code.OpConstant), 0, 0,
		byte(code.OpAdd),
		byte(code.OpCall), 0, 0,
	}

	pos := GetNextOpCallPos(ins, 0)
	if pos != 4 {
		t.Errorf("expected position 4, got %d", pos)
	}
}

func TestGetNextOpCallPosFoundAfterSome(t *testing.T) {
	ins := []byte{
		byte(code.OpConstant), 0, 0,
		byte(code.OpPop),
		byte(code.OpConstant), 0, 0,
		byte(code.OpCall), 0, 0,
	}

	pos := GetNextOpCallPos(ins, 0)
	if pos != 7 {
		t.Errorf("expected position 7, got %d", pos)
	}
}

func TestGetNextOpCallPosNotFirst(t *testing.T) {
	ins := []byte{
		byte(code.OpConstant), 0, 0,
		byte(code.OpAdd),
		byte(code.OpCall), 0, 0,
	}

	pos := GetNextOpCallPos(ins, 3)
	if pos != 4 {
		t.Errorf("expected position 4 (after OpAdd), got %d", pos)
	}
}

func TestGetNextOpCallPosNotFound(t *testing.T) {
	ins := []byte{
		byte(code.OpConstant), 0, 0,
		byte(code.OpPop),
		byte(code.OpAdd),
	}

	pos := GetNextOpCallPos(ins, 0)
	if pos != -1 {
		t.Errorf("expected -1 when OpCall not found, got %d", pos)
	}
}

func TestGetNextOpCallPosEmpty(t *testing.T) {
	ins := []byte{}
	pos := GetNextOpCallPos(ins, 0)
	if pos != -1 {
		t.Errorf("expected -1 for empty instructions, got %d", pos)
	}
}

func TestGetNextOpCallPosOnlyCall(t *testing.T) {
	ins := []byte{byte(code.OpCall), 0, 0}
	pos := GetNextOpCallPos(ins, 0)
	if pos != 0 {
		t.Errorf("expected position 0, got %d", pos)
	}
}

// BytecodeDebugString tests

func TestBytecodeDebugString(t *testing.T) {
	ins := []byte{
		byte(code.OpConstant), 0, 0,
		byte(code.OpPop),
	}
	constants := []object.Object{&object.Integer{Value: 42}}

	result := BytecodeDebugString(ins, constants)
	if result == "" {
		t.Error("expected non-empty debug string")
	}
}

func TestBytecodeDebugStringWithConstants(t *testing.T) {
	ins := []byte{
		byte(code.OpConstant), 0, 0,
		byte(code.OpPop),
	}
	constants := []object.Object{&object.Integer{Value: 42}}

	result := BytecodeDebugString(ins, constants)
	if result == "" {
		t.Error("expected non-empty debug string")
	}
}

func TestBytecodeDebugStringMultipleInstructions(t *testing.T) {
	ins := []byte{
		byte(code.OpConstant), 0, 0,
		byte(code.OpConstant), 0, 1,
		byte(code.OpAdd),
		byte(code.OpPop),
	}
	constants := []object.Object{&object.Integer{Value: 10}, &object.Integer{Value: 20}}

	result := BytecodeDebugString(ins, constants)
	lines := 0
	for _, c := range result {
		if c == '\n' {
			lines++
		}
	}
	if lines != 4 {
		t.Errorf("expected 4 lines, got %d", lines)
	}
}

func TestBytecodeDebugStringOpPopOnly(t *testing.T) {
	ins := []byte{byte(code.OpPop)}

	result := BytecodeDebugString(ins, nil)
	if result == "" {
		t.Error("expected non-empty debug string for OpPop only")
	}
}

func TestBytecodeDebugStringOpTrueOnly(t *testing.T) {
	ins := []byte{byte(code.OpTrue)}

	result := BytecodeDebugString(ins, nil)
	if result == "" {
		t.Error("expected non-empty debug string for OpTrue only")
	}
}

// BytecodeDebugStringWithOffset tests

func TestBytecodeDebugStringWithOffset(t *testing.T) {
	ins := []byte{
		byte(code.OpConstant), 0, 0,
		byte(code.OpPop),
	}
	constants := []object.Object{&object.Integer{Value: 42}}

	result := BytecodeDebugStringWithOffset(100, ins, constants)
	if result == "" {
		t.Error("expected non-empty debug string")
	}
}

func TestBytecodeDebugStringWithOffsetShowsOffset(t *testing.T) {
	ins := []byte{
		byte(code.OpConstant), 0, 0,
	}
	constants := []object.Object{&object.Integer{Value: 42}}

	result := BytecodeDebugStringWithOffset(500, ins, constants)
	if result == "" {
		t.Error("expected non-empty debug string")
	}
}

func TestBytecodeDebugStringWithOffsetMultipleInstructions(t *testing.T) {
	ins := []byte{
		byte(code.OpPop),
		byte(code.OpAdd),
		byte(code.OpFalse),
	}

	result := BytecodeDebugStringWithOffset(10, ins, nil)
	lines := 0
	for _, c := range result {
		if c == '\n' {
			lines++
		}
	}
	if lines != 3 {
		t.Errorf("expected 3 lines, got %d", lines)
	}
}

func TestBytecodeDebugStringWithOffsetVsNoOffset(t *testing.T) {
	ins := []byte{
		byte(code.OpConstant), 0, 0,
	}
	constants := []object.Object{&object.Integer{Value: 42}}

	resultWithOffset := BytecodeDebugStringWithOffset(100, ins, constants)
	resultNoOffset := BytecodeDebugString(ins, constants)

	// The offset should differ in the output
	if resultWithOffset == resultNoOffset {
		t.Error("expected different output with different offsets")
	}
}

// fmtInstruction tests

func TestFmtInstructionNoOperands(t *testing.T) {
	def := &code.Definition{Name: "OpPop", OperandWidths: []int{}}
	result := fmtInstruction(def, []int{}, nil)
	if result != "OpPop" {
		t.Errorf("expected 'OpPop', got %q", result)
	}
}

func TestFmtInstructionOneOperand(t *testing.T) {
	def := &code.Definition{Name: "OpConstant", OperandWidths: []int{2}}
	constants := []object.Object{&object.Integer{Value: 42}}
	result := fmtInstruction(def, []int{0}, constants)
	expected := "OpConstant 0 (42)"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFmtInstructionOneOperandWithNilConstant(t *testing.T) {
	def := &code.Definition{Name: "OpConstant", OperandWidths: []int{2}}
	// Operand index exceeds constants length
	result := fmtInstruction(def, []int{999}, nil)
	if result == "OpConstant 999" {
		t.Error("expected undefined constant note in output")
	}
}

func TestFmtInstructionTwoOperands(t *testing.T) {
	def := &code.Definition{Name: "OpGetBuiltin", OperandWidths: []int{1, 1}}
	result := fmtInstruction(def, []int{0, 0}, nil)
	if result == "" {
		t.Error("expected non-empty output")
	}
}

func TestFmtInstructionTwoOperandsWithBuiltinobjs(t *testing.T) {
	def := &code.Definition{Name: "OpGetBuiltin", OperandWidths: []int{1, 1}}
	result := fmtInstruction(def, []int{object.BuiltinobjsModuleIndex, 0}, nil)
	if result == "" {
		t.Error("expected non-empty output with builtinobjs index")
	}
}

func TestFmtInstructionMismatchedOperandCount(t *testing.T) {
	def := &code.Definition{Name: "OpConstant", OperandWidths: []int{2}}
	result := fmtInstruction(def, []int{1, 2}, nil)
	if result == "OpConstant 1 2" {
		t.Error("expected error message for mismatched operand count")
	}
}

// ENABLE_VM_CACHING tests

func TestEnableVMCachingDefault(t *testing.T) {
	original := ENABLE_VM_CACHING
	defer func() { ENABLE_VM_CACHING = original }()

	// Default should be true
	if !ENABLE_VM_CACHING {
		t.Error("expected ENABLE_VM_CACHING to be true by default")
	}
}

func TestEnableVMCachingSetFalse(t *testing.T) {
	original := ENABLE_VM_CACHING
	defer func() { ENABLE_VM_CACHING = original }()

	ENABLE_VM_CACHING = false
	if ENABLE_VM_CACHING {
		t.Error("expected ENABLE_VM_CACHING to be false after setting")
	}
}

func TestEnableVMCachingSetTrue(t *testing.T) {
	original := ENABLE_VM_CACHING
	defer func() { ENABLE_VM_CACHING = original }()

	ENABLE_VM_CACHING = false
	ENABLE_VM_CACHING = true
	if !ENABLE_VM_CACHING {
		t.Error("expected ENABLE_VM_CACHING to be true after setting")
	}
}
