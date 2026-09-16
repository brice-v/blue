package ml

import (
	"slices"
	"testing"
)

func TestNumel(t *testing.T) {
	cases := []struct {
		name  string
		shape []int
		want  int
	}{
		{"0-D nil shape", nil, 1},
		{"0-D empty shape", []int{}, 1},
		{"size-1", []int{1}, 1},
		{"all size-1", []int{1, 1, 1}, 1},
		{"matrix", []int{2, 3}, 6},
		{"3-D", []int{2, 3, 4}, 24},
		{"zero dim", []int{0}, 0},
		{"empty batch", []int{0, 3}, 0},
		{"zero middle", []int{2, 0, 4}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := dense(nil, tc.shape...).Numel()
			if got != tc.want {
				t.Fatalf("Numel() = %d, want %d", got, tc.want)
			}
			if n := numelOf(tc.shape); n != tc.want {
				t.Fatalf("numelOf(%v) = %d, want %d", tc.shape, n, tc.want)
			}
		})
	}
}

func TestShapeAndStrides(t *testing.T) {
	a := dense([]float32{1, 2, 3, 4, 5, 6}, 2, 3)
	if got := a.Shape(); !slices.Equal(got, []int{2, 3}) {
		t.Fatalf("Shape() = %v, want [2 3]", got)
	}
	if got := a.Strides(); !slices.Equal(got, []int{3, 1}) {
		t.Fatalf("Strides() = %v, want [3 1]", got)
	}

	b := dense(nil, 2, 3, 4)
	if got := b.Strides(); !slices.Equal(got, []int{12, 4, 1}) {
		t.Fatalf("Strides() = %v, want [12 4 1]", got)
	}
}

func TestContiguousStrides(t *testing.T) {
	cases := []struct {
		shape []int
		want  []int
	}{
		{nil, []int{}},
		{[]int{5}, []int{1}},
		{[]int{2, 3}, []int{3, 1}},
		{[]int{2, 3, 4}, []int{12, 4, 1}},
	}
	for _, tc := range cases {
		got := getContiguousStridesFromShape(tc.shape)
		if !slices.Equal(got, tc.want) {
			t.Fatalf("contiguousStrides(%v) = %v, want %v", tc.shape, got, tc.want)
		}
	}
}

func TestIsContiguous(t *testing.T) {
	base := []float32{10, 20, 30, 40, 50, 60}

	cases := []struct {
		name string
		t    *Tensor
		want bool
	}{
		{"packed 2x3", dense(base, 2, 3), true},
		{"transposed view", view(base, []int{3, 2}, []int{1, 3}, 0), false},
		{"strided 1-D slice", view(base, []int{3}, []int{4}, 1), false},
		{"packed with offset", view(base, []int{3}, []int{1}, 2), true},
		{"size-1 dim is ignored", view([]float32{1, 2, 3}, []int{1, 3}, []int{999, 1}, 0), true},
		{"0-D scalar", view(base, nil, nil, 0), true},
		{"empty", view([]float32{}, []int{0, 3}, []int{3, 1}, 0), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.t.IsContiguous(); got != tc.want {
				t.Fatalf("IsContiguous() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestMaterialize(t *testing.T) {
	t.Run("contiguous offset 0 returns the same tensor", func(t *testing.T) {
		a := dense([]float32{1, 2, 3, 4}, 2, 2)
		got := a.materialize()
		if got != a {
			t.Fatal("expected materialize to return the same tensor, no copy")
		}
	})

	t.Run("contiguous with offset re-slices, no copy", func(t *testing.T) {
		base := []float32{10, 20, 30, 40}
		a := view(base, []int{2}, []int{1}, 1) // logical [20, 30]
		got := a.materialize()
		if got == a {
			t.Fatal("expected a new view when offset != 0")
		}
		if got.offset != 0 {
			t.Fatalf("offset = %d, want 0", got.offset)
		}
		if &got.data[0] != &base[1] {
			t.Fatal("expected the re-slice to share the original backing array")
		}
		if got.data[0] != 20 || got.data[1] != 30 {
			t.Fatalf("data = %v, want [20 30 ...]", got.data)
		}
	})

	t.Run("non-contiguous copies in row-major order", func(t *testing.T) {
		base := []float32{1, 2, 3, 4, 5, 6}
		aT := view(base, []int{3, 2}, []int{1, 3}, 0) // logical [[1,4],[2,5],[3,6]]
		got := aT.materialize()
		if !got.IsContiguous() || got.offset != 0 {
			t.Fatalf("expected a packed offset-0 result, got strides=%v offset=%d", got.strides, got.offset)
		}
		if &got.data[0] == &base[0] {
			t.Fatal("expected a fresh buffer for a non-contiguous input")
		}
		if !slices.Equal(got.data, []float32{1, 4, 2, 5, 3, 6}) {
			t.Fatalf("data = %v, want [1 4 2 5 3 6]", got.data)
		}
		if got.gradState != nil {
			t.Fatal("materialized copy should not carry graph state")
		}
	})

	t.Run("0-D scalar with offset", func(t *testing.T) {
		base := []float32{5, 6, 7}
		s := view(base, nil, nil, 2)
		got := s.materialize()
		if got.Numel() != 1 || got.data[0] != 7 {
			t.Fatalf("scalar materialize = %v, want 7", got.data)
		}
	})

	t.Run("empty stays empty", func(t *testing.T) {
		a := view([]float32{}, []int{0, 3}, []int{3, 1}, 0)
		got := a.materialize()
		if got.Numel() != 0 {
			t.Fatalf("Numel() = %d, want 0", got.Numel())
		}
	})
}

func TestCopyStrided(t *testing.T) {
	t.Run("transposed", func(t *testing.T) {
		src := view([]float32{1, 2, 3, 4, 5, 6}, []int{3, 2}, []int{1, 3}, 0)
		dst := make([]float32, 6)
		copyStrided(dst, src)
		if !slices.Equal(dst, []float32{1, 4, 2, 5, 3, 6}) {
			t.Fatalf("dst = %v, want [1 4 2 5 3 6]", dst)
		}
	})

	t.Run("0-D", func(t *testing.T) {
		src := view([]float32{9, 8, 7}, nil, nil, 2)
		dst := make([]float32, 1)
		copyStrided(dst, src)
		if dst[0] != 7 {
			t.Fatalf("dst[0] = %v, want 7", dst[0])
		}
	})
}

func TestItem(t *testing.T) {
	t.Run("single element", func(t *testing.T) {
		got, err := dense([]float32{42}, 1).Item()
		if err != nil {
			t.Fatalf("Item() unexpected error: %v", err)
		}
		if got != 42 {
			t.Fatalf("Item() = %v, want 42", got)
		}
	})

	t.Run("honours offset", func(t *testing.T) {
		a := view([]float32{5, 6, 7}, []int{1}, []int{1}, 2)
		got, err := a.Item()
		if err != nil {
			t.Fatalf("Item() unexpected error: %v", err)
		}
		if got != 7 {
			t.Fatalf("Item() = %v, want 7", got)
		}
	})

	t.Run("multi element errors", func(t *testing.T) {
		if _, err := dense([]float32{1, 2}, 2).Item(); err == nil {
			t.Fatal("expected Item() on a multi-element tensor to error")
		}
	})
}

func TestNewLike(t *testing.T) {
	a := dense([]float32{1, 2, 3, 4}, 2, 2)
	a.dtype = Float64
	a.device = GPU

	got := newLike(a)
	if !slices.Equal(got.shape, []int{2, 2}) {
		t.Fatalf("shape = %v, want [2 2]", got.shape)
	}
	if !slices.Equal(got.strides, []int{2, 1}) {
		t.Fatalf("strides = %v, want [2 1]", got.strides)
	}
	if got.offset != 0 {
		t.Fatalf("offset = %d, want 0", got.offset)
	}
	if got.dtype != Float64 || got.device != GPU {
		t.Fatalf("dtype/device = %v/%v, want Float64/GPU", got.dtype, got.device)
	}
	if len(got.data) != 4 {
		t.Fatalf("len(data) = %d, want 4", len(got.data))
	}
	for i, v := range got.data {
		if v != 0 {
			t.Fatalf("data[%d] = %v, want 0 (newLike must not copy input data)", i, v)
		}
	}
	if &got.data[0] == &a.data[0] {
		t.Fatal("newLike must allocate a fresh buffer")
	}
}

// TestAccessors encodes PyTorch's leaf semantics: a tensor is a leaf when
// requires_grad is false, or when it has no gradFn. Untracked tensors are
// leaves by convention, and a marked tensor with no gradFn is also a leaf.
func TestAccessors(t *testing.T) {
	t.Run("untracked tensor is a leaf", func(t *testing.T) {
		a := dense([]float32{1}, 1)
		if !a.IsLeaf() {
			t.Fatal("an untracked tensor should be a leaf")
		}
		if a.RequiresGrad() {
			t.Fatal("fresh tensor should not require grad")
		}
		if a.Grad() != nil {
			t.Fatal("fresh tensor should have no grad")
		}
	})

	t.Run("marked leaf still is a leaf", func(t *testing.T) {
		a := dense([]float32{1}, 1)
		a.SetRequiresGrad(true)
		if !a.RequiresGrad() {
			t.Fatal("RequiresGrad() = false after SetRequiresGrad(true)")
		}
		if !a.IsLeaf() {
			t.Fatal("a tensor with requires_grad and no gradFn should still be a leaf")
		}
	})

	t.Run("grad round trips", func(t *testing.T) {
		a := dense([]float32{1}, 1)
		a.SetRequiresGrad(true)
		if a.Grad() != nil {
			t.Fatal("Grad() should be nil before any accumulation")
		}
		g := dense([]float32{5}, 1)
		a.gradState.Grad = g
		if a.Grad() != g {
			t.Fatal("Grad() did not return the accumulated gradient")
		}
	})

	t.Run("tensor with gradFn is not a leaf", func(t *testing.T) {
		a := dense([]float32{1}, 1)
		a.gradState = &gradState{
			requiresGrad: true,
			gradFn:       func(*Tensor) []*Tensor { return nil },
		}
		if a.IsLeaf() {
			t.Fatal("a tensor with a gradFn should not be a leaf")
		}
	})

	t.Run("unset keeps the gradient", func(t *testing.T) {
		a := dense([]float32{1}, 1)
		a.SetRequiresGrad(true)
		a.gradState.Grad = dense([]float32{9}, 1)
		a.SetRequiresGrad(false)
		if a.RequiresGrad() {
			t.Fatal("RequiresGrad() = true after SetRequiresGrad(false)")
		}
		if a.Grad() == nil {
			t.Fatal("SetRequiresGrad(false) should preserve the accumulated grad")
		}
	})

	t.Run("unset on an untracked tensor is a no-op", func(t *testing.T) {
		a := dense([]float32{1}, 1)
		a.SetRequiresGrad(false)
		if a.gradState != nil {
			t.Fatal("SetRequiresGrad(false) allocated gradState on an untracked tensor")
		}
	})
}
