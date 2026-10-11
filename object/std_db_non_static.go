//go:build !wasm && !static

package object

import _ "github.com/duckdb/duckdb-go/v2"

const (
	duckdbAvailable = true
	duckdbHelpNote  = ""
)
