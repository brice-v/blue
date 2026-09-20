package ml

import (
	"testing"

	"blue/borncgo/tensor"
)

// TestBindLeafCaptureReplay checks the multi-leaf capture machinery: a graph can
// be recorded from bound leaves (as a backward capture needs), optimised, and
// replayed with fresh tensors bound to those leaves.
func TestBindLeafCaptureReplay(t *testing.T) {
	a := mustTensor(t, []float32{1, 2, 3}, []int{3})
	b := mustTensor(t, []float32{10, 20, 30}, []int{3})
	seed := mustTensor(t, []float32{1, 1, 1}, []int{3})

	tr := NewBackwardTracer(tensor.CPU)
	fa := tr.BindLeaf(a.t.Raw())
	fb := tr.BindLeaf(b.t.Raw())
	fs := tr.BindLeaf(seed.t.Raw())

	// A stand-in backward computation: grad = seed * leaf.
	gradA := tr.Mul(fs, fa)
	tr.Mul(fs, fb)

	id, ok := tr.ids[gradA]
	if !ok {
		t.Fatal("output was not recorded")
	}
	tr.g.output = id
	tr.g.outputs = []int{id}
	g := tr.g
	g.Optimize()

	// Replay with fresh leaves: a=2, seed=5 gives grad = 10.
	na := mustTensor(t, []float32{2, 2, 2}, []int{3})
	nb := mustTensor(t, []float32{3, 3, 3}, []int{3})
	ns := mustTensor(t, []float32{5, 5, 5}, []int{3})
	slots := map[int]*tensor.RawTensor{}
	for _, lb := range tr.leafRaws {
		switch lb.raw {
		case a.t.Raw():
			slots[lb.id] = na.t.Raw()
		case b.t.Raw():
			slots[lb.id] = nb.t.Raw()
		case seed.t.Raw():
			slots[lb.id] = ns.t.Raw()
		}
	}

	raw, err := g.RunWithLeaves(a.be, slots)
	if err != nil {
		t.Fatalf("RunWithLeaves: %v", err)
	}
	got := wrapRaw(a.be, raw).ContiguousData()
	if !closeData([]float32{10, 10, 10}, got) {
		t.Fatalf("replay with rebinding: want [10 10 10], got %v", got)
	}
}

// TestRunWithLeavesRejectsUnbound checks a replay fails loudly rather than
// silently passing a nil tensor into a kernel.
func TestRunWithLeavesRejectsUnbound(t *testing.T) {
	a := mustTensor(t, []float32{1, 2}, []int{2})
	b := mustTensor(t, []float32{3, 4}, []int{2})

	tr := NewBackwardTracer(tensor.CPU)
	fa := tr.BindLeaf(a.t.Raw())
	fb := tr.BindLeaf(b.t.Raw())
	out := tr.Add(fa, fb)
	id, _ := tr.ids[out]
	tr.g.output = id
	g := tr.g

	if _, err := g.RunWithLeaves(a.be, map[int]*tensor.RawTensor{}); err == nil {
		t.Fatal("expected an unbound leaf to be rejected")
	}
}
