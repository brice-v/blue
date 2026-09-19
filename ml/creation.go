package ml

import (
	"fmt"
	"math"
)

func Zeros(shape []int, dtype DType, device Device) (*Tensor, error) {
	return NewTensor(make([]float32, numelOf(shape)), shape, dtype, device)
}

func Ones(shape []int, dtype DType, device Device) (*Tensor, error) {
	data := make([]float32, numelOf(shape))
	for i := range data {
		data[i] = 1
	}
	return NewTensorOwned(data, shape, dtype, device)
}

func Full(shape []int, v float32, dtype DType, device Device) (*Tensor, error) {
	data := make([]float32, numelOf(shape))
	for i := range data {
		data[i] = v
	}
	return NewTensorOwned(data, shape, dtype, device)
}

func Arange(start, end, step float32, dtype DType, device Device) (*Tensor, error) {
	if step == 0 {
		return nil, fmt.Errorf("arange: step must be non-zero")
	}
	n := int(math.Ceil(float64((end - start) / step)))
	if n < 0 {
		n = 0
	}
	data := make([]float32, n)
	for i := range data {
		data[i] = start + float32(i)*step
	}
	return NewTensorOwned(data, []int{n}, dtype, device)
}

func Eye(n int, dtype DType, device Device) (*Tensor, error) {
	data := make([]float32, n*n)
	for i := range n {
		data[i*n+i] = 1
	}
	return NewTensorOwned(data, []int{n, n}, dtype, device)
}
