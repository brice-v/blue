//go:build !cgo && (raylib_no_embed || (!windows && !linux && !darwin) || (!amd64 && !arm64))
// +build !cgo
// +build raylib_no_embed !windows,!linux,!darwin !amd64,!arm64

package rl

import (
	_ "embed"
)

// extractLib extracts the embedded shared library and returns the path to it
func extractLib() (libname string) {
	return
}
