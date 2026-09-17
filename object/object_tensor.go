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

func sliceIntToListInteger(a []int) Object {
	es := make([]Object, len(a))
	for i, e := range a {
		es[i] = NewInteger(int64(e))
	}
	return &List{Elements: es}
}

func (t *Tensor) Get(property string) (Object, error) {
	switch property {
	case "shape":
		return sliceIntToListInteger(t.T.Shape()), nil
	case "strides":
		return sliceIntToListInteger(t.T.Strides()), nil
	case "T":
		// TODO: Cannot implement properly yet until ops.go is created and track is implemented
		tt, err := ml.DefaultBackend.Transpose(t.T, 0, 1)
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
	}
	return nil, fmt.Errorf("unsupported property on tensor: %s", property)
}
