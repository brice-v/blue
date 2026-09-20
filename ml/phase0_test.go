package ml

import (
	"runtime"
	"testing"
	"time"

	"blue/borncgo/tensor"
)

// TestDTypeHonesty checks that creation stores the requested dtype instead of
// always storing float32 and tagging it.
func TestDTypeHonesty(t *testing.T) {
	cases := []struct {
		dtype DType
		raw   tensor.DataType
	}{
		{Float32, tensor.Float32},
		{Float64, tensor.Float64},
		{Int32, tensor.Int32},
		{Int64, tensor.Int64},
		{Uint8, tensor.Uint8},
		{Bool, tensor.Bool},
	}
	for _, tc := range cases {
		z, err := Zeros([]int{2, 2}, tc.dtype, CPU)
		if err != nil {
			t.Fatalf("Zeros(%s): %v", tc.dtype, err)
		}
		if z.DType() != tc.dtype {
			t.Fatalf("Zeros(%s) tag = %s", tc.dtype, z.DType())
		}
		if got := z.Raw().DType(); got != tc.raw {
			t.Fatalf("Zeros(%s) raw dtype = %s, want %s", tc.dtype, got, tc.raw)
		}
		o, err := Ones([]int{2, 2}, tc.dtype, CPU)
		if err != nil {
			t.Fatalf("Ones(%s): %v", tc.dtype, err)
		}
		if got := o.Raw().DType(); got != tc.raw {
			t.Fatalf("Ones(%s) raw dtype = %s, want %s", tc.dtype, got, tc.raw)
		}
		for _, v := range o.ContiguousData() {
			if v != 1 {
				t.Fatalf("Ones(%s) element = %v, want 1", tc.dtype, v)
			}
		}
	}
}

// TestNewTensorDType checks that NewTensor builds a real typed buffer.
func TestNewTensorDType(t *testing.T) {
	i, err := NewTensor([]float32{1, 2, 3}, []int{3}, Int32, CPU)
	if err != nil {
		t.Fatalf("NewTensor int32: %v", err)
	}
	if got := i.Raw().DType(); got != tensor.Int32 {
		t.Fatalf("NewTensor int32 raw dtype = %s", got)
	}
	b, err := NewTensor([]float32{0, 1}, []int{2}, Bool, CPU)
	if err != nil {
		t.Fatalf("NewTensor bool: %v", err)
	}
	if got := b.Raw().DType(); got != tensor.Bool {
		t.Fatalf("NewTensor bool raw dtype = %s", got)
	}
}

// TestClonePreservesGradient checks that clone stays in the graph, like
// torch.Tensor.clone, so gradients flow back to the original.
func TestClonePreservesGradient(t *testing.T) {
	x := dense([]float32{1, 2, 3}, 3)
	x.SetRequiresGrad(true)
	c := x.Clone()
	if !c.RequiresGrad() {
		t.Fatal("clone lost requires_grad")
	}
	s, err := Sum(c, nil, false)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	if err := s.Backward(); err != nil {
		t.Fatalf("Backward: %v", err)
	}
	check(t, "clone grad", x.Grad(), nil, []int{3}, []float32{1, 1, 1})
}

// TestDetachDropsGradient checks that detach is cut out of the graph.
func TestDetachDropsGradient(t *testing.T) {
	x := dense([]float32{1, 2, 3}, 3)
	x.SetRequiresGrad(true)
	d := x.Detach()
	if d.RequiresGrad() {
		t.Fatal("detach kept requires_grad")
	}
	s, err := Sum(d, nil, false)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	if err := s.Backward(); err != nil {
		t.Fatalf("Backward: %v", err)
	}
	if x.Grad() != nil {
		t.Fatalf("detach leaked a gradient to the original: %v", x.Grad())
	}
}

// makeTransientLeaf creates and drops a leaf that requires grad, in its own
// function so the compiler cannot keep it alive past the call.
func makeTransientLeaf(i int) {
	x := dense([]float32{float32(i)}, 1)
	x.SetRequiresGrad(true)
}

// TestTrackedLeavesArePruned checks that dropping leaves lets the finalizer
// shrink the tracked set, so it is bounded by live leaves.
func TestTrackedLeavesArePruned(t *testing.T) {
	base := len(liveLeaves())
	for i := range 2000 {
		makeTransientLeaf(i)
	}
	// Finalizers run asynchronously after a GC, so retry briefly.
	for range 50 {
		runtime.GC()
		if len(liveLeaves()) <= base+100 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("tracked grew to %d from %d, finalizer did not prune", len(liveLeaves()), base)
}

// TestBackwardAfterTransientLeaves checks that a backward still finds the live
// leaves after many transient leaves have come and gone.
func TestBackwardAfterTransientLeaves(t *testing.T) {
	for i := range 500 {
		makeTransientLeaf(i)
	}
	x := dense([]float32{1, 2, 3}, 3)
	x.SetRequiresGrad(true)
	s, err := Sum(x, nil, false)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	if err := s.Backward(); err != nil {
		t.Fatalf("Backward: %v", err)
	}
	check(t, "grad after transients", x.Grad(), nil, []int{3}, []float32{1, 1, 1})
}
