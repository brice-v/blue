package ml

import (
	"fmt"
	"sort"

	"blue/borncgo/tensor"
)

// MaxDim returns the maximum values and their indices along dim, matching
// torch.max(x, dim). The values keep the dim when keepdim is set; the indices
// are int32 with the dim removed (or kept as size 1 when keepdim is set).
func MaxDim(a *Tensor, dim int, keepdim bool) (*Tensor, *Tensor, error) {
	return extremumDimPair(a, dim, keepdim, Max, ArgMax)
}

// MinDim returns the minimum values and their indices along dim, matching
// torch.min(x, dim).
func MinDim(a *Tensor, dim int, keepdim bool) (*Tensor, *Tensor, error) {
	return extremumDimPair(a, dim, keepdim, Min, ArgMin)
}

func extremumDimPair(a *Tensor, dim int, keepdim bool,
	values func(*Tensor, []int, bool) (*Tensor, error),
	indices func(*Tensor, int) (*Tensor, error)) (*Tensor, *Tensor, error) {
	d, err := normalizeDim(a.Shape(), dim)
	if err != nil {
		return nil, nil, err
	}
	v, err := values(a, []int{d}, keepdim)
	if err != nil {
		return nil, nil, err
	}
	idx, err := indices(a, d)
	if err != nil {
		return nil, nil, err
	}
	if keepdim {
		idx, err = Unsqueeze(idx, d)
		if err != nil {
			return nil, nil, err
		}
	}
	return v, idx, nil
}

// reducedCount returns how many elements a reduction over dims covers.
func reducedCount(shape []int, dims []int) int {
	if dims == nil {
		n := 1
		for _, s := range shape {
			n *= s
		}
		return n
	}
	ds, err := normalizeDims(shape, dims)
	if err != nil || len(ds) == 0 {
		n := 1
		for _, s := range shape {
			n *= s
		}
		return n
	}
	n := 1
	for _, d := range ds {
		n *= shape[d]
	}
	return n
}

// Variance is the mean of the squared deviations. unbiased divides by n-1
// instead of n, matching torch.var(unbiased=True).
func Variance(a *Tensor, dims []int, keepdim, unbiased bool) (*Tensor, error) {
	m, err := Mean(a, dims, true)
	if err != nil {
		return nil, err
	}
	d, err := Sub(a, m)
	if err != nil {
		return nil, err
	}
	sq, err := Mul(d, d)
	if err != nil {
		return nil, err
	}
	v, err := Mean(sq, dims, keepdim)
	if err != nil {
		return nil, err
	}
	if !unbiased {
		return v, nil
	}
	n := reducedCount(a.Shape(), dims)
	if n <= 1 {
		return v, nil
	}
	scale := float32(n) / float32(n-1)
	return wrapRaw(v.be, v.be.MulScalar(v.t.Raw(), scale)), nil
}

// Std is the square root of Variance.
func Std(a *Tensor, dims []int, keepdim, unbiased bool) (*Tensor, error) {
	v, err := Variance(a, dims, keepdim, unbiased)
	if err != nil {
		return nil, err
	}
	return Sqrt(v)
}

// Tile repeats a along each dim, matching torch.tile (and Tensor.repeat when
// reps has one entry per dim). It is differentiable because it is built from
// reshape and expand.
func Tile(a *Tensor, reps []int) (*Tensor, error) {
	shape := a.Shape()
	if len(reps) != len(shape) {
		return nil, fmt.Errorf("tile: reps has %d entries for rank %d", len(reps), len(shape))
	}
	for _, r := range reps {
		if r < 0 {
			return nil, fmt.Errorf("tile: negative repetition %d", r)
		}
	}
	inter := make([]int, 0, 2*len(shape))
	for _, s := range shape {
		inter = append(inter, 1, s)
	}
	x, err := Reshape(a, inter...)
	if err != nil {
		return nil, err
	}
	exp := make([]int, 0, 2*len(shape))
	for i, s := range shape {
		exp = append(exp, reps[i], s)
	}
	x, err = BroadcastTo(x, exp)
	if err != nil {
		return nil, err
	}
	out := make([]int, len(shape))
	for i, s := range shape {
		out[i] = reps[i] * s
	}
	return Reshape(x, out...)
}

// Cumsum is the cumulative sum along dim, matching torch.cumsum. It is
// differentiable because it is a matmul with a lower-triangular ones matrix.
func Cumsum(a *Tensor, dim int) (*Tensor, error) {
	shape := a.Shape()
	d, err := normalizeDim(shape, dim)
	if err != nil {
		return nil, err
	}
	rank := len(shape)
	x := a
	var perm []int
	if d != rank-1 {
		perm = make([]int, rank)
		for i := range perm {
			perm[i] = i
		}
		perm[d], perm[rank-1] = perm[rank-1], perm[d]
		x, err = Permute(a, perm...)
		if err != nil {
			return nil, err
		}
	}
	xs := x.Shape()
	n := xs[rank-1]
	lead := 1
	for _, s := range xs[:rank-1] {
		lead *= s
	}
	// Flatten the leading dims so the triangular matmul is a plain 2-D product.
	x2, err := Reshape(x, lead, n)
	if err != nil {
		return nil, err
	}
	// Upper triangular ones, so x @ L gives the inclusive prefix sum:
	// (x @ L)[i] = sum_{j <= i} x[j].
	data := make([]float32, n*n)
	for i := range n {
		for j := i; j < n; j++ {
			data[i*n+j] = 1
		}
	}
	l, err := NewTensor(data, []int{n, n}, a.dtype, a.Device())
	if err != nil {
		return nil, err
	}
	y2, err := MatMul(x2, l)
	if err != nil {
		return nil, err
	}
	y, err := Reshape(y2, xs...)
	if err != nil {
		return nil, err
	}
	if perm != nil {
		y, err = Permute(y, perm...)
		if err != nil {
			return nil, err
		}
	}
	return y, nil
}

// Sort returns the values and indices that sort a along dim, matching
// torch.sort. It is host-computed, so it is inference-only.
func Sort(a *Tensor, dim int, descending bool) (*Tensor, *Tensor, error) {
	d, err := normalizeDim(a.Shape(), dim)
	if err != nil {
		return nil, nil, err
	}
	vals, idx := sortDim(a, d, descending)
	v, err := NewTensor(vals, a.Shape(), a.dtype, a.Device())
	if err != nil {
		return nil, nil, err
	}
	i, err := NewInt32Tensor(idx, a.Shape(), a.Device())
	if err != nil {
		return nil, nil, err
	}
	return v, i, nil
}

func sortDim(a *Tensor, d int, descending bool) ([]float32, []int32) {
	shape := a.Shape()
	reduce := shape[d]
	outer, inner := 1, 1
	for _, s := range shape[:d] {
		outer *= s
	}
	for _, s := range shape[d+1:] {
		inner *= s
	}
	data := a.ContiguousData()
	vals := make([]float32, len(data))
	idx := make([]int32, len(data))
	order := make([]int, reduce)
	for o := 0; o < outer; o++ {
		for in := 0; in < inner; in++ {
			for r := range order {
				order[r] = r
			}
			sort.SliceStable(order, func(x, y int) bool {
				vx := data[(o*reduce+order[x])*inner+in]
				vy := data[(o*reduce+order[y])*inner+in]
				if descending {
					return vx > vy
				}
				return vx < vy
			})
			for r := range reduce {
				src := order[r]
				vals[(o*reduce+r)*inner+in] = data[(o*reduce+src)*inner+in]
				idx[(o*reduce+r)*inner+in] = int32(src)
			}
		}
	}
	return vals, idx
}

// TopK returns the k largest (or smallest) values and their indices along dim,
// matching torch.topk. The result is always sorted. It is host-computed through
// Sort, so it is inference-only.
func TopK(a *Tensor, k, dim int, largest bool) (*Tensor, *Tensor, error) {
	shape := a.Shape()
	d, err := normalizeDim(shape, dim)
	if err != nil {
		return nil, nil, err
	}
	if k < 0 {
		return nil, nil, fmt.Errorf("topk: k must be non-negative")
	}
	if k > shape[d] {
		k = shape[d]
	}
	v, i, err := Sort(a, d, largest)
	if err != nil {
		return nil, nil, err
	}
	vs, err := Slice(v, d, 0, k, 1)
	if err != nil {
		return nil, nil, err
	}
	is, err := Slice(i, d, 0, k, 1)
	if err != nil {
		return nil, nil, err
	}
	return vs, is, nil
}

// Nonzero returns the indices of the nonzero elements as an [n, rank] int32
// tensor, matching torch.nonzero. It is host-computed, so it is inference-only.
func Nonzero(a *Tensor) (*Tensor, error) {
	shape := a.Shape()
	data := a.ContiguousData()
	rank := len(shape)
	strides := make([]int, rank)
	acc := 1
	for i := rank - 1; i >= 0; i-- {
		strides[i] = acc
		acc *= shape[i]
	}
	out := make([]int32, 0)
	rows := 0
	for flat, v := range data {
		if v == 0 {
			continue
		}
		rem := flat
		for i := range rank {
			out = append(out, int32(rem/strides[i]))
			rem %= strides[i]
		}
		rows++
	}
	return NewInt32Tensor(out, []int{rows, rank}, a.Device())
}

// ScatterAdd returns dest with src added into the positions selected by index
// along dim, matching the engine's scatter-add. It is differentiable through the
// engine's tape.
func ScatterAdd(dest *Tensor, dim int, index, src *Tensor) (*Tensor, error) {
	d, err := normalizeDim(dest.Shape(), dim)
	if err != nil {
		return nil, err
	}
	it := index.t.Raw()
	if index.dtype != Int32 {
		it = dest.be.Cast(it, tensor.Int32)
	}
	return wrapRaw(dest.be, dest.be.ScatterAdd(dest.t.Raw(), d, it, src.t.Raw())), nil
}

// LogSoftmax is log(softmax(a, dim)) computed with a max shift for stability.
// The max is detached, which is correct because log_softmax is shift invariant.
func LogSoftmax(a *Tensor, dim int) (*Tensor, error) {
	m, err := Max(a, []int{dim}, true)
	if err != nil {
		return nil, err
	}
	shifted, err := Sub(a, m)
	if err != nil {
		return nil, err
	}
	e, err := Exp(shifted)
	if err != nil {
		return nil, err
	}
	s, err := Sum(e, []int{dim}, true)
	if err != nil {
		return nil, err
	}
	ls, err := Log(s)
	if err != nil {
		return nil, err
	}
	return Sub(shifted, ls)
}

// CrossEntropyOpts controls CrossEntropyWith.
type CrossEntropyOpts struct {
	// Reduction is "mean" (default), "sum", or "none".
	Reduction string
	// HasIgnore turns IgnoreIndex on. Targets equal to IgnoreIndex are dropped
	// from the mean denominator, matching torch's ignore_index.
	HasIgnore   bool
	IgnoreIndex int
	// Weight, when set, is a [classes] tensor applied per sample.
	Weight *Tensor
}

// CrossEntropyWith is a flexible cross-entropy loss over logits [batch, classes]
// and integer targets [batch]. Unlike the engine loss it supports reduction,
// ignore_index, and class weights, matching torch.nn.functional.cross_entropy.
func CrossEntropyWith(logits, target *Tensor, opts CrossEntropyOpts) (*Tensor, error) {
	shape := logits.Shape()
	if len(shape) != 2 {
		return nil, fmt.Errorf("cross_entropy: logits must be 2D, got %v", shape)
	}
	batch := shape[0]

	lp, err := LogSoftmax(logits, 1)
	if err != nil {
		return nil, err
	}
	idx, err := Unsqueeze(target, 1)
	if err != nil {
		return nil, err
	}
	picked, err := Gather(lp, 1, idx)
	if err != nil {
		return nil, err
	}
	loss, err := Neg(picked) // [batch, 1]
	if err != nil {
		return nil, err
	}

	var keep *Tensor
	if opts.HasIgnore {
		ign, err := NewTensor([]float32{float32(opts.IgnoreIndex)}, []int{1}, target.dtype, target.Device())
		if err != nil {
			return nil, err
		}
		keepB, err := Ne(target, ign)
		if err != nil {
			return nil, err
		}
		keepF, err := keepB.Cast(Float32)
		if err != nil {
			return nil, err
		}
		keep, err = Unsqueeze(keepF, 1)
		if err != nil {
			return nil, err
		}
		loss, err = Mul(loss, keep)
		if err != nil {
			return nil, err
		}
	}

	if opts.Weight != nil {
		w, err := Gather(opts.Weight, 0, target) // [batch]
		if err != nil {
			return nil, err
		}
		w, err = Unsqueeze(w, 1)
		if err != nil {
			return nil, err
		}
		loss, err = Mul(loss, w)
		if err != nil {
			return nil, err
		}
	}

	// Reductions return a 1-element tensor of shape [1], matching the engine
	// loss and blue's scalar convention.
	switch opts.Reduction {
	case "none":
		return Reshape(loss, batch)
	case "sum":
		s, err := Sum(loss, nil, true)
		if err != nil {
			return nil, err
		}
		return Reshape(s, 1)
	default: // mean
		var m *Tensor
		if keep != nil {
			s, err := Sum(loss, nil, true)
			if err != nil {
				return nil, err
			}
			cnt, err := Sum(keep, nil, true)
			if err != nil {
				return nil, err
			}
			m, err = Div(s, cnt)
			if err != nil {
				return nil, err
			}
		} else {
			m, err = Mean(loss, nil, true)
			if err != nil {
				return nil, err
			}
		}
		return Reshape(m, 1)
	}
}
