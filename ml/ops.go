package ml

import (
	"fmt"
	"slices"

	"blue/borncgo/tensor"
)

// promoteDType returns the dtype both operands should use, or an error when the
// pair has no common dtype (bool only combines with bool). Ordering follows
// PyTorch: float64 > float32 > int64 > int32.
func promoteDType(a, b DType) (DType, error) {
	if a == b {
		return a, nil
	}
	if a == Bool || b == Bool {
		return Invalid, fmt.Errorf("no common dtype for %s and %s", a, b)
	}
	rank := func(d DType) int {
		switch d {
		case Float64:
			return 4
		case Float32:
			return 3
		case Int64:
			return 2
		case Int32:
			return 1
		default:
			return 0
		}
	}
	if rank(a) >= rank(b) {
		return a, nil
	}
	return b, nil
}

// promote casts both operands to their common dtype so mixed-type arithmetic
// and comparisons work the way PyTorch's promotion does.
func promote(a, b *Tensor) (*Tensor, *Tensor, error) {
	d, err := promoteDType(a.dtype, b.dtype)
	if err != nil {
		return nil, nil, err
	}
	na, err := a.Cast(d)
	if err != nil {
		return nil, nil, err
	}
	nb, err := b.Cast(d)
	if err != nil {
		return nil, nil, err
	}
	return na, nb, nil
}

// Binary arithmetic delegates straight to borncgo, which handles broadcasting
// (including scalar tensors of shape [1]) and records the op on its tape.

func Add(a, b *Tensor) (*Tensor, error) {
	a, b, err := promote(a, b)
	if err != nil {
		return nil, err
	}
	return wrap(a.t.Add(b.t)), nil
}

func Sub(a, b *Tensor) (*Tensor, error) {
	a, b, err := promote(a, b)
	if err != nil {
		return nil, err
	}
	return wrap(a.t.Sub(b.t)), nil
}

func Mul(a, b *Tensor) (*Tensor, error) {
	a, b, err := promote(a, b)
	if err != nil {
		return nil, err
	}
	return wrap(a.t.Mul(b.t)), nil
}

func Div(a, b *Tensor) (*Tensor, error) {
	a, b, err := promote(a, b)
	if err != nil {
		return nil, err
	}
	return wrap(a.t.Div(b.t)), nil
}

func MatMul(a, b *Tensor) (*Tensor, error) {
	a, b, err := promote(a, b)
	if err != nil {
		return nil, err
	}
	return wrap(a.t.MatMul(b.t)), nil
}

// Pow is the one op borncgo lacks, so blue composes it. A non-negative integer
// scalar exponent uses repeated multiplication, which stays differentiable and
// works for a negative base (matching PyTorch's integer-exponent behavior).
// Everything else is exp(b * log(a)).
func Pow(a, b *Tensor) (*Tensor, error) {
	a, b, err := promote(a, b)
	if err != nil {
		return nil, err
	}
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

func Relu(a *Tensor) (*Tensor, error)    { return wrapRaw(a.be, a.be.ReLU(a.t.Raw())), nil }
func Sigmoid(a *Tensor) (*Tensor, error) { return wrapRaw(a.be, a.be.Sigmoid(a.t.Raw())), nil }
func Tanh(a *Tensor) (*Tensor, error)    { return wrapRaw(a.be, a.be.Tanh(a.t.Raw())), nil }

func Softmax(a *Tensor, dim int) (*Tensor, error) { return wrap(a.t.Softmax(dim)), nil }

// Comparisons promote to a common dtype, then return borncgo bool tensors, so
// blue reports dtype "bool" and to_list yields booleans, matching PyTorch.
func promoteCmp(a, b *Tensor, f func(be tensor.Backend, x, y *Tensor) *tensor.RawTensor) (*Tensor, error) {
	a, b, err := promote(a, b)
	if err != nil {
		return nil, err
	}
	return wrapRaw(a.be, f(a.be, a, b)), nil
}

func Eq(a, b *Tensor) (*Tensor, error) {
	return promoteCmp(a, b, func(be tensor.Backend, x, y *Tensor) *tensor.RawTensor {
		return be.Equal(x.t.Raw(), y.t.Raw())
	})
}

func Ne(a, b *Tensor) (*Tensor, error) {
	return promoteCmp(a, b, func(be tensor.Backend, x, y *Tensor) *tensor.RawTensor {
		return be.NotEqual(x.t.Raw(), y.t.Raw())
	})
}

func Gt(a, b *Tensor) (*Tensor, error) {
	return promoteCmp(a, b, func(be tensor.Backend, x, y *Tensor) *tensor.RawTensor {
		return be.Greater(x.t.Raw(), y.t.Raw())
	})
}

func Ge(a, b *Tensor) (*Tensor, error) {
	return promoteCmp(a, b, func(be tensor.Backend, x, y *Tensor) *tensor.RawTensor {
		return be.GreaterEqual(x.t.Raw(), y.t.Raw())
	})
}

func Lt(a, b *Tensor) (*Tensor, error) {
	return promoteCmp(a, b, func(be tensor.Backend, x, y *Tensor) *tensor.RawTensor {
		return be.Lower(x.t.Raw(), y.t.Raw())
	})
}

func Le(a, b *Tensor) (*Tensor, error) {
	return promoteCmp(a, b, func(be tensor.Backend, x, y *Tensor) *tensor.RawTensor {
		return be.LowerEqual(x.t.Raw(), y.t.Raw())
	})
}

// Reductions.

// Sum reduces dims, or every dim when dims is nil. keepdim keeps reduced dims as 1.
func Sum(a *Tensor, dims []int, keepdim bool) (*Tensor, error) {
	return reduce(a, dims, keepdim, func(x *Tensor, d int) *Tensor {
		return wrapRaw(x.be, x.be.SumDim(x.t.Raw(), d, keepdim))
	})
}

// Mean is borncgo's MeanDim applied per dim.
func Mean(a *Tensor, dims []int, keepdim bool) (*Tensor, error) {
	return reduce(a, dims, keepdim, func(x *Tensor, d int) *Tensor {
		return wrapRaw(x.be, x.be.MeanDim(x.t.Raw(), d, keepdim))
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

// ArgMax/ArgMin return int32 index tensors, matching PyTorch (which returns
// int64; blue's index dtype is int32).
func ArgMax(a *Tensor, dim int) (*Tensor, error) {
	return wrapRaw(a.be, a.t.Argmax(dim).Raw()), nil
}

func ArgMin(a *Tensor, dim int) (*Tensor, error) {
	n := a.t.MulScalar(float32(-1))
	return wrapRaw(a.be, n.Argmax(dim).Raw()), nil
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

// gatherAlong gathers indices along dim d. borncgo's Gather is PyTorch-style:
// the index must have the same rank as the input and its shape is the output
// shape, so the index is materialized over the full output.
func gatherAlong(a *Tensor, d int, indices []int) (*Tensor, error) {
	shape := a.Shape()
	outShape := slices.Clone(shape)
	outShape[d] = len(indices)
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
	if len(indices) > 0 {
		for flat := range idx {
			coord := (flat / strides[d]) % len(indices)
			idx[flat] = int32(indices[coord])
		}
	}
	it, err := tensor.FromSlice[int32](idx, tensor.Shape(outShape), a.be)
	if err != nil {
		return nil, err
	}
	return wrapRaw(a.be, a.be.Gather(a.t.Raw(), d, it.Raw())), nil
}

// Slice selects [start:end:step] along dim. Negative start and end count from
// the end of the dim. step defaults to 1. For a negative step the walk starts
// at start and stops before end, and end == -1 means "through index 0", which
// is the usual full reverse.
func Slice(a *Tensor, dim, start, end, step int) (*Tensor, error) {
	shape := a.Shape()
	d, err := normalizeDim(shape, dim)
	if err != nil {
		return nil, err
	}
	if step == 0 {
		return nil, fmt.Errorf("slice: step must be nonzero")
	}
	n := shape[d]
	if start < 0 {
		start += n
	}
	if end < 0 && (step >= 0 || end != -1) {
		end += n
	}
	var indices []int
	if step > 0 {
		if start < 0 {
			start = 0
		}
		if end > n {
			end = n
		}
		for i := start; i < end; i += step {
			indices = append(indices, i)
		}
	} else {
		if start > n-1 {
			start = n - 1
		}
		for i := start; i > end; i += step {
			if i >= 0 && i < n {
				indices = append(indices, i)
			}
		}
	}
	return gatherAlong(a, d, indices)
}

// Select returns the entries at index along dim with the dim removed, covering
// x[i] and x[:, j].
func Select(a *Tensor, dim, index int) (*Tensor, error) {
	shape := a.Shape()
	d, err := normalizeDim(shape, dim)
	if err != nil {
		return nil, err
	}
	n := shape[d]
	if index < 0 {
		index += n
	}
	if index < 0 || index >= n {
		return nil, fmt.Errorf("select: index %d out of range for dim %d of size %d", index, d, n)
	}
	s, err := Slice(a, d, index, index+1, 1)
	if err != nil {
		return nil, err
	}
	return Squeeze(s, []int{d})
}

// IndexSelect gathers rows (or slices) along dim using a 1d index tensor,
// matching torch.index_select.
func IndexSelect(a *Tensor, dim int, index *Tensor) (*Tensor, error) {
	shape := a.Shape()
	d, err := normalizeDim(shape, dim)
	if err != nil {
		return nil, err
	}
	data := index.ContiguousData()
	indices := make([]int, len(data))
	for i, v := range data {
		indices[i] = int(v)
	}
	return gatherAlong(a, d, indices)
}

// MaskedFill replaces elements where mask is true with value, matching
// torch.Tensor.masked_fill.
func MaskedFill(a *Tensor, mask *Tensor, value float32) (*Tensor, error) {
	fill := tensor.Full[float32](tensor.Shape(a.Shape()), value, a.be)
	return Where(mask, wrap(fill), a)
}

// Flip reverses the order of elements along each dim, matching torch.flip. It is
// differentiable because it is built from a negative-step slice (a gather).
func Flip(a *Tensor, dims []int) (*Tensor, error) {
	ds, err := normalizeDims(a.Shape(), dims)
	if err != nil {
		return nil, err
	}
	out := a
	for _, d := range ds {
		n := out.Shape()[d]
		out, err = Slice(out, d, n-1, -1, -1)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// MaskedSelect returns the elements of a where mask is nonzero, flattened to
// 1-D, matching torch.masked_select. The output length depends on the data, so
// this is inference-only: it reads host data and is not differentiable.
func MaskedSelect(a *Tensor, mask *Tensor) (*Tensor, error) {
	if !slices.Equal(a.Shape(), mask.Shape()) {
		return nil, fmt.Errorf("masked_select: shape mismatch %v vs %v", a.Shape(), mask.Shape())
	}
	data := a.ContiguousData()
	m := mask.ContiguousData()
	out := make([]float32, 0, len(data))
	for i := range data {
		if m[i] != 0 {
			out = append(out, data[i])
		}
	}
	return NewTensor(out, []int{len(out)}, a.dtype, a.Device())
}

// Embedding looks up rows of weight by index, matching
// torch.nn.functional.embedding. It is differentiable.
func Embedding(weight, indices *Tensor) (*Tensor, error) {
	it := indices.t.Raw()
	if indices.dtype != Int32 {
		it = weight.be.Cast(it, tensor.Int32)
	}
	return wrapRaw(weight.be, weight.be.Embedding(weight.t.Raw(), it)), nil
}

// BMM is batched matrix multiplication over the last two dims, matching
// torch.bmm.
func BMM(a, b *Tensor) (*Tensor, error) {
	a, b, err := promote(a, b)
	if err != nil {
		return nil, err
	}
	return wrap(a.t.BatchMatMul(b.t)), nil
}

// Silu is the Sigmoid Linear Unit, x * sigmoid(x).
func Silu(a *Tensor) (*Tensor, error) { return wrapRaw(a.be, a.be.SiLU(a.t.Raw())), nil }

func Clamp(a *Tensor, lo, hi float32) (*Tensor, error) {
	return wrapRaw(a.be, a.be.Clamp(a.t.Raw(), lo, hi)), nil
}

// Cat joins tensors along dim. All tensors must share every dimension except
// dim, matching PyTorch's torch.cat.
// this function.
func Cat(tensors []*Tensor, dim int) (*Tensor, error) {
	if len(tensors) == 0 {
		return nil, fmt.Errorf("cat: expected at least one tensor")
	}
	// Promote to a common dtype and device, then hand the raw tensors to
	// borncgo's Cat, which is differentiable and records on its tape.
	be := tensors[0].be
	shape := tensors[0].Shape()
	d, err := normalizeDim(shape, dim)
	if err != nil {
		return nil, err
	}
	dtype := tensors[0].dtype
	raws := make([]*tensor.RawTensor, len(tensors))
	for i, t := range tensors {
		if t.be != be {
			return nil, fmt.Errorf("cat: all tensors must be on the same device")
		}
		// Validate here rather than letting the backend panic: cat is a public
		// op, so a shape mismatch has to come back as an error.
		ts := t.Shape()
		if len(ts) != len(shape) {
			return nil, fmt.Errorf("cat: tensor %d has %d dimensions, want %d", i, len(ts), len(shape))
		}
		for j := range ts {
			if j != d && ts[j] != shape[j] {
				return nil, fmt.Errorf("cat: tensor %d dimension %d is %d, expected %d", i, j, ts[j], shape[j])
			}
		}
		pd, err := promoteDType(dtype, t.dtype)
		if err != nil {
			return nil, err
		}
		dtype = pd
		raws[i] = t.t.Raw()
	}
	if dtype != tensors[0].dtype {
		casted := make([]*tensor.RawTensor, len(tensors))
		for i, t := range tensors {
			casted[i] = be.Cast(t.t.Raw(), dtype)
		}
		raws = casted
	}
	out := wrapRaw(be, be.Cat(raws, d))
	out.dtype = dtype
	return out, nil
}

// Gather selects entries along dim using an int index tensor, matching
// torch.gather: out[i][j] = input[i][index[i][j]] for dim=1. It is
// differentiable, so it can be used inside models.
func Gather(a *Tensor, dim int, index *Tensor) (*Tensor, error) {
	shape := a.Shape()
	d, err := normalizeDim(shape, dim)
	if err != nil {
		return nil, err
	}
	it := index.t.Raw()
	if index.dtype != Int32 {
		it = a.be.Cast(it, tensor.Int32)
	}
	out := wrapRaw(a.be, a.be.Gather(a.t.Raw(), d, it))
	return out, nil
}

func Where(cond, a, b *Tensor) (*Tensor, error) {
	// blue's condition may be a float 0/1 tensor (from ml.tensor) or a bool
	// tensor (from a comparison); borncgo's Where needs a bool.
	c := cond.t.Raw()
	if cond.dtype != Bool {
		c = a.be.Cast(cond.t.Raw(), tensor.Bool)
	}
	return wrapRaw(a.be, a.be.Where(c, a.t.Raw(), b.t.Raw())), nil
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
