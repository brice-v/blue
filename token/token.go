// tokenizer

package token

import "fmt"

// Type is the string representation of the Token
type Type string

// Token is the struct containing the TokenType Type, and the
// Literal value as a string
type Token struct {
	Type    Type
	Literal string

	Filepath       string
	LineNumber     int
	PositionInLine int
}

func (t Token) DisplayForErrorLine() string {
	return fmt.Sprintf("Filepath: %q, LineNumber: %d, PositionInLine: %d", t.Filepath, t.LineNumber, t.PositionInLine)
}

const (
	// ILLEGAL is the string rep. of an illegal tok.
	ILLEGAL = "ILLEGAL"
	// EOF is the string rep. of end of file
	EOF = "EOF"

	// MULTLINE_COMMENT is the string rep. of a multiline comment token
	MULTLINE_COMMENT = "###"
	// DOCSTRING_COMMENT is the string rep. of a doc comment token
	DOCSTRING_COMMENT = "##"

	// Identifiers and literals

	// IDENT is the string rep. of an identifier
	IDENT = "IDENT"
	// INT is the string rep. of an integer tok.
	INT = "INT"
	// FLOAT is the string rep. of a float tok.
	FLOAT = "FLOAT"
	// HEX is the string rep. of a hex tok.
	HEX = "HEX"
	// OCTAL is the string rep. of an octal tok.
	OCTAL = "OCTAL"
	// BINARY is the string rep. of a binary tok.
	BINARY = "BINARY"
	// UINT is the string rep. of a uinteger tok.
	UINT = "UINT"
	// BIGINT is the string rep. of a big int tok.
	BIGINT = "BIGINT"
	// BIGFLOAT is the string rep. of a big float tok.
	BIGFLOAT = "BIGFLOAT"
	// STRING_DOUBLE_QUOTE is the string rep. of a string literal tok. with "
	STRING_DOUBLE_QUOTE = "STRING_DOUBLE_QUOTE"
	// STRING_SINGLE_QUOTE is the string rep. of a string literal tok. with '
	STRING_SINGLE_QUOTE = "STRING_SINGLE_QUOTE"

	// Operators

	// ASSIGN is the string rep. of an assignment tok.
	ASSIGN = "="
	// PERCENTEQ is the string rep. of the modulo equal tok.
	PERCENTEQ = "%="
	// LTEQ is the string rep. of the less than equal tok.
	LTEQ = "<="
	// GTEQ is the string rep. of the greater than equal tok.
	GTEQ = ">="
	// RARROW is the string rep. of the right arrow tok.
	RARROW = "=>"

	// RAW_STRING is the string rep. of the raw string token
	RAW_STRING = `"""`
	// STRINGINTERP is the string interpolation token
	STRINGINTERP = "#{"

	// REGEX is the string rep. of the regex literal start token
	REGEX = "r/"

	// NOTE: ANDEQ, OREQ, BINNOTEQ, and XOREQ might also be used for sets and other data types eventually

	// ANDANDEQ is the string rep. of the boolean and equal tok.
	ANDANDEQ = "&&="
	// OROREQ is the string rep. of the boolean or equal tok.
	OROREQ = "||="
	// ANDEQ is the string rep. of the binary and equal tok.
	ANDEQ = "&="
	// OREQ is the string rep. of the binary or equal tok.
	OREQ = "|="
	// BINNOTEQ is the string rep. of the binary not equal tok.
	BINNOTEQ = "~="
	// XOREQ is the string rep. of the binary xor equal tok.
	XOREQ = "^="

	// MULEQ is the string rep. of the mulitply equal tok.
	MULEQ = "*="
	// PLUSEQ is the string rep. of the plus equal tok.
	PLUSEQ = "+="
	// MINUSEQ is the string rep. of the minus equal tok.
	MINUSEQ = "-="
	// DIVEQ is the string rep. of the div equal token.
	DIVEQ = "/="

	// PLUS is the string rep. of a plus tok.
	PLUS = "+"

	// BANG is the string rep. of a ! tok.
	BANG = "!"
	// STAR is the string rep. of a * tok.
	STAR = "*"
	// FSLASH is the string rep. of a forward slash tok.
	FSLASH = "/"
	// MINUS is the string rep. of a minus tok.
	MINUS = "-"

	// TILDE is the string rep. of a bitwise not tok.
	// this may not be used, but good for the lexer to know
	TILDE = "~"

	// AMPERSAND is the string rep. of an ampersand tok.
	AMPERSAND = "&"
	// HAT is the string rep. of a hat tok.
	HAT = "^"
	// HASH is the string rep. of a number sign tok.
	HASH = "#"
	// PERCENT is the string rep. of a percent tok.
	PERCENT = "%"
	// DOT is the string rep. of a period tok.
	DOT = "."

	// LT is the string rep. of a less than tok.
	LT = "<"
	// GT is the string rep. of a greater than tok.
	GT = ">"

	// EQ is the string rep. of an equal tok.
	EQ = "=="
	// NEQ is the string rep. of a not equal tok.
	NEQ = "!="

	// POW is the string rep. of a power tok. ie. (2 ** 3 == 8)
	POW = "**"
	// RANGE is the string rep. of a range tok.
	RANGE = ".."
	// FDIV is the string rep. of a floor division tok.
	FDIV = "//"
	// RSHIFT is the string rep. of a right shift tok.
	RSHIFT = ">>"
	// LSHIFT is the string rep. of a left shift tok.
	LSHIFT = "<<"

	// Delimeters

	// COMMA is the string rep. of a comma tok.
	COMMA = ","
	// SEMICOLON is the string rep. of a comma tok.
	SEMICOLON = ";"
	// COLON is the string rep. of a colon tok.
	COLON = ":"

	// BACKTICK is the string rep. of a backtick tok.
	BACKTICK = "`"

	// LPAREN is the string rep. of a left paren. tok.
	LPAREN = "("
	// RPAREN is the string rep. of a right paren. tok.
	RPAREN = ")"
	// LBRACE is the string rep. of a left brace tok.
	LBRACE = "{"
	// RBRACE is the string rep. of a right brace tok.
	RBRACE = "}"
	// LBRACKET is the string rep. of a left bracket tok.
	LBRACKET = "["
	// RBRACKET is the string rep. of a right bracket tok.
	RBRACKET = "]"
	// PIPE is the string rep. of the pipe tok.
	PIPE = "|"

	// POWEQ is the string rep. of the pow equal tok.
	POWEQ = "**="
	// ELLIPSE is the string rep. of the ellises tok.
	ELLIPSE = "..."
	// FDIVEQ is the string rep. of the floor div equal tok.
	FDIVEQ = "//="
	// RSHIFTEQ is the string rep. of the right shift equal tok.
	RSHIFTEQ = ">>="
	// LSHIFTEQ is the string rep. of the left shift equal tok.
	LSHIFTEQ = "<<="
	// NONINCRANGE is the string rep. of the non inclusive range token
	NONINCRANGE = "..<"

	// Keywords

	// FUNCTION is the string rep. of a function tok.
	FUNCTION = "FUNCTION"
	// VAR is the string rep. of a var tok.
	VAR = "VAR"
	// VAL is the string rep. of a val tok.
	VAL = "VAL"
	// TRUE is the string rep. of the `true` tok.
	TRUE = "TRUE"
	// FALSE is the string rep. of the `false` tok.
	FALSE = "FALSE"
	// IF is the string rep. of the `if` tok.
	IF = "IF"
	// ELSE is the string rep. of the `else` tok.
	ELSE = "ELSE"
	// RETURN is the string rep of the `return` tok.
	RETURN = "RETURN"
	// FOR is the string rep. of the `for` tok.
	FOR = "FOR"
	// IN is the string rep. of the `in` tok.
	IN = "IN"
	// NOTIN is the string rep. of the `notin` tok.
	NOTIN = "NOTIN"
	// AND is the string rep. of the `and` tok.
	AND = "AND"
	// OR is the string rep. of the `or` tok.
	OR = "OR"
	// NOT is the string rep. of the `not` tok.
	NOT = "NOT"
	// CONST is the string rep. of the `const` tok.
	CONST = "CONST"
	// MATCH is the string rep. of the `match` tok.
	MATCH = "MATCH"
	// NULL_KW is the string rep. of the `null` tok
	NULL_KW = "NULL_KW"

	// IMPORT is the string rep. of the import tok
	IMPORT = "IMPORT"
	// IMPORT_PATH is the string rep. of the import path tok
	IMPORT_PATH = "IMPORT_PATH"
	// FROM is the string rep. of the from import tok
	FROM = "FROM"
	// AS is the string rep. of the as tok
	AS = "AS"

	// TRY is the string rep. of the 'try' keyword token
	TRY = "TRY"
	// CATCH is the string rep. of the 'catch' keyword token
	CATCH = "CATCH"
	// FINALLY is the string rep. of the 'finally' keyword token
	FINALLY = "FINALLY"

	// EVAL is the string rep. of the 'eval' keyword token
	EVAL = "EVAL"
	// SPAWN is the string rep. of the 'spawn' keyword token
	SPAWN = "SPAWN"
	// DEFER is the string rep. of the 'defer' keyword token
	DEFER = "DEFER"
	// SELF is the string rep. of the 'self' keyword token
	SELF = "SELF"

	// BREAK is the string rep. of the 'break' keyword token
	BREAK = "BREAK"
	// CONTINUE is the string rep. of the 'continue' keyword token
	CONTINUE = "CONTINUE"
)

var keywords = map[string]Type{
	"fun":      FUNCTION,
	"var":      VAR,
	"val":      VAL,
	"true":     TRUE,
	"false":    FALSE,
	"if":       IF,
	"else":     ELSE,
	"return":   RETURN,
	"for":      FOR,
	"in":       IN,
	"notin":    NOTIN,
	"and":      AND,
	"or":       OR,
	"not":      NOT,
	"const":    CONST,
	"match":    MATCH,
	"null":     NULL_KW,
	"import":   IMPORT,
	"from":     FROM,
	"as":       AS,
	"try":      TRY,
	"catch":    CATCH,
	"finally":  FINALLY,
	"eval":     EVAL,
	"spawn":    SPAWN,
	"defer":    DEFER,
	"self":     SELF,
	"break":    BREAK,
	"continue": CONTINUE,
}

// LookupIdent will check if the identifer passed in matches one of the
// keywords and if so will return that keyword token type, otherwise
// it will return the IDENT token type
func LookupIdent(ident string) Type {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
