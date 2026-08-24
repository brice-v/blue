//go:build static || wasm

package vm

import (
	"blue/object"
)

// getUIStdBuiltin for static and wasm builds. The fyne toolkit has no
// js/wasm support and static builds drop all GUI dependencies (see
// object/std_ui_static.go), so nothing resolves here.
func getUIStdBuiltin(name string, vm *VM) *object.Builtin {
	return nil
}
