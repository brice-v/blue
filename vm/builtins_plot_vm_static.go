//go:build static

package vm

import "blue/object"

// plotVmBuiltin resolves nothing in a static build: the plot module is omitted,
// so no plot builtin can be vm bound.
func plotVmBuiltin(string, *VM) *object.Builtin { return nil }
