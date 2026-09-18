package ml

import "fmt"

// track records the graph node on out. No-op when nothing requires grad, so
// inference stays cheap.
func track(out *Tensor, op string, inputs []*Tensor, gradFn func(*Tensor) []*Tensor) {
	needs := false
	for _, in := range inputs {
		if in.RequiresGrad() {
			needs = true
			break
		}
	}
	if !needs {
		return
	}
	out.gradState = &gradState{
		requiresGrad: true,
		gradFn:       gradFn,
		prev:         inputs, // positional match with gradFn's return
		op:           op,
	}
}

func (t *Tensor) accumulate(g *Tensor) {
	if t.gradState == nil {
		t.gradState = &gradState{requiresGrad: true}
	}
	if t.gradState.Grad == nil {
		t.gradState.Grad = g // read-only alias; never mutated in place
		return
	}
	t.gradState.Grad = must(mustBackend(t).Add(t.gradState.Grad, g))
}

func (t *Tensor) ZeroGrad() {
	if t.gradState != nil {
		t.gradState.Grad = nil
	}
}

// Detach shares data and drops the graph.
func (t *Tensor) Detach() *Tensor {
	return t.Clone() // Clone already drops gradState
}

func topoOrder(root *Tensor) []*Tensor {
	var order []*Tensor
	seen := map[*Tensor]bool{}
	var visit func(t *Tensor)
	visit = func(t *Tensor) {
		if t == nil || seen[t] {
			return
		}
		seen[t] = true
		if t.gradState != nil {
			for _, p := range t.gradState.prev {
				visit(p)
			}
		}
		order = append(order, t) // post-order: children before parents
	}
	visit(root)
	return order
}

// Backward must be called on a scalar (0d) loss
func (t *Tensor) Backward() error {
	if t.gradState == nil {
		return fmt.Errorf("backward: tensor is not part of a graph")
	}
	if t.Numel() != 1 {
		return fmt.Errorf("backward: expected a scalar loss, got shape %v", t.Shape())
	}
	t.accumulate(onesLike(t)) // seed dLoss/dLoss = 1

	order := topoOrder(t)
	for i := len(order) - 1; i >= 0; i-- {
		n := order[i]
		if n.gradState == nil || n.gradState.gradFn == nil || n.gradState.Grad == nil {
			continue
		}
		grads := n.gradState.gradFn(n.gradState.Grad)
		for j, p := range n.gradState.prev {
			if p.RequiresGrad() {
				p.accumulate(grads[j])
			}
		}
	}
	return nil
}
