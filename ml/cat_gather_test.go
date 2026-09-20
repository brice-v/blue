package ml

import (
	"testing"

	"blue/borncgo/tensor"
)

func mustIntTensor(t *testing.T, data []int32, shape []int) *Tensor {
	t.Helper()
	tt, err := NewInt32Tensor(data, shape, CPU)
	if err != nil {
		t.Fatalf("NewInt32Tensor: %v", err)
	}
	return tt
}

// TestCatForward checks cat against torch.cat semantics for both dims.
func TestCatForward(t *testing.T) {
	a := mustTensor(t, []float32{1, 2, 3, 4, 5, 6}, []int{2, 3})
	b := mustTensor(t, []float32{7, 8, 9}, []int{1, 3})
	c := mustTensor(t, []float32{10, 11, 12}, []int{1, 3})

	rows, err := Cat([]*Tensor{a, b, c}, 0)
	if err != nil {
		t.Fatalf("Cat dim 0: %v", err)
	}
	if got := rows.Shape(); len(got) != 2 || got[0] != 4 || got[1] != 3 {
		t.Fatalf("dim 0 shape = %v, want [4 3]", got)
	}
	want := []float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	if !closeData(want, rows.ContiguousData()) {
		t.Fatalf("dim 0 data = %v, want %v", rows.ContiguousData(), want)
	}

	cols, err := Cat([]*Tensor{a, a}, 1)
	if err != nil {
		t.Fatalf("Cat dim 1: %v", err)
	}
	if got := cols.Shape(); len(got) != 2 || got[0] != 2 || got[1] != 6 {
		t.Fatalf("dim 1 shape = %v, want [2 6]", got)
	}
}

// TestCatRejectsMismatch checks that non-cat dimensions must agree.
func TestCatRejectsMismatch(t *testing.T) {
	a := mustTensor(t, []float32{1, 2, 3, 4}, []int{2, 2})
	b := mustTensor(t, []float32{1, 2, 3, 4, 5, 6}, []int{2, 3})
	if _, err := Cat([]*Tensor{a, b}, 0); err == nil {
		t.Fatal("expected cat to reject mismatched trailing dims")
	}
}

// TestGatherForward checks torch.gather semantics for dim=1.
func TestGatherForward(t *testing.T) {
	a := mustTensor(t, []float32{1, 2, 3, 4, 5, 6}, []int{2, 3})
	idx := mustIntTensor(t, []int32{2, 0, 1, 1}, []int{2, 2})

	g, err := Gather(a, 1, idx)
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	if got := g.Shape(); len(got) != 2 || got[0] != 2 || got[1] != 2 {
		t.Fatalf("gather shape = %v, want [2 2]", got)
	}
	want := []float32{3, 1, 5, 5}
	if !closeData(want, g.ContiguousData()) {
		t.Fatalf("gather data = %v, want %v", g.ContiguousData(), want)
	}
}

// TestGatherBackward checks that gradients scatter back to the selected
// positions, matching torch.gather's backward.
func TestGatherBackward(t *testing.T) {
	x := mustTensor(t, []float32{1, 2, 3}, []int{1, 3})
	x.SetRequiresGrad(true)
	idx := mustIntTensor(t, []int32{2, 0}, []int{1, 2})

	g, err := Gather(x, 1, idx)
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	loss, err := Sum(g, nil, false)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	if err := loss.Backward(); err != nil {
		t.Fatalf("Backward: %v", err)
	}
	// Position 1 was never selected, so its gradient is zero.
	want := []float32{1, 0, 1}
	if !closeData(want, x.Grad().ContiguousData()) {
		t.Fatalf("gather grad = %v, want %v", x.Grad().ContiguousData(), want)
	}
}

// TestCatBackward checks cat's backward sends each operand its own slice.
func TestCatBackward(t *testing.T) {
	p := mustTensor(t, []float32{1, 2}, []int{1, 2})
	q := mustTensor(t, []float32{3, 4}, []int{1, 2})
	p.SetRequiresGrad(true)
	q.SetRequiresGrad(true)

	j, err := Cat([]*Tensor{p, q}, 1)
	if err != nil {
		t.Fatalf("Cat: %v", err)
	}
	loss, err := Sum(j, nil, false)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	if err := loss.Backward(); err != nil {
		t.Fatalf("Backward: %v", err)
	}
	if !closeData([]float32{1, 1}, p.Grad().ContiguousData()) {
		t.Fatalf("p grad = %v, want [1 1]", p.Grad().ContiguousData())
	}
	if !closeData([]float32{1, 1}, q.Grad().ContiguousData()) {
		t.Fatalf("q grad = %v, want [1 1]", q.Grad().ContiguousData())
	}
}

// TestCompiledCatGather checks that cat and gather survive capture, fusion and
// replay, and that replay on the real backend still autodiffs.
func TestCompiledCatGather(t *testing.T) {
	w := mustTensor(t, []float32{1, 0, 0, 0, 1, 0, 0, 0, 1}, []int{3, 3})
	w.SetRequiresGrad(true)
	idx := mustIntTensor(t, []int32{0, 2, 2, 0}, []int{2, 2})

	fwd := func(x *Tensor) (*Tensor, error) {
		h, err := MatMul(x, w)
		if err != nil {
			return nil, err
		}
		h, err = Cat([]*Tensor{h, h}, 1) // [2, 6]
		if err != nil {
			return nil, err
		}
		return Gather(h, 1, idx) // [2, 2]
	}

	ex := mustTensor(t, []float32{1, 2, 3, 4, 5, 6}, []int{2, 3})

	// Eager result and gradient.
	eager, err := fwd(ex)
	if err != nil {
		t.Fatalf("eager forward: %v", err)
	}
	eagerData := eager.ContiguousData()
	eagerLoss, err := Sum(eager, nil, false)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	if err := eagerLoss.Backward(); err != nil {
		t.Fatalf("eager backward: %v", err)
	}
	eagerGrad := w.Grad().ContiguousData()
	w.ZeroGrad()

	// Capture, optimise and replay.
	tr, err := NewTracer(ex)
	if err != nil {
		t.Fatalf("NewTracer: %v", err)
	}
	out, err := fwd(tr.InputTensor())
	if err != nil {
		t.Fatalf("traced forward: %v", err)
	}
	g, err := tr.Finish(out)
	if err != nil {
		t.Fatalf("Finish: %v", err)
	}
	g.Optimize()

	raw, err := g.Run(ex.be, ex.t.Raw())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	got := wrapRaw(ex.be, raw).ContiguousData()
	if !closeData(eagerData, got) {
		t.Fatalf("compiled cat/gather differs from eager:\n want %v\n got  %v", eagerData, got)
	}

	loss, err := Sum(wrapRaw(ex.be, raw), nil, false)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	if err := loss.Backward(); err != nil {
		t.Fatalf("compiled backward: %v", err)
	}
	_ = tensor.CPU
	if !closeData(eagerGrad, w.Grad().ContiguousData()) {
		t.Fatalf("compiled cat/gather grad differs from eager:\n want %v\n got  %v", eagerGrad, w.Grad().ContiguousData())
	}
}
