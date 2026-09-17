package object

import (
	"blue/ml"
	"fmt"
)

func checkListOfFloats(l *List) bool {
	for _, e := range l.Elements {
		switch e.Type() {
		case LIST_OBJ:
			if !checkListOfFloats(e.(*List)) {
				return false
			}
		case FLOAT_OBJ:
		default:
			return false
		}
	}
	return true
}

type tensorData struct {
	data []float32
}

func inferShape(l *List) []int {
	shape := []int{}
	for {
		shape = append(shape, len(l.Elements))
		if len(l.Elements) == 0 {
			return shape
		}
		first := l.Elements[0]
		if first.Type() != LIST_OBJ {
			return shape
		}
		l = first.(*List)
	}
}

func appendToData(tdata *tensorData, l *List, shape []int) error {
	if len(shape) == 0 {
		return fmt.Errorf("unexpected extra list nesting, got a list where a float was expected")
	}
	if len(l.Elements) != shape[0] {
		return fmt.Errorf("level has %d elements but the inferred shape needs %d", len(l.Elements), shape[0])
	}
	for _, e := range l.Elements {
		switch e.Type() {
		case FLOAT_OBJ:
			if len(shape) != 1 {
				return fmt.Errorf("got a float where a list of length %d was expected", shape[1])
			}
			tdata.data = append(tdata.data, float32(e.(*Float).Value))
		case LIST_OBJ:
			if len(shape) == 1 {
				return fmt.Errorf("got a list where a float was expected")
			}
			if err := appendToData(tdata, e.(*List), shape[1:]); err != nil {
				return err
			}
		default:
			return fmt.Errorf("encountered %s, expected list or float", e.Type())
		}
	}
	return nil
}

var MlBuiltins = []*Builtin{
	{
		Name: "_tensor",
		Fun: func(args ...Object) Object {
			err := checkArgCount("tensor", 4, args)
			if err != nil {
				return err
			}
			if args[0].Type() != LIST_OBJ && args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("tensor", 1, "list or float", args[0].Type())
			}
			// TODO: Evenually support int/bool, for float32 vs 64 it all depends on dtype
			// TODO: or coerce ints to float32 as well
			tdata := &tensorData{
				data: []float32{},
			}
			var shape []int
			if args[0].Type() == FLOAT_OBJ {
				tdata.data = append(tdata.data, float32(args[0].(*Float).Value))
				shape = []int{1}
			} else {
				l := args[0].(*List)
				shape = inferShape(l)
				err := appendToData(tdata, l, shape)
				if err != nil {
					return newError("`tensor` error: %s", err.Error())
				}
			}
			err = checkArgType("tensor", 2, STRING_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("tensor", 3, STRING_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("tensor", 4, BOOLEAN_OBJ, args)
			if err != nil {
				return err
			}
			dtypeS := args[1].(*Stringo).Value
			dtype, derr := ml.ParseDType(dtypeS)
			if derr != nil {
				return newError("`tensor` error: %s", derr.Error())
			}
			deviceS := args[2].(*Stringo).Value
			device, derr := ml.ParseDevice(deviceS)
			if derr != nil {
				return newError("`tensor` error: %s", derr.Error())
			}
			requiresGrad := args[3].(*Boolean).Value
			tt, terr := ml.NewTensor(tdata.data, shape, dtype, device)
			if terr != nil {
				return newError("`tensor` error: %s", terr.Error())
			}
			tt.SetRequiresGrad(requiresGrad)
			return &Tensor{T: tt}
		},
		HelpStr: helpStrArgs{
			explanation: "`tensor` returns a TENSOR constructed from the provided lists",
			signature:   "tensor(list) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "tensor() => Tensor{shape: [2,2], strides: TODO}",
		}.String(),
	},
}
