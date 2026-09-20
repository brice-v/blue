package ml

import (
	"fmt"
	"sort"

	"blue/borncgo/tensor"
)

func tag(t *bornTensor, dtype DType, device Device) *Tensor {
	tt := wrap(t)
	tt.dtype = dtype
	tt.device = device
	return tt
}

// isFloat32 reports whether a dtype can use the float32 fast path. Invalid is
// treated as float32 because that is what callers that omit the dtype get.
func isFloat32(dtype DType) bool {
	return dtype == Float32 || dtype == Invalid
}

// fillValue writes a single value into a raw tensor of any supported dtype, so
// creation honors the requested dtype instead of always storing float32.
func fillValue(raw *tensor.RawTensor, v float32) {
	switch raw.DType() {
	case tensor.Float32:
		s := raw.AsFloat32()
		for i := range s {
			s[i] = v
		}
	case tensor.Float64:
		s := raw.AsFloat64()
		for i := range s {
			s[i] = float64(v)
		}
	case tensor.Int32:
		s := raw.AsInt32()
		for i := range s {
			s[i] = int32(v)
		}
	case tensor.Int64:
		s := raw.AsInt64()
		for i := range s {
			s[i] = int64(v)
		}
	case tensor.Uint8:
		s := raw.AsUint8()
		for i := range s {
			s[i] = uint8(v)
		}
	case tensor.Bool:
		s := raw.AsBool()
		for i := range s {
			s[i] = v != 0
		}
	}
}

func Zeros(shape []int, dtype DType, device Device) (*Tensor, error) {
	be, err := backendFor(device)
	if err != nil {
		return nil, err
	}
	if isFloat32(dtype) {
		return tag(tensor.Zeros[float32](tensor.Shape(shape), be), Float32, device), nil
	}
	// NewRaw zeroes the buffer, so no fill is needed.
	return buildRaw(shape, tensor.DataType(dtype), be, nil)
}

func Ones(shape []int, dtype DType, device Device) (*Tensor, error) {
	be, err := backendFor(device)
	if err != nil {
		return nil, err
	}
	if isFloat32(dtype) {
		return tag(tensor.Ones[float32](tensor.Shape(shape), be), Float32, device), nil
	}
	return buildRaw(shape, tensor.DataType(dtype), be, func(raw *tensor.RawTensor) { fillValue(raw, 1) })
}

func Full(shape []int, v float32, dtype DType, device Device) (*Tensor, error) {
	be, err := backendFor(device)
	if err != nil {
		return nil, err
	}
	if isFloat32(dtype) {
		return tag(tensor.Full[float32](tensor.Shape(shape), v, be), Float32, device), nil
	}
	return buildRaw(shape, tensor.DataType(dtype), be, func(raw *tensor.RawTensor) { fillValue(raw, v) })
}

func Eye(n int, dtype DType, device Device) (*Tensor, error) {
	be, err := backendFor(device)
	if err != nil {
		return nil, err
	}
	base := tensor.Eye[float32](n, be)
	if isFloat32(dtype) {
		return tag(base, Float32, device), nil
	}
	return wrapRaw(be, be.Cast(base.Raw(), tensor.DataType(dtype))), nil
}

func Randn(shape []int, dtype DType, device Device) (*Tensor, error) {
	be, err := backendFor(device)
	if err != nil {
		return nil, err
	}
	base := tensor.Randn[float32](tensor.Shape(shape), be)
	if isFloat32(dtype) {
		return tag(base, Float32, device), nil
	}
	return wrapRaw(be, be.Cast(base.Raw(), tensor.DataType(dtype))), nil
}

// Rand returns uniform samples in [0, 1) from the shared engine RNG.
func Rand(shape []int, dtype DType, device Device) (*Tensor, error) {
	be, err := backendFor(device)
	if err != nil {
		return nil, err
	}
	base := tensor.Rand[float32](tensor.Shape(shape), be)
	if isFloat32(dtype) {
		return tag(base, Float32, device), nil
	}
	return wrapRaw(be, be.Cast(base.Raw(), tensor.DataType(dtype))), nil
}

// RandPerm returns a random permutation of 0..n-1 drawn from the shared engine
// RNG, so it is reproducible after ManualSeed. Values are float32 like every
// other blue tensor; index consumers cast to int32.
func RandPerm(n int, device Device) (*Tensor, error) {
	if n < 0 {
		return nil, fmt.Errorf("randperm: n must be non-negative, got %d", n)
	}
	be, err := backendFor(device)
	if err != nil {
		return nil, err
	}
	u := tensor.Rand[float32](tensor.Shape{n}, be)
	keys := u.Data()
	perm := make([]int, n)
	for i := range perm {
		perm[i] = i
	}
	sort.SliceStable(perm, func(i, j int) bool { return keys[perm[i]] < keys[perm[j]] })
	data := make([]float32, n)
	for i, p := range perm {
		data[i] = float32(p)
	}
	return NewTensor(data, []int{n}, Float32, device)
}

// Shuffle returns a copy of a with dim permuted by a random permutation, so
// rows can be shuffled for minibatch training.
func Shuffle(a *Tensor, dim int) (*Tensor, error) {
	shape := a.Shape()
	d, err := normalizeDim(shape, dim)
	if err != nil {
		return nil, err
	}
	perm, err := RandPerm(shape[d], a.device)
	if err != nil {
		return nil, err
	}
	data := perm.ContiguousData()
	indices := make([]int, len(data))
	for i, v := range data {
		indices[i] = int(v)
	}
	return gatherAlong(a, d, indices)
}

// buildRaw allocates a raw tensor of the given dtype and device, fills it, and
// wraps it. The wrapper's Go type parameter stays float32, but the raw buffer
// and dtype tag are the requested type; every op dispatches on the raw dtype.
func buildRaw(shape []int, dtype tensor.DataType, be tensor.Backend, fill func(*tensor.RawTensor)) (*Tensor, error) {
	raw, err := tensor.NewRaw(tensor.Shape(shape), dtype, be.Device())
	if err != nil {
		return nil, err
	}
	if fill != nil {
		fill(raw)
	}
	return wrapRaw(be, raw), nil
}

// NewFloat64Tensor builds a float64-backed tensor.
func NewFloat64Tensor(data []float64, shape []int, device Device) (*Tensor, error) {
	be, err := backendFor(device)
	if err != nil {
		return nil, err
	}
	if len(data) != tensor.Shape(shape).NumElements() {
		return nil, fmt.Errorf("NewFloat64Tensor: data has %d elements, shape %v needs %d", len(data), shape, tensor.Shape(shape).NumElements())
	}
	return buildRaw(shape, tensor.Float64, be, func(raw *tensor.RawTensor) { copy(raw.AsFloat64(), data) })
}

// NewInt32Tensor builds an int32-backed tensor.
func NewInt32Tensor(data []int32, shape []int, device Device) (*Tensor, error) {
	be, err := backendFor(device)
	if err != nil {
		return nil, err
	}
	if len(data) != tensor.Shape(shape).NumElements() {
		return nil, fmt.Errorf("NewInt32Tensor: data has %d elements, shape %v needs %d", len(data), shape, tensor.Shape(shape).NumElements())
	}
	return buildRaw(shape, tensor.Int32, be, func(raw *tensor.RawTensor) { copy(raw.AsInt32(), data) })
}

// NewInt64Tensor builds an int64-backed tensor.
func NewInt64Tensor(data []int64, shape []int, device Device) (*Tensor, error) {
	be, err := backendFor(device)
	if err != nil {
		return nil, err
	}
	if len(data) != tensor.Shape(shape).NumElements() {
		return nil, fmt.Errorf("NewInt64Tensor: data has %d elements, shape %v needs %d", len(data), shape, tensor.Shape(shape).NumElements())
	}
	return buildRaw(shape, tensor.Int64, be, func(raw *tensor.RawTensor) { copy(raw.AsInt64(), data) })
}

// NewUint8Tensor builds a uint8-backed tensor.
func NewUint8Tensor(data []uint8, shape []int, device Device) (*Tensor, error) {
	be, err := backendFor(device)
	if err != nil {
		return nil, err
	}
	if len(data) != tensor.Shape(shape).NumElements() {
		return nil, fmt.Errorf("NewUint8Tensor: data has %d elements, shape %v needs %d", len(data), shape, tensor.Shape(shape).NumElements())
	}
	return buildRaw(shape, tensor.Uint8, be, func(raw *tensor.RawTensor) { copy(raw.AsUint8(), data) })
}

// NewBoolTensor builds a bool-backed tensor.
func NewBoolTensor(data []bool, shape []int, device Device) (*Tensor, error) {
	be, err := backendFor(device)
	if err != nil {
		return nil, err
	}
	if len(data) != tensor.Shape(shape).NumElements() {
		return nil, fmt.Errorf("NewBoolTensor: data has %d elements, shape %v needs %d", len(data), shape, tensor.Shape(shape).NumElements())
	}
	return buildRaw(shape, tensor.Bool, be, func(raw *tensor.RawTensor) { copy(raw.AsBool(), data) })
}

// Arange returns evenly spaced values in [start, end). borncgo's Arange has no
// step, so blue builds the data itself.
func Arange(start, end, step float32, dtype DType, device Device) (*Tensor, error) {
	if step == 0 {
		return nil, fmt.Errorf("arange: step must be nonzero")
	}
	n := 0
	for v := start; (step > 0 && v < end) || (step < 0 && v > end); v += step {
		n++
	}
	data := make([]float32, n)
	v := start
	for i := range data {
		data[i] = v
		v += step
	}
	return NewTensor(data, []int{n}, dtype, device)
}

// ManualSeed seeds borncgo's shared RNG so randn is reproducible.
func ManualSeed(seed int64) {
	tensor.SetSeed(seed)
}
