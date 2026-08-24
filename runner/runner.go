// Package runner holds the shared "run a compiled program image" path used
// by the full blue binary (package cmd) and by the minimal standalone runner
// (cmd/bluerun, built with -tags minivm).
//
// It deliberately avoids importing the lexer/parser/compiler toolchain: the
// only non-trivial dependency is lexer.GetErrorLineMessage, which formats
// token positions against source files (including embedded core/std sources)
// and pulls in nothing heavier than data embeds. Everything else comes from
// the closed runtime set: bluec, vm, object, consts, blueutil.
package runner

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"blue/bluec"
	"blue/blueutil"
	"blue/consts"
	"blue/lexer"
	"blue/object"
	"blue/token"
	"blue/vm"
)

// RunBytecode executes a program image and returns the process exit code:
// 0 on success, 1 when the program raised a VM/runtime error. Errors are
// printed to stderr using the same colored trace formatting as the full
// binary. When printResult is set (the `vm` command behavior), the value of
// the last evaluated expression is written to stdout.
func RunBytecode(bc *bluec.Bytecode, noExec, printResult bool) int {
	globals := make([]object.Object, vm.GlobalsSize)
	v := vm.NewWithGlobalsStore(bc, globals)
	object.NoExec = noExec
	if err := v.Run(); err != nil {
		printVMError(v.TokensForErrorTrace, err.Error())
		return 1
	}
	val := v.LastPoppedStackElem()
	if val.Type() == object.ERROR_OBJ {
		printVMErrorMessageOnly(val.(*object.Error).Message)
		return 1
	}
	if printResult {
		if _, err := fmt.Fprintf(os.Stdout, "%s\n", val.Inspect()); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to write result:", err)
			return 1
		}
	}
	return 0
}

// printVMError renders a runtime error raised by vm.Run(), walking any
// recorded token trace so users get file/line pointers. Images compiled
// without tokens still print the message, just without position info.
func printVMError(tokens []*token.Token, message string) {
	if len(tokens) == 0 {
		consts.ErrorPrinter("%s%s\n", consts.VM_ERROR_PREFIX, message)
		return
	}
	for i, tok := range tokens {
		errorLine := lexer.GetErrorLineMessage(*tok)
		fullMsg := fmt.Sprintf("%s\n%s", message, errorLine)
		blueutil.PrintCustomError(os.Stderr, consts.VM_ERROR_PREFIX, fullMsg, tok.LineNumber, i == 0)
	}
}

// printVMErrorMessageOnly preserves the historical formatting used when a
// run finishes and leaves an ERROR object as its result.
func printVMErrorMessageOnly(message string) {
	var buf bytes.Buffer
	buf.WriteString(message)
	buf.WriteByte('\n')
	msg := fmt.Sprintf("%s%s", consts.VM_ERROR_PREFIX, buf.String())
	splitMsg := strings.Split(msg, "\n")
	for i, s := range splitMsg {
		if i == 0 {
			consts.ErrorPrinter("%s\n", s)
			continue
		}
		delimeter := ""
		if i != len(splitMsg)-1 {
			delimeter = "\n"
		}
		if _, err := fmt.Fprintf(os.Stderr, "%s%s", s, delimeter); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to write to output:", err)
			break
		}
	}
}
