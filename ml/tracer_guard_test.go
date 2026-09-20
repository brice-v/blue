package ml

import "testing"

// TestTracerRejectsHostRead checks that reading a fake tensor's data during a
// capture fails the capture, so a compiled graph can never silently bake in
// zeros for a value the forward read at trace time.
func TestTracerRejectsHostRead(t *testing.T) {
	x := mustTensor(t, []float32{1, 2, 3, 4}, []int{4})

	tr, err := NewTracer(x)
	if err != nil {
		t.Fatalf("NewTracer: %v", err)
	}
	fake := tr.InputTensor()
	// Simulate a host read inside the traced forward, e.g. .item() or to_list().
	_ = fake.ContiguousData()

	if _, err := tr.Finish(fake); err == nil {
		t.Fatal("expected the capture to fail after a host read of a fake tensor")
	}
}

// TestTracerRejectsUnsupportedOp checks an unrecordable op fails the capture.
func TestTracerRejectsUnsupportedOp(t *testing.T) {
	x := mustTensor(t, []float32{1, 2, 3, 4}, []int{4})
	tr, err := NewTracer(x)
	if err != nil {
		t.Fatalf("NewTracer: %v", err)
	}
	fake := tr.InputTensor()
	// A 1D matmul is outside what the tracer records.
	tr.MatMul(fake.t.Raw(), fake.t.Raw())
	if _, err := tr.Finish(fake); err == nil {
		t.Fatal("expected the capture to fail on an unsupported op")
	}
}
