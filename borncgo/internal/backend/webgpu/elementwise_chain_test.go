//go:build !js

package webgpu

import (
	"math"
	"testing"

	"blue/borncgo/internal/tensor"
)

// TestElementwiseChainFused checks the generated WGSL kernel against the
// per-op reference for a chain of arithmetic and transcendental ops.
func TestElementwiseChainFused(t *testing.T) {
	if !computeAvailable {
		t.Skip("WebGPU not available")
	}
	backend, err := New()
	if err != nil {
		t.Fatalf("failed to create backend: %v", err)
	}
	defer backend.Release()

	input := []float32{-2, -0.5, 0, 0.5, 2, 3}
	shape := tensor.Shape{6}
	x := createTensor(t, shape, input)

	// slots: 0 = x. step 0: mul(0,0); 1: add_scalar(1, +1); 2: relu(2);
	// 3: exp(3).
	steps := []tensor.ElementwiseStep{
		{Op: "mul", A: 0, B: 0},
		{Op: "add_scalar", A: 1, Scalar: 1},
		{Op: "relu", A: 2, B: -1},
		{Op: "exp", A: 3, B: -1},
	}

	got := backend.ElementwiseChain([]*tensor.RawTensor{x}, steps, shape, tensor.Float32)
	if got == nil {
		t.Fatal("expected the fused elementwise kernel to accept the chain")
	}
	actual := extractData(t, got)

	for i, v := range input {
		pre := v*v + 1
		if pre < 0 {
			pre = 0
		}
		want := float32(math.Exp(float64(pre)))
		rel := math.Abs(float64(actual[i]-want)) / math.Max(1, math.Abs(float64(want)))
		if rel > 1e-5 {
			t.Errorf("element %d: got %v, want %v (rel %g)", i, actual[i], want, rel)
		}
	}
}

// TestElementwiseChainFusedMultiLeaf checks a chain that reads two leaves.
func TestElementwiseChainFusedMultiLeaf(t *testing.T) {
	if !computeAvailable {
		t.Skip("WebGPU not available")
	}
	backend, err := New()
	if err != nil {
		t.Fatalf("failed to create backend: %v", err)
	}
	defer backend.Release()

	shape := tensor.Shape{4}
	a := createTensor(t, shape, []float32{1, 2, 3, 4})
	b := createTensor(t, shape, []float32{10, 20, 30, 40})

	// (a * b) + a, then relu. slots: 0 = a, 1 = b.
	steps := []tensor.ElementwiseStep{
		{Op: "mul", A: 0, B: 1},
		{Op: "add", A: 2, B: 0},
		{Op: "relu", A: 3, B: -1},
	}
	got := backend.ElementwiseChain([]*tensor.RawTensor{a, b}, steps, shape, tensor.Float32)
	if got == nil {
		t.Fatal("expected the kernel to accept a two-leaf chain")
	}
	actual := extractData(t, got)
	want := []float32{11, 42, 93, 164}
	if !compareSlices(t, want, actual, 1e-5) {
		t.Errorf("multi-leaf chain: want %v, got %v", want, actual)
	}
}

// TestElementwiseChainDeclinesUnsupported checks the backend returns nil (so the
// caller falls back) for shapes it cannot fuse.
func TestElementwiseChainDeclinesUnsupported(t *testing.T) {
	if !computeAvailable {
		t.Skip("WebGPU not available")
	}
	backend, err := New()
	if err != nil {
		t.Fatalf("failed to create backend: %v", err)
	}
	defer backend.Release()

	x := createTensor(t, tensor.Shape{2, 2}, []float32{1, 2, 3, 4})
	steps := []tensor.ElementwiseStep{{Op: "relu", A: 0, B: -1}}
	if got := backend.ElementwiseChain([]*tensor.RawTensor{x}, steps, tensor.Shape{4}, tensor.Float32); got != nil {
		t.Fatal("expected the kernel to decline a shape mismatch")
	}
}
