package vm

import (
	"blue/code"
	"blue/compiler"
	"blue/object"
	"fmt"
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
