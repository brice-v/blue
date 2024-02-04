//go:build static
// +build static

package evaluator

func setupBuiltinsWithEvaluator(name string, newE *Evaluator) {
	if name == "http" {
		_http_builtin_map.Put("_handle", createHttpHandleBuiltin(newE, false))
		_http_builtin_map.Put("_handle_use", createHttpHandleBuiltin(newE, true))
		_http_builtin_map.Put("_handle_ws", createHttpHandleWSBuiltin(newE))
	}
}
