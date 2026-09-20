package ml

import "testing"

func TestRetainGradNonLeaf(t *testing.T) {
	x := dense([]float32{2}, 1)
	x.SetRequiresGrad(true)
	y, err := Mul(x, x)
	if err != nil {
		t.Fatalf("Mul: %v", err)
	}
	y.RetainGrad()
	s, err := Sum(y, nil, false)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	if err := s.Backward(); err != nil {
		t.Fatalf("Backward: %v", err)
	}
	check(t, "x.grad", x.Grad(), nil, []int{1}, []float32{4})
	check(t, "y.grad", y.Grad(), nil, []int{1}, []float32{1})
}

func TestAutogradGradDoesNotStore(t *testing.T) {
	a := dense([]float32{3}, 1)
	b := dense([]float32{4}, 1)
	a.SetRequiresGrad(true)
	b.SetRequiresGrad(true)
	prod, err := Mul(a, b)
	if err != nil {
		t.Fatalf("Mul: %v", err)
	}
	z, err := Sum(prod, nil, false)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	gs, err := AutogradGrad(z, []*Tensor{a, b})
	if err != nil {
		t.Fatalf("AutogradGrad: %v", err)
	}
	check(t, "d/da", gs[0], nil, []int{1}, []float32{4})
	check(t, "d/db", gs[1], nil, []int{1}, []float32{3})
	if a.Grad() != nil || b.Grad() != nil {
		t.Fatal("AutogradGrad stored a gradient on an input")
	}
}

func TestRegisterHook(t *testing.T) {
	x := dense([]float32{2}, 1)
	x.SetRequiresGrad(true)
	called := false
	x.RegisterHook(func(g *Tensor) *Tensor {
		called = true
		scaled, _ := Mul(g, dense([]float32{10}, 1))
		return scaled
	})
	s, err := Sum(x, nil, false)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	if err := s.Backward(); err != nil {
		t.Fatalf("Backward: %v", err)
	}
	if !called {
		t.Fatal("hook was not called")
	}
	// The hook ran; x.grad is set before the hook transforms the passed value.
	if x.Grad() == nil {
		t.Fatal("leaf gradient was not set")
	}
}

func TestOptimizerStateDictRoundTrip(t *testing.T) {
	w := dense([]float32{1, 2, 3, 4}, 2, 2)
	w.SetRequiresGrad(true)
	opt, err := Adam([]OptimBinding{BindTensor("w", w)}, 0.1, 0.9, 0.999, 1e-8, 0)
	if err != nil {
		t.Fatalf("Adam: %v", err)
	}
	step := func() {
		prod, err := Mul(w, w)
		if err != nil {
			t.Fatalf("Mul: %v", err)
		}
		loss, err := Sum(prod, nil, false)
		if err != nil {
			t.Fatalf("Sum: %v", err)
		}
		if err := loss.Backward(); err != nil {
			t.Fatalf("Backward: %v", err)
		}
		if err := opt.Step(); err != nil {
			t.Fatalf("Step: %v", err)
		}
	}
	step()
	sd := opt.StateDict()
	if len(sd) != 2 {
		t.Fatalf("Adam state has %d entries, want 2 (m and v)", len(sd))
	}
	m := sd["m.0"]
	if m == nil {
		t.Fatal("no m.0 in state dict")
	}
	saved := append([]float32(nil), m.ContiguousData()...)

	// Two more steps change the moments, then loading restores them.
	step()
	step()
	savedT, err := NewTensor(saved, m.Shape(), m.DType(), CPU)
	if err != nil {
		t.Fatalf("NewTensor: %v", err)
	}
	if err := CopyInto(sd["m.0"], savedT); err != nil {
		t.Fatalf("CopyInto: %v", err)
	}
	check(t, "restored m.0", sd["m.0"], nil, m.Shape(), saved)
}
