package compiler

import (
	"blue/lexer"
	"blue/parser"
	"strings"
	"testing"
)

func TestCompileLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1", "&object.Integer{Value: 1}"},
		{"1.0", "&object.Float{Value: 1.000000}"},
		{"0x05", "&object.UInteger{Value: 5}"},
	}
	for _, test := range tests {
		l := lexer.New(test.input, "<test:compileobject>")
		p := parser.New(l)
		prog := p.ParseProgram()
		if len(p.Errors()) != 0 {
			t.Fatalf("parser contained errors for input `%s`: errors: %s", test.input, strings.Join(p.Errors(), ","))
		}
		result := Compile(prog)
		if result != test.expected {
			t.Fatalf("result `%s` did not match expected `%s`", result, test.expected)
		}
	}
}
