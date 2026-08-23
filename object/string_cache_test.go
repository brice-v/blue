package object

import (
	"testing"
)

func TestNewStringCachesShortStrings(t *testing.T) {
	// Cached values must return identical pointers
	if NewString("") != EmptyString {
		t.Fatal("expected NewString(\"\") to return EmptyString")
	}
	for _, c := range []byte{' ', '\n', '\t', 'a', '0', '~', '\x00'} {
		first, second := NewString(string(c)), NewString(string(c))
		if first != second {
			t.Fatalf("expected shared object for %q", string(c))
		}
	}
	space := " "
	if NewString(" ") != NewString(space) {
		t.Fatal("expected ' ' cache hit from any source string")
	}

	// Non-cacheable values must preserve contents
	for _, s := range []string{"\u00e9", "ab", "hello world", string(rune(200))} {
		got := NewString(s)
		if got.Value != s {
			t.Fatalf("value mismatch: got %q want %q", got.Value, s)
		}
	}
	// Multi-byte and longer strings must not alias each other
	e1, e2 := NewString("\u00e9"), NewString("\u00e9")
	if e1 == e2 {
		t.Fatal("non-ASCII single-rune string should not be cached/shared")
	}
}

var benchSinkString *Stringo

func BenchmarkStringoAlloc(b *testing.B) {
	short := "e"
	long := "the quick brown fox jumps over the lazy dog"
	b.Run("literal-empty", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			benchSinkString = &Stringo{Value: ""}
		}
	})
	b.Run("newstring-empty", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			benchSinkString = NewString("")
		}
	})
	b.Run("literal-char", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			benchSinkString = &Stringo{Value: short}
		}
	})
	b.Run("newstring-char", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			benchSinkString = NewString(short)
		}
	})
	b.Run("literal-long", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			benchSinkString = &Stringo{Value: long}
		}
	})
	b.Run("newstring-long", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			benchSinkString = NewString(long)
		}
	})
}
