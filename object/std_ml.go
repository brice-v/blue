package object

import (
	"blue/ml"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
)

type tensorData struct {
	// leaves are collected as float64 (which holds int64 exactly up to 2^53)
	// and converted to the requested dtype when the tensor is built.
	data []float64
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
		case FLOAT_OBJ, INTEGER_OBJ, BOOLEAN_OBJ:
			if len(shape) != 1 {
				return fmt.Errorf("got a scalar where a list of length %d was expected", shape[1])
			}
			switch v := e.(type) {
			case *Float:
				tdata.data = append(tdata.data, v.Value)
			case *Integer:
				tdata.data = append(tdata.data, float64(v.Value))
			case *Boolean:
				if v.Value {
					tdata.data = append(tdata.data, 1)
				} else {
					tdata.data = append(tdata.data, 0)
				}
			}
		case LIST_OBJ:
			if len(shape) == 1 {
				return fmt.Errorf("got a list where a scalar was expected")
			}
			if err := appendToData(tdata, e.(*List), shape[1:]); err != nil {
				return err
			}
		default:
			return fmt.Errorf("encountered %s, expected list or scalar", e.Type())
		}
	}
	return nil
}

// newTypedTensor builds a tensor of the requested dtype from collected leaves.
func newTypedTensor(data []float64, shape []int, dtype ml.DType, device ml.Device) (*ml.Tensor, error) {
	switch dtype {
	case ml.Float32:
		d := make([]float32, len(data))
		for i, v := range data {
			d[i] = float32(v)
		}
		return ml.NewTensor(d, shape, dtype, device)
	case ml.Float64:
		return ml.NewFloat64Tensor(data, shape, device)
	case ml.Int32:
		d := make([]int32, len(data))
		for i, v := range data {
			d[i] = int32(v)
		}
		return ml.NewInt32Tensor(d, shape, device)
	case ml.Int64:
		d := make([]int64, len(data))
		for i, v := range data {
			d[i] = int64(v)
		}
		return ml.NewInt64Tensor(d, shape, device)
	case ml.Uint8:
		d := make([]uint8, len(data))
		for i, v := range data {
			d[i] = uint8(v)
		}
		return ml.NewUint8Tensor(d, shape, device)
	case ml.Bool:
		d := make([]bool, len(data))
		for i, v := range data {
			d[i] = v != 0
		}
		return ml.NewBoolTensor(d, shape, device)
	default:
		return nil, fmt.Errorf("unsupported dtype %s", dtype)
	}
}

var MlBuiltins = []*Builtin{
	{
		Name: "_tensor",
		Fun: func(args ...Object) Object {
			err := checkArgCount("tensor", 4, args)
			if err != nil {
				return err
			}
			if args[0].Type() != LIST_OBJ && args[0].Type() != FLOAT_OBJ &&
				args[0].Type() != INTEGER_OBJ && args[0].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("tensor", 1, "list or scalar", args[0].Type())
			}
			tdata := &tensorData{data: []float64{}}
			var shape []int
			switch v := args[0].(type) {
			case *List:
				shape = inferShape(v)
				if err := appendToData(tdata, v, shape); err != nil {
					return newError("`tensor` error: %s", err.Error())
				}
			case *Float:
				tdata.data = append(tdata.data, v.Value)
				shape = []int{1}
			case *Integer:
				tdata.data = append(tdata.data, float64(v.Value))
				shape = []int{1}
			case *Boolean:
				if v.Value {
					tdata.data = append(tdata.data, 1)
				} else {
					tdata.data = append(tdata.data, 0)
				}
				shape = []int{1}
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
			tt, terr := newTypedTensor(tdata.data, shape, dtype, device)
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
			if err := checkArgCount("slice", 5, args); err != nil {
				return err
			}
			if err := checkArgType("slice", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			iv := make([]int, 4)
			for i := range iv {
				n, ok := args[i+1].(*Integer)
				if !ok {
					return newPositionalTypeError("slice", i+2, INTEGER_OBJ, args[i+1].Type())
				}
				iv[i] = int(n.Value)
			}
			out, err := ml.Slice(args[0].(*Tensor).T, iv[0], iv[1], iv[2], iv[3])
			if err != nil {
				return newError("`slice` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`slice` returns a view of `dim` from `start` to `end` with `step`, matching a[start:end:step]",
			signature:   "slice(a: tensor, dim: int, start: int, end: int, step: int=1) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "slice(tensor([[1.0, 2.0], [3.0, 4.0]]), 0, 0, 1, 1) => Tensor{shape: [1 2]}",
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
		Name: "_cat",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("cat", 2, args); err != nil {
				return err
			}
			list, ok := args[0].(*List)
			if !ok {
				return newPositionalTypeError("cat", 1, LIST_OBJ, args[0].Type())
			}
			n, ok := args[1].(*Integer)
			if !ok {
				return newPositionalTypeError("cat", 2, INTEGER_OBJ, args[1].Type())
			}
			tensors := make([]*ml.Tensor, len(list.Elements))
			for i, e := range list.Elements {
				t, ok := e.(*Tensor)
				if !ok {
					return newError("`cat` error: element %d is not a tensor", i)
				}
				tensors[i] = t.T
			}
			out, err := ml.Cat(tensors, int(n.Value))
			if err != nil {
				return newError("`cat` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`cat` joins a list of tensors along a dimension (same as torch.cat)",
			signature:   "cat(tensors: list[tensor], dim: int) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "cat([a, b], 0) => Tensor{shape: [4 3]}",
		}.String(),
	},
	{
		Name: "_gather",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("gather", 3, args); err != nil {
				return err
			}
			if err := checkArgType("gather", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			n, ok := args[1].(*Integer)
			if !ok {
				return newPositionalTypeError("gather", 2, INTEGER_OBJ, args[1].Type())
			}
			if err := checkArgType("gather", 3, TENSOR_OBJ, args); err != nil {
				return err
			}
			out, err := ml.Gather(args[0].(*Tensor).T, int(n.Value), args[2].(*Tensor).T)
			if err != nil {
				return newError("`gather` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`gather` selects entries along a dimension using an index tensor (torch.gather)",
			signature:   "gather(a: tensor, dim: int, index: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "gather(a, 1, idx) => Tensor{shape: [2 3]}",
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
		Name: "_cross_entropy_grad",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("cross_entropy_grad", 2, args); err != nil {
				return err
			}
			if err := checkArgType("cross_entropy_grad", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			if err := checkArgType("cross_entropy_grad", 2, TENSOR_OBJ, args); err != nil {
				return err
			}
			out, err := ml.CrossEntropyGrad(args[0].(*Tensor).T, args[1].(*Tensor).T)
			if err != nil {
				return newError("`cross_entropy_grad` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`cross_entropy_grad` returns the gradient of the mean cross-entropy loss with respect to the logits, for use as a compiled-backward seed",
			signature:   "cross_entropy_grad(logits: tensor, target: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "cross_entropy_grad(tensor([[2.0, 1.0, 0.1]]), tensor([0.0])) => Tensor{shape: [1 3]}",
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
			if err := checkArgCount("optim_sgd", 4, args); err != nil {
				return err
			}
			params, errObj := optimParams("optim_sgd", args[0])
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
			weightDecay, ok := toFloatArg(args[3])
			if !ok {
				return newPositionalTypeError("optim_sgd", 4, FLOAT_OBJ, args[3].Type())
			}
			opt, err := ml.SGD(params, lr, momentum, weightDecay)
			if err != nil {
				return newError("`sgd` error: %s", err.Error())
			}
			return &GoObj[*ml.Optimizer]{Value: opt}
		},
		HelpStr: helpStrArgs{
			explanation: "`optim_sgd` builds an SGD optimizer over tensors or module parameters",
			signature:   "optim_sgd(params: list, lr: float, momentum: float, weight_decay: float) -> GoObj[*ml.Optimizer]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "optim_sgd(params, 0.01, 0.0, 0.0) => GoObj[*ml.Optimizer]",
		}.String(),
	},
	{
		Name: "_optim_adam",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("optim_adam", 6, args); err != nil {
				return err
			}
			params, errObj := optimParams("optim_adam", args[0])
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
			weightDecay, ok := toFloatArg(args[5])
			if !ok {
				return newPositionalTypeError("optim_adam", 6, FLOAT_OBJ, args[5].Type())
			}
			opt, err := ml.Adam(params, lr, beta1, beta2, eps, weightDecay)
			if err != nil {
				return newError("`adam` error: %s", err.Error())
			}
			return &GoObj[*ml.Optimizer]{Value: opt}
		},
		HelpStr: helpStrArgs{
			explanation: "`optim_adam` builds an Adam (AdamW when weight_decay is nonzero) optimizer over tensors or module parameters",
			signature:   "optim_adam(params: list, lr: float, beta1: float, beta2: float, eps: float, weight_decay: float) -> GoObj[*ml.Optimizer]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "optim_adam(params, 0.001, 0.9, 0.999, 1e-8, 0.0) => GoObj[*ml.Optimizer]",
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
			signature:   "linear(in_features: int, out_features: int, dev: str='cpu') -> GoObj[*ml.Module]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "linear(784, 128, 'gpu') => GoObj[*ml.Module]",
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
			signature:   "relu(dev: str='cpu') -> GoObj[*ml.Module]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "relu('cpu') => GoObj[*ml.Module]",
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
			signature:   "sigmoid(dev: str='cpu') -> GoObj[*ml.Module]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "sigmoid('cpu') => GoObj[*ml.Module]",
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
			var out []Object
			if errObj := collectParams("parameters", args[0], map[Object]bool{}, &out); errObj != nil {
				return errObj
			}
			if out == nil {
				out = []Object{}
			}
			return &List{Elements: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`parameters` collects trainable leaves from a model: a map or list is walked recursively, tensors with requires_grad are kept, and nn module parameters are returned as handles",
			signature:   "parameters(model: map|list|tensor|GoObj[*ml.Module]) -> list[tensor|GoObj[*ml.Param]]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "parameters({'w': ml.randn([2, 2], requires_grad=true)}) => list[tensor]",
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
	{
		Name: "_cast",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("cast", 2, args); err != nil {
				return err
			}
			if err := checkArgType("cast", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			if err := checkArgType("cast", 2, STRING_OBJ, args); err != nil {
				return err
			}
			dt, derr := ml.ParseDType(args[1].(*Stringo).Value)
			if derr != nil {
				return newError("`cast` error: %s", derr.Error())
			}
			out, err := args[0].(*Tensor).T.Cast(dt)
			if err != nil {
				return newError("`cast` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`cast` converts a tensor to another dtype",
			signature:   "cast(a: tensor, datatype: str) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "cast(tensor([1.0]), 'int32') => Tensor{shape: [1]}",
		}.String(),
	},
	{
		Name: "_nn_linear_act",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("linear_act", 3, args); err != nil {
				return err
			}
			m, ok := args[0].(*GoObj[ml.Module])
			if !ok {
				return newPositionalTypeErrorForGoObj("linear_act", 1, "*ml.Module", args[0])
			}
			if err := checkArgType("linear_act", 2, TENSOR_OBJ, args); err != nil {
				return err
			}
			if err := checkArgType("linear_act", 3, BOOLEAN_OBJ, args); err != nil {
				return err
			}
			out, err := ml.NNLinearForwardAct(m.Value, args[1].(*Tensor).T, args[2].(*Boolean).Value)
			if err != nil {
				return newError("`linear_act` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`linear_act` runs a Linear layer fused with the following activation (one dispatch)",
			signature:   "linear_act(module: GoObj[*ml.Module], x: tensor, relu: bool) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "linear_act(linear, x, true) => Tensor{shape: [1 128]}",
		}.String(),
	},
	{
		Name: "_trace_begin",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("compile", 1, args); err != nil {
				return err
			}
			if err := checkArgType("compile", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			tr, terr := ml.NewTracer(args[0].(*Tensor).T)
			if terr != nil {
				return newError("`compile` error: %s", terr.Error())
			}
			return &GoObj[*ml.Tracer]{Value: tr}
		},
		HelpStr: helpStrArgs{
			explanation: "`trace_begin` starts a traced capture of a model's forward pass",
			signature:   "trace_begin(example: tensor) -> GoObj[*ml.Tracer]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "trace_begin(x) => GoObj[*ml.Tracer]",
		}.String(),
	},
	{
		Name: "_trace_input",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("compile", 1, args); err != nil {
				return err
			}
			tr, ok := args[0].(*GoObj[*ml.Tracer])
			if !ok {
				return newPositionalTypeErrorForGoObj("compile", 1, "*ml.Tracer", args[0])
			}
			return &Tensor{T: tr.Value.InputTensor()}
		},
		HelpStr: helpStrArgs{
			explanation: "`trace_input` returns the symbolic input for a traced capture",
			signature:   "trace_input(tracer: GoObj[*ml.Tracer]) -> tensor",
			errors:      "InvalidArgCount,PositionalType",
			example:     "trace_input(tracer) => Tensor",
		}.String(),
	},
	{
		Name: "_trace_end",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("compile", 2, args); err != nil {
				return err
			}
			tr, ok := args[0].(*GoObj[*ml.Tracer])
			if !ok {
				return newPositionalTypeErrorForGoObj("compile", 1, "*ml.Tracer", args[0])
			}
			if err := checkArgType("compile", 2, TENSOR_OBJ, args); err != nil {
				return err
			}
			g, gerr := tr.Value.Finish(args[1].(*Tensor).T)
			if gerr != nil {
				return NULL
			}
			return &GoObj[*ml.Graph]{Value: g}
		},
		HelpStr: helpStrArgs{
			explanation: "`trace_end` closes a traced capture at the model output; returns null when the forward could not be traced",
			signature:   "trace_end(tracer: GoObj[*ml.Tracer], out: tensor) -> GoObj[*ml.Graph]|null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "trace_end(tracer, out) => GoObj[*ml.Graph]",
		}.String(),
	},
	{
		Name: "_graph_optimize",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("compile", 1, args); err != nil {
				return err
			}
			g, ok := args[0].(*GoObj[*ml.Graph])
			if !ok {
				return newPositionalTypeErrorForGoObj("compile", 1, "*ml.Graph", args[0])
			}
			g.Value.Optimize()
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`graph_optimize` runs the fusion and pruning passes over a captured graph",
			signature:   "graph_optimize(graph: GoObj[*ml.Graph]) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "graph_optimize(graph) => null",
		}.String(),
	},
	{
		Name: "_compiled_new",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("compile", 0, args); err != nil {
				return err
			}
			return &GoObj[*ml.Compiled]{Value: ml.NewCompiled()}
		},
		HelpStr: helpStrArgs{
			explanation: "`compiled_new` creates a plan cache for a compiled model",
			signature:   "compiled_new() -> GoObj[*ml.Compiled]",
			errors:      "InvalidArgCount",
			example:     "compiled_new() => GoObj[*ml.Compiled]",
		}.String(),
	},
	{
		Name: "_compiled_has_plan",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("compile", 2, args); err != nil {
				return err
			}
			c, ok := args[0].(*GoObj[*ml.Compiled])
			if !ok {
				return newPositionalTypeErrorForGoObj("compile", 1, "*ml.Compiled", args[0])
			}
			if err := checkArgType("compile", 2, TENSOR_OBJ, args); err != nil {
				return err
			}
			return nativeToBooleanObject(c.Value.Plan(args[1].(*Tensor).T) != nil)
		},
		HelpStr: helpStrArgs{
			explanation: "`compiled_has_plan` reports whether a plan exists for a tensor's shape (a guard hit)",
			signature:   "compiled_has_plan(compiled: GoObj[*ml.Compiled], x: tensor) -> bool",
			errors:      "InvalidArgCount,PositionalType",
			example:     "compiled_has_plan(compiled, x) => true",
		}.String(),
	},
	{
		Name: "_compiled_install",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("compile", 3, args); err != nil {
				return err
			}
			c, ok := args[0].(*GoObj[*ml.Compiled])
			if !ok {
				return newPositionalTypeErrorForGoObj("compile", 1, "*ml.Compiled", args[0])
			}
			if err := checkArgType("compile", 2, TENSOR_OBJ, args); err != nil {
				return err
			}
			g, ok := args[2].(*GoObj[*ml.Graph])
			if !ok {
				return newPositionalTypeErrorForGoObj("compile", 3, "*ml.Graph", args[2])
			}
			c.Value.Install(args[1].(*Tensor).T, g.Value)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`compiled_install` installs a captured graph for an input shape",
			signature:   "compiled_install(compiled: GoObj[*ml.Compiled], x: tensor, graph: GoObj[*ml.Graph]) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "compiled_install(compiled, x, graph) => null",
		}.String(),
	},
	{
		Name: "_compiled_run",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("compile", 2, args); err != nil {
				return err
			}
			c, ok := args[0].(*GoObj[*ml.Compiled])
			if !ok {
				return newPositionalTypeErrorForGoObj("compile", 1, "*ml.Compiled", args[0])
			}
			if err := checkArgType("compile", 2, TENSOR_OBJ, args); err != nil {
				return err
			}
			out, rerr := c.Value.Run(args[1].(*Tensor).T)
			if rerr != nil {
				return newError("`compile` error: %s", rerr.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`compiled_run` replays the cached plan for a tensor",
			signature:   "compiled_run(compiled: GoObj[*ml.Compiled], x: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "compiled_run(compiled, x) => Tensor",
		}.String(),
	},
	{
		Name: "_compiled_stats",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("compile", 1, args); err != nil {
				return err
			}
			c, ok := args[0].(*GoObj[*ml.Compiled])
			if !ok {
				return newPositionalTypeErrorForGoObj("compile", 1, "*ml.Compiled", args[0])
			}
			return &Stringo{Value: c.Value.Stats()}
		},
		HelpStr: helpStrArgs{
			explanation: "`compiled_stats` reports plan count and recompilations",
			signature:   "compiled_stats(compiled: GoObj[*ml.Compiled]) -> str",
			errors:      "InvalidArgCount,PositionalType",
			example:     "compiled_stats(compiled) => 'plans=1 compiles=1 backwards=1'",
		}.String(),
	},
	{
		Name: "_compiled_backward",
		Fun: func(args ...Object) Object {
			if err := checkArgsCount("compile", []int{2, 3}, args); err != nil {
				return err
			}
			c, ok := args[0].(*GoObj[*ml.Compiled])
			if !ok {
				return newPositionalTypeErrorForGoObj("compile", 1, "*ml.Compiled", args[0])
			}
			if err := checkArgType("compile", 2, TENSOR_OBJ, args); err != nil {
				return err
			}
			x := args[1].(*Tensor).T
			var err error
			if len(args) == 3 {
				if err := checkArgType("compile", 3, TENSOR_OBJ, args); err != nil {
					return err
				}
				err = c.Value.BackwardWithSeed(x, args[2].(*Tensor).T)
			} else {
				err = c.Value.Backward(x)
			}
			if err != nil {
				return newError("`compile` error: %s", err.Error())
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`compiled_backward` runs the compiled backward for a tensor, filling parameter gradients; an optional seed is the loss gradient w.r.t. the output",
			signature:   "compiled_backward(compiled: GoObj[*ml.Compiled], x: tensor, seed: tensor=null) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "compiled_backward(compiled, x) => null",
		}.String(),
	},
	{
		Name: "_embedding",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("embedding", 2, args); err != nil {
				return err
			}
			if err := checkArgType("embedding", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			if err := checkArgType("embedding", 2, TENSOR_OBJ, args); err != nil {
				return err
			}
			out, err := ml.Embedding(args[0].(*Tensor).T, args[1].(*Tensor).T)
			if err != nil {
				return newError("`embedding` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`embedding` looks up rows of a [vocab, dim] weight tensor by an index tensor, matching torch.nn.functional.embedding",
			signature:   "embedding(weight: tensor, indices: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "embedding(ml.randn([10, 4]), ml.tensor([1, 3])) => Tensor{shape: [2 4]}",
		}.String(),
	},
	{
		Name: "_rand",
		Fun:  creationBuiltin("rand", ml.Rand),
		HelpStr: helpStrArgs{
			explanation: "`rand` returns uniform samples in [0, 1) from the shared engine RNG",
			signature:   "rand(shape: list[int], datatype: str='float32', dev: str='cpu', requires_grad: bool=false) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "rand([2, 2]) => Tensor{shape: [2 2]}",
		}.String(),
	},
	{
		Name: "_randperm",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("randperm", 2, args); err != nil {
				return err
			}
			n, ok := args[0].(*Integer)
			if !ok {
				return newPositionalTypeError("randperm", 1, INTEGER_OBJ, args[0].Type())
			}
			s, ok := args[1].(*Stringo)
			if !ok {
				return newPositionalTypeError("randperm", 2, STRING_OBJ, args[1].Type())
			}
			dev, derr := ml.ParseDevice(s.Value)
			if derr != nil {
				return newError("`randperm` error: %s", derr.Error())
			}
			out, err := ml.RandPerm(int(n.Value), dev)
			if err != nil {
				return newError("`randperm` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`randperm` returns a random permutation of 0..n-1, reproducible after manual_seed",
			signature:   "randperm(n: int, dev: str='cpu') -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "randperm(4) => Tensor{shape: [4]}",
		}.String(),
	},
	{
		Name: "_shuffle",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("shuffle", 2, args); err != nil {
				return err
			}
			if err := checkArgType("shuffle", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			d, ok := args[1].(*Integer)
			if !ok {
				return newPositionalTypeError("shuffle", 2, INTEGER_OBJ, args[1].Type())
			}
			out, err := ml.Shuffle(args[0].(*Tensor).T, int(d.Value))
			if err != nil {
				return newError("`shuffle` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`shuffle` returns a copy of a with dim permuted by a random permutation",
			signature:   "shuffle(a: tensor, dim: int=0) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "shuffle(ml.randn([4, 2]), 0) => Tensor{shape: [4 2]}",
		}.String(),
	},
	{
		Name: "_select",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("select", 3, args); err != nil {
				return err
			}
			if err := checkArgType("select", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			d, ok := args[1].(*Integer)
			if !ok {
				return newPositionalTypeError("select", 2, INTEGER_OBJ, args[1].Type())
			}
			idx, ok := args[2].(*Integer)
			if !ok {
				return newPositionalTypeError("select", 3, INTEGER_OBJ, args[2].Type())
			}
			out, err := ml.Select(args[0].(*Tensor).T, int(d.Value), int(idx.Value))
			if err != nil {
				return newError("`select` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`select` returns the entries at index along dim with that dim removed, covering x[i] and x[:, j]",
			signature:   "select(a: tensor, dim: int, index: int) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "select(ml.randn([3, 4]), 0, 1) => Tensor{shape: [4]}",
		}.String(),
	},
	{
		Name: "_index_select",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("index_select", 3, args); err != nil {
				return err
			}
			if err := checkArgType("index_select", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			d, ok := args[1].(*Integer)
			if !ok {
				return newPositionalTypeError("index_select", 2, INTEGER_OBJ, args[1].Type())
			}
			if err := checkArgType("index_select", 3, TENSOR_OBJ, args); err != nil {
				return err
			}
			out, err := ml.IndexSelect(args[0].(*Tensor).T, int(d.Value), args[2].(*Tensor).T)
			if err != nil {
				return newError("`index_select` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`index_select` gathers rows or slices along dim using a 1d index tensor, matching torch.index_select",
			signature:   "index_select(a: tensor, dim: int, index: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "index_select(ml.randn([3, 4]), 0, ml.tensor([2, 0])) => Tensor{shape: [2 4]}",
		}.String(),
	},
	{
		Name: "_masked_fill",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("masked_fill", 3, args); err != nil {
				return err
			}
			if err := checkArgType("masked_fill", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			if err := checkArgType("masked_fill", 2, TENSOR_OBJ, args); err != nil {
				return err
			}
			v, ok := toFloatArg(args[2])
			if !ok {
				return newPositionalTypeError("masked_fill", 3, FLOAT_OBJ, args[2].Type())
			}
			out, err := ml.MaskedFill(args[0].(*Tensor).T, args[1].(*Tensor).T, v)
			if err != nil {
				return newError("`masked_fill` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`masked_fill` replaces elements where mask is true with value, matching torch.Tensor.masked_fill",
			signature:   "masked_fill(a: tensor, mask: tensor, value: float) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "masked_fill(ml.tensor([1.0, 2.0]), ml.tensor([0.0, 1.0]), 0.0) => Tensor{shape: [2]}",
		}.String(),
	},
	{
		Name: "_optim_set_lr",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("optim_set_lr", 2, args); err != nil {
				return err
			}
			opt, ok := args[0].(*GoObj[*ml.Optimizer])
			if !ok {
				return newPositionalTypeErrorForGoObj("optim_set_lr", 1, "*ml.Optimizer", args[0])
			}
			lr, ok := toFloatArg(args[1])
			if !ok {
				return newPositionalTypeError("optim_set_lr", 2, FLOAT_OBJ, args[1].Type())
			}
			opt.Value.SetLR(lr)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`optim_set_lr` sets the optimizer learning rate, for a schedule driven from blue",
			signature:   "optim_set_lr(optimizer: GoObj[*ml.Optimizer], lr: float) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "optim_set_lr(opt, 0.001) => null",
		}.String(),
	},
	{
		Name: "_optim_clip_grad_norm",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("clip_grad_norm", 2, args); err != nil {
				return err
			}
			bindings, errObj := optimParams("clip_grad_norm", args[0])
			if errObj != nil {
				return errObj
			}
			maxNorm, ok := toFloatArg(args[1])
			if !ok {
				return newPositionalTypeError("clip_grad_norm", 2, FLOAT_OBJ, args[1].Type())
			}
			norm, err := ml.ClipGradNorm(bindings, maxNorm)
			if err != nil {
				return newError("`clip_grad_norm` error: %s", err.Error())
			}
			return &Float{Value: float64(norm)}
		},
		HelpStr: helpStrArgs{
			explanation: "`clip_grad_norm` scales the last backward's gradients in place so their global L2 norm is at most max_norm, and returns the norm before clipping",
			signature:   "clip_grad_norm(params: list, max_norm: float) -> float",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "clip_grad_norm(params, 1.0) => 1.0",
		}.String(),
	},
	{
		Name: "_lr_schedule",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("lr_schedule", 5, args); err != nil {
				return err
			}
			kind, ok := args[0].(*Stringo)
			if !ok {
				return newPositionalTypeError("lr_schedule", 1, STRING_OBJ, args[0].Type())
			}
			base, ok := toFloatArg(args[1])
			if !ok {
				return newPositionalTypeError("lr_schedule", 2, FLOAT_OBJ, args[1].Type())
			}
			step, ok := args[2].(*Integer)
			if !ok {
				return newPositionalTypeError("lr_schedule", 3, INTEGER_OBJ, args[2].Type())
			}
			a, ok := toFloatArg(args[3])
			if !ok {
				return newPositionalTypeError("lr_schedule", 4, FLOAT_OBJ, args[3].Type())
			}
			b, ok := toFloatArg(args[4])
			if !ok {
				return newPositionalTypeError("lr_schedule", 5, FLOAT_OBJ, args[4].Type())
			}
			lr, err := ml.LRSchedule(kind.Value, base, int(step.Value), a, b)
			if err != nil {
				return newError("`lr_schedule` error: %s", err.Error())
			}
			return &Float{Value: float64(lr)}
		},
		HelpStr: helpStrArgs{
			explanation: "`lr_schedule` computes a learning rate for a step; kinds are constant, step (a=step_size, b=gamma), cosine (a=total_steps, b=min_lr), and warmup (a=warmup_steps)",
			signature:   "lr_schedule(kind: str, base_lr: float, step: int, a: float, b: float) -> float",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "lr_schedule('cosine', 0.1, 5, 100.0, 0.0) => 0.0997",
		}.String(),
	},
	{
		Name: "_clone",
		Fun: tensorUnaryBuiltin("clone", func(a *ml.Tensor) (*ml.Tensor, error) {
			return a.Clone(), nil
		}),
		HelpStr: helpStrArgs{
			explanation: "`clone` returns a deep copy that stays in the autograd graph, like torch.Tensor.clone; use detach for a copy cut out of the graph",
			signature:   "clone(a: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType",
			example:     "clone(ml.tensor([1.0, 2.0])) => Tensor{shape: [2]}",
		}.String(),
	},
	{
		Name: "_state_dict",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("state_dict", 1, args); err != nil {
				return err
			}
			state, errObj := stateDictOf(args[0])
			if errObj != nil {
				return errObj
			}
			return CreateMapObjectForGoMap(*state)
		},
		HelpStr: helpStrArgs{
			explanation: "`state_dict` returns a map of dotted names to the tensors in a model, walking maps, lists, tensors, and module parameters",
			signature:   "state_dict(model: map|list|tensor|module) -> map[str]tensor",
			errors:      "InvalidArgCount,PositionalType",
			example:     "state_dict({'w': ml.randn([2, 2], requires_grad=true)}) => {'w': tensor}",
		}.String(),
	},
	{
		Name: "_save_state",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("save_state", 2, args); err != nil {
				return err
			}
			path, ok := args[1].(*Stringo)
			if !ok {
				return newPositionalTypeError("save_state", 2, STRING_OBJ, args[1].Type())
			}
			state, errObj := stateDictOf(args[0])
			if errObj != nil {
				return errObj
			}
			entries := make([]ml.NamedTensor, 0, state.Len())
			for _, k := range state.Keys {
				v, ok := state.Get(k)
				if !ok {
					continue
				}
				t, ok := v.(*Tensor)
				if !ok {
					continue
				}
				entries = append(entries, ml.NamedTensor{Name: k, Tensor: t.T})
			}
			if err := ml.SaveState(entries, path.Value); err != nil {
				return newError("`save_state` error: %s", err.Error())
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`save_state` writes a model's named tensors to one file",
			signature:   "save_state(model: map|list|tensor|module, path: str) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "save_state(model, 'model.state') => null",
		}.String(),
	},
	{
		Name: "_load_state",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("load_state", 2, args); err != nil {
				return err
			}
			path, ok := args[1].(*Stringo)
			if !ok {
				return newPositionalTypeError("load_state", 2, STRING_OBJ, args[1].Type())
			}
			loaded, err := ml.LoadState(path.Value)
			if err != nil {
				return newError("`load_state` error: %s", err.Error())
			}
			state, errObj := stateDictOf(args[0])
			if errObj != nil {
				return errObj
			}
			for _, nt := range loaded {
				v, ok := state.Get(nt.Name)
				if !ok {
					continue
				}
				dst, ok := v.(*Tensor)
				if !ok {
					continue
				}
				if err := ml.CopyInto(dst.T, nt.Tensor); err != nil {
					return newError("`load_state` error for %q: %s", nt.Name, err.Error())
				}
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`load_state` loads named tensors from a file into a model in place, matching names; missing names are skipped and shape mismatches error",
			signature:   "load_state(model: map|list|tensor|module, path: str) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "load_state(model, 'model.state') => null",
		}.String(),
	},
	{
		Name: "_flip",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("flip", 2, args); err != nil {
				return err
			}
			if err := checkArgType("flip", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			l, ok := args[1].(*List)
			if !ok {
				return newPositionalTypeError("flip", 2, LIST_OBJ, args[1].Type())
			}
			dims, err := toIntList("flip", l)
			if err != nil {
				return newError("%s", err.Error())
			}
			out, ferr := ml.Flip(args[0].(*Tensor).T, dims)
			if ferr != nil {
				return newError("`flip` error: %s", ferr.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`flip` reverses elements along each dim, matching torch.flip; it is differentiable",
			signature:   "flip(a: tensor, dims: list[int]) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "flip(ml.tensor([1.0, 2.0, 3.0]), [0]) => Tensor{shape: [3]}",
		}.String(),
	},
	{
		Name: "_masked_select",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("masked_select", 2, args); err != nil {
				return err
			}
			if err := checkArgType("masked_select", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			if err := checkArgType("masked_select", 2, TENSOR_OBJ, args); err != nil {
				return err
			}
			out, err := ml.MaskedSelect(args[0].(*Tensor).T, args[1].(*Tensor).T)
			if err != nil {
				return newError("`masked_select` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`masked_select` returns the elements where mask is nonzero as a 1d tensor; inference-only because the output length depends on data",
			signature:   "masked_select(a: tensor, mask: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "masked_select(ml.tensor([1.0, 2.0]), ml.tensor([0.0, 1.0])) => Tensor{shape: [1]}",
		}.String(),
	},
	// Generic dispatchers. The named ml.* functions in lib/std/ml.b are thin
	// wrappers over these, so one builtin covers a whole family instead of one
	// builtin per op. New ops are added here, not as new entries.
	{
		Name: "_binary",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("binary", 3, args); err != nil {
				return err
			}
			op, ok := args[0].(*Stringo)
			if !ok {
				return newPositionalTypeError("binary", 1, STRING_OBJ, args[0].Type())
			}
			device := ml.CPU
			if t, ok := args[1].(*Tensor); ok {
				device = t.T.Device()
			} else if t, ok := args[2].(*Tensor); ok {
				device = t.T.Device()
			}
			a, ok := asTensorArg(args[1], device)
			if !ok {
				return newPositionalTypeError("binary", 2, TENSOR_OBJ, args[1].Type())
			}
			b, ok := asTensorArg(args[2], device)
			if !ok {
				return newPositionalTypeError("binary", 3, TENSOR_OBJ, args[2].Type())
			}
			var out *ml.Tensor
			var ferr error
			switch op.Value {
			case "add":
				out, ferr = ml.Add(a, b)
			case "sub":
				out, ferr = ml.Sub(a, b)
			case "mul":
				out, ferr = ml.Mul(a, b)
			case "div":
				out, ferr = ml.Div(a, b)
			case "pow":
				out, ferr = ml.Pow(a, b)
			case "matmul":
				out, ferr = ml.MatMul(a, b)
			case "bmm":
				out, ferr = ml.BMM(a, b)
			default:
				return newError("`binary` error: unknown op %q", op.Value)
			}
			if ferr != nil {
				return newError("`%s` error: %s", op.Value, ferr.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`binary` dispatches an elementwise or matrix binary op by name",
			signature:   "binary(op: str, a: tensor, b: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "binary('add', ml.tensor([1.0]), ml.tensor([2.0])) => Tensor{shape: [1]}",
		}.String(),
	},
	{
		Name: "_compare",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("compare", 3, args); err != nil {
				return err
			}
			op, ok := args[0].(*Stringo)
			if !ok {
				return newPositionalTypeError("compare", 1, STRING_OBJ, args[0].Type())
			}
			device := ml.CPU
			if t, ok := args[1].(*Tensor); ok {
				device = t.T.Device()
			} else if t, ok := args[2].(*Tensor); ok {
				device = t.T.Device()
			}
			a, ok := asTensorArg(args[1], device)
			if !ok {
				return newPositionalTypeError("compare", 2, TENSOR_OBJ, args[1].Type())
			}
			b, ok := asTensorArg(args[2], device)
			if !ok {
				return newPositionalTypeError("compare", 3, TENSOR_OBJ, args[2].Type())
			}
			var out *ml.Tensor
			var ferr error
			switch op.Value {
			case "eq":
				out, ferr = ml.Eq(a, b)
			case "ne":
				out, ferr = ml.Ne(a, b)
			case "gt":
				out, ferr = ml.Gt(a, b)
			case "ge":
				out, ferr = ml.Ge(a, b)
			case "lt":
				out, ferr = ml.Lt(a, b)
			case "le":
				out, ferr = ml.Le(a, b)
			default:
				return newError("`compare` error: unknown op %q", op.Value)
			}
			if ferr != nil {
				return newError("`%s` error: %s", op.Value, ferr.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`compare` dispatches an elementwise comparison by name, returning a bool tensor",
			signature:   "compare(op: str, a: tensor, b: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "compare('gt', ml.tensor([2.0]), 0.0) => Tensor{shape: [1]}",
		}.String(),
	},
	{
		Name: "_unary",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("unary", 2, args); err != nil {
				return err
			}
			op, ok := args[0].(*Stringo)
			if !ok {
				return newPositionalTypeError("unary", 1, STRING_OBJ, args[0].Type())
			}
			if err := checkArgType("unary", 2, TENSOR_OBJ, args); err != nil {
				return err
			}
			a := args[1].(*Tensor).T
			var out *ml.Tensor
			var ferr error
			switch op.Value {
			case "neg":
				out, ferr = ml.Neg(a)
			case "relu":
				out, ferr = ml.Relu(a)
			case "exp":
				out, ferr = ml.Exp(a)
			case "log":
				out, ferr = ml.Log(a)
			case "sqrt":
				out, ferr = ml.Sqrt(a)
			case "abs":
				out, ferr = ml.Abs(a)
			case "sigmoid":
				out, ferr = ml.Sigmoid(a)
			case "tanh":
				out, ferr = ml.Tanh(a)
			case "silu":
				out, ferr = ml.Silu(a)
			default:
				return newError("`unary` error: unknown op %q", op.Value)
			}
			if ferr != nil {
				return newError("`%s` error: %s", op.Value, ferr.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`unary` dispatches an elementwise unary op by name",
			signature:   "unary(op: str, a: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "unary('relu', ml.tensor([-1.0, 2.0])) => Tensor{shape: [2]}",
		}.String(),
	},
	{
		Name: "_reduce",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("reduce", 4, args); err != nil {
				return err
			}
			op, ok := args[0].(*Stringo)
			if !ok {
				return newPositionalTypeError("reduce", 1, STRING_OBJ, args[0].Type())
			}
			if err := checkArgType("reduce", 2, TENSOR_OBJ, args); err != nil {
				return err
			}
			t := args[1].(*Tensor).T
			dim, keepdim := args[2], args[3]
			switch op.Value {
			case "sum":
				return tensorReduction("sum", t, dim, keepdim, ml.Sum)
			case "mean":
				return tensorReduction("mean", t, dim, keepdim, ml.Mean)
			case "max":
				return tensorReduction("max", t, dim, keepdim, ml.Max)
			case "min":
				return tensorReduction("min", t, dim, keepdim, ml.Min)
			default:
				return newError("`reduce` error: unknown op %q", op.Value)
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`reduce` dispatches a reduction by name over a dim or list of dims, with keepdim",
			signature:   "reduce(op: str, a: tensor, dim: int|list[int]|null, keepdim: bool) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "reduce('sum', ml.tensor([1.0, 2.0]), null, false) => Tensor{shape: [1]}",
		}.String(),
	},
	{
		Name: "_argreduce",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("argreduce", 3, args); err != nil {
				return err
			}
			op, ok := args[0].(*Stringo)
			if !ok {
				return newPositionalTypeError("argreduce", 1, STRING_OBJ, args[0].Type())
			}
			if err := checkArgType("argreduce", 2, TENSOR_OBJ, args); err != nil {
				return err
			}
			d, ok := args[2].(*Integer)
			if !ok {
				return newPositionalTypeError("argreduce", 3, INTEGER_OBJ, args[2].Type())
			}
			t := args[1].(*Tensor).T
			var out *ml.Tensor
			var ferr error
			switch op.Value {
			case "argmax":
				out, ferr = ml.ArgMax(t, int(d.Value))
			case "argmin":
				out, ferr = ml.ArgMin(t, int(d.Value))
			default:
				return newError("`argreduce` error: unknown op %q", op.Value)
			}
			if ferr != nil {
				return newError("`%s` error: %s", op.Value, ferr.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`argreduce` dispatches an index reduction by name",
			signature:   "argreduce(op: str, a: tensor, dim: int) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "argreduce('argmax', ml.tensor([1.0, 3.0]), 0) => Tensor{shape: [1]}",
		}.String(),
	},
	{
		Name: "_nn_conv2d",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("conv2d", 7, args); err != nil {
				return err
			}
			ints := make([]int, 5)
			for i := range ints {
				n, ok := args[i].(*Integer)
				if !ok {
					return newPositionalTypeError("conv2d", i+1, INTEGER_OBJ, args[i].Type())
				}
				ints[i] = int(n.Value)
			}
			useBias, ok := args[5].(*Boolean)
			if !ok {
				return newPositionalTypeError("conv2d", 6, BOOLEAN_OBJ, args[5].Type())
			}
			devStr, ok := args[6].(*Stringo)
			if !ok {
				return newPositionalTypeError("conv2d", 7, STRING_OBJ, args[6].Type())
			}
			dev, derr := ml.ParseDevice(devStr.Value)
			if derr != nil {
				return newError("`conv2d` error: %s", derr.Error())
			}
			m, err := ml.NNConv2D(ints[0], ints[1], ints[2], ints[3], ints[4], useBias.Value, dev)
			if err != nil {
				return newError("`conv2d` error: %s", err.Error())
			}
			return &GoObj[ml.Module]{Value: m}
		},
		HelpStr: helpStrArgs{
			explanation: "`conv2d` builds a 2D convolution module with a square kernel",
			signature:   "conv2d(in_channels: int, out_channels: int, kernel_size: int, stride: int, padding: int, bias: bool, dev: str) -> module",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "conv2d(1, 8, 3, 1, 1, true, 'cpu') => module",
		}.String(),
	},
	{
		Name: "_nn_maxpool2d",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("maxpool2d", 3, args); err != nil {
				return err
			}
			kernel, ok := args[0].(*Integer)
			if !ok {
				return newPositionalTypeError("maxpool2d", 1, INTEGER_OBJ, args[0].Type())
			}
			stride, ok := args[1].(*Integer)
			if !ok {
				return newPositionalTypeError("maxpool2d", 2, INTEGER_OBJ, args[1].Type())
			}
			devStr, ok := args[2].(*Stringo)
			if !ok {
				return newPositionalTypeError("maxpool2d", 3, STRING_OBJ, args[2].Type())
			}
			dev, derr := ml.ParseDevice(devStr.Value)
			if derr != nil {
				return newError("`maxpool2d` error: %s", derr.Error())
			}
			m, err := ml.NNMaxPool2D(int(kernel.Value), int(stride.Value), dev)
			if err != nil {
				return newError("`maxpool2d` error: %s", err.Error())
			}
			return &GoObj[ml.Module]{Value: m}
		},
		HelpStr: helpStrArgs{
			explanation: "`maxpool2d` builds a 2D max pooling module",
			signature:   "maxpool2d(kernel_size: int, stride: int, dev: str) -> module",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "maxpool2d(2, 2, 'cpu') => module",
		}.String(),
	},
	{
		Name: "_max_dim",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("max_dim", 3, args); err != nil {
				return err
			}
			if err := checkArgType("max_dim", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			d, ok := args[1].(*Integer)
			if !ok {
				return newPositionalTypeError("max_dim", 2, INTEGER_OBJ, args[1].Type())
			}
			keep, ok := args[2].(*Boolean)
			if !ok {
				return newPositionalTypeError("max_dim", 3, BOOLEAN_OBJ, args[2].Type())
			}
			v, i, err := ml.MaxDim(args[0].(*Tensor).T, int(d.Value), keep.Value)
			if err != nil {
				return newError("`max_dim` error: %s", err.Error())
			}
			return &List{Elements: []Object{&Tensor{T: v}, &Tensor{T: i}}}
		},
		HelpStr: helpStrArgs{
			explanation: "`max_dim` returns [values, indices] for the maximum along dim, like torch.max(x, dim)",
			signature:   "max_dim(a: tensor, dim: int, keepdim: bool) -> list[tensor, tensor]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "max_dim(ml.tensor([1.0, 3.0]), 0, false) => [tensor, tensor]",
		}.String(),
	},
	{
		Name: "_min_dim",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("min_dim", 3, args); err != nil {
				return err
			}
			if err := checkArgType("min_dim", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			d, ok := args[1].(*Integer)
			if !ok {
				return newPositionalTypeError("min_dim", 2, INTEGER_OBJ, args[1].Type())
			}
			keep, ok := args[2].(*Boolean)
			if !ok {
				return newPositionalTypeError("min_dim", 3, BOOLEAN_OBJ, args[2].Type())
			}
			v, i, err := ml.MinDim(args[0].(*Tensor).T, int(d.Value), keep.Value)
			if err != nil {
				return newError("`min_dim` error: %s", err.Error())
			}
			return &List{Elements: []Object{&Tensor{T: v}, &Tensor{T: i}}}
		},
		HelpStr: helpStrArgs{
			explanation: "`min_dim` returns [values, indices] for the minimum along dim, like torch.min(x, dim)",
			signature:   "min_dim(a: tensor, dim: int, keepdim: bool) -> list[tensor, tensor]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "min_dim(ml.tensor([1.0, 3.0]), 0, false) => [tensor, tensor]",
		}.String(),
	},
	{
		Name: "_variance",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("variance", 4, args); err != nil {
				return err
			}
			if err := checkArgType("variance", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			dims, errObj := dimsFromObject("variance", args[1])
			if errObj != nil {
				return errObj
			}
			keep, ok := args[2].(*Boolean)
			if !ok {
				return newPositionalTypeError("variance", 3, BOOLEAN_OBJ, args[2].Type())
			}
			unbiased, ok := args[3].(*Boolean)
			if !ok {
				return newPositionalTypeError("variance", 4, BOOLEAN_OBJ, args[3].Type())
			}
			out, err := ml.Variance(args[0].(*Tensor).T, dims, keep.Value, unbiased.Value)
			if err != nil {
				return newError("`variance` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`variance` is the variance over dim; unbiased divides by n-1 like torch.var(unbiased=True)",
			signature:   "variance(a: tensor, dim: int|list[int]|null, keepdim: bool, unbiased: bool) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "variance(ml.tensor([1.0, 2.0, 3.0]), null, false, true) => Tensor{shape: [1]}",
		}.String(),
	},
	{
		Name: "_std",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("std", 4, args); err != nil {
				return err
			}
			if err := checkArgType("std", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			dims, errObj := dimsFromObject("std", args[1])
			if errObj != nil {
				return errObj
			}
			keep, ok := args[2].(*Boolean)
			if !ok {
				return newPositionalTypeError("std", 3, BOOLEAN_OBJ, args[2].Type())
			}
			unbiased, ok := args[3].(*Boolean)
			if !ok {
				return newPositionalTypeError("std", 4, BOOLEAN_OBJ, args[3].Type())
			}
			out, err := ml.Std(args[0].(*Tensor).T, dims, keep.Value, unbiased.Value)
			if err != nil {
				return newError("`std` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`std` is the square root of variance; unbiased divides by n-1",
			signature:   "std(a: tensor, dim: int|list[int]|null, keepdim: bool, unbiased: bool) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "std(ml.tensor([1.0, 2.0, 3.0]), null, false, true) => Tensor{shape: [1]}",
		}.String(),
	},
	{
		Name: "_tile",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("tile", 2, args); err != nil {
				return err
			}
			if err := checkArgType("tile", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			l, ok := args[1].(*List)
			if !ok {
				return newPositionalTypeError("tile", 2, LIST_OBJ, args[1].Type())
			}
			reps, err := toIntList("tile", l)
			if err != nil {
				return newError("%s", err.Error())
			}
			out, ferr := ml.Tile(args[0].(*Tensor).T, reps)
			if ferr != nil {
				return newError("`tile` error: %s", ferr.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`tile` repeats a along each dim, like torch.tile; it is differentiable",
			signature:   "tile(a: tensor, reps: list[int]) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "tile(ml.tensor([[1.0, 2.0]]), [2, 1]) => Tensor{shape: [2 2]}",
		}.String(),
	},
	{
		Name: "_cumsum",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("cumsum", 2, args); err != nil {
				return err
			}
			if err := checkArgType("cumsum", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			d, ok := args[1].(*Integer)
			if !ok {
				return newPositionalTypeError("cumsum", 2, INTEGER_OBJ, args[1].Type())
			}
			out, err := ml.Cumsum(args[0].(*Tensor).T, int(d.Value))
			if err != nil {
				return newError("`cumsum` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`cumsum` is the cumulative sum along dim, like torch.cumsum; it is differentiable",
			signature:   "cumsum(a: tensor, dim: int) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "cumsum(ml.tensor([1.0, 2.0, 3.0]), 0) => Tensor{shape: [3]}",
		}.String(),
	},
	{
		Name: "_sort",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("sort", 3, args); err != nil {
				return err
			}
			if err := checkArgType("sort", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			d, ok := args[1].(*Integer)
			if !ok {
				return newPositionalTypeError("sort", 2, INTEGER_OBJ, args[1].Type())
			}
			desc, ok := args[2].(*Boolean)
			if !ok {
				return newPositionalTypeError("sort", 3, BOOLEAN_OBJ, args[2].Type())
			}
			v, i, err := ml.Sort(args[0].(*Tensor).T, int(d.Value), desc.Value)
			if err != nil {
				return newError("`sort` error: %s", err.Error())
			}
			return &List{Elements: []Object{&Tensor{T: v}, &Tensor{T: i}}}
		},
		HelpStr: helpStrArgs{
			explanation: "`sort` returns [values, indices] sorted along dim; host-computed, so inference-only",
			signature:   "sort(a: tensor, dim: int, descending: bool) -> list[tensor, tensor]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "sort(ml.tensor([3.0, 1.0, 2.0]), 0, false) => [tensor, tensor]",
		}.String(),
	},
	{
		Name: "_topk",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("topk", 4, args); err != nil {
				return err
			}
			if err := checkArgType("topk", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			k, ok := args[1].(*Integer)
			if !ok {
				return newPositionalTypeError("topk", 2, INTEGER_OBJ, args[1].Type())
			}
			d, ok := args[2].(*Integer)
			if !ok {
				return newPositionalTypeError("topk", 3, INTEGER_OBJ, args[2].Type())
			}
			largest, ok := args[3].(*Boolean)
			if !ok {
				return newPositionalTypeError("topk", 4, BOOLEAN_OBJ, args[3].Type())
			}
			v, i, err := ml.TopK(args[0].(*Tensor).T, int(k.Value), int(d.Value), largest.Value)
			if err != nil {
				return newError("`topk` error: %s", err.Error())
			}
			return &List{Elements: []Object{&Tensor{T: v}, &Tensor{T: i}}}
		},
		HelpStr: helpStrArgs{
			explanation: "`topk` returns [values, indices] of the k largest or smallest along dim; host-computed, so inference-only",
			signature:   "topk(a: tensor, k: int, dim: int, largest: bool) -> list[tensor, tensor]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "topk(ml.tensor([1.0, 3.0, 2.0]), 2, 0, true) => [tensor, tensor]",
		}.String(),
	},
	{
		Name: "_nonzero",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("nonzero", 1, args); err != nil {
				return err
			}
			if err := checkArgType("nonzero", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			out, err := ml.Nonzero(args[0].(*Tensor).T)
			if err != nil {
				return newError("`nonzero` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`nonzero` returns the indices of the nonzero elements as [n, rank]; host-computed, so inference-only",
			signature:   "nonzero(a: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "nonzero(ml.tensor([0.0, 2.0, 0.0, 3.0])) => Tensor{shape: [2 1]}",
		}.String(),
	},
	{
		Name: "_scatter_add",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("scatter_add", 4, args); err != nil {
				return err
			}
			if _, ok := args[0].(*Tensor); !ok {
				return newPositionalTypeError("scatter_add", 1, TENSOR_OBJ, args[0].Type())
			}
			d, ok := args[1].(*Integer)
			if !ok {
				return newPositionalTypeError("scatter_add", 2, INTEGER_OBJ, args[1].Type())
			}
			if _, ok := args[2].(*Tensor); !ok {
				return newPositionalTypeError("scatter_add", 3, TENSOR_OBJ, args[2].Type())
			}
			if _, ok := args[3].(*Tensor); !ok {
				return newPositionalTypeError("scatter_add", 4, TENSOR_OBJ, args[3].Type())
			}
			out, err := ml.ScatterAdd(args[0].(*Tensor).T, int(d.Value), args[2].(*Tensor).T, args[3].(*Tensor).T)
			if err != nil {
				return newError("`scatter_add` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`scatter_add` returns dest with src added at the positions selected by index along dim",
			signature:   "scatter_add(dest: tensor, dim: int, index: tensor, src: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "scatter_add(ml.zeros([3, 2]), 0, ml.tensor([[0,1]], datatype=ml.dtype.int32), ml.ones([1, 2])) => tensor",
		}.String(),
	},
	{
		Name: "_cross_entropy_opts",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("cross_entropy", 5, args); err != nil {
				return err
			}
			if err := checkArgType("cross_entropy", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			if err := checkArgType("cross_entropy", 2, TENSOR_OBJ, args); err != nil {
				return err
			}
			opts := ml.CrossEntropyOpts{Reduction: "mean"}
			s, ok := args[2].(*Stringo)
			if !ok {
				return newPositionalTypeError("cross_entropy", 3, STRING_OBJ, args[2].Type())
			}
			opts.Reduction = s.Value
			if _, isNull := args[3].(*Null); !isNull {
				n, ok := args[3].(*Integer)
				if !ok {
					return newPositionalTypeError("cross_entropy", 4, "INTEGER or NULL", args[3].Type())
				}
				opts.HasIgnore = true
				opts.IgnoreIndex = int(n.Value)
			}
			if _, isNull := args[4].(*Null); !isNull {
				w, ok := args[4].(*Tensor)
				if !ok {
					return newPositionalTypeError("cross_entropy", 5, "TENSOR or NULL", args[4].Type())
				}
				opts.Weight = w.T
			}
			out, err := ml.CrossEntropyWith(args[0].(*Tensor).T, args[1].(*Tensor).T, opts)
			if err != nil {
				return newError("`cross_entropy` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`cross_entropy` with reduction, ignore_index, and class weights, like torch.nn.functional.cross_entropy",
			signature:   "cross_entropy(logits: tensor, target: tensor, reduction: str, ignore_index: int|null, weight: tensor|null) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "cross_entropy(logits, target, 'mean', null, null) => tensor",
		}.String(),
	},
	{
		Name: "_optim_state_dict",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("optim_state_dict", 1, args); err != nil {
				return err
			}
			opt, ok := args[0].(*GoObj[*ml.Optimizer])
			if !ok {
				return newPositionalTypeErrorForGoObj("optim_state_dict", 1, "*ml.Optimizer", args[0])
			}
			out := NewOrderedMap[string, Object]()
			for k, t := range opt.Value.StateDict() {
				out.Set(k, &Tensor{T: t})
			}
			return CreateMapObjectForGoMap(*out)
		},
		HelpStr: helpStrArgs{
			explanation: "`optim_state_dict` returns the optimizer's internal buffers as a name to tensor map; the tensors share storage, so saving and loading them checkpoints the optimizer",
			signature:   "optim_state_dict(optimizer: GoObj[*ml.Optimizer]) -> map[str]tensor",
			errors:      "InvalidArgCount,PositionalType",
			example:     "optim_state_dict(opt) => {'m.0': tensor, ...}",
		}.String(),
	},
	{
		Name: "_retain_grad",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("retain_grad", 1, args); err != nil {
				return err
			}
			if err := checkArgType("retain_grad", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			args[0].(*Tensor).T.RetainGrad()
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`retain_grad` keeps a non-leaf tensor's gradient after backward, like torch.Tensor.retain_grad",
			signature:   "retain_grad(a: tensor) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "retain_grad(ml.mul(x, x)) => null",
		}.String(),
	},
	{
		Name: "_autograd_grad",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("autograd_grad", 2, args); err != nil {
				return err
			}
			if err := checkArgType("autograd_grad", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			l, ok := args[1].(*List)
			if !ok {
				return newPositionalTypeError("autograd_grad", 2, LIST_OBJ, args[1].Type())
			}
			inputs := make([]*ml.Tensor, len(l.Elements))
			for i, e := range l.Elements {
				t, ok := e.(*Tensor)
				if !ok {
					return newPositionalTypeError("autograd_grad", 2, TENSOR_OBJ, e.Type())
				}
				inputs[i] = t.T
			}
			grads, err := ml.AutogradGrad(args[0].(*Tensor).T, inputs)
			if err != nil {
				return newError("`autograd_grad` error: %s", err.Error())
			}
			elems := make([]Object, len(grads))
			for i, g := range grads {
				if g == nil {
					elems[i] = NULL
				} else {
					elems[i] = &Tensor{T: g}
				}
			}
			return &List{Elements: elems}
		},
		HelpStr: helpStrArgs{
			explanation: "`autograd_grad` returns d(output)/d(input) for each input without storing them, like torch.autograd.grad",
			signature:   "autograd_grad(output: tensor, inputs: list[tensor]) -> list[tensor|null]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "autograd_grad(ml.sum(y), [x]) => [tensor]",
		}.String(),
	},
	{
		Name: "_contiguous",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("contiguous", 1, args); err != nil {
				return err
			}
			if err := checkArgType("contiguous", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			return &Tensor{T: args[0].(*Tensor).T.Contiguous()}
		},
		HelpStr: helpStrArgs{
			explanation: "`contiguous` returns a tensor with row-major contiguous storage, materializing a view if needed",
			signature:   "contiguous(a: tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType",
			example:     "contiguous(ml.transpose(x, 0, 1)) => tensor",
		}.String(),
	},
	{
		Name: "_view",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("view", 2, args); err != nil {
				return err
			}
			if err := checkArgType("view", 1, TENSOR_OBJ, args); err != nil {
				return err
			}
			l, ok := args[1].(*List)
			if !ok {
				return newPositionalTypeError("view", 2, LIST_OBJ, args[1].Type())
			}
			shape, err := toIntList("view", l)
			if err != nil {
				return newError("%s", err.Error())
			}
			out, ferr := ml.View(args[0].(*Tensor).T, shape)
			if ferr != nil {
				return newError("`view` error: %s", ferr.Error())
			}
			return &Tensor{T: out}
		},
		HelpStr: helpStrArgs{
			explanation: "`view` returns a storage-sharing view with the given shape; inference-only, so it works under no_grad",
			signature:   "view(a: tensor, shape: list[int]) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "view(ml.zeros([4]), [2, 2]) => Tensor{shape: [2 2]}",
		}.String(),
	},
	{
		Name: "_from_matrix",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("from_matrix", 1, args); err != nil {
				return err
			}
			var data []float32
			var shape []int
			switch v := args[0].(type) {
			case *GoObj[*NumMatrix]:
				if v.Value == nil {
					return newError("`from_matrix` error: nil matrix")
				}
				r, c := v.Value.Dims()
				data = matrixToFloat32Data(v.Value)
				shape = []int{r, c}
			case *List:
				m, err := listToMatrix(v)
				if err != nil {
					return newError("`from_matrix` error: %s", err.Error())
				}
				r, c := m.Dims()
				data = matrixToFloat32Data(m)
				shape = []int{r, c}
			case *Tensor:
				return args[0] // already a tensor, nothing to do
			default:
				return newPositionalTypeError("from_matrix", 1, "MATRIX, LIST, or TENSOR", args[0].Type())
			}
			t, err := ml.NewTensor(data, shape, ml.Float32, ml.CPU)
			if err != nil {
				return newError("`from_matrix` error: %s", err.Error())
			}
			return &Tensor{T: t}
		},
		HelpStr: helpStrArgs{
			explanation: "`from_matrix` converts a num matrix or a nested list into an ml float32 tensor, so num results feed straight into ml models",
			signature:   "from_matrix(m: matrix|list|tensor) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "from_matrix(num.eye(2)) => Tensor{shape: [2 2]}",
		}.String(),
	},
	{
		Name: "_from_list",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("from_list", 1, args); err != nil {
				return err
			}
			l, ok := args[0].(*List)
			if !ok {
				return newPositionalTypeError("from_list", 1, LIST_OBJ, args[0].Type())
			}
			data := make([]float32, len(l.Elements))
			for i, e := range l.Elements {
				f, ok := objectToFloat64(e)
				if !ok {
					return newPositionalTypeError("from_list", 1, "LIST of numbers", e.Type())
				}
				data[i] = float32(f)
			}
			t, err := ml.NewTensor(data, []int{len(data)}, ml.Float32, ml.CPU)
			if err != nil {
				return newError("`from_list` error: %s", err.Error())
			}
			return &Tensor{T: t}
		},
		HelpStr: helpStrArgs{
			explanation: "`from_list` converts a list of numbers into a 1d ml tensor, so num statistics and sampling feed straight into ml",
			signature:   "from_list(values: list) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "from_list([1.0, 2.0, 3.0]) => Tensor{shape: [3]}",
		}.String(),
	},
	{
		Name: "_from_matrix_f64",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("from_matrix_f64", 1, args); err != nil {
				return err
			}
			var data []float64
			var shape []int
			switch v := args[0].(type) {
			case *GoObj[*NumMatrix]:
				if v.Value == nil {
					return newError("`from_matrix_f64` error: nil matrix")
				}
				r, c := v.Value.Dims()
				data = matrixToFloat64Data(v.Value)
				shape = []int{r, c}
			case *List:
				m, err := listToMatrix(v)
				if err != nil {
					return newError("`from_matrix_f64` error: %s", err.Error())
				}
				r, c := m.Dims()
				data = matrixToFloat64Data(m)
				shape = []int{r, c}
			default:
				return newPositionalTypeError("from_matrix_f64", 1, "MATRIX or LIST", args[0].Type())
			}
			t, err := ml.NewFloat64Tensor(data, shape, ml.CPU)
			if err != nil {
				return newError("`from_matrix_f64` error: %s", err.Error())
			}
			return &Tensor{T: t}
		},
		HelpStr: helpStrArgs{
			explanation: "`from_matrix_f64` converts a num matrix into an ml float64 tensor, so a round trip through num is exact",
			signature:   "from_matrix_f64(m: matrix|list) -> tensor",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "from_matrix_f64(num.eye(2)) => Tensor{shape: [2 2]}",
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

// optimParams extracts optimizer bindings from a parameter list. Each element
// may be a raw tensor (a leaf the user owns) or a module parameter handle from
// ml.parameters. Raw tensors are bound directly, which is what lets custom
// models train without an nn module.
func optimParams(name string, o Object) ([]ml.OptimBinding, Object) {
	l, ok := o.(*List)
	if !ok {
		return nil, newPositionalTypeError(name, 1, LIST_OBJ, o.Type())
	}
	out := make([]ml.OptimBinding, len(l.Elements))
	for i, e := range l.Elements {
		switch v := e.(type) {
		case *Tensor:
			out[i] = ml.BindTensor(fmt.Sprintf("param_%d", i), v.T)
		case *GoObj[*ml.Param]:
			out[i] = ml.BindParam(v.Value)
		default:
			return nil, newPositionalTypeError(name, 1, "TENSOR or PARAM", e.Type())
		}
	}
	return out, nil
}

// collectParams walks a model (map, list, tensor, or nn module handle) and
// collects trainable leaves. Raw tensors with requires_grad are returned as-is;
// module parameters are returned as handles. Non-tensor values are skipped, so
// a model map may hold config and closures alongside weights. seen guards
// against cycles in self-referential maps and lists.
func collectParams(name string, o Object, seen map[Object]bool, out *[]Object) Object {
	switch v := o.(type) {
	case *Tensor:
		if v.T.RequiresGrad() {
			*out = append(*out, v)
		}
	case *Map:
		if seen[v] {
			return nil
		}
		seen[v] = true
		for _, k := range v.Pairs.Keys {
			mp, ok := v.Pairs.Get(k)
			if !ok {
				continue
			}
			if errObj := collectParams(name, mp.Value, seen, out); errObj != nil {
				return errObj
			}
		}
	case *List:
		if seen[v] {
			return nil
		}
		seen[v] = true
		for _, e := range v.Elements {
			if errObj := collectParams(name, e, seen, out); errObj != nil {
				return errObj
			}
		}
	case *GoObj[ml.Module]:
		for _, p := range ml.NNParameters(v.Value) {
			*out = append(*out, &GoObj[*ml.Param]{Value: p})
		}
	default:
		// Scalars, strings, functions, and null are not parameters.
	}
	return nil
}

// collectState walks a model and records every tensor it finds under a dotted
// name (map keys joined with ".", list indices as numbers). Module parameters
// are recorded as views that share the parameter's storage, so copying into the
// view updates the parameter in place. Non-string map keys and non-tensor
// values are skipped.
func collectState(name string, o Object, seen map[Object]bool, out *OrderedMap2[string, Object]) Object {
	switch v := o.(type) {
	case *Tensor:
		if name != "" {
			out.Set(name, v)
		}
	case *Map:
		if seen[v] {
			return nil
		}
		seen[v] = true
		for _, k := range v.Pairs.Keys {
			mp, ok := v.Pairs.Get(k)
			if !ok {
				continue
			}
			ks, ok := mp.Key.(*Stringo)
			if !ok {
				continue
			}
			child := ks.Value
			if name != "" {
				child = name + "." + ks.Value
			}
			// Internal keys such as a module's "__handle" are not part of the
			// user-facing name; descend into the value without adding the key.
			if strings.HasPrefix(ks.Value, "__") {
				child = name
			}
			if errObj := collectState(child, mp.Value, seen, out); errObj != nil {
				return errObj
			}
		}
	case *List:
		if seen[v] {
			return nil
		}
		seen[v] = true
		for i, e := range v.Elements {
			child := strconv.Itoa(i)
			if name != "" {
				child = name + "." + child
			}
			if errObj := collectState(child, e, seen, out); errObj != nil {
				return errObj
			}
		}
	case *GoObj[ml.Module]:
		for _, p := range ml.NNParameters(v.Value) {
			pn := p.Name()
			if name != "" {
				pn = name + "." + pn
			}
			out.Set(pn, &Tensor{T: ml.WrapRaw(p.Tensor().Backend(), p.Tensor().Raw())})
		}
	default:
		// Scalars, strings, functions, and null carry no state.
	}
	return nil
}

// stateDictOf builds the name to tensor map for a model.
func stateDictOf(model Object) (*OrderedMap2[string, Object], Object) {
	out := NewOrderedMap[string, Object]()
	if errObj := collectState("", model, map[Object]bool{}, out); errObj != nil {
		return nil, errObj
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
