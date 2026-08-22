//go:build !cgo && !raylib_no_embed
// +build !cgo,!raylib_no_embed

package rl

import _ "embed"

//go:embed libs/raylib-6.0_linux_arm64.tar.gz
var embeddedLib []byte
