package ml

import (
	"testing"

	"blue/borncgo/tensor"
)

// bindBackward binds a backward plan's leaves: forward values map to the live
// tensors the forward replay produced, and the extra leaf (the seed) to ones.
func bindBackward(t *testing.T, plan *backwardPlan, fwd *Graph, input *tensor.RawTensor, be tensor.Backend) map[int]*tensor.RawTensor {
	t.Helper()
	slots, err := fwd.replaySlots(be, map[int]*tensor.RawTensor{fwd.input: input})
	if err != nil {
		t.Fatalf("forward replay: %v", err)
	}
	leaves := map[int]*tensor.RawTensor{}
	for fwdID, slot := range plan.leaves {
		leaves[slot] = slots[fwdID]
	}
	// Only the seed leaf gets a gradient of one; constant leaves keep their
	// captured values and must not be overwritten.
	leaves[plan.seed] = onesShape(plan.graph.values[plan.seed].shape, input)
	return leaves
}

func onesShape(shape tensor.Shape, like *tensor.RawTensor) *tensor.RawTensor {
	n := 1
	for _, s := range shape {
		n *= s
	}
	if n == 0 {
		n = 1
	}
	raw, err := tensor.NewRaw(tensor.Shape{n}, tensor.Float32, like.Device())
	if err != nil {
		panic(err)
	}
	for i := range raw.AsFloat32() {
		raw.AsFloat32()[i] = 1
	}
	return raw
}

// TestBuildBackwardMatchesEager differentiates a small MLP by construction and
// checks the parameter gradients match the eager backward.
func TestBuildBackwardMatchesEager(t *testing.T) {
	w1 := mustTensor(t, []float32{0.1, -0.2, 0.3, 0.4, 0.5, -0.6, 0.7, 0.8, -0.9, 1.0, 0.2, -0.3, 0.4, 0.5, -0.6}, []int{3, 5})
	b1 := mustTensor(t, []float32{0.1, -0.1, 0.2, -0.2, 0.3}, []int{5})
	w2 := mustTensor(t, []float32{0.5, -0.5, 0.25, 0.75, -0.25, 0.6, -0.6, 0.1, 0.2, 0.3}, []int{5, 2})
	b2 := mustTensor(t, []float32{0.01, -0.02}, []int{2})
	x := mustTensor(t, []float32{0.2, -0.4, 0.6, 0.8, -1.0, 1.2, 0.3, -0.5, 0.7, -0.9, 1.1, -1.3}, []int{4, 3})
	params := []*Tensor{w1, b1, w2, b2}
	for _, p := range params {
		p.SetRequiresGrad(true)
	}

	fwd := func(v *Tensor) (*Tensor, error) {
		h, err := MatMul(v, w1)
		if err != nil {
			return nil, err
		}
		h, err = Add(h, b1)
		if err != nil {
			return nil, err
		}
		h, err = Relu(h)
		if err != nil {
			return nil, err
		}
		h, err = MatMul(h, w2)
		if err != nil {
			return nil, err
		}
		h, err = Add(h, b2)
		if err != nil {
			return nil, err
		}
		return Sum(h, nil, false)
	}

	// Eager reference gradients.
	loss, err := fwd(x)
	if err != nil {
		t.Fatalf("eager forward: %v", err)
	}
	if err := loss.Backward(); err != nil {
		t.Fatalf("eager backward: %v", err)
	}
	want := make([][]float32, len(params))
	for i, p := range params {
		want[i] = append([]float32(nil), p.Grad().ContiguousData()...)
		p.ZeroGrad()
	}

	// Capture, differentiate, replay.
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
	plan.graph.outputs = append(plan.graph.outputs, plan.graph.output)
	plan.graph.Optimize()
	t.Logf("forward %s, backward %s", fg.Stats(), plan.graph.Stats())

	leaves := bindBackward(t, plan, fg, x.t.Raw(), x.be)

	// Match each parameter to its forward value via the bound raw identity.
	for i, p := range params {
		fwdID := -1
		for id, v := range fg.values {
			if v.raw == p.t.Raw() {
				fwdID = id
				break
			}
		}
		if fwdID < 0 {
			t.Fatalf("parameter %d not found in the forward graph", i)
		}
		slot, ok := plan.paramGrads[fwdID]
		if !ok {
			t.Fatalf("no captured gradient for parameter %d", i)
		}
		grad, err := plan.graph.RunWithLeavesOutput(x.be, leaves, slot)
		if err != nil {
			t.Fatalf("backward replay: %v", err)
		}
		got := wrapRaw(x.be, grad).ContiguousData()
		if !closeData(want[i], got) {
			t.Fatalf("param %d gradient mismatch:\n want %v\n got  %v", i, want[i], got)
		}
	}
}
