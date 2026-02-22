package utils

import (
	"blue/code"
	"blue/object"
	"bytes"
	"fmt"
	"log"
	"math"
	"strings"
)

var ENABLE_VM_CACHING = true

func CheckOverflow(leftVal, rightVal int64) bool {
	result := leftVal + rightVal
	return result-leftVal != rightVal
}

func CheckUnderflow(leftVal, rightVal int64) bool {
	result := leftVal - rightVal
	return result+rightVal != leftVal
}

func CheckOverflowMul(leftVal, rightVal int64) bool {
	if leftVal == 0 || rightVal == 0 || leftVal == 1 || rightVal == 1 {
		return false
	}
	if leftVal == math.MinInt64 || rightVal == math.MinInt64 {
		return true
	}
	result := leftVal * rightVal
	return result/rightVal != leftVal
}

func CheckOverflowPow(leftVal, rightVal int64) bool {
	if leftVal == 0 || rightVal == 0 || leftVal == 1 || rightVal == 1 {
		return false
	}
	if leftVal == math.MinInt64 || rightVal == math.MinInt64 {
		return true
	}
	if rightVal > 63 && leftVal > 1 {
		return true
	}
	return false
}

func IfNameInMapSetEnv(env *object.Environment, m object.OrderedMap2[object.HashKey, object.MapPair], name string) bool {
	for _, k := range m.Keys {
		mp, _ := m.Get(k)
		if mp.Key.Type() == object.STRING_OBJ {
			s := mp.Key.(*object.Stringo).Value
			if name == s {
				env.Set(name, mp.Value)
				return true
			}
		}
	}
	return false
}

func GetNextOpCallPos(ins code.Instructions, ip int) int {
	i := ip
	for i < len(ins) {
		def, err := code.Lookup(ins[i])
		if err != nil {
			log.Fatalf("UNREACHABLE - failed to lookup instruction")
		}
		if def.Name == "OpCall" {
			return i
		}
		_, read := code.ReadOperands(def, ins[i+1:])
		i += 1 + read
	}
	return -1
}

func BytecodeDebugStringWithOffset(offset int, ins code.Instructions, constants []object.Object) string {
	var out bytes.Buffer
	i := 0
	for i < len(ins) {
		def, err := code.Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}
		operands, read := code.ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d %s\n", offset+i, fmtInstruction(def, operands, constants))
		i += 1 + read
	}
	return out.String()
}

func BytecodeDebugString(ins code.Instructions, constants []object.Object) string {
	return BytecodeDebugStringWithOffset(0, ins, constants)
}

func fmtInstruction(def *code.Definition, operands []int, constants []object.Object) string {
	operandCount := len(def.OperandWidths)
	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n",
			len(operands), operandCount)
	}
	switch operandCount {
	case 0:
		return def.Name
	case 1:
		lastPart := ""
		if def.Name == "OpConstant" {
			if operands[0] > len(constants) {
				// Noticed this occurred with offset of core compiled but not without it
				// so added this so that its noted during compile
				// VM would crash anyways if this code ran with undefined reference
				lastPart = " (<nil>) <---------------- UNDEFINED CONSTANT REFERENCE (this really shouldn't happen)"
			} else {
				lastPart = fmt.Sprintf(" (%s)", constants[operands[0]].Inspect())
			}
		}
		return fmt.Sprintf("%s %d%s", def.Name, operands[0], lastPart)
	case 2:
		lastPart := ""
		switch def.Name {
		case "OpGetBuiltin":
			if operands[0] == object.BuiltinobjsModuleIndex {
				lastPart = fmt.Sprintf(" (%s)", object.BuiltinobjsList[operands[1]].Name)
			} else {
				lastPart = fmt.Sprintf(" (%s)", object.AllBuiltins[operands[0]].Builtins[operands[1]].Name)
			}
		case "OpClosure":
			cf := constants[operands[0]].(*object.CompiledFunction)
			lastPart = fmt.Sprintf("\n\t%s", strings.ReplaceAll(BytecodeDebugString(cf.Instructions, constants), "\n", "\n\t"))
			lastPart = strings.TrimSuffix(lastPart, "\n\t")
		}
		return fmt.Sprintf("%s %d %d%s", def.Name, operands[0], operands[1], lastPart)
	}
	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}
