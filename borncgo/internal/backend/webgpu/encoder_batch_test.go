//go:build !js

package webgpu

import (
	"testing"

	"blue/borncgo/internal/tensor"
)

// ─────────────────────────────────────────────────────────────────────────────
// Input buffer cache tests
// ─────────────────────────────────────────────────────────────────────────────

// TestInputBufferCache_HitOnRepeatUse verifies that the same *RawTensor pointer
// returns the same GPU buffer on subsequent ops, i.e. the cache is populated on
// first use and reused on second use.
func TestInputBufferCache_HitOnRepeatUse(t *testing.T) {
	if !IsAvailable() {
		t.Skip("WebGPU not available")
	}

	backend, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer backend.Release()

	weights, err := tensor.NewRaw(tensor.Shape{4, 4}, tensor.Float32, tensor.CPU)
	if err != nil {
		t.Fatal(err)
	}
	for i := range weights.AsFloat32() {
		weights.AsFloat32()[i] = float32(i) * 0.1
	}

	// First op: cache miss — uploads weights to GPU.
	result1 := backend.Add(weights, weights)
	_ = result1.Data()

	cacheAfterFirst := backend.inputBufferCacheSize()
	if cacheAfterFirst == 0 {
		t.Error("expected cache to have at least one entry after first op")
	}

	// Second op with the SAME tensor pointer: must be a cache hit (no re-upload).
	result2 := backend.Add(weights, weights)
	_ = result2.Data()

	cacheAfterSecond := backend.inputBufferCacheSize()
	if cacheAfterSecond != cacheAfterFirst {
		t.Errorf("cache size changed after second op: got %d, want %d (same as after first)",
			cacheAfterSecond, cacheAfterFirst)
	}
}

// TestInputBufferCache_NewTensorNewEntry verifies that two different *RawTensor
// objects (even with identical data) each get their own cache entry.
func TestInputBufferCache_NewTensorNewEntry(t *testing.T) {
	if !IsAvailable() {
		t.Skip("WebGPU not available")
	}

	backend, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer backend.Release()

	a, err := tensor.NewRaw(tensor.Shape{4}, tensor.Float32, tensor.CPU)
	if err != nil {
		t.Fatal(err)
	}
	copy(a.AsFloat32(), []float32{1, 2, 3, 4})

	b2, err := tensor.NewRaw(tensor.Shape{4}, tensor.Float32, tensor.CPU)
	if err != nil {
		t.Fatal(err)
	}
	copy(b2.AsFloat32(), []float32{1, 2, 3, 4}) // Same data, different pointer.

	// Use both tensors in ops to populate cache.
	r1 := backend.Add(a, a)
	_ = r1.Data()
	sizeAfterA := backend.inputBufferCacheSize()

	r2 := backend.Add(b2, b2)
	_ = r2.Data()
	sizeAfterB := backend.inputBufferCacheSize()

	if sizeAfterB <= sizeAfterA {
		t.Errorf("expected cache to grow after second distinct tensor: got %d, want > %d",
			sizeAfterB, sizeAfterA)
	}
}

// TestInputBufferCache_LazyTensorNotCached verifies that lazy (GPU-backed) tensors
// produced by GPU ops are NOT inserted into the input buffer cache. Caching
// transient intermediates would cause use-after-free when their staging buffers
// are released after readback.
func TestInputBufferCache_LazyTensorNotCached(t *testing.T) {
	if !IsAvailable() {
		t.Skip("WebGPU not available")
	}

	backend, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer backend.Release()

	cpu, err := tensor.NewRaw(tensor.Shape{4}, tensor.Float32, tensor.CPU)
	if err != nil {
		t.Fatal(err)
	}
	copy(cpu.AsFloat32(), []float32{1, 2, 3, 4})

	// Produce a lazy GPU tensor.
	lazyResult := backend.Add(cpu, cpu)
	cacheBefore := backend.inputBufferCacheSize()

	// Use the lazy result as input — must NOT add it to the cache.
	_ = backend.Add(lazyResult, lazyResult)

	// Read lazyResult to realize it.
	_ = lazyResult.Data()

	cacheAfter := backend.inputBufferCacheSize()
	if cacheAfter > cacheBefore {
		t.Errorf("lazy tensor was added to cache: size grew from %d to %d",
			cacheBefore, cacheAfter)
	}
}

// TestInputBufferCache_ClearResets verifies that ClearInputBufferCache empties
// the cache and that subsequent ops re-populate it correctly.
func TestInputBufferCache_ClearResets(t *testing.T) {
	if !IsAvailable() {
		t.Skip("WebGPU not available")
	}

	backend, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer backend.Release()

	w, err := tensor.NewRaw(tensor.Shape{4}, tensor.Float32, tensor.CPU)
	if err != nil {
		t.Fatal(err)
	}
	copy(w.AsFloat32(), []float32{1, 2, 3, 4})

	r := backend.Add(w, w)
	_ = r.Data()

	if backend.inputBufferCacheSize() == 0 {
		t.Fatal("cache should be non-empty after op")
	}

	backend.ClearInputBufferCache()

	if backend.inputBufferCacheSize() != 0 {
		t.Errorf("cache should be empty after ClearInputBufferCache, got %d", backend.inputBufferCacheSize())
	}

	// After clear, ops must still produce correct results.
	r2 := backend.Add(w, w)
	got := r2.AsFloat32()
	for i, v := range got {
		want := w.AsFloat32()[i] * 2
		if v != want {
			t.Errorf("element %d: got %f, want %f", i, v, want)
		}
	}
}

// TestInputBufferCache_CorrectResults verifies that cached buffer reuse produces
// numerically correct results across multiple ops with the same tensor.
func TestInputBufferCache_CorrectResults(t *testing.T) {
	if !IsAvailable() {
		t.Skip("WebGPU not available")
	}

	backend, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer backend.Release()

	w, err := tensor.NewRaw(tensor.Shape{4}, tensor.Float32, tensor.CPU)
	if err != nil {
		t.Fatal(err)
	}
	copy(w.AsFloat32(), []float32{1, 2, 3, 4})

	// Run the same op 10 times; all should produce w+w = [2,4,6,8].
	for i := 0; i < 10; i++ {
		result := backend.Add(w, w)
		got := result.AsFloat32()
		expected := []float32{2, 4, 6, 8}
		for j, v := range got {
			if v != expected[j] {
				t.Errorf("iteration %d, element %d: got %f, want %f", i, j, v, expected[j])
			}
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Shared encoder accumulator tests
// ─────────────────────────────────────────────────────────────────────────────

// TestEncoderBatch_AccumulatesMultiplePasses verifies that multiple ops produce
// correct results when accumulated in the shared encoder. The shared encoder
// accumulates compute passes to avoid per-op submit overhead; correctness of the
// final readback is the observable contract. We do not assert the intermediate
// batch count because the GPU driver may auto-flush at any point.
func TestEncoderBatch_AccumulatesMultiplePasses(t *testing.T) {
	if !IsAvailable() {
		t.Skip("WebGPU not available")
	}

	backend, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer backend.Release()

	a, err := tensor.NewRaw(tensor.Shape{16}, tensor.Float32, tensor.CPU)
	if err != nil {
		t.Fatal(err)
	}
	for i := range a.AsFloat32() {
		a.AsFloat32()[i] = 1.0
	}

	// Issue 5 ops without any readback.
	const opsToAccumulate = 5
	result := a
	for i := 0; i < opsToAccumulate; i++ {
		result = backend.Add(result, a)
	}

	// Readback flushes the encoder and returns the result.
	got := result.AsFloat32()
	if len(got) != 16 {
		t.Fatalf("output length: got %d, want 16", len(got))
	}
	// result = a (all 1s), then add a five times → result = 6a = 6.0 per element.
	want := float32(6.0)
	for i, v := range got {
		if v != want {
			t.Errorf("element %d: got %f, want %f", i, v, want)
		}
	}

	// Post-readback contract: active batch must be drained.
	if count := backend.activeBatchCount(); count != 0 {
		t.Errorf("activeBatchCount() = %d after Data(); expected 0", count)
	}
}

// TestEncoderBatch_FlushClearsActiveBatch verifies that after a Data() access,
// the active batch count resets to 0 (encoder was submitted) and the result
// is numerically correct.
func TestEncoderBatch_FlushClearsActiveBatch(t *testing.T) {
	if !IsAvailable() {
		t.Skip("WebGPU not available")
	}

	backend, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer backend.Release()

	a, err := tensor.NewRaw(tensor.Shape{4}, tensor.Float32, tensor.CPU)
	if err != nil {
		t.Fatal(err)
	}
	copy(a.AsFloat32(), []float32{1, 2, 3, 4})

	result := backend.Add(a, a)

	// Readback triggers flushCommands which calls finishActiveBatchLocked.
	// We do not assert the batch count before readback — the GPU driver may
	// have already flushed the op (timing-dependent). The post-readback state
	// is the observable contract.
	got := result.AsFloat32()

	// After readback: active batch must be empty.
	if count := backend.activeBatchCount(); count != 0 {
		t.Errorf("activeBatchCount() = %d after readback; expected 0", count)
	}

	// Numerical correctness: a + a = [2, 4, 6, 8].
	expected := []float32{2, 4, 6, 8}
	for i, v := range got {
		if v != expected[i] {
			t.Errorf("element %d: got %f, want %f", i, v, expected[i])
		}
	}
}

// TestEncoderBatch_AutoFlushAtThreshold verifies that when ops exceed
// maxPendingBeforeFlush, the encoder is sealed and a new one started.
// This is a stress variant of TestAutoFlush_ManyOpsWithoutReadback.
func TestEncoderBatch_AutoFlushAtThreshold(t *testing.T) {
	if !IsAvailable() {
		t.Skip("WebGPU not available")
	}

	backend, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer backend.Release()

	a, err := tensor.NewRaw(tensor.Shape{4}, tensor.Float32, tensor.CPU)
	if err != nil {
		t.Fatal(err)
	}
	copy(a.AsFloat32(), []float32{1, 1, 1, 1})

	// Exceed the threshold by 10 — auto-flush must trigger mid-loop.
	totalOps := maxPendingBeforeFlush + 10

	result := a
	for i := 0; i < totalOps; i++ {
		result = backend.Add(result, a)
	}

	// Must not panic and must produce correct output.
	got := result.AsFloat32()
	if len(got) != 4 {
		t.Fatalf("output length: got %d, want 4", len(got))
	}
	// All 1s + totalOps additions: result = (totalOps+1) * 1.0
	want := float32(totalOps + 1)
	for i, v := range got {
		if v != want {
			t.Errorf("element %d: got %f, want %f", i, v, want)
		}
	}
}

// TestEncoderBatch_MatMulWithCachedWeights verifies that cached weight matrices
// (same tensor reused for matmul) produce correct results across iterations.
// This is the primary use-case: transformer layer forward pass with fixed weights.
func TestEncoderBatch_MatMulWithCachedWeights(t *testing.T) {
	if !IsAvailable() {
		t.Skip("WebGPU not available")
	}

	backend, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer backend.Release()

	// Weight matrix [4, 4].
	weights, err := tensor.NewRaw(tensor.Shape{4, 4}, tensor.Float32, tensor.CPU)
	if err != nil {
		t.Fatal(err)
	}
	// Identity-like: diagonal 2, zero elsewhere.
	wData := weights.AsFloat32()
	for i := range wData {
		row, col := i/4, i%4
		if row == col {
			wData[i] = 2.0
		}
	}

	// Input [4, 4] — all ones.
	input, err := tensor.NewRaw(tensor.Shape{4, 4}, tensor.Float32, tensor.CPU)
	if err != nil {
		t.Fatal(err)
	}
	for i := range input.AsFloat32() {
		input.AsFloat32()[i] = 1.0
	}

	// Simulate 3 forward passes reusing the same weight tensor.
	for pass := 0; pass < 3; pass++ {
		result := backend.MatMul(input, weights)
		got := result.AsFloat32()
		// input * 2I = 2*input, so each diagonal output is 2.0, off-diag 0.
		// Actually input@weights: input[4,4] all-1 @ diag(2)[4,4] = [2,2,2,2] per row.
		if len(got) != 16 {
			t.Fatalf("pass %d: output length = %d, want 16", pass, len(got))
		}
		for i, v := range got {
			want := float32(2.0) // Each row is [2,2,2,2] for all-ones input @ 2*Identity.
			if v != want {
				t.Errorf("pass %d, element %d: got %f, want %f", pass, i, v, want)
			}
		}
	}
}
