package object

import "testing"

func TestRuneAtIndex(t *testing.T) {
	// ASCII: rune index == byte index, result shares the single-byte cache
	got, ok := RuneAtIndex("abcdef", 2)
	if !ok || got != "c" {
		t.Fatalf("ascii: got %q ok=%v", got, ok)
	}
	if NewString(got) != asciiStrings['c'] {
		t.Fatal("expected ascii hit via NewString")
	}

	// Multi-byte runes are returned as their exact byte slice
	s := "aé☃\U0001F600b" // 1 + 1 + 1 + 1(2-byte) ... runes: a é ☃ 😀 b
	cases := map[int64]string{0: "a", 1: "é", 2: "☃", 3: "\U0001F600", 4: "b"}
	for idx, want := range cases {
		got, ok := RuneAtIndex(s, idx)
		if !ok || got != want {
			t.Fatalf("rune %d: got %q ok=%v want %q", idx, got, ok, want)
		}
	}
	if _, ok := RuneAtIndex(s, 5); ok {
		t.Fatal("index past end should fail")
	}

	// Invalid UTF-8 decodes like []rune conversion: one U+FFFD per bad byte
	bad := string([]byte{0xff, 'x', 0xfe})
	for idx := int64(0); idx < 3; idx++ {
		got, ok := RuneAtIndex(bad, idx)
		if !ok {
			t.Fatalf("invalid utf8 index %d should decode", idx)
		}
		if want := string([]rune(bad)[idx]); got != want {
			t.Fatalf("invalid utf8 index %d: got %q want %q", idx, got, want)
		}
	}

	if _, ok := RuneAtIndex("abc", -1); ok {
		t.Fatal("negative index should fail")
	}
	if _, ok := RuneAtIndex("", 0); ok {
		t.Fatal("empty string should fail")
	}
}

func TestRuneRange(t *testing.T) {
	s := "héllo wörld"
	// runes: h é l l o ' ' w ö r l d  => count 11
	tests := []struct {
		start, end int64
		want       string
		wantOK     bool
	}{
		{0, 11, s, true},
		{0, 0, "", true},
		{11, 11, "", true},
		{0, 1, "h", true},
		{0, 2, "hé", true},
		{1, 3, "él", true},
		{6, 8, "wö", true},
		{10, 11, "d", true},
		{2, 2, "", true},
		{12, 12, "", false}, // past end
		{0, 12, "", false},
		{-1, 3, "", false},
		{5, 3, "", false},
	}
	for _, tc := range tests {
		got, ok := RuneRange(s, tc.start, tc.end)
		if ok != tc.wantOK || (ok && got != tc.want) {
			t.Fatalf("RuneRange(%d,%d): got %q ok=%v want %q ok=%v", tc.start, tc.end, got, ok, tc.want, tc.wantOK)
		}
	}

	// Empty source string with an empty range is valid
	if got, ok := RuneRange("", 0, 0); !ok || got != "" {
		t.Fatalf("empty string empty range: got %q ok=%v", got, ok)
	}

	// A window containing invalid UTF-8 reports not-ok so callers fall back
	// to direct []rune slicing (which rewrites each bad byte to U+FFFD)
	bad := string([]byte{'x', 0xff, 0xff, 'y'})
	if _, ok := RuneRange(bad, 1, 3); ok {
		t.Fatal("window with invalid utf8 should report not-ok")
	}
	// Windows outside the invalid region still slice directly
	if got, ok := RuneRange(bad, 3, 4); !ok || got != "y" {
		t.Fatalf("after-invalid window: got %q ok=%v", got, ok)
	}
}
