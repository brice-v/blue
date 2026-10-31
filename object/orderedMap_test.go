package object

import (
	"testing"
)

func TestOrderedMapSmallBasics(t *testing.T) {
	m := NewOrderedMap[string, int]()
	if !m.Small() {
		t.Fatal("new map should use the small representation")
	}

	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)
	if !m.Small() {
		t.Fatal("three entries should stay small")
	}
	if m.Len() != 3 {
		t.Fatalf("len = %d, want 3", m.Len())
	}

	for _, tc := range []struct {
		key string
		val int
	}{{"a", 1}, {"b", 2}, {"c", 3}} {
		got, ok := m.Get(tc.key)
		if !ok || got != tc.val {
			t.Fatalf("Get(%q) = %d, %v, want %d, true", tc.key, got, ok, tc.val)
		}
	}
	if got, ok := m.Get("missing"); ok || got != 0 {
		t.Fatalf("Get(missing) = %d, %v, want 0, false", got, ok)
	}

	m.Set("b", 20)
	if got, _ := m.Get("b"); got != 20 {
		t.Fatalf("overwrite Get(b) = %d, want 20", got)
	}
	if m.Len() != 3 {
		t.Fatalf("overwrite changed len to %d", m.Len())
	}

	m.Delete("b")
	if m.Len() != 2 {
		t.Fatalf("len after delete = %d, want 2", m.Len())
	}
	if _, ok := m.Get("b"); ok {
		t.Fatal("deleted key is still present")
	}
	if m.Keys[0] != "a" || m.Keys[1] != "c" {
		t.Fatalf("keys after delete = %v, want [a c]", m.Keys)
	}

	m.Delete("missing")
	if m.Len() != 2 {
		t.Fatalf("deleting a missing key changed len to %d", m.Len())
	}
}

func TestOrderedMapPromotesAboveThreshold(t *testing.T) {
	m := NewOrderedMap[int, int]()
	for i := range smallMapThreshold {
		m.Set(i, i*10)
	}
	if !m.Small() {
		t.Fatalf("%d entries should stay small", smallMapThreshold)
	}

	m.Set(smallMapThreshold, smallMapThreshold*10)
	if m.Small() {
		t.Fatalf("%d entries should promote", smallMapThreshold+1)
	}
	if m.Len() != smallMapThreshold+1 {
		t.Fatalf("len = %d, want %d", m.Len(), smallMapThreshold+1)
	}

	for i := 0; i <= smallMapThreshold; i++ {
		got, ok := m.Get(i)
		if !ok || got != i*10 {
			t.Fatalf("Get(%d) = %d, %v, want %d, true", i, got, ok, i*10)
		}
		if m.Keys[i] != i {
			t.Fatalf("keys[%d] = %d, want %d", i, m.Keys[i], i)
		}
	}

	m.Delete(4)
	if m.Len() != smallMapThreshold {
		t.Fatalf("len after delete = %d, want %d", m.Len(), smallMapThreshold)
	}
	if _, ok := m.Get(4); ok {
		t.Fatal("deleted key is still present after promotion")
	}
	if m.Keys[4] != 5 {
		t.Fatalf("keys[4] = %d, want 5", m.Keys[4])
	}

	m.Set(100, 1000)
	if got, ok := m.Get(100); !ok || got != 1000 {
		t.Fatalf("Get(100) = %d, %v, want 1000, true", got, ok)
	}
}

func TestOrderedMapSizeHint(t *testing.T) {
	if m := NewOrderedMapWithSize[string, int](smallMapThreshold); !m.Small() {
		t.Fatalf("size hint %d should stay small", smallMapThreshold)
	}
	if m := NewOrderedMapWithSize[string, int](smallMapThreshold + 1); m.Small() {
		t.Fatalf("size hint %d should start promoted", smallMapThreshold+1)
	}
}

func TestMapObjectKeysUseValueEquality(t *testing.T) {
	m := &Map{Pairs: NewPairsMap()}
	m.Set(&Stringo{Value: "a"}, &Integer{Value: 1})
	m.Set(&Stringo{Value: "a"}, &Integer{Value: 2})
	if m.Pairs.Len() != 1 {
		t.Fatalf("equal string keys produced %d entries, want 1", m.Pairs.Len())
	}

	pair, ok := m.Get(&Stringo{Value: "a"})
	if !ok {
		t.Fatal("Get with a fresh equal key failed")
	}
	if got := pair.Value.(*Integer).Value; got != 2 {
		t.Fatalf("value = %d, want 2", got)
	}
	if !m.Contains(&Stringo{Value: "a"}) {
		t.Fatal("Contains with a fresh equal key failed")
	}

	m.Delete(&Stringo{Value: "a"})
	if m.Contains(&Stringo{Value: "a"}) {
		t.Fatal("Delete with a fresh equal key did not remove the entry")
	}
}

func TestMapObjectSmallAndPromotedAgree(t *testing.T) {
	build := func(n int) *Map {
		m := &Map{Pairs: NewPairsMap()}
		for i := range n {
			m.Set(&Stringo{Value: "k" + string(rune('a'+i))}, &Integer{Value: int64(i)})
		}
		return m
	}

	for _, n := range []int{3, smallMapThreshold, smallMapThreshold + 1, 20} {
		m := build(n)
		for i := range n {
			key := &Stringo{Value: "k" + string(rune('a'+i))}
			pair, ok := m.Get(key)
			if !ok {
				t.Fatalf("n=%d: Get(%q) failed", n, key.Value)
			}
			if got := pair.Value.(*Integer).Value; got != int64(i) {
				t.Fatalf("n=%d: Get(%q) = %d, want %d", n, key.Value, got, i)
			}
		}
		if m.Pairs.Len() != n {
			t.Fatalf("n=%d: len = %d", n, m.Pairs.Len())
		}
		m.Delete(&Stringo{Value: "ka"})
		if m.Contains(&Stringo{Value: "ka"}) {
			t.Fatalf("n=%d: deleted key still present", n)
		}
		if m.Pairs.Len() != n-1 {
			t.Fatalf("n=%d: len after delete = %d, want %d", n, m.Pairs.Len(), n-1)
		}
	}
}

func TestObjectsEqual(t *testing.T) {
	tests := []struct {
		name string
		a    Object
		b    Object
		want bool
	}{
		{"ints equal", &Integer{Value: 3}, &Integer{Value: 3}, true},
		{"ints differ", &Integer{Value: 3}, &Integer{Value: 4}, false},
		{"uints equal", &UInteger{Value: 3}, &UInteger{Value: 3}, true},
		{"floats equal", &Float{Value: 1.5}, &Float{Value: 1.5}, true},
		{"strings equal", &Stringo{Value: "x"}, &Stringo{Value: "x"}, true},
		{"strings differ", &Stringo{Value: "x"}, &Stringo{Value: "y"}, false},
		{"bools equal", &Boolean{Value: true}, &Boolean{Value: true}, true},
		{"null equal", NULL, NULL, true},
		{"null vs int", NULL, &Integer{Value: 0}, false},
		{"int vs string", &Integer{Value: 1}, &Stringo{Value: "1"}, false},
		{"nil equal", nil, nil, true},
		{"nil vs int", nil, &Integer{Value: 0}, false},
	}
	for _, tc := range tests {
		if got := FastEqual(tc.a, tc.b); got != tc.want {
			t.Errorf("%s: ObjectsEqual = %v, want %v", tc.name, got, tc.want)
		}
	}
}
