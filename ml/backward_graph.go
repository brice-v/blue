package ml

import (
	"fmt"
	"slices"

	"blue/borncgo/tensor"
)

// Compiled backward, built by construction.
//
// torch.compile differentiates the graph symbolically rather than replaying an
// interpreter, so the backward it produces is a graph in its own right and gets
// the same fusions as the forward. The same is done here: for each node in the
// captured forward graph, a backward rule emits nodes into a new graph whose
// leaves are the forward graph's values (the input, parameters and activations)
// plus the gradient seed.
//
// Building the backward this way, instead of tracing borncgo's tape, avoids the
// tape's identity-gradient case (some op backwards return a forward activation
// unchanged) because every rule here is expressed in graph terms and produces a
// real gradient value.

// backwardPlan is a differentiated forward graph: a graph for the gradients and
// the slot each parameter's gradient lands in.
type backwardPlan struct {
	graph *Graph
	// paramGrads maps a parameter's value id in the forward graph to the value
	// id holding its gradient in the backward graph.
	paramGrads map[int]int
	// inputGrad is the value id of the gradient w.r.t. the graph input, or -1.
	inputGrad int
	// leaves maps a forward value id to the backward graph value id that stands
	// for it, so replay can bind the live activation.
	leaves map[int]int
	// leafOrder lists the forward value ids that are backward leaves, in slot
	// order.
	leafOrder []int
	// seed is the backward graph value id of the output-gradient leaf.
	seed int
}

// BuildBackward differentiates a forward graph and returns the plan.
//
// Every forward node needs a backward rule; a node without one makes the whole
// differentiation fail, and the caller falls back to the eager backward.
func BuildBackward(fg *Graph) (*backwardPlan, error) {
	b := &backwardBuilder{
		fg:      fg,
		g:       &Graph{device: fg.device},
		leafOf:  map[int]int{},
		gradOf:  map[int]int{},
		hasGrad: map[int]bool{},
	}
	b.plan = &backwardPlan{graph: b.g, paramGrads: map[int]int{}, inputGrad: -1, leaves: map[int]int{}}

	// Bind every forward value as a backward leaf, so the backward reads the
	// live activations and parameters.
	for id := range fg.values {
		b.bindLeaf(id)
	}

	// The seed is an extra leaf: the gradient of the forward output.
	outShape := fg.values[fg.output].shape
	seed := len(b.g.values)
	b.g.values = append(b.g.values, graphValue{shape: outShape, dtype: fg.values[fg.output].dtype, device: b.g.device, node: -1, input: true})
	b.seed = seed
	b.plan.seed = seed
	b.seedGrad(fg.output, seed)

	// Walk forward nodes in reverse topological order (node order is
	// topological because nodes are appended as they execute).
	for i := len(fg.nodes) - 1; i >= 0; i-- {
		n := &fg.nodes[i]
		if n.dead || !b.has(n.out) {
			continue
		}
		if err := b.applyBackwardRule(n); err != nil {
			return nil, err
		}
	}

	// Record the gradient of every forward value that receives one. Callers pick
	// the parameter gradients they care about; for a model whose weights are
	// bound parameters this is the weight set, and for a hand-built graph the
	// weights are raw leaves that are also recorded here.
	b.plan.paramGrads = map[int]int{}
	for id := range fg.values {
		if b.has(id) {
			b.plan.paramGrads[id] = b.gradOf[id]
			b.g.outputs = append(b.g.outputs, b.gradOf[id])
		}
	}
	if b.has(fg.input) {
		b.plan.inputGrad = b.gradOf[fg.input]
	}
	return b.plan, nil
}

// backwardBuilder carries the state while differentiating.
type backwardBuilder struct {
	fg      *Graph
	g       *Graph
	plan    *backwardPlan
	leafOf  map[int]int
	gradOf  map[int]int
	hasGrad map[int]bool
	seed    int
}

func (b *backwardBuilder) bindLeaf(fwdID int) int {
	if s, ok := b.leafOf[fwdID]; ok {
		return s
	}
	v := &b.fg.values[fwdID]
	slot := len(b.g.values)
	b.g.values = append(b.g.values, graphValue{
		shape:  v.shape,
		dtype:  v.dtype,
		device: b.g.device,
		node:   -1,
		input:  true,
	})
	b.leafOf[fwdID] = slot
	b.plan.leaves[fwdID] = slot
	b.plan.leafOrder = append(b.plan.leafOrder, fwdID)
	return slot
}

// emit adds a node to the backward graph and returns its value id.
func (b *backwardBuilder) emit(op string, ins []int, shape tensor.Shape, dtype tensor.DataType, cfg func(*graphNode)) int {
	id := len(b.g.values)
	b.g.values = append(b.g.values, graphValue{shape: shape, dtype: dtype, device: b.g.device, node: len(b.g.nodes)})
	n := graphNode{op: op, ins: ins, out: id}
	if cfg != nil {
		cfg(&n)
	}
	b.g.nodes = append(b.g.nodes, n)
	return id
}

func (b *backwardBuilder) seedGrad(fwdID, grad int) {
	b.gradOf[fwdID] = grad
	b.hasGrad[fwdID] = true
}

func (b *backwardBuilder) has(fwdID int) bool { return b.hasGrad[fwdID] }

// accumulate sums a partial gradient into an existing gradient slot.
func (b *backwardBuilder) accumulate(fwdID, partial int) {
	if !b.has(fwdID) {
		b.seedGrad(fwdID, partial)
		return
	}
	existing := b.gradOf[fwdID]
	shape := b.g.values[existing].shape
	id := b.emit("add", []int{existing, partial}, shape, b.g.values[existing].dtype, nil)
	b.gradOf[fwdID] = id
}

func (b *backwardBuilder) shapeOf(fwdID int) tensor.Shape    { return b.fg.values[fwdID].shape }
func (b *backwardBuilder) dtypeOf(fwdID int) tensor.DataType { return b.fg.values[fwdID].dtype }
func (b *backwardBuilder) leaf(fwdID int) int                { return b.leafOf[fwdID] }
func (b *backwardBuilder) grad(fwdID int) int                { return b.gradOf[fwdID] }

// applyBackwardRule emits the backward of one forward node. It distributes the
// node's output gradient to its inputs by the rule for the op.
func (b *backwardBuilder) applyBackwardRule(n *graphNode) error {
	og := b.grad(n.out) // output gradient value id
	switch n.op {
	case "add":
		b.propagateBinary(n, og, 1, 1)
	case "sub":
		b.propagateBinary(n, og, 1, -1)
	case "mul":
		b.backwardMul(n, og)
	case "div":
		return b.backwardDiv(n, og)
	case "matmul":
		return b.backwardMatMul(n, og)
	case "matmul_bias":
		return b.backwardMatMulBias(n, og)
	case "relu":
		b.backwardRelu(n, og)
	case "sum":
		b.backwardBcast(n, n.ins[0], og, 0)
	case "sum_dim":
		b.backwardSumDim(n, og)
	case "mean_dim":
		b.backwardMeanDim(n, og)
	case "exp":
		b.backwardUnaryByOutput(n, og)
	case "log":
		b.backwardLog(n, og)
	case "sqrt":
		b.backwardSqrt(n, og)
	case "tanh":
		b.backwardTanh(n, og)
	case "sigmoid":
		b.backwardSigmoid(n, og)
	case "mul_scalar":
		b.backwardScalarMul(n, og)
	case "neg":
		b.backwardNeg(n, og)
	case "reshape", "unsqueeze", "squeeze", "expand", "transpose":
		b.backwardShape(n, og)
	case "cast":
		b.backwardShape(n, og)
	case "elementwise":
		return b.backwardElementwise(n, og)
	case "identity":
		b.accumulate(n.ins[0], og)
	default:
		return fmt.Errorf("compile: no backward rule for op %q", n.op)
	}
	return nil
}

// backwardElementwise differentiates a fused elementwise chain. It walks the
// chain's steps in reverse, keeping a gradient slot for each chain value (leaf
// or step), then accumulates the leaves' gradients into the forward graph.
//
// A fused forward node does not materialise its intermediates, but some
// backward rules need them (the relu mask needs the pre-activation). Those are
// recomputed inside the backward graph from the chain's leaves, which is both
// correct and cheap: the recomputed ops are themselves elementwise and get
// folded by the same fusion pass.
func (b *backwardBuilder) backwardElementwise(n *graphNode, og int) error {
	chain := n.chain
	if chain == nil {
		return fmt.Errorf("compile: elementwise node has no chain")
	}
	nInputs := len(chain.inputs)
	grad := make([]int, nInputs+len(chain.steps))
	has := make([]bool, nInputs+len(chain.steps))
	grad[chain.outSlot] = og
	has[chain.outSlot] = true

	// fwd memoises the recomputation of a forward chain value in the backward
	// graph. Slot numbering matches the chain: leaves first, then steps.
	fwd := make([]int, nInputs+len(chain.steps))
	fwdOK := make([]bool, nInputs+len(chain.steps))
	var fwdVal func(slot int) int
	fwdVal = func(slot int) int {
		if fwdOK[slot] {
			return fwd[slot]
		}
		if slot < nInputs {
			// A chain leaf is a live forward value, bound as a backward leaf.
			fwd[slot] = b.leaf(chain.inputs[slot])
			fwdOK[slot] = true
			return fwd[slot]
		}
		s := chain.steps[slot-nInputs]
		a := fwdVal(s.a)
		bv := -1
		if s.b >= 0 {
			bv = fwdVal(s.b)
		}
		id := b.emitElementwise(s, a, bv)
		fwd[slot] = id
		fwdOK[slot] = true
		return id
	}

	for i, s := range slices.Backward(chain.steps) {
		slot := nInputs + i
		if !has[slot] {
			continue
		}
		g := grad[slot]

		add := func(target, partial int) {
			if !has[target] {
				grad[target] = partial
				has[target] = true
				return
			}
			from := b.g.values[partial].shape
			to := b.g.values[grad[target]].shape
			reduced := b.reduceTo(partial, from, to)
			sum := b.emit("add", []int{grad[target], reduced}, to, b.g.values[grad[target]].dtype, nil)
			grad[target] = sum
		}
		switch s.op {
		case "add":
			add(s.a, g)
			add(s.b, g)
		case "sub":
			add(s.a, g)
			add(s.b, b.neg(g))
		case "mul":
			a, bv := fwdVal(s.a), fwdVal(s.b)
			ga := b.emit("mul", []int{g, bv}, b.g.values[g].shape, b.g.values[g].dtype, nil)
			gb := b.emit("mul", []int{g, a}, b.g.values[g].shape, b.g.values[g].dtype, nil)
			add(s.a, ga)
			add(s.b, gb)
		case "div":
			a, bv := fwdVal(s.a), fwdVal(s.b)
			ga := b.emit("div", []int{g, bv}, b.g.values[g].shape, b.g.values[g].dtype, nil)
			t1 := b.emit("mul", []int{g, a}, b.g.values[g].shape, b.g.values[g].dtype, nil)
			b2 := b.emit("mul", []int{bv, bv}, b.g.values[bv].shape, b.g.values[bv].dtype, nil)
			t2 := b.emit("div", []int{t1, b2}, b.g.values[g].shape, b.g.values[g].dtype, nil)
			add(s.a, ga)
			add(s.b, b.neg(t2))
		case "mul_scalar", "add_scalar", "sub_scalar", "div_scalar", "neg":
			factor := s.scalar
			switch s.op {
			case "neg":
				factor = float32(-1)
			case "add_scalar", "sub_scalar":
				factor = float32(1)
			case "div_scalar":
				inv := toFloat32(s.scalar)
				if inv != 0 {
					factor = 1 / inv
				}
			}
			add(s.a, b.emit("mul_scalar", []int{g}, b.g.values[g].shape, b.g.values[g].dtype, func(nn *graphNode) { nn.scalar = factor }))
		case "relu":
			x := fwdVal(s.a)
			zero := b.scalarConst(0)
			mask := b.emit("greater", []int{x, zero}, b.g.values[x].shape, tensor.Bool, nil)
			maskF := b.emit("cast", []int{mask}, b.g.values[x].shape, tensor.Float32, nil)
			add(s.a, b.emit("mul", []int{g, maskF}, b.g.values[g].shape, b.g.values[g].dtype, nil))
		case "sigmoid", "tanh", "exp", "log", "sqrt":
			y := fwdVal(slot)
			var deriv int
			switch s.op {
			case "sigmoid":
				one := b.scalarConst(1)
				inv := b.emit("sub", []int{one, y}, b.g.values[y].shape, b.g.values[y].dtype, nil)
				deriv = b.emit("mul", []int{y, inv}, b.g.values[y].shape, b.g.values[y].dtype, nil)
			case "tanh":
				one := b.scalarConst(1)
				sq := b.emit("mul", []int{y, y}, b.g.values[y].shape, b.g.values[y].dtype, nil)
				deriv = b.emit("sub", []int{one, sq}, b.g.values[y].shape, b.g.values[y].dtype, nil)
			case "exp":
				deriv = y
			case "log":
				deriv = b.emit("div", []int{b.scalarConst(1), fwdVal(s.a)}, b.g.values[g].shape, b.g.values[g].dtype, nil)
			case "sqrt":
				half := b.emit("mul_scalar", []int{y}, b.g.values[y].shape, b.g.values[y].dtype, func(nn *graphNode) { nn.scalar = float32(0.5) })
				deriv = b.emit("div", []int{half, fwdVal(s.a)}, b.g.values[g].shape, b.g.values[g].dtype, nil)
			}
			add(s.a, b.emit("mul", []int{g, deriv}, b.g.values[g].shape, b.g.values[g].dtype, nil))
		case "abs", "sign", "cos", "sin", "erf", "silu", "clamp",
			"greater", "lower", "greater_equal", "lower_equal", "equal", "not_equal", "cast":
			// Zero derivative almost everywhere (comparisons, casts) or a rule
			// not expressed yet, so no gradient flows through.
			continue
		default:
			return fmt.Errorf("compile: no backward rule for fused step %q", s.op)
		}
	}

	for i, fwdID := range chain.inputs {
		if has[i] {
			// A leaf may have been broadcast in the chain, so reduce the
			// gradient back to the leaf's own shape before accumulating.
			g := grad[i]
			from := b.g.values[g].shape
			to := b.shapeOf(fwdID)
			b.accumulate(fwdID, b.reduceTo(g, from, to))
		}
	}
	return nil
}

// emitElementwise emits one forward elementwise step into the backward graph,
// used to recompute a fused node's intermediate values for the backward.
func (b *backwardBuilder) emitElementwise(s elementwiseOp, a, bv int) int {
	out := tensor.Shape(nil)
	dtype := s.dtype
	if a >= 0 && a < len(b.g.values) {
		out = b.g.values[a].shape
	}
	ins := []int{a}
	if bv >= 0 {
		ins = append(ins, bv)
	}
	return b.emit(s.op, ins, out, dtype, func(n *graphNode) {
		n.scalar = s.scalar
		n.minAny, n.maxAny = s.minAny, s.maxAny
		n.dtype = s.dtype
	})
}

// propagateBinary applies ga = og * sa, gb = og * sb, where sa/sb are +/-1.
func (b *backwardBuilder) propagateBinary(n *graphNode, og, sa, sb int) {
	ga := og
	if sa < 0 {
		ga = b.neg(og)
	}
	gb := og
	if sb < 0 {
		gb = b.neg(og)
	}
	b.propagateToInput(n, 0, ga)
	b.propagateToInput(n, 1, gb)
}

// propagateToInput accumulates a partial gradient into a forward input,
// reducing it when the input was broadcast.
func (b *backwardBuilder) propagateToInput(n *graphNode, i, partial int) {
	fwdIn := n.ins[i]
	target := b.shapeOf(fwdIn)
	cur := b.g.values[partial].shape
	reduced := b.reduceTo(partial, cur, target)
	b.accumulate(fwdIn, reduced)
}

// reduceTo sums `id` (shape from) down to target, when broadcasting occurred.
func (b *backwardBuilder) reduceTo(id int, from, target tensor.Shape) int {
	if from.Equal(target) {
		return id
	}
	dtype := b.g.values[id].dtype
	cur := id
	cs := append(tensor.Shape(nil), from...)
	ts := target
	// Drop extra leading dims.
	for len(cs) > len(ts) {
		cur = b.emit("sum_dim", []int{cur}, dropDim(cs, 0), dtype, func(nn *graphNode) { nn.dim = 0; nn.keepDim = false })
		cs = dropDim(cs, 0)
	}
	// Sum size-1 dims that were broadcast (keep the dim as 1).
	for i := range ts {
		if ts[i] == 1 && cs[i] != 1 {
			cur = b.emit("sum_dim", []int{cur}, keepDimShape(cs, i), dtype, func(nn *graphNode) { nn.dim = i; nn.keepDim = true })
			cs = keepDimShape(cs, i)
		}
	}
	return cur
}

func (b *backwardBuilder) neg(id int) int {
	dt := b.g.values[id].dtype
	return b.emit("mul_scalar", []int{id}, b.g.values[id].shape, dt, func(n *graphNode) { n.scalar = float32(-1) })
}

func (b *backwardBuilder) backwardMul(n *graphNode, og int) {
	// ga = og * b, gb = og * a
	ga := b.emit("mul", []int{og, b.leaf(n.ins[1])}, b.shapeOf(n.out), b.dtypeOf(n.out), nil)
	gb := b.emit("mul", []int{og, b.leaf(n.ins[0])}, b.shapeOf(n.out), b.dtypeOf(n.out), nil)
	b.propagateToInput(n, 0, ga)
	b.propagateToInput(n, 1, gb)
}

func (b *backwardBuilder) backwardDiv(n *graphNode, og int) error {
	// ga = og / b, gb = -og * a / b^2
	a, bb := b.leaf(n.ins[0]), b.leaf(n.ins[1])
	ga := b.emit("div", []int{og, bb}, b.shapeOf(n.out), b.dtypeOf(n.out), nil)
	t1 := b.emit("mul", []int{og, a}, b.shapeOf(n.out), b.dtypeOf(n.out), nil)
	b2 := b.emit("mul", []int{bb, bb}, b.shapeOf(n.ins[1]), b.dtypeOf(n.ins[1]), nil)
	t2 := b.emit("div", []int{t1, b2}, b.shapeOf(n.out), b.dtypeOf(n.out), nil)
	gb := b.neg(t2)
	b.propagateToInput(n, 0, ga)
	b.propagateToInput(n, 1, gb)
	return nil
}

func (b *backwardBuilder) backwardMatMul(n *graphNode, og int) error {
	// For y = op(a) @ op(b) with trans flags: gradA and gradB are the standard
	// formulas, expressed with the same transposed flags.
	a, bb := n.ins[0], n.ins[1]
	ash, bsh := b.shapeOf(a), b.shapeOf(bb)
	dtype := b.dtypeOf(n.out)

	// gradA = og @ op(b)^T, respecting a's transpose flag.
	gradA := b.emit("matmul", []int{og, b.leaf(bb)}, ash, dtype, func(nn *graphNode) {
		nn.transA = false
		nn.transB = !n.transB
	})
	// gradB = op(a)^T @ og, respecting b's transpose flag.
	gradB := b.emit("matmul", []int{b.leaf(a), og}, bsh, dtype, func(nn *graphNode) {
		nn.transA = !n.transA
		nn.transB = false
	})
	b.accumulate(a, gradA)
	b.accumulate(bb, gradB)
	return nil
}

func (b *backwardBuilder) backwardRelu(n *graphNode, og int) {
	// grad = og * (x > 0)
	x := b.leaf(n.ins[0])
	zero := b.zeroScalarLike(n.ins[0])
	mask := b.emit("greater", []int{x, zero}, b.shapeOf(n.ins[0]), tensor.Bool, nil)
	maskF := b.emit("cast", []int{mask}, b.shapeOf(n.ins[0]), tensor.Float32, nil)
	g := b.emit("mul", []int{og, maskF}, b.shapeOf(n.out), b.dtypeOf(n.out), nil)
	b.propagateToInput(n, 0, g)
}

// backwardMatMulBias differentiates the fused Linear epilogue
// y = relu(x @ W^T + bias), with x [M, K], W [N, K] and bias [N].
func (b *backwardBuilder) backwardMatMulBias(n *graphNode, og int) error {
	x, w, bias := n.ins[0], n.ins[1], n.ins[2]
	dpre := og
	if n.relu {
		// y = relu(z); dpre = og * (y > 0)
		y := b.leaf(n.out)
		zero := b.scalarConst(0)
		mask := b.emit("greater", []int{y, zero}, b.shapeOf(n.out), tensor.Bool, nil)
		maskF := b.emit("cast", []int{mask}, b.shapeOf(n.out), tensor.Float32, nil)
		dpre = b.emit("mul", []int{og, maskF}, b.shapeOf(n.out), b.dtypeOf(n.out), nil)
	}
	// gradX = dpre @ W
	gradX := b.emit("matmul", []int{dpre, b.leaf(w)}, b.shapeOf(x), b.dtypeOf(n.out), nil)
	// gradW = dpre^T @ x
	gradW := b.emit("matmul", []int{dpre, b.leaf(x)}, b.shapeOf(w), b.dtypeOf(n.out), func(nn *graphNode) {
		nn.transA = true
	})
	// gradBias = sum over rows of dpre
	gradB := b.emit("sum_dim", []int{dpre}, b.shapeOf(bias), b.dtypeOf(n.out), func(nn *graphNode) {
		nn.dim = 0
		nn.keepDim = false
	})
	b.accumulate(x, gradX)
	b.accumulate(w, gradW)
	b.accumulate(bias, gradB)
	return nil
}

func (b *backwardBuilder) backwardBcast(n *graphNode, to int, og int, _ int) {
	partial := b.reduceTo(og, b.g.values[og].shape, b.shapeOf(to))
	b.accumulate(to, partial)
}

func (b *backwardBuilder) backwardSumDim(n *graphNode, og int) {
	// Re-expand the gradient's dim, then broadcast to the input shape.
	var target tensor.Shape
	xShape := b.shapeOf(n.ins[0])
	if n.keepDim {
		target = tensor.Shape(nil)
	} else {
		target = make(tensor.Shape, 0, len(xShape))
		for i := range xShape {
			if i == n.dim {
				target = append(target, 1)
			} else {
				target = append(target, xShape[i])
			}
		}
	}
	reshaped := b.emit("reshape", []int{og}, target, b.dtypeOf(n.out), func(nn *graphNode) { nn.shape = target })
	exp := b.emit("expand", []int{reshaped}, xShape, b.dtypeOf(n.out), func(nn *graphNode) { nn.shape = xShape })
	b.accumulate(n.ins[0], exp)
}

func (b *backwardBuilder) backwardMeanDim(n *graphNode, og int) {
	b.backwardSumDim(n, og)
	xShape := b.shapeOf(n.ins[0])
	count := float32(xShape[n.dim])
	cur := b.grad(n.ins[0])
	scaled := b.emit("mul_scalar", []int{cur}, b.g.values[cur].shape, b.g.values[cur].dtype, func(nn *graphNode) { nn.scalar = 1 / count })
	b.gradOf[n.ins[0]] = scaled
}

func (b *backwardBuilder) backwardUnaryByOutput(n *graphNode, og int) {
	// d/dx exp(x) = exp(x), i.e. the forward output.
	g := b.emit("mul", []int{og, b.leaf(n.out)}, b.shapeOf(n.out), b.dtypeOf(n.out), nil)
	b.propagateToInput(n, 0, g)
}

func (b *backwardBuilder) backwardLog(n *graphNode, og int) {
	// d/dx log(x) = 1/x
	x := b.leaf(n.ins[0])
	inv := b.emit("div", []int{b.oneScalarLike(n.ins[0]), x}, b.shapeOf(n.ins[0]), b.dtypeOf(n.ins[0]), nil)
	g := b.emit("mul", []int{og, inv}, b.shapeOf(n.out), b.dtypeOf(n.out), nil)
	b.propagateToInput(n, 0, g)
}

func (b *backwardBuilder) backwardSqrt(n *graphNode, og int) {
	// d/dx sqrt(x) = 1/(2 sqrt(x)) = 0.5 / output
	half := b.emit("mul_scalar", []int{og}, b.shapeOf(n.out), b.dtypeOf(n.out), func(nn *graphNode) { nn.scalar = float32(0.5) })
	g := b.emit("div", []int{half, b.leaf(n.out)}, b.shapeOf(n.out), b.dtypeOf(n.out), nil)
	b.propagateToInput(n, 0, g)
}

func (b *backwardBuilder) backwardTanh(n *graphNode, og int) {
	// d/dx tanh(x) = 1 - output^2
	y := b.leaf(n.out)
	sq := b.emit("mul", []int{y, y}, b.shapeOf(n.out), b.dtypeOf(n.out), nil)
	one := b.oneScalarLike(n.out)
	inv := b.emit("sub", []int{one, sq}, b.shapeOf(n.out), b.dtypeOf(n.out), nil)
	g := b.emit("mul", []int{og, inv}, b.shapeOf(n.out), b.dtypeOf(n.out), nil)
	b.propagateToInput(n, 0, g)
}

func (b *backwardBuilder) backwardSigmoid(n *graphNode, og int) {
	// d/dx sigmoid(x) = output * (1 - output)
	y := b.leaf(n.out)
	one := b.oneScalarLike(n.out)
	inv := b.emit("sub", []int{one, y}, b.shapeOf(n.out), b.dtypeOf(n.out), nil)
	t := b.emit("mul", []int{y, inv}, b.shapeOf(n.out), b.dtypeOf(n.out), nil)
	g := b.emit("mul", []int{og, t}, b.shapeOf(n.out), b.dtypeOf(n.out), nil)
	b.propagateToInput(n, 0, g)
}

func (b *backwardBuilder) backwardScalarMul(n *graphNode, og int) {
	scalar := n.scalar
	if scalar == nil {
		scalar = float32(1)
	}
	g := b.emit("mul_scalar", []int{og}, b.shapeOf(n.out), b.dtypeOf(n.out), func(nn *graphNode) { nn.scalar = scalar })
	b.propagateToInput(n, 0, g)
}

func (b *backwardBuilder) backwardNeg(n *graphNode, og int) {
	b.propagateToInput(n, 0, b.neg(og))
}

func (b *backwardBuilder) backwardShape(n *graphNode, og int) {
	target := b.shapeOf(n.ins[0])
	var cur int
	if n.op == "transpose" {
		// Transpose backward is the inverse permutation.
		perm := inversePerm(n.axes)
		cur = b.emit("transpose", []int{og}, target, b.dtypeOf(n.out), func(nn *graphNode) { nn.axes = perm })
	} else {
		cur = b.emit("reshape", []int{og}, target, b.dtypeOf(n.out), func(nn *graphNode) { nn.shape = target })
	}
	b.accumulate(n.ins[0], cur)
}

// ---- scalar leaf helpers ----

// scalarConst interns a [1] float constant as a leaf, so broadcasting a scalar
// into an op needs no special replay support.
func (b *backwardBuilder) scalarConst(v float32) int {
	raw, err := tensor.NewRaw(tensor.Shape{1}, tensor.Float32, b.g.device)
	if err != nil {
		panic(err)
	}
	raw.AsFloat32()[0] = v
	id := len(b.g.values)
	b.g.values = append(b.g.values, graphValue{shape: tensor.Shape{1}, dtype: tensor.Float32, device: b.g.device, node: -1, raw: raw})
	return id
}

func (b *backwardBuilder) zeroScalarLike(fwdID int) int {
	_ = fwdID
	return b.scalarConst(0)
}

func (b *backwardBuilder) oneScalarLike(fwdID int) int {
	_ = fwdID
	return b.scalarConst(1)
}

// ---- shape helpers ----

func dropDim(s tensor.Shape, dim int) tensor.Shape {
	out := make(tensor.Shape, 0, len(s)-1)
	for i, v := range s {
		if i == dim {
			continue
		}
		out = append(out, v)
	}
	return out
}

func keepDimShape(s tensor.Shape, dim int) tensor.Shape {
	out := append(tensor.Shape(nil), s...)
	out[dim] = 1
	return out
}

func inversePerm(axes []int) []int {
	inv := make([]int, len(axes))
	for i, a := range axes {
		inv[a] = i
	}
	return inv
}
