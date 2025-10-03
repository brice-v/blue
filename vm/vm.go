package vm

import (
	"blue/code"
	"blue/compiler"
	"blue/object"
	"fmt"
	"math/big"
)

const StackSize = 2048

type VM struct {
	constants    []object.Object
	instructions code.Instructions
	stack        []object.Object
	sp           int // Always points to the next value. Top of stack is stack[sp-1]
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		stack:        make([]object.Object, StackSize),
		sp:           0,
	}
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])
		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}
		case code.OpTrue:
			vm.push(object.TRUE)
		case code.OpFalse:
			vm.push(object.FALSE)
		case code.OpNull:
			vm.push(object.NULL)
		case code.OpPop:
			vm.pop()
		case code.OpAdd, code.OpMinus, code.OpStar, code.OpPow, code.OpDiv,
			code.OpFlDiv, code.OpPercent, code.OpCarat, code.OpAmpersand,
			code.OpPipe, code.OpIn, code.OpNotin, code.OpRange, code.OpNonIncRange,
			code.OpAnd, code.OpEqual, code.OpOr, code.OpGreaterThan, code.OpGreaterThanOrEqual:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case code.OpNot:
			err := vm.executeNotOperation()
			if err != nil {
				return err
			}
		case code.OpNeg:
			err := vm.executeNegOperation()
			if err != nil {
				return err
			}
		case code.OpTilde:
			err := vm.executeBitwiseNotOperation()
			if err != nil {
				return err
			}
		case code.OpLshiftPre:
			err := vm.executeLshiftPrefixOperation()
			if err != nil {
				return err
			}
		case code.OpRshiftPost:
			err := vm.executeRshiftPostfixOperation()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow when trying to push %+#v", o.Inspect())
	}
	vm.stack[vm.sp] = o
	vm.sp++
	return nil
}

func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()
	leftType := left.Type()
	rightType := right.Type()
	if leftType == rightType {
		return binaryOperationFunctions[leftType](vm, op, left, right)
	} else if leftType != rightType {
		return vm.executeBinaryOperationDifferentTypes(op, left, right, leftType, rightType)
	}
	return nil
}

func (vm *VM) executeNotOperation() error {
	operand := vm.pop()
	switch operand {
	case object.TRUE:
		return vm.push(object.FALSE)
	case object.FALSE:
		return vm.push(object.TRUE)
	default:
		return vm.push(object.FALSE)
	}
}

func (vm *VM) executeNegOperation() error {
	operand := vm.pop()
	if operand.Type() == object.INTEGER_OBJ {
		value := operand.(*object.Integer).Value
		return vm.push(&object.Integer{Value: -value})
	}
	if operand.Type() == object.FLOAT_OBJ {
		value := operand.(*object.Float).Value
		return vm.push(&object.Float{Value: -value})
	}
	if operand.Type() == object.BIG_INTEGER_OBJ {
		value := operand.(*object.BigInteger).Value
		return vm.push(&object.BigInteger{Value: new(big.Int).Neg(value)})
	}
	if operand.Type() == object.BIG_FLOAT_OBJ {
		value := operand.(*object.BigFloat).Value
		return vm.push(&object.BigFloat{Value: value.Neg()})
	}

	return vm.push(newError("unknown operator: -%s", operand.Type()))
}

func (vm *VM) executeBitwiseNotOperation() error {
	operand := vm.pop()
	switch operand.Type() {
	case object.INTEGER_OBJ:
		value := operand.(*object.Integer).Value
		return vm.push(&object.Integer{Value: ^value})
	case object.UINTEGER_OBJ:
		value := operand.(*object.UInteger).Value
		return vm.push(&object.UInteger{Value: ^value})
	case object.BYTES_OBJ:
		value := operand.(*object.Bytes).Value
		buf := make([]byte, len(value))
		for i, b := range value {
			buf[i] = ^b
		}
		return vm.push(&object.Bytes{Value: buf})
	default:
		return vm.push(newError("unknown operator: ~%s", operand.Type()))
	}
}

func (vm *VM) executeLshiftPrefixOperation() error {
	operand := vm.pop()
	switch operand.Type() {
	case object.LIST_OBJ:
		l := operand.(*object.List)
		listLen := len(l.Elements)
		if listLen == 0 {
			return vm.push(object.NULL)
		}
		e := l.Elements[0]
		if listLen == 1 {
			l.Elements = []object.Object{}
		} else {
			l.Elements = l.Elements[1:listLen]
		}
		return vm.push(e)
	default:
		return vm.push(newError("unknown operator: << %s", operand.Type()))
	}
}

func (vm *VM) executeRshiftPostfixOperation() error {
	operand := vm.pop()
	switch operand.Type() {
	case object.LIST_OBJ:
		l := operand.(*object.List)
		listLen := len(l.Elements)
		if listLen == 0 {
			return vm.push(object.NULL)
		}
		e := l.Elements[listLen-1]
		if listLen == 1 {
			l.Elements = []object.Object{}
		} else {
			l.Elements = l.Elements[0 : listLen-1]
		}
		return vm.push(e)
	default:
		return vm.push(newError("unknown operator: %s >>", operand.Type()))
	}
}
