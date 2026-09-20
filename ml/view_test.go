package ml

import "testing"

// withNoGrad runs fn with tape recording off, so shape ops return views.
func withNoGrad(t *testing.T, fn func()) {
	t.Helper()
	prev := SetGradEnabled(false)
	defer SetGradEnabled(prev)
	fn()
}

func TestTransposeViewSharesStorage(t *testing.T) {
	withNoGrad(t, func() {
		x := dense([]float32{1, 2, 3, 4, 5, 6}, 2, 3)
		tr, err := Transpose(x, 0, 1)
		check(t, "transpose view", tr, err, []int{3, 2}, []float32{1, 4, 2, 5, 3, 6})
		// The view shares the base buffer: writing through the base is visible.
		x.Raw().AsFloat32()[0] = 100
		if got := tr.ContiguousData()[0]; got != 100 {
			t.Fatalf("view did not share storage: got %v", got)
		}
	})
}

func TestReshapeView(t *testing.T) {
	withNoGrad(t, func() {
		x := dense([]float32{1, 2, 3, 4, 5, 6}, 2, 3)
		v, err := Reshape(x, 3, 2)
		check(t, "reshape view", v, err, []int{3, 2}, []float32{1, 2, 3, 4, 5, 6})
		// A contiguous reshape view shares storage.
		x.Raw().AsFloat32()[5] = 99
		if got := v.ContiguousData()[5]; got != 99 {
			t.Fatalf("reshape view did not share storage: got %v", got)
		}
	})
}

func TestSliceView(t *testing.T) {
	withNoGrad(t, func() {
		x := dense([]float32{1, 2, 3, 4, 5, 6}, 2, 3)
		v, err := Slice(x, 1, 1, 3, 1)
		check(t, "slice view", v, err, []int{2, 2}, []float32{2, 3, 5, 6})
		// Inner-dim slice is non-contiguous, so it materializes on read.
		if v.isRawContiguous() {
			t.Fatal("inner-dim slice should be non-contiguous")
		}
	})
}

func TestUnsqueezeSqueezeView(t *testing.T) {
	withNoGrad(t, func() {
		x := dense([]float32{1, 2, 3}, 3)
		u, err := Unsqueeze(x, 0)
		check(t, "unsqueeze view", u, err, []int{1, 3}, []float32{1, 2, 3})
		s, err := Squeeze(u, []int{0})
		check(t, "squeeze view", s, err, []int{3}, []float32{1, 2, 3})
	})
}

func TestOpOnViewMaterializes(t *testing.T) {
	withNoGrad(t, func() {
		x := dense([]float32{1, 2, 3, 4, 5, 6}, 2, 3)
		tr, err := Transpose(x, 0, 1) // [3,2], non-contiguous
		if err != nil {
			t.Fatalf("Transpose: %v", err)
		}
		s, err := Sum(tr, nil, false)
		if err != nil {
			t.Fatalf("Sum: %v", err)
		}
		if got := mustItem(t, s); got != 21 {
			t.Fatalf("sum over a view = %v, want 21", got)
		}
	})
}

func TestContiguousMaterializes(t *testing.T) {
	withNoGrad(t, func() {
		x := dense([]float32{1, 2, 3, 4, 5, 6}, 2, 3)
		tr, err := Transpose(x, 0, 1)
		if err != nil {
			t.Fatalf("Transpose: %v", err)
		}
		c := tr.Contiguous()
		if !c.isRawContiguous() {
			t.Fatal("Contiguous did not produce contiguous storage")
		}
		check(t, "contiguous values", c, nil, []int{3, 2}, []float32{1, 4, 2, 5, 3, 6})
	})
}

func TestViewsRequireNoGrad(t *testing.T) {
	// Under recording, View is refused and shape ops materialize.
	x := dense([]float32{1, 2, 3, 4}, 2, 2)
	if _, err := View(x, []int{4}); err == nil {
		t.Fatal("View should error while a gradient is recorded")
	}
}

func TestRecordingShapeOpsStillDifferentiate(t *testing.T) {
	// The training path is unchanged: transpose materializes and records, so
	// gradients still flow.
	x := dense([]float32{1, 2, 3, 4, 5, 6}, 2, 3)
	x.SetRequiresGrad(true)
	tr, err := Transpose(x, 0, 1)
	if err != nil {
		t.Fatalf("Transpose: %v", err)
	}
	s, err := Sum(tr, nil, false)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	if err := s.Backward(); err != nil {
		t.Fatalf("Backward: %v", err)
	}
	check(t, "transpose grad", x.Grad(), nil, []int{2, 3}, []float32{1, 1, 1, 1, 1, 1})
}
