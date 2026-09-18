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
	// methods (allow a.matmul(b) as an example)
	case "matmul":
		return t.binaryMethod("matmul", ml.MatMul), nil
	case "add":
		return t.binaryMethod("add", ml.Add), nil
	case "sub":
		return t.binaryMethod("sub", ml.Sub), nil
	case "mul":
		return t.binaryMethod("mul", ml.Mul), nil
	case "div":
		return t.binaryMethod("div", ml.Div), nil
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
	case "item":
		return t.noArgMethod("item", func() Object {
			v, err := t.T.Item()
			if err != nil {
				return newError("%s", err.Error())
			}
			return &Float{Value: float64(v)}
		}), nil
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

func (t *Tensor) binaryMethod(name string, f func(a, b *ml.Tensor) (*ml.Tensor, error)) *Builtin {
	return &Builtin{Name: name, Fun: func(args ...Object) Object {
		if len(args) != 1 {
			return newInvalidArgCountError(name, len(args), 1, "")
		}
		other, ok := args[0].(*Tensor)
		if !ok {
			return newPositionalTypeError(name, 1, TENSOR_OBJ, args[0].Type())
		}
		out, err := f(t.T, other.T)
		if err != nil {
			return newError("%s", err.Error())
		}
		return &Tensor{T: out}
	}}
}

func (t *Tensor) unaryMethod(name string, f func(a *ml.Tensor) (*ml.Tensor, error)) *Builtin {
	return &Builtin{Name: name, Fun: func(args ...Object) Object {
		if len(args) != 0 {
			return newInvalidArgCountError(name, len(args), 0, "")
		}
		out, err := f(t.T)
		if err != nil {
			return newError("%s", err.Error())
		}
		return &Tensor{T: out}
	}}
}

func (t *Tensor) noArgMethod(name string, f func() Object) *Builtin {
	return &Builtin{Name: name, Fun: func(args ...Object) Object {
		if len(args) != 0 {
			return newInvalidArgCountError(name, len(args), 0, "")
		}
		return f()
	}}
}
