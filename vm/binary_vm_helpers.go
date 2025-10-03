package vm

import (
	"blue/code"
	"blue/object"
	"blue/utils"
	"bytes"
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/shopspring/decimal"
)

var binaryOperationFunctions = map[object.Type]func(vm *VM, op code.Opcode, left, right object.Object) error{
	object.INTEGER_OBJ: func(vm *VM, op code.Opcode, leftObj, rightObj object.Object) error {
		leftVal := leftObj.(*object.Integer).Value
		rightVal := rightObj.(*object.Integer).Value
		switch op {
		case code.OpAdd:
			overflowed := utils.CheckOverflow(leftVal, rightVal)
			if overflowed {
				left := new(big.Int).SetInt64(leftVal)
				right := new(big.Int).SetInt64(rightVal)
				result := big.NewInt(0)
				return vm.push(&object.BigInteger{Value: result.Add(left, right)})
			}
			return vm.push(&object.Integer{Value: leftVal + rightVal})
		case code.OpMinus:
			underflowed := utils.CheckUnderflow(leftVal, rightVal)
			if underflowed {
				left := new(big.Int).SetInt64(leftVal)
				right := new(big.Int).SetInt64(rightVal)
				result := big.NewInt(0)
				return vm.push(&object.BigInteger{Value: result.Sub(left, right)})
			}
			return vm.push(&object.Integer{Value: leftVal - rightVal})
		case code.OpDiv:
			if rightVal == 0 {
				return vm.push(newError("Division by zero is not allowed"))
			}
			if rightVal > leftVal {
				return vm.push(&object.Integer{Value: 0})
			}
			return vm.push(&object.Integer{Value: leftVal / rightVal})
		case code.OpStar:
			overflowed := utils.CheckOverflowMul(leftVal, rightVal)
			if overflowed {
				left := new(big.Int).SetInt64(leftVal)
				right := new(big.Int).SetInt64(rightVal)
				result := big.NewInt(0)
				return vm.push(&object.BigInteger{Value: result.Mul(left, right)})
			}
			return vm.push(&object.Integer{Value: leftVal * rightVal})
		case code.OpPow:
			overflowed := utils.CheckOverflowPow(leftVal, rightVal)
			if overflowed {
				left := new(big.Int).SetInt64(leftVal)
				right := new(big.Int).SetInt64(rightVal)
				result := big.NewInt(0)
				return vm.push(&object.BigInteger{Value: result.Exp(left, right, nil)})
			}
			return vm.push(&object.Integer{Value: int64(math.Pow(float64(leftVal), float64(rightVal)))})
		case code.OpFlDiv:
			if rightVal == 0 {
				return vm.push(newError("Floor Division by zero is not allowed"))
			}
			if rightVal > leftVal {
				return vm.push(&object.Integer{Value: 0})
			}
			return vm.push(&object.Integer{Value: int64(leftVal / rightVal)})
		case code.OpPercent:
			if rightVal == 0 {
				return vm.push(newError("Modulus by zero is not allowed"))
			}
			if leftVal < 0 || rightVal < 0 {
				left := new(big.Int).SetInt64(leftVal)
				right := new(big.Int).SetInt64(rightVal)
				result := big.NewInt(0)
				return vm.push(&object.BigInteger{Value: result.Mod(left, right)})
			}
			return vm.push(&object.Integer{Value: int64(math.Mod(float64(leftVal), float64(rightVal)))})
		case code.OpGreaterThan:
			return vm.push(nativeToBooleanObject(leftVal > rightVal))
		case code.OpGreaterThanOrEqual:
			return vm.push(nativeToBooleanObject(leftVal >= rightVal))
		case code.OpEqual:
			return vm.push(nativeToBooleanObject(leftVal == rightVal))
		case code.OpNotEqual:
			return vm.push(nativeToBooleanObject(leftVal != rightVal))
		case code.OpRange:
			return vm.push(executeIntegerRangeOperator(leftVal, rightVal))
		case code.OpNonIncRange:
			return vm.push(executeIntegerNonInclusiveRangeOperator(leftVal, rightVal))
		default:
			return vm.push(newError("unknown operator: %s %s %s", leftObj.Type(), code.GetOpName(op), rightObj.Type()))
		}
	},
	object.BIG_INTEGER_OBJ: func(vm *VM, op code.Opcode, left, right object.Object) error {
		var leftVal, rightVal *big.Int
		if lBI, ok := left.(*object.BigInteger); ok {
			leftVal = lBI.Value
		} else if lI, ok := left.(*object.Integer); ok {
			leftVal = new(big.Int).SetInt64(lI.Value)
		} else {
			leftVal = new(big.Int).SetUint64(left.(*object.UInteger).Value)
		}
		if rBI, ok := right.(*object.BigInteger); ok {
			rightVal = rBI.Value
		} else if rI, ok := right.(*object.Integer); ok {
			rightVal = new(big.Int).SetInt64(rI.Value)
		} else {
			rightVal = new(big.Int).SetUint64(right.(*object.UInteger).Value)
		}
		result := big.NewInt(0)
		switch op {
		case code.OpAdd:
			return vm.push(&object.BigInteger{Value: result.Add(leftVal, rightVal)})
		case code.OpMinus:
			return vm.push(&object.BigInteger{Value: result.Sub(leftVal, rightVal)})
		case code.OpDiv:
			return vm.push(&object.BigInteger{Value: result.Div(leftVal, rightVal)})
		case code.OpStar:
			return vm.push(&object.BigInteger{Value: result.Mul(leftVal, rightVal)})
		case code.OpPow:
			return vm.push(&object.BigInteger{Value: result.Exp(leftVal, rightVal, nil)})
		case code.OpFlDiv:
			maybeWanted := new(big.Int)
			floored, _ := result.DivMod(leftVal, rightVal, maybeWanted)
			// Note: Ignoring the modulus here
			return vm.push(&object.BigInteger{Value: floored})
		case code.OpPercent:
			return vm.push(&object.BigInteger{Value: result.Mod(leftVal, rightVal)})
		case code.OpGreaterThan:
			compared := leftVal.Cmp(rightVal)
			return vm.push(nativeToBooleanObject(compared == 1))
		case code.OpGreaterThanOrEqual:
			compared := leftVal.Cmp(rightVal)
			return vm.push(nativeToBooleanObject(compared == 1 || compared == 0))
		case code.OpEqual:
			compared := leftVal.Cmp(rightVal)
			return vm.push(nativeToBooleanObject(compared == 0))
		case code.OpNotEqual:
			compared := leftVal.Cmp(rightVal)
			return vm.push(nativeToBooleanObject(compared != 0))
		default:
			return vm.push(newError("unknown operator: %s %s %s", left.Type(), code.GetOpName(op), right.Type()))
		}
	},
	object.FLOAT_OBJ: func(vm *VM, op code.Opcode, left, right object.Object) error {
		// Only Integers and Floats should be passed into this
		var leftVal, rightVal float64
		if lF, ok := left.(*object.Float); ok {
			leftVal = lF.Value
		} else {
			leftVal = float64(left.(*object.Integer).Value)
		}
		if rF, ok := right.(*object.Float); ok {
			rightVal = rF.Value
		} else {
			rightVal = float64(right.(*object.Integer).Value)
		}
		switch op {
		case code.OpAdd:
			return vm.push(&object.Float{Value: leftVal + rightVal})
		case code.OpMinus:
			return vm.push(&object.Float{Value: leftVal - rightVal})
		case code.OpDiv:
			return vm.push(&object.Float{Value: leftVal / rightVal})
		case code.OpStar:
			return vm.push(&object.Float{Value: leftVal * rightVal})
		case code.OpPow:
			return vm.push(&object.Float{Value: math.Pow(leftVal, rightVal)})
		case code.OpFlDiv:
			return vm.push(&object.Float{Value: math.Floor(leftVal / rightVal)})
		case code.OpPercent:
			return vm.push(&object.Float{Value: math.Mod(leftVal, rightVal)})
		case code.OpGreaterThan:
			return vm.push(nativeToBooleanObject(leftVal > rightVal))
		case code.OpGreaterThanOrEqual:
			return vm.push(nativeToBooleanObject(leftVal >= rightVal))
		case code.OpEqual:
			return vm.push(nativeToBooleanObject(leftVal == rightVal))
		case code.OpNotEqual:
			return vm.push(nativeToBooleanObject(leftVal != rightVal))
		default:
			return vm.push(newError("unknown operator: %s %s %s", left.Type(), code.GetOpName(op), right.Type()))
		}
	},
	object.BIG_FLOAT_OBJ: func(vm *VM, op code.Opcode, left, right object.Object) error {
		var leftVal, rightVal decimal.Decimal
		if lBF, ok := left.(*object.BigFloat); ok {
			leftVal = lBF.Value
		} else if lF, ok := left.(*object.Float); ok {
			leftVal = decimal.NewFromFloat(lF.Value)
		} else if lI, ok := left.(*object.Integer); ok {
			leftVal = decimal.NewFromInt(lI.Value)
		} else if lBI, ok := left.(*object.BigInteger); ok {
			leftVal = decimal.NewFromBigInt(lBI.Value, 0)
		}
		if rBF, ok := right.(*object.BigFloat); ok {
			rightVal = rBF.Value
		} else if rF, ok := right.(*object.Float); ok {
			rightVal = decimal.NewFromFloat(rF.Value)
		} else if rI, ok := right.(*object.Integer); ok {
			rightVal = decimal.NewFromInt(rI.Value)
		} else if rBI, ok := right.(*object.BigInteger); ok {
			rightVal = decimal.NewFromBigInt(rBI.Value, 0)
		}
		switch op {
		case code.OpAdd:
			return vm.push(&object.BigFloat{Value: leftVal.Add(rightVal)})
		case code.OpMinus:
			return vm.push(&object.BigFloat{Value: leftVal.Sub(rightVal)})
		case code.OpDiv:
			return vm.push(&object.BigFloat{Value: leftVal.Div(rightVal)})
		case code.OpStar:
			return vm.push(&object.BigFloat{Value: leftVal.Mul(rightVal)})
		case code.OpPow:
			return vm.push(&object.BigFloat{Value: leftVal.Pow(rightVal)})
		case code.OpFlDiv:
			return vm.push(&object.BigFloat{Value: leftVal.Div(rightVal).Floor()})
		case code.OpPercent:
			return vm.push(&object.BigFloat{Value: leftVal.Mod(rightVal)})
		case code.OpGreaterThan:
			compared := leftVal.Cmp(rightVal)
			return vm.push(nativeToBooleanObject(compared == 1))
		case code.OpGreaterThanOrEqual:
			compared := leftVal.Cmp(rightVal)
			return vm.push(nativeToBooleanObject(compared == 1 || compared == 0))
		case code.OpEqual:
			compared := leftVal.Cmp(rightVal)
			return vm.push(nativeToBooleanObject(compared == 0))
		case code.OpNotEqual:
			compared := leftVal.Cmp(rightVal)
			return vm.push(nativeToBooleanObject(compared != 0))
		default:
			return vm.push(newError("unknown operator: %s %s %s", left.Type(), code.GetOpName(op), right.Type()))
		}
	},
	object.UINTEGER_OBJ: func(vm *VM, op code.Opcode, left, right object.Object) error {
		var leftVal, rightVal uint64
		if lUI, ok := left.(*object.UInteger); ok {
			leftVal = lUI.Value
		} else {
			leftIntVal := left.(*object.Integer).Value
			if leftIntVal < 0 {
				return vm.push(newError("Left Integer was negative, and is not allowed for Unsigned Integer operations. %s %s %s", left.Inspect(), code.GetOpName(op), right.Inspect()))
			}
			leftVal = uint64(leftIntVal)
		}
		if rUI, ok := right.(*object.UInteger); ok {
			rightVal = rUI.Value
		} else {
			rightIntVal := right.(*object.Integer).Value
			if rightIntVal < 0 {
				return vm.push(newError("Right Integer was negative, and is not allowed for Unsigned Integer operations. %s %s %s", left.Inspect(), code.GetOpName(op), right.Inspect()))
			}
		}
		switch op {
		case code.OpAdd:
			return vm.push(&object.UInteger{Value: leftVal + rightVal})
		case code.OpMinus:
			return vm.push(&object.UInteger{Value: leftVal - rightVal})
		case code.OpDiv:
			return vm.push(&object.UInteger{Value: leftVal / rightVal})
		case code.OpStar:
			return vm.push(&object.UInteger{Value: leftVal * rightVal})
		case code.OpPow:
			return vm.push(&object.UInteger{Value: uint64(math.Pow(float64(leftVal), float64(rightVal)))})
		case code.OpFlDiv:
			return vm.push(&object.UInteger{Value: uint64(math.Floor(float64(leftVal) / float64(rightVal)))})
		case code.OpPercent:
			return vm.push(&object.UInteger{Value: uint64(math.Mod(float64(leftVal), float64(rightVal)))})
		case code.OpAmpersand:
			return vm.push(&object.UInteger{Value: leftVal & rightVal})
		case code.OpPipe:
			return vm.push(&object.UInteger{Value: leftVal | rightVal})
		case code.OpCarat:
			return vm.push(&object.UInteger{Value: leftVal ^ rightVal})
		case code.OpRshift:
			return vm.push(&object.UInteger{Value: leftVal >> rightVal})
		case code.OpLshift:
			return vm.push(&object.UInteger{Value: leftVal << rightVal})
		case code.OpGreaterThan:
			return vm.push(nativeToBooleanObject(leftVal > rightVal))
		case code.OpGreaterThanOrEqual:
			return vm.push(nativeToBooleanObject(leftVal >= rightVal))
		case code.OpEqual:
			return vm.push(nativeToBooleanObject(leftVal == rightVal))
		case code.OpNotEqual:
			return vm.push(nativeToBooleanObject(leftVal != rightVal))
		default:
			return vm.push(newError("unknown operator: %s %s %s", left.Type(), code.GetOpName(op), right.Type()))
		}
	},
	object.STRING_OBJ: func(vm *VM, op code.Opcode, left, right object.Object) error {
		leftStr := left.(*object.Stringo).Value
		rightStr := right.(*object.Stringo).Value
		switch op {
		case code.OpAdd:
			return vm.push(&object.Stringo{Value: leftStr + rightStr})
		case code.OpEqual:
			return vm.push(nativeToBooleanObject(leftStr == rightStr))
		case code.OpNotEqual:
			return vm.push(nativeToBooleanObject(leftStr != rightStr))
		case code.OpIn:
			return vm.push(nativeToBooleanObject(strings.Contains(rightStr, leftStr)))
		case code.OpNotin:
			return vm.push(nativeToBooleanObject(!strings.Contains(rightStr, leftStr)))
		case code.OpRange:
			if runeLen(leftStr) != 1 {
				return vm.push(newError("operator .. expects left string to be 1 rune"))
			}
			if runeLen(rightStr) != 1 {
				return vm.push(newError("operator .. expects right string to be 1 rune"))
			}
			lr := []rune(leftStr)[0]
			rr := []rune(rightStr)[0]
			if lr == rr {
				// If they are the same just return vm.push(a list with the single element)
				// because this is the inclusive operator
				return vm.push(&object.List{Elements: []object.Object{left}})
			}
			elements := []object.Object{}
			if lr > rr {
				// Left rune is > so we are descending
				for i := lr; i >= rr; i-- {
					s := string(i)
					elements = append(elements, &object.Stringo{Value: s})
				}
				return vm.push(&object.List{Elements: elements})
			} else {
				// Right rune is > so we are ascending
				for i := lr; i <= rr; i++ {
					s := string(i)
					elements = append(elements, &object.Stringo{Value: s})
				}
				return vm.push(&object.List{Elements: elements})
			}
		case code.OpNonIncRange:
			if runeLen(leftStr) != 1 {
				return vm.push(newError("operator ..< expects left string to be 1 rune"))
			}
			if runeLen(rightStr) != 1 {
				return vm.push(newError("operator ..< expects right string to be 1 rune"))
			}
			lr := []rune(leftStr)[0]
			rr := []rune(rightStr)[0]
			if lr == rr {
				// If they are the same just return vm.push(an empty list because this is non-inclusive)
				return vm.push(&object.List{Elements: []object.Object{}})
			}
			elements := []object.Object{}
			if lr > rr {
				// Left rune is > so we are descending
				for i := lr; i > rr; i-- {
					s := string(i)
					elements = append(elements, &object.Stringo{Value: s})
				}
				return vm.push(&object.List{Elements: elements})
			} else {
				// Right rune is > so we are ascending
				for i := lr; i < rr; i++ {
					s := string(i)
					elements = append(elements, &object.Stringo{Value: s})
				}
				return vm.push(&object.List{Elements: elements})
			}
		default:
			return vm.push(newError("unknown operator: %s %s %s", left.Type(), code.GetOpName(op), right.Type()))
		}
	},
	object.SET_OBJ: func(vm *VM, op code.Opcode, left, right object.Object) error {
		newSet := &object.Set{Elements: object.NewSetElements()}
		if op == code.OpAdd {
			var s *object.Set
			var key uint64
			var obj object.Object
			if left.Type() == object.SET_OBJ {
				// return set with right obj added
				s = left.(*object.Set)
				key = object.HashObject(right)
				obj = right
			} else {
				// return set with left obj added
				s = right.(*object.Set)
				key = object.HashObject(left)
				obj = left
			}
			for _, k := range s.Elements.Keys {
				v, ok := s.Elements.Get(k)
				if ok {
					newSet.Elements.Set(k, v)
				}
			}
			if _, ok := s.Elements.Get(key); !ok {
				// Key does not exist, add new elem
				newSet.Elements.Set(key, object.SetPair{Value: obj, Present: struct{}{}})
			}
			return vm.push(newSet)
		}
		leftE := left.(*object.Set).Elements
		rightE := right.(*object.Set).Elements
		var leftElems *object.OrderedMap2[uint64, object.SetPair]
		var rightElems *object.OrderedMap2[uint64, object.SetPair]
		if leftE.Len() >= rightE.Len() {
			leftElems = leftE
			rightElems = rightE
		} else {
			leftElems = rightE
			rightElems = leftE
		}
		switch op {
		case code.OpPipe:
			// union
			for _, k := range leftElems.Keys {
				v, ok := leftElems.Get(k)
				if !ok {
					continue
				}
				newSet.Elements.Set(k, v)
			}
			for _, k := range rightElems.Keys {
				v, ok := rightElems.Get(k)
				if !ok {
					continue
				}
				newSet.Elements.Set(k, v)
			}
			return vm.push(newSet)
		case code.OpAmpersand:
			// intersect
			for _, k := range leftElems.Keys {
				v, ok := leftElems.Get(k)
				if !ok {
					continue
				}
				_, ok = rightElems.Get(k)
				if !ok {
					continue
				}
				newSet.Elements.Set(k, v)
			}
			return vm.push(newSet)
		case code.OpCarat:
			// symmetric difference
			for _, k := range leftElems.Keys {
				v, ok := leftElems.Get(k)
				if !ok {
					continue
				}
				_, ok = rightElems.Get(k)
				if !ok {
					newSet.Elements.Set(k, v)
				}
			}
			for _, k := range rightElems.Keys {
				v, ok := rightElems.Get(k)
				if !ok {
					continue
				}
				_, ok = leftElems.Get(k)
				if !ok {
					newSet.Elements.Set(k, v)
				}
			}
			return vm.push(newSet)
		case code.OpGreaterThanOrEqual:
			// left is superset of right
			for _, k := range rightE.Keys {
				if _, ok := leftE.Get(k); !ok {
					return vm.push(object.FALSE)
				}
			}
			return vm.push(object.TRUE)
		case code.OpMinus:
			// difference
			for _, k := range leftElems.Keys {
				v, ok := leftElems.Get(k)
				if !ok {
					continue
				}
				_, ok = rightElems.Get(k)
				if !ok {
					newSet.Elements.Set(k, v)
				}
			}
			return vm.push(newSet)
		case code.OpEqual:
			for _, k := range leftElems.Keys {
				_, ok := rightElems.Get(k)
				if !ok {
					return vm.push(object.FALSE)
				}
			}
			return vm.push(object.TRUE)
		case code.OpNotEqual:
			for _, k := range leftElems.Keys {
				_, ok := rightElems.Get(k)
				if !ok {
					return vm.push(object.TRUE)
				}
			}
			return vm.push(object.FALSE)
		default:
			return vm.push(newError("unknown operator: %s %s %s", left.Type(), code.GetOpName(op), right.Type()))
		}
	},
	object.LIST_OBJ: func(vm *VM, op code.Opcode, left, right object.Object) error {
		leftListObj := left.(*object.List)
		rightListObj := right.(*object.List)
		leftElements := leftListObj.Elements
		rightElements := rightListObj.Elements
		leftSize := len(leftElements)
		rightSize := len(rightElements)
		switch op {
		case code.OpAdd:
			newList := make([]object.Object, 0, leftSize+rightSize)
			newList = append(newList, leftElements...)
			newList = append(newList, rightElements...)
			return vm.push(&object.List{Elements: newList})
		case code.OpEqual, code.OpNotEqual:
			return vm.executeDefaultBinaryOperation(op, left, right)
		default:
			return vm.push(newError("unknown operator: %s %s %s", left.Type(), code.GetOpName(op), right.Type()))
		}
	},
	object.BYTES_OBJ: func(vm *VM, op code.Opcode, left, right object.Object) error {
		leftBs := left.(*object.Bytes).Value
		rightBs := right.(*object.Bytes).Value
		switch op {
		case code.OpAmpersand:
			if len(leftBs) != len(rightBs) {
				return vm.push(newError("length of left and right bytes must match to perform bitwise AND operation. got: len(l)=%d, len(r)=%d", len(leftBs), len(rightBs)))
			}
			buf := make([]byte, len(leftBs))
			for i := range leftBs {
				buf[i] = leftBs[i] & rightBs[i]
			}
			return vm.push(&object.Bytes{Value: buf})
		case code.OpPipe:
			if len(leftBs) != len(rightBs) {
				return vm.push(newError("length of left and right bytes must match to perform bitwise OR operation. got: len(l)=%d, len(r)=%d", len(leftBs), len(rightBs)))
			}
			buf := make([]byte, len(leftBs))
			for i := range leftBs {
				buf[i] = leftBs[i] | rightBs[i]
			}
			return vm.push(&object.Bytes{Value: buf})
		case code.OpCarat:
			if len(leftBs) != len(rightBs) {
				return vm.push(newError("length of left and right bytes must match to perform bitwise XOR operation. got: len(l)=%d, len(r)=%d", len(leftBs), len(rightBs)))
			}
			buf := make([]byte, len(leftBs))
			for i := range leftBs {
				buf[i] = leftBs[i] ^ rightBs[i]
			}
			return vm.push(&object.Bytes{Value: buf})
		default:
			return vm.executeDefaultBinaryOperation(op, left, right)
		}
	},
	// TODO: Handle other defaults when type matches (list, set, map)
}

func (vm *VM) executeBinaryOperationDifferentTypes(op code.Opcode, left, right object.Object, leftType, rightType object.Type) error {
	if op == code.OpAdd && (leftType == object.SET_OBJ || rightType == object.SET_OBJ) {
		return binaryOperationFunctions[object.SET_OBJ](vm, op, left, right)
	}
	if leftType == object.BIG_INTEGER_OBJ && rightType == object.INTEGER_OBJ ||
		leftType == object.INTEGER_OBJ && rightType == object.BIG_INTEGER_OBJ ||
		leftType == object.UINTEGER_OBJ && rightType == object.BIG_INTEGER_OBJ ||
		leftType == object.BIG_INTEGER_OBJ && rightType == object.UINTEGER_OBJ {
		return binaryOperationFunctions[object.BIG_INTEGER_OBJ](vm, op, left, right)
	}
	if leftType == object.FLOAT_OBJ && rightType == object.FLOAT_OBJ ||
		leftType == object.INTEGER_OBJ && rightType == object.FLOAT_OBJ ||
		leftType == object.FLOAT_OBJ && rightType == object.INTEGER_OBJ {
		return binaryOperationFunctions[object.FLOAT_OBJ](vm, op, left, right)
	}
	if leftType == object.FLOAT_OBJ && rightType == object.BIG_INTEGER_OBJ ||
		leftType == object.BIG_INTEGER_OBJ && rightType == object.FLOAT_OBJ ||
		leftType == object.FLOAT_OBJ && rightType == object.BIG_FLOAT_OBJ ||
		leftType == object.BIG_FLOAT_OBJ && rightType == object.FLOAT_OBJ ||
		leftType == object.INTEGER_OBJ && rightType == object.BIG_FLOAT_OBJ ||
		leftType == object.BIG_FLOAT_OBJ && rightType == object.INTEGER_OBJ ||
		leftType == object.UINTEGER_OBJ && rightType == object.BIG_FLOAT_OBJ ||
		leftType == object.BIG_FLOAT_OBJ && rightType == object.UINTEGER_OBJ ||
		leftType == object.BIG_FLOAT_OBJ && rightType == object.BIG_INTEGER_OBJ ||
		leftType == object.BIG_INTEGER_OBJ && rightType == object.BIG_FLOAT_OBJ {
		return binaryOperationFunctions[object.BIG_FLOAT_OBJ](vm, op, left, right)
	}
	if leftType == object.INTEGER_OBJ && rightType == object.UINTEGER_OBJ ||
		leftType == object.UINTEGER_OBJ && rightType == object.INTEGER_OBJ {
		return binaryOperationFunctions[object.UINTEGER_OBJ](vm, op, left, right)
	}
	if left.Type() == object.STRING_OBJ && right.Type() == object.INTEGER_OBJ ||
		left.Type() == object.INTEGER_OBJ && right.Type() == object.STRING_OBJ ||
		left.Type() == object.STRING_OBJ && right.Type() == object.UINTEGER_OBJ ||
		left.Type() == object.UINTEGER_OBJ && right.Type() == object.STRING_OBJ {
		return vm.executeBinaryStringAndIntOrUintOperation(op, left, right)
	}
	// TODO: More to handle here
	return fmt.Errorf("handle %s %s %s", leftType, code.GetOpName(op), rightType)
}

func (vm *VM) executeBinaryStringAndIntOrUintOperation(op code.Opcode, left, right object.Object) error {
	var strToBuild string
	var amount uint64
	if s, ok := left.(*object.Stringo); ok {
		strToBuild = s.Value
	} else if lI, ok := left.(*object.Integer); ok {
		amount = uint64(lI.Value)
	} else {
		amount = left.(*object.UInteger).Value
	}
	if s, ok := right.(*object.Stringo); ok {
		strToBuild = s.Value
	} else if rI, ok := right.(*object.Integer); ok {
		amount = uint64(rI.Value)
	} else {
		amount = right.(*object.UInteger).Value
	}
	switch op {
	case code.OpStar:
		var out bytes.Buffer
		var i uint64
		for i = 0; i < amount; i++ {
			out.WriteString(strToBuild)
		}
		return vm.push(&object.Stringo{Value: out.String()})
	default:
		return vm.executeDefaultBinaryOperation(op, left, right)
	}
}

func (vm *VM) executeDefaultBinaryOperation(op code.Opcode, left, right object.Object) error {
	switch {
	case op == code.OpEqual:
		return vm.push(nativeToBooleanObject(object.HashObject(left) == object.HashObject(right)))
	case op == code.OpNotEqual:
		return vm.push(nativeToBooleanObject(object.HashObject(left) != object.HashObject(right)))
	case op == code.OpAnd:
		leftBool, ok := left.(*object.Boolean)
		if !ok {
			return vm.push(newError("unknown operator: %s %s %s", left.Type(), code.GetOpName(op), right.Type()))
		}
		rightBool, ok := right.(*object.Boolean)
		if !ok {
			return vm.push(newError("unknown operator: %s %s %s", left.Type(), code.GetOpName(op), right.Type()))
		}
		return vm.push(nativeToBooleanObject(leftBool.Value && rightBool.Value))
	case op == code.OpOr:
		if left == object.NULL {
			// Null coalescing operator returns right side if left is null
			return vm.push(right)
		}
		leftBool, ok := left.(*object.Boolean)
		if !ok {
			return vm.push(newError("unknown operator: %s %s %s", left.Type(), code.GetOpName(op), right.Type()))
		}
		rightBool, ok := right.(*object.Boolean)
		if !ok {
			return vm.push(newError("unknown operator: %s %s %s", left.Type(), code.GetOpName(op), right.Type()))
		}
		return vm.push(nativeToBooleanObject(leftBool.Value || rightBool.Value))
	case (op == code.OpIn || op == code.OpNotin) && (right.Type() == object.LIST_OBJ || right.Type() == object.SET_OBJ || right.Type() == object.MAP_OBJ):
		// return e.evalInOrNotinInfixExpression(operator, left, right)
		return fmt.Errorf("handle this here ----")
	case left.Type() != right.Type():
		return vm.push(newError("type mismatch: %s %s %s", left.Type(), code.GetOpName(op), right.Type()))
	default:
		return vm.push(newError("unknown operator: %s %s %s", left.Type(), code.GetOpName(op), right.Type()))
	}
}
