package ml

import (
	"path/filepath"
	"testing"
)

func TestStateRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "model.state")
	w := dense([]float32{1, 2, 3, 4}, 2, 2)
	b := dense([]float32{1, 0, 1, 0}, 2, 2)
	if err := SaveState([]NamedTensor{{Name: "w", Tensor: w}, {Name: "b", Tensor: b}}, path); err != nil {
		t.Fatalf("SaveState: %v", err)
	}
	loaded, err := LoadState(path)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if len(loaded) != 2 || loaded[0].Name != "w" || loaded[1].Name != "b" {
		t.Fatalf("names round-tripped wrong: %+v", loaded)
	}
	check(t, "loaded w", loaded[0].Tensor, nil, []int{2, 2}, []float32{1, 2, 3, 4})
	check(t, "loaded b", loaded[1].Tensor, nil, []int{2, 2}, []float32{1, 0, 1, 0})
}

func TestCopyIntoPlace(t *testing.T) {
	src := dense([]float32{1, 2, 3, 4}, 2, 2)
	dst := dense([]float32{0, 0, 0, 0}, 2, 2)
	if err := CopyInto(dst, src); err != nil {
		t.Fatalf("CopyInto: %v", err)
	}
	check(t, "copied", dst, nil, []int{2, 2}, []float32{1, 2, 3, 4})
}

func TestCopyIntoDType(t *testing.T) {
	src := dense([]float32{1, 0, 1}, 3)
	dst, err := NewTensor([]float32{0, 0, 0}, []int{3}, Bool, CPU)
	if err != nil {
		t.Fatalf("NewTensor: %v", err)
	}
	if err := CopyInto(dst, src); err != nil {
		t.Fatalf("CopyInto: %v", err)
	}
	check(t, "bool copied", dst, nil, []int{3}, []float32{1, 0, 1})
}

func TestCopyIntoShapeMismatch(t *testing.T) {
	src := dense([]float32{1, 2, 3}, 3)
	dst := dense([]float32{0, 0}, 2)
	if err := CopyInto(dst, src); err == nil {
		t.Fatal("CopyInto with mismatched shapes should error")
	}
}
