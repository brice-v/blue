package ml

import (
	"math"
	"slices"
	"testing"
)

// eps is the tolerance for float32 comparisons.
const eps = 1e-5

// dense builds a packed, offset-0 tensor for tests.
func dense(data []float32, shape ...int) *Tensor {
	return &Tensor{
		data:    slices.Clone(data),
		shape:   slices.Clone(shape),
		strides: contiguousStrides(shape),
		offset:  0,
		dtype:   Float32,
		device:  CPU,
	}
}

// view builds an arbitrary strided view, used for non-contiguous inputs.
func view(data []float32, shape, strides []int, offset int) *Tensor {
	return &Tensor{
		data:    data,
		shape:   shape,
		strides: strides,
		offset:  offset,
		dtype:   Float32,
		device:  CPU,
	}
}

// scalar is a 0-D tensor, which is how ops receive a scalar operand.
func scalar(v float32) *Tensor { return dense([]float32{v}) }

// check asserts shape and logical (materialized) values.
func check(t *testing.T, name string, got *Tensor, err error, wantShape []int, wantData []float32) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", name, err)
	}
	if got == nil {
		t.Fatalf("%s: got nil tensor, want shape=%v data=%v", name, wantShape, wantData)
	}
	if !slices.Equal(got.shape, wantShape) {
		t.Fatalf("%s: shape = %v, want %v", name, got.shape, wantShape)
	}
	m := got.materialize()
	if len(m.data) != len(wantData) {
		t.Fatalf("%s: len(data) = %d, want %d (got %v)", name, len(m.data), len(wantData), m.data)
	}
	for i := range wantData {
		if math.Abs(float64(m.data[i]-wantData[i])) > eps {
			t.Fatalf("%s: data[%d] = %v, want %v (got %v)", name, i, m.data[i], wantData[i], m.data)
		}
	}
}

// checkErr asserts the op rejected the input.
func checkErr(t *testing.T, name string, got *Tensor, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected an error, got tensor %v", name, got)
	}
}

func TestDefaultBackend(t *testing.T) {
	if DefaultBackend == nil {
		t.Fatal("DefaultBackend is nil")
	}
	var _ Backend = CPUBackend{}
}

func TestMatMul(t *testing.T) {
	be := CPUBackend{}

	t.Run("2x2", func(t *testing.T) {
		a := dense([]float32{1, 2, 3, 4}, 2, 2)
		b := dense([]float32{5, 6, 7, 8}, 2, 2)
		got, err := be.MatMul(a, b)
		check(t, "matmul", got, err, []int{2, 2}, []float32{19, 22, 43, 50})
	})

	t.Run("non-contiguous left", func(t *testing.T) {
		// Logical [[1,4],[2,5],[3,6]], a transposed view of [1..6].
		aT := view([]float32{1, 2, 3, 4, 5, 6}, []int{3, 2}, []int{1, 3}, 0)
		b := dense([]float32{7, 8, 9, 10}, 2, 2)
		got, err := be.MatMul(aT, b)
		check(t, "matmul non-contiguous", got, err, []int{3, 2}, []float32{43, 48, 59, 66, 75, 84})
	})

	t.Run("inner dim mismatch", func(t *testing.T) {
		a := dense([]float32{1, 2, 3, 4, 5, 6}, 2, 3)
		b := dense([]float32{1, 2, 3, 4}, 2, 2)
		got, err := be.MatMul(a, b)
		checkErr(t, "matmul mismatch", got, err)
	})
}

func TestElementwise(t *testing.T) {
	be := CPUBackend{}
	a := dense([]float32{1, 2, 3, 4}, 2, 2)
	b := dense([]float32{10, 20, 30, 40}, 2, 2)

	cases := []struct {
		name string
		fn   func(*Tensor, *Tensor) (*Tensor, error)
		want []float32
	}{
		{"add", be.Add, []float32{11, 22, 33, 44}},
		{"sub", be.Sub, []float32{-9, -18, -27, -36}},
		{"mul", be.Mul, []float32{10, 40, 90, 160}},
		{"div", be.Div, []float32{0.1, 0.1, 0.1, 0.1}},
	}
	for _, tc := range cases {
		got, err := tc.fn(a, b)
		check(t, tc.name, got, err, []int{2, 2}, tc.want)
	}
}

func TestElementwiseScalar(t *testing.T) {
	be := CPUBackend{}
	a := dense([]float32{1, 2, 3, 4}, 2, 2)
	s := scalar(2)

	cases := []struct {
		name string
		fn   func(*Tensor, *Tensor) (*Tensor, error)
		want []float32
	}{
		{"add scalar", be.Add, []float32{3, 4, 5, 6}},
		{"sub scalar", be.Sub, []float32{-1, 0, 1, 2}},
		{"mul scalar", be.Mul, []float32{2, 4, 6, 8}},
		{"div scalar", be.Div, []float32{0.5, 1, 1.5, 2}},
	}
	for _, tc := range cases {
		got, err := tc.fn(a, s)
		check(t, tc.name, got, err, []int{2, 2}, tc.want)
	}
}

func TestElementwiseNonContiguous(t *testing.T) {
	be := CPUBackend{}
	// Logical [[1,4],[2,5],[3,6]].
	aT := view([]float32{1, 2, 3, 4, 5, 6}, []int{3, 2}, []int{1, 3}, 0)
	ones := dense([]float32{1, 1, 1, 1, 1, 1}, 3, 2)
	got, err := be.Add(aT, ones)
	check(t, "add non-contiguous", got, err, []int{3, 2}, []float32{2, 5, 3, 6, 4, 7})
}

func TestElementwiseErrors(t *testing.T) {
	be := CPUBackend{}

	t.Run("shape mismatch", func(t *testing.T) {
		a := dense([]float32{1, 2, 3}, 3)
		b := dense([]float32{1, 2}, 2)
		got, err := be.Add(a, b)
		checkErr(t, "add mismatch", got, err)
	})

	t.Run("div by zero scalar", func(t *testing.T) {
		a := dense([]float32{1, 2}, 2)
		got, err := be.Div(a, scalar(0))
		checkErr(t, "div by zero", got, err)
	})
}

func TestUnary(t *testing.T) {
	be := CPUBackend{}
	a := dense([]float32{1, 2, 4}, 3)

	got, err := be.Exp(a)
	check(t, "exp", got, err, []int{3},
		[]float32{float32(math.Exp(1)), float32(math.Exp(2)), float32(math.Exp(4))})

	got, err = be.Log(a)
	check(t, "log", got, err, []int{3},
		[]float32{float32(math.Log(1)), float32(math.Log(2)), float32(math.Log(4))})

	got, err = be.Sqrt(a)
	check(t, "sqrt", got, err, []int{3},
		[]float32{float32(math.Sqrt(1)), float32(math.Sqrt(2)), float32(math.Sqrt(4))})
}

func TestSum(t *testing.T) {
	be := CPUBackend{}
	a := dense([]float32{1, 2, 3, 4, 5, 6}, 2, 3)

	t.Run("dim 0", func(t *testing.T) {
		got, err := be.Sum(a, 0, false)
		check(t, "sum dim0", got, err, []int{3}, []float32{5, 7, 9})
	})
	t.Run("dim 1", func(t *testing.T) {
		got, err := be.Sum(a, 1, false)
		check(t, "sum dim1", got, err, []int{2}, []float32{6, 15})
	})
	t.Run("dim 0 keepdim", func(t *testing.T) {
		got, err := be.Sum(a, 0, true)
		check(t, "sum dim0 keepdim", got, err, []int{1, 3}, []float32{5, 7, 9})
	})
	t.Run("dim 1 keepdim", func(t *testing.T) {
		got, err := be.Sum(a, 1, true)
		check(t, "sum dim1 keepdim", got, err, []int{2, 1}, []float32{6, 15})
	})
}

func TestMax(t *testing.T) {
	be := CPUBackend{}
	a := dense([]float32{1, 2, 3, 4, 5, 6}, 2, 3)

	t.Run("dim 0", func(t *testing.T) {
		got, err := be.Max(a, 0, false)
		check(t, "max dim0", got, err, []int{3}, []float32{4, 5, 6})
	})
	t.Run("dim 1", func(t *testing.T) {
		got, err := be.Max(a, 1, false)
		check(t, "max dim1", got, err, []int{2}, []float32{3, 6})
	})
	t.Run("dim 0 keepdim", func(t *testing.T) {
		got, err := be.Max(a, 0, true)
		check(t, "max dim0 keepdim", got, err, []int{1, 3}, []float32{4, 5, 6})
	})
	t.Run("dim 1 keepdim", func(t *testing.T) {
		got, err := be.Max(a, 1, true)
		check(t, "max dim1 keepdim", got, err, []int{2, 1}, []float32{3, 6})
	})
}

func TestSoftmax(t *testing.T) {
	be := CPUBackend{}

	t.Run("single row", func(t *testing.T) {
		a := dense([]float32{1, 2, 3}, 1, 3)
		got, err := be.Softmax(a, 1)
		check(t, "softmax", got, err, []int{1, 3},
			[]float32{0.09003057, 0.24472847, 0.66524096})
	})

	t.Run("shift invariant", func(t *testing.T) {
		// Same result as [1,2,3]; a naive exp() overflows here.
		a := dense([]float32{1000, 1001, 1002}, 1, 3)
		got, err := be.Softmax(a, 1)
		check(t, "softmax shift", got, err, []int{1, 3},
			[]float32{0.09003057, 0.24472847, 0.66524096})
	})

	t.Run("two rows", func(t *testing.T) {
		a := dense([]float32{1, 2, 3, 10, 10, 10}, 2, 3)
		got, err := be.Softmax(a, 1)
		check(t, "softmax rows", got, err, []int{2, 3},
			[]float32{0.09003057, 0.24472847, 0.66524096, 1.0 / 3, 1.0 / 3, 1.0 / 3})
	})
}

func TestReshape(t *testing.T) {
	be := CPUBackend{}
	a := dense([]float32{1, 2, 3, 4, 5, 6}, 2, 3)

	t.Run("3x2", func(t *testing.T) {
		got, err := be.Reshape(a, 3, 2)
		check(t, "reshape 3x2", got, err, []int{3, 2}, []float32{1, 2, 3, 4, 5, 6})
	})

	t.Run("flat", func(t *testing.T) {
		got, err := be.Reshape(a, 6)
		check(t, "reshape flat", got, err, []int{6}, []float32{1, 2, 3, 4, 5, 6})
	})

	t.Run("contiguous is a view", func(t *testing.T) {
		got, err := be.Reshape(a, 3, 2)
		if err != nil || got == nil {
			t.Fatalf("reshape failed: %v", err)
		}
		if len(got.data) == 0 || &got.data[0] != &a.data[0] {
			t.Fatal("reshape of a contiguous tensor should be a view sharing data")
		}
	})

	t.Run("numel mismatch", func(t *testing.T) {
		got, err := be.Reshape(a, 2, 2)
		checkErr(t, "reshape mismatch", got, err)
	})

	t.Run("non-contiguous", func(t *testing.T) {
		// Logical [[1,4],[2,5],[3,6]], flattened in row-major order.
		aT := view([]float32{1, 2, 3, 4, 5, 6}, []int{3, 2}, []int{1, 3}, 0)
		got, err := be.Reshape(aT, 6)
		check(t, "reshape non-contiguous", got, err, []int{6}, []float32{1, 4, 2, 5, 3, 6})
	})
}

func TestTranspose(t *testing.T) {
	be := CPUBackend{}
	a := dense([]float32{1, 2, 3, 4, 5, 6}, 2, 3)

	t.Run("2x3 to 3x2", func(t *testing.T) {
		got, err := be.Transpose(a, 0, 1)
		check(t, "transpose", got, err, []int{3, 2}, []float32{1, 4, 2, 5, 3, 6})
	})

	t.Run("is a view", func(t *testing.T) {
		got, err := be.Transpose(a, 0, 1)
		if err != nil || got == nil {
			t.Fatalf("transpose failed: %v", err)
		}
		if len(got.data) == 0 || &got.data[0] != &a.data[0] {
			t.Fatal("transpose should share data with its input")
		}
	})

	t.Run("bad dim", func(t *testing.T) {
		got, err := be.Transpose(a, 0, 5)
		checkErr(t, "transpose bad dim", got, err)
	})
}
