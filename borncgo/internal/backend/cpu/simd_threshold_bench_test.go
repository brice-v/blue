package cpu

import (
	"fmt"
	"testing"
)

// thresholdSizes covers the sub-threshold range and a few super-threshold sizes
// to make the SIMD vs scalar crossover visible in benchmark output.
// Sizes below simdMinLen (32) exercise the scalar-fallback path; sizes at and
// above it exercise the SIMD path when a kernel is available.
var thresholdSizes = []int{1, 4, 8, 16, 32, 64, 128, 256}

// BenchmarkSIMDThreshold_AddInplaceF32 benchmarks addInplaceFloat32 across
// sizes that span the simdMinLen boundary.  Run with and without
// GOEXPERIMENT=simd to compare SIMD dispatch vs scalar fallback:
//
//	go test -bench=BenchmarkSIMDThreshold_AddInplaceF32 -benchmem ./internal/backend/cpu/...
//	GOEXPERIMENT=simd go test -bench=BenchmarkSIMDThreshold_AddInplaceF32 -benchmem ./internal/backend/cpu/...
//
// Sizes below simdMinLen (32) always use the scalar path even when a SIMD
// kernel is registered, confirming the threshold is effective.
func BenchmarkSIMDThreshold_AddInplaceF32(b *testing.B) {
	for _, n := range thresholdSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			a := createRandomFloat32Slice(n)
			bSlice := createRandomFloat32Slice(n)
			b.SetBytes(int64(n * 4))
			b.ResetTimer()
			for b.Loop() {
				addInplaceFloat32(a, bSlice)
			}
		})
	}
}

// BenchmarkSIMDThreshold_AddInplaceF32_ForcedScalar benchmarks addInplaceFloat32
// with the SIMD kernel disabled, showing the scalar-only cost at each size.
// This is the baseline that SIMD must beat to be profitable.
func BenchmarkSIMDThreshold_AddInplaceF32_ForcedScalar(b *testing.B) {
	saved := simdAddInplaceFloat32
	simdAddInplaceFloat32 = nil
	b.Cleanup(func() { simdAddInplaceFloat32 = saved })

	for _, n := range thresholdSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			a := createRandomFloat32Slice(n)
			bSlice := createRandomFloat32Slice(n)
			b.SetBytes(int64(n * 4))
			b.ResetTimer()
			for b.Loop() {
				addInplaceFloat32(a, bSlice)
			}
		})
	}
}

// BenchmarkSIMDThreshold_AddVectorizedF32 benchmarks addVectorizedFloat32
// across the simdMinLen boundary.
func BenchmarkSIMDThreshold_AddVectorizedF32(b *testing.B) {
	for _, n := range thresholdSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			a := createRandomFloat32Slice(n)
			bSlice := createRandomFloat32Slice(n)
			dst := make([]float32, n)
			b.SetBytes(int64(n * 4))
			b.ResetTimer()
			for b.Loop() {
				addVectorizedFloat32(dst, a, bSlice)
			}
		})
	}
}

// BenchmarkSIMDThreshold_AddVectorizedF32_ForcedScalar is the forced-scalar
// baseline for addVectorizedFloat32.
func BenchmarkSIMDThreshold_AddVectorizedF32_ForcedScalar(b *testing.B) {
	saved := simdAddVectorizedFloat32
	simdAddVectorizedFloat32 = nil
	b.Cleanup(func() { simdAddVectorizedFloat32 = saved })

	for _, n := range thresholdSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			a := createRandomFloat32Slice(n)
			bSlice := createRandomFloat32Slice(n)
			dst := make([]float32, n)
			b.SetBytes(int64(n * 4))
			b.ResetTimer()
			for b.Loop() {
				addVectorizedFloat32(dst, a, bSlice)
			}
		})
	}
}

// BenchmarkSIMDThreshold_MulInplaceF32 benchmarks mulInplaceFloat32 across
// the simdMinLen boundary.  Mul has the same per-element cost as Add on
// modern CPUs, so the crossover should be at the same slice length.
func BenchmarkSIMDThreshold_MulInplaceF32(b *testing.B) {
	for _, n := range thresholdSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			a := createRandomFloat32Slice(n)
			bSlice := createRandomFloat32Slice(n)
			b.SetBytes(int64(n * 4))
			b.ResetTimer()
			for b.Loop() {
				mulInplaceFloat32(a, bSlice)
			}
		})
	}
}

// BenchmarkSIMDThreshold_MulInplaceF32_ForcedScalar is the forced-scalar
// baseline for mulInplaceFloat32.
func BenchmarkSIMDThreshold_MulInplaceF32_ForcedScalar(b *testing.B) {
	saved := simdMulInplaceFloat32
	simdMulInplaceFloat32 = nil
	b.Cleanup(func() { simdMulInplaceFloat32 = saved })

	for _, n := range thresholdSizes {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			a := createRandomFloat32Slice(n)
			bSlice := createRandomFloat32Slice(n)
			b.SetBytes(int64(n * 4))
			b.ResetTimer()
			for b.Loop() {
				mulInplaceFloat32(a, bSlice)
			}
		})
	}
}
