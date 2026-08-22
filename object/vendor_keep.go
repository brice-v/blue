//go:build required
// +build required

package object

// This file is never compiled: normal builds do not set the "required" tag.
// However, 'go mod tidy' and 'go mod vendor' scan imports with every build
// tag satisfied, so the blank imports below force those directories to be
// copied into vendor/. Without this, raylib's RGFW backend sources
// ('-tags rgfw') are missing from vendor/ because they are only referenced
// from C #include directives behind preprocessor macros that vendoring
// cannot see.
//
// Mirrors upstream raylib-go's cgo_vendor.go, which keeps the equivalent
// external/glfw directories vendored the same way but omits external/RGFW.
import (
	_ "github.com/gen2brain/raylib-go/raylib/external/RGFW"
	_ "github.com/gen2brain/raylib-go/raylib/external/RGFW/deps"
)
