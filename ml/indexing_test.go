package ml

import "testing"

func TestSliceStep(t *testing.T) {
	x := dense([]float32{0, 1, 2, 3, 4, 5}, 6)
	out, err := Slice(x, 0, 1, 6, 2)
	check(t, "slice step 2", out, err, []int{3}, []float32{1, 3, 5})
}

func TestSliceNegativeStep(t *testing.T) {
	x := dense([]float32{0, 1, 2, 3, 4, 5}, 6)
	out, err := Slice(x, 0, 5, -1, -1)
	check(t, "slice reverse", out, err, []int{6}, []float32{5, 4, 3, 2, 1, 0})
}

func TestSliceNegativeBounds(t *testing.T) {
	x := dense([]float32{0, 1, 2, 3, 4, 5}, 6)
	out, err := Slice(x, 0, -2, 6, 1)
	check(t, "slice negative start", out, err, []int{2}, []float32{4, 5})
}

func TestSliceStepZero(t *testing.T) {
	x := dense([]float32{1, 2, 3}, 3)
	if _, err := Slice(x, 0, 0, 3, 0); err == nil {
		t.Fatal("Slice with step 0 should error")
	}
}

func TestSelect(t *testing.T) {
	x := dense([]float32{1, 2, 3, 4, 5, 6}, 2, 3)
	out, err := Select(x, 0, 1)
	check(t, "select row", out, err, []int{3}, []float32{4, 5, 6})
}

func TestSelectNegative(t *testing.T) {
	x := dense([]float32{1, 2, 3, 4, 5, 6}, 2, 3)
	out, err := Select(x, 1, -1)
	check(t, "select last col", out, err, []int{2}, []float32{3, 6})
}

func TestIndexSelect(t *testing.T) {
	x := dense([]float32{1, 2, 3, 4, 5, 6}, 3, 2)
	idx := mustIntTensor(t, []int32{2, 0}, []int{2})
	out, err := IndexSelect(x, 0, idx)
	check(t, "index_select", out, err, []int{2, 2}, []float32{5, 6, 1, 2})
}

func TestFlip(t *testing.T) {
	x := dense([]float32{1, 2, 3, 4, 5, 6}, 2, 3)
	out, err := Flip(x, []int{0})
	check(t, "flip rows", out, err, []int{2, 3}, []float32{4, 5, 6, 1, 2, 3})
	out, err = Flip(x, []int{1})
	check(t, "flip cols", out, err, []int{2, 3}, []float32{3, 2, 1, 6, 5, 4})
}

func TestMaskedSelect(t *testing.T) {
	x := dense([]float32{1, 2, 3, 4}, 2, 2)
	m := dense([]float32{1, 0, 0, 1}, 2, 2)
	out, err := MaskedSelect(x, m)
	check(t, "masked_select", out, err, []int{2}, []float32{1, 4})
}

func TestMaskedFill(t *testing.T) {
	x := dense([]float32{1, 2, 3, 4}, 2, 2)
	mask := dense([]float32{1, 0, 0, 1}, 2, 2)
	out, err := MaskedFill(x, mask, 0)
	check(t, "masked_fill", out, err, []int{2, 2}, []float32{0, 2, 3, 0})
}

func TestEmbedding(t *testing.T) {
	w := dense([]float32{1, 2, 3, 4, 5, 6}, 3, 2)
	idx := mustIntTensor(t, []int32{2, 0}, []int{2})
	out, err := Embedding(w, idx)
	check(t, "embedding", out, err, []int{2, 2}, []float32{5, 6, 1, 2})
}

func TestEmbeddingBackward(t *testing.T) {
	w := dense([]float32{1, 2, 3, 4, 5, 6}, 3, 2)
	w.SetRequiresGrad(true)
	idx := mustIntTensor(t, []int32{2, 0}, []int{2})
	emb, err := Embedding(w, idx)
	if err != nil {
		t.Fatalf("Embedding() error: %v", err)
	}
	loss, err := Sum(emb, nil, false)
	if err != nil {
		t.Fatalf("Sum() error: %v", err)
	}
	if err := loss.Backward(); err != nil {
		t.Fatalf("Backward() error: %v", err)
	}
	// Rows 0 and 2 are touched once each, row 1 is untouched.
	check(t, "embedding grad", w.Grad(), nil, []int{3, 2}, []float32{1, 1, 0, 0, 1, 1})
}

func TestBMM(t *testing.T) {
	a := dense([]float32{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, 2, 2, 3)
	b := dense([]float32{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, 2, 3, 2)
	out, err := BMM(a, b)
	check(t, "bmm", out, err, []int{2, 2, 2}, []float32{3, 3, 3, 3, 3, 3, 3, 3})
}

func TestSilu(t *testing.T) {
	x := dense([]float32{0, 1}, 2)
	out, err := Silu(x)
	check(t, "silu", out, err, []int{2}, []float32{0, 0.7310586})
}

func TestRandPermIsPermutation(t *testing.T) {
	ManualSeed(7)
	p, err := RandPerm(8, CPU)
	if err != nil {
		t.Fatalf("RandPerm() error: %v", err)
	}
	seen := map[int]bool{}
	for _, v := range p.ContiguousData() {
		i := int(v)
		if i < 0 || i >= 8 || seen[i] {
			t.Fatalf("RandPerm produced %v, not a permutation", p.ContiguousData())
		}
		seen[i] = true
	}
}

func TestRandPermDeterministic(t *testing.T) {
	ManualSeed(11)
	a, _ := RandPerm(6, CPU)
	ManualSeed(11)
	b, _ := RandPerm(6, CPU)
	check(t, "randperm deterministic", b, nil, []int{6}, a.ContiguousData())
}

func TestShuffleRowsAreAPermutation(t *testing.T) {
	ManualSeed(3)
	x := dense([]float32{0, 1, 2, 3, 4, 5, 6, 7}, 4, 2)
	out, err := Shuffle(x, 0)
	if err != nil {
		t.Fatalf("Shuffle() error: %v", err)
	}
	// Every original row must appear exactly once: check the first column.
	seen := map[float32]bool{}
	got := []float32{out.ContiguousData()[0], out.ContiguousData()[2], out.ContiguousData()[4], out.ContiguousData()[6]}
	for _, v := range got {
		if seen[v] {
			t.Fatalf("Shuffle duplicated a row: %v", got)
		}
		seen[v] = true
	}
	if len(seen) != 4 {
		t.Fatalf("Shuffle lost a row: %v", got)
	}
}
