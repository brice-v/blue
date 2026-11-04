package b_program_test

import (
	"blue/ast"
	"blue/compiler"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"blue/vm"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"
)

// Test-only glue: full-build binaries install vm.EvalHook from their own
// entrypoints (see cmd/util.go). Tests run against the raw VM so they
// install it here to exercise programs using eval/to_num/load.
func TestMain(m *testing.M) {
	vm.EvalHook = evalForTests
	vm.ParseNumHook = parseNumForTests
	os.Exit(m.Run())
}

func evalForTests(src string) object.Object {
	l := lexer.New(src, "<internal:string>")
	p := parser.New(l)
	prog := p.ParseProgram()
	if p.HasErrors() {
		return &object.Error{Message: "failed to `eval` string"}
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

func parseNumForTests(src string) object.Object {
	negate := false
	if strings.HasPrefix(src, "-") {
		negate = true
		src = strings.TrimPrefix(src, "-")
	}
	l := lexer.New(src, "<internal:string>")
	p := parser.New(l)
	prog := p.ParseProgram()
	if p.HasErrors() {
		return &object.Error{Message: fmt.Sprintf("failed to `parse_num` string, found '%d' parser errors", len(p.ErrorMessages()))}
	}
	if len(prog.Statements) != 1 {
		return &object.Error{Message: fmt.Sprintf("failed to `parse_num` string, expected 1 statement got %d", len(prog.Statements))}
	}
	stmt := prog.Statements[0]
	expr, ok := stmt.(*ast.ExpressionStatement)
	if !ok {
		return &object.Error{Message: fmt.Sprintf("failed to `parse_num` string, expected *ast.ExpressionStatement got %T", stmt)}
	}
	switch v := expr.Expression.(type) {
	case *ast.IntegerLiteral:
		if negate {
			return object.NewInteger(-v.Value)
		}
		return object.NewInteger(v.Value)
	case *ast.BigIntegerLiteral:
		if negate {
			return &object.BigInteger{Value: new(big.Int).Neg(v.Value)}
		}
		return &object.BigInteger{Value: v.Value}
	case *ast.FloatLiteral:
		if negate {
			return &object.Float{Value: -v.Value}
		}
		return &object.Float{Value: v.Value}
	case *ast.BigFloatLiteral:
		if negate {
			return &object.BigFloat{Value: v.Value.Neg()}
		}
		return &object.BigFloat{Value: v.Value}
	case *ast.UIntegerLiteral:
		return &object.UInteger{Value: v.Value}
	case *ast.BinaryLiteral:
		return &object.UInteger{Value: v.Value}
	case *ast.OctalLiteral:
		return &object.UInteger{Value: v.Value}
	case *ast.HexLiteral:
		return &object.UInteger{Value: v.Value}
	}
	return &object.Error{Message: fmt.Sprintf("failed to `parse_num` string, expected numeric literal got %T", expr.Expression)}
}
