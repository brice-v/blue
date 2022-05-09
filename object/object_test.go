package object

import "testing"

func TestStringHashKey(t *testing.T) {
	hello1 := &Stringo{Value: "Hello World"}
	hello2 := &Stringo{Value: "Hello World"}
	diff1 := &Stringo{Value: "My name is johnny"}
	diff2 := &Stringo{Value: "My name is johnny"}
	if hello1.HashKey() != hello2.HashKey() {
		t.Errorf("strings with same content have different hash keys")
	}
	if diff1.HashKey() != diff2.HashKey() {
		t.Errorf("strings with same content have different hash keys")
	}
	if hello1.HashKey() == diff1.HashKey() {
		t.Errorf("strings with different content have same hash keys")
	}
}
