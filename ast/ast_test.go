package ast

import (
	"blue/token"
	"math/big"
	"testing"

	"github.com/shopspring/decimal"
)

func TestString(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&VarStatement{
				Token: token.Token{Type: token.VAR, Literal: "var"},
				Names: []*Identifier{{
					Token: token.Token{Type: token.IDENT, Literal: "myVar"},
					Value: "myVar",
				}},
				Value: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "anotherVar"},
					Value: "anotherVar",
				},
				AssignmentToken: token.Token{Type: token.ASSIGN, Literal: "="},
			},
		},
	}

	if program.String() != "var myVar = anotherVar" {
		t.Errorf("program.String() wrong. got=%q", program.String())
	}
}

func TestString2(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&VarStatement{
				Token: token.Token{Type: token.VAR, Literal: "var"},
				Names: []*Identifier{{
					Token: token.Token{Type: token.IDENT, Literal: "myVar"},
					Value: "myVar",
				}},
				Value: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "anotherVar"},
					Value: "anotherVar",
				},
				AssignmentToken: token.Token{Type: token.PLUSEQ, Literal: "+="},
			},
		},
	}

	if program.String() != "var myVar += anotherVar" {
		t.Errorf("program.String() wrong. got=%q", program.String())
	}
}

func TestProgramTokenToken(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&ExpressionStatement{
				Token: token.Token{Type: token.IDENT, Literal: "hello"},
			},
		},
	}
	tt := program.TokenToken()
	if tt.Type != "" {
		t.Errorf("expected empty token, got %v", tt)
	}
}

func TestProgramTokenLiteralEmpty(t *testing.T) {
	program := &Program{
		Statements: []Statement{},
	}
	result := program.TokenLiteral()
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestProgramStringEmpty(t *testing.T) {
	program := &Program{
		Statements: []Statement{},
	}
	result := program.String()
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestProgramStringMultipleStatements(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&ExpressionStatement{
				Token: token.Token{Type: token.IDENT, Literal: "a"},
				Expression: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "a"},
					Value: "a",
				},
			},
			&ExpressionStatement{
				Token: token.Token{Type: token.IDENT, Literal: "b"},
				Expression: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "b"},
					Value: "b",
				},
			},
		},
	}
	result := program.String()
	expected := "a\nb"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// --- Identifier Tests ---

func TestIdentifierTokenLiteral(t *testing.T) {
	ident := &Identifier{
		Token: token.Token{Type: token.IDENT, Literal: "myIdentifier"},
		Value: "myIdentifier",
	}
	if ident.TokenLiteral() != "myIdentifier" {
		t.Errorf("TokenLiteral() = %q, want %q", ident.TokenLiteral(), "myIdentifier")
	}
}

func TestIdentifierTokenToken(t *testing.T) {
	expectedToken := token.Token{Type: token.IDENT, Literal: "x"}
	ident := &Identifier{
		Token: expectedToken,
		Value: "x",
	}
	got := ident.TokenToken()
	if got.Type != expectedToken.Type || got.Literal != expectedToken.Literal {
		t.Errorf("TokenToken() = %v, want %v", got, expectedToken)
	}
}

func TestIdentifierString(t *testing.T) {
	ident := &Identifier{
		Token: token.Token{Type: token.IDENT, Literal: "foo"},
		Value: "foo",
	}
	if ident.String() != "foo" {
		t.Errorf("String() = %q, want %q", ident.String(), "foo")
	}
}

// --- Null Tests ---

func TestNullTokenLiteral(t *testing.T) {
	n := &Null{
		Token: token.Token{Type: token.NULL_KW, Literal: "null"},
	}
	if n.TokenLiteral() != "null" {
		t.Errorf("TokenLiteral() = %q, want %q", n.TokenLiteral(), "null")
	}
}

func TestNullString(t *testing.T) {
	n := &Null{
		Token: token.Token{Type: token.NULL_KW, Literal: "null"},
	}
	if n.String() != "null" {
		t.Errorf("String() = %q, want %q", n.String(), "null")
	}
}

// --- Boolean Tests ---

func TestBooleanTokenLiteral(t *testing.T) {
	b := &Boolean{
		Token: token.Token{Type: token.TRUE, Literal: "true"},
		Value: true,
	}
	if b.TokenLiteral() != "true" {
		t.Errorf("TokenLiteral() = %q, want %q", b.TokenLiteral(), "true")
	}
}

func TestBooleanStringTrue(t *testing.T) {
	b := &Boolean{
		Token: token.Token{Type: token.TRUE, Literal: "true"},
		Value: true,
	}
	if b.String() != "true" {
		t.Errorf("String() = %q, want %q", b.String(), "true")
	}
}

func TestBooleanStringFalse(t *testing.T) {
	b := &Boolean{
		Token: token.Token{Type: token.FALSE, Literal: "false"},
		Value: false,
	}
	if b.String() != "false" {
		t.Errorf("String() = %q, want %q", b.String(), "false")
	}
}

// --- PrefixExpression Tests ---

func TestPrefixExpressionTokenLiteral(t *testing.T) {
	pe := &PrefixExpression{
		Token:    token.Token{Type: token.BANG, Literal: "!"},
		Operator: "!",
		Right:    &Boolean{Token: token.Token{Type: token.TRUE, Literal: "true"}, Value: true},
	}
	if pe.TokenLiteral() != "!" {
		t.Errorf("TokenLiteral() = %q, want %q", pe.TokenLiteral(), "!")
	}
}

func TestPrefixExpressionString(t *testing.T) {
	pe := &PrefixExpression{
		Token:    token.Token{Type: token.MINUS, Literal: "-"},
		Operator: "-",
		Right:    &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "5"}, Value: 5},
	}
	expected := "(-5)"
	if pe.String() != expected {
		t.Errorf("String() = %q, want %q", pe.String(), expected)
	}
}

func TestPrefixExpressionWithBoolean(t *testing.T) {
	pe := &PrefixExpression{
		Token:    token.Token{Type: token.BANG, Literal: "!"},
		Operator: "!",
		Right:    &Boolean{Token: token.Token{Type: token.TRUE, Literal: "true"}, Value: true},
	}
	expected := "(!true)"
	if pe.String() != expected {
		t.Errorf("String() = %q, want %q", pe.String(), expected)
	}
}

// --- PostfixExpression Tests ---

func TestPostfixExpressionTokenLiteral(t *testing.T) {
	pe := &PostfixExpression{
		Token:    token.Token{Type: token.IDENT, Literal: "x"},
		Operator: "!",
		Left:     &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
	}
	if pe.TokenLiteral() != "x" {
		t.Errorf("TokenLiteral() = %q, want %q", pe.TokenLiteral(), "x")
	}
}

func TestPostfixExpressionString(t *testing.T) {
	pe := &PostfixExpression{
		Token:    token.Token{Type: token.IDENT, Literal: "x"},
		Operator: "!",
		Left:     &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
	}
	expected := "(x!)"
	if pe.String() != expected {
		t.Errorf("String() = %q, want %q", pe.String(), expected)
	}
}

// --- InfixExpression Tests ---

func TestInfixExpressionTokenLiteral(t *testing.T) {
	ie := &InfixExpression{
		Token:    token.Token{Type: token.PLUS, Literal: "+"},
		Operator: "+",
		Left:     &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "3"}, Value: 3},
		Right:    &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "4"}, Value: 4},
	}
	if ie.TokenLiteral() != "+" {
		t.Errorf("TokenLiteral() = %q, want %q", ie.TokenLiteral(), "+")
	}
}

func TestInfixExpressionString(t *testing.T) {
	ie := &InfixExpression{
		Token:    token.Token{Type: token.PLUS, Literal: "+"},
		Operator: "+",
		Left:     &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "3"}, Value: 3},
		Right:    &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "4"}, Value: 4},
	}
	expected := "(3 + 4)"
	if ie.String() != expected {
		t.Errorf("String() = %q, want %q", ie.String(), expected)
	}
}

func TestInfixExpressionWithBoolean(t *testing.T) {
	ie := &InfixExpression{
		Token:    token.Token{Type: token.EQ, Literal: "=="},
		Operator: "==",
		Left:     &Boolean{Token: token.Token{Type: token.TRUE, Literal: "true"}, Value: true},
		Right:    &Boolean{Token: token.Token{Type: token.FALSE, Literal: "false"}, Value: false},
	}
	expected := "(true == false)"
	if ie.String() != expected {
		t.Errorf("String() = %q, want %q", ie.String(), expected)
	}
}

// --- IfExpression Tests ---

func TestIfExpressionTokenLiteral(t *testing.T) {
	ie := &IfExpression{
		Token: token.Token{Type: token.IF, Literal: "if"},
		Conditions: []Expression{
			&Boolean{Token: token.Token{Type: token.TRUE, Literal: "true"}, Value: true},
		},
		Consequences: []*BlockStatement{
			{
				Token:      token.Token{Type: token.LBRACE, Literal: "{"},
				Statements: []Statement{},
			},
		},
	}
	if ie.TokenLiteral() != "if" {
		t.Errorf("TokenLiteral() = %q, want %q", ie.TokenLiteral(), "if")
	}
}

func TestIfExpressionStringSimple(t *testing.T) {
	ie := &IfExpression{
		Token: token.Token{Type: token.IF, Literal: "if"},
		Conditions: []Expression{
			&Boolean{Token: token.Token{Type: token.TRUE, Literal: "true"}, Value: true},
		},
		Consequences: []*BlockStatement{
			{
				Token:      token.Token{Type: token.LBRACE, Literal: "{"},
				Statements: []Statement{
					&ExpressionStatement{
						Token:      token.Token{Type: token.IDENT, Literal: "x"},
						Expression: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
					},
				},
			},
		},
	}
	result := ie.String()
	if result == "" {
		t.Errorf("String() should not be empty")
	}
}

func TestIfExpressionStringWithElse(t *testing.T) {
	ie := &IfExpression{
		Token: token.Token{Type: token.IF, Literal: "if"},
		Conditions: []Expression{
			&Boolean{Token: token.Token{Type: token.TRUE, Literal: "true"}, Value: true},
		},
		Consequences: []*BlockStatement{
			{
				Token:      token.Token{Type: token.LBRACE, Literal: "{"},
				Statements: []Statement{},
			},
		},
		Alternative: &BlockStatement{
			Token:      token.Token{Type: token.LBRACE, Literal: "{"},
			Statements: []Statement{},
		},
	}
	result := ie.String()
	if result == "" {
		t.Errorf("String() should not be empty")
	}
}

func TestIfExpressionStringMultipleConditions(t *testing.T) {
	ie := &IfExpression{
		Token: token.Token{Type: token.IF, Literal: "if"},
		Conditions: []Expression{
			&Boolean{Token: token.Token{Type: token.TRUE, Literal: "true"}, Value: true},
			&Boolean{Token: token.Token{Type: token.FALSE, Literal: "false"}, Value: false},
		},
		Consequences: []*BlockStatement{
			{
				Token:      token.Token{Type: token.LBRACE, Literal: "{"},
				Statements: []Statement{},
			},
			{
				Token:      token.Token{Type: token.LBRACE, Literal: "{"},
				Statements: []Statement{},
			},
		},
	}
	result := ie.String()
	if result == "" {
		t.Errorf("String() should not be empty")
	}
}

// --- MatchExpression Tests ---

func TestMatchExpressionTokenLiteral(t *testing.T) {
	me := &MatchExpression{
		Token: token.Token{Type: token.MATCH, Literal: "match"},
		OptionalValue: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "x"},
			Value: "x",
		},
		Conditions: []Expression{
			&Boolean{Token: token.Token{Type: token.TRUE, Literal: "true"}, Value: true},
		},
		Consequences: []*BlockStatement{
			{
				Token:      token.Token{Type: token.LBRACE, Literal: "{"},
				Statements: []Statement{},
			},
		},
	}
	if me.TokenLiteral() != "match" {
		t.Errorf("TokenLiteral() = %q, want %q", me.TokenLiteral(), "match")
	}
}

func TestMatchExpressionString(t *testing.T) {
	me := &MatchExpression{
		Token: token.Token{Type: token.MATCH, Literal: "match"},
		OptionalValue: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "x"},
			Value: "x",
		},
		Conditions: []Expression{
			&Boolean{Token: token.Token{Type: token.TRUE, Literal: "true"}, Value: true},
		},
		Consequences: []*BlockStatement{
			{
				Token:      token.Token{Type: token.LBRACE, Literal: "{"},
				Statements: []Statement{},
			},
		},
	}
	result := me.String()
	if result == "" {
		t.Errorf("String() should not be empty")
	}
}

// --- CallExpression Tests ---

func TestCallExpressionTokenLiteral(t *testing.T) {
	ce := &CallExpression{
		Token:    token.Token{Type: token.LPAREN, Literal: "("},
		Function: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "foo"}, Value: "foo"},
		Arguments: []Expression{
			&IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "1"}, Value: 1},
		},
	}
	if ce.TokenLiteral() != "(" {
		t.Errorf("TokenLiteral() = %q, want %q", ce.TokenLiteral(), "(")
	}
}

func TestCallExpressionStringNoArgs(t *testing.T) {
	ce := &CallExpression{
		Token:     token.Token{Type: token.LPAREN, Literal: "("},
		Function:  &Identifier{Token: token.Token{Type: token.IDENT, Literal: "foo"}, Value: "foo"},
		Arguments: []Expression{},
	}
	expected := "foo()"
	if ce.String() != expected {
		t.Errorf("String() = %q, want %q", ce.String(), expected)
	}
}

func TestCallExpressionStringWithArgs(t *testing.T) {
	ce := &CallExpression{
		Token:    token.Token{Type: token.LPAREN, Literal: "("},
		Function: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "foo"}, Value: "foo"},
		Arguments: []Expression{
			&IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "1"}, Value: 1},
			&IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "2"}, Value: 2},
		},
	}
	expected := "foo(1, 2)"
	if ce.String() != expected {
		t.Errorf("String() = %q, want %q", ce.String(), expected)
	}
}

// --- IndexExpression Tests ---

func TestIndexExpressionTokenLiteral(t *testing.T) {
	ie := &IndexExpression{
		Token: token.Token{Type: token.LBRACKET, Literal: "["},
		Left:  &Identifier{Token: token.Token{Type: token.IDENT, Literal: "arr"}, Value: "arr"},
		Index: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "0"}, Value: 0},
	}
	if ie.TokenLiteral() != "[" {
		t.Errorf("TokenLiteral() = %q, want %q", ie.TokenLiteral(), "[")
	}
}

func TestIndexExpressionStringBracket(t *testing.T) {
	ie := &IndexExpression{
		Token: token.Token{Type: token.LBRACKET, Literal: "["},
		Left:  &Identifier{Token: token.Token{Type: token.IDENT, Literal: "arr"}, Value: "arr"},
		Index: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "0"}, Value: 0},
	}
	expected := "(arr[0])"
	if ie.String() != expected {
		t.Errorf("String() = %q, want %q", ie.String(), expected)
	}
}

func TestIndexExpressionStringDot(t *testing.T) {
	ie := &IndexExpression{
		Token: token.Token{Type: token.DOT, Literal: "."},
		Left:  &Identifier{Token: token.Token{Type: token.IDENT, Literal: "obj"}, Value: "obj"},
		Index: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "field"}, Value: "'field'"},
	}
	expected := "(obj.field)"
	if ie.String() != expected {
		t.Errorf("String() = %q, want %q", ie.String(), expected)
	}
}

// --- AssignmentExpression Tests ---

func TestAssignmentExpressionTokenLiteral(t *testing.T) {
	ae := &AssignmentExpression{
		Token: token.Token{Type: token.ASSIGN, Literal: "="},
		Left:  &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
		Value: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "42"}, Value: 42},
	}
	if ae.TokenLiteral() != "=" {
		t.Errorf("TokenLiteral() = %q, want %q", ae.TokenLiteral(), "=")
	}
}

func TestAssignmentExpressionString(t *testing.T) {
	ae := &AssignmentExpression{
		Token: token.Token{Type: token.ASSIGN, Literal: "="},
		Left:  &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
		Value: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "42"}, Value: 42},
	}
	expected := "x = 42"
	if ae.String() != expected {
		t.Errorf("String() = %q, want %q", ae.String(), expected)
	}
}

// --- EvalExpression Tests ---

func TestEvalExpressionTokenLiteral(t *testing.T) {
	ee := &EvalExpression{
		Token:     token.Token{Type: token.EVAL, Literal: "eval"},
		StrToEval: &StringLiteral{Token: token.Token{Type: token.STRING_DOUBLE_QUOTE, Literal: "\""}, Value: "x + 1"},
	}
	if ee.TokenLiteral() != "eval" {
		t.Errorf("TokenLiteral() = %q, want %q", ee.TokenLiteral(), "eval")
	}
}

func TestEvalExpressionString(t *testing.T) {
	ee := &EvalExpression{
		Token:     token.Token{Type: token.EVAL, Literal: "eval"},
		StrToEval: &StringLiteral{Token: token.Token{Type: token.STRING_DOUBLE_QUOTE, Literal: "\""}, Value: "x + 1"},
	}
	// StringLiteral.String() includes its own quotes, so we get nested quotes
	expected := `eval(""x + 1"")`
	if ee.String() != expected {
		t.Errorf("String() = %q, want %q", ee.String(), expected)
	}
}

// --- SpawnExpression Tests ---

func TestSpawnExpressionTokenLiteral(t *testing.T) {
	se := &SpawnExpression{
		Token: token.Token{Type: token.SPAWN, Literal: "spawn"},
		Arguments: []Expression{
			&Identifier{Token: token.Token{Type: token.IDENT, Literal: "foo"}, Value: "foo"},
		},
	}
	if se.TokenLiteral() != "spawn" {
		t.Errorf("TokenLiteral() = %q, want %q", se.TokenLiteral(), "spawn")
	}
}

func TestSpawnExpressionString(t *testing.T) {
	se := &SpawnExpression{
		Token:     token.Token{Type: token.SPAWN, Literal: "spawn"},
		Arguments: []Expression{},
	}
	expected := "spawn()"
	if se.String() != expected {
		t.Errorf("String() = %q, want %q", se.String(), expected)
	}
}

func TestSpawnExpressionStringWithArgs(t *testing.T) {
	se := &SpawnExpression{
		Token: token.Token{Type: token.SPAWN, Literal: "spawn"},
		Arguments: []Expression{
			&Identifier{Token: token.Token{Type: token.IDENT, Literal: "foo"}, Value: "foo"},
			&Identifier{Token: token.Token{Type: token.IDENT, Literal: "bar"}, Value: "bar"},
		},
	}
	expected := "spawn(foo, bar)"
	if se.String() != expected {
		t.Errorf("String() = %q, want %q", se.String(), expected)
	}
}

// --- DeferExpression Tests ---

func TestDeferExpressionTokenLiteral(t *testing.T) {
	de := &DeferExpression{
		Token: token.Token{Type: token.DEFER, Literal: "defer"},
		Arguments: []Expression{
			&Identifier{Token: token.Token{Type: token.IDENT, Literal: "foo"}, Value: "foo"},
		},
	}
	if de.TokenLiteral() != "defer" {
		t.Errorf("TokenLiteral() = %q, want %q", de.TokenLiteral(), "defer")
	}
}

func TestDeferExpressionString(t *testing.T) {
	de := &DeferExpression{
		Token:     token.Token{Type: token.DEFER, Literal: "defer"},
		Arguments: []Expression{},
	}
	expected := "defer()"
	if de.String() != expected {
		t.Errorf("String() = %q, want %q", de.String(), expected)
	}
}

// --- SelfExpression Tests ---

func TestSelfExpressionTokenLiteral(t *testing.T) {
	se := &SelfExpression{
		Token: token.Token{Type: token.SELF, Literal: "self"},
	}
	if se.TokenLiteral() != "self" {
		t.Errorf("TokenLiteral() = %q, want %q", se.TokenLiteral(), "self")
	}
}

func TestSelfExpressionString(t *testing.T) {
	se := &SelfExpression{
		Token: token.Token{Type: token.SELF, Literal: "self"},
	}
	if se.String() != "self()" {
		t.Errorf("String() = %q, want %q", se.String(), "self()")
	}
}

// --- IntegerLiteral Tests ---

func TestIntegerLiteralTokenLiteral(t *testing.T) {
	il := &IntegerLiteral{
		Token: token.Token{Type: token.INT, Literal: "42"},
		Value: 42,
	}
	if il.TokenLiteral() != "42" {
		t.Errorf("TokenLiteral() = %q, want %q", il.TokenLiteral(), "42")
	}
}

func TestIntegerLiteralString(t *testing.T) {
	il := &IntegerLiteral{
		Token: token.Token{Type: token.INT, Literal: "42"},
		Value: 42,
	}
	if il.String() != "42" {
		t.Errorf("String() = %q, want %q", il.String(), "42")
	}
}

// --- FloatLiteral Tests ---

func TestFloatLiteralTokenLiteral(t *testing.T) {
	fl := &FloatLiteral{
		Token: token.Token{Type: token.FLOAT, Literal: "3.14"},
		Value: 3.14,
	}
	if fl.TokenLiteral() != "3.14" {
		t.Errorf("TokenLiteral() = %q, want %q", fl.TokenLiteral(), "3.14")
	}
}

func TestFloatLiteralString(t *testing.T) {
	fl := &FloatLiteral{
		Token: token.Token{Type: token.FLOAT, Literal: "3.14"},
		Value: 3.14,
	}
	if fl.String() != "3.14" {
		t.Errorf("String() = %q, want %q", fl.String(), "3.14")
	}
}

// --- BigIntegerLiteral Tests ---

func TestBigIntegerLiteralTokenLiteral(t *testing.T) {
	val, _ := big.NewInt(0).SetString("123456789012345678901234567890", 10)
	bil := &BigIntegerLiteral{
		Token: token.Token{Type: token.BIGINT, Literal: "123456789012345678901234567890"},
		Value: val,
	}
	if bil.TokenLiteral() != "123456789012345678901234567890" {
		t.Errorf("TokenLiteral() = %q, want %q", bil.TokenLiteral(), "123456789012345678901234567890")
	}
}

func TestBigIntegerLiteralString(t *testing.T) {
	val := big.NewInt(999999999999)
	bil := &BigIntegerLiteral{
		Token: token.Token{Type: token.BIGINT, Literal: "999999999999"},
		Value: val,
	}
	if bil.String() != "999999999999" {
		t.Errorf("String() = %q, want %q", bil.String(), "999999999999")
	}
}

// --- BigFloatLiteral Tests ---

func TestBigFloatLiteralTokenLiteral(t *testing.T) {
	d, _ := decimal.NewFromString("123.456")
	bfl := &BigFloatLiteral{
		Token: token.Token{Type: token.BIGFLOAT, Literal: "123.456"},
		Value: d,
	}
	if bfl.TokenLiteral() != "123.456" {
		t.Errorf("TokenLiteral() = %q, want %q", bfl.TokenLiteral(), "123.456")
	}
}

func TestBigFloatLiteralString(t *testing.T) {
	d, _ := decimal.NewFromString("999.999")
	bfl := &BigFloatLiteral{
		Token: token.Token{Type: token.BIGFLOAT, Literal: "999.999"},
		Value: d,
	}
	if bfl.String() != "999.999" {
		t.Errorf("String() = %q, want %q", bfl.String(), "999.999")
	}
}

// --- HexLiteral Tests ---

func TestHexLiteralTokenLiteral(t *testing.T) {
	hl := &HexLiteral{
		Token: token.Token{Type: token.HEX, Literal: "0xFF"},
		Value: 0xFF,
	}
	if hl.TokenLiteral() != "0xFF" {
		t.Errorf("TokenLiteral() = %q, want %q", hl.TokenLiteral(), "0xFF")
	}
}

func TestHexLiteralString(t *testing.T) {
	hl := &HexLiteral{
		Token: token.Token{Type: token.HEX, Literal: "0xFF"},
		Value: 0xFF,
	}
	if hl.String() != "0xFF" {
		t.Errorf("String() = %q, want %q", hl.String(), "0xFF")
	}
}

// --- OctalLiteral Tests ---

func TestOctalLiteralTokenLiteral(t *testing.T) {
	ol := &OctalLiteral{
		Token: token.Token{Type: token.OCTAL, Literal: "0o77"},
		Value: 0o77,
	}
	if ol.TokenLiteral() != "0o77" {
		t.Errorf("TokenLiteral() = %q, want %q", ol.TokenLiteral(), "0o77")
	}
}

func TestOctalLiteralString(t *testing.T) {
	ol := &OctalLiteral{
		Token: token.Token{Type: token.OCTAL, Literal: "0o77"},
		Value: 0o77,
	}
	if ol.String() != "0o77" {
		t.Errorf("String() = %q, want %q", ol.String(), "0o77")
	}
}

// --- BinaryLiteral Tests ---

func TestBinaryLiteralTokenLiteral(t *testing.T) {
	bl := &BinaryLiteral{
		Token: token.Token{Type: token.BINARY, Literal: "0b1010"},
		Value: 0b1010,
	}
	if bl.TokenLiteral() != "0b1010" {
		t.Errorf("TokenLiteral() = %q, want %q", bl.TokenLiteral(), "0b1010")
	}
}

func TestBinaryLiteralString(t *testing.T) {
	bl := &BinaryLiteral{
		Token: token.Token{Type: token.BINARY, Literal: "0b1010"},
		Value: 0b1010,
	}
	if bl.String() != "0b1010" {
		t.Errorf("String() = %q, want %q", bl.String(), "0b1010")
	}
}

// --- UIntegerLiteral Tests ---

func TestUIntegerLiteralTokenLiteral(t *testing.T) {
	ul := &UIntegerLiteral{
		Token: token.Token{Type: token.UINT, Literal: "42"},
		Value: 42,
	}
	if ul.TokenLiteral() != "42" {
		t.Errorf("TokenLiteral() = %q, want %q", ul.TokenLiteral(), "42")
	}
}

func TestUIntegerLiteralString(t *testing.T) {
	ul := &UIntegerLiteral{
		Token: token.Token{Type: token.UINT, Literal: "42"},
		Value: 42,
	}
	if ul.String() != "42" {
		t.Errorf("String() = %q, want %q", ul.String(), "42")
	}
}

// --- FunctionLiteral Tests ---

func TestFunctionLiteralTokenLiteral(t *testing.T) {
	fl := &FunctionLiteral{
		Token: token.Token{Type: token.FUNCTION, Literal: "fun"},
		Body: &BlockStatement{
			Token: token.Token{Type: token.LBRACE, Literal: "{"},
		},
	}
	if fl.TokenLiteral() != "fun" {
		t.Errorf("TokenLiteral() = %q, want %q", fl.TokenLiteral(), "fun")
	}
}

func TestFunctionLiteralStringNoParams(t *testing.T) {
	fl := &FunctionLiteral{
		Token: token.Token{Type: token.FUNCTION, Literal: "fun"},
		Body: &BlockStatement{
			Token: token.Token{Type: token.LBRACE, Literal: "{"},
		},
	}
	result := fl.String()
	if result == "" {
		t.Errorf("String() should not be empty")
	}
}

func TestFunctionLiteralStringWithParams(t *testing.T) {
	fl := &FunctionLiteral{
		Token: token.Token{Type: token.FUNCTION, Literal: "fun"},
		Parameters: []*Identifier{
			{Token: token.Token{Type: token.IDENT, Literal: "a"}, Value: "a"},
			{Token: token.Token{Type: token.IDENT, Literal: "b"}, Value: "b"},
		},
		Body: &BlockStatement{
			Token: token.Token{Type: token.LBRACE, Literal: "{"},
		},
	}
	result := fl.String()
	if result == "" {
		t.Errorf("String() should not be empty")
	}
}

// --- ExecStringLiteral Tests ---

func TestExecStringLiteralTokenLiteral(t *testing.T) {
	esl := &ExecStringLiteral{
		Token: token.Token{Type: token.BACKTICK, Literal: "`"},
		Value: "ls -la",
	}
	if esl.TokenLiteral() != "`" {
		t.Errorf("TokenLiteral() = %q, want %q", esl.TokenLiteral(), "`")
	}
}

func TestExecStringLiteralString(t *testing.T) {
	esl := &ExecStringLiteral{
		Token: token.Token{Type: token.BACKTICK, Literal: "`"},
		Value: "ls -la",
	}
	expected := "`ls -la`"
	if esl.String() != expected {
		t.Errorf("String() = %q, want %q", esl.String(), expected)
	}
}

// --- StringLiteral Tests ---

func TestStringLiteralDoubleQuote(t *testing.T) {
	sl := &StringLiteral{
		Token: token.Token{Type: token.STRING_DOUBLE_QUOTE, Literal: "\""},
		Value: "hello",
	}
	expected := `"hello"`
	if sl.String() != expected {
		t.Errorf("String() = %q, want %q", sl.String(), expected)
	}
}

func TestStringLiteralSingleQuote(t *testing.T) {
	sl := &StringLiteral{
		Token: token.Token{Type: token.STRING_SINGLE_QUOTE, Literal: "'"},
		Value: "hello",
	}
	expected := `'hello'`
	if sl.String() != expected {
		t.Errorf("String() = %q, want %q", sl.String(), expected)
	}
}

func TestStringLiteralStringWithoutQuotesDouble(t *testing.T) {
	sl := &StringLiteral{
		Token: token.Token{Type: token.STRING_DOUBLE_QUOTE, Literal: "\""},
		Value: "hello",
	}
	if sl.StringWithoutQuotes() != "hello" {
		t.Errorf("StringWithoutQuotes() = %q, want %q", sl.StringWithoutQuotes(), "hello")
	}
}

func TestStringLiteralStringWithoutQuotesSingle(t *testing.T) {
	sl := &StringLiteral{
		Token: token.Token{Type: token.STRING_SINGLE_QUOTE, Literal: "'"},
		Value: "hello",
	}
	if sl.StringWithoutQuotes() != "hello" {
		t.Errorf("StringWithoutQuotes() = %q, want %q", sl.StringWithoutQuotes(), "hello")
	}
}

// --- RegexLiteral Tests ---

func TestRegexLiteralTokenLiteral(t *testing.T) {
	rl := &RegexLiteral{
		Token: token.Token{Type: token.REGEX, Literal: "r/"},
		Value: "^[a-z]+$",
	}
	if rl.TokenLiteral() != "r/" {
		t.Errorf("TokenLiteral() = %q, want %q", rl.TokenLiteral(), "r/")
	}
}

func TestRegexLiteralString(t *testing.T) {
	rl := &RegexLiteral{
		Token: token.Token{Type: token.REGEX, Literal: "r/"},
		Value: "^[a-z]+$",
	}
	expected := `r/^[a-z]+$/`
	if rl.String() != expected {
		t.Errorf("String() = %q, want %q", rl.String(), expected)
	}
}

func TestRegexLiteralStringWithBackslash(t *testing.T) {
	rl := &RegexLiteral{
		Token: token.Token{Type: token.REGEX, Literal: "r/"},
		Value: "\\d+",
	}
	expected := `r/\\d+/`
	if rl.String() != expected {
		t.Errorf("String() = %q, want %q", rl.String(), expected)
	}
}

// --- ListLiteral Tests ---

func TestListLiteralTokenLiteral(t *testing.T) {
	ll := &ListLiteral{
		Token:    token.Token{Type: token.LBRACKET, Literal: "["},
		Elements: []Expression{},
	}
	if ll.TokenLiteral() != "[" {
		t.Errorf("TokenLiteral() = %q, want %q", ll.TokenLiteral(), "[")
	}
}

func TestListLiteralStringEmpty(t *testing.T) {
	ll := &ListLiteral{
		Token:    token.Token{Type: token.LBRACKET, Literal: "["},
		Elements: []Expression{},
	}
	expected := "[]"
	if ll.String() != expected {
		t.Errorf("String() = %q, want %q", ll.String(), expected)
	}
}

func TestListLiteralStringWithElements(t *testing.T) {
	ll := &ListLiteral{
		Token: token.Token{Type: token.LBRACKET, Literal: "["},
		Elements: []Expression{
			&IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "1"}, Value: 1},
			&IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "2"}, Value: 2},
			&IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "3"}, Value: 3},
		},
	}
	expected := "[1, 2, 3]"
	if ll.String() != expected {
		t.Errorf("String() = %q, want %q", ll.String(), expected)
	}
}

// --- ListCompLiteral Tests ---

func TestListCompLiteralString(t *testing.T) {
	lcl := &ListCompLiteral{
		Token:               token.Token{},
		NonEvaluatedProgram: "[x for x in range(10)]",
	}
	expected := "[x for x in range(10)]"
	if lcl.String() != expected {
		t.Errorf("String() = %q, want %q", lcl.String(), expected)
	}
}

func TestListCompLiteralTokenLiteral(t *testing.T) {
	lcl := &ListCompLiteral{
		Token:               token.Token{},
		NonEvaluatedProgram: "test",
	}
	if lcl.TokenLiteral() != "" {
		t.Errorf("TokenLiteral() = %q, want empty", lcl.TokenLiteral())
	}
}

func TestListCompLiteralTokenToken(t *testing.T) {
	lcl := &ListCompLiteral{}
	tt := lcl.TokenToken()
	if tt.Type != "" {
		t.Errorf("TokenToken() Type should be empty, got %q", tt.Type)
	}
}

// --- MapLiteral Tests ---

func TestMapLiteralTokenLiteral(t *testing.T) {
	ml := &MapLiteral{
		Token:      token.Token{Type: token.LBRACE, Literal: "{"},
		Pairs:      make(map[Expression]Expression),
		PairsIndex: make(map[int]Expression),
	}
	if ml.TokenLiteral() != "{" {
		t.Errorf("TokenLiteral() = %q, want %q", ml.TokenLiteral(), "{")
	}
}

func TestMapLiteralStringEmpty(t *testing.T) {
	ml := &MapLiteral{
		Token:      token.Token{Type: token.LBRACE, Literal: "{"},
		Pairs:      make(map[Expression]Expression),
		PairsIndex: make(map[int]Expression),
	}
	expected := "{}"
	if ml.String() != expected {
		t.Errorf("String() = %q, want %q", ml.String(), expected)
	}
}

func TestMapLiteralStringWithPairs(t *testing.T) {
	key1 := &StringLiteral{Token: token.Token{Type: token.STRING_DOUBLE_QUOTE, Literal: "\""}, Value: "name"}
	val1 := &StringLiteral{Token: token.Token{Type: token.STRING_DOUBLE_QUOTE, Literal: "\""}, Value: "alice"}
	key2 := &StringLiteral{Token: token.Token{Type: token.STRING_DOUBLE_QUOTE, Literal: "\""}, Value: "age"}
	val2 := &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "30"}, Value: 30}

	ml := &MapLiteral{
		Token:      token.Token{Type: token.LBRACE, Literal: "{"},
		Pairs:      map[Expression]Expression{key1: val1, key2: val2},
		PairsIndex: map[int]Expression{0: key1, 1: key2},
	}
	result := ml.String()
	if result == "" {
		t.Errorf("String() should not be empty")
	}
}

// --- MapCompLiteral Tests ---

func TestMapCompLiteralString(t *testing.T) {
	mcl := &MapCompLiteral{
		NonEvaluatedProgram: "{x: x for x in range(10)}",
	}
	expected := "{x: x for x in range(10)}"
	if mcl.String() != expected {
		t.Errorf("String() = %q, want %q", mcl.String(), expected)
	}
}

func TestMapCompLiteralTokenToken(t *testing.T) {
	mcl := &MapCompLiteral{}
	tt := mcl.TokenToken()
	if tt.Type != "" {
		t.Errorf("TokenToken() Type should be empty, got %q", tt.Type)
	}
}

// --- SetLiteral Tests ---

func TestSetLiteralTokenLiteral(t *testing.T) {
	sl := &SetLiteral{
		Token:    token.Token{Type: token.LBRACE, Literal: "{"},
		Elements: []Expression{},
	}
	if sl.TokenLiteral() != "{" {
		t.Errorf("TokenLiteral() = %q, want %q", sl.TokenLiteral(), "{")
	}
}

func TestSetLiteralStringEmpty(t *testing.T) {
	sl := &SetLiteral{
		Token:    token.Token{Type: token.LBRACE, Literal: "{"},
		Elements: []Expression{},
	}
	expected := "{}"
	if sl.String() != expected {
		t.Errorf("String() = %q, want %q", sl.String(), expected)
	}
}

func TestSetLiteralStringWithElements(t *testing.T) {
	sl := &SetLiteral{
		Token: token.Token{Type: token.LBRACE, Literal: "{"},
		Elements: []Expression{
			&IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "1"}, Value: 1},
			&IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "2"}, Value: 2},
		},
	}
	expected := "{1, 2}"
	if sl.String() != expected {
		t.Errorf("String() = %q, want %q", sl.String(), expected)
	}
}

// --- SetCompLiteral Tests ---

func TestSetCompLiteralString(t *testing.T) {
	scl := &SetCompLiteral{
		NonEvaluatedProgram: "{x for x in range(10)}",
	}
	expected := "{x for x in range(10)}"
	if scl.String() != expected {
		t.Errorf("String() = %q, want %q", scl.String(), expected)
	}
}

func TestSetCompLiteralTokenToken(t *testing.T) {
	scl := &SetCompLiteral{}
	tt := scl.TokenToken()
	if tt.Type != "" {
		t.Errorf("TokenToken() Type should be empty, got %q", tt.Type)
	}
}

// --- StructLiteral Tests ---

func TestStructLiteralTokenLiteral(t *testing.T) {
	sl := &StructLiteral{
		Token:  token.Token{Type: token.ATLBRACE, Literal: "@{"},
		Fields: []string{"name", "age"},
		Values: []Expression{
			&StringLiteral{Token: token.Token{Type: token.STRING_DOUBLE_QUOTE, Literal: "\""}, Value: "alice"},
			&IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "30"}, Value: 30},
		},
	}
	if sl.TokenLiteral() != "@{" {
		t.Errorf("TokenLiteral() = %q, want %q", sl.TokenLiteral(), "@{")
	}
}

func TestStructLiteralString(t *testing.T) {
	sl := &StructLiteral{
		Token:  token.Token{Type: token.ATLBRACE, Literal: "@{"},
		Fields: []string{"name", "age"},
		Values: []Expression{
			&StringLiteral{Token: token.Token{Type: token.STRING_DOUBLE_QUOTE, Literal: "\""}, Value: "alice"},
			&IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "30"}, Value: 30},
		},
	}
	result := sl.String()
	if result == "" {
		t.Errorf("String() should not be empty")
	}
}

func TestStructLiteralStringEmpty(t *testing.T) {
	sl := &StructLiteral{
		Token:  token.Token{Type: token.ATLBRACE, Literal: "@{"},
		Fields: []string{},
		Values: []Expression{},
	}
	expected := "@{}"
	if sl.String() != expected {
		t.Errorf("String() = %q, want %q", sl.String(), expected)
	}
}

// --- VarStatement Tests ---

func TestVarStatementTokenLiteral(t *testing.T) {
	vs := &VarStatement{
		Token: token.Token{Type: token.VAR, Literal: "var"},
		Names: []*Identifier{
			{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
		},
		Value: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "1"}, Value: 1},
		AssignmentToken: token.Token{Type: token.ASSIGN, Literal: "="},
	}
	if vs.TokenLiteral() != "var" {
		t.Errorf("TokenLiteral() = %q, want %q", vs.TokenLiteral(), "var")
	}
}

func TestVarStatementStringSimple(t *testing.T) {
	vs := &VarStatement{
		Token: token.Token{Type: token.VAR, Literal: "var"},
		Names: []*Identifier{
			{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
		},
		Value: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "1"}, Value: 1},
		AssignmentToken: token.Token{Type: token.ASSIGN, Literal: "="},
	}
	expected := "var x = 1"
	if vs.String() != expected {
		t.Errorf("String() = %q, want %q", vs.String(), expected)
	}
}

func TestVarStatementStringMultipleNames(t *testing.T) {
	vs := &VarStatement{
		Token: token.Token{Type: token.VAR, Literal: "var"},
		Names: []*Identifier{
			{Token: token.Token{Type: token.IDENT, Literal: "a"}, Value: "a"},
			{Token: token.Token{Type: token.IDENT, Literal: "b"}, Value: "b"},
		},
		Value: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "1"}, Value: 1},
		AssignmentToken: token.Token{Type: token.ASSIGN, Literal: "="},
	}
	expected := "var a, b = 1"
	if vs.String() != expected {
		t.Errorf("String() = %q, want %q", vs.String(), expected)
	}
}

func TestVarStatementStringListDestructor(t *testing.T) {
	vs := &VarStatement{
		Token: token.Token{Type: token.VAR, Literal: "var"},
		Names: []*Identifier{
			{Token: token.Token{Type: token.IDENT, Literal: "a"}, Value: "a"},
			{Token: token.Token{Type: token.IDENT, Literal: "b"}, Value: "b"},
		},
		Value: &ListLiteral{
			Token:    token.Token{Type: token.LBRACKET, Literal: "["},
			Elements: []Expression{},
		},
		IsListDestructor: true,
		AssignmentToken:  token.Token{Type: token.ASSIGN, Literal: "="},
	}
	expected := "var [a, b] = []"
	if vs.String() != expected {
		t.Errorf("String() = %q, want %q", vs.String(), expected)
	}
}

func TestVarStatementStringMapDestructor(t *testing.T) {
	key := &StringLiteral{Token: token.Token{Type: token.STRING_DOUBLE_QUOTE, Literal: "\""}, Value: "name"}
	vs := &VarStatement{
		Token: token.Token{Type: token.VAR, Literal: "var"},
		Names: []*Identifier{
			{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
		},
		Value: &MapLiteral{
			Token:      token.Token{Type: token.LBRACE, Literal: "{"},
			Pairs:      map[Expression]Expression{key: &StringLiteral{Token: token.Token{Type: token.STRING_DOUBLE_QUOTE, Literal: "\""}, Value: "alice"}},
			PairsIndex: map[int]Expression{0: key},
		},
		IsMapDestructor: true,
		AssignmentToken: token.Token{Type: token.ASSIGN, Literal: "="},
	}
	result := vs.String()
	if result == "" {
		t.Errorf("String() should not be empty")
	}
}

func TestVarStatementIsValStatement(t *testing.T) {
	vs := &VarStatement{}
	if vs.IsValStatement() {
		t.Errorf("IsValStatement() should be false for VarStatement")
	}
}

func TestVarStatementVVToken(t *testing.T) {
	expected := token.Token{Type: token.VAR, Literal: "var"}
	vs := &VarStatement{Token: expected}
	if vs.VVToken().Type != expected.Type {
		t.Errorf("VVToken() = %v, want %v", vs.VVToken(), expected)
	}
}

func TestVarStatementVVNames(t *testing.T) {
	names := []*Identifier{
		{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
	}
	vs := &VarStatement{Names: names}
	got := vs.VVNames()
	if len(got) != 1 || got[0].Value != "x" {
		t.Errorf("VVNames() = %v, want [%v]", got, "x")
	}
}

func TestVarStatementVVValue(t *testing.T) {
	val := &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "1"}, Value: 1}
	vs := &VarStatement{Value: val}
	if vs.VVValue() != val {
		t.Errorf("VVValue() mismatch")
	}
}

func TestVarStatementVVIsMapDestructor(t *testing.T) {
	vs := &VarStatement{IsMapDestructor: true}
	if !vs.VVIsMapDestructor() {
		t.Errorf("VVIsMapDestructor() should be true")
	}
}

func TestVarStatementVVIsListDestructor(t *testing.T) {
	vs := &VarStatement{IsListDestructor: true}
	if !vs.VVIsListDestructor() {
		t.Errorf("VVIsListDestructor() should be true")
	}
}

func TestVarStatementWithPlusEq(t *testing.T) {
	vs := &VarStatement{
		Token: token.Token{Type: token.VAR, Literal: "var"},
		Names: []*Identifier{
			{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
		},
		Value: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "1"}, Value: 1},
		AssignmentToken: token.Token{Type: token.PLUSEQ, Literal: "+="},
	}
	expected := "var x += 1"
	if vs.String() != expected {
		t.Errorf("String() = %q, want %q", vs.String(), expected)
	}
}

// --- ValStatement Tests ---

func TestValStatementTokenLiteral(t *testing.T) {
	vs := &ValStatement{
		Token: token.Token{Type: token.VAL, Literal: "val"},
		Names: []*Identifier{
			{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
		},
		Value: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "1"}, Value: 1},
	}
	if vs.TokenLiteral() != "val" {
		t.Errorf("TokenLiteral() = %q, want %q", vs.TokenLiteral(), "val")
	}
}

func TestValStatementStringSimple(t *testing.T) {
	vs := &ValStatement{
		Token: token.Token{Type: token.VAL, Literal: "val"},
		Names: []*Identifier{
			{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
		},
		Value: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "1"}, Value: 1},
	}
	expected := "val x = 1"
	if vs.String() != expected {
		t.Errorf("String() = %q, want %q", vs.String(), expected)
	}
}

func TestValStatementStringMultipleNames(t *testing.T) {
	vs := &ValStatement{
		Token: token.Token{Type: token.VAL, Literal: "val"},
		Names: []*Identifier{
			{Token: token.Token{Type: token.IDENT, Literal: "a"}, Value: "a"},
			{Token: token.Token{Type: token.IDENT, Literal: "b"}, Value: "b"},
		},
		Value: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "1"}, Value: 1},
	}
	expected := "val a, b = 1"
	if vs.String() != expected {
		t.Errorf("String() = %q, want %q", vs.String(), expected)
	}
}

func TestValStatementStringListDestructor(t *testing.T) {
	vs := &ValStatement{
		Token: token.Token{Type: token.VAL, Literal: "val"},
		Names: []*Identifier{
			{Token: token.Token{Type: token.IDENT, Literal: "a"}, Value: "a"},
			{Token: token.Token{Type: token.IDENT, Literal: "b"}, Value: "b"},
		},
		Value: &ListLiteral{
			Token:    token.Token{Type: token.LBRACKET, Literal: "["},
			Elements: []Expression{},
		},
		IsListDestructor: true,
	}
	expected := "val [a, b] = []"
	if vs.String() != expected {
		t.Errorf("String() = %q, want %q", vs.String(), expected)
	}
}

func TestValStatementIsValStatement(t *testing.T) {
	vs := &ValStatement{}
	if !vs.IsValStatement() {
		t.Errorf("IsValStatement() should be true for ValStatement")
	}
}

func TestValStatementVVToken(t *testing.T) {
	expected := token.Token{Type: token.VAL, Literal: "val"}
	vs := &ValStatement{Token: expected}
	if vs.VVToken().Type != expected.Type {
		t.Errorf("VVToken() = %v, want %v", vs.VVToken(), expected)
	}
}

func TestValStatementVVNames(t *testing.T) {
	names := []*Identifier{
		{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
	}
	vs := &ValStatement{Names: names}
	got := vs.VVNames()
	if len(got) != 1 || got[0].Value != "x" {
		t.Errorf("VVNames() = %v, want [%v]", got, "x")
	}
}

func TestValStatementVVValue(t *testing.T) {
	val := &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "1"}, Value: 1}
	vs := &ValStatement{Value: val}
	if vs.VVValue() != val {
		t.Errorf("VVValue() mismatch")
	}
}

func TestValStatementVVIsMapDestructor(t *testing.T) {
	vs := &ValStatement{IsMapDestructor: true}
	if !vs.VVIsMapDestructor() {
		t.Errorf("VVIsMapDestructor() should be true")
	}
}

func TestValStatementVVIsListDestructor(t *testing.T) {
	vs := &ValStatement{IsListDestructor: true}
	if !vs.VVIsListDestructor() {
		t.Errorf("VVIsListDestructor() should be true")
	}
}

// --- FunctionStatement Tests ---

func TestFunctionStatementTokenLiteral(t *testing.T) {
	fs := &FunctionStatement{
		Token:  token.Token{Type: token.FUNCTION, Literal: "fun"},
		Name:   &Identifier{Token: token.Token{Type: token.IDENT, Literal: "foo"}, Value: "foo"},
		Body:   &BlockStatement{Token: token.Token{Type: token.LBRACE, Literal: "{"}, Statements: []Statement{}},
	}
	if fs.TokenLiteral() != "fun" {
		t.Errorf("TokenLiteral() = %q, want %q", fs.TokenLiteral(), "fun")
	}
}

func TestFunctionStatementString(t *testing.T) {
	fs := &FunctionStatement{
		Token:  token.Token{Type: token.FUNCTION, Literal: "fun"},
		Name:   &Identifier{Token: token.Token{Type: token.IDENT, Literal: "foo"}, Value: "foo"},
		Body:   &BlockStatement{Token: token.Token{Type: token.LBRACE, Literal: "{"}, Statements: []Statement{}},
	}
	result := fs.String()
	if result == "" {
		t.Errorf("String() should not be empty")
	}
}

func TestFunctionStatementStringWithParams(t *testing.T) {
	fs := &FunctionStatement{
		Token:      token.Token{Type: token.FUNCTION, Literal: "fun"},
		Name:       &Identifier{Token: token.Token{Type: token.IDENT, Literal: "foo"}, Value: "foo"},
		Parameters: []*Identifier{{Token: token.Token{Type: token.IDENT, Literal: "a"}, Value: "a"}},
		Body:       &BlockStatement{Token: token.Token{Type: token.LBRACE, Literal: "{"}, Statements: []Statement{}},
	}
	result := fs.String()
	if result == "" {
		t.Errorf("String() should not be empty")
	}
}

// --- ReturnStatement Tests ---

func TestReturnStatementTokenLiteral(t *testing.T) {
	rs := &ReturnStatement{
		Token: token.Token{Type: token.RETURN, Literal: "return"},
	}
	if rs.TokenLiteral() != "return" {
		t.Errorf("TokenLiteral() = %q, want %q", rs.TokenLiteral(), "return")
	}
}

func TestReturnStatementStringNoValue(t *testing.T) {
	rs := &ReturnStatement{
		Token: token.Token{Type: token.RETURN, Literal: "return"},
	}
	expected := "return "
	if rs.String() != expected {
		t.Errorf("String() = %q, want %q", rs.String(), expected)
	}
}

func TestReturnStatementStringWithValue(t *testing.T) {
	rs := &ReturnStatement{
		Token:       token.Token{Type: token.RETURN, Literal: "return"},
		ReturnValue: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "42"}, Value: 42},
	}
	expected := "return 42"
	if rs.String() != expected {
		t.Errorf("String() = %q, want %q", rs.String(), expected)
	}
}

// --- TryCatchStatement Tests ---

func TestTryCatchStatementTokenLiteral(t *testing.T) {
	tcs := &TryCatchStatement{
		Token:           token.Token{Type: token.TRY, Literal: "try"},
		TryBlock:        &BlockStatement{Token: token.Token{Type: token.LBRACE, Literal: "{"}, Statements: []Statement{}},
		CatchIdentifier: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "e"}, Value: "e"},
		CatchBlock:      &BlockStatement{Token: token.Token{Type: token.LBRACE, Literal: "{"}, Statements: []Statement{}},
	}
	if tcs.TokenLiteral() != "try" {
		t.Errorf("TokenLiteral() = %q, want %q", tcs.TokenLiteral(), "try")
	}
}

func TestTryCatchStatementStringNoFinally(t *testing.T) {
	tcs := &TryCatchStatement{
		Token:           token.Token{Type: token.TRY, Literal: "try"},
		TryBlock:        &BlockStatement{Token: token.Token{Type: token.LBRACE, Literal: "{"}, Statements: []Statement{}},
		CatchIdentifier: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "e"}, Value: "e"},
		CatchBlock:      &BlockStatement{Token: token.Token{Type: token.LBRACE, Literal: "{"}, Statements: []Statement{}},
	}
	result := tcs.String()
	if result == "" {
		t.Errorf("String() should not be empty")
	}
}

func TestTryCatchStatementStringWithFinally(t *testing.T) {
	tcs := &TryCatchStatement{
		Token:           token.Token{Type: token.TRY, Literal: "try"},
		TryBlock:        &BlockStatement{Token: token.Token{Type: token.LBRACE, Literal: "{"}, Statements: []Statement{}},
		CatchIdentifier: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "e"}, Value: "e"},
		CatchBlock:      &BlockStatement{Token: token.Token{Type: token.LBRACE, Literal: "{"}, Statements: []Statement{}},
		FinallyBlock:    &BlockStatement{Token: token.Token{Type: token.LBRACE, Literal: "{"}, Statements: []Statement{}},
	}
	result := tcs.String()
	if result == "" {
		t.Errorf("String() should not be empty")
	}
}

// --- ExpressionStatement Tests ---

func TestExpressionStatementTokenLiteral(t *testing.T) {
	es := &ExpressionStatement{
		Token:      token.Token{Type: token.IDENT, Literal: "x"},
		Expression: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
	}
	if es.TokenLiteral() != "x" {
		t.Errorf("TokenLiteral() = %q, want %q", es.TokenLiteral(), "x")
	}
}

func TestExpressionStatementString(t *testing.T) {
	es := &ExpressionStatement{
		Token:      token.Token{Type: token.IDENT, Literal: "x"},
		Expression: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
	}
	if es.String() != "x" {
		t.Errorf("String() = %q, want %q", es.String(), "x")
	}
}

func TestExpressionStatementStringEmpty(t *testing.T) {
	es := &ExpressionStatement{
		Token:      token.Token{Type: token.SEMICOLON, Literal: ";"},
		Expression: nil,
	}
	if es.String() != "" {
		t.Errorf("String() = %q, want empty", es.String())
	}
}

// --- BlockStatement Tests ---

func TestBlockStatementTokenLiteral(t *testing.T) {
	bs := &BlockStatement{
		Token: token.Token{Type: token.LBRACE, Literal: "{"},
	}
	if bs.TokenLiteral() != "{" {
		t.Errorf("TokenLiteral() = %q, want %q", bs.TokenLiteral(), "{")
	}
}

func TestBlockStatementStringEmpty(t *testing.T) {
	bs := &BlockStatement{
		Token:      token.Token{Type: token.LBRACE, Literal: "{"},
		Statements: []Statement{},
	}
	if bs.String() != "" {
		t.Errorf("String() = %q, want empty", bs.String())
	}
}

func TestBlockStatementStringMultiple(t *testing.T) {
	bs := &BlockStatement{
		Token: token.Token{Type: token.LBRACE, Literal: "{"},
		Statements: []Statement{
			&ExpressionStatement{
				Token:      token.Token{Type: token.IDENT, Literal: "a"},
				Expression: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "a"}, Value: "a"},
			},
			&ExpressionStatement{
				Token:      token.Token{Type: token.IDENT, Literal: "b"},
				Expression: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "b"}, Value: "b"},
			},
		},
	}
	expected := "a\nb"
	if bs.String() != expected {
		t.Errorf("String() = %q, want %q", bs.String(), expected)
	}
}

func TestBlockStatementExpressionString(t *testing.T) {
	bs := &BlockStatement{
		Token: token.Token{Type: token.LBRACE, Literal: "{"},
		Statements: []Statement{
			&ExpressionStatement{
				Token:      token.Token{Type: token.IDENT, Literal: "a"},
				Expression: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "a"}, Value: "a"},
			},
			&ExpressionStatement{
				Token:      token.Token{Type: token.IDENT, Literal: "b"},
				Expression: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "b"}, Value: "b"},
			},
		},
	}
	expected := "a;, b;"
	if bs.ExpressionString() != expected {
		t.Errorf("ExpressionString() = %q, want %q", bs.ExpressionString(), expected)
	}
}

func TestBlockStatementExpressionStringSingle(t *testing.T) {
	bs := &BlockStatement{
		Token: token.Token{Type: token.LBRACE, Literal: "{"},
		Statements: []Statement{
			&ExpressionStatement{
				Token:      token.Token{Type: token.IDENT, Literal: "a"},
				Expression: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "a"}, Value: "a"},
			},
		},
	}
	expected := "a;"
	if bs.ExpressionString() != expected {
		t.Errorf("ExpressionString() = %q, want %q", bs.ExpressionString(), expected)
	}
}

// --- ImportStatement Tests ---

func TestImportStatementTokenLiteral(t *testing.T) {
	is := &ImportStatement{
		Token: token.Token{Type: token.IMPORT, Literal: "import"},
		Path:  &Identifier{Token: token.Token{Type: token.IDENT, Literal: "foo"}, Value: "foo"},
	}
	if is.TokenLiteral() != "import" {
		t.Errorf("TokenLiteral() = %q, want %q", is.TokenLiteral(), "import")
	}
}

func TestImportStatementStringSimple(t *testing.T) {
	is := &ImportStatement{
		Token: token.Token{Type: token.IMPORT, Literal: "import"},
		Path:  &Identifier{Token: token.Token{Type: token.IDENT, Literal: "foo"}, Value: "foo"},
	}
	expected := "import foo"
	if is.String() != expected {
		t.Errorf("String() = %q, want %q", is.String(), expected)
	}
}

func TestImportStatementStringWithAlias(t *testing.T) {
	is := &ImportStatement{
		Token: token.Token{Type: token.IMPORT, Literal: "import"},
		Path:  &Identifier{Token: token.Token{Type: token.IDENT, Literal: "foo"}, Value: "foo"},
		Alias: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "bar"}, Value: "bar"},
	}
	expected := "import foo as bar"
	if is.String() != expected {
		t.Errorf("String() = %q, want %q", is.String(), expected)
	}
}

func TestImportStatementStringFromImport(t *testing.T) {
	is := &ImportStatement{
		Token:          token.Token{Type: token.IMPORT, Literal: "import"},
		Path:           &Identifier{Token: token.Token{Type: token.IDENT, Literal: "foo"}, Value: "foo"},
		IdentsToImport: []*Identifier{{Token: token.Token{Type: token.IDENT, Literal: "a"}, Value: "a"}},
	}
	expected := "from foo import [a]"
	if is.String() != expected {
		t.Errorf("String() = %q, want %q", is.String(), expected)
	}
}

func TestImportStatementStringFromImportStar(t *testing.T) {
	is := &ImportStatement{
		Token:          token.Token{Type: token.IMPORT, Literal: "import"},
		Path:           &Identifier{Token: token.Token{Type: token.IDENT, Literal: "foo"}, Value: "foo"},
		IdentsToImport: []*Identifier{},
		ImportAll:      true,
	}
	expected := "from foo import *"
	if is.String() != expected {
		t.Errorf("String() = %q, want %q", is.String(), expected)
	}
}

func TestImportStatementStringMultipleIdents(t *testing.T) {
	is := &ImportStatement{
		Token: token.Token{Type: token.IMPORT, Literal: "import"},
		Path:  &Identifier{Token: token.Token{Type: token.IDENT, Literal: "foo"}, Value: "foo"},
		IdentsToImport: []*Identifier{
			{Token: token.Token{Type: token.IDENT, Literal: "a"}, Value: "a"},
			{Token: token.Token{Type: token.IDENT, Literal: "b"}, Value: "b"},
		},
	}
	expected := "from foo import [a, b]"
	if is.String() != expected {
		t.Errorf("String() = %q, want %q", is.String(), expected)
	}
}

// --- BreakStatement Tests ---

func TestBreakStatementTokenLiteral(t *testing.T) {
	bs := &BreakStatement{
		Token: token.Token{Type: token.BREAK, Literal: "break"},
	}
	if bs.TokenLiteral() != "break" {
		t.Errorf("TokenLiteral() = %q, want %q", bs.TokenLiteral(), "break")
	}
}

func TestBreakStatementString(t *testing.T) {
	bs := &BreakStatement{
		Token: token.Token{Type: token.BREAK, Literal: "break"},
	}
	if bs.String() != "break" {
		t.Errorf("String() = %q, want %q", bs.String(), "break")
	}
}

// --- ContinueStatement Tests ---

func TestContinueStatementTokenLiteral(t *testing.T) {
	cs := &ContinueStatement{
		Token: token.Token{Type: token.CONTINUE, Literal: "continue"},
	}
	if cs.TokenLiteral() != "continue" {
		t.Errorf("TokenLiteral() = %q, want %q", cs.TokenLiteral(), "continue")
	}
}

func TestContinueStatementString(t *testing.T) {
	cs := &ContinueStatement{
		Token: token.Token{Type: token.CONTINUE, Literal: "continue"},
	}
	if cs.String() != "continue" {
		t.Errorf("String() = %q, want %q", cs.String(), "continue")
	}
}

// --- ForStatement Tests ---

func TestForStatementTokenLiteral(t *testing.T) {
	fs := &ForStatement{
		Token: token.Token{Type: token.FOR, Literal: "for"},
		Condition: &Boolean{Token: token.Token{Type: token.TRUE, Literal: "true"}, Value: true},
		Consequence: &BlockStatement{
			Token:      token.Token{Type: token.LBRACE, Literal: "{"},
			Statements: []Statement{},
		},
	}
	if fs.TokenLiteral() != "for" {
		t.Errorf("TokenLiteral() = %q, want %q", fs.TokenLiteral(), "for")
	}
}

func TestForStatementStringSimple(t *testing.T) {
	fs := &ForStatement{
		Token: token.Token{Type: token.FOR, Literal: "for"},
		Condition: &Boolean{Token: token.Token{Type: token.TRUE, Literal: "true"}, Value: true},
		Consequence: &BlockStatement{
			Token:      token.Token{Type: token.LBRACE, Literal: "{"},
			Statements: []Statement{},
		},
	}
	result := fs.String()
	if result == "" {
		t.Errorf("String() should not be empty")
	}
}

func TestForStatementStringWithVar(t *testing.T) {
	fs := &ForStatement{
		Token: token.Token{Type: token.FOR, Literal: "for"},
		Initializer: &VarStatement{
			Token: token.Token{Type: token.VAR, Literal: "var"},
			Names: []*Identifier{{Token: token.Token{Type: token.IDENT, Literal: "i"}, Value: "i"}},
			Value: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "0"}, Value: 0},
			AssignmentToken: token.Token{Type: token.ASSIGN, Literal: "="},
		},
		Condition: &InfixExpression{
			Token:    token.Token{Type: token.LT, Literal: "<"},
			Operator: "<",
			Left:     &Identifier{Token: token.Token{Type: token.IDENT, Literal: "i"}, Value: "i"},
			Right:    &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "10"}, Value: 10},
		},
		PostExp: &InfixExpression{
			Token:    token.Token{Type: token.PLUS, Literal: "+"},
			Operator: "+=",
			Left:     &Identifier{Token: token.Token{Type: token.IDENT, Literal: "i"}, Value: "i"},
			Right:    &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "1"}, Value: 1},
		},
		Consequence: &BlockStatement{
			Token:      token.Token{Type: token.LBRACE, Literal: "{"},
			Statements: []Statement{},
		},
		UsesVar: true,
	}
	result := fs.String()
	if result == "" {
		t.Errorf("String() should not be empty")
	}
}
