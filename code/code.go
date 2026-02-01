package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instructions []byte
type Opcode byte

func (ins Instructions) String() string {
	var out bytes.Buffer
	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}
		operands, read := ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d %s\n", i, ins.FmtInstruction(def, operands))
		i += 1 + read
	}
	return out.String()
}

func (ins Instructions) FmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)
	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n",
			len(operands), operandCount)
	}
	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	case 2:
		return fmt.Sprintf("%s %d %d", def.Name, operands[0], operands[1])
	}
	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

const (
	OpConstant Opcode = iota
	OpTrue
	OpFalse
	OpNull
	OpAdd
	OpPop
	OpMinus
	OpStar
	OpPow
	OpDiv
	OpFlDiv
	OpPercent
	OpCarat
	OpAmpersand
	OpPipe
	OpIn
	OpNotin
	OpRange
	OpNonIncRange
	OpLshift
	OpRshift
	OpEqual
	OpNotEqual
	OpGreaterThan
	OpGreaterThanOrEqual
	OpOr
	OpAnd
	OpNeg
	OpNot
	OpTilde
	OpLshiftPre
	OpRshiftPost
	OpJumpNotTruthy
	OpJump
	OpSetGlobal
	OpSetGlobalImm
	OpGetGlobal
	OpGetGlobalImm
	OpList
	OpMap
	OpSet
	OpIndex
	OpCall
	OpReturnValue
	OpReturn
	OpSetLocal
	OpSetLocalImm
	OpGetLocal
	OpGetLocalImm
	OpClosure
	OpGetFree
	OpGetFreeImm
	OpGetBuiltin
	OpStringInterp
	OpIndexSet
	OpTry
	OpCatch
	OpCatchEnd
	OpFinally
	OpFinallyEnd
	OpListCompLiteral
	OpSetCompLiteral
	OpMapCompLiteral
	OpExecString
	OpMatchValue
	OpMatchAny
	OpEval
	OpDefaultArgs
	OpCoreCompiled
	OpSlice
)

type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[Opcode]*Definition{
	OpConstant:           {"OpConstant", []int{2}},
	OpTrue:               {"OpTrue", []int{}},
	OpFalse:              {"OpFalse", []int{}},
	OpNull:               {"OpNull", []int{}},
	OpAdd:                {"OpAdd", []int{}},
	OpPop:                {"OpPop", []int{}},
	OpMinus:              {"OpMinus", []int{}},
	OpStar:               {"OpStar", []int{}},
	OpPow:                {"OpPow", []int{}},
	OpDiv:                {"OpDiv", []int{}},
	OpFlDiv:              {"OpFlDiv", []int{}},
	OpPercent:            {"OpPercent", []int{}},
	OpCarat:              {"OpCarat", []int{}},
	OpAmpersand:          {"OpAmpersand", []int{}},
	OpPipe:               {"OpPipe", []int{}},
	OpIn:                 {"OpIn", []int{}},
	OpNotin:              {"OpNotin", []int{}},
	OpRange:              {"OpRange", []int{}},
	OpNonIncRange:        {"OpNonIncRange", []int{}},
	OpLshift:             {"OpLshift", []int{}},
	OpRshift:             {"OpRshift", []int{}},
	OpEqual:              {"OpEqual", []int{}},
	OpNotEqual:           {"OpNotEqual", []int{}},
	OpGreaterThan:        {"OpGreaterThan", []int{}},
	OpGreaterThanOrEqual: {"OpGreaterThanOrEqual", []int{}},
	OpAnd:                {"OpAnd", []int{}},
	OpOr:                 {"OpOr", []int{}},
	OpNeg:                {"OpNeg", []int{}},
	OpNot:                {"OpNot", []int{}},
	OpTilde:              {"OpTilde", []int{}},
	OpLshiftPre:          {"OpLshiftPre", []int{}},
	OpRshiftPost:         {"OpRshiftPost", []int{}},
	OpJumpNotTruthy:      {"OpJumpNotTruthy", []int{2}},
	OpJump:               {"OpJump", []int{2}},
	OpSetGlobal:          {"OpSetGlobal", []int{2}},
	OpSetGlobalImm:       {"OpSetGlobalImm", []int{2}},
	OpGetGlobal:          {"OpGetGlobal", []int{2}},
	OpGetGlobalImm:       {"OpGetGlobalImm", []int{2}},
	OpList:               {"OpList", []int{2}},
	OpMap:                {"OpMap", []int{2}},
	OpSet:                {"OpSet", []int{2}},
	OpIndex:              {"OpIndex", []int{}},
	OpCall:               {"OpCall", []int{1}},
	OpReturnValue:        {"OpReturnValue", []int{}},
	OpReturn:             {"OpReturn", []int{}},
	OpSetLocal:           {"OpSetLocal", []int{1}},
	OpSetLocalImm:        {"OpSetLocalImm", []int{1}},
	OpGetLocal:           {"OpGetLocal", []int{1}},
	OpGetLocalImm:        {"OpGetLocalImm", []int{1}},
	OpClosure:            {"OpClosure", []int{2, 1}},
	OpGetFree:            {"OpGetFree", []int{1}},
	OpGetFreeImm:         {"OpGetFreeImm", []int{1}},
	OpGetBuiltin:         {"OpGetBuiltin", []int{1, 1}},
	OpStringInterp:       {"OpStringInterp", []int{2, 1}},
	OpIndexSet:           {"OpIndexSet", []int{}},
	OpTry:                {"OpTry", []int{}},
	OpCatch:              {"OpCatch", []int{}},
	OpCatchEnd:           {"OpCatchEnd", []int{}},
	OpFinally:            {"OpFinally", []int{}},
	OpFinallyEnd:         {"OpFinallyEnd", []int{}},
	OpListCompLiteral:    {"OpListCompLiteral", []int{}},
	OpSetCompLiteral:     {"OpSetCompLiteral", []int{}},
	OpMapCompLiteral:     {"OpMapCompLiteral", []int{}},
	OpExecString:         {"OpExecString", []int{2}},
	OpMatchValue:         {"OpMatchValue", []int{}},
	OpMatchAny:           {"OpMatchAny", []int{}},
	OpEval:               {"OpEval", []int{}},
	OpDefaultArgs:        {"OpDefaultArgs", []int{2}},
	OpCoreCompiled:       {"OpCoreCompiled", []int{}},
	OpSlice:              {"OpSlice", []int{}},
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}

func GetOpName(op Opcode) string {
	def, ok := definitions[op]
	if !ok {
		return ""
	}
	return def.Name
}

func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}
	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}
	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)
	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		case 1:
			instruction[offset] = byte(o)
		}
		offset += width
	}
	return instruction
}

func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0
	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		case 1:
			operands[i] = int(ReadUint8(ins[offset:]))
		}
		offset += width
	}
	return operands, offset
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

func ReadUint8(ins Instructions) uint8 {
	return uint8(ins[0])
}
