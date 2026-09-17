package object

import (
	"blue/ml"
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
