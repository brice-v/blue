package ml

import (
	"fmt"
	"math"
	"slices"
)

type DType uint8

const (
	Float32 DType = iota
	Float64
	Int32
	Bool
)

func (d DType) String() string {
	switch d {
	case Float32:
		return "float32"
	case Float64:
		return "float64"
	case Int32:
		return "int32"
	case Bool:
		return "bool"
	default:
		panic(fmt.Sprintf("unsupported dtype: %d", d))
	}
}

type Device uint8

const (
	CPU Device = iota
	GPU
)

func (d Device) String() string {
	switch d {
	case CPU:
		return "cpu"
	case GPU:
		return "gpu"
	default:
		panic(fmt.Sprintf("unsupported device: %d", d))
	}
}

// Tensor is a strided view over a flat buffer
type Tensor struct {
	data    []float32 // backing store (starting with f32 only - eventually create storage struct based on dtype and put here)
	shape   []int
	strides []int
	offset  int
	dtype   DType
	device  Device

	gradState *gradState
}

func (t *Tensor) Data() []float32 {
	return t.data
}

func (t *Tensor) Shape() []int {
	return t.shape
}

func (t *Tensor) Strides() []int {
	return t.strides
}

func (t *Tensor) Offset() int {
	return t.offset
}

func (t *Tensor) DType() DType {
	return t.dtype
}

func (t *Tensor) Device() Device {
	return t.device
}

func (t *Tensor) String() string {
	return fmt.Sprintf("Tensor{shape: %v, strides: %v, offset: %d, dtype: %s, device: %s}", t.shape, t.strides, t.offset, t.dtype, t.device)
}

type gradState struct {
	requiresGrad bool
	Grad         *Tensor
	gradFn       func(outputGrad *Tensor) []*Tensor // grads for prev
	prev         []*Tensor
	op           string
}

type Backend interface {
	MatMul(a, b *Tensor) (*Tensor, error)
	Add(a, b *Tensor) (*Tensor, error)
	Sub(a, b *Tensor) (*Tensor, error)
	Mul(a, b *Tensor) (*Tensor, error)
	Div(a, b *Tensor) (*Tensor, error)
	Exp(a *Tensor) (*Tensor, error)
	Log(a *Tensor) (*Tensor, error)
	Sqrt(a *Tensor) (*Tensor, error)
	Sum(a *Tensor, dim int, keepdim bool) (*Tensor, error)
	Max(a *Tensor, dim int, keepdim bool) (*Tensor, error)
	Softmax(a *Tensor, dim int) (*Tensor, error)
	Reshape(a *Tensor, shape ...int) (*Tensor, error)
	Transpose(a *Tensor, dim0, dim1 int) (*Tensor, error)
}

// auto grad helpers

// IsLeaf reports whether t is a leaf, matching PyTorch's is_leaf: true when
// requires_grad is false, or when t was not produced by an op (no gradFn).
// Untracked tensors such as input data are therefore leaves by convention.
func (t *Tensor) IsLeaf() bool {
	return !t.RequiresGrad() || t.gradState.gradFn == nil
}

func (t *Tensor) RequiresGrad() bool {
	return t.gradState != nil && t.gradState.requiresGrad
}

func (t *Tensor) Grad() *Tensor {
	if t.gradState == nil {
		return nil
	}
	return t.gradState.Grad
}

func (t *Tensor) SetRequiresGrad(on bool) {
	if !on {
		if t.gradState != nil {
			t.gradState.requiresGrad = on
		}
		return
	}
	if t.gradState == nil {
		t.gradState = &gradState{}
	}
	t.gradState.requiresGrad = on
}

// Numel is the number of elements
// the product of the shape's dimensions
func (t *Tensor) Numel() int {
	return numelOf(t.shape)
}

// IsContiguous is true if walking through the logical indices in row major order
// visits contiguous slots of the backing slice
func (t *Tensor) IsContiguous() bool {
	expected := 1
	for i, v := range slices.Backward(t.shape) {
		if v != 1 && t.strides[i] != expected {
			return false
		}
		expected *= v
	}
	return true
}

// materialize returns a contiguous offset-0 view of the tensor
func (t *Tensor) materialize() *Tensor {
	if t.IsContiguous() {
		if t.offset == 0 {
			return t
		}
		// slice data on the offset so new tensor becomes offset-0
		return &Tensor{
			data:    t.data[t.offset:],
			shape:   slices.Clone(t.shape),
			strides: slices.Clone(t.strides),
			offset:  0,
			dtype:   t.dtype,
			device:  t.device,
		}
	}
	tt := newLike(t)
	copyStrided(tt.data, t)
	return tt
}

func getContiguousStridesFromShape(shape []int) []int {
	strides := make([]int, len(shape))
	acc := 1
	for i, v := range slices.Backward(shape) {
		strides[i] = acc
		acc *= v
	}
	return strides
}

func copyStrided(dst []float32, src *Tensor) {
	pos := 0
	var walk func(dim, srcIdx int)
	walk = func(dim, srcIdx int) {
		if dim == len(src.shape) {
			dst[pos] = src.data[srcIdx]
			pos++
			return
		}
		for i := 0; i < src.shape[dim]; i++ {
			walk(dim+1, srcIdx+i*src.strides[dim])
		}
	}
	walk(0, src.offset)
}

func (t *Tensor) Item() (float32, error) {
	if t.Numel() != 1 || t.offset < 0 || t.offset >= len(t.data) {
		return float32(math.NaN()), fmt.Errorf("Item: tensor does not hold exactly 1 element")
	}
	return t.data[t.offset], nil
}

// other helpers

func numelOf(shape []int) int {
	n := 1 // 0-d (empty shape) returns 1
	for _, d := range shape {
		n *= d
	}
	return n
}

// newLike does not copy data, returns equivalent of materialize'd tensor in with allocated space for data
func newLike(a *Tensor) *Tensor {
	return &Tensor{
		data:    make([]float32, a.Numel()),
		shape:   slices.Clone(a.shape),
		strides: getContiguousStridesFromShape(a.shape),
		dtype:   a.dtype,
		device:  a.device,
	}
}
