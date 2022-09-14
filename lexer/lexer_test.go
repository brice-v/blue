package lexer

import (
	"testing"

	"blue/token"
)

func TestNextToken(t *testing.T) {
	input := `=+(){},;
import name
import name.foo.bar`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.ASSIGN, "="},
		{token.PLUS, "+"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.COMMA, ","},
		{token.SEMICOLON, ";"},
		{token.IMPORT, "import"},
		{token.IMPORT_PATH, "name"},
		{token.IMPORT, "import"},
		{token.IMPORT_PATH, "name.foo.bar"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d] - tokenType wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d] - tokenLiteral wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken1(t *testing.T) {
	input := `var name = 5;
	val word = 10;
	var add = fun(x, y) {
		x + y;	
	};
	
	val ans = add(name, word);`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.VAR, "var"},
		{token.IDENT, "name"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.VAL, "val"},
		{token.IDENT, "word"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.VAR, "var"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.FUNCTION, "fun"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.VAL, "val"},
		{token.IDENT, "ans"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "name"},
		{token.COMMA, ","},
		{token.IDENT, "word"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d] - tokenType wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d] - tokenLiteral wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken2(t *testing.T) {
	input := `!-/*5;
	5 < 10 > 5;`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.FSLASH, "/"},
		{token.STAR, "*"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.GT, ">"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d] - tokenType wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d] - tokenLiteral wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken3(t *testing.T) {
	input := `if (5 < 10) {return true;} else {return false;}`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d] - tokenType wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d] - tokenLiteral wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken4(t *testing.T) {
	input := `10 == 10; 10 != 9;`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.INT, "10"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.INT, "10"},
		{token.NEQ, "!="},
		{token.INT, "9"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d] - tokenType wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d] - tokenLiteral wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken5(t *testing.T) {
	input := `[]|&^#
	%.~
	###`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.LBRACKET, "["},
		{token.RBRACKET, "]"},
		{token.PIPE, "|"},
		{token.AMPERSAND, "&"},
		{token.HAT, "^"},
		{token.HASH, "#"},
		{token.PERCENT, "%"},
		{token.DOT, "."},
		{token.TILDE, "~"},
		{token.MULTLINE_COMMENT, "###"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d] - tokenType wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d] - tokenLiteral wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextTokenMulti(t *testing.T) {
	input := `**..//>><<
	%=
	<=
	>=
	=>
	&=
	|=
	~=
	^=
	*=
	+=
	-=
	/=
	`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.POW, "**"},
		{token.RANGE, ".."},
		{token.FDIV, "//"},
		{token.RSHIFT, ">>"},
		{token.LSHIFT, "<<"},
		{token.PERCENTEQ, "%="},
		{token.LTEQ, "<="},
		{token.GTEQ, ">="},
		{token.RARROW, "=>"},
		{token.ANDEQ, "&="},
		{token.OREQ, "|="},
		{token.BINNOTEQ, "~="},
		{token.XOREQ, "^="},
		{token.MULEQ, "*="},
		{token.PLUSEQ, "+="},
		{token.MINUSEQ, "-="},
		{token.DIVEQ, "/="},

		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d] - tokenType wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d] - tokenLiteral wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextTokenMulti1(t *testing.T) {
	input := `**=
	...
	//=
	>>=
	<<=
	..<
	`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.POWEQ, "**="},
		{token.ELLIPSE, "..."},
		{token.FDIVEQ, "//="},
		{token.RSHIFTEQ, ">>="},
		{token.LSHIFTEQ, "<<="},
		{token.NONINCRANGE, "..<"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d] - tokenType wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d] - tokenLiteral wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func testNextTokenBacktick(t *testing.T) {
	input := "`"

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.BACKTICK, "`"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d] - tokenType wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d] - tokenLiteral wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextTokenNumbers(t *testing.T) {
	input := `0x1234_1234
	0o777_777
	0b1100_1100
	1234_1234
	0.1234_1234
	1234.1234
	12.12.12
	12_1234.12345_12
	_1234`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.HEX, "0x1234_1234"},
		{token.OCTAL, "0o777_777"},
		{token.BINARY, "0b1100_1100"},
		{token.INT, "1234_1234"},
		{token.FLOAT, "0.1234_1234"},
		{token.FLOAT, "1234.1234"},
		{token.FLOAT, "12.12"},
		{token.DOT, "."},
		{token.INT, "12"},
		{token.FLOAT, "12_1234.12345_12"},
		{token.IDENT, "_1234"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d] - tokenType wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d] - tokenLiteral wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextTokenNewKeywords(t *testing.T) {
	input := `for in and or not const match`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.FOR, "for"},
		{token.IN, "in"},
		{token.AND, "and"},
		{token.OR, "or"},
		{token.NOT, "not"},
		{token.CONST, "const"},
		{token.MATCH, "match"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d] - tokenType wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d] - tokenLiteral wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// func TestNextTokenNewTypeKeywords(t *testing.T) {
// 	input := `int uint type str obj
// 	enum list map any bool char rune`

// 	tests := []struct {
// 		expectedType    token.Type
// 		expectedLiteral string
// 	}{
// 		{token.INT_T, "int"},
// 		{token.UINT_T, "uint"},
// 		{token.TYPE_T, "type"},
// 		{token.STR_T, "str"},
// 		{token.OBJ_T, "obj"},
// 		{token.ENUM_T, "enum"},
// 		{token.LIST_T, "list"},
// 		{token.MAP_T, "map"},
// 		{token.ANY_T, "any"},
// 		{token.BOOL_T, "bool"},
// 		{token.CHAR_T, "char"},
// 		{token.RUNE_T, "rune"},
// 		{token.EOF, ""},
// 	}

// 	l := New(input)

// 	for i, tt := range tests {
// 		tok := l.NextToken()

// 		if tok.Type != tt.expectedType {
// 			t.Fatalf("test[%d] - tokenType wrong. expected=%q, got=%q",
// 				i, tt.expectedType, tok.Type)
// 		}

// 		if tok.Literal != tt.expectedLiteral {
// 			t.Fatalf("test[%d] - tokenLiteral wrong. expected=%q, got=%q",
// 				i, tt.expectedLiteral, tok.Literal)
// 		}
// 	}
// }

func TestNextTokenStrings(t *testing.T) {
	input := `"Hello #{world}!";'Hello #{world}!';"""Hello #{world}!"""`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.STRING_DOUBLE_QUOTE, "Hello #{world}!"},
		{token.SEMICOLON, ";"},
		{token.STRING_SINGLE_QUOTE, "Hello #{world}!"},
		{token.SEMICOLON, ";"},
		{token.RAW_STRING, "Hello #{world}!"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d] - tokenType wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d] - tokenLiteral wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestUnicodeIdentifiers(t *testing.T) {
	input := `ΣŁØÅ
ÆÐ`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "ΣŁØÅ"},
		{token.IDENT, "ÆÐ"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d] - tokenType wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d] - tokenLiteral wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestTryCatchStatement(t *testing.T) {
	input := `try {} catch (e) {} finally {}`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.TRY, "try"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.CATCH, "catch"},
		{token.LPAREN, "("},
		{token.IDENT, "e"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.FINALLY, "finally"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d] - tokenType wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d] - tokenLiteral wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestEvalExpression(t *testing.T) {
	input := `eval()`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.EVAL, "eval"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d] - tokenType wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d] - tokenLiteral wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestVarWithNum(t *testing.T) {
	input := `var abc123 = 1;`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.VAR, "var"},
		{token.IDENT, "abc123"},
		{token.ASSIGN, "="},
		{token.INT, "1"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d] - tokenType wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d] - tokenLiteral wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
