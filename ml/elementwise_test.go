package ml

import (
	"testing"
)

// TestElementwiseFusionFires captures an elementwise chain and checks it folds
// into a single node and replays to the same values.
func TestElementwiseFusionFires(t *testing.T) {
	x := mustTensor(t, []float32{0.5, -1.0, 2.0, -3.0, 0.25, 4.0}, []int{2, 3})

	fwd := func(v *Tensor) (*Tensor, error) {
		// A chain of five elementwise ops: these should become one node.
		h, err := Mul(v, v)
		if err != nil {
			return nil, err
		}
		h = wrapRaw(h.be, h.be.AddScalar(h.t.Raw(), float32(1)))
		h, err = Sqrt(h)
		if err != nil {
			return nil, err
		}
		h, err = Relu(h)
		if err != nil {
			return nil, err
		}
		return Exp(h)
	}

	want, err := fwd(x)
	if err != nil {
		t.Fatalf("eager: %v", err)
	}
	wantData := want.ContiguousData()

	tr, err := NewTracer(x)
	if err != nil {
		t.Fatalf("NewTracer: %v", err)
	}
	out, err := fwd(tr.InputTensor())
	if err != nil {
		t.Fatalf("trace: %v", err)
	}
	g, err := tr.Finish(out)
	if err != nil {
		t.Fatalf("Finish: %v", err)
	}
	g.Optimize()

	live := map[string]int{}
	for i := range g.nodes {
		if !g.nodes[i].dead {
			live[g.nodes[i].op]++
		}
	}
	if live["elementwise"] != 1 {
		t.Fatalf("expected one fused elementwise node, got %v (stats=%s)", live, g.Stats())
	}
	for _, op := range []string{"mul", "add_scalar", "sqrt", "relu", "exp"} {
		if live[op] != 0 {
			t.Fatalf("op %q survived fusion: %v", op, live)
		}
	}

	raw, err := g.Run(x.be, x.t.Raw())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	got := wrapRaw(x.be, raw).ContiguousData()
	if !closeData(wantData, got) {
		t.Fatalf("fused elementwise differs from eager:\n want %v\n got  %v", wantData, got)
	}
}

// TestElementwiseChainBackward checks gradients still flow through a fused
// elementwise chain (replay runs real ops, so the tape is unaffected).
func TestElementwiseChainBackward(t *testing.T) {
	x := mustTensor(t, []float32{0.5, 1.5, 2.5, 3.5}, []int{4})
	x.SetRequiresGrad(true)

	fwd := func(v *Tensor) (*Tensor, error) {
		h, err := Mul(v, v)
		if err != nil {
			return nil, err
		}
		h = wrapRaw(h.be, h.be.AddScalar(h.t.Raw(), float32(1)))
		h, err = Relu(h)
		if err != nil {
			return nil, err
		}
		return Log(h)
	}

	// Eager gradient.
	e, err := fwd(x)
	if err != nil {
		t.Fatalf("eager: %v", err)
	}
	es, err := Sum(e, nil, false)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	if err := es.Backward(); err != nil {
		t.Fatalf("eager backward: %v", err)
	}
	eagerGrad := x.Grad().ContiguousData()
	x.ZeroGrad()

	// Compiled gradient.
	tr, err := NewTracer(x)
	if err != nil {
		t.Fatalf("NewTracer: %v", err)
	}
	out, err := fwd(tr.InputTensor())
	if err != nil {
		t.Fatalf("trace: %v", err)
	}
	g, err := tr.Finish(out)
	if err != nil {
		t.Fatalf("Finish: %v", err)
	}
	g.Optimize()
	raw, err := g.Run(x.be, x.t.Raw())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	cs, err := Sum(wrapRaw(x.be, raw), nil, false)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	if err := cs.Backward(); err != nil {
		t.Fatalf("compiled backward: %v", err)
	}
	if !closeData(eagerGrad, x.Grad().ContiguousData()) {
		t.Fatalf("fused chain grad differs:\n want %v\n got  %v", eagerGrad, x.Grad().ContiguousData())
	}
}
