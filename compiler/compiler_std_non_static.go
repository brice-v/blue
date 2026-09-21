//go:build !static

package compiler

import "blue/lib"

func init() {
	_std_mods["ui"] = &StdModFile{File: lib.ReadStdFileToString("ui.b")}
	_std_mods["gg"] = &StdModFile{File: lib.ReadStdFileToString("gg.b")}
	_std_mods["plot"] = &StdModFile{File: lib.ReadStdFileToString("plot.b")}
}
