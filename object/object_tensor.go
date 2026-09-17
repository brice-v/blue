package object

import (
	"blue/ml"
	"fmt"
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
	return nil, fmt.Errorf("TODO: Tensor Encoding not supported yet")
}

func (t *Tensor) IType() iType {
	return i_TENSOR_OBJ
}

func (t *Tensor) Clone() Object {
	panic("TODO: Support Tensor Clone, likely need to implement clone function in ml/tensor.go")
}
