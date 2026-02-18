//go:build !static
// +build !static

package compiler

import "blue/lib"

func init() {
	_std_mods["ui"] = &StdModFile{File: lib.ReadStdFileToString("ui.b")}
	_std_mods["gg"] = &StdModFile{File: lib.ReadStdFileToString("gg.b")}
}
