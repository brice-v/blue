package ml

import (
	"math"
	"testing"
)

// TestCrossEntropyHandComputed checks the default reduction is mean and the
// value matches the hand-computed mean cross-entropy for logits [1, 2, 3] with
// target class 2: -log(softmax([1,2,3])[2]) = 0.40760595.
func TestCrossEntropyHandComputed(t *testing.T) {
	logits := dense([]float32{1, 2, 3}, 1, 3)
	target := mustIntTensor(t, []int32{2}, []int{1})
	loss, err := CrossEntropy(logits, target)
	if err != nil {
		t.Fatalf("CrossEntropy: %v", err)
	}
	got, err := loss.Item()
	if err != nil {
		t.Fatalf("Item: %v", err)
	}
	if math.Abs(float64(got)-0.40760595) > 1e-5 {
		t.Fatalf("cross entropy = %v, want 0.40760595", got)
	}
}

// TestCrossEntropyMeanOverBatch checks the reduction averages over the batch:
// two identical rows give the same value as one.
func TestCrossEntropyMeanOverBatch(t *testing.T) {
	logits := dense([]float32{1, 2, 3, 1, 2, 3}, 2, 3)
	target := mustIntTensor(t, []int32{2, 2}, []int{2})
	loss, err := CrossEntropy(logits, target)
	if err != nil {
		t.Fatalf("CrossEntropy: %v", err)
	}
	got, err := loss.Item()
	if err != nil {
		t.Fatalf("Item: %v", err)
	}
	if math.Abs(float64(got)-0.40760595) > 1e-5 {
		t.Fatalf("mean cross entropy = %v, want 0.40760595", got)
	}
}
