package object

import (
	"strconv"
	"testing"
)

func TestInternStringRouting(t *testing.T) {
	if InternString("") != EmptyString {
		t.Fatal("empty string should route to EmptyString")
	}
	if InternString(" ") != asciiStrings[' '] {
		t.Fatal("single ASCII byte should route to asciiStrings")
	}
}

func TestInternStringRepeats(t *testing.T) {
	// Consecutive lookups of the same value must return the same object
	for _, s := range []string{"é", "hello", "key-123", string(rune(0x2603))} {
		first, second := InternString(s), InternString(s)
		if first != second {
			t.Fatalf("expected shared object for %q", s)
		}
		if first.Value != s {
			t.Fatalf("value mismatch: got %q want %q", first.Value, s)
		}
	}
}

func TestInternStringLongNotCached(t *testing.T) {
	long := "this string is definitely longer than thirty-two bytes"
	a, b := InternString(long), InternString(long)
	if a == b {
		t.Fatal("long strings must not be interned/shared")
	}
	if a.Value != long || b.Value != long {
		t.Fatal("long string contents must be preserved")
	}
}

func TestInternInt(t *testing.T) {
	// Small non-negative ints share the ascii single-char objects
	if InternInt(7) != asciiStrings['7'] {
		t.Fatal("digit should route to asciiStrings")
	}
	// Repeats return the same object with no content drift
	for _, v := range []int64{-1, -99999, 1000000, 9223372036854775807} {
		want := strconv.FormatInt(v, 10)
		first, second := InternInt(v), InternInt(v)
		if first != second {
			t.Fatalf("expected shared object for %d", v)
		}
		if first.Value != want {
			t.Fatalf("value mismatch: got %q want %q", first.Value, want)
		}
	}
}

var benchSinkIntern *Stringo

func BenchmarkInternString(b *testing.B) {
	b.Run("hit", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			benchSinkIntern = InternString("repeated-value")
		}
	})
	b.Run("int-hit", func(b *testing.B) {
		v := int64(1234567)
		b.ReportAllocs()
		for range b.N {
			benchSinkIntern = InternInt(v)
		}
	})
}
