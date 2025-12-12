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

func newPositionalTypeError(funName string, pos int, expectedType object.Type, currentType object.Type) *object.Error {
	return newError("PositionalTypeError: `%s` expects argument %d to be %s. got=%s", funName, pos, expectedType, currentType)
}

func newInvalidArgCountError(funName string, got, want int, otherCount string) *object.Error {
	if otherCount == "" {
		return newError("InvalidArgCountError: `%s` wrong number of args. got=%d, want=%d", funName, got, want)
	}
	return newError("InvalidArgCountError: `%s` wrong number of args. got=%d, want=%d %s", funName, got, want, otherCount)
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

// goObjectToBlueObject will only work for simple go types
func goObjectToBlueObject(goObject any) (object.Object, error) {
	switch obj := goObject.(type) {
	case string:
		return &object.Stringo{Value: obj}, nil
	case int:
		return &object.Integer{Value: int64(obj)}, nil
	case int64:
		return &object.Integer{Value: obj}, nil
	case uint:
		return &object.UInteger{Value: uint64(obj)}, nil
	case uint64:
		return &object.UInteger{Value: obj}, nil
	case float32:
		return &object.Float{Value: float64(obj)}, nil
	case float64:
		return &object.Float{Value: obj}, nil
	case bool:
		x := nativeToBooleanObject(obj)
		return x, nil
	case nil:
		return object.NULL, nil
	case []any:
		l := &object.List{Elements: make([]object.Object, len(obj))}
		for i, e := range obj {
			val, err := goObjectToBlueObject(e)
			if err != nil {
				return nil, err
			}
			l.Elements[i] = val
		}
		return l, nil
	case map[string]any:
		m := &object.Map{Pairs: object.NewPairsMap()}
		for k, v := range obj {
			key := &object.Stringo{Value: k}
			hashKey := object.HashObject(key)
			hk := object.HashKey{
				Type:  object.STRING_OBJ,
				Value: hashKey,
			}
			val, err := goObjectToBlueObject(v)
			if err != nil {
				return nil, err
			}
			m.Pairs.Set(hk, object.MapPair{
				Key:   key,
				Value: val,
			})
		}
		return m, nil
	case *object.OrderedMap2[string, any]:
		m := &object.Map{Pairs: object.NewPairsMap()}
		for _, k := range obj.Keys {
			v, _ := obj.Get(k)
			key := &object.Stringo{Value: k}
			hk := object.HashKey{
				Type:  object.STRING_OBJ,
				Value: object.HashObject(key),
			}
			val, err := goObjectToBlueObject(v)
			if err != nil {
				return nil, err
			}
			m.Pairs.Set(hk, object.MapPair{
				Key:   key,
				Value: val,
			})
		}
		return m, nil
	case *object.OrderedMap2[int64, any]:
		m := &object.Map{Pairs: object.NewPairsMap()}
		for _, k := range obj.Keys {
			v, _ := obj.Get(k)
			key := &object.Integer{Value: k}
			hk := object.HashKey{
				Type:  object.INTEGER_OBJ,
				Value: object.HashObject(key),
			}
			val, err := goObjectToBlueObject(v)
			if err != nil {
				return nil, err
			}
			m.Pairs.Set(hk, object.MapPair{
				Key:   key,
				Value: val,
			})
		}
		return m, nil
	case *object.OrderedMap2[float64, any]:
		m := &object.Map{Pairs: object.NewPairsMap()}
		for _, k := range obj.Keys {
			v, _ := obj.Get(k)
			key := &object.Float{Value: k}
			hk := object.HashKey{
				Type:  object.FLOAT_OBJ,
				Value: object.HashObject(key),
			}
			val, err := goObjectToBlueObject(v)
			if err != nil {
				return nil, err
			}
			m.Pairs.Set(hk, object.MapPair{
				Key:   key,
				Value: val,
			})
		}
		return m, nil
	case *object.OrderedMap2[uint64, object.SetPairGo]:
		set := &object.Set{Elements: object.NewOrderedMap[uint64, object.SetPair]()}
		for _, k := range obj.Keys {
			v, _ := obj.Get(k)
			val, err := goObjectToBlueObject(v.Value)
			if err != nil {
				return nil, err
			}
			set.Elements.Set(k, object.SetPair{Value: val, Present: v.Present})
		}
		return set, nil
	default:
		return nil, fmt.Errorf("goObjectToBlueObject: TODO: Type currently unsupported: (%T)", obj)
	}
}

func blueObjectToGoObject(blueObject object.Object) (any, error) {
	if blueObject == nil {
		return nil, fmt.Errorf("blueObjectToGoObject: blueObject must not be nil")
	}
	switch blueObject.Type() {
	case object.STRING_OBJ:
		return blueObject.(*object.Stringo).Value, nil
	case object.INTEGER_OBJ:
		return blueObject.(*object.Integer).Value, nil
	case object.FLOAT_OBJ:
		return blueObject.(*object.Float).Value, nil
	case object.NULL_OBJ:
		return nil, nil
	case object.BOOLEAN_OBJ:
		return blueObject.(*object.Boolean).Value, nil
	case object.MAP_OBJ:
		m := blueObject.(*object.Map)
		allInts := true
		allFloats := true
		allStrings := true
		for _, k := range m.Pairs.Keys {
			mp, _ := m.Pairs.Get(k)
			allInts = allInts && mp.Key.Type() == object.INTEGER_OBJ
			allFloats = allFloats && mp.Key.Type() == object.FLOAT_OBJ
			allStrings = allStrings && mp.Key.Type() == object.STRING_OBJ
		}
		if !allStrings && !allFloats && !allInts {
			return nil, fmt.Errorf("blueObjectToGoObject: Map must only have STRING, INTEGER, or FLOAT keys")
		}
		if allStrings {
			pairs := object.NewOrderedMap[string, any]()
			for _, k := range m.Pairs.Keys {
				mp, _ := m.Pairs.Get(k)
				if mp.Value.Type() == object.MAP_OBJ {
					return nil, fmt.Errorf("blueObjectToGoObject: Map must not have map values. got=%s", mp.Value.Type())
				}
				val, err := blueObjectToGoObject(mp.Value)
				if err != nil {
					return nil, err
				}
				pairs.Set(mp.Key.(*object.Stringo).Value, val)
			}
			return pairs, nil
		} else if allInts {
			pairs := object.NewOrderedMap[int64, any]()
			for _, k := range m.Pairs.Keys {
				mp, _ := m.Pairs.Get(k)
				if mp.Value.Type() == object.MAP_OBJ {
					return nil, fmt.Errorf("blueObjectToGoObject: Map must not have map values. got=%s", mp.Value.Type())
				}
				val, err := blueObjectToGoObject(mp.Value)
				if err != nil {
					return nil, err
				}
				pairs.Set(mp.Key.(*object.Integer).Value, val)
			}
			return pairs, nil
		} else {
			// Floats
			pairs := object.NewOrderedMap[float64, any]()
			for _, k := range m.Pairs.Keys {
				mp, _ := m.Pairs.Get(k)
				if mp.Value.Type() == object.MAP_OBJ {
					return nil, fmt.Errorf("blueObjectToGoObject: Map must not have map values. got=%s", mp.Value.Type())
				}
				val, err := blueObjectToGoObject(mp.Value)
				if err != nil {
					return nil, err
				}
				pairs.Set(mp.Key.(*object.Float).Value, val)
			}
			return pairs, nil
		}
	case object.LIST_OBJ:
		l := blueObject.(*object.List).Elements
		elements := make([]any, len(l))
		for i, e := range l {
			if e.Type() == object.LIST_OBJ {
				return nil, fmt.Errorf("blueObjectToGoObject: List of lists unsupported")
			}
			val, err := blueObjectToGoObject(e)
			if err != nil {
				return nil, err
			}
			elements[i] = val
		}
		return elements, nil
	case object.SET_OBJ:
		s := blueObject.(*object.Set)
		set := object.NewOrderedMap[uint64, object.SetPairGo]()
		for _, k := range s.Elements.Keys {
			v, _ := s.Elements.Get(k)
			val, err := blueObjectToGoObject(v.Value)
			if err != nil {
				return nil, err
			}
			set.Set(k, object.SetPairGo{Value: val, Present: struct{}{}})
		}
		return set, nil
	default:
		return nil, fmt.Errorf("blueObjectToGoObject: TODO: Type currently unsupported: %s (%T)", blueObject.Type(), blueObject)
	}
}
