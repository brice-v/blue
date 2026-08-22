//go:build !cgo && !raylib_no_embed
// +build !cgo,!raylib_no_embed

package rl

import _ "embed"

//go:embed libs/raylib-6.0_winarm64_msvc16.tar.gz
var embeddedLib []byte
