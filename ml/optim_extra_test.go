package ml

import (
	"math"
	"testing"
)

func TestClipGradNorm(t *testing.T) {
	x := dense([]float32{1, 2}, 2)
	x.SetRequiresGrad(true)
	w := dense([]float32{3, 4}, 2)
	prod, err := Mul(x, w)
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
	// grad(x) = w = [3, 4], whose L2 norm is 5.
	b := BindTensor("x", x)
	norm, err := ClipGradNorm([]OptimBinding{b}, 1)
	if err != nil {
		t.Fatalf("ClipGradNorm: %v", err)
	}
	if math.Abs(float64(norm)-5) > 1e-5 {
		t.Fatalf("pre-clip norm = %v, want 5", norm)
	}
	g := lastGradsFor(x.be)[b.Param.Tensor().Raw()]
	check(t, "clipped grad", wrapRaw(x.be, g), nil, []int{2}, []float32{0.6, 0.8})
}

func TestClipGradNormBelowThresholdIsNoop(t *testing.T) {
	x := dense([]float32{1, 1}, 2)
	x.SetRequiresGrad(true)
	w := dense([]float32{1, 1}, 2)
	prod, err := Mul(x, w)
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
	b := BindTensor("x", x)
	if _, err := ClipGradNorm([]OptimBinding{b}, 10); err != nil {
		t.Fatalf("ClipGradNorm: %v", err)
	}
	g := lastGradsFor(x.be)[b.Param.Tensor().Raw()]
	check(t, "unclipped grad", wrapRaw(x.be, g), nil, []int{2}, []float32{1, 1})
}

func TestLRSchedule(t *testing.T) {
	cases := []struct {
		kind string
		base float32
		step int
		a, b float32
		want float32
	}{
		{"constant", 0.1, 50, 0, 0, 0.1},
		{"step", 1.0, 0, 10, 0.1, 1.0},
		{"step", 1.0, 10, 10, 0.1, 0.1},
		{"step", 1.0, 25, 10, 0.1, 0.01},
		{"warmup", 1.0, 0, 10, 0, 0.1},
		{"warmup", 1.0, 9, 10, 0, 1.0},
		{"warmup", 1.0, 50, 10, 0, 1.0},
		{"cosine", 1.0, 0, 10, 0, 1.0},
		{"cosine", 1.0, 10, 10, 0, 0.0},
		{"cosine", 1.0, 5, 10, 0, 0.5},
	}
	for _, tc := range cases {
		got, err := LRSchedule(tc.kind, tc.base, tc.step, tc.a, tc.b)
		if err != nil {
			t.Fatalf("LRSchedule(%s, %d): %v", tc.kind, tc.step, err)
		}
		if math.Abs(float64(got-tc.want)) > 1e-5 {
			t.Fatalf("LRSchedule(%s, %d) = %v, want %v", tc.kind, tc.step, got, tc.want)
		}
	}
}

func TestLRScheduleUnknownKind(t *testing.T) {
	if _, err := LRSchedule("nope", 1, 0, 1, 1); err == nil {
		t.Fatal("unknown schedule should error")
	}
}
