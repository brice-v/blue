package ml

import (
	"math"
	"testing"
)

func TestMaxDim(t *testing.T) {
	x := dense([]float32{1, 3, 2, 4, 0, 5}, 2, 3)
	v, i, err := MaxDim(x, 1, false)
	if err != nil {
		t.Fatalf("MaxDim: %v", err)
	}
	check(t, "max values", v, nil, []int{2}, []float32{3, 5})
	check(t, "max indices", i, nil, []int{2}, []float32{1, 2})
}

func TestMinDim(t *testing.T) {
	x := dense([]float32{1, 3, 2, 4, 0, 5}, 2, 3)
	v, i, err := MinDim(x, 1, false)
	if err != nil {
		t.Fatalf("MinDim: %v", err)
	}
	check(t, "min values", v, nil, []int{2}, []float32{1, 0})
	check(t, "min indices", i, nil, []int{2}, []float32{0, 1})
}

func TestVarianceUnbiased(t *testing.T) {
	x := dense([]float32{1, 2, 3}, 1, 3)
	pop, err := Variance(x, []int{1}, false, false)
	if err != nil {
		t.Fatalf("Variance pop: %v", err)
	}
	if math.Abs(float64(mustItem(t, pop))-0.6666667) > 1e-5 {
		t.Fatalf("population variance = %v", mustItem(t, pop))
	}
	unb, err := Variance(x, []int{1}, false, true)
	if err != nil {
		t.Fatalf("Variance unbiased: %v", err)
	}
	if math.Abs(float64(mustItem(t, unb))-1.0) > 1e-5 {
		t.Fatalf("unbiased variance = %v", mustItem(t, unb))
	}
}

func TestTile(t *testing.T) {
	x := dense([]float32{1, 2}, 1, 2)
	out, err := Tile(x, []int{2, 2})
	check(t, "tile", out, err, []int{2, 4}, []float32{1, 2, 1, 2, 1, 2, 1, 2})
}

func TestCumsum(t *testing.T) {
	x := dense([]float32{1, 2, 3}, 3)
	out, err := Cumsum(x, 0)
	check(t, "cumsum 1d", out, err, []int{3}, []float32{1, 3, 6})

	m := dense([]float32{1, 2, 3, 4, 5, 6}, 2, 3)
	out, err = Cumsum(m, 1)
	check(t, "cumsum dim1", out, err, []int{2, 3}, []float32{1, 3, 6, 4, 9, 15})
	out, err = Cumsum(m, 0)
	check(t, "cumsum dim0", out, err, []int{2, 3}, []float32{1, 2, 3, 5, 7, 9})
}

func TestCumsumBackward(t *testing.T) {
	x := dense([]float32{1, 2, 3}, 3)
	x.SetRequiresGrad(true)
	s, err := Sum(mustCumsum(t, x), nil, false)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	if err := s.Backward(); err != nil {
		t.Fatalf("Backward: %v", err)
	}
	// d(sum(cumsum(x)))/dx = [3, 2, 1]
	check(t, "cumsum grad", x.Grad(), nil, []int{3}, []float32{3, 2, 1})
}

func TestSort(t *testing.T) {
	x := dense([]float32{3, 1, 2}, 3)
	v, i, err := Sort(x, 0, false)
	if err != nil {
		t.Fatalf("Sort: %v", err)
	}
	check(t, "sort values", v, nil, []int{3}, []float32{1, 2, 3})
	check(t, "sort indices", i, nil, []int{3}, []float32{1, 2, 0})
	v, _, err = Sort(x, 0, true)
	check(t, "sort descending", v, err, []int{3}, []float32{3, 2, 1})
}

func TestTopK(t *testing.T) {
	x := dense([]float32{3, 1, 2}, 3)
	v, i, err := TopK(x, 2, 0, true)
	if err != nil {
		t.Fatalf("TopK: %v", err)
	}
	check(t, "topk values", v, nil, []int{2}, []float32{3, 2})
	check(t, "topk indices", i, nil, []int{2}, []float32{0, 2})
}

func TestNonzero(t *testing.T) {
	x := dense([]float32{0, 2, 0, 3}, 4)
	out, err := Nonzero(x)
	check(t, "nonzero", out, err, []int{2, 1}, []float32{1, 3})
}

func TestScatterAdd(t *testing.T) {
	dest := dense([]float32{0, 0, 0, 0, 0, 0}, 3, 2)
	idx := mustIntTensor(t, []int32{0, 1}, []int{1, 2})
	src := dense([]float32{1, 1}, 1, 2)
	out, err := ScatterAdd(dest, 0, idx, src)
	check(t, "scatter_add", out, err, []int{3, 2}, []float32{1, 0, 0, 1, 0, 0})
}

func TestCrossEntropyWithReduction(t *testing.T) {
	logits := dense([]float32{1, 2, 3, 1, 2, 3}, 2, 3)
	target := mustIntTensor(t, []int32{2, 2}, []int{2})

	mean, err := CrossEntropyWith(logits, target, CrossEntropyOpts{Reduction: "mean"})
	if err != nil {
		t.Fatalf("mean: %v", err)
	}
	if math.Abs(float64(mustItem(t, mean))-0.40760595) > 1e-5 {
		t.Fatalf("mean = %v", mustItem(t, mean))
	}
	sum, err := CrossEntropyWith(logits, target, CrossEntropyOpts{Reduction: "sum"})
	if err != nil {
		t.Fatalf("sum: %v", err)
	}
	if math.Abs(float64(mustItem(t, sum))-0.8152119) > 1e-5 {
		t.Fatalf("sum = %v", mustItem(t, sum))
	}
	none, err := CrossEntropyWith(logits, target, CrossEntropyOpts{Reduction: "none"})
	check(t, "none", none, err, []int{2}, []float32{0.40760595, 0.40760595})
}

func TestCrossEntropyWithIgnore(t *testing.T) {
	logits := dense([]float32{1, 2, 3, 1, 2, 3}, 2, 3)
	target := mustIntTensor(t, []int32{2, 0}, []int{2})
	// Row 0 is ignored, so the mean is just row 1: -log_softmax([1,2,3])[0] = 2.4076.
	got, err := CrossEntropyWith(logits, target, CrossEntropyOpts{Reduction: "mean", HasIgnore: true, IgnoreIndex: 2})
	if err != nil {
		t.Fatalf("ignore: %v", err)
	}
	if math.Abs(float64(mustItem(t, got))-2.4076059) > 1e-4 {
		t.Fatalf("ignore mean = %v", mustItem(t, got))
	}
}

func TestCrossEntropyWithWeight(t *testing.T) {
	logits := dense([]float32{1, 2, 3}, 1, 3)
	target := mustIntTensor(t, []int32{2}, []int{1})
	w := dense([]float32{1, 1, 2}, 3)
	got, err := CrossEntropyWith(logits, target, CrossEntropyOpts{Reduction: "mean", Weight: w})
	if err != nil {
		t.Fatalf("weight: %v", err)
	}
	if math.Abs(float64(mustItem(t, got))-0.8152119) > 1e-5 {
		t.Fatalf("weighted = %v", mustItem(t, got))
	}
}

func mustItem(t *testing.T, x *Tensor) float32 {
	t.Helper()
	v, err := x.Item()
	if err != nil {
		t.Fatalf("Item: %v", err)
	}
	return v
}

func mustCumsum(t *testing.T, x *Tensor) *Tensor {
	t.Helper()
	out, err := Cumsum(x, 0)
	if err != nil {
		t.Fatalf("Cumsum: %v", err)
	}
	return out
}
