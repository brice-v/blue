package ml

import "testing"

// TestBackwardFusionFoldsElementwise checks that the constructed backward graph
// gets the same elementwise chain fusion as the forward: a relu backward plus
// its mask cast and multiply should collapse toward a single node.
func TestBackwardFusionFoldsElementwise(t *testing.T) {
	w := mustTensor(t, []float32{0.1, -0.2, 0.3, 0.4, 0.5, -0.6, 0.7, 0.8, -0.9, 1.0, 0.2, -0.3}, []int{3, 4})
	x := mustTensor(t, []float32{0.2, -0.4, 0.6, 0.8, -1.0, 1.2}, []int{2, 3})
	w.SetRequiresGrad(true)

	fwd := func(v *Tensor) (*Tensor, error) {
		h, err := MatMul(v, w)
		if err != nil {
			return nil, err
		}
		h, err = Relu(h)
		if err != nil {
			return nil, err
		}
		return Sum(h, nil, false)
	}

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

	plan, err := BuildBackward(fg)
	if err != nil {
		t.Fatalf("BuildBackward: %v", err)
	}
	before := 0
	for i := range plan.graph.nodes {
		if !plan.graph.nodes[i].dead {
			before++
		}
	}
	plan.graph.outputs = append(plan.graph.outputs, plan.graph.output)
	plan.graph.Optimize()
	after := 0
	fused := 0
	for i := range plan.graph.nodes {
		if !plan.graph.nodes[i].dead {
			after++
			if plan.graph.nodes[i].op == "elementwise" {
				fused++
			}
		}
	}
	t.Logf("backward nodes %d -> %d (elementwise nodes %d, stats %s)", before, after, fused, plan.graph.Stats())
	if fused == 0 {
		t.Fatalf("expected the backward graph to contain fused elementwise nodes")
	}
	if after >= before {
		t.Fatalf("expected fusing to reduce the backward node count, got %d -> %d", before, after)
	}
}
