package ast

import (
	"blue/token"
	"testing"
)

func TestString(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&VarStatement{
				Token: token.Token{Type: token.VAR, Literal: "var"},
				Name: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "myVar"},
					Value: "myVar",
				},
				Value: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "anotherVar"},
					Value: "anotherVar",
				},
				AssignmentToken: token.Token{Type: token.ASSIGN, Literal: "="},
			},
		},
	}

	if program.String() != "var myVar = anotherVar;\n" {
		t.Errorf("program.String() wrong. got=%q", program.String())
	}
}

func TestString2(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&VarStatement{
				Token: token.Token{Type: token.VAR, Literal: "var"},
				Name: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "myVar"},
					Value: "myVar",
				},
				Value: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "anotherVar"},
					Value: "anotherVar",
				},
				AssignmentToken: token.Token{Type: token.PLUSEQ, Literal: "+="},
			},
		},
	}

	if program.String() != "var myVar += anotherVar;\n" {
		t.Errorf("program.String() wrong. got=%q", program.String())
	}
}
