package ml

import (
	"math"
	"slices"
	"testing"
)

const eps = 1e-5

func dense(data []float32, shape ...int) *Tensor {
	t, err := NewTensor(data, shape, Float32, CPU)
	if err != nil {
		panic(err)
	}
	return t
}

func scalar(v float32) *Tensor {
	return dense([]float32{v}, 1)
}

func check(t *testing.T, name string, got *Tensor, err error, wantShape []int, wantData []float32) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", name, err)
	}
	if got == nil {
		t.Fatalf("%s: got nil tensor, want shape=%v data=%v", name, wantShape, wantData)
	}
	if !slices.Equal(got.Shape(), wantShape) {
		t.Fatalf("%s: shape = %v, want %v", name, got.Shape(), wantShape)
	}
	data := got.ContiguousData()
	if len(data) != len(wantData) {
		t.Fatalf("%s: len(data) = %d, want %d (got %v)", name, len(data), len(wantData), data)
	}
	for i := range wantData {
		if math.Abs(float64(data[i]-wantData[i])) > eps {
			t.Fatalf("%s: data[%d] = %v, want %v (got %v)", name, i, data[i], wantData[i], data)
		}
	}
}

func TestTensorBasics(t *testing.T) {
	x := dense([]float32{1, 2, 3, 4, 5, 6}, 2, 3)
	if !slices.Equal(x.Shape(), []int{2, 3}) {
		t.Fatalf("shape = %v", x.Shape())
	}
	if x.Numel() != 6 || !x.IsContiguous() {
		t.Fatalf("numel=%d contiguous=%v", x.Numel(), x.IsContiguous())
	}
	if x.DType() != Float32 || x.Device() != CPU {
		t.Fatalf("dtype=%s device=%s", x.DType(), x.Device())
	}
	if v, err := dense([]float32{7}, 1).Item(); err != nil || v != 7 {
		t.Fatalf("item = %v, %v", v, err)
	}
}

func TestMatMulAddDelegates(t *testing.T) {
	a := dense([]float32{1, 2, 3, 4, 5, 6}, 2, 3)
	b := dense([]float32{7, 8, 9, 10, 11, 12}, 3, 2)
	c, err := MatMul(a, b)
	check(t, "matmul", c, err, []int{2, 2}, []float32{58, 64, 139, 154})

	s, err := Add(c, scalar(1))
	check(t, "add scalar", s, err, []int{2, 2}, []float32{59, 65, 140, 155})
}

func TestAutogradMatMul(t *testing.T) {
	w := dense([]float32{2}, 1, 1)
	w.SetRequiresGrad(true)
	x := dense([]float32{3}, 1, 1)

	y, err := MatMul(x, w)
	if err != nil {
		t.Fatalf("MatMul: %v", err)
	}
	if err := y.Backward(); err != nil {
		t.Fatalf("Backward: %v", err)
	}
	check(t, "dy/dw", w.Grad(), nil, []int{1, 1}, []float32{3})
}

func TestAutogradSumMul(t *testing.T) {
	x := dense([]float32{1, 2, 3}, 1, 3)
	x.SetRequiresGrad(true)

	sq, err := Mul(x, x)
	if err != nil {
		t.Fatalf("Mul: %v", err)
	}
	loss, err := Sum(sq, nil, false)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	if err := loss.Backward(); err != nil {
		t.Fatalf("Backward: %v", err)
	}
	check(t, "d(x^2)/dx", x.Grad(), nil, []int{1, 3}, []float32{2, 4, 6})
}

func TestComparisonsReturnBool(t *testing.T) {
	a := dense([]float32{1, 2, 3}, 3)
	b := dense([]float32{1, 0, 4}, 3)
	got, err := Eq(a, b)
	check(t, "eq", got, err, []int{3}, []float32{1, 0, 0})
	if got.DType() != Bool {
		t.Fatalf("eq dtype = %s, want bool", got.DType())
	}
	got, err = Gt(a, b)
	check(t, "gt", got, err, []int{3}, []float32{0, 1, 0})
}

func TestSoftmaxDelegates(t *testing.T) {
	x := dense([]float32{1, 2, 3}, 1, 3)
	s, err := Softmax(x, 1)
	check(t, "softmax", s, err, []int{1, 3},
		[]float32{0.09003057, 0.24472847, 0.66524096})
}

func TestReductions(t *testing.T) {
	x := dense([]float32{1, 2, 3, 4, 5, 6}, 2, 3)

	s, err := Sum(x, []int{0}, false)
	check(t, "sum", s, err, []int{3}, []float32{5, 7, 9})

	all, err := Sum(x, nil, false)
	check(t, "sum all", all, err, []int{}, []float32{21})

	m, err := Mean(x, []int{0}, false)
	check(t, "mean", m, err, []int{3}, []float32{2.5, 3.5, 4.5})

	mx, err := Max(x, []int{1}, false)
	check(t, "max", mx, err, []int{2}, []float32{3, 6})

	mn, err := Min(x, []int{1}, false)
	check(t, "min", mn, err, []int{2}, []float32{1, 4})

	am, err := ArgMax(x, 1)
	check(t, "argmax", am, err, []int{2}, []float32{2, 2})
}

func TestCreationOps(t *testing.T) {
	z, _ := Zeros([]int{2, 2}, Float32, CPU)
	check(t, "zeros", z, nil, []int{2, 2}, []float32{0, 0, 0, 0})

	o, _ := Ones([]int{2, 2}, Float32, CPU)
	check(t, "ones", o, nil, []int{2, 2}, []float32{1, 1, 1, 1})

	f, _ := Full([]int{2}, 5, Float32, CPU)
	check(t, "full", f, nil, []int{2}, []float32{5, 5})

	e, _ := Eye(2, Float32, CPU)
	check(t, "eye", e, nil, []int{2, 2}, []float32{1, 0, 0, 1})

	r, _ := Arange(0, 5, 1, Float32, CPU)
	check(t, "arange", r, nil, []int{5}, []float32{0, 1, 2, 3, 4})

	rn, _ := Randn([]int{2, 3}, Float32, CPU)
	if !slices.Equal(rn.Shape(), []int{2, 3}) {
		t.Fatalf("randn shape = %v", rn.Shape())
	}
}

func TestPowNegativeBaseIntegerExponent(t *testing.T) {
	x := dense([]float32{-2, 3}, 1, 2)
	y, err := Pow(x, scalar(2))
	check(t, "pow", y, err, []int{1, 2}, []float32{4, 9})

	x.SetRequiresGrad(true)
	out, err := Pow(x, scalar(2))
	if err != nil {
		t.Fatalf("Pow: %v", err)
	}
	loss, err := Sum(out, nil, false)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	if err := loss.Backward(); err != nil {
		t.Fatalf("Backward: %v", err)
	}
	check(t, "pow grad", x.Grad(), nil, []int{1, 2}, []float32{-4, 6})
}

func TestShapeOps(t *testing.T) {
	x := dense([]float32{1, 2, 3, 4, 5, 6}, 2, 3)

	r, err := Reshape(x, 3, 2)
	check(t, "reshape", r, err, []int{3, 2}, []float32{1, 2, 3, 4, 5, 6})

	tp, err := Transpose(x, 0, 1)
	check(t, "transpose", tp, err, []int{3, 2}, []float32{1, 4, 2, 5, 3, 6})

	fl, err := Flatten(x)
	check(t, "flatten", fl, err, []int{6}, []float32{1, 2, 3, 4, 5, 6})
}

func TestTensorRepr(t *testing.T) {
	f32 := func(data []float32, shape ...int) *Tensor {
		tt, err := NewTensor(data, shape, Float32, CPU)
		if err != nil {
			t.Fatal(err)
		}
		return tt
	}
	tests := []struct {
		name string
		got  *Tensor
		want string
	}{
		{
			name: "scalar int",
			got: func() *Tensor {
				tt, err := NewInt64Tensor([]int64{7}, nil, CPU)
				if err != nil {
					t.Fatal(err)
				}
				return tt
			}(),
			want: "tensor(7)",
		},
		{
			name: "scalar float",
			got:  f32([]float32{7}, []int{}...),
			want: "tensor(7.)",
		},
		{
			name: "vector float",
			got:  f32([]float32{1, 2, 3}, 3),
			want: "tensor([1., 2., 3.])",
		},
		{
			name: "matrix float",
			got:  f32([]float32{1, 2, 3, 4, 5, 6}, 2, 3),
			want: "tensor([[1., 2., 3.],\n        [4., 5., 6.]])",
		},
		{
			name: "non integral precision",
			got:  f32([]float32{1.5, 2.25}, 2),
			want: "tensor([1.5000, 2.2500])",
		},
		{
			name: "bool padding",
			got: func() *Tensor {
				tt, err := NewBoolTensor([]bool{true, false}, []int{2}, CPU)
				if err != nil {
					t.Fatal(err)
				}
				return tt
			}(),
			want: "tensor([ true, false])",
		},
		{
			name: "int32 suffix",
			got: func() *Tensor {
				tt, err := NewInt32Tensor([]int32{1, 2, 3}, []int{3}, CPU)
				if err != nil {
					t.Fatal(err)
				}
				return tt
			}(),
			want: "tensor([1, 2, 3], dtype=int32)",
		},
		{
			name: "float64 suffix",
			got: func() *Tensor {
				tt, err := NewFloat64Tensor([]float64{1, 2}, []int{2}, CPU)
				if err != nil {
					t.Fatal(err)
				}
				return tt
			}(),
			want: "tensor([1., 2.], dtype=float64)",
		},
		{
			name: "requires grad",
			got: func() *Tensor {
				tt := f32([]float32{1, 2}, 2)
				tt.SetRequiresGrad(true)
				return tt
			}(),
			want: "tensor([1., 2.], requires_grad=True)",
		},
		{
			name: "three dims",
			got:  f32([]float32{1, 2, 3, 4, 5, 6, 7, 8}, 2, 2, 2),
			want: "tensor([[[1., 2.],\n         [3., 4.]],\n\n        [[5., 6.],\n         [7., 8.]]])",
		},
		{
			name: "summarized int",
			got: func() *Tensor {
				data := make([]int64, 2000)
				for i := range data {
					data[i] = int64(i)
				}
				tt, err := NewInt64Tensor(data, []int{2000}, CPU)
				if err != nil {
					t.Fatal(err)
				}
				return tt
			}(),
			want: "tensor([   0,    1,    2,  ..., 1997, 1998, 1999])",
		},
	}
	for _, tt := range tests {
		if got := tt.got.String(); got != tt.want {
			t.Errorf("%s:\n got: %q\nwant: %q", tt.name, got, tt.want)
		}
	}
}

func TestNewScalarTensorDTypePolicy(t *testing.T) {
	references := map[string]*Tensor{}
	for _, dtype := range []DType{Float32, Float64, Int32, Int64} {
		reference, err := Zeros([]int{1}, dtype, CPU)
		if err != nil {
			t.Fatalf("Zeros(%s): %v", dtype, err)
		}
		references[dtype.String()] = reference
	}

	tests := []struct {
		name      string
		value     any
		reference *Tensor
		want      DType
	}{
		{name: "default integer", value: int64(1), want: Int64},
		{name: "default float", value: float64(1), want: Float32},
		{name: "integer in float32", value: int64(1), reference: references["float32"], want: Float32},
		{name: "integer in float64", value: int64(1), reference: references["float64"], want: Float64},
		{name: "integer in int32", value: int64(1), reference: references["int32"], want: Int32},
		{name: "integer in int64", value: int64(1), reference: references["int64"], want: Int64},
		{name: "out of range int32", value: int64(1) << 31, reference: references["int32"], want: Int64},
		{name: "float in float64", value: float64(0.5), reference: references["float64"], want: Float64},
		{name: "float in int64", value: float64(0.5), reference: references["int64"], want: Float32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewScalarTensor(tt.value, tt.reference)
			if err != nil {
				t.Fatalf("NewScalarTensor: %v", err)
			}
			if got.DType() != tt.want {
				t.Fatalf("NewScalarTensor dtype = %s, want %s", got.DType(), tt.want)
			}
		})
	}
}

func TestTensorLen(t *testing.T) {
	tests := []struct {
		name    string
		data    []float32
		shape   []int
		want    int
		wantErr string
	}{
		{name: "one dimensional", data: []float32{1, 2, 3}, shape: []int{3}, want: 3},
		{name: "two dimensional uses dim 0", data: []float32{1, 2, 3, 4, 5, 6}, shape: []int{2, 3}, want: 2},
		{name: "three dimensional uses dim 0", data: make([]float32, 24), shape: []int{2, 3, 4}, want: 2},
		{name: "zero dimensional is an error", data: []float32{7}, shape: []int{}, wantErr: "len() of a 0-d tensor"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tn, err := NewTensor(tt.data, tt.shape, Float32, CPU)
			if err != nil {
				t.Fatalf("NewTensor: %v", err)
			}
			got, err := tn.Len()
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("Len() = %d, want error %q", got, tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("Len() error = %q, want %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Len(): %v", err)
			}
			if got != tt.want {
				t.Fatalf("Len() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestTensorLenTracksViews(t *testing.T) {
	base, err := NewTensor([]float32{1, 2, 3, 4, 5, 6}, []int{2, 3}, Float32, CPU)
	if err != nil {
		t.Fatalf("NewTensor: %v", err)
	}
	transposed, err := Transpose(base, 0, 1)
	if err != nil {
		t.Fatalf("Transpose: %v", err)
	}
	if got, err := transposed.Len(); err != nil || got != 3 {
		t.Fatalf("len(transposed) = %d, %v, want 3", got, err)
	}
	flat, err := Flatten(base)
	if err != nil {
		t.Fatalf("Flatten: %v", err)
	}
	if got, err := flat.Len(); err != nil || got != 6 {
		t.Fatalf("len(flattened) = %d, %v, want 6", got, err)
	}
	reduced, err := Sum(base, []int{0, 1}, false)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	if got, err := reduced.Len(); err == nil || got != 0 {
		t.Fatalf("len(0-d sum) = %d, %v, want an error", got, err)
	}
}
