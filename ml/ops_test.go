package ml

import "testing"

func TestSumOp(t *testing.T) {
	x := dense([]float32{1, 2, 3}, 1, 3)
	x.SetRequiresGrad(true)

	out, err := Sum(x, []int{1}, false) // shape [1], numel 1
	check(t, "sum forward", out, err, []int{1}, []float32{6})

	if err := out.Backward(); err != nil {
		t.Fatalf("Backward() error: %v", err)
	}
	check(t, "sum grad", x.Grad(), nil, []int{1, 3}, []float32{1, 1, 1})
}

func TestSumOpKeepdim(t *testing.T) {
	x := dense([]float32{1, 2, 3, 4, 5, 6}, 2, 3)
	out, err := Sum(x, []int{0}, true)
	check(t, "sum keepdim", out, err, []int{1, 3}, []float32{5, 7, 9})
}

func TestMaxOp(t *testing.T) {
	// Max/Min are host-computed because borncgo has no max reduction, so this
	// checks the values only.
	x := dense([]float32{1, 3, 2}, 1, 3)
	out, err := Max(x, []int{1}, false)
	check(t, "max forward", out, err, []int{1}, []float32{3})
}

func TestSoftmaxOp(t *testing.T) {
	x := dense([]float32{1, 2, 3}, 1, 3)
	s, err := Softmax(x, 1)
	check(t, "softmax forward", s, err, []int{1, 3},
		[]float32{0.09003057, 0.24472847, 0.66524096})
}

// TestSoftmaxOpBackward chains softmax into mul+sum so the loss is scalar, then
// checks the analytic gradient s_i * (w_i - sum_k w_k s_k).
func TestSoftmaxOpBackward(t *testing.T) {
	x := dense([]float32{1, 2, 3}, 1, 3)
	x.SetRequiresGrad(true)

	s, err := Softmax(x, 1)
	if err != nil {
		t.Fatalf("Softmax() error: %v", err)
	}
	p, err := Mul(s, dense([]float32{1, 0, 0}, 1, 3))
	if err != nil {
		t.Fatalf("Mul() error: %v", err)
	}
	loss, err := Sum(p, []int{1}, false)
	if err != nil {
		t.Fatalf("Sum() error: %v", err)
	}
	if err := loss.Backward(); err != nil {
		t.Fatalf("Backward() error: %v", err)
	}
	check(t, "softmax grad", x.Grad(), nil, []int{1, 3},
		[]float32{0.08192507, -0.02203304, -0.05989203})
}

func TestEqOp(t *testing.T) {
	a := dense([]float32{1, 2, 3}, 3)
	b := dense([]float32{1, 0, 3}, 3)

	got, err := Eq(a, b)
	check(t, "eq", got, err, []int{3}, []float32{1, 0, 1})
	if got.DType() != Bool {
		t.Fatalf("Eq dtype = %s, want bool", got.DType())
	}

	got, err = Eq(a, scalar(2)) // scalar broadcast
	check(t, "eq scalar", got, err, []int{3}, []float32{0, 1, 0})

	// comparisons are not differentiable, so Eq must not build a graph node
	a.SetRequiresGrad(true)
	got, err = Eq(a, b)
	if err != nil {
		t.Fatalf("Eq() error: %v", err)
	}
	if got.RequiresGrad() {
		t.Fatal("Eq should not be tracked by autograd")
	}
}

func TestNegOp(t *testing.T) {
	x := dense([]float32{-1, 0, 2}, 3)
	got, err := Neg(x)
	check(t, "neg forward", got, err, []int{3}, []float32{1, 0, -2})

	// d(-x)/dx = -1, so a seed of 1 comes back as -1
	y := dense([]float32{-3}, 1, 1)
	y.SetRequiresGrad(true)
	out, err := Neg(y)
	if err != nil {
		t.Fatalf("Neg() error: %v", err)
	}
	if err := out.Backward(); err != nil {
		t.Fatalf("Backward() error: %v", err)
	}
	check(t, "neg grad", y.Grad(), nil, []int{1, 1}, []float32{-1})
}

func TestCompareOps(t *testing.T) {
	a := dense([]float32{1, 2, 3}, 3)
	b := dense([]float32{1, 0, 4}, 3)

	cases := []struct {
		name string
		fn   func(*Tensor, *Tensor) (*Tensor, error)
		want []float32
	}{
		{"eq", Eq, []float32{1, 0, 0}},
		{"ne", Ne, []float32{0, 1, 1}},
		{"gt", Gt, []float32{0, 1, 0}},
		{"ge", Ge, []float32{1, 1, 0}},
		{"lt", Lt, []float32{0, 0, 1}},
		{"le", Le, []float32{1, 0, 1}},
	}
	for _, tc := range cases {
		got, err := tc.fn(a, b)
		check(t, tc.name, got, err, []int{3}, tc.want)
		if got.DType() != Bool {
			t.Fatalf("%s dtype = %s, want bool", tc.name, got.DType())
		}

		// comparisons never build a graph, even when an input requires grad
		a.SetRequiresGrad(true)
		tracked, err := tc.fn(a, b)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if tracked.RequiresGrad() {
			t.Fatalf("%s should not be tracked by autograd", tc.name)
		}
		a.SetRequiresGrad(false)
	}

	// scalar broadcast on the left
	got, err := Ge(dense([]float32{2}, 1), a) // 2 >= [1, 2, 3]
	check(t, "ge scalar left", got, err, []int{3}, []float32{1, 1, 0})
	got, err = Ne(dense([]float32{2}, 1), a) // 2 != [1, 2, 3]
	check(t, "ne scalar left", got, err, []int{3}, []float32{1, 0, 1})
	got, err = Lt(dense([]float32{2}, 1), a) // 2 < [1, 2, 3]
	check(t, "lt scalar left", got, err, []int{3}, []float32{0, 0, 1})
	got, err = Le(dense([]float32{2}, 1), a) // 2 <= [1, 2, 3]
	check(t, "le scalar left", got, err, []int{3}, []float32{0, 1, 1})
}

func TestSumAllOp(t *testing.T) {
	x := dense([]float32{1, 2, 3, 4}, 2, 2)
	x.SetRequiresGrad(true)

	out, err := Sum(x, nil, false) // all dims -> 0-d
	check(t, "sum all", out, err, []int{}, []float32{10})

	if err := out.Backward(); err != nil {
		t.Fatalf("Backward() error: %v", err)
	}
	check(t, "sum all grad", x.Grad(), nil, []int{2, 2}, []float32{1, 1, 1, 1})
}

func TestSumMultiDimOp(t *testing.T) {
	x := dense([]float32{1, 2, 3, 4, 5, 6}, 2, 3)

	out, err := Sum(x, []int{0, 1}, false)
	check(t, "sum dims 0,1", out, err, []int{}, []float32{21})

	out, err = Sum(x, []int{1, 0}, true)
	check(t, "sum dims keepdim", out, err, []int{1, 1}, []float32{21})
}
