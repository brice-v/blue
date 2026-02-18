package vm

import "blue/object"

func GetStdBuiltinWithVm(mod, name string, vm *VM) func(args ...object.Object) object.Object {
	switch mod {
	case "http":
		switch name {
		case "_handle":
			return createHttpHandleBuiltin(vm, false).Fun
		case "_handle_use":
			return createHttpHandleBuiltin(vm, true).Fun
		case "_handle_ws":
			return createHttpHandleWSBuiltin(vm).Fun
		default:
			panic("GetStdBuiltinWithVm called with incorrect builtin function name '" + name + "' for module: " + mod)
		}
	case "ui":
		switch name {
		case "_button":
			return createUIButtonBuiltin(vm).Fun
		case "_check_box":
			return createUICheckBoxBuiltin(vm).Fun
		case "_radio_group":
			return createUIRadioBuiltin(vm).Fun
		case "_option_select":
			return createUIOptionSelectBuiltin(vm).Fun
		case "_form":
			return createUIFormBuiltin(vm).Fun
		case "_toolbar_action":
			return createUIToolbarAction(vm).Fun
		default:
			panic("GetStdBuiltinWithVm called with incorrect builtin function name '" + name + "' for module: " + mod)
		}
	}
	panic("GetStdBuiltinWithVm called with incorrect module: " + mod)
}

func createHttpHandleBuiltin(vm *VM, isHandleUse bool) *object.Builtin {
	return &object.Builtin{
		Fun: func(args ...object.Object) object.Object {
			return object.NULL
		},
	}
}

func createHttpHandleWSBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Fun: func(args ...object.Object) object.Object {
			return object.NULL
		},
	}
}

func createUIButtonBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Fun: func(args ...object.Object) object.Object {
			return object.NULL
		},
	}
}

func createUICheckBoxBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Fun: func(args ...object.Object) object.Object {
			return object.NULL
		},
	}
}

func createUIRadioBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Fun: func(args ...object.Object) object.Object {
			return object.NULL
		},
	}
}

func createUIOptionSelectBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Fun: func(args ...object.Object) object.Object {
			return object.NULL
		},
	}
}

func createUIFormBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Fun: func(args ...object.Object) object.Object {
			return object.NULL
		},
	}
}

func createUIToolbarAction(vm *VM) *object.Builtin {
	return &object.Builtin{
		Fun: func(args ...object.Object) object.Object {
			return object.NULL
		},
	}
}
