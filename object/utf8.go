package object

import (
	"unicode/utf8"
)

// RuneAtIndex returns the substring of s starting at rune index idx,
// mirroring string([]rune(s)[idx]) semantics (invalid bytes decode to
// utf8.RuneError, exactly like a []rune conversion) without materializing
// a rune slice. ok is false when idx is negative or past the last rune,
// letting callers emit their own out-of-bounds error.
func RuneAtIndex(s string, idx int64) (string, bool) {
	if idx < 0 {
		return "", false
	}
	var n int64
	for off := 0; off < len(s); {
		if b := s[off]; b < utf8.RuneSelf {
			// ASCII fast path: skip the decode call entirely.
			if n == idx {
				return s[off : off+1], true
			}
			n++
			off++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[off:])
		if n == idx {
			if r == utf8.RuneError && size == 1 {
				// Invalid encoding: match []rune conversion output.
				return string(utf8.RuneError), true
			}
			return s[off : off+size], true
		}
		n++
		off += size
	}
	return "", false
}

// RuneRange returns the substring of s covering runes [start, end),
// mirroring string([]rune(s)[start:end]) without allocating rune slices.
// ok is false when the range falls outside the rune count or when the
// window contains invalid UTF-8 (whose replacement-char rewriting needs
// a []rune conversion); callers can fall back to direct slicing there.
func RuneRange(s string, start, end int64) (string, bool) {
	if start < 0 || end < start {
		return "", false
	}
	byteStart := -1
	var n int64
	for off := 0; off < len(s); {
		size := 1
		if s[off] >= utf8.RuneSelf {
			r, decodeSize := utf8.DecodeRuneInString(s[off:])
			size = decodeSize
			if r == utf8.RuneError && size == 1 && n >= start && n < end {
				// Invalid byte inside the requested window.
				return "", false
			}
		}
		if n == start {
			byteStart = off
			if start == end {
				return "", true
			}
		} else if n == end {
			return s[byteStart:off], true
		}
		n++
		off += size
	}
	if n == end {
		if byteStart < 0 {
			// start == end == rune count
			return "", true
		}
		return s[byteStart:], true
	}
	return "", false
}
