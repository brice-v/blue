package ml

import (
	"fmt"
	"slices"

	"blue/borncgo/tensor"
)

// Storage-sharing views
//
// A view is a Tensor whose raw tensor shares its buffer with another tensor but
// reports a different shape, strides, and offset. Views are created by the shape
// ops (reshape, transpose, permute, unsqueeze, squeeze, flatten, slice, select)
// when no gradient is being recorded, that is under no_grad or at inference.
//
// The reason for the recording condition: the engine's kernels read buffers
// linearly and assume contiguous data, and the engine's shape ops record on the
// tape. A view created instead of a recorded op would break the gradient chain.
// So during training the shape ops stay as they are (materialize and record),
// and views are an inference optimization.
//
// Any op that needs contiguous data calls contig, which returns the tensor
// itself when its raw layout is already contiguous and otherwise materializes a
// fresh buffer. ContiguousData also reads through strides, so a view always
// reports the right values.

// canView reports whether shape ops may return storage-sharing views. Views are
// only safe when no gradient is recorded and the tensor is on the CPU, because
// the engine kernels are contiguous-only and the GPU backend owns its buffers.
func canView() bool {
	return !recording
}

// isRawContiguous reports whether the tensor's buffer is row-major contiguous
// for its shape. Size-1 dims are ignored, as usual.
func (t *Tensor) isRawContiguous() bool {
	s, st := t.t.Shape(), t.t.Raw().Strides()
	acc := 1
	for i, v := range slices.Backward(s) {
		if v != 1 && st[i] != acc {
			return false
		}
		acc *= v
	}
	return true
}

// Contiguous returns a tensor with row-major contiguous storage. A contiguous
// tensor (including a contiguous view) is returned as is; a non-contiguous view
// is materialized into a fresh buffer. Under training no views exist, so this is
// a no-op.
func (t *Tensor) Contiguous() *Tensor {
	if t.isRawContiguous() {
		return t
	}
	out, err := NewTensor(t.stridedData(), t.Shape(), t.dtype, t.device)
	if err != nil {
		return t
	}
	if t.requiresGrad {
		out.SetRequiresGrad(true)
	}
	return out
}

// contig is the internal alias used at op entry points.
func (t *Tensor) contig() *Tensor { return t.Contiguous() }

// stridedData reads a non-contiguous view into a fresh row-major slice. The
// strides are relative to the view's offset, so the absolute buffer index is
// the offset plus the strided element index.
func (t *Tensor) stridedData() []float32 {
	raw := t.t.Raw()
	shape := t.t.Shape()
	strides := raw.Strides()
	baseOffset := raw.Offset() / raw.DType().Size()
	n := t.Numel()
	out := make([]float32, n)
	row := rowMajorStrides(shape)
	for flat := range n {
		rem := flat
		src := 0
		for i := range shape {
			c := rem / row[i]
			rem %= row[i]
			src += c * strides[i]
		}
		out[flat] = raw.Float32At(baseOffset + src)
	}
	return out
}

func rowMajorStrides(shape []int) []int {
	out := make([]int, len(shape))
	acc := 1
	for i := len(shape) - 1; i >= 0; i-- {
		out[i] = acc
		acc *= shape[i]
	}
	return out
}

// resolveViewShape fills in a single -1 and validates the element count.
func resolveViewShape(shape []int, current []int) ([]int, error) {
	total := 1
	for _, s := range current {
		total *= s
	}
	out := make([]int, len(shape))
	neg := -1
	prod := 1
	for i, s := range shape {
		if s == -1 {
			if neg >= 0 {
				return nil, fmt.Errorf("reshape: only one dimension can be -1")
			}
			neg = i
			continue
		}
		if s < 0 {
			return nil, fmt.Errorf("reshape: negative dimension %d", s)
		}
		out[i] = s
		prod *= s
	}
	if neg >= 0 {
		if prod == 0 || total%prod != 0 {
			return nil, fmt.Errorf("reshape: cannot infer -1 for shape %v from %v", shape, current)
		}
		out[neg] = total / prod
		prod *= out[neg]
	}
	if prod != total {
		return nil, fmt.Errorf("reshape: shape %v has %d elements, want %d", shape, prod, total)
	}
	return out, nil
}

// viewOf builds a view tensor over a's storage. strides are in elements and
// offsetElems is relative to a's storage start; the raw offset is in bytes.
func viewOf(a *Tensor, shape []int, strides []int, offsetElems int) *Tensor {
	size := a.t.Raw().DType().Size()
	raw := a.t.Raw().ViewAs(tensor.Shape(shape), strides, a.t.Raw().Offset()+offsetElems*size)
	out := wrapRaw(a.be, raw)
	out.dtype = a.dtype
	out.device = a.device
	return out
}

// View returns a storage-sharing view of a with the given shape. It is an
// explicit form of what reshape does under no_grad. It is inference-only: it
// errors when a gradient is being recorded or the tensor is not on the CPU,
// because a view does not record on the tape.
func View(a *Tensor, shape []int) (*Tensor, error) {
	if !canView() {
		return nil, fmt.Errorf("view: only available under no_grad or at inference")
	}
	if a.device != CPU {
		return nil, fmt.Errorf("view: only available on cpu")
	}
	return viewReshape(a, shape)
}

func viewReshape(a *Tensor, shape []int) (*Tensor, error) {
	s, err := resolveViewShape(shape, a.Shape())
	if err != nil {
		return nil, err
	}
	if !a.isRawContiguous() {
		return nil, fmt.Errorf("reshape: non-contiguous input")
	}
	return viewOf(a, s, rowMajorStrides(s), 0), nil
}

func viewPermute(a *Tensor, perm []int) (*Tensor, error) {
	rank := len(a.Shape())
	if len(perm) != rank {
		return nil, fmt.Errorf("permute: %d axes for rank %d", len(perm), rank)
	}
	shape := a.t.Shape()
	strides := a.t.Raw().Strides()
	ns := make([]int, rank)
	nst := make([]int, rank)
	seen := make([]bool, rank)
	for i, p := range perm {
		if p < 0 {
			p += rank
		}
		if p < 0 || p >= rank || seen[p] {
			return nil, fmt.Errorf("permute: invalid axis %d", p)
		}
		seen[p] = true
		ns[i] = shape[p]
		nst[i] = strides[p]
	}
	return viewOf(a, ns, nst, 0), nil
}

// viewSlice builds a strided view for [start:end:step] along dim. A negative
// step produces a negative stride, which is not contiguous, so the view is
// materialized on first use.
func viewSlice(a *Tensor, dim, start, end, step int) (*Tensor, error) {
	shape := a.t.Shape()
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
	var count int
	if step > 0 {
		if start < 0 {
			start = 0
		}
		if end > n {
			end = n
		}
		if end > start {
			count = (end - start + step - 1) / step
		}
	} else {
		if start > n-1 {
			start = n - 1
		}
		for i := start; i > end; i += step {
			if i >= 0 && i < n {
				count++
			}
		}
	}
	ns := make([]int, len(shape))
	copy(ns, shape)
	ns[d] = count
	nst := append([]int(nil), a.t.Raw().Strides()...)
	nst[d] *= step
	return viewOf(a, ns, nst, start*a.t.Raw().Strides()[d]), nil
}

// viewSqueeze drops the size-1 dims in drop, producing a lower-rank view.
func viewSqueeze(a *Tensor, drop []int) (*Tensor, error) {
	remove := map[int]bool{}
	for _, d := range drop {
		remove[d] = true
	}
	shape := a.t.Shape()
	strides := a.t.Raw().Strides()
	ns := make([]int, 0, len(shape))
	nst := make([]int, 0, len(shape))
	for i := range shape {
		if remove[i] {
			continue
		}
		ns = append(ns, shape[i])
		nst = append(nst, strides[i])
	}
	return viewOf(a, ns, nst, 0), nil
}

func viewUnsqueeze(a *Tensor, dim int) (*Tensor, error) {
	rank := len(a.Shape())
	if dim < 0 {
		dim += rank + 1
	}
	if dim < 0 || dim > rank {
		return nil, fmt.Errorf("unsqueeze: dim %d out of range for rank %d", dim, rank)
	}
	ns := make([]int, 0, rank+1)
	nst := make([]int, 0, rank+1)
	strides := a.t.Raw().Strides()
	for i := 0; i < rank+1; i++ {
		if i == dim {
			ns = append(ns, 1)
			nst = append(nst, 1)
			continue
		}
		src := i
		if i > dim {
			src = i - 1
		}
		ns = append(ns, a.t.Shape()[src])
		nst = append(nst, strides[src])
	}
	return viewOf(a, ns, nst, 0), nil
}
