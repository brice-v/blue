//go:build !static
// +build !static

package evaluator

import (
	"blue/lib"
	"blue/object"
)

func init() {
	_std_mods["ui"] = StdModFileAndBuiltins{File: lib.ReadStdFileToString("ui.b"), Builtins: _ui_builtin_map}
	_std_mods["gg"] = StdModFileAndBuiltins{File: lib.ReadStdFileToString("gg.b"), Builtins: _gg_builtin_map}
}

func setupBuiltinsWithEvaluator(name string, newE *Evaluator) {
	switch name {
	case "http":
		_http_builtin_map.Put("_handle", createHttpHandleBuiltin(newE, false))
		_http_builtin_map.Put("_handle_use", createHttpHandleBuiltin(newE, true))
		_http_builtin_map.Put("_handle_ws", createHttpHandleWSBuiltin(newE))
	case "ui":
		_ui_builtin_map.Put("_button", createUIButtonBuiltin(newE))
		_ui_builtin_map.Put("_check_box", createUICheckBoxBuiltin(newE))
		_ui_builtin_map.Put("_radio_group", createUIRadioBuiltin(newE))
		_ui_builtin_map.Put("_option_select", createUIOptionSelectBuiltin(newE))
		_ui_builtin_map.Put("_form", createUIFormBuiltin(newE))
		_ui_builtin_map.Put("_toolbar_action", createUIToolbarAction(newE))
	}
}

var _gg_builtin_map = NewBuiltinMap(object.GgBuiltins)

var _ui_builtin_map = NewBuiltinMap(object.UiBuiltins)
