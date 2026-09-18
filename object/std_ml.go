package object

import (
	"blue/ml"
	"fmt"
)

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
			example:     "TODO",
		}.String(),
	},
	{
		Name: "_matmul",
		Fun:  tensorBinaryBuiltin("matmul", ml.MatMul),
		HelpStr: helpStrArgs{
			explanation: "`matmul` returns tensor @ tensor",
			signature:   "matmul(a: tensor, b: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "TODO",
		}.String(),
	},
	{
		Name: "_add",
		Fun:  tensorBinaryBuiltin("add", ml.Add),
		HelpStr: helpStrArgs{
			explanation: "`add` returns tensor + tensor",
			signature:   "add(a: tensor, b: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "TODO",
		}.String(),
	},
	{
		Name: "_sub",
		Fun:  tensorBinaryBuiltin("sub", ml.Sub),
		HelpStr: helpStrArgs{
			explanation: "`sub` returns tensor - tensor",
			signature:   "sub(a: tensor, b: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "TODO",
		}.String(),
	},
	{
		Name: "_mul",
		Fun:  tensorBinaryBuiltin("mul", ml.Mul),
		HelpStr: helpStrArgs{
			explanation: "`mul` returns tensor * tensor",
			signature:   "mul(a: tensor, b: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "TODO",
		}.String(),
	},
	{
		Name: "_div",
		Fun:  tensorBinaryBuiltin("div", ml.Div),
		HelpStr: helpStrArgs{
			explanation: "`div` returns tensor / tensor",
			signature:   "div(a: tensor, b: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "TODO",
		}.String(),
	},
	{
		Name: "_relu",
		Fun:  tensorUnaryBuiltin("relu", ml.Relu),
		HelpStr: helpStrArgs{
			explanation: "`relu` returns tensor.relu()",
			signature:   "relu(a: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "TODO",
		}.String(),
	},
	{
		Name: "_exp",
		Fun:  tensorUnaryBuiltin("exp", ml.Exp),
		HelpStr: helpStrArgs{
			explanation: "`exp` returns tensor.exp()",
			signature:   "relu(a: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "TODO",
		}.String(),
	},
	{
		Name: "_log",
		Fun:  tensorUnaryBuiltin("log", ml.Log),
		HelpStr: helpStrArgs{
			explanation: "`log` returns tensor.log()",
			signature:   "log(a: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "TODO",
		}.String(),
	},
	{
		Name: "_sqrt",
		Fun:  tensorUnaryBuiltin("sqrt", ml.Sqrt),
		HelpStr: helpStrArgs{
			explanation: "`sqrt` returns tensor.sqrt()",
			signature:   "sqrt(a: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "TODO",
		}.String(),
	},
	{
		Name: "_reshape",
		Fun: func(args ...Object) Object {
			err := checkArgCount("reshape", 2, args)
			if err != nil {
				return err
			}
			err = checkArgType("reshape", 1, TENSOR_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("reshape", 2, LIST_OBJ, args)
			if err != nil {
				return err
			}
			is, ierr := toIntList("reshape", args[1].(*List))
			if ierr != nil {
				return newError("%s", ierr.Error())
			}
			out, terr := ml.Reshape(args[0].(*Tensor).T, is...)
			if terr != nil {
				return newError("`reshape` error: %s", terr.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`reshape` TODO",
			signature:   "reshape(a: tensor, shape: list[int]) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "TODO",
		}.String(),
	},
	{
		Name: "_transpose",
		Fun: func(args ...Object) Object {
			err := checkArgCount("transpose", 2, args)
			if err != nil {
				return err
			}
			err = checkArgType("transpose", 1, TENSOR_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("transpose", 2, INTEGER_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("transpose", 3, INTEGER_OBJ, args)
			if err != nil {
				return err
			}
			out, terr := ml.Transpose(args[0].(*Tensor).T, int(args[1].(*Integer).Value), int(args[2].(*Integer).Value))
			if terr != nil {
				return newError("`transpose` error: %s", terr.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`transpose` TODO",
			signature:   "transpose(a: tensor, dim0: int, dim1: int) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "TODO",
		}.String(),
	},
}

func tensorBinaryBuiltin(name string, f func(a, b *ml.Tensor) (*ml.Tensor, error)) func(...Object) Object {
	return func(args ...Object) Object {
		err := checkArgCount(name, 2, args)
		if err != nil {
			return err
		}
		err = checkArgType(name, 1, TENSOR_OBJ, args)
		if err != nil {
			return err
		}
		err = checkArgType(name, 2, TENSOR_OBJ, args)
		if err != nil {
			return err
		}
		out, ferr := f(args[0].(*Tensor).T, args[1].(*Tensor).T)
		if ferr != nil {
			return newError("`%s` error: %s", name, ferr.Error())
		}
		return &Tensor{T: out}
	}
}

func tensorUnaryBuiltin(name string, f func(a *ml.Tensor) (*ml.Tensor, error)) func(...Object) Object {
	return func(args ...Object) Object {
		err := checkArgCount(name, 1, args)
		if err != nil {
			return err
		}
		err = checkArgType(name, 1, TENSOR_OBJ, args)
		if err != nil {
			return err
		}
		out, ferr := f(args[0].(*Tensor).T)
		if ferr != nil {
			return newError("`%s` error: %s", name, ferr.Error())
		}
		return &Tensor{T: out}
	}
}

func toIntList(name string, l *List) ([]int, error) {
	is := make([]int, len(l.Elements))
	for i, e := range l.Elements {
		if e.Type() != INTEGER_OBJ {
			return nil, fmt.Errorf("`%s` error: expected INTEGER in list, found %s", name, e.Type())
		}
		is[i] = int(e.(*Integer).Value)
	}
	return is, nil
}
