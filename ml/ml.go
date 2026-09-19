package ml

import (
	"fmt"
	"math"
	"slices"
	"strings"
)

type DType uint8

const (
	Float32 DType = iota
	Float64
	Int32
	Bool
	Invalid
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

func ParseDType(s string) (DType, error) {
	s = strings.ToLower(s)
	switch s {
	case "float32":
		return Float32, nil
	case "float64":
		return Float64, nil
	case "int32":
		return Int32, nil
	case "bool":
		return Bool, nil
	default:
		return Invalid, fmt.Errorf("invalid DType: %s", s)
	}
}

type Device uint8

const (
	CPU Device = iota
	GPU
	INVALID
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

func ParseDevice(s string) (Device, error) {
	s = strings.ToLower(s)
	switch s {
	case "cpu":
		return CPU, nil
	case "gpu":
		return GPU, nil
	default:
		return INVALID, fmt.Errorf("invalid Device: %s", s)
	}
}

// Storage owns the backing buffer and the dtype/device it lives on. Views over
// the same buffer share one Storage; only a copy (materialize of a
// non-contiguous view, Clone, NewTensor) creates a new one.
type Storage struct {
	data   []float32
	dtype  DType
	device Device
}

// Tensor is a strided view over a Storage.
type Tensor struct {
	storage *Storage

	shape   []int
	strides []int
	offset  int

	gradState *gradState
}

func (t *Tensor) RawData() []float32 {
	return t.storage.data
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
	return t.storage.dtype
}

func (t *Tensor) Device() Device {
	return t.storage.device
}

func (t *Tensor) Clone() *Tensor {
	return &Tensor{
		storage: &Storage{
			data:   slices.Clone(t.storage.data),
			dtype:  t.storage.dtype,
			device: t.storage.device,
		},
		shape:   slices.Clone(t.shape),
		strides: slices.Clone(t.strides),
		offset:  t.offset,
	}
}

func (t *Tensor) SetGrad(tt *Tensor) {
	if t.gradState == nil {
		t.gradState = &gradState{}
	}
	t.gradState.Grad = tt
}

// ContiguousData returns the tensor's logical elements in row-major order.
// Non-contiguous views are packed into a fresh buffer, so the result always
// has length Numel(). The returned slice may alias the backing store when t
// is already contiguous with offset 0; callers must not mutate it in that
// case. Used by serialization, which only reads the values.
func (t *Tensor) ContiguousData() []float32 {
	m := t.materialize()
	return m.storage.data[:m.Numel()]
}

func checkNewTensorConstruction(data []float32, shape []int, dtype DType, device Device) error {
	if dtype > Bool || dtype < Float32 {
		return fmt.Errorf("NewTensor: unsupported dtype %d", dtype)
	}
	if device > GPU {
		return fmt.Errorf("NewTensor: unsupported device %d", device)
	}
	for _, d := range shape {
		if d < 0 {
			return fmt.Errorf("NewTensor: negative dimension %d in shape %v", d, shape)
		}
	}
	n := numelOf(shape)
	if len(data) != n {
		return fmt.Errorf("NewTensor: data has %d elements but shape %v needs %d", len(data), shape, n)
	}
	return nil
}

// NewTensor builds a contiguous, offset-0 tensor from data and shape. It
// copies data so the returned tensor owns its storage. An error is returned
// when a dimension is negative, when len(data) does not equal the product of
// shape, or when dtype/device are not known. This is the constructor the
// object package uses to rebuild a tensor after decoding.
func NewTensor(data []float32, shape []int, dtype DType, device Device) (*Tensor, error) {
	err := checkNewTensorConstruction(data, shape, dtype, device)
	if err != nil {
		return nil, err
	}
	return &Tensor{
		storage: &Storage{
			data:   slices.Clone(data),
			dtype:  dtype,
			device: device,
		},
		shape:   slices.Clone(shape),
		strides: getContiguousStridesFromShape(shape),
	}, nil
}

// NewTensorOwned is the same as above but without cloning data
// ownership will now be by this tensor so it must not be modified (note: this is currently only used by decode so its safe)
func NewTensorOwned(data []float32, shape []int, dtype DType, device Device) (*Tensor, error) {
	err := checkNewTensorConstruction(data, shape, dtype, device)
	if err != nil {
		return nil, err
	}
	return &Tensor{
		storage: &Storage{
			data:   data,
			dtype:  dtype,
			device: device,
		},
		shape:   shape,
		strides: getContiguousStridesFromShape(shape),
	}, nil
}

func (t *Tensor) String() string {
	return fmt.Sprintf("Tensor{shape: %v, strides: %v, offset: %d, dtype: %s, device: %s}", t.shape, t.strides, t.offset, t.storage.dtype, t.storage.device)
}

type gradState struct {
	requiresGrad bool
	Grad         *Tensor
	gradFn       func(outputGrad *Tensor) ([]*Tensor, error) // grads for prev
	prev         []*Tensor
	op           string
}

var backends = map[Device]Backend{}

func register(d Device, b Backend) {
	backends[d] = b
}

type Backend interface {
	MatMul(a, b *Tensor) (*Tensor, error)
	Add(a, b *Tensor) (*Tensor, error)
	Sub(a, b *Tensor) (*Tensor, error)
	Mul(a, b *Tensor) (*Tensor, error)
	Div(a, b *Tensor) (*Tensor, error)
	Neg(a *Tensor) (*Tensor, error)
	Exp(a *Tensor) (*Tensor, error)
	Log(a *Tensor) (*Tensor, error)
	Sqrt(a *Tensor) (*Tensor, error)
	Relu(a *Tensor) (*Tensor, error)
	Eq(a, b *Tensor) (*Tensor, error)
	Ne(a, b *Tensor) (*Tensor, error)
	Gt(a, b *Tensor) (*Tensor, error)
	Ge(a, b *Tensor) (*Tensor, error)
	Lt(a, b *Tensor) (*Tensor, error)
	Le(a, b *Tensor) (*Tensor, error)
	Pow(a, b *Tensor) (*Tensor, error)
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
			storage: &Storage{
				data:   t.storage.data[t.offset:],
				dtype:  t.storage.dtype,
				device: t.storage.device,
			},
			shape:   slices.Clone(t.shape),
			strides: slices.Clone(t.strides),
			offset:  0,
		}
	}
	tt := newLike(t)
	copyStrided(tt.storage.data, t)
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
			dst[pos] = src.storage.data[srcIdx]
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
	if t.Numel() != 1 || t.offset < 0 || t.offset >= len(t.storage.data) {
		return float32(math.NaN()), fmt.Errorf("Item: tensor does not hold exactly 1 element")
	}
	return t.storage.data[t.offset], nil
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
		storage: &Storage{
			data:   make([]float32, a.Numel()),
			dtype:  a.storage.dtype,
			device: a.storage.device,
		},
		shape:   slices.Clone(a.shape),
		strides: getContiguousStridesFromShape(a.shape),
	}
}

// backendForAll resolves the backend for a set of inputs and rejects mixed
// devices, so an op can never silently run on the wrong one.
func backendForAll(in ...*Tensor) (Backend, error) {
	if len(in) == 0 {
		return nil, fmt.Errorf("no inputs")
	}
	d := in[0].Device()
	for _, t := range in[1:] {
		if t.Device() != d {
			return nil, fmt.Errorf("device mismatch: %s and %s", d, t.Device())
		}
	}
	b, ok := backends[d]
	if !ok {
		return nil, fmt.Errorf("no backend registered for device %s", d)
	}
	return b, nil
}

func onesLike(t *Tensor) *Tensor {
	out := newLike(t)
	for i := range out.storage.data {
		out.storage.data[i] = 1
	}
	return out
}

func zerosLike(t *Tensor) *Tensor {
	return newLike(t)
}

// unbroadcast reduces g back to the shape of like. Only scalar broadcast exists
// today, so a 0d operand's gradient is the sum of all of g's elements.
func unbroadcast(be Backend, g, like *Tensor) (*Tensor, error) {
	if like.Numel() == g.Numel() {
		return g, nil
	}
	return sumAll(be, g)
}

func sumAll(be Backend, g *Tensor) (*Tensor, error) {
	out := g
	for out.Numel() > 1 {
		var err error
		out, err = be.Sum(out, 0, false)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func scalarLike(a *Tensor, v float32) *Tensor {
	out := newLike(a)
	for i := range out.storage.data {
		out.storage.data[i] = v
	}
	return out
}

func resolveDim(shape []int, dim int) int {
	// allow backwards indexing
	if dim < 0 {
		dim += len(shape)
	}
	return dim
}

// normalizeDims resolves negative dims and expands nil to ever dim. An empty,
// non-nil slice means reduce nothing.
func normalizeDims(shape []int, dims []int) ([]int, error) {
	if dims == nil {
		all := make([]int, len(shape))
		for i := range shape {
			all[i] = i
		}
		return all, nil
	}
	out := make([]int, len(dims))
	for i, d := range dims {
		rd := resolveDim(shape, d)
		if rd < 0 || rd >= len(shape) {
			return nil, fmt.Errorf("dim %d out of range for shape %v", d, shape)
		}
		out[i] = rd
	}
	return out, nil
}

// reduceSum runs the single dim kernel once per dim, highest index first so that
// dropping an axis never shifts a lower axies that has yet to be reduced
func reduceSum(be Backend, a *Tensor, dims []int, keepdim bool) (*Tensor, error) {
	ds, err := normalizeDims(a.Shape(), dims)
	if err != nil {
		return nil, err
	}
	if len(ds) == 0 {
		// summing over no dims is the identity, copy
		cp := newLike(a)
		copyStrided(cp.storage.data, a.materialize())
		return cp, nil
	}
	slices.Sort(ds)
	slices.Reverse(ds)
	out := a
	for _, d := range ds {
		out, err = be.Sum(out, d, keepdim)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// broadcastTo expands g to shape, using stride 0 on axes where g has size 1.
// Requires len(g.shape) == len(shape). materialize() turns it into a dense tensor
// for any backend op that needs contiguous data.
func broadcastTo(g *Tensor, shape []int) (*Tensor, error) {
	if len(g.shape) != len(shape) {
		return nil, fmt.Errorf("broadcastTo: rank %d vs %d", len(g.shape), len(shape))
	}
	out := &Tensor{
		storage: g.storage, // shares the backing array, only the strides change
		shape:   slices.Clone(shape),
		strides: make([]int, len(shape)),
		offset:  g.offset,
	}
	for i := range shape {
		switch g.shape[i] {
		case shape[i]:
			out.strides[i] = g.strides[i]
		case 1:
			out.strides[i] = 0
		default:
			return nil, fmt.Errorf("broadcastTo: cannot expand %v to %v", g.shape, shape)
		}
	}
	return out, nil
}

// expandToDim brings a reduced gradient back to a's shape. If keepdim was
// false, the reduced axis is first re-inserted as size 1 so the ranks match.
func expandToDims(be Backend, g, a *Tensor, dims []int, keepdim bool) (*Tensor, error) {
	shape := a.Shape()
	if !keepdim {
		ds, err := normalizeDims(shape, dims)
		if err != nil {
			return nil, err
		}
		newShape := slices.Clone(shape)
		for _, d := range ds {
			newShape[d] = 1
		}
		g, err = be.Reshape(g, newShape...)
		if err != nil {
			return nil, err
		}
	}
	b, err := broadcastTo(g, shape)
	if err != nil {
		return nil, err
	}
	return b.materialize(), nil
}
