package ml

import (
	"fmt"
	"slices"

	"blue/borncgo/tensor"
)

// Binary arithmetic delegates straight to borncgo, which handles broadcasting
// (including scalar tensors of shape [1]) and records the op on its tape.

func Add(a, b *Tensor) (*Tensor, error) { return wrap(a.t.Add(b.t)), nil }
func Sub(a, b *Tensor) (*Tensor, error) { return wrap(a.t.Sub(b.t)), nil }
func Mul(a, b *Tensor) (*Tensor, error) { return wrap(a.t.Mul(b.t)), nil }
func Div(a, b *Tensor) (*Tensor, error) { return wrap(a.t.Div(b.t)), nil }

func MatMul(a, b *Tensor) (*Tensor, error) { return wrap(a.t.MatMul(b.t)), nil }

// Pow is the one op borncgo lacks, so blue composes it. A non-negative integer
// scalar exponent uses repeated multiplication, which stays differentiable and
// works for a negative base (matching PyTorch's integer-exponent behavior).
// Everything else is exp(b * log(a)).
func Pow(a, b *Tensor) (*Tensor, error) {
	if b.Numel() == 1 {
		e := b.t.Data()[0]
		if e == float32(int64(e)) && e >= 0 && e <= 64 {
			return powInt(a, int(e)), nil
		}
	}
	m := b.t.Mul(a.t.Log())
	return wrap(m.Exp()), nil
}

func powInt(a *Tensor, n int) *Tensor {
	out := tensor.Ones[float32](tensor.Shape(a.Shape()), engine)
	base := a.t
	for n > 0 {
		if n&1 == 1 {
			out = out.Mul(base)
		}
		n >>= 1
		if n > 0 {
			base = base.Mul(base)
		}
	}
	return wrap(out)
}

func Neg(a *Tensor) (*Tensor, error) { return wrap(a.t.MulScalar(float32(-1))), nil }

// Unary math and activations.

func Exp(a *Tensor) (*Tensor, error)  { return wrap(a.t.Exp()), nil }
func Log(a *Tensor) (*Tensor, error)  { return wrap(a.t.Log()), nil }
func Sqrt(a *Tensor) (*Tensor, error) { return wrap(a.t.Sqrt()), nil }
func Abs(a *Tensor) (*Tensor, error)  { return wrap(a.t.Abs()), nil }

func Relu(a *Tensor) (*Tensor, error)    { return wrapRaw(engine.ReLU(a.t.Raw())), nil }
func Sigmoid(a *Tensor) (*Tensor, error) { return wrapRaw(engine.Sigmoid(a.t.Raw())), nil }
func Tanh(a *Tensor) (*Tensor, error)    { return wrapRaw(engine.Tanh(a.t.Raw())), nil }

func Softmax(a *Tensor, dim int) (*Tensor, error) { return wrap(a.t.Softmax(dim)), nil }

// Comparisons return borncgo bool tensors, so blue reports dtype "bool" and
// to_list yields booleans, matching PyTorch.

func Eq(a, b *Tensor) (*Tensor, error) { return wrapRaw(engine.Equal(a.t.Raw(), b.t.Raw())), nil }
func Ne(a, b *Tensor) (*Tensor, error) {
	return wrapRaw(engine.NotEqual(a.t.Raw(), b.t.Raw())), nil
}
func Gt(a, b *Tensor) (*Tensor, error) {
	return wrapRaw(engine.Greater(a.t.Raw(), b.t.Raw())), nil
}
func Ge(a, b *Tensor) (*Tensor, error) {
	return wrapRaw(engine.GreaterEqual(a.t.Raw(), b.t.Raw())), nil
}
func Lt(a, b *Tensor) (*Tensor, error) { return wrapRaw(engine.Lower(a.t.Raw(), b.t.Raw())), nil }
func Le(a, b *Tensor) (*Tensor, error) {
	return wrapRaw(engine.LowerEqual(a.t.Raw(), b.t.Raw())), nil
}

// Reductions.

// Sum reduces dims, or every dim when dims is nil. keepdim keeps reduced dims as 1.
func Sum(a *Tensor, dims []int, keepdim bool) (*Tensor, error) {
	return reduce(a, dims, keepdim, func(x *Tensor, d int) *Tensor {
		return wrapRaw(engine.SumDim(x.t.Raw(), d, keepdim))
	})
}

// Mean is borncgo's MeanDim applied per dim.
func Mean(a *Tensor, dims []int, keepdim bool) (*Tensor, error) {
	return reduce(a, dims, keepdim, func(x *Tensor, d int) *Tensor {
		return wrapRaw(engine.MeanDim(x.t.Raw(), d, keepdim))
	})
}

func reduce(a *Tensor, dims []int, keepdim bool, f func(*Tensor, int) *Tensor) (*Tensor, error) {
	x := asFloat(a)
	ds, err := normalizeDims(x.Shape(), dims)
	if err != nil {
		return nil, err
	}
	if len(ds) == 0 {
		return x.Clone(), nil
	}
	out := x
	for _, d := range descendingDims(ds) {
		out = f(out, d)
	}
	return out, nil
}

// Max/Min are computed on the host because borncgo has no max reduction.
func Max(a *Tensor, dims []int, keepdim bool) (*Tensor, error) {
	return extremum(a, dims, keepdim, func(v, best float32) bool { return v > best })
}

func Min(a *Tensor, dims []int, keepdim bool) (*Tensor, error) {
	return extremum(a, dims, keepdim, func(v, best float32) bool { return v < best })
}

func extremum(a *Tensor, dims []int, keepdim bool, better func(v, best float32) bool) (*Tensor, error) {
	ds, err := normalizeDims(a.Shape(), dims)
	if err != nil {
		return nil, err
	}
	if len(ds) == 0 {
		return a.Clone(), nil
	}
	out := a
	for _, d := range descendingDims(ds) {
		out, err = extremumDim(out, d, keepdim, better)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func extremumDim(a *Tensor, dim int, keepdim bool, better func(v, best float32) bool) (*Tensor, error) {
	shape := a.Shape()
	d, err := normalizeDim(shape, dim)
	if err != nil {
		return nil, err
	}
	reduce := shape[d]
	outer, inner := 1, 1
	for _, s := range shape[:d] {
		outer *= s
	}
	for _, s := range shape[d+1:] {
		inner *= s
	}
	data := a.ContiguousData()
	out := make([]float32, outer*inner)
	for o := 0; o < outer; o++ {
		for in := 0; in < inner; in++ {
			best := data[o*reduce*inner+in]
			for r := 1; r < reduce; r++ {
				v := data[(o*reduce+r)*inner+in]
				if better(v, best) {
					best = v
				}
			}
			out[o*inner+in] = best
		}
	}
	ttShape := slices.Clone(shape)
	if keepdim {
		ttShape[d] = 1
	} else {
		ttShape = slices.Delete(ttShape, d, d+1)
	}
	return NewTensor(out, ttShape, a.dtype, a.Device())
}

// ArgMax/ArgMin return float32 tensors, matching blue's existing surface.
func ArgMax(a *Tensor, dim int) (*Tensor, error) {
	return wrap(a.t.Argmax(dim).Float32()), nil
}

func ArgMin(a *Tensor, dim int) (*Tensor, error) {
	n := a.t.MulScalar(float32(-1))
	return wrap(n.Argmax(dim).Float32()), nil
}

// Shape ops.

func Reshape(a *Tensor, shape ...int) (*Tensor, error) { return wrap(a.t.Reshape(shape...)), nil }

func Transpose(a *Tensor, dim0, dim1 int) (*Tensor, error) {
	n := len(a.Shape())
	d0, err := normalizeDim(a.Shape(), dim0)
	if err != nil {
		return nil, err
	}
	d1, err := normalizeDim(a.Shape(), dim1)
	if err != nil {
		return nil, err
	}
	perm := make([]int, n)
	for i := range perm {
		perm[i] = i
	}
	perm[d0], perm[d1] = perm[d1], perm[d0]
	return permuted(a, perm), nil
}

func Permute(a *Tensor, perm ...int) (*Tensor, error) {
	return permuted(a, perm), nil
}

// permuted transposes the data with borncgo and records the permuted strides so
// the reported view metadata matches PyTorch.
func permuted(a *Tensor, perm []int) *Tensor {
	out := wrap(a.t.Transpose(perm...))
	st := a.Strides()
	vs := make([]int, len(perm))
	for i, p := range perm {
		vs[i] = st[p]
	}
	out.viewStrides = vs
	return out
}

func Flatten(a *Tensor) (*Tensor, error) { return wrap(a.t.Reshape(a.Numel())), nil }

func Unsqueeze(a *Tensor, dim int) (*Tensor, error) { return wrap(a.t.Unsqueeze(dim)), nil }

// Squeeze drops size-1 dims. dims nil drops every size-1 dim.
func Squeeze(a *Tensor, dims []int) (*Tensor, error) {
	shape := a.Shape()
	var drop []int
	if dims == nil {
		for i, s := range shape {
			if s == 1 {
				drop = append(drop, i)
			}
		}
	} else {
		for _, d := range dims {
			rd, err := normalizeDim(shape, d)
			if err != nil {
				return nil, err
			}
			if shape[rd] != 1 {
				return nil, fmt.Errorf("squeeze: dim %d has size %d, not 1", d, shape[rd])
			}
			drop = append(drop, rd)
		}
	}
	out := a
	for _, d := range descendingDims(drop) {
		out = wrap(out.t.Squeeze(d))
	}
	return out, nil
}

func BroadcastTo(a *Tensor, shape []int) (*Tensor, error) {
	return wrap(a.t.Expand(tensor.Shape(shape))), nil
}

// Slice selects [start:end) along dim using borncgo's Gather.
func Slice(a *Tensor, dim, start, end int) (*Tensor, error) {
	shape := a.Shape()
	d, err := normalizeDim(shape, dim)
	if err != nil {
		return nil, err
	}
	n := shape[d]
	if start < 0 {
		start += n
	}
	if end < 0 {
		end += n
	}
	if start < 0 {
		start = 0
	}
	if end > n {
		end = n
	}
	if end < start {
		end = start
	}
	idx := make([]int32, end-start)
	for i := range idx {
		idx[i] = int32(start + i)
	}
	it, err := tensor.FromSlice[int32](idx, tensor.Shape{len(idx)}, engine)
	if err != nil {
		return nil, err
	}
	return wrapRaw(engine.Gather(a.t.Raw(), d, it.Raw())), nil
}

func Clamp(a *Tensor, lo, hi float32) (*Tensor, error) {
	return wrapRaw(engine.Clamp(a.t.Raw(), lo, hi)), nil
}

func Where(cond, a, b *Tensor) (*Tensor, error) {
	return wrapRaw(engine.Where(cond.t.Raw(), a.t.Raw(), b.t.Raw())), nil
}

// OneHot turns a 1d label tensor into [n, classes]. Labels are data, not
// differentiable.
func OneHot(labels *Tensor, classes int) (*Tensor, error) {
	data := labels.ContiguousData()
	n := len(data)
	out := make([]float32, n*classes)
	for i, v := range data {
		c := int(v)
		if c < 0 || c >= classes {
			return nil, fmt.Errorf("onehot: label %d out of range for %d classes", c, classes)
		}
		out[i*classes+c] = 1
	}
	return NewTensor(out, []int{n, classes}, Float32, labels.Device())
}
