package vm

import (
	"blue/code"
	"blue/compiler"
	"blue/object"
	"fmt"
	"math/big"
)

const (
	StackSize   = 2048
	GlobalsSize = 65536
	MaxFrames   = 1024
)

type VM struct {
	constants []object.Object
	stack     []object.Object
	sp        int // Always points to the next value. Top of stack is stack[sp-1]

	globals []object.Object

	frames      []*Frame
	framesIndex int
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainFrame := NewFrame(mainFn)
	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame
	return &VM{
		constants: bytecode.Constants,
		stack:     make([]object.Object, StackSize),
		sp:        0,

		globals: make([]object.Object, GlobalsSize),

		frames:      frames,
		framesIndex: 1,
	}
}

func NewWithGlobalsStore(bytecode *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode
	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++
		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])
		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
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
		case code.OpJump:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1
		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().ip = pos - 1
			}
		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			vm.globals[globalIndex] = vm.pop()
		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}
		case code.OpList:
			numElems := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			list := vm.buildList(vm.sp-numElems, vm.sp)
			vm.sp = vm.sp - numElems
			err := vm.push(list)
			if err != nil {
				return err
			}
		case code.OpMap:
			numPairs := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			m := vm.buildMap(vm.sp-numPairs, vm.sp)
			err := vm.push(m)
			if err != nil {
				return err
			}
		case code.OpSet:
			numElems := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			set := vm.buildSet(vm.sp-numElems, vm.sp)
			vm.sp = vm.sp - numElems
			err := vm.push(set)
			if err != nil {
				return err
			}
		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()
			err := vm.executeIndexExpression(left, index)
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
	isTrue := isTruthy(operand)
	if isTrue {
		return vm.push(object.FALSE)
	} else {
		return vm.push(object.TRUE)
	}
	// switch operand {
	// case object.TRUE:
	// 	return vm.push(object.FALSE)
	// case object.FALSE:
	// 	return vm.push(object.TRUE)
	// case object.NULL:
	// 	return vm.push(object.TRUE)
	// default:
	// 	return vm.push(object.FALSE)
	// }
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

func (vm *VM) buildList(startIndex, endIndex int) object.Object {
	elements := make([]object.Object, endIndex-startIndex)
	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]
	}
	return &object.List{Elements: elements}
}

func (vm *VM) buildSet(startIndex, endIndex int) object.Object {
	setMap := object.NewSetElementsWithSize(endIndex - startIndex)
	for i := startIndex; i < endIndex; i++ {
		elem := vm.stack[i]
		hashKey := object.HashObject(elem)
		setMap.Set(hashKey, object.SetPair{Value: elem, Present: struct{}{}})
	}
	return &object.Set{Elements: setMap}
}

func (vm *VM) buildMap(startIndex, endIndex int) object.Object {
	pairs := object.NewPairsMapWithSize((endIndex - startIndex) / 2)
	for i := startIndex; i < endIndex; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]
		if isError(key) {
			return key
		} else if isError(value) {
			return value
		}
		// TODO: Missing some logic here from evaluator
		if ok := object.IsHashable(key); !ok {
			return newError("unusable as a map key: %s", key.Type())
		}
		hk := object.HashObject(key)
		hashed := object.HashKey{Type: key.Type(), Value: hk}

		pairs.Set(hashed, object.MapPair{Key: key, Value: value})
	}
	return &object.Map{Pairs: pairs}
}

func (vm *VM) executeIndexExpression(left, indx object.Object) error {
	switch {
	case left.Type() == object.LIST_OBJ:
		return vm.executeListIndexExpression(left, indx)
	case left.Type() == object.SET_OBJ:
		return vm.executeSetIndexExpression(left, indx)
	case left.Type() == object.MAP_OBJ:
		return vm.executeMapIndexExpression(left, indx)
	case left.Type() == object.STRING_OBJ:
		return vm.executeStringIndexExpression(left, indx)
	// case left.Type() == object.MODULE_OBJ:
	// 	return e.evalModuleIndexExpression(left, indx)
	// case left.Type() == object.PROCESS_OBJ && indx.Type() == object.STRING_OBJ:
	// 	return e.evalProcessIndexExpression(left, indx)
	// case left.Type() == object.BLUE_STRUCT_OBJ && indx.Type() == object.STRING_OBJ:
	// 	return e.evalBlueStructIndexExpression(left, indx)
	default:
		return vm.push(newError("index operator not supported: %s.%s", left.Type(), indx.Type()))
	}
}
