package convert

import (
	"strings"
	"unsafe"
)

// ToBytePtr converts a Go string to a null-terminated C-style string by just appending a null byte,
// if s doesn't already contain one.
func ToBytePtr(s string) *byte {
	size := len(s) + 1
	if index := strings.IndexByte(s, 0); index != -1 {
		size = index + 1
	}

	result := make([]byte, size)
	copy(result, s)
	return &result[0]
}

// ToBytePtrNullable does the same thing as [ToBytePtr], except that an empty string returns nil.
func ToBytePtrNullable(s string) *byte {
	if len(s) == 0 {
		return nil
	}
	return ToBytePtr(s)
}

// ToString converts a null-terminated C-style string into a Go string.
func ToString(p *byte) string {
	if p == nil {
		return ""
	}
	i := 0
	for ptr := unsafe.Pointer(p); *(*byte)(unsafe.Add(ptr, i)) != 0; i++ {
	}
	return string(unsafe.Slice(p, i))
}

// ToStringSlice converts a null-terminated list of C-style strings to a slice of Go strings.
func ToStringSlice(pointers **byte) []string {
	if pointers == nil {
		return nil
	}

	strings := make([]string, 0)

	for ptr := unsafe.Pointer(pointers); *(**byte)(ptr) != nil; ptr = unsafe.Add(ptr, 8) {
		strings = append(strings, ToString(*(**byte)(ptr)))
	}

	return strings
}
