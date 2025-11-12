package vm

import (
	"blue/code"
	"blue/object"
	"bytes"
	"fmt"
	"unicode/utf8"
)

func executeIntegerRangeOperator(leftVal, rightVal int64) object.Object {
	var i int64

	if leftVal < rightVal {
		size := rightVal - leftVal
		listElems := make([]object.Object, 0, size)
		for i = leftVal; i <= rightVal; i++ {
			listElems = append(listElems, &object.Integer{Value: i})
		}
		return &object.List{Elements: listElems}
	} else if rightVal < leftVal {
		size := leftVal - rightVal
		listElems := make([]object.Object, 0, size)
		for i = leftVal; i >= rightVal; i-- {
			listElems = append(listElems, &object.Integer{Value: i})
		}
		return &object.List{Elements: listElems}
	}
	// When they are equal just return a value (leftVal in this case)
	return &object.List{Elements: []object.Object{&object.Integer{Value: leftVal}}}
}

func executeIntegerNonInclusiveRangeOperator(leftVal, rightVal int64) object.Object {
	var i int64

	if leftVal < rightVal {
		size := rightVal - leftVal
		listElems := make([]object.Object, 0, size-1)
		for i = leftVal; i < rightVal; i++ {
			listElems = append(listElems, &object.Integer{Value: i})
		}
		return &object.List{Elements: listElems}
	} else if rightVal < leftVal {
		size := leftVal - rightVal
		listElems := make([]object.Object, 0, size-1)
		for i = leftVal; i > rightVal; i-- {
			listElems = append(listElems, &object.Integer{Value: i})
		}
		return &object.List{Elements: listElems}
	}
	return &object.List{Elements: []object.Object{}}
}

func newError(format string, a ...any) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func runeLen(s string) int {
	return utf8.RuneCountInString(s)
}

func nativeToBooleanObject(b bool) *object.Boolean {
	if b {
		return object.TRUE
	}
	return object.FALSE
}

// for now everything that is not null or false returns true
func isTruthy(obj object.Object) bool {
	switch obj {
	case object.NULL:
		return false
	case object.TRUE:
		return true
	case object.FALSE:
		return false
	default:
		if obj.Type() == object.MAP_OBJ {
			return len(obj.(*object.Map).Pairs.Keys) > 0
		}
		if obj.Type() == object.LIST_OBJ {
			return len(obj.(*object.List).Elements) > 0
		}
		if obj.Type() == object.SET_OBJ {
			return len(obj.(*object.Set).Elements.Keys) > 0
		}
		return true
	}
}

// isError is the helper function to determine if an object is an error
func isError(obj object.Object) bool {
	if obj != nil {
		_, isError := obj.(*object.Error)
		return isError
	}
	return false
}

func (vm *VM) stackString() string {
	var out bytes.Buffer
	for i, o := range vm.stack {
		if o != nil {
			fmt.Fprintf(&out, "%d: %s\n", i, o.Inspect())
		}
	}
	return out.String()
}

func matches(left, right object.Object) bool {
	if left.Type() != right.Type() {
		return false
	}
	// TODO: Support other types that can have an _ to signify ignore
	switch left.Type() {
	case object.MAP_OBJ:
		return matchMap(left.(*object.Map), right.(*object.Map))
	default:
		if isPrimitive(left.Type()) {
			return matchPrimitive(left, right)
		}
		return object.HashObject(left) == object.HashObject(right)
	}
}

func isPrimitive(t object.Type) bool {
	switch t {
	case object.STRING_OBJ, object.INTEGER_OBJ, object.UINTEGER_OBJ, object.FLOAT_OBJ, object.BIG_INTEGER_OBJ, object.BIG_FLOAT_OBJ, object.BOOLEAN_OBJ:
		return true
	default:
		return false
	}
}

func matchMap(left, right *object.Map) bool {
	keys := left.Pairs.Keys
	for _, k := range keys {
		leftPair, _ := left.Pairs.Get(k)
		if leftPair.Key == object.VM_IGNORE {
			continue
		}
		rightPair, ok := right.Pairs.Get(k)
		if !ok {
			return false
		}
		// Both keys matched at this point, check values
		if leftPair.Value == object.VM_IGNORE {
			continue
		}
		if leftPair.Value.Type() != rightPair.Value.Type() {
			return false
		}
		if isPrimitive(leftPair.Value.Type()) && matchPrimitive(leftPair.Value, rightPair.Value) {
			continue
		} else if leftPair.Value.Type() == object.MAP_OBJ && matchMap(leftPair.Value.(*object.Map), rightPair.Value.(*object.Map)) {
			continue
		} else if object.HashObject(leftPair.Value) == object.HashObject(rightPair.Value) {
			continue
		} else {
			return false
		}
	}
	return true
}

func matchPrimitive(left, right object.Object) bool {
	switch left.Type() {
	case object.STRING_OBJ:
		return left.(*object.Stringo).Value == right.(*object.Stringo).Value
	case object.INTEGER_OBJ:
		return left.(*object.Integer).Value == right.(*object.Integer).Value
	case object.UINTEGER_OBJ:
		return left.(*object.UInteger).Value == right.(*object.UInteger).Value
	case object.FLOAT_OBJ:
		return left.(*object.Float).Value == right.(*object.Float).Value
	case object.BIG_INTEGER_OBJ:
		return left.(*object.BigInteger).Value.Cmp(right.(*object.BigInteger).Value) == 0
	case object.BIG_FLOAT_OBJ:
		return left.(*object.BigFloat).Value.Equal(right.(*object.BigFloat).Value)
	case object.BOOLEAN_OBJ:
		return left.(*object.Boolean).Value == right.(*object.Boolean).Value
	default:
		return false
	}
}

func isBooleanOperator(op code.Opcode) bool {
	return op == code.OpEqual || op == code.OpNotEqual || op == code.OpAnd || op == code.OpOr || op == code.OpNot
}
