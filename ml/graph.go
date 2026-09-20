package ml

import (
	"fmt"

	"blue/borncgo/nn"
	"blue/borncgo/tensor"
)

// The graph compiler captures a model's forward pass once and replays it. It is
// a small, honest version of torch.compile: a tracing backend records the ops,
// the graph is optimised (fused / deduplicated / pruned), and the optimised
// graph is replayed with real kernels on the real backend. Backward stays with
// borncgo's tape: replay runs real ops, so the tape records them as usual.

// graphValue is one slot in a captured graph: the external input, a bound
// constant (a weight), or an intermediate produced by a node.
type graphValue struct {
	shape  tensor.Shape
	dtype  tensor.DataType
	device tensor.Device
	node   int // producing node, or -1

	// A value is bound either to a rebindable parameter (read fresh on every
	// replay, since optimizers replace the parameter's tensor), or to a fixed
	// raw constant, or it is the graph input.
	param *nn.Parameter[tensor.Backend]
	raw   *tensor.RawTensor
	input bool
}

// graphNode is one recorded op. Only the fields relevant to its op are used.
type graphNode struct {
	op   string
	ins  []int
	out  int
	dead bool

	dim     int
	keepDim bool
	axes    []int
	shape   tensor.Shape
	dtype   tensor.DataType
	scalar  any
	minAny  any
	maxAny  any
	relu    bool
	transA  bool
	transB  bool

	// chain carries the payload of a fused `elementwise` node.
	chain *elementwiseChain
}

func (n *graphNode) key() string {
	return fmt.Sprintf("%s|%v|%d|%v|%v|%v|%v|%v|%v|%v|%v|%v",
		n.op, n.ins, n.dim, n.keepDim, n.axes, n.shape, n.dtype, n.scalar, n.minAny, n.maxAny, n.relu, n.transA)
}

// Graph is a captured forward computation plus its bound constants.
type Graph struct {
	values []graphValue
	nodes  []graphNode
	input  int
	output int

	// outputs holds extra roots for a multi-output graph (a backward capture has
	// one gradient per leaf). output is always the primary root.
	outputs []int

	device tensor.Device

	fused int // nodes removed by the fusion pass, for diagnostics
}

// Tracer is a borncgo tensor.Backend that records ops instead of running them.
// It hands out fake raw tensors that carry only shape and dtype, so the model's
// forward code (which only reads shapes) runs unchanged.
type Tracer struct {
	g      *Graph
	ids    map[*tensor.RawTensor]int
	err    error
	dummy  *tensor.RawTensor
	input  int
	inputR *tensor.RawTensor

	// fakes holds the fake raw handed out for each bound leaf, so repeated
	// references to the same leaf map to the same graph value.
	fakes map[int]*tensor.RawTensor
	// leafRaws records the leaves' real raws, in graph-value order, so a
	// backward replay can bind fresh tensors to them.
	leafRaws []leafBinding
}

// leafBinding ties a bound leaf's graph value to the raw it stood in for.
type leafBinding struct {
	id  int
	raw *tensor.RawTensor
}

// paramInterner is implemented by the tracer so layer code can bind a weight as
// a rebindable parameter rather than a fixed raw.
type paramInterner interface {
	InternParam(p *nn.Parameter[tensor.Backend])
}

// Tracer must satisfy the full backend interface.
var _ tensor.Backend = (*Tracer)(nil)

// NewTracer starts a capture for an input shaped like example.
func NewTracer(example *Tensor) (*Tracer, error) {
	dev := example.t.Device()
	raw, err := tensor.NewRaw(tensor.Shape(example.Shape()), example.dtype, dev)
	if err != nil {
		return nil, err
	}
	t := &Tracer{g: &Graph{device: dev}, ids: map[*tensor.RawTensor]int{}}
	id := t.addValue(graphValue{shape: tensor.Shape(example.Shape()), dtype: example.dtype, device: dev, node: -1, input: true})
	raw.SetDataHook(func() { t.fail("host read") })
	t.ids[raw] = id
	t.input, t.inputR = id, raw
	return t, nil
}

// NewBackwardTracer starts a capture of a backward pass. Known raws (the
// forward activations, parameters and gradient seed) become graph leaves, so
// the recorded backward graph can be replayed with fresh tensors for them.
func NewBackwardTracer(dev tensor.Device) *Tracer {
	return &Tracer{g: &Graph{device: dev}, ids: map[*tensor.RawTensor]int{}}
}

// BindLeaf interns an existing raw as a graph leaf and returns the fake raw
// the backward should use in its place. The leaf is replayable: it is a graph
// input slot, not a frozen constant.
func (t *Tracer) BindLeaf(raw *tensor.RawTensor) *tensor.RawTensor {
	if id, ok := t.ids[raw]; ok {
		return t.fakeFor(id, raw.Shape(), raw.DType())
	}
	id := t.addValue(graphValue{
		shape:  raw.Shape(),
		dtype:  raw.DType(),
		device: t.g.device,
		node:   -1,
		input:  true,
	})
	// Remember raws that were bound so the caller can fill the replay slots.
	t.leafRaws = append(t.leafRaws, leafBinding{id: id, raw: raw})
	t.ids[raw] = id
	return t.fakeFor(id, raw.Shape(), raw.DType())
}

// fakeFor returns the fake raw for a value id, creating it on first use.
func (t *Tracer) fakeFor(id int, shape tensor.Shape, dtype tensor.DataType) *tensor.RawTensor {
	if r, ok := t.fakes[id]; ok {
		return r
	}
	r, err := tensor.NewRaw(shape, dtype, t.g.device)
	if err != nil {
		t.fail("bind")
		return t.dummyRaw()
	}
	r.SetDataHook(func() { t.fail("host read") })
	if t.fakes == nil {
		t.fakes = map[int]*tensor.RawTensor{}
	}
	t.fakes[id] = r
	t.ids[r] = id
	return r
}

// InputTensor is the fake input the model's forward should be called with.
func (t *Tracer) InputTensor() *Tensor {
	return wrapRaw(t, t.inputR)
}

// Finish closes the capture at the given output.
func (t *Tracer) Finish(out *Tensor) (*Graph, error) {
	if t.err != nil {
		return nil, t.err
	}
	id, ok := t.ids[out.t.Raw()]
	if !ok {
		return nil, fmt.Errorf("compile: output tensor was not produced while tracing")
	}
	t.g.output = id
	return t.g, nil
}

// InternParam binds a parameter's current raw to a rebindable slot.
func (t *Tracer) InternParam(p *nn.Parameter[tensor.Backend]) {
	if p == nil {
		return
	}
	raw := p.Tensor().Raw()
	if _, ok := t.ids[raw]; ok {
		return
	}
	id := t.addValue(graphValue{shape: p.Tensor().Shape(), dtype: p.Tensor().DType(), device: raw.Device(), node: -1, param: p})
	t.ids[raw] = id
}

func (t *Tracer) addValue(v graphValue) int {
	t.g.values = append(t.g.values, v)
	return len(t.g.values) - 1
}

// slot returns the graph id for a raw, interning unseen raws as constants.
func (t *Tracer) slot(x *tensor.RawTensor) int {
	if id, ok := t.ids[x]; ok {
		return id
	}
	id := t.addValue(graphValue{shape: x.Shape(), dtype: x.DType(), device: t.g.device, node: -1, raw: x})
	t.ids[x] = id
	return id
}

func (t *Tracer) emit(op string, ins []int, shape tensor.Shape, dtype tensor.DataType, cfg func(*graphNode)) *tensor.RawTensor {
	if t.err != nil {
		return t.dummyRaw()
	}
	id := t.addValue(graphValue{shape: shape, dtype: dtype, device: t.g.device, node: len(t.g.nodes)})
	n := graphNode{op: op, ins: ins, out: id}
	if cfg != nil {
		cfg(&n)
	}
	t.g.nodes = append(t.g.nodes, n)
	raw, err := tensor.NewRaw(shape, dtype, t.g.device)
	if err != nil {
		t.fail(op)
		return t.dummyRaw()
	}
	// A fake tensor has no real data. Reading it (a host read such as .item() or
	// to_list() inside the traced forward) must fail the capture rather than
	// silently bake in zeros, so the caller can fall back to eager execution.
	raw.SetDataHook(func() { t.fail("host read") })
	t.ids[raw] = id
	return raw
}

func (t *Tracer) dummyRaw() *tensor.RawTensor {
	if t.dummy == nil {
		r, err := tensor.NewRaw(tensor.Shape{1}, tensor.Float32, t.g.device)
		if err != nil {
			panic(err)
		}
		t.dummy = r
	}
	return t.dummy
}

func (t *Tracer) fail(op string) {
	if t.err == nil {
		t.err = fmt.Errorf("compile: op %q is not supported inside a traced region", op)
	}
}

// Unsupported: convolution and scatter ops are not recorded yet.
func (t *Tracer) unsupported(op string) *tensor.RawTensor {
	t.fail(op)
	return t.dummyRaw()
}

// ---- tensor.Backend implementation ----

func (t *Tracer) Name() string          { return "Trace" }
func (t *Tracer) Device() tensor.Device { return t.g.device }

func (t *Tracer) binary(op string, a, b *tensor.RawTensor) *tensor.RawTensor {
	if t.err != nil {
		return t.dummyRaw()
	}
	shape, _, err := tensor.BroadcastShapes(a.Shape(), b.Shape())
	if err != nil {
		t.fail(op)
		return t.dummyRaw()
	}
	return t.emit(op, []int{t.slot(a), t.slot(b)}, shape, a.DType(), nil)
}

func (t *Tracer) Add(a, b *tensor.RawTensor) *tensor.RawTensor { return t.binary("add", a, b) }
func (t *Tracer) Sub(a, b *tensor.RawTensor) *tensor.RawTensor { return t.binary("sub", a, b) }
func (t *Tracer) Mul(a, b *tensor.RawTensor) *tensor.RawTensor { return t.binary("mul", a, b) }
func (t *Tracer) Div(a, b *tensor.RawTensor) *tensor.RawTensor { return t.binary("div", a, b) }

func (t *Tracer) MatMul(a, b *tensor.RawTensor) *tensor.RawTensor {
	if t.err != nil {
		return t.dummyRaw()
	}
	shape, err := matmulShape(a.Shape(), b.Shape(), false, false)
	if err != nil {
		t.fail("matmul")
		return t.dummyRaw()
	}
	return t.emit("matmul", []int{t.slot(a), t.slot(b)}, shape, a.DType(), nil)
}

// MatMulTransposed is the fused-path multiply used by autodiff's backward.
func (t *Tracer) MatMulTransposed(a, b *tensor.RawTensor, transA, transB bool) *tensor.RawTensor {
	if t.err != nil {
		return t.dummyRaw()
	}
	shape, err := matmulShape(a.Shape(), b.Shape(), transA, transB)
	if err != nil {
		t.fail("matmul_transposed")
		return t.dummyRaw()
	}
	return t.emit("matmul", []int{t.slot(a), t.slot(b)}, shape, a.DType(), func(n *graphNode) {
		n.transA, n.transB = transA, transB
	})
}

// MatMulBias records the fused Linear epilogue y = relu(x @ W^T + bias).
func (t *Tracer) MatMulBias(a, other, bias *tensor.RawTensor, relu bool) *tensor.RawTensor {
	if t.err != nil {
		return t.dummyRaw()
	}
	if len(a.Shape()) != 2 || len(other.Shape()) != 2 {
		t.fail("matmul_bias")
		return t.dummyRaw()
	}
	out := tensor.Shape{a.Shape()[0], other.Shape()[0]}
	return t.emit("matmul_bias", []int{t.slot(a), t.slot(other), t.slot(bias)}, out, a.DType(), func(n *graphNode) {
		n.relu = relu
	})
}

func (t *Tracer) BatchMatMul(a, b *tensor.RawTensor) *tensor.RawTensor {
	if t.err != nil {
		return t.dummyRaw()
	}
	shape, _, err := tensor.BroadcastShapesMatMul(a.Shape(), b.Shape())
	if err != nil {
		t.fail("batch_matmul")
		return t.dummyRaw()
	}
	return t.emit("batch_matmul", []int{t.slot(a), t.slot(b)}, shape, a.DType(), nil)
}
func (t *Tracer) Conv2D(input, kernel *tensor.RawTensor, stride, padding int) *tensor.RawTensor {
	return t.unsupported("conv2d")
}
func (t *Tracer) MaxPool2D(input *tensor.RawTensor, kernelSize, stride int) *tensor.RawTensor {
	return t.unsupported("maxpool2d")
}
func (t *Tracer) Conv2DInputBackward(input, kernel, grad *tensor.RawTensor, stride, padding int) *tensor.RawTensor {
	return t.unsupported("conv2d_input_backward")
}
func (t *Tracer) Conv2DKernelBackward(input, kernel, grad *tensor.RawTensor, stride, padding int) *tensor.RawTensor {
	return t.unsupported("conv2d_kernel_backward")
}
func (t *Tracer) MaxPool2DBackward(input, grad *tensor.RawTensor, maxIndices []int, kernelSize, stride int) *tensor.RawTensor {
	return t.unsupported("maxpool2d_backward")
}

func (t *Tracer) Reshape(x *tensor.RawTensor, newShape tensor.Shape) *tensor.RawTensor {
	if t.err != nil {
		return t.dummyRaw()
	}
	shape, err := resolveShape(newShape, x.Shape(), x.NumElements())
	if err != nil {
		t.fail("reshape")
		return t.dummyRaw()
	}
	return t.emit("reshape", []int{t.slot(x)}, shape, x.DType(), func(n *graphNode) { n.shape = shape })
}

func (t *Tracer) Transpose(x *tensor.RawTensor, axes ...int) *tensor.RawTensor {
	if t.err != nil {
		return t.dummyRaw()
	}
	shape := x.Shape()
	perm := transposePerm(len(shape), axes)
	out := make(tensor.Shape, len(shape))
	for i, p := range perm {
		out[i] = shape[p]
	}
	return t.emit("transpose", []int{t.slot(x)}, out, x.DType(), func(n *graphNode) { n.axes = perm })
}

func (t *Tracer) scalarBinary(op string, x *tensor.RawTensor, scalar any) *tensor.RawTensor {
	if t.err != nil {
		return t.dummyRaw()
	}
	return t.emit(op, []int{t.slot(x)}, x.Shape(), x.DType(), func(n *graphNode) { n.scalar = scalar })
}

func (t *Tracer) MulScalar(x *tensor.RawTensor, scalar any) *tensor.RawTensor {
	return t.scalarBinary("mul_scalar", x, scalar)
}
func (t *Tracer) AddScalar(x *tensor.RawTensor, scalar any) *tensor.RawTensor {
	return t.scalarBinary("add_scalar", x, scalar)
}
func (t *Tracer) SubScalar(x *tensor.RawTensor, scalar any) *tensor.RawTensor {
	return t.scalarBinary("sub_scalar", x, scalar)
}
func (t *Tracer) DivScalar(x *tensor.RawTensor, scalar any) *tensor.RawTensor {
	return t.scalarBinary("div_scalar", x, scalar)
}

func (t *Tracer) unary(op string, x *tensor.RawTensor) *tensor.RawTensor {
	if t.err != nil {
		return t.dummyRaw()
	}
	return t.emit(op, []int{t.slot(x)}, x.Shape(), x.DType(), nil)
}

func (t *Tracer) Exp(x *tensor.RawTensor) *tensor.RawTensor   { return t.unary("exp", x) }
func (t *Tracer) Log(x *tensor.RawTensor) *tensor.RawTensor   { return t.unary("log", x) }
func (t *Tracer) Sqrt(x *tensor.RawTensor) *tensor.RawTensor  { return t.unary("sqrt", x) }
func (t *Tracer) Rsqrt(x *tensor.RawTensor) *tensor.RawTensor { return t.unary("rsqrt", x) }
func (t *Tracer) Cos(x *tensor.RawTensor) *tensor.RawTensor   { return t.unary("cos", x) }
func (t *Tracer) Sin(x *tensor.RawTensor) *tensor.RawTensor   { return t.unary("sin", x) }
func (t *Tracer) Erf(x *tensor.RawTensor) *tensor.RawTensor   { return t.unary("erf", x) }
func (t *Tracer) Sign(x *tensor.RawTensor) *tensor.RawTensor  { return t.unary("sign", x) }
func (t *Tracer) Abs(x *tensor.RawTensor) *tensor.RawTensor   { return t.unary("abs", x) }

func (t *Tracer) Clamp(x *tensor.RawTensor, minBound, maxBound any) *tensor.RawTensor {
	if t.err != nil {
		return t.dummyRaw()
	}
	return t.emit("clamp", []int{t.slot(x)}, x.Shape(), x.DType(), func(n *graphNode) {
		n.minAny, n.maxAny = minBound, maxBound
	})
}

func (t *Tracer) ReLU(x *tensor.RawTensor) *tensor.RawTensor    { return t.unary("relu", x) }
func (t *Tracer) Sigmoid(x *tensor.RawTensor) *tensor.RawTensor { return t.unary("sigmoid", x) }
func (t *Tracer) Tanh(x *tensor.RawTensor) *tensor.RawTensor    { return t.unary("tanh", x) }
func (t *Tracer) SiLU(x *tensor.RawTensor) *tensor.RawTensor    { return t.unary("silu", x) }

func (t *Tracer) Softmax(x *tensor.RawTensor, dim int) *tensor.RawTensor {
	if t.err != nil {
		return t.dummyRaw()
	}
	return t.emit("softmax", []int{t.slot(x)}, x.Shape(), x.DType(), func(n *graphNode) { n.dim = dim })
}

func (t *Tracer) comparison(op string, a, b *tensor.RawTensor) *tensor.RawTensor {
	if t.err != nil {
		return t.dummyRaw()
	}
	shape, _, err := tensor.BroadcastShapes(a.Shape(), b.Shape())
	if err != nil {
		t.fail(op)
		return t.dummyRaw()
	}
	return t.emit(op, []int{t.slot(a), t.slot(b)}, shape, tensor.Bool, nil)
}

func (t *Tracer) Greater(a, b *tensor.RawTensor) *tensor.RawTensor {
	return t.comparison("greater", a, b)
}
func (t *Tracer) Lower(a, b *tensor.RawTensor) *tensor.RawTensor { return t.comparison("lower", a, b) }
func (t *Tracer) GreaterEqual(a, b *tensor.RawTensor) *tensor.RawTensor {
	return t.comparison("greater_equal", a, b)
}
func (t *Tracer) LowerEqual(a, b *tensor.RawTensor) *tensor.RawTensor {
	return t.comparison("lower_equal", a, b)
}
func (t *Tracer) Equal(a, b *tensor.RawTensor) *tensor.RawTensor { return t.comparison("equal", a, b) }
func (t *Tracer) NotEqual(a, b *tensor.RawTensor) *tensor.RawTensor {
	return t.comparison("not_equal", a, b)
}

func (t *Tracer) Or(a, b *tensor.RawTensor) *tensor.RawTensor  { return t.binary("or", a, b) }
func (t *Tracer) And(a, b *tensor.RawTensor) *tensor.RawTensor { return t.binary("and", a, b) }
func (t *Tracer) Not(x *tensor.RawTensor) *tensor.RawTensor    { return t.unary("not", x) }

func (t *Tracer) Sum(x *tensor.RawTensor) *tensor.RawTensor {
	if t.err != nil {
		return t.dummyRaw()
	}
	return t.emit("sum", []int{t.slot(x)}, tensor.Shape{}, x.DType(), nil)
}

func (t *Tracer) SumDim(x *tensor.RawTensor, dim int, keepDim bool) *tensor.RawTensor {
	return t.reduceDim("sum_dim", x, dim, keepDim)
}
func (t *Tracer) MeanDim(x *tensor.RawTensor, dim int, keepDim bool) *tensor.RawTensor {
	return t.reduceDim("mean_dim", x, dim, keepDim)
}

func (t *Tracer) reduceDim(op string, x *tensor.RawTensor, dim int, keepDim bool) *tensor.RawTensor {
	if t.err != nil {
		return t.dummyRaw()
	}
	shape := x.Shape()
	if dim < 0 {
		dim += len(shape)
	}
	if dim < 0 || dim >= len(shape) {
		t.fail(op)
		return t.dummyRaw()
	}
	out := make(tensor.Shape, 0, len(shape))
	for i, s := range shape {
		if i == dim {
			if keepDim {
				out = append(out, 1)
			}
			continue
		}
		out = append(out, s)
	}
	return t.emit(op, []int{t.slot(x)}, out, x.DType(), func(n *graphNode) {
		n.dim, n.keepDim = dim, keepDim
	})
}

func (t *Tracer) Argmax(x *tensor.RawTensor, dim int) *tensor.RawTensor {
	if t.err != nil {
		return t.dummyRaw()
	}
	shape := x.Shape()
	if dim < 0 {
		dim += len(shape)
	}
	out := make(tensor.Shape, 0, len(shape)-1)
	for i, s := range shape {
		if i != dim {
			out = append(out, s)
		}
	}
	return t.emit("argmax", []int{t.slot(x)}, out, tensor.Int32, func(n *graphNode) { n.dim = dim })
}

func (t *Tracer) Cat(tensors []*tensor.RawTensor, dim int) *tensor.RawTensor {
	if t.err != nil {
		return t.dummyRaw()
	}
	if len(tensors) == 0 {
		t.fail("cat")
		return t.dummyRaw()
	}
	shape := append(tensor.Shape(nil), tensors[0].Shape()...)
	if dim < 0 {
		dim += len(shape)
	}
	for _, x := range tensors[1:] {
		s := x.Shape()
		if len(s) != len(shape) {
			t.fail("cat")
			return t.dummyRaw()
		}
		for i := range s {
			if i == dim {
				shape[i] += s[i]
			} else if s[i] != shape[i] {
				t.fail("cat")
				return t.dummyRaw()
			}
		}
	}
	ins := make([]int, len(tensors))
	for i, x := range tensors {
		ins[i] = t.slot(x)
	}
	return t.emit("cat", ins, shape, tensors[0].DType(), func(n *graphNode) { n.dim = dim })
}

func (t *Tracer) Chunk(x *tensor.RawTensor, n, dim int) []*tensor.RawTensor {
	t.fail("chunk")
	return []*tensor.RawTensor{t.dummyRaw()}
}

func (t *Tracer) Unsqueeze(x *tensor.RawTensor, dim int) *tensor.RawTensor {
	if t.err != nil {
		return t.dummyRaw()
	}
	shape := x.Shape()
	if dim < 0 {
		dim += len(shape) + 1
	}
	out := make(tensor.Shape, 0, len(shape)+1)
	out = append(out, shape[:dim]...)
	out = append(out, 1)
	out = append(out, shape[dim:]...)
	return t.emit("unsqueeze", []int{t.slot(x)}, out, x.DType(), func(n *graphNode) { n.dim = dim })
}

func (t *Tracer) Squeeze(x *tensor.RawTensor, dim int) *tensor.RawTensor {
	if t.err != nil {
		return t.dummyRaw()
	}
	shape := x.Shape()
	if dim < 0 {
		dim += len(shape)
	}
	out := make(tensor.Shape, 0, len(shape))
	for i, s := range shape {
		if i == dim && s == 1 {
			continue
		}
		out = append(out, s)
	}
	return t.emit("squeeze", []int{t.slot(x)}, out, x.DType(), func(n *graphNode) { n.dim = dim })
}

func (t *Tracer) Gather(x *tensor.RawTensor, dim int, index *tensor.RawTensor) *tensor.RawTensor {
	if t.err != nil {
		return t.dummyRaw()
	}
	shape := append(tensor.Shape(nil), x.Shape()...)
	ish := index.Shape()
	if dim < 0 {
		dim += len(shape)
	}
	out := tensor.Shape{}
	for i := 0; i < dim && i < len(shape); i++ {
		out = append(out, shape[i])
	}
	out = append(out, ish...)
	for i := dim + 1; i < len(shape); i++ {
		out = append(out, shape[i])
	}
	return t.emit("gather", []int{t.slot(x), t.slot(index)}, out, x.DType(), func(n *graphNode) { n.dim = dim })
}

func (t *Tracer) Where(condition, x, y *tensor.RawTensor) *tensor.RawTensor {
	if t.err != nil {
		return t.dummyRaw()
	}
	s1, _, err1 := tensor.BroadcastShapes(condition.Shape(), x.Shape())
	s2, _, err2 := tensor.BroadcastShapes(s1, y.Shape())
	if err1 != nil || err2 != nil {
		t.fail("where")
		return t.dummyRaw()
	}
	return t.emit("where", []int{t.slot(condition), t.slot(x), t.slot(y)}, s2, x.DType(), nil)
}

func (t *Tracer) Embedding(weight, indices *tensor.RawTensor) *tensor.RawTensor {
	if t.err != nil {
		return t.dummyRaw()
	}
	w := weight.Shape()
	out := append(append(tensor.Shape(nil), indices.Shape()...), w[1:]...)
	return t.emit("embedding", []int{t.slot(weight), t.slot(indices)}, out, weight.DType(), nil)
}

func (t *Tracer) SelectAdd(dest *tensor.RawTensor, dim int, indices *tensor.RawTensor, src *tensor.RawTensor) *tensor.RawTensor {
	return t.unsupported("select_add")
}
func (t *Tracer) ScatterAdd(dest *tensor.RawTensor, dim int, indices *tensor.RawTensor, src *tensor.RawTensor) *tensor.RawTensor {
	return t.unsupported("scatter_add")
}

func (t *Tracer) Expand(x *tensor.RawTensor, shape tensor.Shape) *tensor.RawTensor {
	if t.err != nil {
		return t.dummyRaw()
	}
	return t.emit("expand", []int{t.slot(x)}, shape, x.DType(), func(n *graphNode) { n.shape = shape })
}

func (t *Tracer) Cast(x *tensor.RawTensor, dtype tensor.DataType) *tensor.RawTensor {
	if t.err != nil {
		return t.dummyRaw()
	}
	return t.emit("cast", []int{t.slot(x)}, x.Shape(), dtype, func(n *graphNode) { n.dtype = dtype })
}

// Materialize is never valid during a capture; a traced tensor has no data.
func (t *Tracer) Materialize(x *tensor.RawTensor) ([]byte, error) {
	return nil, fmt.Errorf("compile: cannot read tensor data while tracing")
}

// ---- shape helpers ----

func matmulShape(a, b tensor.Shape, transA, transB bool) (tensor.Shape, error) {
	if len(a) != 2 || len(b) != 2 {
		return nil, fmt.Errorf("matmul: only 2D is supported in traced regions")
	}
	am, ak := a[0], a[1]
	bk, bn := b[0], b[1]
	if transA {
		am, ak = ak, am
	}
	if transB {
		bk, bn = bn, bk
	}
	if ak != bk {
		return nil, fmt.Errorf("matmul: inner dims differ (%d vs %d)", ak, bk)
	}
	return tensor.Shape{am, bn}, nil
}

func transposePerm(ndim int, axes []int) []int {
	if len(axes) == 0 {
		perm := make([]int, ndim)
		for i := range perm {
			perm[i] = ndim - 1 - i
		}
		return perm
	}
	return append([]int(nil), axes...)
}

// resolveShape fills in a -1 dim and validates the element count.
func resolveShape(newShape tensor.Shape, old tensor.Shape, numel int) (tensor.Shape, error) {
	out := append(tensor.Shape(nil), newShape...)
	neg := -1
	prod := 1
	for i, d := range out {
		if d == -1 {
			if neg >= 0 {
				return nil, fmt.Errorf("reshape: only one -1 dim is allowed")
			}
			neg = i
			continue
		}
		if d <= 0 {
			return nil, fmt.Errorf("reshape: invalid dim %d", d)
		}
		prod *= d
	}
	if neg >= 0 {
		if prod == 0 || numel%prod != 0 {
			return nil, fmt.Errorf("reshape: cannot infer dim (numel %d, known %d)", numel, prod)
		}
		out[neg] = numel / prod
	} else if prod != numel {
		return nil, fmt.Errorf("reshape: %v has %d elements, want %d", old, numel, prod)
	}
	return out, nil
}
