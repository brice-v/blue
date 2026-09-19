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
	return tag(tensor.Zeros[float32](tensor.Shape(shape), engine), dtype, device), nil
}

func Ones(shape []int, dtype DType, device Device) (*Tensor, error) {
	return tag(tensor.Ones[float32](tensor.Shape(shape), engine), dtype, device), nil
}

func Full(shape []int, v float32, dtype DType, device Device) (*Tensor, error) {
	return tag(tensor.Full[float32](tensor.Shape(shape), v, engine), dtype, device), nil
}

func Eye(n int, dtype DType, device Device) (*Tensor, error) {
	return tag(tensor.Eye[float32](n, engine), dtype, device), nil
}

func Randn(shape []int, dtype DType, device Device) (*Tensor, error) {
	return tag(tensor.Randn[float32](tensor.Shape(shape), engine), dtype, device), nil
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
