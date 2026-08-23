package object

import (
	"strconv"
	"sync/atomic"
	"unicode/utf8"
)

// Dynamic-string interning: a fixed-size direct-mapped cache of recently
// seen dynamic strings (str(int) results, indexed characters, map keys,
// short interpolation results). Repeats return the shared read-only
// Stringo instead of allocating another wrapper. Slots use atomic pointers
// so concurrent processes can race safely; a race only wastes an entry,
// never misbehaves. Memory stays bounded because the table never grows.

const (
	internTableSize = 4096 // power of two
	internMaxLen    = 32   // longer dynamic strings rarely repeat exactly
)

var internTable [internTableSize]atomic.Pointer[Stringo]

const (
	internFnvOffset = 2166136261
	internFnvPrime  = 16777619
)

func internHashString(s string) uint32 {
	h := uint32(internFnvOffset)
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= internFnvPrime
	}
	return h
}

func internHashBytes(b []byte) uint32 {
	h := uint32(internFnvOffset)
	for i := 0; i < len(b); i++ {
		h ^= uint32(b[i])
		h *= internFnvPrime
	}
	return h
}

// internBytesEq compares without allocating (string(b) == s would allocate
// on some paths).
func internBytesEq(s string, b []byte) bool {
	if len(s) != len(b) {
		return false
	}
	for i := 0; i < len(b); i++ {
		if s[i] != b[i] {
			return false
		}
	}
	return true
}

// InternString returns a Stringo holding s, sharing cached objects for ""
// and single-byte ASCII strings and interning short dynamic strings so
// repeated values reuse one object. Strings longer than internMaxLen are
// never interned.
func InternString(s string) *Stringo {
	switch len(s) {
	case 0:
		return EmptyString
	case 1:
		if b := s[0]; b < utf8.RuneSelf {
			return asciiStrings[b]
		}
	default:
		if len(s) > internMaxLen {
			return &Stringo{Value: s}
		}
	}
	slot := &internTable[internHashString(s)&(internTableSize-1)]
	if e := slot.Load(); e != nil && e.Value == s {
		return e
	}
	n := &Stringo{Value: s}
	slot.Store(n)
	return n
}

// InternInt formats v as a decimal string and interns it. Formatting goes
// through a stack buffer so cache hits perform no heap allocation at all,
// which is the common case for repeated str(int) results.
func InternInt(v int64) *Stringo {
	if v >= 0 && v < 10 {
		return asciiStrings['0'+v]
	}
	var buf [20]byte // enough for any int64 including sign
	b := strconv.AppendInt(buf[:0], v, 10)
	slot := &internTable[internHashBytes(b)&(internTableSize-1)]
	if e := slot.Load(); e != nil && internBytesEq(e.Value, b) {
		return e
	}
	n := &Stringo{Value: string(b)}
	slot.Store(n)
	return n
}
