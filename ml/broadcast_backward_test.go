package ml

import (
	"math"
	"testing"
)

// TestBroadcastMulBackward checks the gradient of x * scale where scale is
// broadcast from [1, N] to [M, N], the shape rmsnorm uses. A wrong reduction
// here would let any broadcast model train in the wrong direction.
func TestBroadcastMulBackward(t *testing.T) {
	M, N := 3, 4
	xData := []float32{0.5, -1.0, 2.0, 0.25, 1.5, 0.75, -0.5, 1.0, -2.0, 0.3, 0.9, -1.2}
	sData := []float32{1.0, 2.0, -0.5, 3.0}

	x := mustTensor(t, xData, []int{M, N})
	s := mustTensor(t, sData, []int{1, N})
	x.SetRequiresGrad(true)
	s.SetRequiresGrad(true)

	out, err := Mul(x, s)
	if err != nil {
		t.Fatalf("Mul: %v", err)
	}
	loss, err := Sum(out, nil, false)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	if err := loss.Backward(); err != nil {
		t.Fatalf("Backward: %v", err)
	}

	// d(sum(x*s))/dx = s broadcast; d/ds = sum over rows of x.
	wantX := make([]float32, M*N)
	for i := 0; i < M; i++ {
		for j := 0; j < N; j++ {
			wantX[i*N+j] = sData[j]
		}
	}
	wantS := make([]float32, N)
	for j := 0; j < N; j++ {
		var acc float32
		for i := 0; i < M; i++ {
			acc += xData[i*N+j]
		}
		wantS[j] = acc
	}

	if !closeData(wantX, x.Grad().ContiguousData()) {
		t.Fatalf("grad x mismatch:\n want %v\n got  %v", wantX, x.Grad().ContiguousData())
	}
	if !closeData(wantS, s.Grad().ContiguousData()) {
		t.Fatalf("grad scale mismatch:\n want %v\n got  %v", wantS, s.Grad().ContiguousData())
	}
}

// TestScalarDivBackward checks dividing by a [1] scalar broadcasts and reduces
// correctly, the case that motivated the reduceBroadcast fix.
func TestScalarDivBackward(t *testing.T) {
	M, N := 2, 3
	xData := []float32{1, 2, 3, 4, 5, 6}
	x := mustTensor(t, xData, []int{M, N})
	d := mustTensor(t, []float32{2}, []int{1})
	x.SetRequiresGrad(true)
	d.SetRequiresGrad(true)

	q, err := Div(x, d)
	if err != nil {
		t.Fatalf("Div: %v", err)
	}
	loss, err := Sum(q, nil, false)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	if err := loss.Backward(); err != nil {
		t.Fatalf("Backward: %v", err)
	}

	wantX := make([]float32, M*N)
	for i := range wantX {
		wantX[i] = 0.5
	}
	if !closeData(wantX, x.Grad().ContiguousData()) {
		t.Fatalf("grad x mismatch:\n want %v\n got  %v", wantX, x.Grad().ContiguousData())
	}
	// d/dd sum(x/d) = -sum(x)/d^2 = -21/4.
	want := float32(-21.0 / 4.0)
	if math.Abs(float64(d.Grad().ContiguousData()[0]-want)) > 1e-5 {
		t.Fatalf("grad divisor = %v, want %v", d.Grad().ContiguousData()[0], want)
	}
}
