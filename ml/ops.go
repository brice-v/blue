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
	// A uniform non-negative integer exponent (scalar or a whole tensor of the
	// same value) uses repeated multiplication: correct for a negative base and
	// differentiable, and it never calls log on a non-positive value.
	if n, ok := uniformIntExponent(b); ok && n <= 64 {
		return powInt(a, n), nil
	}
	// General case needs log(a). borncgo's Log panics on non-positive input, so
	// reject it rather than crash.
	if minFloat(a.ContiguousData()) <= 0 {
		return nil, fmt.Errorf("pow: a non-integer exponent with a non-positive base is unsupported")
	}
	m := b.t.Mul(a.t.Log())
	return wrap(m.Exp()), nil
}

// uniformIntExponent reports whether every element of b is the same
// non-negative integer.
func uniformIntExponent(b *Tensor) (int, bool) {
	data := b.ContiguousData()
	if len(data) == 0 {
		return 0, false
	}
	v := data[0]
	if v < 0 || v != float32(int64(v)) {
		return 0, false
	}
	for _, x := range data {
		if x != v {
			return 0, false
		}
	}
	return int(v), true
}

func minFloat(data []float32) float32 {
	if len(data) == 0 {
		return 0
	}
	m := data[0]
	for _, v := range data {
		if v < m {
			m = v
		}
	}
	return m
}

func powInt(a *Tensor, n int) *Tensor {
	out := tensor.Ones[float32](tensor.Shape(a.Shape()), a.be)
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

// In-place ops compute the result and copy it into a's buffer, keeping the
// tensor's identity (PyTorch in-place semantics). They are not tape-aware.
func AddInPlace(a, b *Tensor) error { return copyInto(a, Add, b) }
func SubInPlace(a, b *Tensor) error { return copyInto(a, Sub, b) }
func MulInPlace(a, b *Tensor) error { return copyInto(a, Mul, b) }
func DivInPlace(a, b *Tensor) error { return copyInto(a, Div, b) }

func copyInto(a *Tensor, f func(a, b *Tensor) (*Tensor, error), b *Tensor) error {
	out, err := f(a, b)
	if err != nil {
		return err
	}
	dst := a.t.Raw().AsFloat32()
	src := out.ContiguousData()
	if len(src) != len(dst) {
		return fmt.Errorf("in-place: size mismatch %d vs %d", len(src), len(dst))
	}
	copy(dst, src)
	return nil
}

// Unary math and activations.

func Exp(a *Tensor) (*Tensor, error)  { return wrap(a.t.Exp()), nil }
func Log(a *Tensor) (*Tensor, error)  { return wrap(a.t.Log()), nil }
func Sqrt(a *Tensor) (*Tensor, error) { return wrap(a.t.Sqrt()), nil }
func Abs(a *Tensor) (*Tensor, error)  { return wrap(a.t.Abs()), nil }

func Relu(a *Tensor) (*Tensor, error)    { return wrapRaw(a.be.ReLU(a.t.Raw())), nil }
func Sigmoid(a *Tensor) (*Tensor, error) { return wrapRaw(a.be.Sigmoid(a.t.Raw())), nil }
func Tanh(a *Tensor) (*Tensor, error)    { return wrapRaw(a.be.Tanh(a.t.Raw())), nil }

func Softmax(a *Tensor, dim int) (*Tensor, error) { return wrap(a.t.Softmax(dim)), nil }

// Comparisons return borncgo bool tensors, so blue reports dtype "bool" and
// to_list yields booleans, matching PyTorch.

func Eq(a, b *Tensor) (*Tensor, error) { return wrapRaw(a.be.Equal(a.t.Raw(), b.t.Raw())), nil }
func Ne(a, b *Tensor) (*Tensor, error) {
	return wrapRaw(a.be.NotEqual(a.t.Raw(), b.t.Raw())), nil
}
func Gt(a, b *Tensor) (*Tensor, error) {
	return wrapRaw(a.be.Greater(a.t.Raw(), b.t.Raw())), nil
}
func Ge(a, b *Tensor) (*Tensor, error) {
	return wrapRaw(a.be.GreaterEqual(a.t.Raw(), b.t.Raw())), nil
}
func Lt(a, b *Tensor) (*Tensor, error) { return wrapRaw(a.be.Lower(a.t.Raw(), b.t.Raw())), nil }
func Le(a, b *Tensor) (*Tensor, error) {
	return wrapRaw(a.be.LowerEqual(a.t.Raw(), b.t.Raw())), nil
}

// Reductions.

// Sum reduces dims, or every dim when dims is nil. keepdim keeps reduced dims as 1.
func Sum(a *Tensor, dims []int, keepdim bool) (*Tensor, error) {
	return reduce(a, dims, keepdim, func(x *Tensor, d int) *Tensor {
		return wrapRaw(a.be.SumDim(x.t.Raw(), d, keepdim))
	})
}

// Mean is borncgo's MeanDim applied per dim.
func Mean(a *Tensor, dims []int, keepdim bool) (*Tensor, error) {
	return reduce(a, dims, keepdim, func(x *Tensor, d int) *Tensor {
		return wrapRaw(a.be.MeanDim(x.t.Raw(), d, keepdim))
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
	// borncgo's Gather is PyTorch-style: the index must have the same rank as
	// the input and its shape is the output shape. Build an index that walks
	// start..end along dim.
	outShape := slices.Clone(shape)
	outShape[d] = end - start
	total := 1
	for _, s := range outShape {
		total *= s
	}
	strides := make([]int, len(outShape))
	acc := 1
	for i := len(outShape) - 1; i >= 0; i-- {
		strides[i] = acc
		acc *= outShape[i]
	}
	idx := make([]int32, total)
	for flat := range idx {
		coord := (flat / strides[d]) % (end - start)
		idx[flat] = int32(start + coord)
	}
	it, err := tensor.FromSlice[int32](idx, tensor.Shape(outShape), a.be)
	if err != nil {
		return nil, err
	}
	return wrapRaw(a.be.Gather(a.t.Raw(), d, it.Raw())), nil
}

func Clamp(a *Tensor, lo, hi float32) (*Tensor, error) {
	return wrapRaw(a.be.Clamp(a.t.Raw(), lo, hi)), nil
}

func Where(cond, a, b *Tensor) (*Tensor, error) {
	// blue's condition may be a float 0/1 tensor (from ml.tensor) or a bool
	// tensor (from a comparison); borncgo's Where needs a bool.
	c := cond.t.Raw()
	if cond.dtype != Bool {
		c = a.be.Cast(cond.t.Raw(), tensor.Bool)
	}
	return wrapRaw(a.be.Where(c, a.t.Raw(), b.t.Raw())), nil
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
