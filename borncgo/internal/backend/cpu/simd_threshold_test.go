package cpu

import (
	"testing"
)

// TestSIMDMinLen_IsPositive verifies the threshold constant has a valid value.
func TestSIMDMinLen_IsPositive(t *testing.T) {
	if simdMinLen <= 0 {
		t.Fatalf("simdMinLen must be > 0, got %d", simdMinLen)
	}
}

// TestSIMDThreshold_DispatchSkippedBelowMinLen verifies that a registered SIMD
// kernel is NOT invoked for slices shorter than simdMinLen.  The test installs
// a sentinel kernel that panics; if the dispatch guard is correct the panic
// never fires for small slices.
func TestSIMDThreshold_DispatchSkippedBelowMinLen(t *testing.T) {
	// Sentinel kernel that panics when called — should never be reached for
	// slices shorter than simdMinLen.
	sentinel := func(a, _ []float32) {
		t.Fatalf("SIMD kernel invoked for len=%d, want scalar path (simdMinLen=%d)", len(a), simdMinLen)
	}

	saved := simdAddInplaceFloat32
	simdAddInplaceFloat32 = sentinel
	t.Cleanup(func() { simdAddInplaceFloat32 = saved })

	// All lengths strictly below simdMinLen must use the scalar path.
	for n := 0; n < simdMinLen; n++ {
		a := make([]float32, n)
		b := make([]float32, n)
		for i := range a {
			a[i] = float32(i + 1)
			b[i] = float32(i + 1)
		}
		addInplaceFloat32(a, b) // must not invoke sentinel
		// Verify scalar result is correct.
		for i := range a {
			want := float32(i+1) * 2
			if a[i] != want {
				t.Fatalf("n=%d, i=%d: got %v, want %v", n, i, a[i], want)
			}
		}
	}
}

// TestSIMDThreshold_DispatchUsedAtMinLen verifies that a registered SIMD
// kernel IS invoked once the slice length reaches simdMinLen.
func TestSIMDThreshold_DispatchUsedAtMinLen(t *testing.T) {
	called := false
	fakeKernel := func(a, b []float32) {
		called = true
		// Perform the actual operation so the rest of the test can check results.
		for i := range a {
			a[i] += b[i]
		}
	}

	saved := simdAddInplaceFloat32
	simdAddInplaceFloat32 = fakeKernel
	t.Cleanup(func() { simdAddInplaceFloat32 = saved })

	a := make([]float32, simdMinLen)
	b := make([]float32, simdMinLen)
	for i := range a {
		a[i] = 1
		b[i] = 1
	}

	addInplaceFloat32(a, b)

	if !called {
		t.Fatalf("SIMD kernel was not invoked for len=%d (simdMinLen=%d)", simdMinLen, simdMinLen)
	}
	for i, v := range a {
		if v != 2 {
			t.Fatalf("i=%d: got %v, want 2", i, v)
		}
	}
}
