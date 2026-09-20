package ml

import (
	"fmt"

	"blue/borncgo/tensor"
)

func tag(t *bornTensor, dtype DType, device Device) *Tensor {
	tt := wrap(t)
	tt.dtype = dtype
	tt.device = device
	return tt
}

func Zeros(shape []int, dtype DType, device Device) (*Tensor, error) {
	be, err := backendFor(device)
	if err != nil {
		return nil, err
	}
	return tag(tensor.Zeros[float32](tensor.Shape(shape), be), dtype, device), nil
}

func Ones(shape []int, dtype DType, device Device) (*Tensor, error) {
	be, err := backendFor(device)
	if err != nil {
		return nil, err
	}
	return tag(tensor.Ones[float32](tensor.Shape(shape), be), dtype, device), nil
}

func Full(shape []int, v float32, dtype DType, device Device) (*Tensor, error) {
	be, err := backendFor(device)
	if err != nil {
		return nil, err
	}
	return tag(tensor.Full[float32](tensor.Shape(shape), v, be), dtype, device), nil
}

func Eye(n int, dtype DType, device Device) (*Tensor, error) {
	be, err := backendFor(device)
	if err != nil {
		return nil, err
	}
	return tag(tensor.Eye[float32](n, be), dtype, device), nil
}

func Randn(shape []int, dtype DType, device Device) (*Tensor, error) {
	be, err := backendFor(device)
	if err != nil {
		return nil, err
	}
	return tag(tensor.Randn[float32](tensor.Shape(shape), be), dtype, device), nil
}

// buildRaw allocates a raw tensor of the given dtype and device, fills it, and
// wraps it. The wrapper's Go type parameter stays float32, but the raw buffer
// and dtype tag are the requested type; every op dispatches on the raw dtype.
func buildRaw(shape []int, dtype tensor.DataType, be tensor.Backend, fill func(*tensor.RawTensor)) (*Tensor, error) {
	raw, err := tensor.NewRaw(tensor.Shape(shape), dtype, be.Device())
	if err != nil {
		return nil, err
	}
	fill(raw)
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
