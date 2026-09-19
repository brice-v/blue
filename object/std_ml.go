package object

import (
	"blue/ml"
	"fmt"
	"math"
	"slices"
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
			// TODO: eventually support int/bool leaves; float32 vs float64 depends on dtype
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
			explanation: "`tensor` builds a tensor from nested lists, inferring the shape from the nesting",
			signature:   "tensor(data: list, datatype: str='float32', dev: str='cpu', requires_grad: bool=false) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "tensor([[1.0, 2.0], [3.0, 4.0]]) => Tensor{shape: [2 2]}",
		}.String(),
	},
	{
		Name: "_matmul",
		Fun:  tensorBinaryBuiltin("matmul", ml.MatMul),
		HelpStr: helpStrArgs{
			explanation: "`matmul` returns the matrix product of two tensors",
			signature:   "matmul(a: tensor, b: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "matmul(tensor([[1.0, 2.0]]), tensor([[3.0], [4.0]])) => Tensor{shape: [1 1]}",
		}.String(),
	},
	{
		Name: "_add",
		Fun:  tensorBinaryBuiltin("add", ml.Add),
		HelpStr: helpStrArgs{
			explanation: "`add` returns the elementwise sum of two tensors, with scalar broadcast",
			signature:   "add(a: tensor, b: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "add(tensor([[1.0, 2.0]]), tensor([[3.0, 4.0]])) => Tensor{shape: [1 2]}",
		}.String(),
	},
	{
		Name: "_sub",
		Fun:  tensorBinaryBuiltin("sub", ml.Sub),
		HelpStr: helpStrArgs{
			explanation: "`sub` returns the elementwise difference of two tensors, with scalar broadcast",
			signature:   "sub(a: tensor, b: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "sub(tensor([[1.0, 2.0]]), tensor([[1.0, 1.0]])) => Tensor{shape: [1 2]}",
		}.String(),
	},
	{
		Name: "_mul",
		Fun:  tensorBinaryBuiltin("mul", ml.Mul),
		HelpStr: helpStrArgs{
			explanation: "`mul` returns the elementwise product of two tensors, with scalar broadcast",
			signature:   "mul(a: tensor, b: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "mul(tensor([[1.0, 2.0]]), tensor([[3.0, 4.0]])) => Tensor{shape: [1 2]}",
		}.String(),
	},
	{
		Name: "_div",
		Fun:  tensorBinaryBuiltin("div", ml.Div),
		HelpStr: helpStrArgs{
			explanation: "`div` returns the elementwise quotient of two tensors, with scalar broadcast",
			signature:   "div(a: tensor, b: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "div(tensor([[2.0, 4.0]]), tensor([[2.0, 2.0]])) => Tensor{shape: [1 2]}",
		}.String(),
	},
	{
		Name: "_pow",
		Fun:  tensorBinaryBuiltin("pow", ml.Pow),
		HelpStr: helpStrArgs{
			explanation: "`pow` returns the elementwise power of two tensors, with scalar broadcast; a negative base is allowed when the exponent is an integer",
			signature:   "pow(a: tensor, b: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "pow(tensor([[2.0, 4.0]]), tensor([[2.0, 2.0]])) => Tensor{shape: [1 2]}",
		}.String(),
	},
	{
		Name: "_neg",
		Fun:  tensorUnaryBuiltin("neg", ml.Neg),
		HelpStr: helpStrArgs{
			explanation: "`neg` returns the negation of each element",
			signature:   "neg(a: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "neg(tensor([[1.0, -2.0]])) => Tensor{shape: [1 2]}",
		}.String(),
	},
	{
		Name: "_relu",
		Fun:  tensorUnaryBuiltin("relu", ml.Relu),
		HelpStr: helpStrArgs{
			explanation: "`relu` returns max(0, x) elementwise",
			signature:   "relu(a: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "relu(tensor([[-1.0, 2.0]])) => Tensor{shape: [1 2]}",
		}.String(),
	},
	{
		Name: "_exp",
		Fun:  tensorUnaryBuiltin("exp", ml.Exp),
		HelpStr: helpStrArgs{
			explanation: "`exp` returns e to the power of each element",
			signature:   "exp(a: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "exp(tensor([[0.0]])) => Tensor{shape: [1 1]}",
		}.String(),
	},
	{
		Name: "_log",
		Fun:  tensorUnaryBuiltin("log", ml.Log),
		HelpStr: helpStrArgs{
			explanation: "`log` returns the natural logarithm of each element",
			signature:   "log(a: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "log(tensor([[1.0]])) => Tensor{shape: [1 1]}",
		}.String(),
	},
	{
		Name: "_sqrt",
		Fun:  tensorUnaryBuiltin("sqrt", ml.Sqrt),
		HelpStr: helpStrArgs{
			explanation: "`sqrt` returns the square root of each element",
			signature:   "sqrt(a: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "sqrt(tensor([[4.0]])) => Tensor{shape: [1 1]}",
		}.String(),
	},
	{
		Name: "_eq",
		Fun:  tensorBinaryBuiltin("eq", ml.Eq),
		HelpStr: helpStrArgs{
			explanation: "`eq` returns a bool tensor that is true where a == b, with scalar broadcast; comparisons are not differentiable",
			signature:   "eq(a: tensor, b: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "eq(tensor([[1.0, 3.0]]), tensor([[2.0, 3.0]])) => Tensor{shape: [1 2]}",
		}.String(),
	},
	{
		Name: "_ne",
		Fun:  tensorBinaryBuiltin("ne", ml.Ne),
		HelpStr: helpStrArgs{
			explanation: "`ne` returns a bool tensor that is true where a != b, with scalar broadcast; comparisons are not differentiable",
			signature:   "ne(a: tensor, b: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "ne(tensor([[1.0, 3.0]]), tensor([[1.0, 3.0]])) => Tensor{shape: [1 2]}",
		}.String(),
	},
	{
		Name: "_gt",
		Fun:  tensorBinaryBuiltin("gt", ml.Gt),
		HelpStr: helpStrArgs{
			explanation: "`gt` returns a bool tensor that is true where a > b, with scalar broadcast; comparisons are not differentiable",
			signature:   "gt(a: tensor, b: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "gt(tensor([[1.0, 3.0]]), tensor([[2.0, 2.0]])) => Tensor{shape: [1 2]}",
		}.String(),
	},
	{
		Name: "_ge",
		Fun:  tensorBinaryBuiltin("ge", ml.Ge),
		HelpStr: helpStrArgs{
			explanation: "`ge` returns a bool tensor that is true where a >= b, with scalar broadcast; comparisons are not differentiable",
			signature:   "ge(a: tensor, b: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "ge(tensor([[1.0, 3.0]]), tensor([[2.0, 2.0]])) => Tensor{shape: [1 2]}",
		}.String(),
	},
	{
		Name: "_lt",
		Fun:  tensorBinaryBuiltin("lt", ml.Lt),
		HelpStr: helpStrArgs{
			explanation: "`lt` returns a bool tensor that is true where a < b, with scalar broadcast; comparisons are not differentiable",
			signature:   "lt(a: tensor, b: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "lt(tensor([[1.0, 3.0]]), tensor([[2.0, 2.0]])) => Tensor{shape: [1 2]}",
		}.String(),
	},
	{
		Name: "_le",
		Fun:  tensorBinaryBuiltin("le", ml.Le),
		HelpStr: helpStrArgs{
			explanation: "`le` returns a bool tensor that is true where a <= b, with scalar broadcast; comparisons are not differentiable",
			signature:   "le(a: tensor, b: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "le(tensor([[1.0, 3.0]]), tensor([[2.0, 2.0]])) => Tensor{shape: [1 2]}",
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
			explanation: "`reshape` returns a view of the tensor with a new shape; the element count must match",
			signature:   "reshape(a: tensor, shape: list[int]) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "reshape(tensor([[1.0, 2.0], [3.0, 4.0]]), [4]) => Tensor{shape: [4]}",
		}.String(),
	},
	{
		Name: "_transpose",
		Fun: func(args ...Object) Object {
			err := checkArgCount("transpose", 3, args)
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
			explanation: "`transpose` returns a view with two dimensions swapped (metadata only)",
			signature:   "transpose(a: tensor, dim0: int=0, dim1: int=1) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "transpose(tensor([[1.0, 2.0], [3.0, 4.0]]), 0, 1) => Tensor{shape: [2 2]}",
		}.String(),
	},
	{
		Name: "_equal",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("equal", 2, args); err != nil {
				return err
			}
			if err := checkArgType("equal", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			if err := checkArgType("equal", 2, TENSOR_OBJ, args); err != nil {
				return err
			}
			a, b := args[0].(*Tensor).T, args[1].(*Tensor).T
			if !slices.Equal(a.Shape(), b.Shape()) || a.DType() != b.DType() {
				return FALSE
			}
			ad, bd := a.ContiguousData(), b.ContiguousData()
			for i := range ad {
				if ad[i] != bd[i] {
					return FALSE
				}
			}
			return TRUE
		},
		HelpStr: helpStrArgs{
			explanation: "`equal` returns true when two tensors have the same shape, dtype, and elements; this is the value check to use in tests, since `==` is the elementwise comparison operator",
			signature:   "equal(a: tensor, b: tensor) -> bool",
			errors:      "InvalidArgCount,PositionalType",
			example:     "equal(tensor([[1.0, 2.0]]), tensor([[1.0, 2.0]])) => true",
		}.String(),
	},
	{
		Name: "_allclose",
		Fun: func(args ...Object) Object {
			err := checkArgCount("allclose", 4, args)
			if err != nil {
				return err
			}
			err = checkArgType("allclose", 1, TENSOR_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("allclose", 2, TENSOR_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("allclose", 3, FLOAT_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("allclose", 4, FLOAT_OBJ, args)
			if err != nil {
				return err
			}
			a, b := args[0].(*Tensor).T, args[1].(*Tensor).T
			rtol, atol := args[2].(*Float).Value, args[3].(*Float).Value
			if !slices.Equal(a.Shape(), b.Shape()) {
				return FALSE
			}
			ad, bd := a.ContiguousData(), b.ContiguousData()
			for i := range ad {
				diff := math.Abs(float64(ad[i]) - float64(bd[i]))
				limit := atol + rtol*math.Abs(float64(bd[i]))
				if diff > limit {
					return FALSE
				}
			}
			return TRUE
		},
		HelpStr: helpStrArgs{
			explanation: "`allclose` returns true when two tensors have the same shape and every element satisfies |a - b| <= atol + rtol*|b|; use it for computed results such as softmax, exp, and sqrt",
			signature:   "allclose(a: tensor, b: tensor, rtol: float=1e-5, atol: float=1e-8) -> bool",
			errors:      "InvalidArgCount,PositionalType",
			example:     "allclose(tensor([[1.0]]), tensor([[1.000001]])) => true",
		}.String(),
	},
	{
		Name: "_sum",
		Fun: func(args ...Object) Object {
			err := checkArgCount("sum", 3, args)
			if err != nil {
				return err
			}
			err = checkArgType("sum", 1, TENSOR_OBJ, args)
			if err != nil {
				return err
			}
			return tensorSum(args[0].(*Tensor).T, args[1], args[2])
		},
	},
	{
		Name: "_mean",
		Fun:  reductionBuiltin("mean", ml.Mean),
		HelpStr: helpStrArgs{
			explanation: "`mean` returns the average over `dim`, or over every element when `dim` is null",
			signature:   "mean(a: tensor, dim: int|list[int]|null=null, keepdim: bool=false) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "mean(tensor([[1.0, 2.0], [3.0, 4.0]]), 0) => Tensor{shape: [2]}",
		}.String(),
	},
	{
		Name: "_max",
		Fun:  reductionBuiltin("max", ml.Max),
		HelpStr: helpStrArgs{
			explanation: "`max` returns the maximum over `dim`, or over every element when `dim` is null",
			signature:   "max(a: tensor, dim: int|list[int]|null=null, keepdim: bool=false) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "max(tensor([[1.0, 3.0]]), 1) => Tensor{shape: [1]}",
		}.String(),
	},
	{
		Name: "_min",
		Fun:  reductionBuiltin("min", ml.Min),
		HelpStr: helpStrArgs{
			explanation: "`min` returns the minimum over `dim`, or over every element when `dim` is null",
			signature:   "min(a: tensor, dim: int|list[int]|null=null, keepdim: bool=false) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "min(tensor([[1.0, 3.0]]), 1) => Tensor{shape: [1]}",
		}.String(),
	},
	{
		Name: "_argmax",
		Fun:  argBuiltin("argmax", ml.ArgMax),
		HelpStr: helpStrArgs{
			explanation: "`argmax` returns the index of the maximum along `dim`",
			signature:   "argmax(a: tensor, dim: int) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "argmax(tensor([[1.0, 3.0]]), 1) => Tensor{shape: [1]}",
		}.String(),
	},
	{
		Name: "_argmin",
		Fun:  argBuiltin("argmin", ml.ArgMin),
		HelpStr: helpStrArgs{
			explanation: "`argmin` returns the index of the minimum along `dim`",
			signature:   "argmin(a: tensor, dim: int) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "argmin(tensor([[1.0, 3.0]]), 1) => Tensor{shape: [1]}",
		}.String(),
	},
	{
		Name: "_softmax",
		Fun:  argBuiltin("softmax", ml.Softmax),
		HelpStr: helpStrArgs{
			explanation: "`softmax` returns a numerically stable softmax along `dim`",
			signature:   "softmax(a: tensor, dim: int) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "softmax(tensor([[1.0, 2.0, 3.0]]), 1) => Tensor{shape: [1 3]}",
		}.String(),
	},
}

// asTensorArg accepts either a tensor or a scalar (int, float, or bool), which
// is coerced to a one element tensor on device. This lets the module functions
// take scalars the way PyTorch does, for example `ml.gt(a, 0.0)`.
func asTensorArg(o Object, device ml.Device) (*ml.Tensor, bool) {
	if t, ok := o.(*Tensor); ok {
		return t.T, true
	}
	var v float32
	switch n := o.(type) {
	case *Integer:
		v = float32(n.Value)
	case *Float:
		v = float32(n.Value)
	case *Boolean:
		if n.Value {
			v = 1
		}
	default:
		return nil, false
	}
	t, err := ml.NewTensor([]float32{v}, []int{1}, ml.Float32, device)
	if err != nil {
		return nil, false
	}
	return t, true
}

func tensorBinaryBuiltin(name string, f func(a, b *ml.Tensor) (*ml.Tensor, error)) func(...Object) Object {
	return func(args ...Object) Object {
		err := checkArgCount(name, 2, args)
		if err != nil {
			return err
		}
		// a scalar on either side is built on the other operand's device
		device := ml.CPU
		if t, ok := args[0].(*Tensor); ok {
			device = t.T.Device()
		} else if t, ok := args[1].(*Tensor); ok {
			device = t.T.Device()
		}
		a, ok := asTensorArg(args[0], device)
		if !ok {
			return newPositionalTypeError(name, 1, TENSOR_OBJ, args[0].Type())
		}
		b, ok := asTensorArg(args[1], device)
		if !ok {
			return newPositionalTypeError(name, 2, TENSOR_OBJ, args[1].Type())
		}
		out, ferr := f(a, b)
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

// tensorReduction is the shared code of the module-level reductions.
// dim/keepdim are the raw arguments, nil means unset.
func tensorReduction(name string, t *ml.Tensor, dim, keepdim Object,
	f func(*ml.Tensor, []int, bool) (*ml.Tensor, error)) Object {
	dims, err := dimsFromObject(name, dim)
	if err != nil {
		return err
	}
	keep := false
	if keepdim != nil {
		if keepdim.Type() != BOOLEAN_OBJ {
			return newPositionalTypeError(name, 2, BOOLEAN_OBJ, keepdim.Type())
		}
		keep = keepdim.(*Boolean).Value
	}
	out, serr := f(t, dims, keep)
	if serr != nil {
		return newError("`%s` error: %s", name, serr.Error())
	}
	return &Tensor{T: out}
}

func tensorSum(t *ml.Tensor, dim, keepdim Object) Object {
	return tensorReduction("sum", t, dim, keepdim, ml.Sum)
}

func reductionBuiltin(name string, f func(*ml.Tensor, []int, bool) (*ml.Tensor, error)) func(...Object) Object {
	return func(args ...Object) Object {
		if err := checkArgCount(name, 3, args); err != nil {
			return err
		}
		if err := checkArgType(name, 1, TENSOR_OBJ, args); err != nil {
			return err
		}
		return tensorReduction(name, args[0].(*Tensor).T, args[1], args[2], f)
	}
}

// argBuiltin builds a single-dim index reduction like argmax(a, dim).
func argBuiltin(name string, f func(*ml.Tensor, int) (*ml.Tensor, error)) func(...Object) Object {
	return func(args ...Object) Object {
		if err := checkArgCount(name, 2, args); err != nil {
			return err
		}
		if err := checkArgType(name, 1, TENSOR_OBJ, args); err != nil {
			return err
		}
		d, ok := args[1].(*Integer)
		if !ok {
			return newPositionalTypeError(name, 2, INTEGER_OBJ, args[1].Type())
		}
		out, err := f(args[0].(*Tensor).T, int(d.Value))
		if err != nil {
			return newError("`%s` error: %s", name, err.Error())
		}
		return &Tensor{T: out}
	}
}

func dimsFromObject(name string, o Object) ([]int, Object) {
	switch v := o.(type) {
	case nil, *Null:
		return nil, nil
	case *Integer:
		return []int{int(v.Value)}, nil
	case *List:
		dims := make([]int, len(v.Elements))
		for i, e := range v.Elements {
			if e.Type() != INTEGER_OBJ {
				return nil, newPositionalTypeError(name, 2, INTEGER_OBJ, e.Type())
			}
			dims[i] = int(e.(*Integer).Value)
		}
		return dims, nil
	default:
		return nil, newPositionalTypeError(name, 2, INTEGER_OBJ, o.Type())
	}
}
