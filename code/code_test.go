package code

import (
	"testing"
)

func TestLookupValidOpcode(t *testing.T) {
	def, err := Lookup(byte(OpConstant))
	if err != nil {
		t.Fatalf("Lookup(OpConstant) returned error: %v", err)
	}
	if def.Name != "OpConstant" {
		t.Errorf("Name = %q, want %q", def.Name, "OpConstant")
	}
	if len(def.OperandWidths) != 1 || def.OperandWidths[0] != 2 {
		t.Errorf("OperandWidths = %v, want []int{2}", def.OperandWidths)
	}
}

func TestLookupAllOpcodes(t *testing.T) {
	for op := OpConstant; op <= OpNotInCatch; op++ {
		def, err := Lookup(byte(op))
		if err != nil {
			t.Errorf("Lookup(%d) returned error: %v", op, err)
		}
		if def == nil {
			t.Errorf("Lookup(%d) returned nil definition", op)
		}
		if def.Name == "" {
			t.Errorf("Lookup(%d) returned empty name", op)
		}
	}
}

func TestLookupInvalidOpcode(t *testing.T) {
	_, err := Lookup(255)
	if err == nil {
		t.Fatal("expected error for invalid opcode, got nil")
	}
	expectedMsg := "opcode 255 undefined"
	if err.Error() != expectedMsg {
		t.Errorf("error = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestLookupOpTrue(t *testing.T) {
	def, err := Lookup(byte(OpTrue))
	if err != nil {
		t.Fatalf("Lookup(OpTrue) returned error: %v", err)
	}
	if def.Name != "OpTrue" {
		t.Errorf("Name = %q, want %q", def.Name, "OpTrue")
	}
	if len(def.OperandWidths) != 0 {
		t.Errorf("OperandWidths = %v, want []int{}", def.OperandWidths)
	}
}

func TestLookupOpFalse(t *testing.T) {
	def, err := Lookup(byte(OpFalse))
	if err != nil {
		t.Fatalf("Lookup(OpFalse) returned error: %v", err)
	}
	if def.Name != "OpFalse" {
		t.Errorf("Name = %q, want %q", def.Name, "OpFalse")
	}
}

func TestLookupOpAdd(t *testing.T) {
	def, err := Lookup(byte(OpAdd))
	if err != nil {
		t.Fatalf("Lookup(OpAdd) returned error: %v", err)
	}
	if def.Name != "OpAdd" {
		t.Errorf("Name = %q, want %q", def.Name, "OpAdd")
	}
	if len(def.OperandWidths) != 0 {
		t.Errorf("OperandWidths = %v, want []int{}", def.OperandWidths)
	}
}

func TestLookupOpJumpNotTruthy(t *testing.T) {
	def, err := Lookup(byte(OpJumpNotTruthy))
	if err != nil {
		t.Fatalf("Lookup(OpJumpNotTruthy) returned error: %v", err)
	}
	if def.Name != "OpJumpNotTruthy" {
		t.Errorf("Name = %q, want %q", def.Name, "OpJumpNotTruthy")
	}
	if len(def.OperandWidths) != 1 || def.OperandWidths[0] != 2 {
		t.Errorf("OperandWidths = %v, want []int{2}", def.OperandWidths)
	}
}

func TestLookupOpCall(t *testing.T) {
	def, err := Lookup(byte(OpCall))
	if err != nil {
		t.Fatalf("Lookup(OpCall) returned error: %v", err)
	}
	if def.Name != "OpCall" {
		t.Errorf("Name = %q, want %q", def.Name, "OpCall")
	}
	if len(def.OperandWidths) != 1 || def.OperandWidths[0] != 1 {
		t.Errorf("OperandWidths = %v, want []int{1}", def.OperandWidths)
	}
}

func TestLookupOpClosure(t *testing.T) {
	def, err := Lookup(byte(OpClosure))
	if err != nil {
		t.Fatalf("Lookup(OpClosure) returned error: %v", err)
	}
	if def.Name != "OpClosure" {
		t.Errorf("Name = %q, want %q", def.Name, "OpClosure")
	}
	if len(def.OperandWidths) != 2 || def.OperandWidths[0] != 2 || def.OperandWidths[1] != 1 {
		t.Errorf("OperandWidths = %v, want []int{2, 1}", def.OperandWidths)
	}
}

func TestLookupOpGetBuiltin(t *testing.T) {
	def, err := Lookup(byte(OpGetBuiltin))
	if err != nil {
		t.Fatalf("Lookup(OpGetBuiltin) returned error: %v", err)
	}
	if def.Name != "OpGetBuiltin" {
		t.Errorf("Name = %q, want %q", def.Name, "OpGetBuiltin")
	}
	if len(def.OperandWidths) != 2 || def.OperandWidths[0] != 1 || def.OperandWidths[1] != 1 {
		t.Errorf("OperandWidths = %v, want []int{1, 1}", def.OperandWidths)
	}
}

func TestLookupOpStringInterp(t *testing.T) {
	def, err := Lookup(byte(OpStringInterp))
	if err != nil {
		t.Fatalf("Lookup(OpStringInterp) returned error: %v", err)
	}
	if def.Name != "OpStringInterp" {
		t.Errorf("Name = %q, want %q", def.Name, "OpStringInterp")
	}
	if len(def.OperandWidths) != 2 || def.OperandWidths[0] != 2 || def.OperandWidths[1] != 1 {
		t.Errorf("OperandWidths = %v, want []int{2, 1}", def.OperandWidths)
	}
}

func TestGetOpName(t *testing.T) {
	name := GetOpName(OpConstant)
	if name != "OpConstant" {
		t.Errorf("GetOpName(OpConstant) = %q, want %q", name, "OpConstant")
	}
}

func TestGetOpNameAllOpcodes(t *testing.T) {
	for op := OpConstant; op <= OpNotInCatch; op++ {
		name := GetOpName(op)
		if name == "" {
			t.Errorf("GetOpName(%d) returned empty string", op)
		}
	}
}

func TestGetOpNameInvalidOpcode(t *testing.T) {
	name := GetOpName(Opcode(255))
	if name != "" {
		t.Errorf("GetOpName(255) = %q, want empty string", name)
	}
}

func TestMakeOpConstant(t *testing.T) {
	instr := Make(OpConstant, 42)
	if len(instr) != 3 {
		t.Errorf("len(Make(OpConstant, 42)) = %d, want 3", len(instr))
	}
	if instr[0] != byte(OpConstant) {
		t.Errorf("instr[0] = %d, want %d", instr[0], OpConstant)
	}
}

func TestMakeOpTrue(t *testing.T) {
	instr := Make(OpTrue)
	if len(instr) != 1 {
		t.Errorf("len(Make(OpTrue)) = %d, want 1", len(instr))
	}
	if instr[0] != byte(OpTrue) {
		t.Errorf("instr[0] = %d, want %d", instr[0], OpTrue)
	}
}

func TestMakeOpCall(t *testing.T) {
	instr := Make(OpCall, 3)
	if len(instr) != 2 {
		t.Errorf("len(Make(OpCall, 3)) = %d, want 2", len(instr))
	}
	if instr[0] != byte(OpCall) {
		t.Errorf("instr[0] = %d, want %d", instr[0], OpCall)
	}
	if instr[1] != 3 {
		t.Errorf("instr[1] = %d, want 3", instr[1])
	}
}

func TestMakeOpJumpNotTruthy(t *testing.T) {
	instr := Make(OpJumpNotTruthy, 100)
	if len(instr) != 3 {
		t.Errorf("len(Make(OpJumpNotTruthy, 100)) = %d, want 3", len(instr))
	}
}

func TestMakeOpClosure(t *testing.T) {
	instr := Make(OpClosure, 10, 5)
	if len(instr) != 4 {
		t.Errorf("len(Make(OpClosure, 10, 5)) = %d, want 4", len(instr))
	}
}

func TestMakeOpGetBuiltin(t *testing.T) {
	instr := Make(OpGetBuiltin, 0, 5)
	if len(instr) != 3 {
		t.Errorf("len(Make(OpGetBuiltin, 0, 5)) = %d, want 3", len(instr))
	}
}

func TestMakeOpStringInterp(t *testing.T) {
	instr := Make(OpStringInterp, 10, 3)
	if len(instr) != 4 {
		t.Errorf("len(Make(OpStringInterp, 10, 3)) = %d, want 4", len(instr))
	}
}

func TestMakeInvalidOpcode(t *testing.T) {
	instr := Make(Opcode(255))
	if len(instr) != 0 {
		t.Errorf("Make(255) = %v, want empty slice", instr)
	}
}

func TestReadUint16(t *testing.T) {
	ins := Instructions{0x00, 0x2A}
	result := ReadUint16(ins)
	if result != 42 {
		t.Errorf("ReadUint16([0x00, 0x2A]) = %d, want 42", result)
	}
}

func TestReadUint16Max(t *testing.T) {
	ins := Instructions{0xFF, 0xFF}
	result := ReadUint16(ins)
	if result != 65535 {
		t.Errorf("ReadUint16([0xFF, 0xFF]) = %d, want 65535", result)
	}
}

func TestReadUint16BigEndian(t *testing.T) {
	// BigEndian: 0x01 0x02 = 258
	ins := Instructions{0x01, 0x02}
	result := ReadUint16(ins)
	if result != 258 {
		t.Errorf("ReadUint16([0x01, 0x02]) = %d, want 258", result)
	}
}

func TestReadUint8(t *testing.T) {
	ins := Instructions{0x42}
	result := ReadUint8(ins)
	if result != 66 {
		t.Errorf("ReadUint8([0x42]) = %d, want 66", result)
	}
}

func TestReadUint8Max(t *testing.T) {
	ins := Instructions{0xFF}
	result := ReadUint8(ins)
	if result != 255 {
		t.Errorf("ReadUint8([0xFF]) = %d, want 255", result)
	}
}

func TestReadOperandsSingleOperand(t *testing.T) {
	def := &Definition{Name: "OpConstant", OperandWidths: []int{2}}
	// Pass only the operand bytes (skip opcode)
	operands, read := ReadOperands(def, Instructions{0x00, 0x2A})
	if len(operands) != 1 || operands[0] != 42 {
		t.Errorf("operands = %v, want [42]", operands)
	}
	if read != 2 {
		t.Errorf("read = %d, want 2", read)
	}
}

func TestReadOperandsTwoOperands(t *testing.T) {
	def := &Definition{Name: "OpClosure", OperandWidths: []int{2, 1}}
	// Pass only the operand bytes (skip opcode)
	operands, read := ReadOperands(def, Instructions{0x00, 0x0A, 0x05})
	if len(operands) != 2 {
		t.Errorf("operands = %v, want 2 operands", operands)
	}
	if operands[0] != 10 || operands[1] != 5 {
		t.Errorf("operands = %v, want [10, 5]", operands)
	}
	if read != 3 {
		t.Errorf("read = %d, want 3", read)
	}
}

func TestReadOperandsNoOperands(t *testing.T) {
	def := &Definition{Name: "OpTrue", OperandWidths: []int{}}
	operands, read := ReadOperands(def, Instructions{})
	if len(operands) != 0 {
		t.Errorf("operands = %v, want empty", operands)
	}
	if read != 0 {
		t.Errorf("read = %d, want 0", read)
	}
}

func TestReadOperandsSingleByteOperand(t *testing.T) {
	def := &Definition{Name: "OpCall", OperandWidths: []int{1}}
	// Pass only the operand bytes (skip opcode)
	operands, read := ReadOperands(def, Instructions{0x03})
	if len(operands) != 1 || operands[0] != 3 {
		t.Errorf("operands = %v, want [3]", operands)
	}
	if read != 1 {
		t.Errorf("read = %d, want 1", read)
	}
}

func TestInstructionsString(t *testing.T) {
	instr1 := Make(OpConstant, 42)
	instr2 := Make(OpTrue)
	ins := Instructions(instr1).Concat(Instructions(instr2))
	str := ins.String()
	if str == "" {
		t.Error("String() should not be empty")
	}
}



func TestFmtInstructionNoOperands(t *testing.T) {
	def := &Definition{Name: "OpTrue", OperandWidths: []int{}}
	result := Instructions{}.FmtInstruction(def, nil)
	if result != "OpTrue" {
		t.Errorf("FmtInstruction = %q, want %q", result, "OpTrue")
	}
}

func TestFmtInstructionOneOperand(t *testing.T) {
	def := &Definition{Name: "OpConstant", OperandWidths: []int{2}}
	result := Instructions{}.FmtInstruction(def, []int{42})
	if result != "OpConstant 42" {
		t.Errorf("FmtInstruction = %q, want %q", result, "OpConstant 42")
	}
}

func TestFmtInstructionTwoOperands(t *testing.T) {
	def := &Definition{Name: "OpClosure", OperandWidths: []int{2, 1}}
	result := Instructions{}.FmtInstruction(def, []int{10, 5})
	if result != "OpClosure 10 5" {
		t.Errorf("FmtInstruction = %q, want %q", result, "OpClosure 10 5")
	}
}

func TestFmtInstructionOperandMismatch(t *testing.T) {
	def := &Definition{Name: "OpConstant", OperandWidths: []int{2}}
	result := Instructions{}.FmtInstruction(def, []int{1, 2})
	if result == "" {
		t.Error("FmtInstruction should return error message for operand mismatch")
	}
}

func TestFmtInstructionUnhandledOperandCount(t *testing.T) {
	def := &Definition{Name: "OpUnknown", OperandWidths: []int{}}
	result := Instructions{}.FmtInstruction(def, []int{1, 2, 3})
	if result == "" {
		t.Error("FmtInstruction should return error message for unhandled operand count")
	}
}

func TestMakeOpPop(t *testing.T) {
	instr := Make(OpPop)
	if len(instr) != 1 {
		t.Errorf("len(Make(OpPop)) = %d, want 1", len(instr))
	}
	if instr[0] != byte(OpPop) {
		t.Errorf("instr[0] = %d, want %d", instr[0], OpPop)
	}
}

func TestMakeOpReturnValue(t *testing.T) {
	instr := Make(OpReturnValue)
	if len(instr) != 1 {
		t.Errorf("len(Make(OpReturnValue)) = %d, want 1", len(instr))
	}
}

func TestMakeOpReturn(t *testing.T) {
	instr := Make(OpReturn)
	if len(instr) != 1 {
		t.Errorf("len(Make(OpReturn)) = %d, want 1", len(instr))
	}
}

func TestMakeOpSetGlobal(t *testing.T) {
	instr := Make(OpSetGlobal, 10)
	if len(instr) != 3 {
		t.Errorf("len(Make(OpSetGlobal, 10)) = %d, want 3", len(instr))
	}
}

func TestMakeOpGetGlobal(t *testing.T) {
	instr := Make(OpGetGlobal, 5)
	if len(instr) != 3 {
		t.Errorf("len(Make(OpGetGlobal, 5)) = %d, want 3", len(instr))
	}
}

func TestMakeOpSetLocal(t *testing.T) {
	instr := Make(OpSetLocal, 3)
	if len(instr) != 2 {
		t.Errorf("len(Make(OpSetLocal, 3)) = %d, want 2", len(instr))
	}
}

func TestMakeOpGetLocal(t *testing.T) {
	instr := Make(OpGetLocal, 3)
	if len(instr) != 2 {
		t.Errorf("len(Make(OpGetLocal, 3)) = %d, want 2", len(instr))
	}
}

func TestMakeOpList(t *testing.T) {
	instr := Make(OpList, 5)
	if len(instr) != 3 {
		t.Errorf("len(Make(OpList, 5)) = %d, want 3", len(instr))
	}
}

func TestMakeOpMap(t *testing.T) {
	instr := Make(OpMap, 3)
	if len(instr) != 3 {
		t.Errorf("len(Make(OpMap, 3)) = %d, want 3", len(instr))
	}
}

func TestMakeOpSet(t *testing.T) {
	instr := Make(OpSet, 4)
	if len(instr) != 3 {
		t.Errorf("len(Make(OpSet, 4)) = %d, want 3", len(instr))
	}
}

func TestMakeOpSetCompLiteral(t *testing.T) {
	instr := Make(OpSetCompLiteral)
	if len(instr) != 1 {
		t.Errorf("len(Make(OpSetCompLiteral)) = %d, want 1", len(instr))
	}
}

func TestMakeOpMapCompLiteral(t *testing.T) {
	instr := Make(OpMapCompLiteral)
	if len(instr) != 1 {
		t.Errorf("len(Make(OpMapCompLiteral)) = %d, want 1", len(instr))
	}
}

func TestMakeOpExecString(t *testing.T) {
	instr := Make(OpExecString, 10)
	if len(instr) != 3 {
		t.Errorf("len(Make(OpExecString, 10)) = %d, want 3", len(instr))
	}
}

func TestMakeOpDefaultArgs(t *testing.T) {
	instr := Make(OpDefaultArgs, 5)
	if len(instr) != 3 {
		t.Errorf("len(Make(OpDefaultArgs, 5)) = %d, want 3", len(instr))
	}
}

func TestMakeOpStruct(t *testing.T) {
	instr := Make(OpStruct, 3)
	if len(instr) != 3 {
		t.Errorf("len(Make(OpStruct, 3)) = %d, want 3", len(instr))
	}
}

func TestMakeOpGetListIndex(t *testing.T) {
	instr := Make(OpGetListIndex, 2)
	if len(instr) != 2 {
		t.Errorf("len(Make(OpGetListIndex, 2)) = %d, want 2", len(instr))
	}
}

func TestMakeOpDefer(t *testing.T) {
	instr := Make(OpDefer, 1)
	if len(instr) != 2 {
		t.Errorf("len(Make(OpDefer, 1)) = %d, want 2", len(instr))
	}
}

func TestMakeOpSpawn(t *testing.T) {
	instr := Make(OpSpawn, 2)
	if len(instr) != 2 {
		t.Errorf("len(Make(OpSpawn, 2)) = %d, want 2", len(instr))
	}
}

func TestMakeOpGetGlobalImmOrSpecial(t *testing.T) {
	instr := Make(OpGetGlobalImmOrSpecial, 10, 5)
	if len(instr) != 4 {
		t.Errorf("len(Make(OpGetGlobalImmOrSpecial, 10, 5)) = %d, want 4", len(instr))
	}
}

func TestMakeOpGetFunctionParameterSpecial(t *testing.T) {
	instr := Make(OpGetFunctionParameterSpecial, 3, 1)
	if len(instr) != 3 {
		t.Errorf("len(Make(OpGetFunctionParameterSpecial, 3, 1)) = %d, want 3", len(instr))
	}
}

func TestMakeOpGetFunctionParameterSpecial2(t *testing.T) {
	instr := Make(OpGetFunctionParameterSpecial2, 7)
	if len(instr) != 3 {
		t.Errorf("len(Make(OpGetFunctionParameterSpecial2, 7)) = %d, want 3", len(instr))
	}
}

func TestMakeOpSpecialIndexHelper(t *testing.T) {
	instr := Make(OpSpecialIndexHelper, 5)
	if len(instr) != 3 {
		t.Errorf("len(Make(OpSpecialIndexHelper, 5)) = %d, want 3", len(instr))
	}
}

func TestMakeOpNode(t *testing.T) {
	instr := Make(OpNode, 3)
	if len(instr) != 3 {
		t.Errorf("len(Make(OpNode, 3)) = %d, want 3", len(instr))
	}
}

func TestMakeOpJump(t *testing.T) {
	instr := Make(OpJump, 200)
	if len(instr) != 3 {
		t.Errorf("len(Make(OpJump, 200)) = %d, want 3", len(instr))
	}
}

func TestMakeOpJumpNotTruthyAndPushTrue(t *testing.T) {
	instr := Make(OpJumpNotTruthyAndPushTrue, 50)
	if len(instr) != 3 {
		t.Errorf("len(Make(OpJumpNotTruthyAndPushTrue, 50)) = %d, want 3", len(instr))
	}
}

func TestMakeOpJumpNotTruthyAndPushFalse(t *testing.T) {
	instr := Make(OpJumpNotTruthyAndPushFalse, 50)
	if len(instr) != 3 {
		t.Errorf("len(Make(OpJumpNotTruthyAndPushFalse, 50)) = %d, want 3", len(instr))
	}
}

// Helper method to concatenate Instructions
func (ins Instructions) Concat(other Instructions) Instructions {
	result := make(Instructions, 0, len(ins)+len(other))
	return append(result, ins...)
}
