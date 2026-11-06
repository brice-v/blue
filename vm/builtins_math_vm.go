package vm

import (
	"blue/object"
	"math/rand"
	"time"
)

func mathVmBuiltin(name string, vm *VM) *object.Builtin {
	switch name {
	case "_rand":
		return createMathRandBuiltin(vm)
	case "_seed":
		return createMathSeedBuiltin(vm)
	case "_gauss":
		return createMathGaussBuiltin(vm)
	case "_weighted_choice":
		return createMathWeightedChoiceBuiltin(vm)
	default:
		return nil
	}
}

func (vm *VM) mathRNG() *rand.Rand {
	if vm.mathRand == nil {
		vm.mathRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	return vm.mathRand
}

func createMathRandBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "_rand",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("rand", len(args), 0, "")
			}
			return &object.Float{Value: vm.mathRNG().Float64()}
		},
	}
}

func createMathSeedBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "_seed",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("seed", len(args), 1, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("seed", 1, object.INTEGER_OBJ, args[0].Type())
			}
			vm.mathRand = rand.New(rand.NewSource(args[0].(*object.Integer).Value))
			return object.TRUE
		},
	}
}

func createMathGaussBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "_gauss",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("gauss", len(args), 2, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("gauss", 1, object.FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("gauss", 2, object.FLOAT_OBJ, args[1].Type())
			}
			mean := args[0].(*object.Float).Value
			stddev := args[1].(*object.Float).Value
			v := vm.mathRNG().NormFloat64()
			return &object.Float{Value: mean + stddev*v}
		},
	}
}

func createMathWeightedChoiceBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "_weighted_choice",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("weighted_choice", len(args), 2, "")
			}
			items := args[0]
			weights := args[1]
			if items.Type() != object.LIST_OBJ {
				return newPositionalTypeError("weighted_choice", 1, object.LIST_OBJ, items.Type())
			}
			if weights.Type() != object.LIST_OBJ {
				return newPositionalTypeError("weighted_choice", 2, object.LIST_OBJ, weights.Type())
			}
			itemList := items.(*object.List)
			weightList := weights.(*object.List)
			if len(itemList.Elements) != len(weightList.Elements) {
				return newError("`weighted_choice` error: items and weights must have the same length")
			}
			var totalWeight float64
			for _, w := range weightList.Elements {
				if w.Type() != object.FLOAT_OBJ {
					return newError("`weighted_choice` error: all weights must be FLOAT")
				}
				totalWeight += w.(*object.Float).Value
			}
			if totalWeight == 0 {
				return newError("`weighted_choice` error: weight sum must not be zero")
			}
			r := vm.mathRNG().Float64() * totalWeight
			var cumulative float64
			for i, w := range weightList.Elements {
				cumulative += w.(*object.Float).Value
				if r <= cumulative {
					return itemList.Elements[i]
				}
			}
			return itemList.Elements[len(itemList.Elements)-1]
		},
	}
}
