package ml

import "testing"

// TestCompiledBackwardGPU checks the compiled backward is correct when the
// forward is captured and replayed on the GPU, so the backward's fused
// elementwise chains run through the generated WGSL kernel.
func TestCompiledBackwardGPU(t *testing.T) {
	if !GPUAvailable() {
		t.Skip("GPU not available")
	}
	l1, err := NNLinear(4, 6, GPU)
	if err != nil {
		t.Fatalf("NNLinear: %v", err)
	}
	l2, err := NNLinear(6, 2, GPU)
	if err != nil {
		t.Fatalf("NNLinear: %v", err)
	}
	fwd := func(x *Tensor) (*Tensor, error) {
		h := NNForward(l1, x)
		h, err := Relu(h)
		if err != nil {
			return nil, err
		}
		return NNForward(l2, h), nil
	}
	xg, err := mustTensor(t, []float32{
		0.1, 0.2, 0.3, 0.4,
		-0.5, 0.6, -0.7, 0.8,
	}, []int{2, 4}).To(GPU)
	if err != nil {
		t.Fatalf("to gpu: %v", err)
	}

	// Eager gradient of the sum of the output.
	logits, err := fwd(xg)
	if err != nil {
		t.Fatalf("eager forward: %v", err)
	}
	loss, err := Sum(logits, nil, false)
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
			want[p] = append([]float32(nil), wrapRaw(xg.be, lastGradsFor(xg.be)[p.Tensor().Raw()]).ContiguousData()...)
		}
	}

	c := NewCompiled()
	tr, err := NewTracer(xg)
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
	c.Install(xg, fg)
	if err := c.Backward(xg); err != nil {
		t.Fatalf("compiled gpu backward: %v", err)
	}
	bp := c.BackwardPlan(xg)
	t.Logf("forward %s | backward %s", fg.Stats(), bp.graph.Stats())

	for _, p := range params {
		g := lastGradsFor(xg.be)[p.Tensor().Raw()]
		if g == nil {
			t.Fatalf("compiled gpu backward left a parameter without a gradient")
		}
		got := wrapRaw(xg.be, g).ContiguousData()
		if !closeData(want[p], got) {
			t.Fatalf("gpu compiled gradient differs:\n want %v\n got  %v", want[p], got)
		}
	}
}
