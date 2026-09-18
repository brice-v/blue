package ml

import (
	"fmt"
	"math"
	"slices"
)

type CPUBackend struct{}

// for compile time checks
var _ Backend = CPUBackend{}

func init() {
	register(CPU, CPUBackend{})
}

// requireCPU returns an error unless every tensor is on the CPU device.
func requireCPU(op string, ts ...*Tensor) error {
	for _, t := range ts {
		if t.Device() != CPU {
			return fmt.Errorf("%s: expected cpu tensors, got %s", op, t.Device())
		}
	}
	return nil
}

func (cpu CPUBackend) MatMul(a, b *Tensor) (*Tensor, error) {
	if err := requireCPU("matmul", a, b); err != nil {
		return nil, err
	}
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
	if err := requireCPU("add", a, b); err != nil {
		return nil, err
	}
	a, b = a.materialize(), b.materialize()
	if a.Numel() == 1 && b.Numel() != 1 { // scalar on the left, broadcast
		tt := newLike(b)
		s := a.data[0]
		for i, v := range b.data {
			tt.data[i] = s + v
		}
		return tt, nil
	}
	if b.Numel() == 1 && a.Numel() != 1 { // scalar on the right, broadcast
		tt := newLike(a)
		s := b.data[0]
		for i, v := range a.data {
			tt.data[i] = v + s
		}
		return tt, nil
	}
	if a.Numel() != b.Numel() {
		return nil, fmt.Errorf("add: shape mismatch: %v %v", a.Numel(), b.Numel())
	}
	tt := newLike(a)
	for i, v := range a.data {
		tt.data[i] = v + b.data[i]
	}
	return tt, nil
}

func (cpu CPUBackend) Sub(a, b *Tensor) (*Tensor, error) {
	if err := requireCPU("sub", a, b); err != nil {
		return nil, err
	}
	a, b = a.materialize(), b.materialize()
	if a.Numel() == 1 && b.Numel() != 1 { // scalar on the left, broadcast
		tt := newLike(b)
		s := a.data[0]
		for i, v := range b.data {
			tt.data[i] = s - v
		}
		return tt, nil
	}
	if b.Numel() == 1 && a.Numel() != 1 { // scalar on the right, broadcast
		tt := newLike(a)
		s := b.data[0]
		for i, v := range a.data {
			tt.data[i] = v - s
		}
		return tt, nil
	}
	if a.Numel() != b.Numel() {
		return nil, fmt.Errorf("sub: shape mismatch: %v %v", a.Numel(), b.Numel())
	}
	tt := newLike(a)
	for i, v := range a.data {
		tt.data[i] = v - b.data[i]
	}
	return tt, nil
}

func (cpu CPUBackend) Mul(a, b *Tensor) (*Tensor, error) {
	if err := requireCPU("mul", a, b); err != nil {
		return nil, err
	}
	a, b = a.materialize(), b.materialize()
	if a.Numel() == 1 && b.Numel() != 1 { // scalar on the left, broadcast
		tt := newLike(b)
		s := a.data[0]
		for i, v := range b.data {
			tt.data[i] = s * v
		}
		return tt, nil
	}
	if b.Numel() == 1 && a.Numel() != 1 { // scalar on the right, broadcast
		tt := newLike(a)
		s := b.data[0]
		for i, v := range a.data {
			tt.data[i] = v * s
		}
		return tt, nil
	}
	if a.Numel() != b.Numel() {
		return nil, fmt.Errorf("mul: shape mismatch: %v %v", a.Numel(), b.Numel())
	}
	tt := newLike(a)
	for i, v := range a.data {
		tt.data[i] = v * b.data[i]
	}
	return tt, nil
}

func (cpu CPUBackend) Div(a, b *Tensor) (*Tensor, error) {
	if err := requireCPU("div", a, b); err != nil {
		return nil, err
	}
	a, b = a.materialize(), b.materialize()
	if a.Numel() == 1 && b.Numel() != 1 { // scalar on the left, broadcast
		tt := newLike(b)
		s := a.data[0]
		for i, v := range b.data {
			if v == 0 {
				return nil, fmt.Errorf("div: divide by 0")
			}
			tt.data[i] = s / v
		}
		return tt, nil
	}
	if b.Numel() == 1 && a.Numel() != 1 { // scalar on the right, broadcast
		tt := newLike(a)
		s := b.data[0]
		if s == 0 {
			return nil, fmt.Errorf("div: divide by 0")
		}
		for i, v := range a.data {
			tt.data[i] = v / s
		}
		return tt, nil
	}
	if a.Numel() != b.Numel() {
		return nil, fmt.Errorf("div: shape mismatch: %v %v", a.Numel(), b.Numel())
	}
	tt := newLike(a)
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
	if err := requireCPU("exp", a); err != nil {
		return nil, err
	}
	return mapUnary(a, func(v float32) float32 {
		return float32(math.Exp(float64(v)))
	}), nil
}

func (cpu CPUBackend) Log(a *Tensor) (*Tensor, error) {
	if err := requireCPU("log", a); err != nil {
		return nil, err
	}
	return mapUnary(a, func(v float32) float32 {
		return float32(math.Log(float64(v)))
	}), nil
}

func (cpu CPUBackend) Sqrt(a *Tensor) (*Tensor, error) {
	if err := requireCPU("sqrt", a); err != nil {
		return nil, err
	}
	return mapUnary(a, func(v float32) float32 {
		return float32(math.Sqrt(float64(v)))
	}), nil
}

// outer, reduce, inner, dim, error
// materializes tensor first then returns if valid
func oride(a *Tensor, dimension int, opName string, canReduceBeZero bool) (outer, reduce, inner, dim int, err error) {
	if err = requireCPU(opName, a); err != nil {
		return
	}
	a = a.materialize()
	if len(a.shape) == 0 {
		err = fmt.Errorf("%s: cannot reduce a 0d tensor", opName)
		return
	}
	dimension = resolveDim(a.shape, dimension)
	if dimension < 0 || dimension >= len(a.shape) {
		err = fmt.Errorf("%s: dim %d out of range for shape %v", opName, dimension, a.shape)
		return
	}
	dim = dimension

	outer, reduce, inner = 1, a.shape[dim], 1
	if !canReduceBeZero && reduce == 0 {
		err = fmt.Errorf("%s: cannot reduce an empty dimension", opName)
		return
	}
	for _, d := range a.shape[:dim] {
		outer *= d
	}
	for _, d := range a.shape[dim+1:] {
		inner *= d
	}
	return
}

func (cpu CPUBackend) Sum(a *Tensor, dim int, keepdim bool) (*Tensor, error) {
	outer, reduce, inner, dim, err := oride(a, dim, "sum", true)
	if err != nil {
		return nil, err
	}

	ttShape := slices.Clone(a.shape)
	if keepdim {
		ttShape[dim] = 1
	} else {
		ttShape = slices.Delete(ttShape, dim, dim+1)
	}
	tt := &Tensor{
		data:    make([]float32, outer*inner),
		shape:   ttShape,
		strides: getContiguousStridesFromShape(ttShape),
		dtype:   a.dtype,
		device:  a.device,
	}
	for o := range outer {
		for in := range inner {
			var s float32
			for r := range reduce {
				s += a.data[(o*reduce+r)*inner+in]
			}
			tt.data[o*inner+in] = s
		}
	}
	return tt, nil
}

func (cpu CPUBackend) Max(a *Tensor, dim int, keepdim bool) (*Tensor, error) {
	outer, reduce, inner, dim, err := oride(a, dim, "max", false)
	if err != nil {
		return nil, err
	}

	ttShape := slices.Clone(a.shape)
	if keepdim {
		ttShape[dim] = 1
	} else {
		ttShape = slices.Delete(ttShape, dim, dim+1)
	}
	tt := &Tensor{
		data:    make([]float32, outer*inner),
		shape:   ttShape,
		strides: getContiguousStridesFromShape(ttShape),
		dtype:   a.dtype,
		device:  a.device,
	}
	for o := range outer {
		for in := range inner {
			m := a.data[(o*reduce)*inner+in]
			for r := 1; r < reduce; r++ {
				if v := a.data[(o*reduce+r)*inner+in]; v > m {
					m = v
				}
			}
			tt.data[o*inner+in] = m
		}
	}
	return tt, nil
}

func (cpu CPUBackend) Softmax(a *Tensor, dim int) (*Tensor, error) {
	outer, reduce, inner, dim, err := oride(a, dim, "softmax", false)
	if err != nil {
		return nil, err
	}

	tt := newLike(a)
	for o := range outer {
		for in := range inner {
			base := o*reduce*inner + in
			m := a.data[base]
			for r := 1; r < reduce; r++ {
				if v := a.data[base+r*inner]; v > m {
					m = v
				}
			}
			var sum float32
			for r := range reduce {
				e := float32(math.Exp(float64(a.data[base+r*inner] - m)))
				tt.data[base+r*inner] = e
				sum += e
			}
			for r := range reduce {
				tt.data[base+r*inner] /= sum
			}
		}
	}
	return tt, nil
}

func (cpu CPUBackend) Neg(a *Tensor) (*Tensor, error) {
	if err := requireCPU("neg", a); err != nil {
		return nil, err
	}
	return mapUnary(a, func(v float32) float32 { return -v }), nil
}

func (cpu CPUBackend) Relu(a *Tensor) (*Tensor, error) {
	if err := requireCPU("relu", a); err != nil {
		return nil, err
	}
	return mapUnary(a, func(v float32) float32 {
		if v < 0 {
			return 0
		}
		return v
	}), nil
}

func (cpu CPUBackend) Greater(a, b *Tensor) (*Tensor, error) {
	if err := requireCPU("greater", a, b); err != nil {
		return nil, err
	}
	a, b = a.materialize(), b.materialize()
	if a.Numel() == 1 && b.Numel() != 1 { // scalar on the left, broadcast
		tt := newLike(b)
		tt.dtype = Bool
		s := a.data[0]
		for i, v := range b.data {
			if s > v {
				tt.data[i] = 1
			}
		}
		return tt, nil
	}
	if b.Numel() == 1 && a.Numel() != 1 { // scalar on the right, broadcast
		tt := newLike(a)
		tt.dtype = Bool
		s := b.data[0]
		for i, v := range a.data {
			if v > s {
				tt.data[i] = 1
			}
		}
		return tt, nil
	}
	if a.Numel() != b.Numel() {
		return nil, fmt.Errorf("greater: shape mismatch: %v %v", a.Numel(), b.Numel())
	}
	tt := newLike(a)
	tt.dtype = Bool // 0.0 / 1.0 in the same float32 buffer
	for i, v := range a.data {
		if v > b.data[i] {
			tt.data[i] = 1
		}
	}
	return tt, nil
}

func (cpu CPUBackend) Eq(a, b *Tensor) (*Tensor, error) {
	if err := requireCPU("eq", a, b); err != nil {
		return nil, err
	}
	a, b = a.materialize(), b.materialize()
	if a.Numel() == 1 && b.Numel() != 1 { // scalar on the left, broadcast
		tt := newLike(b)
		tt.dtype = Bool
		s := a.data[0]
		for i, v := range b.data {
			if s == v {
				tt.data[i] = 1
			}
		}
		return tt, nil
	}
	if b.Numel() == 1 && a.Numel() != 1 { // scalar on the right, broadcast
		tt := newLike(a)
		tt.dtype = Bool
		s := b.data[0]
		for i, v := range a.data {
			if v == s {
				tt.data[i] = 1
			}
		}
		return tt, nil
	}
	if a.Numel() != b.Numel() {
		return nil, fmt.Errorf("eq: shape mismatch: %v %v", a.Numel(), b.Numel())
	}
	tt := newLike(a)
	tt.dtype = Bool // 0.0 / 1.0 in the same float32 buffer
	for i, v := range a.data {
		if v == b.data[i] {
			tt.data[i] = 1
		}
	}
	return tt, nil
}

// view ops (does not materialize)

func (cpu CPUBackend) Reshape(a *Tensor, shape ...int) (*Tensor, error) {
	if err := requireCPU("reshape", a); err != nil {
		return nil, err
	}
	if numelOf(shape) != a.Numel() {
		return nil, fmt.Errorf("reshape: cannot reshape %v into %v", a.shape, shape)
	}
	src := a
	if !src.IsContiguous() {
		src = src.materialize()
	}
	return &Tensor{
		data:    src.data, // same backing array
		shape:   slices.Clone(shape),
		strides: getContiguousStridesFromShape(shape),
		offset:  src.offset,
		dtype:   src.dtype,
		device:  src.device,
	}, nil
}

func (cpu CPUBackend) Transpose(a *Tensor, dim0, dim1 int) (*Tensor, error) {
	if err := requireCPU("transpose", a); err != nil {
		return nil, err
	}
	n := len(a.shape)
	if dim0 < 0 {
		dim0 += n
	}
	if dim1 < 0 {
		dim1 += n
	}
	if dim0 < 0 || dim0 >= n || dim1 < 0 || dim1 >= n {
		return nil, fmt.Errorf("transpose: dims %d,%d out of range for shape %v", dim0, dim1, a.shape)
	}
	tt := &Tensor{
		data:    a.data, // same backing array
		shape:   slices.Clone(a.shape),
		strides: slices.Clone(a.strides),
		offset:  a.offset,
		dtype:   a.dtype,
		device:  a.device,
	}
	tt.shape[dim0], tt.shape[dim1] = tt.shape[dim1], tt.shape[dim0]
	tt.strides[dim0], tt.strides[dim1] = tt.strides[dim1], tt.strides[dim0]
	return tt, nil
}
