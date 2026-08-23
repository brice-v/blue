//go:build wasm

package vm

import (
	"blue/object"
)

// getUIStdBuiltin for wasm builds. The fyne toolkit has no js/wasm support
// (and no browser build registers ui builtins), so nothing resolves here.
func getUIStdBuiltin(name string, vm *VM) *object.Builtin {
	return nil
}
