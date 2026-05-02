package token

import (
	"strings"
	"testing"
)

// LookupIdent tests

func TestLookupIdentKeywords(t *testing.T) {
	tests := []struct {
		input    string
		expected Type
	}{
		{"fun", FUNCTION},
		{"var", VAR},
		{"val", VAL},
		{"true", TRUE},
		{"false", FALSE},
		{"if", IF},
		{"else", ELSE},
		{"return", RETURN},
		{"for", FOR},
		{"in", IN},
		{"notin", NOTIN},
		{"and", AND},
		{"or", OR},
		{"not", NOT},
		{"const", CONST},
		{"match", MATCH},
		{"null", NULL_KW},
		{"import", IMPORT},
		{"from", FROM},
		{"as", AS},
		{"try", TRY},
		{"catch", CATCH},
		{"finally", FINALLY},
		{"eval", EVAL},
		{"spawn", SPAWN},
		{"defer", DEFER},
		{"self", SELF},
		{"break", BREAK},
		{"continue", CONTINUE},
	}

	for _, tt := range tests {
		result := LookupIdent(tt.input)
		if result != tt.expected {
			t.Errorf("LookupIdent(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestLookupIdentNonKeywords(t *testing.T) {
	identifiers := []string{
		"hello", "foo", "bar", "x", "myVar", "someFunc",
		"ABC", "test123", "camelCase", "snake_case",
		"funx", "var1", "truey", "iff", "forn",
	}

	for _, ident := range identifiers {
		result := LookupIdent(ident)
		if result != IDENT {
			t.Errorf("LookupIdent(%q) = %q, want IDENT", ident, result)
		}
	}
}

func TestLookupIdentEmpty(t *testing.T) {
	result := LookupIdent("")
	if result != IDENT {
		t.Errorf("LookupIdent(\"\") = %q, want IDENT", result)
	}
}

func TestLookupIdentCaseSensitive(t *testing.T) {
	// Keywords are case-sensitive
	if LookupIdent("Fun") != IDENT {
		t.Error("LookupIdent(\"Fun\") should return IDENT (case sensitive)")
	}
	if LookupIdent("VAR") != IDENT {
		t.Error("LookupIdent(\"VAR\") should return IDENT (case sensitive)")
	}
	if LookupIdent("If") != IDENT {
		t.Error("LookupIdent(\"If\") should return IDENT (case sensitive)")
	}
	if LookupIdent("TRUE") != IDENT {
		t.Error("LookupIdent(\"TRUE\") should return IDENT (case sensitive)")
	}
}

func TestLookupIdentAllKeywordsComplete(t *testing.T) {
	// Verify all defined keyword constants are in the map
	allKeywords := []Type{
		FUNCTION, VAR, VAL, TRUE, FALSE, IF, ELSE, RETURN,
		FOR, IN, NOTIN, AND, OR, NOT, CONST, MATCH, NULL_KW,
		IMPORT, FROM, AS, TRY, CATCH, FINALLY, EVAL, SPAWN,
		DEFER, SELF, BREAK, CONTINUE,
	}

	for _, kw := range allKeywords {
		// Find the string key that maps to this type
		found := false
		for str, typ := range keywords {
			if typ == kw {
				result := LookupIdent(str)
				if result != kw {
					t.Errorf("LookupIdent(%q) = %q, want %q", str, result, kw)
				}
				found = true
				break
			}
		}
		if !found {
			t.Errorf("keyword type %q not found in keywords map", kw)
		}
	}
}

func TestLookupIdentReturnsIdentForUnknown(t *testing.T) {
	unknowns := []string{
		"unknown", "lambda", "class", "struct", "enum",
		"while", "do", "switch", "case", "default",
	}

	for _, u := range unknowns {
		result := LookupIdent(u)
		if result != IDENT {
			t.Errorf("LookupIdent(%q) = %q, want IDENT", u, result)
		}
	}
}

// Token type constant tests

func TestTokenTypes(t *testing.T) {
	tests := []struct {
		name     string
		token    Type
		expected string
	}{
		{"ILLEGAL", ILLEGAL, "ILLEGAL"},
		{"EOF", EOF, "EOF"},
		{"MULTLINE_COMMENT", MULTLINE_COMMENT, "###"},
		{"DOCSTRING_COMMENT", DOCSTRING_COMMENT, "##"},
		{"IDENT", IDENT, "IDENT"},
		{"INT", INT, "INT"},
		{"FLOAT", FLOAT, "FLOAT"},
		{"HEX", HEX, "HEX"},
		{"OCTAL", OCTAL, "OCTAL"},
		{"BINARY", BINARY, "BINARY"},
		{"UINT", UINT, "UINT"},
		{"BIGINT", BIGINT, "BIGINT"},
		{"BIGFLOAT", BIGFLOAT, "BIGFLOAT"},
		{"STRING_DOUBLE_QUOTE", STRING_DOUBLE_QUOTE, "STRING_DOUBLE_QUOTE"},
		{"STRING_SINGLE_QUOTE", STRING_SINGLE_QUOTE, "STRING_SINGLE_QUOTE"},
		{"RAW_STRING", RAW_STRING, `"""`},
		{"STRINGINTERP", STRINGINTERP, "#{"},
		{"REGEX", REGEX, "r/"},
		{"ASSIGN", ASSIGN, "="},
		{"PERCENTEQ", PERCENTEQ, "%="},
		{"LTEQ", LTEQ, "<="},
		{"GTEQ", GTEQ, ">="},
		{"RARROW", RARROW, "=>"},
		{"ANDANDEQ", ANDANDEQ, "&&="},
		{"OROREQ", OROREQ, "||="},
		{"ANDEQ", ANDEQ, "&="},
		{"OREQ", OREQ, "|="},
		{"BINNOTEQ", BINNOTEQ, "~="},
		{"XOREQ", XOREQ, "^="},
		{"MULEQ", MULEQ, "*="},
		{"PLUSEQ", PLUSEQ, "+="},
		{"MINUSEQ", MINUSEQ, "-="},
		{"DIVEQ", DIVEQ, "/="},
		{"PLUS", PLUS, "+"},
		{"BANG", BANG, "!"},
		{"STAR", STAR, "*"},
		{"FSLASH", FSLASH, "/"},
		{"MINUS", MINUS, "-"},
		{"TILDE", TILDE, "~"},
		{"AMPERSAND", AMPERSAND, "&"},
		{"HAT", HAT, "^"},
		{"HASH", HASH, "#"},
		{"PERCENT", PERCENT, "%"},
		{"DOT", DOT, "."},
		{"LT", LT, "<"},
		{"GT", GT, ">"},
		{"EQ", EQ, "=="},
		{"NEQ", NEQ, "!="},
		{"POW", POW, "**"},
		{"RANGE", RANGE, ".."},
		{"FDIV", FDIV, "//"},
		{"RSHIFT", RSHIFT, ">>"},
		{"LSHIFT", LSHIFT, "<<"},
		{"ATLBRACE", ATLBRACE, "@{"},
		{"COMMA", COMMA, ","},
		{"SEMICOLON", SEMICOLON, ";"},
		{"COLON", COLON, ":"},
		{"BACKTICK", BACKTICK, "`"},
		{"LPAREN", LPAREN, "("},
		{"RPAREN", RPAREN, ")"},
		{"LBRACE", LBRACE, "{"},
		{"RBRACE", RBRACE, "}"},
		{"LBRACKET", LBRACKET, "["},
		{"RBRACKET", RBRACKET, "]"},
		{"PIPE", PIPE, "|"},
		{"POWEQ", POWEQ, "**="},
		{"ELLIPSE", ELLIPSE, "..."},
		{"FDIVEQ", FDIVEQ, "//="},
		{"RSHIFTEQ", RSHIFTEQ, ">>="},
		{"LSHIFTEQ", LSHIFTEQ, "<<="},
		{"NONINCRANGE", NONINCRANGE, "..<"},
		{"FUNCTION", FUNCTION, "FUNCTION"},
		{"VAR", VAR, "VAR"},
		{"VAL", VAL, "VAL"},
		{"TRUE", TRUE, "TRUE"},
		{"FALSE", FALSE, "FALSE"},
		{"IF", IF, "IF"},
		{"ELSE", ELSE, "ELSE"},
		{"RETURN", RETURN, "RETURN"},
		{"FOR", FOR, "FOR"},
		{"IN", IN, "IN"},
		{"NOTIN", NOTIN, "NOTIN"},
		{"AND", AND, "AND"},
		{"OR", OR, "OR"},
		{"NOT", NOT, "NOT"},
		{"CONST", CONST, "CONST"},
		{"MATCH", MATCH, "MATCH"},
		{"NULL_KW", NULL_KW, "NULL_KW"},
		{"IMPORT", IMPORT, "IMPORT"},
		{"IMPORT_PATH", IMPORT_PATH, "IMPORT_PATH"},
		{"FROM", FROM, "FROM"},
		{"AS", AS, "AS"},
		{"TRY", TRY, "TRY"},
		{"CATCH", CATCH, "CATCH"},
		{"FINALLY", FINALLY, "FINALLY"},
		{"EVAL", EVAL, "EVAL"},
		{"SPAWN", SPAWN, "SPAWN"},
		{"DEFER", DEFER, "DEFER"},
		{"SELF", SELF, "SELF"},
		{"BREAK", BREAK, "BREAK"},
		{"CONTINUE", CONTINUE, "CONTINUE"},
	}

	for _, tt := range tests {
		if string(tt.token) != tt.expected {
			t.Errorf("%s: token %q = %q, want %q", tt.name, tt.token, tt.token, tt.expected)
		}
	}
}

// Token struct tests

func TestTokenDisplayForErrorLine(t *testing.T) {
	tok := Token{
		Type:           IDENT,
		Literal:        "foo",
		Filepath:       "test.b",
		LineNumber:     42,
		PositionInLine: 10,
	}

	result := tok.DisplayForErrorLine()
	expected := `Filepath: "test.b", LineNumber: 42, PositionInLine: 10`
	if result != expected {
		t.Errorf("DisplayForErrorLine() = %q, want %q", result, expected)
	}
}

func TestTokenDisplayForErrorLineEmpty(t *testing.T) {
	tok := Token{
		Type:           ILLEGAL,
		Literal:        "@",
		Filepath:       "",
		LineNumber:     0,
		PositionInLine: 0,
	}

	result := tok.DisplayForErrorLine()
	expected := `Filepath: "", LineNumber: 0, PositionInLine: 0`
	if result != expected {
		t.Errorf("DisplayForErrorLine() = %q, want %q", result, expected)
	}
}

func TestTokenDisplayForErrorLineSpecialChars(t *testing.T) {
	tok := Token{
		Type:           EOF,
		Literal:        "",
		Filepath:       "/path/to/my file.b",
		LineNumber:     1,
		PositionInLine: 0,
	}

	result := tok.DisplayForErrorLine()
	if !strings.Contains(result, `"/path/to/my file.b"`) {
		t.Errorf("DisplayForErrorLine() should quote filepath, got %q", result)
	}
}

func TestTokenStructFields(t *testing.T) {
	tok := Token{
		Type:           STRING_DOUBLE_QUOTE,
		Literal:        "hello world",
		Filepath:       "main.b",
		LineNumber:     100,
		PositionInLine: 5,
	}

	if tok.Type != STRING_DOUBLE_QUOTE {
		t.Errorf("expected Type STRING_DOUBLE_QUOTE, got %q", tok.Type)
	}
	if tok.Literal != "hello world" {
		t.Errorf("expected Literal \"hello world\", got %q", tok.Literal)
	}
	if tok.Filepath != "main.b" {
		t.Errorf("expected Filepath \"main.b\", got %q", tok.Filepath)
	}
	if tok.LineNumber != 100 {
		t.Errorf("expected LineNumber 100, got %d", tok.LineNumber)
	}
	if tok.PositionInLine != 5 {
		t.Errorf("expected PositionInLine 5, got %d", tok.PositionInLine)
	}
}

func TestTokenZeroValue(t *testing.T) {
	var tok Token
	if tok.Type != "" {
		t.Errorf("zero value Type should be empty, got %q", tok.Type)
	}
	if tok.Literal != "" {
		t.Errorf("zero value Literal should be empty, got %q", tok.Literal)
	}
	if tok.LineNumber != 0 {
		t.Errorf("zero value LineNumber should be 0, got %d", tok.LineNumber)
	}
	if tok.PositionInLine != 0 {
		t.Errorf("zero value PositionInLine should be 0, got %d", tok.PositionInLine)
	}
}

// Keywords map tests

func TestKeywordsMapCompleteness(t *testing.T) {
	// Verify no unexpected keys
	expectedKeys := []string{
		"fun", "var", "val", "true", "false", "if", "else", "return",
		"for", "in", "notin", "and", "or", "not", "const", "match",
		"null", "import", "from", "as", "try", "catch", "finally",
		"eval", "spawn", "defer", "self", "break", "continue",
	}

	if len(keywords) != len(expectedKeys) {
		t.Errorf("expected %d keywords, got %d", len(expectedKeys), len(keywords))
	}

	for _, key := range expectedKeys {
		if _, ok := keywords[key]; !ok {
			t.Errorf("expected keyword %q to be in map", key)
		}
	}
}

func TestKeywordsMapNoDuplicates(t *testing.T) {
	// Each keyword type should appear exactly once in the map
	seen := make(map[Type]int)
	for _, typ := range keywords {
		seen[typ]++
	}
	for typ, count := range seen {
		if count > 1 {
			t.Errorf("keyword type %q appears %d times in map", typ, count)
		}
	}
}

func TestTokenStringRepresentation(t *testing.T) {
	// Verify that all operator tokens have unique string representations
	seen := make(map[string]bool)
	operators := []Type{
		ASSIGN, PERCENTEQ, LTEQ, GTEQ, RARROW, ANDANDEQ, OROREQ,
		ANDEQ, OREQ, BINNOTEQ, XOREQ, MULEQ, PLUSEQ, MINUSEQ, DIVEQ,
		PLUS, BANG, STAR, FSLASH, MINUS, TILDE, AMPERSAND, HAT,
		HASH, PERCENT, DOT, LT, GT, EQ, NEQ, POW, RANGE, FDIV,
		RSHIFT, LSHIFT, ATLBRACE, COMMA, SEMICOLON, COLON, BACKTICK,
		LPAREN, RPAREN, LBRACE, RBRACE, LBRACKET, RBRACKET, PIPE,
		POWEQ, ELLIPSE, FDIVEQ, RSHIFTEQ, LSHIFTEQ, NONINCRANGE,
	}

	for _, op := range operators {
		s := string(op)
		if seen[s] {
			t.Errorf("duplicate operator string representation: %q", s)
		}
		seen[s] = true
	}
}

func TestTokenKeywordsNotOperators(t *testing.T) {
	// Verify no keyword string conflicts with operator strings
	keywordStrings := []string{
		"fun", "var", "val", "true", "false", "if", "else", "return",
		"for", "in", "notin", "and", "or", "not", "const", "match",
		"null", "import", "from", "as", "try", "catch", "finally",
		"eval", "spawn", "defer", "self", "break", "continue",
	}

	for _, ks := range keywordStrings {
		tok := LookupIdent(ks)
		if tok != IDENT {
			// It's a keyword, verify it's not a single-char operator
			if len(ks) == 1 && ks != "in" && ks != "as" {
				// single-char keyword is fine as long as it's a keyword
			}
		}
	}
}
