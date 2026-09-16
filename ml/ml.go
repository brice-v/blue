package ml

import "slices"

type DType uint8

const (
	Float32 DType = iota
	Float64
	Int32
	Bool
)

type Device uint8

const (
	CPU Device = iota
	GPU
)

// Tensor is a strided view over a flat buffer
type Tensor struct {
	data    []float32 // backing store (starting with f32 only, TBD use generics for other dtypes?)
	shape   []int
	strides []int
	offset  int
	dtype   DType
	device  Device

	gradState *gradState
}

type gradState struct {
	requiresGrad bool
	Grad         *Tensor
	gradFn       func(outputGrad *Tensor) []*Tensor // grads for prev
	prev         []*Tensor
	op           string
}

type Backend interface {
	MatMul(a, b *Tensor) *Tensor
	Add(a, b *Tensor) *Tensor
	Sub(a, b *Tensor) *Tensor
	Mul(a, b *Tensor) *Tensor
	Div(a, b *Tensor) *Tensor
	Exp(a, b *Tensor) *Tensor
	Log(a, b *Tensor) *Tensor
	Sqrt(a, b *Tensor) *Tensor
	Sum(a, b *Tensor) *Tensor
	Max(a, b *Tensor) *Tensor
	Softmax(a, b *Tensor) *Tensor
	Reshape(a, b *Tensor) *Tensor
	Transpose(a, b *Tensor) *Tensor
}

// auto grad helpers

func (t *Tensor) IsLeaf() bool {
	return t.gradState == nil
}

func (t *Tensor) RequiresGrad() bool {
	return !t.IsLeaf() && t.gradState.requiresGrad
}

func (t *Tensor) Grad() *Tensor {
	if t.IsLeaf() {
		return nil
	}
	return t.gradState.Grad
}

func (t *Tensor) SetRequiresGrad(on bool) {
	if !on {
		if !t.IsLeaf() {
			t.gradState.requiresGrad = on
		}
		return
	}
	if t.IsLeaf() {
		t.gradState = &gradState{}
	}
	t.gradState.requiresGrad = on
}

// view helpers

func (t *Tensor) Shape() []int {
	return t.shape
}

func (t *Tensor) Strides() []int {
	return t.strides
}

// Numel is the number of elements
// the product of the shape's dimensions
func (t *Tensor) Numel() int {
	n := 1 // 0-d (empty shape) returns 1
	for _, d := range t.shape {
		n *= d
	}
	return n
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
	tt := &Tensor{
		data:    make([]float32, t.Numel()),
		shape:   slices.Clone(t.shape),
		strides: contiguousStrides(t.shape),
		offset:  0,
		dtype:   t.dtype,
		device:  t.device,
	}
	copyStrided(tt.data, t)
	return tt
}

func contiguousStrides(shape []int) []int {
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

func (t *Tensor) Item() float32 {
	if t.Numel() != 1 || t.offset < 0 || t.offset >= len(t.data) {
		panic("Item: tensor does not hold exactly 1 element")
	}
	return t.data[t.offset]
}
