package ml

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"blue/borncgo/tensor"
)

// This file holds the plan cache behind ml.compile: a cache of optimised graphs
// keyed on the input's device/dtype/shape, so a model can be called with many
// shapes and recompiles on a guard miss. torch.compile behaves this way: a
// graph is specialised for the shapes it was traced with, and a shape change
// triggers a new capture instead of an error.

// Compiled is the plan cache behind ml.compile. Blue drives capture (it owns
// the model closure) and installs the optimised graph here.
type Compiled struct {
	mu    sync.Mutex
	plans map[string]*Graph
	// backwards holds the differentiated form of each plan, built lazily the
	// first time a compiled backward runs for that signature.
	backwards map[string]*backwardPlan
	order     []string // insertion order, for a bounded cache

	compiles       int
	backwardsBuilt int
}

// NewCompiled builds an empty compiled-plan cache.
func NewCompiled() *Compiled {
	return &Compiled{plans: map[string]*Graph{}, backwards: map[string]*backwardPlan{}}
}

// BackwardPlan returns the differentiated form of the plan for x, building it
// on first use. It returns nil when x has no plan or the forward graph contains
// an op with no backward rule.
func (c *Compiled) BackwardPlan(x *Tensor) *backwardPlan {
	sig := signature(x)
	c.mu.Lock()
	if bp, ok := c.backwards[sig]; ok {
		c.mu.Unlock()
		return bp
	}
	g := c.plans[sig]
	c.mu.Unlock()
	if g == nil {
		return nil
	}
	bp, err := BuildBackward(g)
	if err != nil {
		return nil
	}
	// Give every gradient value a root so dead-code elimination keeps it.
	for _, id := range bp.paramGrads {
		bp.graph.outputs = append(bp.graph.outputs, id)
	}
	bp.graph.Optimize()

	c.mu.Lock()
	if existing, ok := c.backwards[sig]; ok {
		c.mu.Unlock()
		return existing
	}
	c.backwards[sig] = bp
	c.backwardsBuilt++
	c.mu.Unlock()
	return bp
}

// signature is the guard key: device, dtype and shape of the input.
func signature(t *Tensor) string {
	var b strings.Builder
	b.WriteString(t.Device().String())
	b.WriteByte('/')
	b.WriteString(t.DType().String())
	b.WriteByte('/')
	for i, s := range t.Shape() {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.Itoa(s))
	}
	return b.String()
}

// Plan returns the graph for x's signature, or nil on a guard miss.
func (c *Compiled) Plan(x *Tensor) *Graph {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.plans[signature(x)]
}

// Install records a newly captured graph for the input that produced it. A
// bounded number of plans is kept, so a model called with many distinct shapes
// evicts the oldest plan rather than growing without limit.
func (c *Compiled) Install(x *Tensor, g *Graph) {
	c.mu.Lock()
	defer c.mu.Unlock()
	sig := signature(x)
	if _, ok := c.plans[sig]; ok {
		return
	}
	c.compiles++
	c.plans[sig] = g
	c.order = append(c.order, sig)
	const maxPlans = 16
	if len(c.order) > maxPlans {
		old := c.order[0]
		c.order = c.order[1:]
		delete(c.plans, old)
	}
}

// Run executes the plan for x, erroring on a guard miss (the caller should
// capture and Install first).
func (c *Compiled) Run(x *Tensor) (*Tensor, error) {
	g := c.Plan(x)
	if g == nil {
		return nil, fmt.Errorf("compile: no plan for input %s", signature(x))
	}
	raw, err := g.Run(x.Backend(), x.Raw())
	if err != nil {
		return nil, err
	}
	return wrapRaw(x.Backend(), raw), nil
}

// Stats reports plan count and number of compilations.
func (c *Compiled) Stats() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return fmt.Sprintf("plans=%d compiles=%d backwards=%d", len(c.plans), c.compiles, c.backwardsBuilt)
}

// Backward runs the compiled backward for x, storing the resulting parameter
// gradients where borncgo's optimizer expects them (lastGrads) and on each
// tracked tensor's .grad.
//
// The seed is the gradient of the loss with respect to the forward output. A
// nil seed means "gradient of the sum of the output", matching a scalar loss
// over the output. Callers with a real loss (cross-entropy, MSE) pass its
// gradient so the compiled backward matches the training objective.
func (c *Compiled) Backward(x *Tensor) error {
	return c.BackwardWithSeed(x, nil)
}

// BackwardWithSeed is Backward with an explicit output-gradient seed.
func (c *Compiled) BackwardWithSeed(x *Tensor, seed *Tensor) error {
	bp := c.BackwardPlan(x)
	if bp == nil {
		return fmt.Errorf("compile: no backward plan for input %s", signature(x))
	}
	fg := c.Plan(x)
	if fg == nil {
		return fmt.Errorf("compile: no plan for input %s", signature(x))
	}
	be := x.Backend()

	// Replay the forward to materialise the activations the backward leaves
	// stand for.
	slots, err := fg.replaySlots(be, map[int]*tensor.RawTensor{fg.input: x.Raw()})
	if err != nil {
		return err
	}
	leaves := make(map[int]*tensor.RawTensor, len(bp.leaves)+1)
	for fwdID, slot := range bp.leaves {
		leaves[slot] = slots[fwdID]
	}
	if seed != nil {
		if !tensor.Shape(seed.Shape()).Equal(fg.values[fg.output].shape) {
			return fmt.Errorf("compile: seed shape %v does not match output shape %v", seed.Shape(), fg.values[fg.output].shape)
		}
		leaves[bp.seed] = seed.Raw()
	} else {
		leaves[bp.seed] = onesShapeRaw(fg.values[fg.output].shape, be)
	}

	grads := make(map[*tensor.RawTensor]*tensor.RawTensor, len(bp.paramGrads))
	for fwdID, slot := range bp.paramGrads {
		g, err := bp.graph.RunWithLeavesOutput(be, leaves, slot)
		if err != nil {
			return err
		}
		// Key by the tensor the optimizer knows: a parameter's raw, or a raw
		// leaf's own value.
		v := &fg.values[fwdID]
		switch {
		case v.param != nil:
			grads[v.param.Tensor().Raw()] = g
		case v.raw != nil:
			grads[v.raw] = g
		}
	}

	// Store the gradients where the optimizer reads them, then hand each
	// tracked tensor its .grad, matching Tensor.Backward.
	setLastGrads(be, grads)
	for _, tt := range liveLeaves() {
		if !tt.requiresGrad {
			continue
		}
		if g, ok := grads[tt.t.Raw()]; ok && g != nil {
			tt.grad = wrapRaw(be, g)
		}
	}
	return nil
}

// onesShapeRaw returns a ones tensor of the given shape and device, used as the
// backward seed. The shape must be preserved: a flattened seed would make the
// first matmul see a 1D operand.
func onesShapeRaw(shape tensor.Shape, be tensor.Backend) *tensor.RawTensor {
	s := shape
	if s.NumElements() == 0 {
		s = tensor.Shape{1}
	}
	raw, err := tensor.NewRaw(s, tensor.Float32, be.Device())
	if err != nil {
		panic(err)
	}
	for i := range raw.AsFloat32() {
		raw.AsFloat32()[i] = 1
	}
	return raw
}
