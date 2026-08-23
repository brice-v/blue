//go:build js && wasm

// Command wasmmain is the blue language runtime compiled to WebAssembly for
// the browser playground (see make_wasm.sh and the playground directory).
//
// The interpreter runs entirely inside the wasm sandbox: no filesystem,
// process, or shell access is wired up, shell execution is hard disabled via
// object.NoExec, and the only capability exposed to JavaScript is running
// blue source code and collecting its output.
//
// Output capture works together with the playground page: go_js_wasm_exec
// routes every write to fds 1/2 through fs.writeSync, and the page wraps
// that single choke point, exposing the collected text back via the
// blueCollectOutput callback.
package main

import (
	"fmt"
	"strings"
	"syscall/js"
	"time"

	"blue/blueutil"
	"blue/compiler"
	"blue/consts"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"blue/vm"

	"github.com/gookit/color"
)

const sourceName = "<playground>"

func main() {
	// Plain text output only, no ANSI escape sequences in playground output.
	color.Disable()
	// Belt and braces: never allow shell outs even if a builtin tries.
	object.NoExec = true

	js.Global().Set("blueVersion", js.FuncOf(func(this js.Value, args []js.Value) any {
		return "blue v" + consts.VERSION + " (wasm)"
	}))
	js.Global().Set("blueRun", js.FuncOf(blueRun))

	// Keep the go runtime alive so registered callbacks keep working. A
	// plain select{} would trip the deadlock detector while a callback is
	// suspended, so keep a (very cheap) timer pending instead.
	for {
		time.Sleep(time.Hour)
	}
}

// blueRun is the js entry point. It takes a single string of blue source
// code and returns {error, result}. Everything the program prints flows to
// stdout/stderr where the playground page captures it and attaches it to
// the returned object as the output field.
func blueRun(this js.Value, args []js.Value) any {
	result := map[string]any{
		"error":  "",
		"result": "",
	}
	if len(args) != 1 || args[0].Type() != js.TypeString {
		result["error"] = "blueRun expects a single string argument"
		return result
	}

	runErr, resultVal := safeRunProgram(args[0].String())
	result["error"] = runErr
	result["result"] = resultVal
	return result
}

// safeRunProgram runs the given source, converting any go panic into an
// error message so one bad program cannot take down the whole wasm module.
func safeRunProgram(src string) (errMsg string, resultVal string) {
	defer func() {
		if r := recover(); r != nil {
			errMsg = fmt.Sprintf("[ERROR] VMError: internal error: %v", r)
			resultVal = ""
		}
	}()
	return runProgram(src)
}

// evalSourceString is the wasm build's implementation of vm.EvalHook. It
// mirrors the original vm.Str behavior: lex, parse, compile and run the
// source on a fresh VM, returning its last stack value.
func evalSourceString(src string) object.Object {
	l := lexer.New(src, "<internal:string>")
	p := parser.New(l)
	prog := p.ParseProgram()
	if p.HasErrors() {
		return &object.Error{Message: fmt.Sprintf("failed to `eval` string, found '%d' parser errors", len(p.ErrorMessages()))}
	}
	c := compiler.New()
	if err := c.Compile(prog); err != nil {
		return &object.Error{Message: "compiler error in `eval` string: " + err.Error()}
	}
	vmInstance := vm.New(c.Bytecode())
	if err := vmInstance.Run(); err != nil {
		return &object.Error{Message: "vm error in `eval` string: " + err.Error()}
	}
	return vmInstance.LastPoppedStackElem()
}

// runProgram lexes, parses, compiles and runs the given source. Program
// output flows to stdout/stderr where the page collects it. It mirrors
// cmd.vmFileOrString but never exits or writes to the real terminal.
func runProgram(src string) (errMsg string, resultVal string) {
	l := lexer.New(src, sourceName)
	p := parser.NewWithStopAfterFirst(l)
	program := p.ParseProgram()
	if p.HasErrors() {
		var msg strings.Builder
		p.PrintParserErrors(&msg)
		return msg.String(), ""
	}

	constants := object.NewObjectConstants()
	symbolTable := compiler.NewSymbolTable()
	for i, v := range object.AllBuiltins[0].Builtins {
		symbolTable.DefineBuiltin(i, v.Name, 0, v.Help())
	}
	for i, v := range object.BuiltinobjsList {
		symbolTable.DefineBuiltin(i, v.Name, object.BuiltinobjsModuleIndex, v.Builtin.Help())
	}
	c := compiler.NewWithStateAndCore(symbolTable, constants)
	if err := c.Compile(program); err != nil {
		errToPrint, _, _ := strings.Cut(err.Error(), "\n"+consts.INTERNAL_ERROR_PATTERN)
		var trace strings.Builder
		writeCompilerStackTrace(&trace, c.ErrorTrace)
		return consts.COMPILER_ERROR_PREFIX + errToPrint + "\n" + trace.String(), ""
	}

	v := vm.NewWithGlobalsStore(c.Bytecode(), make([]object.Object, vm.GlobalsSize))
	// The wasm build ships the full toolchain, so runtime evaluation
	// (`eval`, `to_num`, `load`) keeps working: install the hook before
	// running, mirroring what cmd does for the desktop binary.
	vm.EvalHook = evalSourceString
	if err := v.Run(); err != nil {
		if v.TokensForErrorTrace == nil {
			return consts.VM_ERROR_PREFIX + err.Error() + "\n", ""
		}
		var msg strings.Builder
		for i, tok := range v.TokensForErrorTrace {
			errorLine := lexer.GetErrorLineMessage(*tok)
			fullMsg := fmt.Sprintf("%s\n%s", err.Error(), errorLine)
			blueutil.PrintCustomError(&msg, consts.VM_ERROR_PREFIX, fullMsg, tok.LineNumber, i == 0)
		}
		return msg.String(), ""
	}

	val := v.LastPoppedStackElem()
	if val != nil && val.Type() == object.ERROR_OBJ {
		errorObj := val.(*object.Error)
		return consts.VM_ERROR_PREFIX + errorObj.Message + "\n", ""
	}
	if val != nil {
		resultVal = val.Inspect()
	}
	return "", resultVal
}

// writeCompilerStackTrace writes the deduplicated compiler error trace lines
// the way compiler.Compiler.PrintStackTrace does, but without touching the
// real stdout.
func writeCompilerStackTrace(out *strings.Builder, traces []string) {
	const ignoreStr = `Filepath: "", LineNumber: 0, PositionInLine: 0
`
	prevS := ""
	for _, s := range traces {
		if s != prevS && s != ignoreStr {
			out.WriteString(s)
		}
		prevS = s
	}
}
