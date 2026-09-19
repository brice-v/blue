package half

import (
	"math"
	"testing"
)

func TestFloat16ToFloat32(t *testing.T) {
	cases := []struct {
		name string
		in   uint16
		want float32
	}{
		{"zero", 0x0000, 0},
		{"one", 0x3C00, 1},
		{"neg-one", 0xBC00, -1},
		{"two", 0x4000, 2},
		{"half", 0x3800, 0.5},
		{"neg-two", 0xC000, -2},
		{"max-normal", 0x7BFF, 65504},
		{"min-positive-normal", 0x0400, float32(math.Ldexp(1, -14))},
		{"smallest-subnormal", 0x0001, float32(math.Ldexp(1, -24))},
	}
	for _, c := range cases {
		if got := Float16ToFloat32(c.in); got != c.want {
			t.Errorf("%s: Float16ToFloat32(0x%04X) = %v, want %v", c.name, c.in, got, c.want)
		}
	}

	// -0.0 differs from +0.0 only in the sign bit, which == cannot observe
	// (-0.0 == 0.0), so assert the sign is preserved through the conversion.
	if got := Float16ToFloat32(0x8000); got != 0 || !math.Signbit(float64(got)) {
		t.Errorf("neg-zero: Float16ToFloat32(0x8000) = %v (signbit %v), want -0", got, math.Signbit(float64(got)))
	}
}

func TestFloat16ToFloat32Special(t *testing.T) {
	if got := Float16ToFloat32(0x7C00); !math.IsInf(float64(got), 1) {
		t.Errorf("+Inf: got %v, want +Inf", got)
	}
	if got := Float16ToFloat32(0xFC00); !math.IsInf(float64(got), -1) {
		t.Errorf("-Inf: got %v, want -Inf", got)
	}
	if got := Float16ToFloat32(0x7E00); !math.IsNaN(float64(got)) {
		t.Errorf("NaN: got %v, want NaN", got)
	}
}

func TestBFloat16ToFloat32(t *testing.T) {
	cases := []struct {
		name string
		in   uint16
		want float32
	}{
		{"zero", 0x0000, 0},
		{"one", 0x3F80, 1},
		{"neg-one", 0xBF80, -1},
		{"two", 0x4000, 2},
		{"three", 0x4040, 3},
		{"one-and-half", 0x3FC0, 1.5}, // exercises a mantissa bit
		{"min-positive-normal", 0x0080, float32(math.Ldexp(1, -126))},
		{"smallest-subnormal", 0x0001, float32(math.Ldexp(1, -133))},
	}
	for _, c := range cases {
		if got := BFloat16ToFloat32(c.in); got != c.want {
			t.Errorf("%s: BFloat16ToFloat32(0x%04X) = %v, want %v", c.name, c.in, got, c.want)
		}
	}

	// -0.0 differs from +0.0 only in the sign bit, which == cannot observe
	// (-0.0 == 0.0), so assert the sign is preserved through the conversion.
	if got := BFloat16ToFloat32(0x8000); got != 0 || !math.Signbit(float64(got)) {
		t.Errorf("neg-zero: BFloat16ToFloat32(0x8000) = %v (signbit %v), want -0", got, math.Signbit(float64(got)))
	}
}

func TestBFloat16ToFloat32Special(t *testing.T) {
	if got := BFloat16ToFloat32(0x7F80); !math.IsInf(float64(got), 1) {
		t.Errorf("+Inf: got %v, want +Inf", got)
	}
	if got := BFloat16ToFloat32(0xFF80); !math.IsInf(float64(got), -1) {
		t.Errorf("-Inf: got %v, want -Inf", got)
	}
	if got := BFloat16ToFloat32(0x7FC0); !math.IsNaN(float64(got)) {
		t.Errorf("NaN: got %v, want NaN", got)
	}
}
