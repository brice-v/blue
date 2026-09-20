package ml

import (
	"fmt"

	"blue/borncgo/tensor"
)

// This file replays and optimises captured graphs (see graph.go for capture).

// matMulTransposedBackend is implemented by backends that multiply with
// transposed operands without materialising the transpose.
type matMulTransposedBackend interface {
	MatMulTransposed(a, b *tensor.RawTensor, transA, transB bool) *tensor.RawTensor
}

// Optimize runs the compilation passes: fuse the Linear epilogue
// (matmul + bias + optional relu), fold elementwise chains, deduplicate
// identical nodes, then drop nodes whose result is not needed.
func (g *Graph) Optimize() {
	g.fused = g.fuseMatMulBias()
	g.fused += g.fuseMatMulRelu()
	g.fused += g.fuseElementwise()
	g.cse()
	g.eliminateDead()
}

// Stats reports the shape of the compiled graph.
func (g *Graph) Stats() string {
	live := 0
	for i := range g.nodes {
		if !g.nodes[i].dead {
			live++
		}
	}
	return fmt.Sprintf("nodes=%d fused=%d", live, g.fused)
}

// Run replays the graph on a real backend. The tape of a differentiating
// backend records the replayed ops, so backward runs exactly as it would have
// eagerly.
func (g *Graph) Run(be tensor.Backend, input *tensor.RawTensor) (*tensor.RawTensor, error) {
	// Shapes were baked in at capture time. Guard them rather than silently
	// reusing a graph compiled for a different batch size.
	in := g.values[g.input]
	if !input.Shape().Equal(in.shape) {
		return nil, fmt.Errorf("compile: graph captured for input shape %v, got %v", in.shape, input.Shape())
	}
	return g.RunWithLeaves(be, map[int]*tensor.RawTensor{g.input: input})
}

// RunWithLeaves replays the graph with explicit bindings for its input slots.
// A backward graph has many leaves (the forward activations, parameters and the
// gradient seed), so they are supplied as a slot map rather than one input.
func (g *Graph) RunWithLeaves(be tensor.Backend, leaves map[int]*tensor.RawTensor) (*tensor.RawTensor, error) {
	return g.RunWithLeavesOutput(be, leaves, g.output)
}

// RunWithLeavesOutput replays the graph and returns the tensor at `slot`.
func (g *Graph) RunWithLeavesOutput(be tensor.Backend, leaves map[int]*tensor.RawTensor, slot int) (*tensor.RawTensor, error) {
	slots, err := g.replaySlots(be, leaves)
	if err != nil {
		return nil, err
	}
	out := slots[slot]
	if out == nil {
		return nil, fmt.Errorf("compile: graph value %d was not produced", slot)
	}
	return out, nil
}

// replaySlots runs the graph and returns every value slot, so callers (like a
// backward replay) can read activations as well as the output.
func (g *Graph) replaySlots(be tensor.Backend, leaves map[int]*tensor.RawTensor) ([]*tensor.RawTensor, error) {
	slots := make([]*tensor.RawTensor, len(g.values))
	for i := range g.values {
		v := &g.values[i]
		switch {
		case v.input:
			slots[i] = leaves[i]
		case v.param != nil:
			slots[i] = v.param.Tensor().Raw()
		case v.raw != nil:
			slots[i] = v.raw
		}
	}
	for i, r := range leaves {
		if i >= 0 && i < len(slots) {
			slots[i] = r
		}
	}
	for i := range g.nodes {
		n := &g.nodes[i]
		if n.dead {
			continue
		}
		for _, in := range n.ins {
			if in < 0 || in >= len(slots) {
				return nil, fmt.Errorf("compile: op %q reads out-of-range value %d", n.op, in)
			}
			if slots[in] == nil {
				return nil, fmt.Errorf("compile: op %q reads unbound value %d", n.op, in)
			}
		}
		if err := g.runNode(be, n, slots); err != nil {
			return nil, err
		}
	}
	return slots, nil
}

func (g *Graph) runNode(be tensor.Backend, n *graphNode, slots []*tensor.RawTensor) error {
	in := func(i int) *tensor.RawTensor {
		if i < 0 || i >= len(n.ins) {
			return nil
		}
		return slots[n.ins[i]]
	}
	var out *tensor.RawTensor
	switch n.op {
	case "add":
		out = be.Add(in(0), in(1))
	case "sub":
		out = be.Sub(in(0), in(1))
	case "mul":
		out = be.Mul(in(0), in(1))
	case "div":
		out = be.Div(in(0), in(1))
	case "or":
		out = be.Or(in(0), in(1))
	case "and":
		out = be.And(in(0), in(1))
	case "not":
		out = be.Not(in(0))
	case "matmul":
		out = g.runMatMul(be, n, in)
	case "batch_matmul":
		out = be.BatchMatMul(in(0), in(1))
	case "matmul_bias":
		out = runMatMulBias(be, in(0), in(1), in(2), n.relu)
	case "reshape":
		out = be.Reshape(in(0), n.shape)
	case "transpose":
		out = be.Transpose(in(0), n.axes...)
	case "unsqueeze":
		out = be.Unsqueeze(in(0), n.dim)
	case "squeeze":
		out = be.Squeeze(in(0), n.dim)
	case "expand":
		out = be.Expand(in(0), n.shape)
	case "cat":
		parts := make([]*tensor.RawTensor, len(n.ins))
		for i := range n.ins {
			parts[i] = in(i)
		}
		out = be.Cat(parts, n.dim)
	case "gather":
		out = be.Gather(in(0), n.dim, in(1))
	case "where":
		out = be.Where(in(0), in(1), in(2))
	case "embedding":
		out = be.Embedding(in(0), in(1))
	case "mul_scalar":
		out = be.MulScalar(in(0), n.scalar)
	case "add_scalar":
		out = be.AddScalar(in(0), n.scalar)
	case "sub_scalar":
		out = be.SubScalar(in(0), n.scalar)
	case "div_scalar":
		out = be.DivScalar(in(0), n.scalar)
	case "exp":
		out = be.Exp(in(0))
	case "log":
		out = be.Log(in(0))
	case "sqrt":
		out = be.Sqrt(in(0))
	case "rsqrt":
		out = be.Rsqrt(in(0))
	case "cos":
		out = be.Cos(in(0))
	case "sin":
		out = be.Sin(in(0))
	case "erf":
		out = be.Erf(in(0))
	case "sign":
		out = be.Sign(in(0))
	case "abs":
		out = be.Abs(in(0))
	case "clamp":
		out = be.Clamp(in(0), n.minAny, n.maxAny)
	case "relu":
		out = be.ReLU(in(0))
	case "sigmoid":
		out = be.Sigmoid(in(0))
	case "tanh":
		out = be.Tanh(in(0))
	case "silu":
		out = be.SiLU(in(0))
	case "softmax":
		out = be.Softmax(in(0), n.dim)
	case "sum":
		out = be.Sum(in(0))
	case "sum_dim":
		out = be.SumDim(in(0), n.dim, n.keepDim)
	case "mean_dim":
		out = be.MeanDim(in(0), n.dim, n.keepDim)
	case "argmax":
		out = be.Argmax(in(0), n.dim)
	case "greater":
		out = be.Greater(in(0), in(1))
	case "lower":
		out = be.Lower(in(0), in(1))
	case "greater_equal":
		out = be.GreaterEqual(in(0), in(1))
	case "lower_equal":
		out = be.LowerEqual(in(0), in(1))
	case "equal":
		out = be.Equal(in(0), in(1))
	case "not_equal":
		out = be.NotEqual(in(0), in(1))
	case "cast":
		out = be.Cast(in(0), n.dtype)
	case "elementwise":
		out = runElementwise(be, n, slots)
	default:
		return fmt.Errorf("compile: cannot replay unknown op %q", n.op)
	}
	if out == nil {
		return fmt.Errorf("compile: op %q produced no output", n.op)
	}
	slots[n.out] = out
	return nil
}

func (g *Graph) runMatMul(be tensor.Backend, n *graphNode, in func(int) *tensor.RawTensor) *tensor.RawTensor {
	if !n.transA && !n.transB {
		return be.MatMul(in(0), in(1))
	}
	if mt, ok := be.(matMulTransposedBackend); ok {
		return mt.MatMulTransposed(in(0), in(1), n.transA, n.transB)
	}
	x, y := in(0), in(1)
	if n.transA {
		x = be.Transpose(x, 1, 0)
	}
	if n.transB {
		y = be.Transpose(y, 1, 0)
	}
	return be.MatMul(x, y)
}

func runMatMulBias(be tensor.Backend, x, w, bias *tensor.RawTensor, relu bool) *tensor.RawTensor {
	if mb, ok := be.(matMulBiasBackend); ok {
		return mb.MatMulBias(x, w, bias, relu)
	}
	wT := be.Transpose(w, 1, 0)
	out := be.MatMul(x, wT)
	bias2 := be.Reshape(bias, tensor.Shape{1, w.Shape()[0]})
	out = be.Add(out, bias2)
	if relu {
		out = be.ReLU(out)
	}
	return out
}

// ---- optimisation passes ----

// consumers maps each value id to the nodes that read it.
func (g *Graph) consumers() [][]int {
	consumers := make([][]int, len(g.values))
	for i := range g.nodes {
		if g.nodes[i].dead {
			continue
		}
		for _, in := range g.nodes[i].ins {
			consumers[in] = append(consumers[in], i)
		}
	}
	return consumers
}

// fuseMatMulBias rewrites the eager Linear pattern
// transpose(w) -> matmul(x, wT) -> add(bias) [-> relu]
// into a single fused matmul_bias node.
func (g *Graph) fuseMatMulBias() int {
	consumers := g.consumers()
	fused := 0
	for i := range g.nodes {
		n := &g.nodes[i]
		if n.dead || n.op != "matmul" || len(n.ins) != 2 {
			continue
		}
		if len(consumers[n.out]) != 1 {
			continue
		}
		a := &g.nodes[consumers[n.out][0]]
		if a.dead || a.op != "add" {
			continue
		}
		var biasID int
		switch {
		case a.ins[0] == n.out:
			biasID = a.ins[1]
		case a.ins[1] == n.out:
			biasID = a.ins[0]
		default:
			continue
		}
		outShape := g.values[n.out].shape
		if len(outShape) != 2 {
			continue
		}
		bias, ok := g.biasSource(biasID, outShape[1])
		if !ok {
			continue
		}
		wTID := n.ins[1]
		if len(consumers[wTID]) != 1 {
			continue
		}
		w, ok := g.transposeSource(wTID)
		if !ok {
			continue
		}

		finalID := a.out
		relu := false
		if len(consumers[a.out]) == 1 {
			r := &g.nodes[consumers[a.out][0]]
			if !r.dead && r.op == "relu" {
				relu = true
				finalID = r.out
				r.dead = true
			}
		}

		n.op = "matmul_bias"
		n.ins = []int{n.ins[0], w, bias}
		n.relu = relu
		n.out = finalID
		a.dead = true
		g.values[finalID].node = i
		fused++
	}
	return fused
}

// fuseMatMulRelu folds a ReLU that only consumes a matmul_bias into the fused
// kernel's activation flag.
func (g *Graph) fuseMatMulRelu() int {
	consumers := g.consumers()
	fused := 0
	for i := range g.nodes {
		n := &g.nodes[i]
		if n.dead || n.op != "matmul_bias" || n.relu {
			continue
		}
		if len(consumers[n.out]) != 1 {
			continue
		}
		r := &g.nodes[consumers[n.out][0]]
		if r.dead || r.op != "relu" {
			continue
		}
		n.relu = true
		n.out = r.out
		r.dead = true
		g.values[r.out].node = i
		fused++
	}
	return fused
}

// biasSource returns the id of a [N] bias value, looking through the [1, N]
// reshape nn.Linear inserts for broadcasting.
func (g *Graph) biasSource(id, n int) (int, bool) {
	v := &g.values[id]
	if (v.raw != nil || v.param != nil) && len(v.shape) == 1 && v.shape[0] == n {
		return id, true
	}
	if v.node >= 0 {
		nd := &g.nodes[v.node]
		if !nd.dead && nd.op == "reshape" && len(nd.ins) == 1 {
			iv := &g.values[nd.ins[0]]
			if (iv.raw != nil || iv.param != nil) && len(iv.shape) == 1 && iv.shape[0] == n {
				return nd.ins[0], true
			}
		}
	}
	return 0, false
}

// transposeSource returns the input of a transpose node, when that input is a
// bound constant or parameter (a weight).
func (g *Graph) transposeSource(id int) (int, bool) {
	v := &g.values[id]
	if v.node < 0 {
		return 0, false
	}
	n := &g.nodes[v.node]
	if n.dead || n.op != "transpose" || len(n.ins) != 1 {
		return 0, false
	}
	iv := &g.values[n.ins[0]]
	if iv.raw == nil && iv.param == nil {
		return 0, false
	}
	return n.ins[0], true
}

// cse replaces identical nodes with a single one.
func (g *Graph) cse() {
	seen := map[string]int{}
	remap := map[int]int{}
	for i := range g.nodes {
		n := &g.nodes[i]
		if n.dead {
			continue
		}
		for j, in := range n.ins {
			if r, ok := remap[in]; ok {
				n.ins[j] = r
			}
		}
		if prev, ok := seen[n.key()]; ok {
			remap[n.out] = prev
			n.dead = true
			continue
		}
		seen[n.key()] = n.out
	}
	if r, ok := remap[g.output]; ok {
		g.output = r
	}
	for i, id := range g.outputs {
		if r, ok := remap[id]; ok {
			g.outputs[i] = r
		}
	}
}

// eliminateDead prunes nodes whose result does not reach an output.
func (g *Graph) eliminateDead() {
	used := make([]bool, len(g.values))
	var mark func(int)
	mark = func(id int) {
		if id < 0 || id >= len(used) || used[id] {
			return
		}
		used[id] = true
		v := &g.values[id]
		if v.node >= 0 {
			nd := &g.nodes[v.node]
			if !nd.dead {
				for _, in := range nd.ins {
					mark(in)
				}
			}
		}
	}
	mark(g.output)
	for _, id := range g.outputs {
		mark(id)
	}
	for i := range g.nodes {
		if !g.nodes[i].dead && !used[g.nodes[i].out] {
			g.nodes[i].dead = true
		}
	}
}
