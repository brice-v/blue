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
	{
		Name: "_abs",
		Fun:  tensorUnaryBuiltin("abs", ml.Abs),
		HelpStr: helpStrArgs{
			explanation: "`abs` returns the absolute value of each element",
			signature:   "abs(a: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "abs(tensor([[-1.0, 2.0]])) => Tensor{shape: [1 2]}",
		}.String(),
	},
	{
		Name: "_sigmoid",
		Fun:  tensorUnaryBuiltin("sigmoid", ml.Sigmoid),
		HelpStr: helpStrArgs{
			explanation: "`sigmoid` returns 1/(1+exp(-x)) elementwise",
			signature:   "sigmoid(a: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "sigmoid(tensor([0.0])) => Tensor{shape: [1]}",
		}.String(),
	},
	{
		Name: "_tanh",
		Fun:  tensorUnaryBuiltin("tanh", ml.Tanh),
		HelpStr: helpStrArgs{
			explanation: "`tanh` returns the hyperbolic tangent elementwise",
			signature:   "tanh(a: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "tanh(tensor([0.0])) => Tensor{shape: [1]}",
		}.String(),
	},
	{
		Name: "_flatten",
		Fun:  tensorUnaryBuiltin("flatten", ml.Flatten),
		HelpStr: helpStrArgs{
			explanation: "`flatten` returns a 1d view of the tensor",
			signature:   "flatten(a: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "flatten(tensor([[1.0, 2.0]])) => Tensor{shape: [2]}",
		}.String(),
	},
	{
		Name: "_unsqueeze",
		Fun:  argBuiltin("unsqueeze", ml.Unsqueeze),
		HelpStr: helpStrArgs{
			explanation: "`unsqueeze` inserts a size-1 dimension at `dim`",
			signature:   "unsqueeze(a: tensor, dim: int) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "unsqueeze(tensor([1.0, 2.0]), 0) => Tensor{shape: [1 2]}",
		}.String(),
	},
	{
		Name: "_permute",
		Fun: listBuiltin("permute", func(a *ml.Tensor, dims []int) (*ml.Tensor, error) {
			return ml.Permute(a, dims...)
		}),
		HelpStr: helpStrArgs{
			explanation: "`permute` reorders the dimensions",
			signature:   "permute(a: tensor, dims: list[int]) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "permute(tensor([[1.0, 2.0]]), [1, 0]) => Tensor{shape: [2 1]}",
		}.String(),
	},
	{
		Name: "_broadcast_to",
		Fun:  listBuiltin("broadcast_to", ml.BroadcastTo),
		HelpStr: helpStrArgs{
			explanation: "`broadcast_to` expands size-1 dims to the given shape",
			signature:   "broadcast_to(a: tensor, shape: list[int]) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "broadcast_to(tensor([[1.0, 2.0]]), [3, 2]) => Tensor{shape: [3 2]}",
		}.String(),
	},
	{
		Name: "_squeeze",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("squeeze", 2, args); err != nil {
				return err
			}
			if err := checkArgType("squeeze", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			var dims []int
			if args[1].Type() != NULL_OBJ {
				parsed, errObj := dimsFromObject("squeeze", args[1])
				if errObj != nil {
					return errObj
				}
				dims = parsed
			}
			out, err := ml.Squeeze(args[0].(*Tensor).T, dims)
			if err != nil {
				return newError("`squeeze` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`squeeze` removes size-1 dimensions; `dim` null removes every size-1 dim",
			signature:   "squeeze(a: tensor, dim: int|list[int]|null=null) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "squeeze(tensor([[[1.0, 2.0]]])) => Tensor{shape: [2]}",
		}.String(),
	},
	{
		Name: "_zeros",
		Fun:  creationBuiltin("zeros", ml.Zeros),
		HelpStr: helpStrArgs{
			explanation: "`zeros` returns a tensor of zeros with the given shape",
			signature:   "zeros(shape: list[int], datatype: str='float32', dev: str='cpu', requires_grad: bool=false) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "zeros([2, 2]) => Tensor{shape: [2 2]}",
		}.String(),
	},
	{
		Name: "_ones",
		Fun:  creationBuiltin("ones", ml.Ones),
		HelpStr: helpStrArgs{
			explanation: "`ones` returns a tensor of ones with the given shape",
			signature:   "ones(shape: list[int], datatype: str='float32', dev: str='cpu', requires_grad: bool=false) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "ones([2, 2]) => Tensor{shape: [2 2]}",
		}.String(),
	},
	{
		Name: "_randn",
		Fun:  creationBuiltin("randn", ml.Randn),
		HelpStr: helpStrArgs{
			explanation: "`randn` returns a tensor of standard normal samples",
			signature:   "randn(shape: list[int], datatype: str='float32', dev: str='cpu', requires_grad: bool=false) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "randn([2, 2]) => Tensor{shape: [2 2]}",
		}.String(),
	},
	{
		Name: "_full",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("full", 5, args); err != nil {
				return err
			}
			if err := checkArgType("full", 1, LIST_OBJ, args); err != nil {
				return err
			}
			v, ok := toFloatArg(args[1])
			if !ok {
				return newPositionalTypeError("full", 2, FLOAT_OBJ, args[1].Type())
			}
			if err := checkArgType("full", 3, STRING_OBJ, args); err != nil {
				return err
			}
			if err := checkArgType("full", 4, STRING_OBJ, args); err != nil {
				return err
			}
			if err := checkArgType("full", 5, BOOLEAN_OBJ, args); err != nil {
				return err
			}
			shape, err := toIntList("full", args[0].(*List))
			if err != nil {
				return newError("%s", err.Error())
			}
			dtype, derr := ml.ParseDType(args[2].(*Stringo).Value)
			if derr != nil {
				return newError("`full` error: %s", derr.Error())
			}
			device, derr := ml.ParseDevice(args[3].(*Stringo).Value)
			if derr != nil {
				return newError("`full` error: %s", derr.Error())
			}
			tt, ferr := ml.Full(shape, v, dtype, device)
			if ferr != nil {
				return newError("`full` error: %s", ferr.Error())
			}
			tt.SetRequiresGrad(args[4].(*Boolean).Value)
			return &Tensor{T: tt}
		},
		HelpStr: helpStrArgs{
			explanation: "`full` returns a tensor filled with a value",
			signature:   "full(shape: list[int], fill_value: float, datatype: str='float32', dev: str='cpu', requires_grad: bool=false) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "full([2, 2], 5.0) => Tensor{shape: [2 2]}",
		}.String(),
	},
	{
		Name: "_arange",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("arange", 5, args); err != nil {
				return err
			}
			start, ok := toFloatArg(args[0])
			if !ok {
				return newPositionalTypeError("arange", 1, FLOAT_OBJ, args[0].Type())
			}
			end, ok := toFloatArg(args[1])
			if !ok {
				return newPositionalTypeError("arange", 2, FLOAT_OBJ, args[1].Type())
			}
			step, ok := toFloatArg(args[2])
			if !ok {
				return newPositionalTypeError("arange", 3, FLOAT_OBJ, args[2].Type())
			}
			if err := checkArgType("arange", 4, STRING_OBJ, args); err != nil {
				return err
			}
			if err := checkArgType("arange", 5, STRING_OBJ, args); err != nil {
				return err
			}
			dtype, derr := ml.ParseDType(args[3].(*Stringo).Value)
			if derr != nil {
				return newError("`arange` error: %s", derr.Error())
			}
			device, derr := ml.ParseDevice(args[4].(*Stringo).Value)
			if derr != nil {
				return newError("`arange` error: %s", derr.Error())
			}
			tt, ferr := ml.Arange(start, end, step, dtype, device)
			if ferr != nil {
				return newError("`arange` error: %s", ferr.Error())
			}
			return &Tensor{T: tt}
		},
		HelpStr: helpStrArgs{
			explanation: "`arange` returns evenly spaced values in [start, end)",
			signature:   "arange(start: float, end: float, step: float=1.0, datatype: str='float32', dev: str='cpu') -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "arange(0.0, 5.0, 1.0) => Tensor{shape: [5]}",
		}.String(),
	},
	{
		Name: "_eye",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("eye", 3, args); err != nil {
				return err
			}
			n, ok := args[0].(*Integer)
			if !ok {
				return newPositionalTypeError("eye", 1, INTEGER_OBJ, args[0].Type())
			}
			if err := checkArgType("eye", 2, STRING_OBJ, args); err != nil {
				return err
			}
			if err := checkArgType("eye", 3, STRING_OBJ, args); err != nil {
				return err
			}
			dtype, derr := ml.ParseDType(args[1].(*Stringo).Value)
			if derr != nil {
				return newError("`eye` error: %s", derr.Error())
			}
			device, derr := ml.ParseDevice(args[2].(*Stringo).Value)
			if derr != nil {
				return newError("`eye` error: %s", derr.Error())
			}
			tt, ferr := ml.Eye(int(n.Value), dtype, device)
			if ferr != nil {
				return newError("`eye` error: %s", ferr.Error())
			}
			return &Tensor{T: tt}
		},
		HelpStr: helpStrArgs{
			explanation: "`eye` returns an n by n identity matrix",
			signature:   "eye(n: int, datatype: str='float32', dev: str='cpu') -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "eye(2) => Tensor{shape: [2 2]}",
		}.String(),
	},
	{
		Name: "_manual_seed",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("manual_seed", 1, args); err != nil {
				return err
			}
			n, ok := args[0].(*Integer)
			if !ok {
				return newPositionalTypeError("manual_seed", 1, INTEGER_OBJ, args[0].Type())
			}
			ml.ManualSeed(n.Value)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`manual_seed` seeds the shared random generator",
			signature:   "manual_seed(seed: int) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "manual_seed(0) => null",
		}.String(),
	},
	{
		Name: "_slice",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("slice", 4, args); err != nil {
				return err
			}
			if err := checkArgType("slice", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			iv := make([]int, 3)
			for i := range iv {
				n, ok := args[i+1].(*Integer)
				if !ok {
					return newPositionalTypeError("slice", i+2, INTEGER_OBJ, args[i+1].Type())
				}
				iv[i] = int(n.Value)
			}
			out, err := ml.Slice(args[0].(*Tensor).T, iv[0], iv[1], iv[2])
			if err != nil {
				return newError("`slice` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`slice` returns a view of `dim` from `start` to `end`",
			signature:   "slice(a: tensor, dim: int, start: int, end: int) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "slice(tensor([[1.0, 2.0], [3.0, 4.0]]), 0, 0, 1) => Tensor{shape: [1 2]}",
		}.String(),
	},
	{
		Name: "_clamp",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("clamp", 3, args); err != nil {
				return err
			}
			if err := checkArgType("clamp", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			lo, ok := toFloatArg(args[1])
			if !ok {
				return newPositionalTypeError("clamp", 2, FLOAT_OBJ, args[1].Type())
			}
			hi, ok := toFloatArg(args[2])
			if !ok {
				return newPositionalTypeError("clamp", 3, FLOAT_OBJ, args[2].Type())
			}
			out, err := ml.Clamp(args[0].(*Tensor).T, lo, hi)
			if err != nil {
				return newError("`clamp` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`clamp` limits each element to [min, max]",
			signature:   "clamp(a: tensor, min: float, max: float) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "clamp(tensor([[-5.0, 0.5, 5.0]]), 0.0, 1.0) => Tensor{shape: [1 3]}",
		}.String(),
	},
	{
		Name: "_where",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("where", 3, args); err != nil {
				return err
			}
			if err := checkArgType("where", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			if err := checkArgType("where", 2, TENSOR_OBJ, args); err != nil {
				return err
			}
			if err := checkArgType("where", 3, TENSOR_OBJ, args); err != nil {
				return err
			}
			out, err := ml.Where(args[0].(*Tensor).T, args[1].(*Tensor).T, args[2].(*Tensor).T)
			if err != nil {
				return newError("`where` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`where` selects a where condition is true and b otherwise",
			signature:   "where(condition: tensor, a: tensor, b: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "where(tensor([[1.0, 0.0]]), a, b) => Tensor{shape: [1 2]}",
		}.String(),
	},
	{
		Name: "_onehot",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("onehot", 2, args); err != nil {
				return err
			}
			if err := checkArgType("onehot", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			n, ok := args[1].(*Integer)
			if !ok {
				return newPositionalTypeError("onehot", 2, INTEGER_OBJ, args[1].Type())
			}
			out, err := ml.OneHot(args[0].(*Tensor).T, int(n.Value))
			if err != nil {
				return newError("`onehot` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`onehot` turns a 1d label tensor into a [n, classes] matrix",
			signature:   "onehot(labels: tensor, classes: int) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "onehot(tensor([0.0, 2.0]), 3) => Tensor{shape: [2 3]}",
		}.String(),
	},
	{
		Name: "_set_requires_grad",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("set_requires_grad", 2, args); err != nil {
				return err
			}
			if err := checkArgType("set_requires_grad", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			if err := checkArgType("set_requires_grad", 2, BOOLEAN_OBJ, args); err != nil {
				return err
			}
			args[0].(*Tensor).T.SetRequiresGrad(args[1].(*Boolean).Value)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`set_requires_grad` turns gradient tracking on or off for a tensor",
			signature:   "set_requires_grad(a: tensor, flag: bool) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "set_requires_grad(t, true) => null",
		}.String(),
	},
	{
		Name: "_requires_grad",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("requires_grad", 1, args); err != nil {
				return err
			}
			if err := checkArgType("requires_grad", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			return nativeToBooleanObject(args[0].(*Tensor).T.RequiresGrad())
		},
		HelpStr: helpStrArgs{
			explanation: "`requires_grad` reports whether a tensor is tracked",
			signature:   "requires_grad(a: tensor) -> bool",
			errors:      "InvalidArgCount,PositionalType",
			example:     "requires_grad(t) => true",
		}.String(),
	},
	{
		Name: "_backward",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("backward", 1, args); err != nil {
				return err
			}
			if err := checkArgType("backward", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			if err := args[0].(*Tensor).T.Backward(); err != nil {
				return newError("`backward` error: %s", err.Error())
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`backward` runs reverse-mode autograd from a scalar tensor",
			signature:   "backward(a: tensor) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "backward(loss) => null",
		}.String(),
	},
	{
		Name: "_grad",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("grad", 1, args); err != nil {
				return err
			}
			if err := checkArgType("grad", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			g := args[0].(*Tensor).T.Grad()
			if g == nil {
				return NULL
			}
			return &Tensor{T: g}
		},
		HelpStr: helpStrArgs{
			explanation: "`grad` returns the accumulated gradient or null",
			signature:   "grad(a: tensor) -> tensor|null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "grad(t) => Tensor{shape: [1 1]}",
		}.String(),
	},
	{
		Name: "_zero_grad",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("zero_grad", 1, args); err != nil {
				return err
			}
			if err := checkArgType("zero_grad", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			args[0].(*Tensor).T.ZeroGrad()
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`zero_grad` clears a tensor's gradient",
			signature:   "zero_grad(a: tensor) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "zero_grad(t) => null",
		}.String(),
	},
	{
		Name: "_set_grad_enabled",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("set_grad_enabled", 1, args); err != nil {
				return err
			}
			if err := checkArgType("set_grad_enabled", 1, BOOLEAN_OBJ, args); err != nil {
				return err
			}
			return nativeToBooleanObject(ml.SetGradEnabled(args[0].(*Boolean).Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`set_grad_enabled` turns graph building on or off and returns the previous value",
			signature:   "set_grad_enabled(flag: bool) -> bool",
			errors:      "InvalidArgCount,PositionalType",
			example:     "set_grad_enabled(false) => true",
		}.String(),
	},
	{
		Name: "_cross_entropy",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("cross_entropy", 2, args); err != nil {
				return err
			}
			if err := checkArgType("cross_entropy", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			if err := checkArgType("cross_entropy", 2, TENSOR_OBJ, args); err != nil {
				return err
			}
			out, err := ml.CrossEntropy(args[0].(*Tensor).T, args[1].(*Tensor).T)
			if err != nil {
				return newError("`cross_entropy` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`cross_entropy` returns the mean cross-entropy loss between logits [batch, classes] and class indices [batch]",
			signature:   "cross_entropy(logits: tensor, target: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "cross_entropy(tensor([[2.0, 1.0, 0.1]]), tensor([0.0])) => Tensor{shape: [1]}",
		}.String(),
	},
	{
		Name: "_mse_loss",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("mse_loss", 2, args); err != nil {
				return err
			}
			if err := checkArgType("mse_loss", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			if err := checkArgType("mse_loss", 2, TENSOR_OBJ, args); err != nil {
				return err
			}
			out, err := ml.MSE(args[0].(*Tensor).T, args[1].(*Tensor).T)
			if err != nil {
				return newError("`mse_loss` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`mse_loss` returns the mean squared error between two tensors of equal shape",
			signature:   "mse_loss(predictions: tensor, targets: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "mse_loss(tensor([[1.0, 2.0]]), tensor([[0.0, 0.0]])) => Tensor{shape: [1]}",
		}.String(),
	},
	{
		Name: "_add_",
		Fun:  inPlaceBuiltin("add_", ml.AddInPlace),
		HelpStr: helpStrArgs{
			explanation: "`add_` adds b into a in place",
			signature:   "add_(a: tensor, b: tensor) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "add_(a, b) => null",
		}.String(),
	},
	{
		Name: "_sub_",
		Fun:  inPlaceBuiltin("sub_", ml.SubInPlace),
		HelpStr: helpStrArgs{
			explanation: "`sub_` subtracts b from a in place",
			signature:   "sub_(a: tensor, b: tensor) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "sub_(a, b) => null",
		}.String(),
	},
	{
		Name: "_mul_",
		Fun:  inPlaceBuiltin("mul_", ml.MulInPlace),
		HelpStr: helpStrArgs{
			explanation: "`mul_` multiplies a by b in place",
			signature:   "mul_(a: tensor, b: tensor) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "mul_(a, b) => null",
		}.String(),
	},
	{
		Name: "_div_",
		Fun:  inPlaceBuiltin("div_", ml.DivInPlace),
		HelpStr: helpStrArgs{
			explanation: "`div_` divides a by b in place",
			signature:   "div_(a: tensor, b: tensor) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "div_(a, b) => null",
		}.String(),
	},
	{
		Name: "_save",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("save", 2, args); err != nil {
				return err
			}
			if err := checkArgType("save", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			if err := checkArgType("save", 2, STRING_OBJ, args); err != nil {
				return err
			}
			if err := ml.Save(args[0].(*Tensor).T, args[1].(*Stringo).Value); err != nil {
				return newError("`save` error: %s", err.Error())
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`save` writes a tensor to a file",
			signature:   "save(a: tensor, path: str) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "save(a, \"a.bin\") => null",
		}.String(),
	},
	{
		Name: "_load",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("load", 2, args); err != nil {
				return err
			}
			if err := checkArgType("load", 1, STRING_OBJ, args); err != nil {
				return err
			}
			if err := checkArgType("load", 2, STRING_OBJ, args); err != nil {
				return err
			}
			device, derr := ml.ParseDevice(args[1].(*Stringo).Value)
			if derr != nil {
				return newError("`load` error: %s", derr.Error())
			}
			t, err := ml.Load(args[0].(*Stringo).Value, device)
			if err != nil {
				return newError("`load` error: %s", err.Error())
			}
			return &Tensor{T: t}
		},
		HelpStr: helpStrArgs{
			explanation: "`load` reads a tensor written by `save` onto a device",
			signature:   "load(path: str, dev: str='cpu') -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "load(\"a.bin\") => Tensor{shape: [2 2]}",
		}.String(),
	},
	{
		Name: "_optim_sgd",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("optim_sgd", 3, args); err != nil {
				return err
			}
			params, errObj := paramList("optim_sgd", args[0])
			if errObj != nil {
				return errObj
			}
			lr, ok := toFloatArg(args[1])
			if !ok {
				return newPositionalTypeError("optim_sgd", 2, FLOAT_OBJ, args[1].Type())
			}
			momentum, ok := toFloatArg(args[2])
			if !ok {
				return newPositionalTypeError("optim_sgd", 3, FLOAT_OBJ, args[2].Type())
			}
			opt, err := ml.SGD(params, lr, momentum)
			if err != nil {
				return newError("`sgd` error: %s", err.Error())
			}
			return &GoObj[*ml.Optimizer]{Value: opt}
		},
		HelpStr: helpStrArgs{
			explanation: "`optim_sgd` builds a borncgo SGD optimizer over a list of parameter tensors",
			signature:   "optim_sgd(params: list[tensor], lr: float, momentum: float) -> GoObj[*ml.Optimizer]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "optim_sgd(params, 0.01, 0.0) => GoObj[*ml.Optimizer]",
		}.String(),
	},
	{
		Name: "_optim_adam",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("optim_adam", 5, args); err != nil {
				return err
			}
			params, errObj := paramList("optim_adam", args[0])
			if errObj != nil {
				return errObj
			}
			lr, ok := toFloatArg(args[1])
			if !ok {
				return newPositionalTypeError("optim_adam", 2, FLOAT_OBJ, args[1].Type())
			}
			beta1, ok := toFloatArg(args[2])
			if !ok {
				return newPositionalTypeError("optim_adam", 3, FLOAT_OBJ, args[2].Type())
			}
			beta2, ok := toFloatArg(args[3])
			if !ok {
				return newPositionalTypeError("optim_adam", 4, FLOAT_OBJ, args[3].Type())
			}
			eps, ok := toFloatArg(args[4])
			if !ok {
				return newPositionalTypeError("optim_adam", 5, FLOAT_OBJ, args[4].Type())
			}
			opt, err := ml.Adam(params, lr, beta1, beta2, eps)
			if err != nil {
				return newError("`adam` error: %s", err.Error())
			}
			return &GoObj[*ml.Optimizer]{Value: opt}
		},
		HelpStr: helpStrArgs{
			explanation: "`optim_adam` builds a borncgo Adam optimizer over a list of parameter tensors",
			signature:   "optim_adam(params: list[tensor], lr: float, beta1: float, beta2: float, eps: float) -> GoObj[*ml.Optimizer]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "optim_adam(params, 0.001, 0.9, 0.999, 1e-8) => GoObj[*ml.Optimizer]",
		}.String(),
	},
	{
		Name: "_optim_step",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("optim_step", 1, args); err != nil {
				return err
			}
			opt, ok := args[0].(*GoObj[*ml.Optimizer])
			if !ok {
				return newPositionalTypeErrorForGoObj("optim_step", 1, "*ml.Optimizer", args[0])
			}
			if err := opt.Value.Step(); err != nil {
				return newError("`step` error: %s", err.Error())
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`optim_step` applies one optimizer update from the last backward",
			signature:   "optim_step(optimizer: GoObj[*ml.Optimizer]) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "optim_step(opt) => null",
		}.String(),
	},
	{
		Name: "_optim_zero_grad",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("optim_zero_grad", 1, args); err != nil {
				return err
			}
			opt, ok := args[0].(*GoObj[*ml.Optimizer])
			if !ok {
				return newPositionalTypeErrorForGoObj("optim_zero_grad", 1, "*ml.Optimizer", args[0])
			}
			opt.Value.ZeroGrad()
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`optim_zero_grad` clears the optimizer's parameter gradients",
			signature:   "optim_zero_grad(optimizer: GoObj[*ml.Optimizer]) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "optim_zero_grad(opt) => null",
		}.String(),
	},
	{
		Name: "_nn_linear",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("linear", 3, args); err != nil {
				return err
			}
			in, ok := args[0].(*Integer)
			if !ok {
				return newPositionalTypeError("linear", 1, INTEGER_OBJ, args[0].Type())
			}
			out, ok := args[1].(*Integer)
			if !ok {
				return newPositionalTypeError("linear", 2, INTEGER_OBJ, args[1].Type())
			}
			if err := checkArgType("linear", 3, STRING_OBJ, args); err != nil {
				return err
			}
			dev, derr := ml.ParseDevice(args[2].(*Stringo).Value)
			if derr != nil {
				return newError("`linear` error: %s", derr.Error())
			}
			m, err := ml.NNLinear(int(in.Value), int(out.Value), dev)
			if err != nil {
				return newError("`linear` error: %s", err.Error())
			}
			return &GoObj[ml.Module]{Value: m}
		},
		HelpStr: helpStrArgs{
			explanation: "`linear` builds a borncgo Linear layer on a device",
			signature:   "linear(in_features: int, out_features: int, dev: str='cpu') -> GoObj[*ml.Linear]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "linear(784, 128, 'gpu') => GoObj[*ml.Linear]",
		}.String(),
	},
	{
		Name: "_nn_relu",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("relu", 1, args); err != nil {
				return err
			}
			if err := checkArgType("relu", 1, STRING_OBJ, args); err != nil {
				return err
			}
			return &GoObj[ml.Module]{Value: ml.NNReLU()}
		},
		HelpStr: helpStrArgs{
			explanation: "`relu` builds a borncgo ReLU module",
			signature:   "relu(dev: str='cpu') -> GoObj[*ml.ReLU]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "relu('cpu') => GoObj[*ml.ReLU]",
		}.String(),
	},
	{
		Name: "_nn_sigmoid",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("sigmoid", 1, args); err != nil {
				return err
			}
			if err := checkArgType("sigmoid", 1, STRING_OBJ, args); err != nil {
				return err
			}
			return &GoObj[ml.Module]{Value: ml.NNSigmoid()}
		},
		HelpStr: helpStrArgs{
			explanation: "`sigmoid` builds a borncgo Sigmoid module",
			signature:   "sigmoid(dev: str='cpu') -> GoObj[*ml.Sigmoid]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "sigmoid('cpu') => GoObj[*ml.Sigmoid]",
		}.String(),
	},
	{
		Name: "_nn_forward",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("forward", 2, args); err != nil {
				return err
			}
			m, ok := args[0].(*GoObj[ml.Module])
			if !ok {
				return newPositionalTypeErrorForGoObj("forward", 1, "*ml.Module", args[0])
			}
			if err := checkArgType("forward", 2, TENSOR_OBJ, args); err != nil {
				return err
			}
			return &Tensor{T: ml.NNForward(m.Value, args[1].(*Tensor).T)}
		},
		HelpStr: helpStrArgs{
			explanation: "`forward` runs a borncgo module on a tensor",
			signature:   "forward(module: GoObj[*ml.Module], x: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType",
			example:     "forward(linear, x) => Tensor{shape: [1 128]}",
		}.String(),
	},
	{
		Name: "_nn_parameters",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("parameters", 1, args); err != nil {
				return err
			}
			m, ok := args[0].(*GoObj[ml.Module])
			if !ok {
				return newPositionalTypeErrorForGoObj("parameters", 1, "*ml.Module", args[0])
			}
			ps := ml.NNParameters(m.Value)
			elems := make([]Object, len(ps))
			for i, p := range ps {
				elems[i] = &GoObj[*ml.Param]{Value: p}
			}
			return &List{Elements: elems}
		},
		HelpStr: helpStrArgs{
			explanation: "`parameters` returns a module's trainable parameters",
			signature:   "parameters(module: GoObj[*ml.Module]) -> list[GoObj[*ml.Param]]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "parameters(linear) => list[GoObj[*ml.Param]]",
		}.String(),
	},
	{
		Name: "_gpu_is_available",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("is_available", 0, args); err != nil {
				return err
			}
			return nativeToBooleanObject(ml.GPUAvailable())
		},
		HelpStr: helpStrArgs{
			explanation: "`is_available` reports whether a GPU backend is available",
			signature:   "is_available() -> bool",
			errors:      "InvalidArgCount",
			example:     "is_available() => true",
		}.String(),
	},
	{
		Name: "_nn_to",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("nn_to", 2, args); err != nil {
				return err
			}
			m, ok := args[0].(*GoObj[ml.Module])
			if !ok {
				return newPositionalTypeErrorForGoObj("nn_to", 1, "*ml.Module", args[0])
			}
			if err := checkArgType("nn_to", 2, STRING_OBJ, args); err != nil {
				return err
			}
			dev, derr := ml.ParseDevice(args[1].(*Stringo).Value)
			if derr != nil {
				return newError("`to` error: %s", derr.Error())
			}
			if err := ml.NNTo(m.Value, dev); err != nil {
				return newError("`to` error: %s", err.Error())
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`nn_to` moves a module's parameters to a device",
			signature:   "nn_to(module: GoObj[*ml.Module], dev: str) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "nn_to(linear, 'gpu') => null",
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

func toFloatArg(o Object) (float32, bool) {
	switch v := o.(type) {
	case *Float:
		return float32(v.Value), true
	case *Integer:
		return float32(v.Value), true
	}
	return 0, false
}

// paramList extracts a list of borncgo parameter handles (from ml.nn.parameters).
func paramList(name string, o Object) ([]*ml.Param, Object) {
	l, ok := o.(*List)
	if !ok {
		return nil, newPositionalTypeError(name, 1, LIST_OBJ, o.Type())
	}
	out := make([]*ml.Param, len(l.Elements))
	for i, e := range l.Elements {
		g, ok := e.(*GoObj[*ml.Param])
		if !ok {
			return nil, newPositionalTypeErrorForGoObj(name, 1, "*ml.Param", e)
		}
		out[i] = g.Value
	}
	return out, nil
}

// listBuiltin builds a module function taking (tensor, list[int]).
func listBuiltin(name string, f func(*ml.Tensor, []int) (*ml.Tensor, error)) func(...Object) Object {
	return func(args ...Object) Object {
		if err := checkArgCount(name, 2, args); err != nil {
			return err
		}
		if err := checkArgType(name, 1, TENSOR_OBJ, args); err != nil {
			return err
		}
		l, ok := args[1].(*List)
		if !ok {
			return newPositionalTypeError(name, 2, LIST_OBJ, args[1].Type())
		}
		dims, err := toIntList(name, l)
		if err != nil {
			return newError("%s", err.Error())
		}
		out, ferr := f(args[0].(*Tensor).T, dims)
		if ferr != nil {
			return newError("`%s` error: %s", name, ferr.Error())
		}
		return &Tensor{T: out}
	}
}

// creationBuiltin parses (shape, datatype, dev, requires_grad) and calls f.
func creationBuiltin(name string, f func([]int, ml.DType, ml.Device) (*ml.Tensor, error)) func(...Object) Object {
	return func(args ...Object) Object {
		if err := checkArgCount(name, 4, args); err != nil {
			return err
		}
		if err := checkArgType(name, 1, LIST_OBJ, args); err != nil {
			return err
		}
		if err := checkArgType(name, 2, STRING_OBJ, args); err != nil {
			return err
		}
		if err := checkArgType(name, 3, STRING_OBJ, args); err != nil {
			return err
		}
		if err := checkArgType(name, 4, BOOLEAN_OBJ, args); err != nil {
			return err
		}
		shape, err := toIntList(name, args[0].(*List))
		if err != nil {
			return newError("%s", err.Error())
		}
		dtype, derr := ml.ParseDType(args[1].(*Stringo).Value)
		if derr != nil {
			return newError("`%s` error: %s", name, derr.Error())
		}
		device, derr := ml.ParseDevice(args[2].(*Stringo).Value)
		if derr != nil {
			return newError("`%s` error: %s", name, derr.Error())
		}
		tt, ferr := f(shape, dtype, device)
		if ferr != nil {
			return newError("`%s` error: %s", name, ferr.Error())
		}
		tt.SetRequiresGrad(args[3].(*Boolean).Value)
		return &Tensor{T: tt}
	}
}

// inPlaceBuiltin builds (tensor, tensor-or-scalar) -> null, mutating a.
func inPlaceBuiltin(name string, f func(a, b *ml.Tensor) error) func(...Object) Object {
	return func(args ...Object) Object {
		if err := checkArgCount(name, 2, args); err != nil {
			return err
		}
		at, ok := args[0].(*Tensor)
		if !ok {
			return newPositionalTypeError(name, 1, TENSOR_OBJ, args[0].Type())
		}
		b, ok := asTensorArg(args[1], at.T.Device())
		if !ok {
			return newPositionalTypeError(name, 2, TENSOR_OBJ, args[1].Type())
		}
		if err := f(at.T, b); err != nil {
			return newError("`%s` error: %s", name, err.Error())
		}
		return NULL
	}
}
