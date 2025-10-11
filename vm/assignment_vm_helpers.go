package vm

import (
	"blue/object"
	"fmt"
)

func (vm *VM) executeIndexSetOperator(indexable object.Object, index object.Object, rightValue object.Object) error {
	if m, ok := indexable.(*object.Map); ok {
		return vm.executeMapIndexSetOperator(m, index, rightValue)
	}
	return fmt.Errorf("'%s' (%T) is not indexable", indexable.Inspect(), indexable)
}

func (vm *VM) executeMapIndexSetOperator(m *object.Map, indx object.Object, rightValue object.Object) error {
	if ok := object.IsHashable(indx); !ok {
		return vm.push(newError("unusable as a map key: %s", indx.Type()))
	}
	hashed := object.HashObject(indx)
	key := object.HashKey{Type: indx.Type(), Value: hashed}
	m.Pairs.Set(key, object.MapPair{Key: indx, Value: rightValue})
	return nil
}
