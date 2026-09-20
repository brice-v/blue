package object

import (
	"blue/ml"
	"fmt"
	"hash/maphash"
)

const TENSOR_OBJ Type = "TENSOR"

type Tensor struct {
	T *ml.Tensor
}

func (t *Tensor) Type() Type {
	return TENSOR_OBJ
}

func (t *Tensor) Inspect() string {
	return t.T.String()
}

func (t *Tensor) Help() string {
	return createHelpStringForObject("Tensor", "is the object that represents tensor values", t)
}

func (t *Tensor) Encode() ([]byte, error) {
	return marshalObjectWrapper(t)
}

func (t *Tensor) IType() iType {
	return i_TENSOR_OBJ
}

func (t *Tensor) Clone() Object {
	return &Tensor{T: t.T.Clone()}
}

func (t *Tensor) hashTensor() uint64 {
	hasher := newHasher()
	maphash.WriteComparable(hasher, uint8(t.T.DType()))
	maphash.WriteComparable(hasher, uint8(t.T.Device()))
	for _, d := range t.T.Shape() {
		maphash.WriteComparable(hasher, d)
	}
	for _, v := range t.T.ContiguousData() {
		maphash.WriteComparable(hasher, v)
	}
	return hasher.Sum64()
}

func intList(a []int) Object {
	es := make([]Object, len(a))
	for i, e := range a {
		es[i] = NewInteger(int64(e))
	}
	return &List{Elements: es}
}

func (t *Tensor) Get(property string) (Object, error) {
	switch property {
	case "shape":
		return intList(t.T.Shape()), nil
	case "strides":
		return intList(t.T.Strides()), nil
	case "grad":
		if g := t.T.Grad(); g != nil {
			return &Tensor{T: g}, nil
		}
		return NULL, nil
	case "T":
		if len(t.T.Shape()) <= 1 {
			return t, nil
		}
		tt, err := ml.Transpose(t.T, 0, 1)
		if err != nil {
			return nil, err
		}
		return &Tensor{T: tt}, nil
	case "offset":
		return NewInteger(int64(t.T.Offset())), nil
	case "dtype":
		return &Stringo{Value: t.T.DType().String()}, nil
	case "device":
		return &Stringo{Value: t.T.Device().String()}, nil
	case "ndim":
		return NewInteger(int64(len(t.T.Shape()))), nil
	case "requires_grad":
		return nativeToBooleanObject(t.T.RequiresGrad()), nil
	// Binary arithmetic and comparisons have one spelling each: the operators
	// (+ - * / ** @ == != > >= < <=) and the ml.* functions. Methods are not
	// duplicated here, so there is exactly one way to add or compare tensors.
	case "neg":
		return t.unaryMethod("neg", ml.Neg), nil
	case "relu":
		return t.unaryMethod("relu", ml.Relu), nil
	case "exp":
		return t.unaryMethod("exp", ml.Exp), nil
	case "log":
		return t.unaryMethod("log", ml.Log), nil
	case "sqrt":
		return t.unaryMethod("sqrt", ml.Sqrt), nil
	case "backward":
		return t.noArgMethod("backward", func() Object {
			if err := t.T.Backward(); err != nil {
				return newError("%s", err.Error())
			}
			return NULL
		}), nil
	case "zero_grad":
		return t.noArgMethod("zero_grad", func() Object {
			t.T.ZeroGrad()
			return NULL
		}), nil
	case "to_list":
		return t.noArgMethod("to_list", func() Object {
			return t.ToList()
		}), nil
	case "item":
		return t.noArgMethod("item", func() Object {
			v, err := t.T.Item()
			if err != nil {
				return newError("%s", err.Error())
			}
			return &Float{Value: float64(v)}
		}), nil
	case "sum":
		return t.sumMethod(), nil
	case "mean":
		return t.reductionMethod("mean", ml.Mean), nil
	case "max":
		return t.reductionMethod("max", ml.Max), nil
	case "min":
		return t.reductionMethod("min", ml.Min), nil
	case "argmax":
		return t.intMethod("argmax", ml.ArgMax), nil
	case "argmin":
		return t.intMethod("argmin", ml.ArgMin), nil
	case "softmax":
		return t.intMethod("softmax", ml.Softmax), nil
	case "reshape":
		return t.listMethod("reshape", func(a *ml.Tensor, dims []int) (*ml.Tensor, error) {
			return ml.Reshape(a, dims...)
		}), nil
	case "permute":
		return t.listMethod("permute", func(a *ml.Tensor, dims []int) (*ml.Tensor, error) {
			return ml.Permute(a, dims...)
		}), nil
	case "broadcast_to":
		return t.listMethod("broadcast_to", ml.BroadcastTo), nil
	case "unsqueeze":
		return t.intMethod("unsqueeze", ml.Unsqueeze), nil
	case "squeeze":
		return t.squeezeMethod(), nil
	case "flatten":
		return t.unaryMethod("flatten", ml.Flatten), nil
	case "abs":
		return t.unaryMethod("abs", ml.Abs), nil
	case "sigmoid":
		return t.unaryMethod("sigmoid", ml.Sigmoid), nil
	case "tanh":
		return t.unaryMethod("tanh", ml.Tanh), nil
	case "detach":
		return t.noArgMethod("detach", func() Object {
			return &Tensor{T: t.T.Detach()}
		}), nil
	case "clone":
		return t.noArgMethod("clone", func() Object {
			return &Tensor{T: t.T.Clone()}
		}), nil
	case "retain_grad":
		return t.noArgMethod("retain_grad", func() Object {
			t.T.RetainGrad()
			return NULL
		}), nil
	case "to":
		return t.toMethod(), nil
	case "cast":
		return t.castMethod(), nil
	}
	return nil, fmt.Errorf("unsupported property on tensor: %s", property)
}

func (t *Tensor) Set(property string, val Object) error {
	switch property {
	case "requires_grad":
		b, ok := val.(*Boolean)
		if !ok {
			return fmt.Errorf("requires_grad must be set to boolean. got=%s", val.Type())
		}
		t.T.SetRequiresGrad(b.Value)
		return nil
	case "grad":
		g, ok := val.(*Tensor)
		if !ok {
			return fmt.Errorf("grad must be set to tensor. got=%s", val.Type())
		}
		t.T.SetGrad(g.T)
		return nil
	}
	return fmt.Errorf("unsupported property on tensor: %s", property)
}

// ToList returns the tensor's values as nested blue lists, one level per
// dimension, so the nesting matches the shape. A 0-D tensor returns a single
// value. Values match the tensor's dtype.
func (t *Tensor) ToList() Object {
	data := t.T.ContiguousData()
	pos := 0
	var build func(dims []int) Object
	build = func(dims []int) Object {
		if len(dims) == 0 {
			v := data[pos]
			pos++
			return scalarObject(t.T.DType(), v)
		}
		elems := make([]Object, dims[0])
		for i := range elems {
			elems[i] = build(dims[1:])
		}
		return &List{Elements: elems}
	}
	return build(t.T.Shape())
}

func (t *Tensor) unaryMethod(name string, f func(a *ml.Tensor) (*ml.Tensor, error)) *Builtin {
	return &Builtin{
		Name: name,
		Fun: func(args ...Object) Object {
			err := checkArgCount(name, 0, args)
			if err != nil {
				return err
			}
			out, ferr := f(t.T)
			if ferr != nil {
				return newError("%s", ferr.Error())
			}
			return &Tensor{T: out}
		},
	}
}

func (t *Tensor) noArgMethod(name string, f func() Object) *Builtin {
	return &Builtin{
		Name: name,
		Fun: func(args ...Object) Object {
			err := checkArgCount(name, 0, args)
			if err != nil {
				return err
			}
			return f()
		},
	}
}

// scalarObject converts a stored float32 into the blue value that matches the
// tensor's dtype.
func scalarObject(dt ml.DType, v float32) Object {
	switch dt {
	case ml.Bool:
		return nativeToBooleanObject(v != 0)
	case ml.Int32:
		return NewInteger(int64(v))
	default:
		return &Float{Value: float64(v)}
	}
}

func (t *Tensor) sumMethod() *Builtin {
	return &Builtin{
		Name: "sum",
		Fun: func(args ...Object) Object {
			vals, err := bindArgs("sum", args, "dim", "keepdim")
			if err != nil {
				return err
			}
			return tensorSum(t.T, vals["dim"], vals["keepdim"])
		},
	}
}

// reductionMethod builds a.sum/mean/max/min(dim=null, keepdim=false).
func (t *Tensor) reductionMethod(name string, f func(*ml.Tensor, []int, bool) (*ml.Tensor, error)) *Builtin {
	return &Builtin{
		Name: name,
		Fun: func(args ...Object) Object {
			vals, err := bindArgs(name, args, "dim", "keepdim")
			if err != nil {
				return err
			}
			return tensorReduction(name, t.T, vals["dim"], vals["keepdim"], f)
		},
	}
}

// intMethod builds a single-int-argument method like a.argmax(dim).
func (t *Tensor) intMethod(name string, f func(*ml.Tensor, int) (*ml.Tensor, error)) *Builtin {
	return &Builtin{
		Name: name,
		Fun: func(args ...Object) Object {
			if err := checkArgCount(name, 1, args); err != nil {
				return err
			}
			n, ok := args[0].(*Integer)
			if !ok {
				return newPositionalTypeError(name, 1, INTEGER_OBJ, args[0].Type())
			}
			out, err := f(t.T, int(n.Value))
			if err != nil {
				return newError("`%s` error: %s", name, err.Error())
			}
			return &Tensor{T: out}
		},
	}
}

// listMethod builds a method that takes one list[int], like a.reshape([2, 3]).
func (t *Tensor) listMethod(name string, f func(*ml.Tensor, []int) (*ml.Tensor, error)) *Builtin {
	return &Builtin{
		Name: name,
		Fun: func(args ...Object) Object {
			if err := checkArgCount(name, 1, args); err != nil {
				return err
			}
			l, ok := args[0].(*List)
			if !ok {
				return newPositionalTypeError(name, 1, LIST_OBJ, args[0].Type())
			}
			dims, err := toIntList(name, l)
			if err != nil {
				return newError("%s", err.Error())
			}
			out, ferr := f(t.T, dims)
			if ferr != nil {
				return newError("`%s` error: %s", name, ferr.Error())
			}
			return &Tensor{T: out}
		},
	}
}

// toMethod builds a.to(dev_or_dtype), like PyTorch's Tensor.to: a device name
// moves the tensor, a dtype name casts it.
func (t *Tensor) toMethod() *Builtin {
	return &Builtin{
		Name: "to",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("to", 1, args); err != nil {
				return err
			}
			s, ok := args[0].(*Stringo)
			if !ok {
				return newPositionalTypeError("to", 1, STRING_OBJ, args[0].Type())
			}
			if dev, derr := ml.ParseDevice(s.Value); derr == nil {
				out, err := t.T.To(dev)
				if err != nil {
					return newError("`to` error: %s", err.Error())
				}
				return &Tensor{T: out}
			}
			dt, derr := ml.ParseDType(s.Value)
			if derr != nil {
				return newError("`to` error: expected a device (cpu/gpu) or dtype (float32/float64/int32/int64/uint8/bool), got %q", s.Value)
			}
			out, err := t.T.Cast(dt)
			if err != nil {
				return newError("`to` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
	}
}

// castMethod builds a.cast(dtype), an explicit dtype conversion.
func (t *Tensor) castMethod() *Builtin {
	return &Builtin{
		Name: "cast",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("cast", 1, args); err != nil {
				return err
			}
			s, ok := args[0].(*Stringo)
			if !ok {
				return newPositionalTypeError("cast", 1, STRING_OBJ, args[0].Type())
			}
			dt, derr := ml.ParseDType(s.Value)
			if derr != nil {
				return newError("`cast` error: %s", derr.Error())
			}
			out, err := t.T.Cast(dt)
			if err != nil {
				return newError("`cast` error: %s", err.Error())
			}
			return &Tensor{T: out}
		},
	}
}

// squeezeMethod builds a.squeeze(dim=null): null drops every size-1 dim.
func (t *Tensor) squeezeMethod() *Builtin {
	return &Builtin{
		Name: "squeeze",
		Fun: func(args ...Object) Object {
			vals, err := bindArgs("squeeze", args, "dim")
			if err != nil {
				return err
			}
			var dims []int
			if d, ok := vals["dim"]; ok && d != nil {
				parsed, errObj := dimsFromObject("squeeze", d)
				if errObj != nil {
					return errObj
				}
				dims = parsed
			}
			out, ferr := ml.Squeeze(t.T, dims)
			if ferr != nil {
				return newError("`squeeze` error: %s", ferr.Error())
			}
			return &Tensor{T: out}
		},
	}
}
