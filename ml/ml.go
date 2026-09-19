package ml

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"blue/borncgo/autodiff"
	"blue/borncgo/backend/cpu"
	"blue/borncgo/tensor"
)

// DType and Device are borncgo's runtime type/device tags, re-exported so the
// rest of blue does not import borncgo directly.
type DType = tensor.DataType

const (
	Float32 = tensor.Float32
	Float64 = tensor.Float64
	Int32   = tensor.Int32
	Bool    = tensor.Bool
	Invalid = tensor.DataType(-1)
)

func ParseDType(s string) (DType, error) {
	switch strings.ToLower(s) {
	case "float32":
		return Float32, nil
	case "float64":
		return Float64, nil
	case "int32":
		return Int32, nil
	case "bool":
		return Bool, nil
	default:
		return Invalid, fmt.Errorf("invalid DType: %s", s)
	}
}

// Device is blue's own device tag (lowercase names, matching blue) mapped onto
// borncgo's device enum at the boundary.
type Device int

const (
	CPU Device = iota
	GPU
	INVALID
)

func (d Device) String() string {
	switch d {
	case CPU:
		return "cpu"
	case GPU:
		return "gpu"
	default:
		return "unknown"
	}
}

func ParseDevice(s string) (Device, error) {
	switch strings.ToLower(s) {
	case "cpu":
		return CPU, nil
	case "gpu":
		return GPU, nil
	default:
		return INVALID, fmt.Errorf("invalid Device: %s", s)
	}
}

// engine is the single borncgo autodiff backend every tensor shares. Its tape
// records operations, and Backward walks it. Delegating autograd to borncgo is
// the whole point: blue only shapes the PyTorch-facing surface.
type engineB = *autodiff.Backend[*cpu.Backend]
type bornTensor = tensor.Tensor[float32, engineB]

var engine engineB

func init() {
	engine = autodiff.New(cpu.New())
	engine.Tape().StartRecording()
}

// Tensor is the blue-facing PyTorch-style tensor. It owns a borncgo tensor for
// storage and compute, and keeps the gradient bookkeeping PyTorch exposes
// (requires_grad and .grad).
type Tensor struct {
	t *bornTensor

	requiresGrad bool
	grad         *Tensor
	dtype        DType
	device       Device

	// viewStrides lets view ops report PyTorch-style strides even though
	// borncgo materializes the data behind them. nil means "use the raw strides".
	viewStrides []int
}

func wrap(t *bornTensor) *Tensor {
	return &Tensor{t: t, dtype: t.DType(), device: CPU}
}

func wrapRaw(raw *tensor.RawTensor) *Tensor {
	return wrap(tensor.New[float32](raw, engine))
}

func (t *Tensor) Shape() []int { return []int(t.t.Shape()) }
func (t *Tensor) Strides() []int {
	if t.viewStrides != nil {
		return t.viewStrides
	}
	return t.t.Raw().Strides()
}
func (t *Tensor) Offset() int    { return 0 }
func (t *Tensor) DType() DType   { return t.dtype }
func (t *Tensor) Device() Device { return t.device }
func (t *Tensor) Numel() int     { return t.t.NumElements() }

func (t *Tensor) IsContiguous() bool {
	s, st := t.Shape(), t.Strides()
	acc := 1
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] != 1 && st[i] != acc {
			return false
		}
		acc *= s[i]
	}
	return true
}

func (t *Tensor) String() string {
	return fmt.Sprintf("Tensor{shape: %v, strides: %v, offset: %d, dtype: %s, device: %s}",
		t.Shape(), t.Strides(), 0, t.DType(), t.Device())
}

// ContiguousData returns the logical elements in row-major order as float32.
// Bool and int payloads (comparison and index results) are converted to
// 0/1/values so the object layer can read every dtype the same way.
func (t *Tensor) ContiguousData() []float32 {
	raw := t.t.Raw()
	switch raw.DType() {
	case tensor.Bool:
		b := raw.AsBool()
		out := make([]float32, len(b))
		for i, v := range b {
			if v {
				out[i] = 1
			}
		}
		return out
	case tensor.Int32:
		n := raw.AsInt32()
		out := make([]float32, len(n))
		for i, v := range n {
			out[i] = float32(v)
		}
		return out
	case tensor.Int64:
		n := raw.AsInt64()
		out := make([]float32, len(n))
		for i, v := range n {
			out[i] = float32(v)
		}
		return out
	case tensor.Float64:
		f := raw.AsFloat64()
		out := make([]float32, len(f))
		for i, v := range f {
			out[i] = float32(v)
		}
		return out
	default:
		return slices.Clone(raw.AsFloat32())
	}
}

func (t *Tensor) RawData() []float32 { return t.ContiguousData() }

func (t *Tensor) Item() (float32, error) {
	if t.Numel() != 1 {
		return float32(math.NaN()), fmt.Errorf("Item: tensor does not hold exactly 1 element")
	}
	return t.t.Data()[0], nil
}

func (t *Tensor) RequiresGrad() bool { return t.requiresGrad }

func (t *Tensor) SetRequiresGrad(on bool) {
	t.requiresGrad = on
	if on {
		tracked[t] = true
	} else {
		delete(tracked, t)
	}
}

func (t *Tensor) Grad() *Tensor     { return t.grad }
func (t *Tensor) SetGrad(g *Tensor) { t.grad = g }
func (t *Tensor) ZeroGrad()         { t.grad = nil }

// Detach shares data and drops the graph (borncgo clones the raw buffer).
func (t *Tensor) Detach() *Tensor { return wrap(t.t.Detach()) }

// Clone is a detached deep copy, matching the previous blue behavior.
func (t *Tensor) Clone() *Tensor { return wrap(t.t.Detach()) }

// tracked holds every leaf that requires grad, so Backward can pick its
// gradient out of borncgo's map.
var tracked = map[*Tensor]bool{}

// Backward runs reverse-mode autodiff on borncgo's tape and stores the leaf
// gradients on the tensors.
//
// borncgo's tape seeds the *last recorded operation*, not the tensor passed in,
// so we append one connected op (loss + 0) to make the loss graph the walk root
// even when other operations ran after the loss was built.
func (t *Tensor) Backward() error {
	if t.Numel() != 1 {
		return fmt.Errorf("backward: expected a scalar loss, got shape %v", t.Shape())
	}
	if engine.Tape().NumOps() == 0 {
		return fmt.Errorf("backward: tensor is not part of a graph")
	}

	zero := tensor.Zeros[float32](tensor.Shape(t.Shape()), engine)
	sentinel := t.t.Add(zero)

	var grads map[*tensor.RawTensor]*tensor.RawTensor
	err := func() (e error) {
		defer func() {
			if r := recover(); r != nil {
				e = fmt.Errorf("%v", r)
			}
		}()
		grads = autodiff.Backward(sentinel, engine)
		return nil
	}()
	if err != nil {
		return err
	}
	engine.Tape().Clear()

	// Accumulate outside the tape so gradient bookkeeping never becomes part of
	// the next graph.
	was := engine.Tape().IsRecording()
	engine.Tape().StopRecording()
	defer func() {
		if was {
			engine.Tape().StartRecording()
		}
	}()

	for tt := range tracked {
		if !tt.requiresGrad {
			continue
		}
		g, ok := grads[tt.t.Raw()]
		if !ok || g == nil {
			continue
		}
		gt := wrapRaw(g)
		if tt.grad == nil {
			tt.grad = gt
		} else {
			tt.grad = wrapRaw(engine.Add(tt.grad.t.Raw(), gt.t.Raw()))
		}
	}
	return nil
}

// SetGradEnabled turns tape recording on or off and returns the previous value,
// so a no_grad wrapper can restore it. This is how no_grad is implemented now.
func SetGradEnabled(on bool) bool {
	was := engine.Tape().IsRecording()
	if on {
		engine.Tape().StartRecording()
	} else {
		engine.Tape().StopRecording()
	}
	return was
}

// NewTensor builds a tensor, keeping dtype as a blue-facing tag (borncgo always
// stores float32 today).
func NewTensor(data []float32, shape []int, dtype DType, device Device) (*Tensor, error) {
	bt, err := tensor.FromSlice[float32](data, tensor.Shape(shape), engine)
	if err != nil {
		return nil, err
	}
	tt := wrap(bt)
	tt.dtype = dtype
	tt.device = device
	return tt, nil
}

// NewTensorOwned is retained for the object layer's decoder.
func NewTensorOwned(data []float32, shape []int, dtype DType, device Device) (*Tensor, error) {
	return NewTensor(data, shape, dtype, device)
}

// helpers shared by the ops.

func normalizeDim(shape []int, dim int) (int, error) {
	if dim < 0 {
		dim += len(shape)
	}
	if dim < 0 || dim >= len(shape) {
		return 0, fmt.Errorf("dim %d out of range for shape %v", dim, shape)
	}
	return dim, nil
}

func normalizeDims(shape []int, dims []int) ([]int, error) {
	if dims == nil {
		all := make([]int, len(shape))
		for i := range shape {
			all[i] = i
		}
		return all, nil
	}
	out := make([]int, len(dims))
	for i, d := range dims {
		rd, err := normalizeDim(shape, d)
		if err != nil {
			return nil, err
		}
		out[i] = rd
	}
	return out, nil
}

// descendingDims sorts high to low so dropping an axis never shifts one still
// to be reduced.
func descendingDims(dims []int) []int {
	out := slices.Clone(dims)
	slices.Sort(out)
	slices.Reverse(out)
	return out
}

// asFloat casts a non-float tensor (bool masks, int indices) to float32 so it
// can take part in arithmetic and reductions.
func asFloat(a *Tensor) *Tensor {
	if a.dtype == Float32 {
		return a
	}
	f := wrapRaw(engine.Cast(a.t.Raw(), tensor.Float32))
	f.dtype = Float32
	return f
}
