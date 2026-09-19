package cpu

// simdMinLen is the minimum slice length for which SIMD dispatch is profitable.
//
// Below this threshold, the overhead of an indirect function call through a
// non-nil function pointer, combined with the kernel prologue and loop setup
// inside the SIMD kernel, exceeds the savings from vectorised arithmetic.
// For element-wise ops like Add, Sub, Mul and Div:
//
//   - AVX processes 8 float32 per vector iteration; for n < 8 the kernel does
//     zero SIMD iterations and falls through to its scalar tail — pure call
//     overhead with no vectorisation benefit.
//   - AVX-512 processes 16 float32 per vector iteration; the same applies for
//     n < 16.
//
// 32 covers two full AVX vectors (2 × 8) and two full AVX-512 half-widths
// (2 × 16), giving the register pipeline enough work to amortize the
// function-pointer call overhead on both µarchs. This value aligns with the
// threshold used for SIMD dispatch in the Go standard library (e.g. bytes,
// strings) for similar loop-body costs.
//
// Determined empirically via BenchmarkSIMDThreshold (simd_threshold_bench_test.go):
// run with GOEXPERIMENT=simd on an AVX2/AVX-512 host to confirm the crossover.
const simdMinLen = 32
