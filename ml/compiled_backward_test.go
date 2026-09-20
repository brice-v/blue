package ml

import (
	"testing"
)

// TestCompiledBackwardDrivesOptimizer checks the compiled backward produces the
// same parameter gradients as the eager backward.
func TestCompiledBackwardDrivesOptimizer(t *testing.T) {
	l1, err := NNLinear(4, 8, CPU)
	if err != nil {
		t.Fatalf("NNLinear: %v", err)
	}
	l2, err := NNLinear(8, 2, CPU)
	if err != nil {
		t.Fatalf("NNLinear: %v", err)
	}
	layers := []Module{l1, l2}

	fwd := func(x *Tensor) (*Tensor, error) {
		h := NNForward(l1, x)
		h, err := Relu(h)
		if err != nil {
			return nil, err
		}
		return Sum(NNForward(l2, h), nil, false)
	}

	x := mustTensor(t, []float32{
		0.1, 0.2, 0.3, 0.4,
		-0.5, 0.6, -0.7, 0.8,
		0.9, -1.0, 1.1, -1.2,
	}, []int{3, 4})

	// Eager reference gradients, read from the map the optimizer consumes.
	loss, err := fwd(x)
	if err != nil {
		t.Fatalf("eager forward: %v", err)
	}
	if err := loss.Backward(); err != nil {
		t.Fatalf("eager backward: %v", err)
	}
	want := map[*Param][]float32{}
	params := []*Param{}
	for _, m := range layers {
		for _, p := range NNParameters(m) {
			params = append(params, p)
			g, ok := lastGradsFor(x.be)[p.Tensor().Raw()]
			if !ok || g == nil {
				t.Fatalf("eager backward produced no gradient for a parameter")
			}
			want[p] = append([]float32(nil), wrapRaw(x.be, g).ContiguousData()...)
		}
	}

	// Compiled path.
	c := NewCompiled()
	tr, err := NewTracer(x)
	if err != nil {
		t.Fatalf("NewTracer: %v", err)
	}
	out, err := fwd(tr.InputTensor())
	if err != nil {
		t.Fatalf("trace: %v", err)
	}
	fg, err := tr.Finish(out)
	if err != nil {
		t.Fatalf("Finish: %v", err)
	}
	fg.Optimize()
	c.Install(x, fg)

	if err := c.Backward(x); err != nil {
		t.Fatalf("compiled Backward: %v", err)
	}
	bp := c.BackwardPlan(x)
	t.Logf("stats: %s | forward %s | backward %s", c.Stats(), fg.Stats(), bp.graph.Stats())

	for _, p := range params {
		g, ok := lastGradsFor(x.be)[p.Tensor().Raw()]
		if !ok || g == nil {
			t.Fatalf("compiled backward produced no gradient for a parameter")
		}
		got := wrapRaw(x.be, g).ContiguousData()
		if !closeData(want[p], got) {
			t.Fatalf("compiled param gradient differs:\n want %v\n got  %v", want[p], got)
		}
	}
}
