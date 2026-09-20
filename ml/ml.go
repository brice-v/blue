package ml

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"sync"
	"weak"

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
	Int64   = tensor.Int64
	Uint8   = tensor.Uint8
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
	case "int64":
		return Int64, nil
	case "uint8":
		return Uint8, nil
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

// backends hold one borncgo autodiff backend per device. Tensors remember the
// backend they were created on, so ops run on the right device and each device
// keeps its own tape. Delegating autograd to borncgo is the whole point: blue
// only shapes the PyTorch-facing surface.
type bornTensor = tensor.Tensor[float32, tensor.Backend]

var (
	cpuEngine tensor.Backend
	gpuEngine tensor.Backend
	gpuErr    error
	recording = true
)

func init() {
	cpuEngine = autodiff.New(cpu.New())
	setRecording(cpuEngine, true)
}

// tapeOf returns the gradient tape of a backend that supports backprop.
func tapeOf(be tensor.Backend) *autodiff.GradientTape {
	if bc, ok := be.(autodiff.BackwardCapable); ok {
		return bc.GetTape()
	}
	return nil
}

func setRecording(be tensor.Backend, on bool) {
	if tp := tapeOf(be); tp != nil {
		if on {
			tp.StartRecording()
		} else {
			tp.StopRecording()
		}
	}
}

// ensureGPU lazily creates the WebGPU autodiff backend. The implementation is
// compiled in only for cgo, non-static builds (see gpu_cgo.go / gpu_stub.go).
func ensureGPU() tensor.Backend {
	if gpuEngine != nil || gpuErr != nil {
		return gpuEngine
	}
	be, err := newGPUEngine()
	if err != nil {
		gpuErr = err
		return nil
	}
	gpuEngine = be
	setRecording(gpuEngine, recording)
	return gpuEngine
}

// backendFor returns the backend for a blue device, erroring when the device is
// unavailable (for example gpu without a working WebGPU adapter).
func backendFor(d Device) (tensor.Backend, error) {
	switch d {
	case CPU:
		return cpuEngine, nil
	case GPU:
		if be := ensureGPU(); be != nil {
			return be, nil
		}
		return nil, fmt.Errorf("gpu device unavailable: %w", gpuErr)
	default:
		return nil, fmt.Errorf("unknown device %d", d)
	}
}

// backendForBorn maps a borncgo device back to the backend that owns it.
func backendForBorn(d tensor.Device) tensor.Backend {
	if d == tensor.CPU {
		return cpuEngine
	}
	if be := ensureGPU(); be != nil {
		return be
	}
	return cpuEngine
}

func mlDeviceOf(d tensor.Device) Device {
	if d == tensor.CPU {
		return CPU
	}
	return GPU
}

// Tensor is the blue-facing PyTorch-style tensor. It owns a borncgo tensor for
// storage and compute, and keeps the gradient bookkeeping PyTorch exposes
// (requires_grad and .grad).
type Tensor struct {
	t  *bornTensor
	be tensor.Backend

	requiresGrad bool
	grad         *Tensor
	dtype        DType
	device       Device

	// retainGrad keeps a non-leaf tensor's gradient after Backward, matching
	// torch.Tensor.retain_grad. hooks run during Backward, matching
	// torch.Tensor.register_hook.
	retainGrad bool
	hooks      []func(*Tensor) *Tensor

	// viewStrides lets view ops (transpose/permute) report PyTorch-style strides
	// even though borncgo materializes the data behind them. nil means "use the
	// raw strides". See Strides and Offset for the materialization caveat.
	viewStrides []int
}

func wrap(t *bornTensor) *Tensor {
	// Keep the backend that actually produced the tensor rather than mapping the
	// device back to an engine. They agree for real backends, but a tracing
	// backend needs to stay attached so recorded ops keep flowing through it.
	d := t.Device()
	return &Tensor{t: t, be: t.Backend(), dtype: t.DType(), device: mlDeviceOf(d)}
}

func wrapRaw(be tensor.Backend, raw *tensor.RawTensor) *Tensor {
	return wrap(tensor.New[float32](raw, be))
}

func (t *Tensor) Shape() []int { return []int(t.t.Shape()) }

// Raw exposes the underlying borncgo raw tensor.
func (t *Tensor) Raw() *tensor.RawTensor { return t.t.Raw() }

// Backend exposes the backend a tensor was produced on, so the graph compiler
// can replay on the same engine.
func (t *Tensor) Backend() tensor.Backend { return t.be }

// WrapRaw wraps a raw tensor produced by a known backend.
func WrapRaw(be tensor.Backend, raw *tensor.RawTensor) *Tensor { return wrapRaw(be, raw) }

// Strides reports the tensor's strides. View ops (transpose/permute) record
// PyTorch-style strides in viewStrides, because borncgo materializes the data
// behind them; the reported layout matches what a PyTorch user expects even
// though the underlying buffer is contiguous.
func (t *Tensor) Strides() []int {
	if t.viewStrides != nil {
		return t.viewStrides
	}
	return t.t.Raw().Strides()
}

// Offset is always 0. borncgo materializes view ops, so every blue tensor owns
// a buffer whose first element is index 0. PyTorch reports a nonzero offset for
// some views; blue never creates those views.
func (t *Tensor) Offset() int    { return 0 }
func (t *Tensor) DType() DType   { return t.dtype }
func (t *Tensor) Device() Device { return t.device }
func (t *Tensor) Numel() int     { return t.t.NumElements() }

func (t *Tensor) IsContiguous() bool {
	s, st := t.Shape(), t.Strides()
	acc := 1
	for i, v := range slices.Backward(s) {
		if v != 1 && st[i] != acc {
			return false
		}
		acc *= v
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
	if !t.isRawContiguous() {
		return t.stridedData()
	}
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
	case tensor.Uint8:
		n := raw.AsUint8()
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
	trackedMu.Lock()
	if on {
		tracked[weak.Make(t)] = struct{}{}
	} else {
		delete(tracked, weak.Make(t))
	}
	trackedMu.Unlock()
}

func (t *Tensor) Grad() *Tensor     { return t.grad }
func (t *Tensor) SetGrad(g *Tensor) { t.grad = g }
func (t *Tensor) ZeroGrad()         { t.grad = nil }

// Detach returns a copy that shares the value but is cut out of the graph, like
// torch.Tensor.detach.
func (t *Tensor) Detach() *Tensor { return wrap(t.t.Detach()) }

// Clone returns a deep copy that stays connected to the autograd graph, like
// torch.Tensor.clone: the copy records an identity op, so gradients flow back to
// the original. Use Detach for a copy cut out of the graph. Non-float tensors
// have nothing to differentiate, so they are deep copied without tracking.
func (t *Tensor) Clone() *Tensor {
	if t.dtype != Float32 && t.dtype != Float64 {
		return wrap(t.t.Detach())
	}
	out := wrapRaw(t.be, t.be.AddScalar(t.t.Raw(), float32(0)))
	out.requiresGrad = t.requiresGrad
	return out
}

// tracked holds a weak reference to every leaf that requires grad, so Backward
// can attach its gradient to the blue tensor. Because the reference is weak,
// the set does not keep leaves alive and is bounded by live leaves rather than
// by every leaf ever created. liveLeaves prunes entries whose tensor has been
// collected. The mutex guards the map.
var (
	trackedMu sync.Mutex
	tracked   = map[weak.Pointer[Tensor]]struct{}{}
)

// liveLeaves snapshots the tracked leaves, dropping the ones the garbage
// collector has reclaimed, so the graph walk does not hold the lock while it
// calls into the engine.
func liveLeaves() []*Tensor {
	trackedMu.Lock()
	defer trackedMu.Unlock()
	out := make([]*Tensor, 0, len(tracked))
	for wp := range tracked {
		if t := wp.Value(); t != nil {
			out = append(out, t)
		} else {
			delete(tracked, wp)
		}
	}
	return out
}

// retained holds a weak reference to tensors that asked for a non-leaf gradient
// (RetainGrad) or registered a hook, so Backward can find them after the graph
// walk. Like tracked, it does not keep tensors alive.
var (
	retainedMu sync.Mutex
	retained   = map[weak.Pointer[Tensor]]struct{}{}
)

func liveRetained() []*Tensor {
	retainedMu.Lock()
	defer retainedMu.Unlock()
	out := make([]*Tensor, 0, len(retained))
	for wp := range retained {
		if t := wp.Value(); t != nil {
			out = append(out, t)
		} else {
			delete(retained, wp)
		}
	}
	return out
}

// RetainGrad keeps this tensor's gradient after Backward even though it is not a
// leaf, matching torch.Tensor.retain_grad.
func (t *Tensor) RetainGrad() {
	t.retainGrad = true
	retainedMu.Lock()
	retained[weak.Make(t)] = struct{}{}
	retainedMu.Unlock()
}

// RegisterHook adds a function that Backward calls with this tensor's gradient.
// The value a hook returns replaces the gradient passed to the next hook,
// matching torch.Tensor.register_hook.
func (t *Tensor) RegisterHook(fn func(grad *Tensor) *Tensor) {
	t.hooks = append(t.hooks, fn)
	retainedMu.Lock()
	retained[weak.Make(t)] = struct{}{}
	retainedMu.Unlock()
}

// Backward runs reverse-mode autodiff on borncgo's tape and stores the leaf
// gradients on the tensors.
//
// borncgo's tape seeds the *last recorded operation*, not the tensor passed in,
// so we append one connected op (loss + 0) to make the loss graph the walk root
// even when other operations ran after the loss was built.
// lastGradsByBackend holds the raw gradient map from the most recent Backward,
// keyed by the backend that produced it. borncgo's optimizers consume the map
// directly, so it is kept until the next Backward on that backend. Keying by
// backend keeps a CPU graph and a GPU graph from clobbering each other; two
// models on the same backend still share the most recent backward, which is the
// single-graph-at-a-time model blue documents.
var (
	lastGradsMu        sync.Mutex
	lastGradsByBackend = map[tensor.Backend]map[*tensor.RawTensor]*tensor.RawTensor{}
)

func setLastGrads(be tensor.Backend, grads map[*tensor.RawTensor]*tensor.RawTensor) {
	lastGradsMu.Lock()
	lastGradsByBackend[be] = grads
	lastGradsMu.Unlock()
}

func lastGradsFor(be tensor.Backend) map[*tensor.RawTensor]*tensor.RawTensor {
	lastGradsMu.Lock()
	defer lastGradsMu.Unlock()
	return lastGradsByBackend[be]
}

func (t *Tensor) Backward() error {
	if t.Numel() != 1 {
		return fmt.Errorf("backward: expected a scalar loss, got shape %v", t.Shape())
	}
	be := t.be
	tp := tapeOf(be)
	if tp == nil {
		return fmt.Errorf("backward: no tape for device %s", t.Device())
	}
	if tp.NumOps() == 0 {
		return fmt.Errorf("backward: tensor is not part of a graph")
	}

	// Free the previous step's gradient buffers (held for the optimizer) before
	// computing new ones, so GPU grad memory does not accumulate across steps.
	if prev := lastGradsFor(be); prev != nil {
		autodiff.ReleaseGradients(prev)
		setLastGrads(be, nil)
	}

	one := tensor.Ones[float32](tensor.Shape(t.Shape()), be)

	var grads map[*tensor.RawTensor]*tensor.RawTensor
	err := func() (e error) {
		defer func() {
			if r := recover(); r != nil {
				e = fmt.Errorf("%v", r)
			}
		}()
		grads = tp.BackwardFrom(t.t.Raw(), one.Raw(), be)
		return nil
	}()
	if err != nil {
		return err
	}
	tp.Clear()
	setLastGrads(be, grads)

	// Accumulate outside the tape so gradient bookkeeping never becomes part of
	// the next graph.
	was := tp.IsRecording()
	tp.StopRecording()
	defer func() {
		if was {
			tp.StartRecording()
		}
	}()

	// Attach gradients to leaves and to retained non-leaves, and run hooks.
	processed := map[*Tensor]bool{}
	attach := func(tt *Tensor) {
		if processed[tt] {
			return
		}
		processed[tt] = true
		g, ok := grads[tt.t.Raw()]
		if !ok || g == nil {
			return
		}
		gt := wrapRaw(be, g)
		if tt.requiresGrad || tt.retainGrad {
			if tt.grad == nil {
				tt.grad = gt
			} else {
				tt.grad = wrapRaw(be, be.Add(tt.grad.t.Raw(), gt.t.Raw()))
			}
		}
		for _, h := range tt.hooks {
			gt = h(gt)
		}
	}
	for _, tt := range liveLeaves() {
		attach(tt)
	}
	for _, tt := range liveRetained() {
		attach(tt)
	}
	return nil
}

// AutogradGrad computes the gradients of a scalar output with respect to the
// given inputs and returns them without storing anything on the inputs,
// matching torch.autograd.grad. The tape is left intact, so a later Backward
// still works. An input that is not part of the graph gets a nil entry.
func AutogradGrad(output *Tensor, inputs []*Tensor) ([]*Tensor, error) {
	if output.Numel() != 1 {
		return nil, fmt.Errorf("autograd_grad: output must be a scalar, got shape %v", output.Shape())
	}
	be := output.be
	tp := tapeOf(be)
	if tp == nil {
		return nil, fmt.Errorf("autograd_grad: no tape for device %s", output.Device())
	}
	if tp.NumOps() == 0 {
		return nil, fmt.Errorf("autograd_grad: output is not part of a graph")
	}
	one := tensor.Ones[float32](tensor.Shape(output.Shape()), be)
	var grads map[*tensor.RawTensor]*tensor.RawTensor
	err := func() (e error) {
		defer func() {
			if r := recover(); r != nil {
				e = fmt.Errorf("%v", r)
			}
		}()
		grads = tp.BackwardFrom(output.t.Raw(), one.Raw(), be)
		return nil
	}()
	if err != nil {
		return nil, err
	}
	out := make([]*Tensor, len(inputs))
	for i, in := range inputs {
		if g, ok := grads[in.t.Raw()]; ok && g != nil {
			out[i] = wrapRaw(be, g)
		}
	}
	return out, nil
}

// SetGradEnabled turns tape recording on or off and returns the previous value,
// so a no_grad wrapper can restore it. This is how no_grad is implemented now.
func SetGradEnabled(on bool) bool {
	prev := recording
	recording = on
	setRecording(cpuEngine, on)
	if gpuEngine != nil {
		setRecording(gpuEngine, on)
	}
	return prev
}

// NewTensor builds a tensor of the requested dtype from float32 data. The
// storage matches the dtype, so a float64 tensor holds a float64 buffer and an
// int32 tensor holds an int32 buffer; the dtype tag is not cosmetic. Float32
// (and the zero value Invalid) is the fast path.
func NewTensor(data []float32, shape []int, dtype DType, device Device) (*Tensor, error) {
	switch dtype {
	case Float32, Invalid:
		return newFloat32Tensor(data, shape, device)
	case Float64:
		d := make([]float64, len(data))
		for i, v := range data {
			d[i] = float64(v)
		}
		return NewFloat64Tensor(d, shape, device)
	case Int32:
		d := make([]int32, len(data))
		for i, v := range data {
			d[i] = int32(v)
		}
		return NewInt32Tensor(d, shape, device)
	case Int64:
		d := make([]int64, len(data))
		for i, v := range data {
			d[i] = int64(v)
		}
		return NewInt64Tensor(d, shape, device)
	case Uint8:
		d := make([]uint8, len(data))
		for i, v := range data {
			d[i] = uint8(v)
		}
		return NewUint8Tensor(d, shape, device)
	case Bool:
		d := make([]bool, len(data))
		for i, v := range data {
			d[i] = v != 0
		}
		return NewBoolTensor(d, shape, device)
	default:
		return nil, fmt.Errorf("NewTensor: unsupported dtype %s", dtype)
	}
}

func newFloat32Tensor(data []float32, shape []int, device Device) (*Tensor, error) {
	be, err := backendFor(device)
	if err != nil {
		return nil, err
	}
	bt, err := tensor.FromSlice[float32](data, tensor.Shape(shape), be)
	if err != nil {
		return nil, err
	}
	tt := wrap(bt)
	tt.dtype = Float32
	tt.device = device
	return tt, nil
}

// NewTensorOwned is retained for the object layer's decoder.
func NewTensorOwned(data []float32, shape []int, dtype DType, device Device) (*Tensor, error) {
	return NewTensor(data, shape, dtype, device)
}

// GPUAvailable reports whether a GPU backend can be created on this machine.
func GPUAvailable() bool {
	return gpuIsAvailable()
}

// To returns a copy of the tensor on the requested device. Like PyTorch's
// Tensor.to, this is a data transfer, not a graph operation. It preserves the
// dtype tag but the new storage is the given device's default (float32).
func (t *Tensor) To(device Device) (*Tensor, error) {
	if t.device == device {
		return t, nil
	}
	out, err := NewTensor(t.ContiguousData(), t.Shape(), t.dtype, device)
	if err != nil {
		return nil, err
	}
	// A device move is a host copy, so the graph is not carried across devices
	// (the engine has no cross-device autodiff). The tracking flag is preserved
	// so the moved tensor can start a new graph.
	if t.requiresGrad {
		out.SetRequiresGrad(true)
	}
	return out, nil
}

// Cast returns a tensor with a different dtype on the same device, matching
// PyTorch's to(dtype).
func (t *Tensor) Cast(dtype DType) (*Tensor, error) {
	if t.dtype == dtype {
		return t, nil
	}
	return wrapRaw(t.be, t.be.Cast(t.t.Raw(), dtype)), nil
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
	a = a.contig()
	if a.dtype == Float32 {
		return a
	}
	f := wrapRaw(a.be, a.be.Cast(a.t.Raw(), tensor.Float32))
	f.dtype = Float32
	return f
}
