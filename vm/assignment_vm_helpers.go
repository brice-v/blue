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
	} else if bs, ok := indexable.(*object.BlueStruct); ok {
		return vm.executeStructIndexSetOperator(bs, index, rightValue)
	}
	return fmt.Errorf("'%s' (%T) is not indexable", indexable.Inspect(), indexable)
}

func (vm *VM) executeStructIndexSetOperator(bs *object.BlueStruct, indx, rightValue object.Object) error {
	indexField, ok := indx.(*object.Stringo)
	if !ok {
		return fmt.Errorf("index operator not supported: BLUE_STRUCT.%s", indx.Inspect())
	}
	fieldName := indexField.Value
	orig, origIndex := bs.Get(fieldName)
	if orig == nil {
		return fmt.Errorf("field name `%s` not found on blue struct: %s", fieldName, bs.Inspect())
	}
	return bs.Set(origIndex, rightValue)
}

func (vm *VM) executeMapIndexSetOperator(m *object.Map, indx, rightValue object.Object) error {
	if m.IsEnvBuiltin {
		if indx.Type() != object.STRING_OBJ {
			return fmt.Errorf("ENV requires string key")
		}
		if rightValue.Type() != object.STRING_OBJ && rightValue.Type() != object.NULL_OBJ {
			return fmt.Errorf("ENV requires string value or null")
		}
		k := indx.(*object.Stringo).Value
		if rightValue == object.NULL {
			err := os.Unsetenv(k)
			if err != nil {
				return fmt.Errorf("ENV unset error: %s", err.Error())
			}
		} else {
			v := rightValue.(*object.Stringo).Value
			err := os.Setenv(k, v)
			if err != nil {
				return fmt.Errorf("ENV set error: %s", err.Error())
			}
		}
		object.BuiltinobjsList[object.EnvBuiltinobjsListIndex].Builtin.Obj = object.PopulateENVObj()
		if rightValue == object.NULL {
			return nil
		}
	} else {
		if ok := object.IsHashable(indx); !ok {
			return fmt.Errorf("unusable as a map key: %s", indx.Type())
		}
	}
	hashed := object.HashObject(indx)
	key := object.HashKey{Type: indx.Type(), Value: hashed}
	m.Pairs.Set(key, object.MapPair{Key: indx, Value: rightValue})
	return nil
}

func (vm *VM) executeListIndexSetOperator(l *object.List, indx, rightValue object.Object) error {
	idx, ok := indx.(*object.Integer)
	if !ok {
		return fmt.Errorf("cannot index list with %s", indx.Type())
	}
	indexInt := int(idx.Value)
	listLen := len(l.Elements)
	if indexInt > listLen || indexInt < 0 {
		return fmt.Errorf("index out of bounds: %d", idx.Value)
	}
	if indexInt == listLen {
		l.Elements = append(l.Elements, object.NULL)
	}
	l.Elements[indexInt] = rightValue
	return nil
}

func (vm *VM) executeStringIndexSetOperator(str *object.Stringo, indx, rightValue object.Object) error {
	if rightValue.Type() != object.STRING_OBJ {
		return fmt.Errorf("cannot assign %s to STRING", rightValue.Type())
	}
	if indx.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("cannot index string with %s", indx.Type())
	}
	s := str.Value
	c := rightValue.(*object.Stringo).Value
	indxInt := int(indx.(*object.Integer).Value)
	if runeLen(c) != 1 {
		return fmt.Errorf("string index assignment value must be 1 character long. got=%d", runeLen(c))
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
