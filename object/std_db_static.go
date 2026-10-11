//go:build !wasm && static

package object

import "blue/consts"

const (
	duckdbAvailable = false
	duckdbHelpNote  = " (" + consts.DuckDbLinkNote + ")"
)
