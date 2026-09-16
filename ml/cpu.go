package ml

import (
	"fmt"
	"math"
)

type CPUBackend struct{}

var DefaultBackend Backend = CPUBackend{}

func (cpu CPUBackend) MatMul(a, b *Tensor) (*Tensor, error) {
	a, b = a.materialize(), b.materialize()
	if len(a.shape) != 2 || len(b.shape) != 2 {
		return nil, fmt.Errorf("matmul: expected 2d tensors got %v and %v", a.shape, b.shape)
	}
	m, k1 := a.shape[0], a.shape[1]
	k2, n := b.shape[0], b.shape[1]
	if k1 != k2 {
		return nil, fmt.Errorf("matmul: inner dimensions differ: %v x %v", a.shape, b.shape)
	}
	tt := &Tensor{
		data:    make([]float32, m*n),
		shape:   []int{m, n},
		strides: []int{n, 1},
		dtype:   a.dtype,
		device:  a.device,
	}
	for i := range m {
		for p := range k1 {
			av := a.data[i*k1+p] // packed after materialize
			for j := range n {
				tt.data[i*n+j] += av * b.data[p*n+j]
			}
		}
	}
	return tt, nil
}

func (cpu CPUBackend) Add(a, b *Tensor) (*Tensor, error) {
	a = a.materialize()
	tt := newLike(a)
	if b.Numel() == 1 {
		s := b.materialize().data[0]
		for i, v := range a.data {
			tt.data[i] = v + s
		}
		return tt, nil
	}
	b = b.materialize()
	if a.Numel() != b.Numel() {
		return nil, fmt.Errorf("add: shape mismatch: %v %v", a.Numel(), b.Numel())
	}
	for i, v := range a.data {
		tt.data[i] = v + b.data[i]
	}
	return tt, nil
}

func (cpu CPUBackend) Sub(a, b *Tensor) (*Tensor, error) {
	a = a.materialize()
	tt := newLike(a)
	if b.Numel() == 1 {
		s := b.materialize().data[0]
		for i, v := range a.data {
			tt.data[i] = v - s
		}
		return tt, nil
	}
	b = b.materialize()
	if a.Numel() != b.Numel() {
		return nil, fmt.Errorf("sub: shape mismatch: %v %v", a.Numel(), b.Numel())
	}
	for i, v := range a.data {
		tt.data[i] = v - b.data[i]
	}
	return tt, nil
}

func (cpu CPUBackend) Mul(a, b *Tensor) (*Tensor, error) {
	a = a.materialize()
	tt := newLike(a)
	if b.Numel() == 1 {
		s := b.materialize().data[0]
		for i, v := range a.data {
			tt.data[i] = v * s
		}
		return tt, nil
	}
	b = b.materialize()
	if a.Numel() != b.Numel() {
		return nil, fmt.Errorf("mul: shape mismatch: %v %v", a.Numel(), b.Numel())
	}
	for i, v := range a.data {
		tt.data[i] = v * b.data[i]
	}
	return tt, nil
}

func (cpu CPUBackend) Div(a, b *Tensor) (*Tensor, error) {
	a = a.materialize()
	tt := newLike(a)
	if b.Numel() == 1 {
		s := b.materialize().data[0]
		if s == 0 {
			return nil, fmt.Errorf("div: divide by 0")
		}
		for i, v := range a.data {
			tt.data[i] = v / s
		}
		return tt, nil
	}
	b = b.materialize()
	if a.Numel() != b.Numel() {
		return nil, fmt.Errorf("div: shape mismatch: %v %v", a.Numel(), b.Numel())
	}
	for i, v := range a.data {
		bb := b.data[i]
		if bb == 0 {
			return nil, fmt.Errorf("div: divide by 0")
		}
		tt.data[i] = v / b.data[i]
	}
	return tt, nil
}

func mapUnary(a *Tensor, f func(float32) float32) *Tensor {
	a = a.materialize()
	tt := newLike(a)
	for i := range tt.data {
		tt.data[i] = f(a.data[i])
	}
	return tt
}

func (cpu CPUBackend) Exp(a *Tensor) (*Tensor, error) {
	return mapUnary(a, func(v float32) float32 {
		return float32(math.Exp(float64(v)))
	}), nil
}

func (cpu CPUBackend) Log(a *Tensor) (*Tensor, error) {
	return mapUnary(a, func(v float32) float32 {
		return float32(math.Log(float64(v)))
	}), nil
}

func (cpu CPUBackend) Sqrt(a *Tensor) (*Tensor, error) {
	return mapUnary(a, func(v float32) float32 {
		return float32(math.Sqrt(float64(v)))
	}), nil
}

func (cpu CPUBackend) Sum(a *Tensor, dim int, keepdim bool) (*Tensor, error) {
	return nil, nil
}

func (cpu CPUBackend) Max(a *Tensor, dim int, keepdim bool) (*Tensor, error) {
	return nil, nil
}

func (cpu CPUBackend) Softmax(a *Tensor, dim int) (*Tensor, error) {
	return nil, nil
}

// view ops (does not materialize)

func (cpu CPUBackend) Reshape(a *Tensor, shape ...int) (*Tensor, error) {
	return nil, nil
}

func (cpu CPUBackend) Transpose(a *Tensor, dim0, dim1 int) (*Tensor, error) {
	return nil, nil
}
