package ml

import (
	"fmt"
	"sort"
	"strings"

	"blue/borncgo/tensor"
)

// Elementwise chain fusion. A run of elementwise ops (add, sub, mul, div, the
// unary math ops and the activations) whose intermediates are not needed
// elsewhere is folded into one `elementwise` node. The node records the chain
// and its leaves, so it can run as a single kernel instead of one dispatch and
// one full-tensor round trip per op. This is the generic win: it applies to any
// model, not just the Linear pattern.
//
// Correctness never depends on the fusion: when the backend has a fused kernel
// the chain runs in one pass, otherwise the same chain runs as the original
// per-op sequence.

// elementwiseOp is one step in a fused elementwise chain.
type elementwiseOp struct {
	op     string
	a, b   int // operands into the chain's value list; b is -1 for unary
	scalar any
	minAny any
	maxAny any
	shape  tensor.Shape
	dtype  tensor.DataType
}

// elementwiseChain is the payload of an `elementwise` node. The chain's value
// list is the inputs followed by the steps, so step i has slot
// len(inputs)+i.
type elementwiseChain struct {
	inputs  []int // graph value ids of the leaves, by chain-local index
	steps   []elementwiseOp
	outSlot int
}

// elementwiseOps reports whether an op is fusible elementwise. Comparisons are
// included because their result, stored as 0/1, commonly feeds a multiply (the
// relu backward mask is the motivating case); a bool intermediate in a fused
// chain is kept as a float 0/1.
func elementwiseOps(op string) bool {
	switch op {
	case "add", "sub", "mul", "div", "neg",
		"exp", "log", "sqrt", "rsqrt", "cos", "sin", "erf", "sign", "abs",
		"clamp", "relu", "sigmoid", "tanh", "silu",
		"mul_scalar", "add_scalar", "sub_scalar", "div_scalar",
		"greater", "lower", "greater_equal", "lower_equal", "equal", "not_equal",
		"cast":
		return true
	default:
		return false
	}
}

// fuseElementwise folds maximal elementwise chains into single nodes.
func (g *Graph) fuseElementwise() int {
	consumers := g.consumers()
	visited := make([]bool, len(g.nodes))
	fused := 0

	// Iterate to a fixed point: folding a chain can expose another one.
	for pass := 0; pass < 8; pass++ {
		progress := false
		for i := range g.nodes {
			n := &g.nodes[i]
			if n.dead || visited[i] || !elementwiseOps(n.op) {
				continue
			}
			// Start only at a chain root: a node whose output is consumed by
			// something that is not elementwise (or is the graph output).
			// Walking backwards from a mid-chain node would miss its consumers.
			if g.hasElementwiseConsumer(i, consumers) {
				continue
			}
			chain, members, ok := g.buildChain(i, consumers)
			if !ok || len(members) < 2 {
				continue
			}
			g.replaceWithElementwise(i, chain, members)
			for _, m := range members {
				visited[m] = true
			}
			fused += len(members) - 1
			progress = true
		}
		if !progress {
			break
		}
	}
	return fused
}

// hasElementwiseConsumer reports whether some consumer of node i is itself an
// elementwise op, meaning i is mid-chain and not a chain root.
func (g *Graph) hasElementwiseConsumer(i int, consumers [][]int) bool {
	n := &g.nodes[i]
	for _, c := range consumers[n.out] {
		cn := &g.nodes[c]
		if !cn.dead && elementwiseOps(cn.op) {
			return true
		}
	}
	return false
}

// buildChain walks the elementwise subgraph reachable backwards from start,
// collecting the ops that can be folded. A node joins the chain only when every
// consumer of its result is itself elementwise or is the graph output.
func (g *Graph) buildChain(start int, consumers [][]int) (*elementwiseChain, []int, bool) {
	var members []int
	seen := map[int]bool{}

	var walk func(i int) bool
	walk = func(i int) bool {
		if seen[i] {
			return true
		}
		n := &g.nodes[i]
		if n.dead || !elementwiseOps(n.op) {
			return false
		}
		if i != start {
			// An intermediate value may only feed other elementwise ops, or the
			// graph output; otherwise it has to stay materialised.
			for _, c := range consumers[n.out] {
				cn := &g.nodes[c]
				if cn.dead {
					continue
				}
				if cn.op != "elementwise" && !elementwiseOps(cn.op) && cn.out != g.output {
					return false
				}
			}
		}
		seen[i] = true
		for _, in := range n.ins {
			v := &g.values[in]
			if v.node >= 0 && !g.nodes[v.node].dead && elementwiseOps(g.nodes[v.node].op) {
				if !walk(v.node) {
					return false
				}
			}
		}
		// Inputs are appended before this node, so members is topological.
		members = append(members, i)
		return true
	}
	if !walk(start) {
		return nil, nil, false
	}
	chain, ok := g.chainFrom(members, start)
	if !ok {
		return nil, nil, false
	}
	return chain, members, true
}

// chainFrom turns member node indices into a replayable chain.
func (g *Graph) chainFrom(members []int, root int) (*elementwiseChain, bool) {
	nInputs := len(members)
	// Pass 1: note which nodes are inside the chain, so a member's output is
	// never mistaken for a leaf.
	memberNode := make(map[int]bool, nInputs)
	for _, ni := range members {
		memberNode[ni] = true
	}
	index := make(map[int]int, nInputs) // node index -> chain slot
	leaf := map[int]int{}               // graph value id -> chain input index
	chain := &elementwiseChain{}

	// Pass 2: collect the leaves (operands not produced inside the chain), in
	// first-seen order.
	for _, ni := range members {
		n := &g.nodes[ni]
		for _, in := range n.ins {
			v := &g.values[in]
			if v.node >= 0 && memberNode[v.node] {
				continue
			}
			if _, ok := leaf[in]; ok {
				continue
			}
			leaf[in] = len(chain.inputs)
			chain.inputs = append(chain.inputs, in)
		}
	}
	// Now that the leaves are known, step slots follow the leaves.
	for i, ni := range members {
		index[ni] = len(chain.inputs) + i
	}

	// Pass 3: build the steps with fully resolved operands.
	slotFor := func(vid int) int {
		v := &g.values[vid]
		if v.node >= 0 {
			if s, ok := index[v.node]; ok {
				return s
			}
		}
		return leaf[vid]
	}
	for _, ni := range members {
		n := &g.nodes[ni]
		step := elementwiseOp{op: n.op, a: slotFor(n.ins[0]), b: -1, scalar: n.scalar, minAny: n.minAny, maxAny: n.maxAny}
		if len(n.ins) > 1 {
			step.b = slotFor(n.ins[1])
		}
		step.shape = g.values[n.out].shape
		step.dtype = g.values[n.out].dtype
		chain.steps = append(chain.steps, step)
	}

	outSlot, ok := index[root]
	if !ok {
		return nil, false
	}
	chain.outSlot = outSlot
	return chain, true
}

// replaceWithElementwise swaps the member nodes for one elementwise node.
func (g *Graph) replaceWithElementwise(root int, chain *elementwiseChain, members []int) {
	rn := &g.nodes[root]
	memberOut := map[int]bool{}
	for _, mi := range members {
		if mi == root {
			continue
		}
		g.nodes[mi].dead = true
		memberOut[g.nodes[mi].out] = true
	}
	// Consumers that read any member except the root now read the fused node.
	for i := range g.nodes {
		cn := &g.nodes[i]
		if cn.dead {
			continue
		}
		for j, in := range cn.ins {
			if in != rn.out && memberOut[in] {
				cn.ins[j] = rn.out
			}
		}
	}
	rn.op = "elementwise"
	rn.ins = chain.inputs
	rn.chain = chain
	g.values[rn.out].node = root
}

// elementwiseBackend is the backend capability the fused chain needs. It is
// declared in borncgo/tensor so every backend can implement it the same way.
type elementwiseBackend = tensor.ElementwiseChainBackend

// runElementwise runs a fused chain: one kernel when the backend has one,
// otherwise the original per-op sequence, so correctness never depends on it.
func runElementwise(be tensor.Backend, n *graphNode, slots []*tensor.RawTensor) *tensor.RawTensor {
	chain := n.chain
	inputs := make([]*tensor.RawTensor, len(chain.inputs))
	for i, vid := range chain.inputs {
		inputs[i] = slots[vid]
	}
	if eb, ok := be.(elementwiseBackend); ok {
		if out := eb.ElementwiseChain(inputs, chain.backendSteps(), chain.outShape(), chain.OutDType()); out != nil {
			return out
		}
	}
	return chain.runEager(be, inputs)
}

// toFloat32 coerces a scalar operand (as stored on a graph node) to float32.
func toFloat32(v any) float32 {
	switch x := v.(type) {
	case float32:
		return x
	case float64:
		return float32(x)
	case int:
		return float32(x)
	case int32:
		return float32(x)
	case int64:
		return float32(x)
	default:
		return 0
	}
}

// backendSteps converts the chain to the backend-facing step form.
func (c *elementwiseChain) backendSteps() []tensor.ElementwiseStep {
	steps := make([]tensor.ElementwiseStep, len(c.steps))
	for i, s := range c.steps {
		steps[i] = tensor.ElementwiseStep{
			Op:     s.op,
			A:      s.a,
			B:      s.b,
			Scalar: toFloat32(s.scalar),
			DType:  s.dtype,
		}
		if v, ok := s.minAny.(float32); ok {
			steps[i].Min = v
		}
		if v, ok := s.maxAny.(float32); ok {
			steps[i].Max = v
		}
	}
	return steps
}

// outStep returns the step index backing the chain's output slot.
func (c *elementwiseChain) outStep() int { return c.outSlot - len(c.inputs) }

func (c *elementwiseChain) outShape() tensor.Shape    { return c.steps[c.outStep()].shape }
func (c *elementwiseChain) OutDType() tensor.DataType { return c.steps[c.outStep()].dtype }

// runEager executes the chain with the individual backend ops.
func (c *elementwiseChain) runEager(be tensor.Backend, inputs []*tensor.RawTensor) *tensor.RawTensor {
	vals := make([]*tensor.RawTensor, len(inputs)+len(c.steps))
	copy(vals, inputs)
	for i, s := range c.steps {
		a := vals[s.a]
		var b *tensor.RawTensor
		if s.b >= 0 {
			b = vals[s.b]
		}
		vals[len(inputs)+i] = applyElementwise(be, s, a, b)
	}
	return vals[c.outSlot]
}

func applyElementwise(be tensor.Backend, s elementwiseOp, a, b *tensor.RawTensor) *tensor.RawTensor {
	// A fused chain may carry a bool (a comparison kept as 0/1) into an
	// arithmetic op, so coerce bool operands to float32 here rather than
	// requiring every backend's arithmetic to accept bool.
	a = coerceFloat(be, a)
	b = coerceFloat(be, b)
	switch s.op {
	case "add":
		return be.Add(a, b)
	case "sub":
		return be.Sub(a, b)
	case "mul":
		return be.Mul(a, b)
	case "div":
		return be.Div(a, b)
	case "exp":
		return be.Exp(a)
	case "log":
		return be.Log(a)
	case "sqrt":
		return be.Sqrt(a)
	case "rsqrt":
		return be.Rsqrt(a)
	case "cos":
		return be.Cos(a)
	case "sin":
		return be.Sin(a)
	case "erf":
		return be.Erf(a)
	case "sign":
		return be.Sign(a)
	case "abs":
		return be.Abs(a)
	case "relu":
		return be.ReLU(a)
	case "sigmoid":
		return be.Sigmoid(a)
	case "tanh":
		return be.Tanh(a)
	case "silu":
		return be.SiLU(a)
	case "clamp":
		return be.Clamp(a, s.minAny, s.maxAny)
	case "neg":
		return be.MulScalar(a, float32(-1))
	case "greater", "lower", "greater_equal", "lower_equal", "equal", "not_equal":
		return compareElementwise(be, s.op, a, b)
	case "cast":
		return be.Cast(a, s.dtype)
	case "mul_scalar":
		return be.MulScalar(a, s.scalar)
	case "add_scalar":
		return be.AddScalar(a, s.scalar)
	case "sub_scalar":
		return be.SubScalar(a, s.scalar)
	case "div_scalar":
		return be.DivScalar(a, s.scalar)
	default:
		panic("elementwise: unknown op " + s.op)
	}
}

// coerceFloat casts a bool operand to float32 so a comparison kept as 0/1 can
// feed arithmetic inside a fused chain. Other dtypes pass through.
func coerceFloat(be tensor.Backend, x *tensor.RawTensor) *tensor.RawTensor {
	if x != nil && x.DType() == tensor.Bool {
		return be.Cast(x, tensor.Float32)
	}
	return x
}

func compareElementwise(be tensor.Backend, op string, a, b *tensor.RawTensor) *tensor.RawTensor {
	switch op {
	case "greater":
		return be.Greater(a, b)
	case "lower":
		return be.Lower(a, b)
	case "greater_equal":
		return be.GreaterEqual(a, b)
	case "lower_equal":
		return be.LowerEqual(a, b)
	case "equal":
		return be.Equal(a, b)
	default:
		return be.NotEqual(a, b)
	}
}

// SourceName is a stable identifier for the chain, used as a kernel cache key.
func (c *elementwiseChain) SourceName() string {
	var b strings.Builder
	for i, s := range c.steps {
		if i > 0 {
			b.WriteByte(';')
		}
		b.WriteString(s.op)
	}
	return b.String()
}

// Summary is used by diagnostics.
func (c *elementwiseChain) Summary() string {
	out := append([]int(nil), c.inputs...)
	sort.Ints(out)
	return fmt.Sprintf("inputs=%v steps=%d", out, len(c.steps))
}
