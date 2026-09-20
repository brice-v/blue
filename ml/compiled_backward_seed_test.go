package ml

import "testing"

// TestCompiledBackwardWithSeed checks that an explicit output-gradient seed
// produces the same parameter gradients as the eager backward of the same
// objective.
func TestCompiledBackwardWithSeed(t *testing.T) {
	l1, _ := NNLinear(4, 8, CPU)
	l2, _ := NNLinear(8, 3, CPU)
	fwd := func(x *Tensor) (*Tensor, error) {
		h := NNForward(l1, x)
		h, _ = Relu(h)
		return NNForward(l2, h), nil
	}
	x := mustTensor(t, []float32{
		0.1, 0.2, 0.3, 0.4,
		0.5, -0.6, 0.7, -0.8,
		1.0, 0.9, -0.8, 0.7,
	}, []int{3, 4})

	// Eager: a weighted sum of the output, so the seed is the weights.
	weights := mustTensor(t, []float32{0.5, -1.0, 2.0, 1.5, 0.25, -0.75, 1.0, 0.5, -1.5}, []int{3, 3})
	logits, err := fwd(x)
	if err != nil {
		t.Fatalf("eager forward: %v", err)
	}
	weighted, err := Mul(logits, weights)
	if err != nil {
		t.Fatalf("Mul: %v", err)
	}
	loss, err := Sum(weighted, nil, false)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	if err := loss.Backward(); err != nil {
		t.Fatalf("eager backward: %v", err)
	}
	want := map[*Param][]float32{}
	var params []*Param
	for _, m := range []Module{l1, l2} {
		for _, p := range NNParameters(m) {
			params = append(params, p)
			g := lastGradsFor(x.be)[p.Tensor().Raw()]
			want[p] = append([]float32(nil), wrapRaw(x.be, g).ContiguousData()...)
		}
	}

	// Compiled: same forward, seeded with the weights.
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
	if err := c.BackwardWithSeed(x, weights); err != nil {
		t.Fatalf("BackwardWithSeed: %v", err)
	}
	for _, p := range params {
		g := lastGradsFor(x.be)[p.Tensor().Raw()]
		if g == nil {
			t.Fatalf("compiled backward left a parameter without a gradient")
		}
		got := wrapRaw(x.be, g).ContiguousData()
		if !closeData(want[p], got) {
			t.Fatalf("seeded compiled gradient differs:\n want %v\n got  %v", want[p], got)
		}
	}
}

// TestCompiledBackwardRejectsBadSeed checks a mismatched seed is reported.
func TestCompiledBackwardRejectsBadSeed(t *testing.T) {
	l, _ := NNLinear(4, 3, CPU)
	x := mustTensor(t, []float32{0.1, 0.2, 0.3, 0.4}, []int{1, 4})
	fwd := func(v *Tensor) (*Tensor, error) { return NNForward(l, v), nil }

	c := NewCompiled()
	tr, _ := NewTracer(x)
	out, _ := fwd(tr.InputTensor())
	fg, _ := tr.Finish(out)
	fg.Optimize()
	c.Install(x, fg)

	bad := mustTensor(t, []float32{1, 1}, []int{2})
	if err := c.BackwardWithSeed(x, bad); err == nil {
		t.Fatal("expected a shape mismatch to be reported")
	}
}
