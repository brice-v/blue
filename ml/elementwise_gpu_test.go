package ml

import (
	"testing"
)

// TestCompiledElementwiseGPUMatchesEager captures an elementwise chain and
// replays it on the GPU backend, checking the fused kernel matches eager. It
// skips when no GPU adapter exists.
func TestCompiledElementwiseGPUMatchesEager(t *testing.T) {
	if !GPUAvailable() {
		t.Skip("GPU not available")
	}

	x, err := NewTensor([]float32{-2, -0.5, 0, 0.5, 2, 3}, []int{6}, Float32, CPU)
	if err != nil {
		t.Fatalf("NewTensor: %v", err)
	}
	xg, err := x.To(GPU)
	if err != nil {
		t.Fatalf("to gpu: %v", err)
	}

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
		return Exp(h)
	}

	want, err := fwd(xg)
	if err != nil {
		t.Fatalf("eager gpu: %v", err)
	}
	wantData := want.ContiguousData()

	tr, err := NewTracer(xg)
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

	// The chain must have folded into a single elementwise node.
	live := map[string]int{}
	for i := range g.nodes {
		if !g.nodes[i].dead {
			live[g.nodes[i].op]++
		}
	}
	if live["elementwise"] != 1 {
		t.Fatalf("expected a single fused elementwise node, got %v", live)
	}

	raw, err := g.Run(xg.be, xg.t.Raw())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	got := wrapRaw(xg.be, raw).ContiguousData()
	if !closeData(wantData, got) {
		t.Fatalf("gpu fused elementwise differs from eager:\n want %v\n got  %v", wantData, got)
	}
}
