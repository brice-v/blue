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

	l := New(input, "<internal: test>")

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

	l := New(input, "<internal: test>")

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
		{token.NOT, "!"},
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

	l := New(input, "<internal: test>")

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

	l := New(input, "<internal: test>")

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

	l := New(input, "<internal: test>")

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

	l := New(input, "<internal: test>")

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
	&&=
	||=
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
		{token.ANDANDEQ, "&&="},
		{token.OROREQ, "||="},

		{token.EOF, ""},
	}

	l := New(input, "<internal: test>")

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
	@{
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
		{token.ATLBRACE, "@{"},
		{token.EOF, ""},
	}

	l := New(input, "<internal: test>")

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

	l := New(input, "<internal: test>")

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

	l := New(input, "<internal: test>")

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

	l := New(input, "<internal: test>")

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

// 	l := New(input, "<internal: test>")

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
	input := `"Hello #{world}!";'Hello #{world}!';"""Hello #{world}!""")`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.STRING_DOUBLE_QUOTE, "Hello #{world}!"},
		{token.SEMICOLON, ";"},
		{token.STRING_SINGLE_QUOTE, "Hello #{world}!"},
		{token.SEMICOLON, ";"},
		{token.RAW_STRING, "Hello #{world}!"},
		{token.RPAREN, ")"},
		{token.EOF, ""},
	}

	l := New(input, "<internal: test>")

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

	l := New(input, "<internal: test>")

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

	l := New(input, "<internal: test>")

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

	l := New(input, "<internal: test>")

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

	l := New(input, "<internal: test>")

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

func TestNextTokenEscapeNewline(t *testing.T) {
	input := `"hello\nworld"`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.STRING_DOUBLE_QUOTE {
		t.Fatalf("expected STRING_DOUBLE_QUOTE, got %v", tok.Type)
	}
	if tok.Literal != "hello\nworld" {
		t.Errorf("expected escaped newline, got %q", tok.Literal)
	}
}

func TestNextTokenEscapeCarriageReturn(t *testing.T) {
	input := `"hello\rworld"`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.STRING_DOUBLE_QUOTE {
		t.Fatalf("expected STRING_DOUBLE_QUOTE, got %v", tok.Type)
	}
	if tok.Literal != "hello\rworld" {
		t.Errorf("expected escaped carriage return, got %q", tok.Literal)
	}
}

func TestNextTokenEscapeTab(t *testing.T) {
	input := `"hello\tworld"`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.STRING_DOUBLE_QUOTE {
		t.Fatalf("expected STRING_DOUBLE_QUOTE, got %v", tok.Type)
	}
	if tok.Literal != "hello\tworld" {
		t.Errorf("expected escaped tab, got %q", tok.Literal)
	}
}

func TestNextTokenEscapeBackslash(t *testing.T) {
	input := `"hello\\world"`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.STRING_DOUBLE_QUOTE {
		t.Fatalf("expected STRING_DOUBLE_QUOTE, got %v", tok.Type)
	}
	if tok.Literal != "hello\\world" {
		t.Errorf("expected escaped backslash, got %q", tok.Literal)
	}
}

func TestNextTokenEscapeQuote(t *testing.T) {
	input := `"say \"hi\""`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.STRING_DOUBLE_QUOTE {
		t.Fatalf("expected STRING_DOUBLE_QUOTE, got %v", tok.Type)
	}
	if tok.Literal != `say "hi"` {
		t.Errorf("expected escaped quote, got %q", tok.Literal)
	}
}

func TestNextTokenEscapeHex(t *testing.T) {
	input := `"\x41"`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.STRING_DOUBLE_QUOTE {
		t.Fatalf("expected STRING_DOUBLE_QUOTE, got %v", tok.Type)
	}
	if tok.Literal != "A" {
		t.Errorf("expected hex decoded A, got %q", tok.Literal)
	}
}

func TestNextTokenUnfinishedString(t *testing.T) {
	input := `"hello`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.ILLEGAL {
		t.Fatalf("expected ILLEGAL for unterminated string, got %v", tok.Type)
	}
}

func TestNextTokenUnfinishedSingleQuoteString(t *testing.T) {
	input := `'hello`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.ILLEGAL {
		t.Fatalf("expected ILLEGAL for unterminated single quote string, got %v", tok.Type)
	}
}

func TestNextTokenRegex(t *testing.T) {
	input := `r/^[a-z]+$/`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.REGEX {
		t.Fatalf("expected REGEX, got %v", tok.Type)
	}
	if tok.Literal != "^[a-z]+$" {
		t.Errorf("expected regex literal, got %q", tok.Literal)
	}
}

func TestNextTokenRegexWithEscapedSlash(t *testing.T) {
	input := `r/\/test/`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.REGEX {
		t.Fatalf("expected REGEX, got %v", tok.Type)
	}
	if tok.Literal != "/test" {
		t.Errorf("expected regex with escaped slash, got %q", tok.Literal)
	}
}

func TestNextTokenRegexUnfinished(t *testing.T) {
	input := `r/unclosed`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.ILLEGAL {
		t.Fatalf("expected ILLEGAL for unterminated regex, got %v", tok.Type)
	}
}

func TestNextTokenBigInt(t *testing.T) {
	input := `123456789012345678901234567890n`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.BIGINT {
		t.Fatalf("expected BIGINT, got %v", tok.Type)
	}
	if tok.Literal != "123456789012345678901234567890n" {
		t.Errorf("expected bigint literal, got %q", tok.Literal)
	}
}

func TestNextTokenBigFloat(t *testing.T) {
	input := `12345678901234567890.1234567890n`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.BIGFLOAT {
		t.Fatalf("expected BIGFLOAT, got %v", tok.Type)
	}
}

func TestNextTokenUint(t *testing.T) {
	input := `0u1234567890`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.UINT {
		t.Fatalf("expected UINT, got %v", tok.Type)
	}
}

func TestNextTokenFromImport(t *testing.T) {
	input := `from foo import [bar, baz]`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.FROM, "from"},
		{token.IMPORT_PATH, "foo"},
		{token.IMPORT, "import"},
		{token.LBRACKET, "["},
		{token.IDENT, "bar"},
		{token.COMMA, ","},
		{token.IDENT, "baz"},
		{token.RBRACKET, "]"},
		{token.EOF, ""},
	}

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

func TestNextTokenMultiLineComment(t *testing.T) {
	input := `### this is a comment ###`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.MULTLINE_COMMENT {
		t.Fatalf("expected MULTLINE_COMMENT, got %v", tok.Type)
	}
}

func TestNextTokenDocStringComment(t *testing.T) {
	input := `## this is a doc comment`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.DOCSTRING_COMMENT {
		t.Fatalf("expected DOCSTRING_COMMENT, got %v", tok.Type)
	}
	expectedLiteral := " this is a doc comment"
	if tok.Literal != expectedLiteral {
		t.Errorf("expected literal %q, got %q", expectedLiteral, tok.Literal)
	}
}

func TestNextTokenStringInterpolation(t *testing.T) {
	input := `"hello #{name}!"`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.STRING_DOUBLE_QUOTE {
		t.Fatalf("expected STRING_DOUBLE_QUOTE, got %v", tok.Type)
	}
}

func TestNextTokenBacktickExec(t *testing.T) {
	input := "`ls -la`"
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.BACKTICK {
		t.Fatalf("expected BACKTICK, got %v", tok.Type)
	}
	if tok.Literal != "ls -la" {
		t.Errorf("expected exec literal 'ls -la', got %q", tok.Literal)
	}
}

func TestNextTokenSpawnDefer(t *testing.T) {
	input := `spawn(foo) defer(bar)`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.SPAWN, "spawn"},
		{token.LPAREN, "("},
		{token.IDENT, "foo"},
		{token.RPAREN, ")"},
		{token.DEFER, "defer"},
		{token.LPAREN, "("},
		{token.IDENT, "bar"},
		{token.RPAREN, ")"},
		{token.EOF, ""},
	}

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

func TestNextTokenBreakContinue(t *testing.T) {
	input := `for (var i = 0; i < 10; i += 1) { break; continue; }`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.FOR, "for"},
		{token.LPAREN, "("},
		{token.VAR, "var"},
		{token.IDENT, "i"},
		{token.ASSIGN, "="},
		{token.INT, "0"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "i"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "i"},
		{token.PLUSEQ, "+="},
		{token.INT, "1"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.BREAK, "break"},
		{token.SEMICOLON, ";"},
		{token.CONTINUE, "continue"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}

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

func TestNextTokenNotIn(t *testing.T) {
	input := `x notin y`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "x"},
		{token.NOTIN, "notin"},
		{token.IDENT, "y"},
		{token.EOF, ""},
	}

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

func TestNextTokenNull(t *testing.T) {
	input := `val x = null`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.VAL, "val"},
		{token.IDENT, "x"},
		{token.ASSIGN, "="},
		{token.NULL_KW, "null"},
		{token.EOF, ""},
	}

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

func TestNextTokenSelf(t *testing.T) {
	input := `self()`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.SELF, "self"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.EOF, ""},
	}

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

func TestNextTokenIllegalChar(t *testing.T) {
	input := `@`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.ILLEGAL {
		t.Fatalf("expected ILLEGAL for @, got %v", tok.Type)
	}
}

func TestNextTokenLineNumberTracking(t *testing.T) {
	input := `line1
line2
line3`
	l := New(input, "test.b")

	tok := l.NextToken()
	if tok.Type != token.IDENT || tok.Literal != "line1" {
		t.Fatalf("expected IDENT line1, got %v %q", tok.Type, tok.Literal)
	}
	if tok.LineNumber != 0 {
		t.Errorf("expected line 0, got %d", tok.LineNumber)
	}

	tok = l.NextToken()
	if tok.Type != token.IDENT || tok.Literal != "line2" {
		t.Fatalf("expected IDENT line2, got %v %q", tok.Type, tok.Literal)
	}
	if tok.LineNumber != 1 {
		t.Errorf("expected line 1, got %d", tok.LineNumber)
	}

	tok = l.NextToken()
	if tok.Type != token.IDENT || tok.Literal != "line3" {
		t.Fatalf("expected IDENT line3, got %v %q", tok.Type, tok.Literal)
	}
	if tok.LineNumber != 2 {
		t.Errorf("expected line 2, got %d", tok.LineNumber)
	}
}

func TestNextTokenPositionTracking(t *testing.T) {
	input := `  hello`
	l := New(input, "test.b")
	tok := l.NextToken()
	if tok.Type != token.IDENT || tok.Literal != "hello" {
		t.Fatalf("expected IDENT hello, got %v %q", tok.Type, tok.Literal)
	}
	if tok.PositionInLine != 3 {
		t.Errorf("expected position 3, got %d", tok.PositionInLine)
	}
}

func TestNextTokenWhitespaceOnly(t *testing.T) {
	input := "   \t\n  "
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.EOF {
		t.Fatalf("expected EOF for whitespace-only input, got %v", tok.Type)
	}
}

func TestNextTokenEmptyInput(t *testing.T) {
	input := ``
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.EOF {
		t.Fatalf("expected EOF for empty input, got %v", tok.Type)
	}
}

func TestNextTokenComments(t *testing.T) {
	input := `## this is a doc comment
## another doc comment`
	l := New(input, "<internal: test>")

	// First ## starts docstring comment
	tok := l.NextToken()
	if tok.Type != token.DOCSTRING_COMMENT {
		t.Fatalf("expected DOCSTRING_COMMENT, got %v", tok.Type)
	}
	if tok.Literal != " this is a doc comment" {
		t.Errorf("expected ' this is a doc comment', got %q", tok.Literal)
	}

	// Second ## starts another docstring comment
	tok = l.NextToken()
	if tok.Type != token.DOCSTRING_COMMENT {
		t.Fatalf("expected second DOCSTRING_COMMENT, got %v", tok.Type)
	}
	if tok.Literal != " another doc comment" {
		t.Errorf("expected ' another doc comment', got %q", tok.Literal)
	}
}

func TestNextTokenSingleQuoteEscapes(t *testing.T) {
	input := `'hello\nworld'`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.STRING_SINGLE_QUOTE {
		t.Fatalf("expected STRING_SINGLE_QUOTE, got %v", tok.Type)
	}
	if tok.Literal != "hello\nworld" {
		t.Errorf("expected escaped newline in single quote, got %q", tok.Literal)
	}
}

func TestNextTokenHexWithUnderscores(t *testing.T) {
	input := `0xFF_FF_FF`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.HEX {
		t.Fatalf("expected HEX, got %v", tok.Type)
	}
}

func TestNextTokenOctalWithUnderscores(t *testing.T) {
	input := `0o77_77`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.OCTAL {
		t.Fatalf("expected OCTAL, got %v", tok.Type)
	}
}

func TestNextTokenBinaryWithUnderscores(t *testing.T) {
	input := `0b1010_1010`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.BINARY {
		t.Fatalf("expected BINARY, got %v", tok.Type)
	}
}

func TestNextTokenScientificNotation(t *testing.T) {
	input := `1e10`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.FLOAT {
		t.Fatalf("expected FLOAT for scientific notation, got %v", tok.Type)
	}
}

func TestNextTokenScientificNotationWithExp(t *testing.T) {
	input := `1.5e-10`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.FLOAT {
		t.Fatalf("expected FLOAT for scientific notation with sign, got %v", tok.Type)
	}
}

func TestNextTokenScientificNotationUpperExp(t *testing.T) {
	input := `1E+10`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.FLOAT {
		t.Fatalf("expected FLOAT for scientific notation uppercase E, got %v", tok.Type)
	}
}

func TestNextTokenDoubleDotRange(t *testing.T) {
	input := `1..10`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.INT, "1"},
		{token.RANGE, ".."},
		{token.INT, "10"},
		{token.EOF, ""},
	}

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

func TestNextTokenNonInclusiveRange(t *testing.T) {
	input := `1..<10`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.INT, "1"},
		{token.NONINCRANGE, "..<"},
		{token.INT, "10"},
		{token.EOF, ""},
	}

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

func TestNextTokenEllipsis(t *testing.T) {
	input := `func(args...)`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "func"},
		{token.LPAREN, "("},
		{token.IDENT, "args"},
		{token.ELLIPSE, "..."},
		{token.RPAREN, ")"},
		{token.EOF, ""},
	}

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

func TestNextTokenFloorDiv(t *testing.T) {
	input := `a // b`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "a"},
		{token.FDIV, "//"},
		{token.IDENT, "b"},
		{token.EOF, ""},
	}

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

func TestNextTokenFloorDivEq(t *testing.T) {
	input := `a //= 1`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "a"},
		{token.FDIVEQ, "//="},
		{token.INT, "1"},
		{token.EOF, ""},
	}

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

func TestNextTokenPowerEq(t *testing.T) {
	input := `a **= 2`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "a"},
		{token.POWEQ, "**="},
		{token.INT, "2"},
		{token.EOF, ""},
	}

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

func TestNextTokenRightArrow(t *testing.T) {
	input := `a => b`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "a"},
		{token.RARROW, "=>"},
		{token.IDENT, "b"},
		{token.EOF, ""},
	}

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

func TestNextTokenPower(t *testing.T) {
	input := `2 ** 3`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.INT, "2"},
		{token.POW, "**"},
		{token.INT, "3"},
		{token.EOF, ""},
	}

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

func TestNextTokenShiftOperators(t *testing.T) {
	input := `a << 1 >> 2`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "a"},
		{token.LSHIFT, "<<"},
		{token.INT, "1"},
		{token.RSHIFT, ">>"},
		{token.INT, "2"},
		{token.EOF, ""},
	}

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

func TestNextTokenShiftEqOperators(t *testing.T) {
	input := `a <<= 1 >>= 2`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "a"},
		{token.LSHIFTEQ, "<<="},
		{token.INT, "1"},
		{token.RSHIFTEQ, ">>="},
		{token.INT, "2"},
		{token.EOF, ""},
	}

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

func TestNextTokenLogicalAndOr(t *testing.T) {
	input := `a && b || c`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "a"},
		{token.AND, "&&"},
		{token.IDENT, "b"},
		{token.OR, "||"},
		{token.IDENT, "c"},
		{token.EOF, ""},
	}

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

func TestNextTokenLogicalAndOrEq(t *testing.T) {
	input := `a &&= b ||= c`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "a"},
		{token.ANDANDEQ, "&&="},
		{token.IDENT, "b"},
		{token.OROREQ, "||="},
		{token.IDENT, "c"},
		{token.EOF, ""},
	}

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

func TestNextTokenBitwiseAssignments(t *testing.T) {
	input := `a &= b |= c ~= d ^= e`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "a"},
		{token.ANDEQ, "&="},
		{token.IDENT, "b"},
		{token.OREQ, "|="},
		{token.IDENT, "c"},
		{token.BINNOTEQ, "~="},
		{token.IDENT, "d"},
		{token.XOREQ, "^="},
		{token.IDENT, "e"},
		{token.EOF, ""},
	}

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

func TestNextTokenCompoundAssignments(t *testing.T) {
	input := `a += b -= c *= d /= e %= f`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "a"},
		{token.PLUSEQ, "+="},
		{token.IDENT, "b"},
		{token.MINUSEQ, "-="},
		{token.IDENT, "c"},
		{token.MULEQ, "*="},
		{token.IDENT, "d"},
		{token.DIVEQ, "/="},
		{token.IDENT, "e"},
		{token.PERCENTEQ, "%="},
		{token.IDENT, "f"},
		{token.EOF, ""},
	}

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

func TestNextTokenIdentifierWithSpecialChars(t *testing.T) {
	input := `foo? bar!`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "foo?"},
		{token.IDENT, "bar!"},
		{token.EOF, ""},
	}

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

func TestNextTokenStructLiteral(t *testing.T) {
	input := `@{name: "alice"}`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.ATLBRACE, "@{"},
		{token.IDENT, "name"},
		{token.COLON, ":"},
		{token.STRING_DOUBLE_QUOTE, "alice"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}

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

func TestNextTokenMapLiteral(t *testing.T) {
	input := `{"key": "value"}`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.LBRACE, "{"},
		{token.STRING_DOUBLE_QUOTE, "key"},
		{token.COLON, ":"},
		{token.STRING_DOUBLE_QUOTE, "value"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}

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

func TestNextTokenListLiteral(t *testing.T) {
	input := `[1, 2, 3]`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.INT, "2"},
		{token.COMMA, ","},
		{token.INT, "3"},
		{token.RBRACKET, "]"},
		{token.EOF, ""},
	}

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

func TestNextTokenPipe(t *testing.T) {
	input := `a | b`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "a"},
		{token.PIPE, "|"},
		{token.IDENT, "b"},
		{token.EOF, ""},
	}

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

func TestNextTokenHash(t *testing.T) {
	input := `# not a comment`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.HASH {
		t.Fatalf("expected HASH, got %v", tok.Type)
	}
}

func TestNextTokenColon(t *testing.T) {
	input := `key: value`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "key"},
		{token.COLON, ":"},
		{token.IDENT, "value"},
		{token.EOF, ""},
	}

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

func TestNextTokenMatchExpression(t *testing.T) {
	input := `match x { true => { 1 } false => { 0 } }`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.MATCH, "match"},
		{token.IDENT, "x"},
		{token.LBRACE, "{"},
		{token.TRUE, "true"},
		{token.RARROW, "=>"},
		{token.LBRACE, "{"},
		{token.INT, "1"},
		{token.RBRACE, "}"},
		{token.FALSE, "false"},
		{token.RARROW, "=>"},
		{token.LBRACE, "{"},
		{token.INT, "0"},
		{token.RBRACE, "}"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}

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

func TestNextTokenTryCatchFinally(t *testing.T) {
	input := `try { 1 } catch (e) { 2 } finally { 3 }`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.TRY, "try"},
		{token.LBRACE, "{"},
		{token.INT, "1"},
		{token.RBRACE, "}"},
		{token.CATCH, "catch"},
		{token.LPAREN, "("},
		{token.IDENT, "e"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.INT, "2"},
		{token.RBRACE, "}"},
		{token.FINALLY, "finally"},
		{token.LBRACE, "{"},
		{token.INT, "3"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}

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

func TestNextTokenForLoopWithVar(t *testing.T) {
	input := `for (var i = 0; i < 10; i += 1) { x }`
	l := New(input, "<internal: test>")

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.FOR, "for"},
		{token.LPAREN, "("},
		{token.VAR, "var"},
		{token.IDENT, "i"},
		{token.ASSIGN, "="},
		{token.INT, "0"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "i"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "i"},
		{token.PLUSEQ, "+="},
		{token.INT, "1"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}

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

func TestNextTokenRawString(t *testing.T) {
	input := `"""hello world"""`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.RAW_STRING {
		t.Fatalf("expected RAW_STRING, got %v", tok.Type)
	}
	if tok.Literal != "hello world" {
		t.Errorf("expected 'hello world', got %q", tok.Literal)
	}
}

func TestNextTokenRawStringWithInterpolation(t *testing.T) {
	input := `"""hello #{name}!"""`
	l := New(input, "<internal: test>")
	tok := l.NextToken()
	if tok.Type != token.RAW_STRING {
		t.Fatalf("expected RAW_STRING, got %v", tok.Type)
	}
	if tok.Literal != "hello #{name}!" {
		t.Errorf("expected 'hello #{name}!', got %q", tok.Literal)
	}
}

func TestNextTokenMultipleOnNewlines(t *testing.T) {
	input := `a
b
c`
	l := New(input, "test.b")

	tests := []string{"a", "b", "c"}
	for i, expected := range tests {
		tok := l.NextToken()
		if tok.Type != token.IDENT {
			t.Fatalf("test[%d] - expected IDENT, got %v", i, tok.Type)
		}
		if tok.Literal != expected {
			t.Fatalf("test[%d] - expected %q, got %q", i, expected, tok.Literal)
		}
		if tok.LineNumber != i {
			t.Fatalf("test[%d] - expected line %d, got %d", i, i, tok.LineNumber)
		}
	}
}
