package parser

import (
	"blue/ast"
	"blue/lexer"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

// checkParserErrors prints out any and all error messages
// that happened after parsing otherwise it will return
// and the program will continue
func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func TestVarStatements(t *testing.T) {
	input := `
	var x = 5;
	var y = 10;
	var foobar = 100;`

	l := lexer.New(input, "<internal: test>")
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}
	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d", len(program.Statements))
	}

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foobar"},
	}
	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testVarStatement(t, stmt, tt.expectedIdentifier) {
			return
		}
	}
}

func testVarStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "var" {
		t.Errorf("s.TokenLiteral not `var`. got=%q", s.TokenLiteral())
		return false
	}

	varStmt, ok := s.(*ast.VarStatement)
	if !ok {
		t.Errorf("s not *ast.VarStatement. got=%T", s)
		return false
	}

	if varStmt.Names[0].Value != name {
		t.Errorf("varStmt.Names[0].Value not `%s`. got=%s", name, varStmt.Names[0].Value)
		return false
	}

	if varStmt.Names[0].TokenLiteral() != name {
		t.Errorf("s.Name not `%s`. got=%s", name, varStmt.Names[0])
		return false
	}

	return true
}

func TestValStatements(t *testing.T) {
	input := `
	val x = 5;
	val y = 10;
	val foobar = 100;`

	l := lexer.New(input, "<internal: test>")
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}
	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d", len(program.Statements))
	}

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foobar"},
	}
	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testValStatement(t, stmt, tt.expectedIdentifier) {
			return
		}
	}
}

func testValStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "val" {
		t.Errorf("s.TokenLiteral not `val`. got=%q", s.TokenLiteral())
		return false
	}

	valStmt, ok := s.(*ast.ValStatement)
	if !ok {
		t.Errorf("s not *ast.ValStatement. got=%T", s)
		return false
	}

	if valStmt.Names[0].Value != name {
		t.Errorf("valStmt.Names[0].Value not `%s`. got=%s", name, valStmt.Names[0].Value)
		return false
	}

	if valStmt.Names[0].TokenLiteral() != name {
		t.Errorf("s.Name not `%s`. got=%s", name, valStmt.Names[0])
		return false
	}

	return true
}

func TestAssignmentExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"x = 5", "x = 5"},
		{"y = true", "y = true"},
		{"foobar = y", "foobar = y"},
		{"[1, 2, 3][1] = 4", "([1, 2, 3][1]) = 4"},
		{`{"a": 1}["b"] = 2`, `({"a": 1}["b"]) = 2` + ""},
		{"x += 5", "x += 5"},
		{"y += true", "y += true"},
		{"foobar += y", "foobar += y"},
		{"[1, 2, 3][1] += 4", "([1, 2, 3][1]) += 4"},
		{`{"a": 1}["b"] += 2`, `({"a": 1}["b"]) += 2` + ""},
		{"x -= 5", "x -= 5"},
		{"y -= true", "y -= true"},
		{"foobar -= y", "foobar -= y"},
		{"[1, 2, 3][1] -= 4", "([1, 2, 3][1]) -= 4"},
		{`{"a": 1}["b"] -= 2`, `({"a": 1}["b"]) -= 2` + ""},
		{"x *= 5", "x *= 5"},
		{"y *= true", "y *= true"},
		{"foobar *= y", "foobar *= y"},
		{"[1, 2, 3][1] *= 4", "([1, 2, 3][1]) *= 4"},
		{`{"a": 1}["b"] *= 2`, `({"a": 1}["b"]) *= 2` + ""},
		{"x /= 5", "x /= 5"},
		{"y /= true", "y /= true"},
		{"foobar /= y", "foobar /= y"},
		{"[1, 2, 3][1] /= 4", "([1, 2, 3][1]) /= 4"},
		{`{"a": 1}["b"] /= 2`, `({"a": 1}["b"]) /= 2` + ""},
		{"x //= 5", "x //= 5"},
		{"y //= true", "y //= true"},
		{"foobar //= y", "foobar //= y"},
		{"[1, 2, 3][1] //= 4", "([1, 2, 3][1]) //= 4"},
		{`{"a": 1}["b"] //= 2`, `({"a": 1}["b"]) //= 2` + ""},
		{"x **= 5", "x **= 5"},
		{"y **= true", "y **= true"},
		{"foobar **= y", "foobar **= y"},
		{"[1, 2, 3][1] **= 4", "([1, 2, 3][1]) **= 4"},
		{`{"a": 1}["b"] **= 2`, `({"a": 1}["b"]) **= 2` + ""},
		{"x &= 5", "x &= 5"},
		{"y &= true", "y &= true"},
		{"foobar &= y", "foobar &= y"},
		{"[1, 2, 3][1] &= 4", "([1, 2, 3][1]) &= 4"},
		{`{"a": 1}["b"] &= 2`, `({"a": 1}["b"]) &= 2` + ""},
		{"x |= 5", "x |= 5"},
		{"y |= true", "y |= true"},
		{"foobar |= y", "foobar |= y"},
		{"[1, 2, 3][1] |= 4", "([1, 2, 3][1]) |= 4"},
		{`{"a": 1}["b"] |= 2`, `({"a": 1}["b"]) |= 2` + ""},
		{"x ~= 5", "x ~= 5"},
		{"y ~= true", "y ~= true"},
		{"foobar ~= y", "foobar ~= y"},
		{"[1, 2, 3][1] ~= 4", "([1, 2, 3][1]) ~= 4"},
		{`{"a": 1}["b"] ~= 2`, `({"a": 1}["b"]) ~= 2` + ""},
		{"x >>= 5", "x >>= 5"},
		{"y >>= true", "y >>= true"},
		{"foobar >>= y", "foobar >>= y"},
		{"[1, 2, 3][1] >>= 4", "([1, 2, 3][1]) >>= 4"},
		{`{"a": 1}["b"] >>= 2`, `({"a": 1}["b"]) >>= 2` + ""},
		{"x <<= 5", "x <<= 5"},
		{"y <<= true", "y <<= true"},
		{"foobar <<= y", "foobar <<= y"},
		{"[1, 2, 3][1] <<= 4", "([1, 2, 3][1]) <<= 4"},
		{`{"a": 1}["b"] <<= 2`, `({"a": 1}["b"]) <<= 2` + ""},
		{"x %= 5", "x %= 5"},
		{"y %= true", "y %= true"},
		{"foobar %= y", "foobar %= y"},
		{"[1, 2, 3][1] %= 4", "([1, 2, 3][1]) %= 4"},
		{`{"a": 1}["b"] %= 2`, `({"a": 1}["b"]) %= 2` + ""},
		{"x ^= 5", "x ^= 5"},
		{"y ^= true", "y ^= true"},
		{"foobar ^= y", "foobar ^= y"},
		{"[1, 2, 3][1] ^= 4", "([1, 2, 3][1]) ^= 4"},
		{`{"a": 1}["b"] ^= 2`, `({"a": 1}["b"]) ^= 2` + ""},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input, "<internal: test>")
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if tt.expected != program.String() {
			t.Errorf("tt.expected != program.String(). want=%q, got=%q", tt.expected, program.String())
		}
	}
}

func TestVarStatementsWithOtherAssignmentTokens(t *testing.T) {
	input := `
	var x += 1;
	var y -= 2;
	var z /= 3;
	var a //= 4;
	var b *= 5;
	var c **= 6;
	var d &= 7;
	var e |= 8;
	var f ~= 9;
	var g >>= 10;
	var h <<= 11;
	var i %= 12;
	var j ^= 13;
	`

	l := lexer.New(input, "<internal: test>")
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}
	if len(program.Statements) != 13 {
		t.Fatalf("program.Statements does not contain 13 statements. got=%d", len(program.Statements))
	}

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"z"},
		{"a"},
		{"b"},
		{"c"},
		{"d"},
		{"e"},
		{"f"},
		{"g"},
		{"h"},
		{"i"},
		{"j"},
	}
	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testVarStatement(t, stmt, tt.expectedIdentifier) {
			return
		}
	}
}

func TestReturnStatements(t *testing.T) {
	input := `return 5;
	return 10;
	return 999_12.1234;`

	l := lexer.New(input, "<internal: test>")
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d", len(program.Statements))
	}

	for _, stmt := range program.Statements {
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("stmt not *ast.ReturnStatement. got=%T", stmt)
			continue
		}
		if returnStmt.TokenLiteral() != "return" {
			t.Errorf("returnStmt.TokenLiteral() not `return` got `%q`", returnStmt.TokenLiteral())
		}
	}
}

func TestIdentifierExpressions(t *testing.T) {
	input := "foobar;"
	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program does not have enough statements. got=%d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not an *ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	ident, ok := stmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("expression not *ast.Identifier. got=%T", stmt.Expression)
	}

	if ident.Value != "foobar" {
		t.Errorf("ident.Value not %s. got=%s", "foobar", ident.Value)
	}
	if ident.TokenLiteral() != "foobar" {
		t.Errorf("ident.TokenLiteral() not %s. got=%s", "foobar", ident.TokenLiteral())
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5_5;"
	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program does not have enough statements. got=%d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp is not an *ast.IntegerLiteral. got=%T", stmt.Expression)
	}
	if literal.Value != 55 {
		t.Fatalf("literal.Value not %d. got %d", 55, literal.Value)
	}
}

func TestFloatLiteralExpression(t *testing.T) {
	input := "5.0_1;"
	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program does not have enough statements. got=%d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.FloatLiteral)
	if !ok {
		t.Fatalf("exp is not an *ast.FloatLiteral. got=%T", stmt.Expression)
	}
	if literal.Value != 5.01 {
		t.Fatalf("literal.Value not %f. got %f", 5.01, literal.Value)
	}

}

func TestHexLiteralExpression(t *testing.T) {
	input := "0x1234_1234;"
	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program does not have enough statements. got=%d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.HexLiteral)
	if !ok {
		t.Fatalf("exp is not an *ast.HexLiteral. got=%T", stmt.Expression)
	}
	if literal.Value != 0x12341234 {
		t.Fatalf("literal.Value not %x. got %x", 0x12341234, literal.Value)
	}

}

func TestOctalLiteralExpression(t *testing.T) {
	input := "0o777_111;"
	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program does not have enough statements. got=%d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.OctalLiteral)
	if !ok {
		t.Fatalf("exp is not an *ast.OctalLiteral. got=%T", stmt.Expression)
	}
	if literal.Value != 0777111 {
		t.Fatalf("literal.Value not %o. got %o", 0777111, literal.Value)
	}

}

func TestBinaryLiteralExpression(t *testing.T) {
	input := "0b1111_0000;"
	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program does not have enough statements. got=%d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.BinaryLiteral)
	if !ok {
		t.Fatalf("exp is not an *ast.BinaryLiteral. got=%T", stmt.Expression)
	}
	if literal.Value != 240 {
		t.Fatalf("literal.Value not %b. got %b", 240, literal.Value)
	}
}

func TestUIntLiteralExpression(t *testing.T) {
	input := "0u1234;"
	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program does not have enough statements. got=%d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.UIntegerLiteral)
	if !ok {
		t.Fatalf("exp is not an *ast.UIntegerLiteral. got=%T", stmt.Expression)
	}
	if literal.Value != uint64(1234) {
		t.Fatalf("literal.Value not %d. got %d", uint64(1234), literal.Value)
	}
}

func TestBigIntLiteralExpression(t *testing.T) {
	input := "2134n;"
	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program does not have enough statements. got=%d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.BigIntegerLiteral)
	if !ok {
		t.Fatalf("exp is not an *ast.BigIntegerLiteral. got=%T", stmt.Expression)
	}
	bi := new(big.Int).SetInt64(2134)
	if !reflect.DeepEqual(literal.Value, bi) {
		t.Fatalf("literal.Value not %d. got %d", bi, literal.Value)
	}
}

func TestBigFloatLiteralExpression(t *testing.T) {
	input := "2134.1234n ;"
	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program does not have enough statements. got=%d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.BigFloatLiteral)
	if !ok {
		t.Fatalf("exp is not an *ast.BigFloatLiteral. got=%T", stmt.Expression)
	}
	bf := decimal.NewFromFloat(2134.1234)
	if !reflect.DeepEqual(literal.Value, bf) {
		t.Fatalf("literal.Value not %d. got %d", bf, literal.Value)
	}
}

func TestParsingPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input    string
		operator string
		value    interface{}
	}{
		{"not 5;", "not", 5},
		{"-15;", "-", 15},
		{"~1;", "~", 1},
		{"not true;", "not", true},
	}

	for _, tt := range prefixTests {
		l := lexer.New(tt.input, "<internal: test>")
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d", 1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("stmt is not an ast.PrefixExpression. got=%T", stmt.Expression)
		}
		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not `%s`. got `%s`", tt.operator, exp.Operator)
		}
		if !testLiteralExpression(t, exp.Right, tt.value) {
			return
		}
	}
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integ, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il not *ast.IntegerLiteral. got=%T", il)
		return false
	}
	if integ.Value != value {
		t.Errorf("integ.Value not %d. got=%d", value, integ.Value)
		return false
	}
	tokenLiteral := strings.Replace(integ.TokenLiteral(), "_", "", -1)
	if tokenLiteral != fmt.Sprintf("%d", value) {
		t.Errorf("integ.TokenLiteral() not %d. got=%s", value, integ.TokenLiteral())
		return false
	}
	return true
}

func TestParsingInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  int64
		operator   string
		rightValue int64
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 ^ 5;", 5, "^", 5},
		{"5 & 5;", 5, "&", 5},
		{"5 | 5;", 5, "|", 5},
		{"5 % 5;", 5, "%", 5},
		{"5 and 5;", 5, "and", 5},
		{"5 or 5;", 5, "or", 5},
		{"5 != 5;", 5, "!=", 5},
		{"5 <= 5;", 5, "<=", 5},
		{"5 >= 5;", 5, ">=", 5},
		{"5 // 5;", 5, "//", 5},
		{"5 == 5;", 5, "==", 5},
		{"5..5;", 5, "..", 5},
		{"5..<5;", 5, "..<", 5},
		{"5 in 5", 5, "in", 5},
	}
	for _, tt := range infixTests {
		l := lexer.New(tt.input, "<internal: test>")
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d", 1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.InfixExpression)
		if !ok {
			t.Fatalf("stmt is not an ast.InfixExpression. got=%T", stmt.Expression)
		}

		if !testInfixExpression(t, exp, tt.leftValue, tt.operator, tt.rightValue) {
			return
		}
		if !testInfixExpression(t, exp, tt.rightValue, tt.operator, tt.rightValue) {
			return
		}
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"not-a",
			"(not(-a))",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
		{
			"a + b - c",
			"((a + b) - c)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a + b / c",
			"(a + (b / c))",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f)",
		},
		{
			"3 + 4; -5 * 5",
			"(3 + 4)\n((-5) * 5)",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4))",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"5 >= 4 == 3 <= 4",
			"((5 >= 4) == (3 <= 4))",
		},
		{
			"5 <= 4 != 3 >= 4",
			"((5 <= 4) != (3 >= 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"true",
			"true",
		},
		{
			"false",
			"false",
		},
		{
			"3 > 5 == false",
			"((3 > 5) == false)",
		},
		{
			"3 < 5 == true",
			"((3 < 5) == true)",
		},
		{
			"1 + (2 + 3) + 4",
			"((1 + (2 + 3)) + 4)",
		},
		{
			"(5 + 5) * 2",
			"((5 + 5) * 2)",
		},
		{
			"2 / (5 + 5)",
			"(2 / (5 + 5))",
		},
		{
			"(5 + 5) * 2 * (5 + 5)",
			"(((5 + 5) * 2) * (5 + 5))",
		},
		{
			"-(5 + 5)",
			"(-(5 + 5))",
		},
		{
			"not(true == true)",
			"(not(true == true))",
		},
		{
			"a + add(b * c) + d",
			"((a + add((b * c))) + d)",
		},
		{
			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))",
			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))",
		},
		{
			"add(a + b + c * d / f + g)",
			"add((((a + b) + ((c * d) / f)) + g))",
		},
	}
	for i, tt := range tests {
		l := lexer.New(tt.input, "<internal: test>")
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		if actual != tt.expected {
			t.Errorf("[%d] expected=%q, got=%q", i, tt.expected, actual)
		}
	}
}

// testIdentifier is a helper expression to test for an identifier
func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Fatalf("exp not *ast.Identifier. got=%T", exp)
		return false
	}

	if ident.Value != value {
		t.Errorf("value not %s. got=%s", value, ident.Value)
		return false
	}

	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral() not %s. got=%s", value, ident.TokenLiteral())
		return false
	}
	return true
}

// testLiteralExpression switches on the expression type
// and tries to test a variety of literal tests (if they
// exist for the type)
func testLiteralExpression(
	t *testing.T,
	exp ast.Expression,
	expected interface{},
) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	}
	t.Errorf("type of exp not handled. got=%T", exp)
	return false
}

func testInfixExpression(
	t *testing.T,
	exp ast.Expression,
	left interface{},
	operator string,
	right interface{},
) bool {
	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("exp is not ast.OperatorExpression. got=%T(%s)", exp, exp)
		return false
	}

	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}

	if opExp.Operator != operator {
		t.Errorf("exp.Operator is not `%s`. got=%q", operator, opExp.Operator)
		return false
	}

	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}
	return true
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	bo, ok := exp.(*ast.Boolean)
	if !ok {
		t.Errorf("exp not *ast.Boolean. got=%T", exp)
		return false
	}

	if bo.Value != value {
		t.Errorf("bo.Value not %t. got=%t", value, bo.Value)
		return false
	}

	if bo.TokenLiteral() != fmt.Sprintf("%t", value) {
		t.Errorf("bo.TokenLiteral() not %t. got=%s", value, bo.TokenLiteral())
		return false
	}
	return true
}

func TestIfExpression(t *testing.T) {
	input := `if (x < y) { x }`

	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n", 1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.IfExpression. got=%T",
			stmt.Expression)
	}

	if !testInfixExpression(t, exp.Conditions[0], "x", "<", "y") {
		return
	}

	if len(exp.Consequences[0].Statements) != 1 {
		t.Errorf("consequence is not 1 statements. got=%d\n",
			len(exp.Consequences[0].Statements))
	}

	consequence, ok := exp.Consequences[0].Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not ast.ExpressionStatement. got=%T",
			exp.Consequences[0].Statements[0])
	}

	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	if exp.Alternative != nil {
		t.Errorf("exp.Alternative.Statements was not nil. got=%+v", exp.Alternative)
	}
}

func TestIfElseExpression(t *testing.T) {
	input := `if (x < y) { x } else { y }`

	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
			1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.IfExpression. got=%T", stmt.Expression)
	}

	if !testInfixExpression(t, exp.Conditions[0], "x", "<", "y") {
		return
	}

	if len(exp.Consequences[0].Statements) != 1 {
		t.Errorf("consequence is not 1 statements. got=%d\n",
			len(exp.Consequences[0].Statements))
	}

	consequence, ok := exp.Consequences[0].Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not ast.ExpressionStatement. got=%T",
			exp.Consequences[0].Statements[0])
	}

	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	if len(exp.Alternative.Statements) != 1 {
		t.Errorf("exp.Alternative.Statements does not contain 1 statements. got=%d\n",
			len(exp.Alternative.Statements))
	}

	alternative, ok := exp.Alternative.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not ast.ExpressionStatement. got=%T",
			exp.Alternative.Statements[0])
	}

	if !testIdentifier(t, alternative.Expression, "y") {
		return
	}
}

func TestFunctionLiteralParsing(t *testing.T) {
	input := `fun(x, y) { x + y; }`

	l := lexer.New(input, "<internal: test>")
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Body does not contain %d statements. got=%d", 1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	function, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.FunctionLiteral. got=%T", stmt.Expression)
	}

	if len(function.Parameters) != 2 {
		t.Fatalf("function literal parameters wrong. want 2, got=%d", len(function.Parameters))
	}

	testLiteralExpression(t, function.Parameters[0], "x")
	testLiteralExpression(t, function.Parameters[1], "y")

	if len(function.Body.Statements) != 1 {
		t.Fatalf("function.Body.Statements does not have 1 statement. got=%d", len(function.Body.Statements))
	}

	bodyStmt, ok := function.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("function body stmt is not ast.ExpressionStatement. got=%T", function.Body.Statements[0])
	}

	testInfixExpression(t, bodyStmt.Expression, "x", "+", "y")
}

func TestFunctionStatementParsing(t *testing.T) {
	input := `fun name(x, y) { x + y; }`

	l := lexer.New(input, "<internal: test>")
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Body does not contain %d statements. got=%d", 1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.FunctionStatement. got=%T", program.Statements[0])
	}

	name := stmt.Name.String()
	if name != "name" {
		t.Fatalf("stmt.Name is not %s. got=%s", "name", stmt.Name.String())
	}

	if len(stmt.Parameters) != 2 {
		t.Fatalf("function literal parameters wrong. want 2, got=%d", len(stmt.Parameters))
	}

	testLiteralExpression(t, stmt.Parameters[0], "x")
	testLiteralExpression(t, stmt.Parameters[1], "y")

	if len(stmt.Body.Statements) != 1 {
		t.Fatalf("stmt.Body.Statements does not have 1 statement. got=%d", len(stmt.Body.Statements))
	}

	bodyStmt, ok := stmt.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt body stmt is not ast.ExpressionStatement. got=%T", stmt.Body.Statements[0])
	}

	testInfixExpression(t, bodyStmt.Expression, "x", "+", "y")
}

func TestFunctionParameterParsing(t *testing.T) {
	tests := []struct {
		input          string
		expectedParams []string
	}{
		{input: "fun() {};", expectedParams: []string{}},
		{input: "fun(x) {};", expectedParams: []string{"x"}},
		{input: "fun(x, y, z) {};", expectedParams: []string{"x", "y", "z"}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input, "<internal: test>")
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		function := stmt.Expression.(*ast.FunctionLiteral)

		if len(function.Parameters) != len(tt.expectedParams) {
			t.Errorf("length parameters wrong. want %d, got=%d\n",
				len(tt.expectedParams), len(function.Parameters))
		}

		for i, ident := range tt.expectedParams {
			testLiteralExpression(t, function.Parameters[i], ident)
		}
	}
}

func TestLambdaLiteralParsing(t *testing.T) {
	input := `|x, y| => { x + y; }`

	l := lexer.New(input, "<internal: test>")
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Body does not contain %d statements. got=%d", 1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	function, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.FunctionLiteral. got=%T", stmt.Expression)
	}

	if len(function.Parameters) != 2 {
		t.Fatalf("function literal parameters wrong. want 2, got=%d", len(function.Parameters))
	}

	testLiteralExpression(t, function.Parameters[0], "x")
	testLiteralExpression(t, function.Parameters[1], "y")

	if len(function.Body.Statements) != 1 {
		t.Fatalf("function.Body.Statements does not have 1 statement. got=%d", len(function.Body.Statements))
	}

	bodyStmt, ok := function.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("function body stmt is not ast.ExpressionStatement. got=%T", function.Body.Statements[0])
	}

	testInfixExpression(t, bodyStmt.Expression, "x", "+", "y")
}

func TestCallExpressionParsing(t *testing.T) {
	input := "add(1, 2 * 3, 4 + 5);"

	l := lexer.New(input, "<internal: test>")
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n", 1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.CallExpression. got=%T", stmt.Expression)
	}

	if !testIdentifier(t, exp.Function, "add") {
		return
	}

	if len(exp.Arguments) != 3 {
		t.Fatalf("wrong length of arguments. got=%d", len(exp.Arguments))
	}

	testLiteralExpression(t, exp.Arguments[0], 1)
	testInfixExpression(t, exp.Arguments[1], 2, "*", 3)
	testInfixExpression(t, exp.Arguments[2], 4, "+", 5)
}

func TestBrokenParsing(t *testing.T) {
	input := `fun abc(x) { x + y };  abc(4);`
	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if len(program.Statements) != 2 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n", 2, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.FunctionStatement. got=%T", program.Statements[0])
	}

	stmt2, ok := program.Statements[1].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt2 is not *ast.ExpressionStatement. got=%T", program.Statements[1])
	}

	name := stmt.Name.String()
	if name != "abc" {
		t.Fatalf("stmt.Name is not %s. got=%s", "abc", stmt.Name.String())
	}

	if len(stmt.Parameters) != 1 {
		t.Fatalf("function literal parameters wrong. want 1, got=%d", len(stmt.Parameters))
	}

	testLiteralExpression(t, stmt.Parameters[0], "x")

	if len(stmt.Body.Statements) != 1 {
		t.Fatalf("stmt.Body.Statements does not have 1 statement. got=%d", len(stmt.Body.Statements))
	}

	bodyStmt, ok := stmt.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt body stmt is not ast.ExpressionStatement. got=%T", stmt.Body.Statements[0])
	}

	testInfixExpression(t, bodyStmt.Expression, "x", "+", "y")

	exp, ok := stmt2.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.CallExpression. got=%T", stmt2.Expression)
	}

	if !testIdentifier(t, exp.Function, "abc") {
		return
	}

	if len(exp.Arguments) != 1 {
		t.Fatalf("wrong length of arguments. got=%d", len(exp.Arguments))
	}

	testLiteralExpression(t, exp.Arguments[0], 4)
}

func TestStringLiteralExpression(t *testing.T) {
	input := `"Hello #{world}";`
	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program does not have enough statements. got=%d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("exp is not an *ast.StringLiteral. got=%T", stmt.Expression)
	}
	if literal.Value != `Hello #{world}` {
		t.Fatalf("literal.Value not %s. got %s", `Hello #{world}`, literal.Value)
	}
	if len(literal.InterpolationValues) != 1 {
		t.Fatalf("literal.InterpolationValues not %d. got=%d", 1, len(literal.InterpolationValues))
	}
	ident, ok := literal.InterpolationValues[0].(*ast.Identifier)
	if !ok {
		t.Fatalf("literal.InterpolationValues[0] is not *ast.Identifier. got=%T", literal.InterpolationValues[0])
	}
	if ident.Value != "world" {
		t.Fatalf("ident.Literal is not %s. got %s", "world", ident.Value)
	}
}

func TestStringLiteralExpression1(t *testing.T) {
	input := `"Hello #{x + y}";`
	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program does not have enough statements. got=%d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("exp is not an *ast.StringLiteral. got=%T", stmt.Expression)
	}
	if literal.Value != `Hello #{x + y}` {
		t.Fatalf("literal.Value not %s. got %s", `Hello #{x + y}`, literal.Value)
	}
	if len(literal.InterpolationValues) != 1 {
		t.Fatalf("literal.InterpolationValues not %d. got=%d", 1, len(literal.InterpolationValues))
	}
	exp, ok := literal.InterpolationValues[0].(*ast.InfixExpression)
	if !ok {
		t.Fatalf("literal.InterpolationValues[0] is not *ast.InfixExpression. got=%T", literal.InterpolationValues[0])
	}
	if !testInfixExpression(t, exp, "x", "+", "y") {
		return
	}
}

func TestStringLiteralExpression2(t *testing.T) {
	input := `"Hello #{x + y}  #{world}";`
	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program does not have enough statements. got=%d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("exp is not an *ast.StringLiteral. got=%T", stmt.Expression)
	}
	if literal.Value != `Hello #{x + y}  #{world}` {
		t.Fatalf("literal.Value not %s. got %s", `Hello #{x + y}  #{world}`, literal.Value)
	}
	if len(literal.InterpolationValues) != 2 {
		t.Fatalf("literal.InterpolationValues not %d. got=%d", 2, len(literal.InterpolationValues))
	}
	exp, ok := literal.InterpolationValues[0].(*ast.InfixExpression)
	if !ok {
		t.Fatalf("literal.InterpolationValues[0] is not *ast.InfixExpression. got=%T", literal.InterpolationValues[0])
	}
	if !testInfixExpression(t, exp, "x", "+", "y") {
		return
	}
	exp1, ok := literal.InterpolationValues[1].(*ast.Identifier)
	if !ok {
		t.Fatalf("literal.InterpolationValues[1] is not an identifier. got=%T", literal.InterpolationValues[1])
	}
	if !testIdentifier(t, exp1, "world") {
		return
	}
}

func TestParsingListLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast *ast.ExpressionStatement. got=%T", program.Statements[0])
	}
	list, ok := stmt.Expression.(*ast.ListLiteral)
	if !ok {
		t.Fatalf("exp not ast.ListLiteral. got=%T", stmt.Expression)
	}

	if len(list.Elements) != 3 {
		t.Fatalf("len(list.Elements) not 3. got=%d", len(list.Elements))
	}

	testIntegerLiteral(t, list.Elements[0], 1)
	testInfixExpression(t, list.Elements[1], 2, "*", 2)
	testInfixExpression(t, list.Elements[2], 3, "+", 3)
}

func TestParsingMapLiteralsStringKeys(t *testing.T) {
	input := `{"one": 1, "two": 2, "three": 3}`

	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	myMap, ok := stmt.Expression.(*ast.MapLiteral)
	if !ok {
		t.Fatalf("exp is not ast.MapLiteral. got=%T", stmt.Expression)
	}

	if len(myMap.Pairs) != 3 {
		t.Errorf("myMap.Pairs has wrong length. got=%d", len(myMap.Pairs))
	}

	expected := map[string]int64{
		"one":   1,
		"two":   2,
		"three": 3,
	}

	for key, value := range myMap.Pairs {
		literal, ok := key.(*ast.StringLiteral)
		if !ok {
			t.Errorf("key is not ast.StringLiteral. got=%T", key)
		}

		expectedValue := expected[literal.StringWithoutQuotes()]

		testIntegerLiteral(t, value, expectedValue)
	}
}

func TestParsingMapLiteralsIdentifierKeys(t *testing.T) {
	input := "{one: 1, two: 2, three: 3}"

	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	myMap, ok := stmt.Expression.(*ast.MapLiteral)
	if !ok {
		t.Fatalf("exp is not ast.MapLiteral. got=%T", stmt.Expression)
	}

	if len(myMap.Pairs) != 3 {
		t.Errorf("myMap.Pairs has wrong length. got=%d", len(myMap.Pairs))
	}

	expected := map[string]int64{
		"one":   1,
		"two":   2,
		"three": 3,
	}

	for key, value := range myMap.Pairs {
		literal, ok := key.(*ast.Identifier)
		if !ok {
			t.Errorf("key is not ast.Identifier. got=%T", key)
		}

		expectedValue := expected[literal.String()]

		testIntegerLiteral(t, value, expectedValue)
	}
}

func TestParsingMapLiteralsBooleanKeys(t *testing.T) {
	input := `{true: 1, false: 2}`

	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	myMap, ok := stmt.Expression.(*ast.MapLiteral)
	if !ok {
		t.Fatalf("exp is not ast.MapLiteral. got=%T", stmt.Expression)
	}

	expected := map[string]int64{
		"true":  1,
		"false": 2,
	}

	if len(myMap.Pairs) != len(expected) {
		t.Errorf("myMap.Pairs has wrong length. got=%d", len(myMap.Pairs))
	}

	for key, value := range myMap.Pairs {
		boolean, ok := key.(*ast.Boolean)
		if !ok {
			t.Errorf("key is not ast.BooleanLiteral. got=%T", key)
			continue
		}

		expectedValue := expected[boolean.String()]
		testIntegerLiteral(t, value, expectedValue)
	}
}

func TestParsingMapLiteralsIntegerKeys(t *testing.T) {
	input := `{1: 1, 2: 2, 3: 3}`

	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	myMap, ok := stmt.Expression.(*ast.MapLiteral)
	if !ok {
		t.Fatalf("exp is not ast.MapLiteral. got=%T", stmt.Expression)
	}

	expected := map[string]int64{
		"1": 1,
		"2": 2,
		"3": 3,
	}

	if len(myMap.Pairs) != len(expected) {
		t.Errorf("myMap.Pairs has wrong length. got=%d", len(myMap.Pairs))
	}

	for key, value := range myMap.Pairs {
		integer, ok := key.(*ast.IntegerLiteral)
		if !ok {
			t.Errorf("key is not ast.IntegerLiteral. got=%T", key)
			continue
		}

		expectedValue := expected[integer.String()]

		testIntegerLiteral(t, value, expectedValue)
	}
}

func TestParsingEmptyMapLiteral(t *testing.T) {
	input := "{}"

	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	myMap, ok := stmt.Expression.(*ast.MapLiteral)
	if !ok {
		t.Fatalf("exp is not ast.MapLiteral. got=%T", stmt.Expression)
	}

	if len(myMap.Pairs) != 0 {
		t.Errorf("myMap.Pairs has wrong length. got=%d", len(myMap.Pairs))
	}
}

func TestParsingMapLiteralsWithExpressions(t *testing.T) {
	input := `{"one": 0 + 1, "two": 10 - 8, "three": 15 / 5}`

	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	myMap, ok := stmt.Expression.(*ast.MapLiteral)
	if !ok {
		t.Fatalf("exp is not ast.MapLiteral. got=%T", stmt.Expression)
	}

	if len(myMap.Pairs) != 3 {
		t.Errorf("myMap.Pairs has wrong length. got=%d", len(myMap.Pairs))
	}

	tests := map[string]func(ast.Expression){
		"one": func(e ast.Expression) {
			testInfixExpression(t, e, 0, "+", 1)
		},
		"two": func(e ast.Expression) {
			testInfixExpression(t, e, 10, "-", 8)
		},
		"three": func(e ast.Expression) {
			testInfixExpression(t, e, 15, "/", 5)
		},
	}

	for key, value := range myMap.Pairs {
		literal, ok := key.(*ast.StringLiteral)
		if !ok {
			t.Errorf("key is not ast.StringLiteral. got=%T", key)
			continue
		}

		testFunc, ok := tests[literal.StringWithoutQuotes()]
		if !ok {
			t.Errorf("No test function for key %q found", literal.String())
			continue
		}

		testFunc(value)
	}
}

func TestParsingIndexExpressions(t *testing.T) {
	input := "mylist[1_1 + 1_1]"

	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not an *ast.ExpressionStatment. got=%T", program.Statements[0])
	}
	indexExp, ok := stmt.Expression.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("indxExp not *ast.IndexExpression. got=%T", stmt.Expression)
	}

	if !testIdentifier(t, indexExp.Left, "mylist") {
		return
	}

	if !testInfixExpression(t, indexExp.Index, 11, "+", 11) {
		return
	}
}

func TestParsingMemberAccessExpression(t *testing.T) {
	input := "test.foo"

	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not an *ast.ExpressionStatement. got=%T", program.Statements[0])
	}
	t.Logf("stmt: %#v", stmt)

	exp, ok := stmt.Expression.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("exp not *ast.IndexExpression. got=%T", stmt.Expression)
	}

	ident, ok := exp.Left.(*ast.Identifier)
	if !ok {
		t.Fatalf("exp.Left not *ast.Identifier. got=%T", stmt.Expression)
	}

	if !testIdentifier(t, ident, "test") {
		return
	}

	index, ok := exp.Index.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("exp.Index not *ast.StringLiteral. got=%T", stmt.Expression)
	}

	if index.Value != "foo" {
		t.Fatalf("index.Value != \"foo\"")
	}
}

func TestForExpression(t *testing.T) {
	input := `for (x < y) { var z = x + y; };`

	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d", 1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.ForExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.ForExpression. got=%T", stmt.Expression)
	}

	if !testInfixExpression(t, exp.Condition, "x", "<", "y") {
		return
	}

	if len(exp.Consequence.Statements) != 1 {
		t.Fatalf("exp.Consequence.Statements does not contain %d statements. got=%d", 1, len(exp.Consequence.Statements))
	}

	cons, ok := exp.Consequence.Statements[0].(*ast.VarStatement)
	if !ok {
		t.Fatalf("exp.Consequence.Statements[0] is not *ast.VarStatement. got=%T", exp.Consequence.Statements[0])
	}

	if !testVarStatement(t, cons, "z") {
		return
	}

}

func TestBrokenParsingOfFile(t *testing.T) {
	input := `import time
	import psutil
	
	fun handle_ws_realtime_cpus(ws) {
		var sub = pubsub.subscribe('/realtime/cpus');
		println("ws = #{ws}, sub = #{sub}")
		for (true) {
			val topic_msg = sub.recv();
			match topic_msg {
				{topic: '/realtime/cpus', msg: _} => {
					ws.send(topic_msg.msg);
				},
				_ => {
					println("exiting /realtime/cpus");
					sub.unsubscribe();
					return NULL;
				},
			}
		}
	}
	
	fun cpus() {
		val usage = psutil.cpu.percent();
		println("/cpus called: usage = #{usage}")
		usage.to_json()
	}
	
	fun realtime_cpus(ws) {
		println("/realtime/cpus called: ws=#{ws}");
		val y = spawn(handle_ws_realtime_cpus, [ws]);
		println(y);
	}`

	l := lexer.New(input, "<internal: test>")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 5 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d", 5, len(program.Statements))
	}
}
