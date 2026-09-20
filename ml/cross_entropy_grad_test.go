package ml

import "testing"

// TestCrossEntropyGradMatchesBackward checks the seed helper equals the gradient
// the eager cross-entropy backward produces.
func TestCrossEntropyGradMatchesBackward(t *testing.T) {
	logits := mustTensor(t, []float32{
		2.0, 1.0, 0.1,
		-1.0, 0.5, 2.5,
		0.3, -0.2, 1.1,
	}, []int{3, 3})
	logits.SetRequiresGrad(true)
	target := mustTensor(t, []float32{0, 2, 1}, []int{3})

	loss, err := CrossEntropy(logits, target)
	if err != nil {
		t.Fatalf("CrossEntropy: %v", err)
	}
	if err := loss.Backward(); err != nil {
		t.Fatalf("Backward: %v", err)
	}
	want := append([]float32(nil), logits.Grad().ContiguousData()...)

	grad, err := CrossEntropyGrad(logits, target)
	if err != nil {
		t.Fatalf("CrossEntropyGrad: %v", err)
	}
	got := grad.ContiguousData()
	if !closeData(want, got) {
		t.Fatalf("cross-entropy seed differs from eager grad:\n want %v\n got  %v", want, got)
	}
}
