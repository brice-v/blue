package vm

import (
	"blue/object"
	"fmt"
	"os"
	"strings"
)

func (vm *VM) executeIndexSetOperator(indexable object.Object, index object.Object, rightValue object.Object) error {
	if m, ok := indexable.(*object.Map); ok {
		return vm.executeMapIndexSetOperator(m, index, rightValue)
	} else if l, ok := indexable.(*object.List); ok {
		return vm.executeListIndexSetOperator(l, index, rightValue)
	} else if s, ok := indexable.(*object.Stringo); ok {
		return vm.executeStringIndexSetOperator(s, index, rightValue)
	}
	return fmt.Errorf("'%s' (%T) is not indexable", indexable.Inspect(), indexable)
}

func (vm *VM) executeMapIndexSetOperator(m *object.Map, indx object.Object, rightValue object.Object) error {
	if m.IsEnvBuiltin {
		if indx.Type() != object.STRING_OBJ {
			return vm.push(newError("ENV requires string key"))
		}
		if rightValue.Type() != object.STRING_OBJ && rightValue.Type() != object.NULL_OBJ {
			return vm.push(newError("ENV requires string value or null"))
		}
		k := indx.(*object.Stringo).Value
		if rightValue == object.NULL {
			err := os.Unsetenv(k)
			if err != nil {
				return vm.push(newError("ENV unset error: %s", err.Error()))
			}
		} else {
			v := rightValue.(*object.Stringo).Value
			err := os.Setenv(k, v)
			if err != nil {
				return vm.push(newError("ENV set error: %s", err.Error()))
			}
		}
		object.BuiltinobjsList[object.EnvBuiltinobjsListIndex].Builtin.Obj = object.PopulateENVObj()
		if rightValue == object.NULL {
			return nil
		}
	} else {
		if ok := object.IsHashable(indx); !ok {
			return vm.push(newError("unusable as a map key: %s", indx.Type()))
		}
	}
	hashed := object.HashObject(indx)
	key := object.HashKey{Type: indx.Type(), Value: hashed}
	m.Pairs.Set(key, object.MapPair{Key: indx, Value: rightValue})
	return nil
}

func (vm *VM) executeListIndexSetOperator(l *object.List, indx object.Object, rightValue object.Object) error {
	idx, ok := indx.(*object.Integer)
	if !ok {
		return vm.push(newError("cannot index list with %s", indx.Type()))
	}
	indexInt := int(idx.Value)
	listLen := len(l.Elements)
	if indexInt > listLen || indexInt < 0 {
		return vm.push(newError("index out of bounds: %d", idx.Value))
	}
	if indexInt == listLen {
		l.Elements = append(l.Elements, object.NULL)
	}
	l.Elements[indexInt] = rightValue
	return nil
}

func (vm *VM) executeStringIndexSetOperator(str *object.Stringo, indx object.Object, rightValue object.Object) error {
	if rightValue.Type() != object.STRING_OBJ {
		return vm.push(newError("cannot assign %s to STRING", rightValue.Type()))
	}
	if indx.Type() != object.INTEGER_OBJ {
		return vm.push(newError("cannot index string with %s", indx.Type()))
	}
	s := str.Value
	c := rightValue.(*object.Stringo).Value
	indxInt := int(indx.(*object.Integer).Value)
	if runeLen(c) != 1 {
		return vm.push(newError("string index assignment value must be 1 character long. got=%d", runeLen(c)))
	}
	sb := strings.Builder{}
	for i, ch := range s {
		if i == indxInt {
			sb.WriteString(c)
		} else {
			sb.WriteRune(ch)
		}
	}
	str.Value = sb.String()
	return nil
}
