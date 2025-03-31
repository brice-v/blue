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
	ILLEGAL Type = "ILLEGAL"
	// EOF is the string rep. of end of file
	EOF Type = "EOF"

	// MULTLINE_COMMENT is the string rep. of a multiline comment token
	MULTLINE_COMMENT Type = "###"
	// DOCSTRING_COMMENT is the string rep. of a doc comment token
	DOCSTRING_COMMENT Type = "##"

	// Identifiers and literals

	// IDENT is the string rep. of an identifier
	IDENT Type = "IDENT"
	// INT is the string rep. of an integer tok.
	INT Type = "INT"
	// FLOAT is the string rep. of a float tok.
	FLOAT Type = "FLOAT"
	// HEX is the string rep. of a hex tok.
	HEX Type = "HEX"
	// OCTAL is the string rep. of an octal tok.
	OCTAL Type = "OCTAL"
	// BINARY is the string rep. of a binary tok.
	BINARY Type = "BINARY"
	// UINT is the string rep. of a uinteger tok.
	UINT Type = "UINT"
	// BIGINT is the string rep. of a big int tok.
	BIGINT Type = "BIGINT"
	// BIGFLOAT is the string rep. of a big float tok.
	BIGFLOAT Type = "BIGFLOAT"
	// STRING_DOUBLE_QUOTE is the string rep. of a string literal tok. with "
	STRING_DOUBLE_QUOTE Type = "STRING_DOUBLE_QUOTE"
	// STRING_SINGLE_QUOTE is the string rep. of a string literal tok. with '
	STRING_SINGLE_QUOTE Type = "STRING_SINGLE_QUOTE"

	// Operators

	// ASSIGN is the string rep. of an assignment tok.
	ASSIGN Type = "="
	// PERCENTEQ is the string rep. of the modulo equal tok.
	PERCENTEQ Type = "%="
	// LTEQ is the string rep. of the less than equal tok.
	LTEQ Type = "<="
	// GTEQ is the string rep. of the greater than equal tok.
	GTEQ Type = ">="
	// RARROW is the string rep. of the right arrow tok.
	RARROW Type = "=>"

	// RAW_STRING is the string rep. of the raw string token
	RAW_STRING Type = `"""`
	// STRINGINTERP is the string interpolation token
	STRINGINTERP Type = "#{"

	// REGEX is the string rep. of the regex literal start token
	REGEX Type = "r/"

	// NOTE: ANDEQ, OREQ, BINNOTEQ, and XOREQ might also be used for sets and other data types eventually

	// ANDANDEQ is the string rep. of the boolean and equal tok.
	ANDANDEQ Type = "&&="
	// OROREQ is the string rep. of the boolean or equal tok.
	OROREQ Type = "||="
	// ANDEQ is the string rep. of the binary and equal tok.
	ANDEQ Type = "&="
	// OREQ is the string rep. of the binary or equal tok.
	OREQ Type = "|="
	// BINNOTEQ is the string rep. of the binary not equal tok.
	BINNOTEQ Type = "~="
	// XOREQ is the string rep. of the binary xor equal tok.
	XOREQ Type = "^="

	// MULEQ is the string rep. of the mulitply equal tok.
	MULEQ Type = "*="
	// PLUSEQ is the string rep. of the plus equal tok.
	PLUSEQ Type = "+="
	// MINUSEQ is the string rep. of the minus equal tok.
	MINUSEQ Type = "-="
	// DIVEQ is the string rep. of the div equal token.
	DIVEQ Type = "/="

	// PLUS is the string rep. of a plus tok.
	PLUS Type = "+"

	// BANG is the string rep. of a ! tok.
	BANG Type = "!"
	// STAR is the string rep. of a * tok.
	STAR Type = "*"
	// FSLASH is the string rep. of a forward slash tok.
	FSLASH Type = "/"
	// MINUS is the string rep. of a minus tok.
	MINUS Type = "-"

	// TILDE is the string rep. of a bitwise not tok.
	// this may not be used, but good for the lexer to know
	TILDE Type = "~"

	// AMPERSAND is the string rep. of an ampersand tok.
	AMPERSAND Type = "&"
	// HAT is the string rep. of a hat tok.
	HAT Type = "^"
	// HASH is the string rep. of a number sign tok.
	HASH Type = "#"
	// PERCENT is the string rep. of a percent tok.
	PERCENT Type = "%"
	// DOT is the string rep. of a period tok.
	DOT Type = "."

	// LT is the string rep. of a less than tok.
	LT Type = "<"
	// GT is the string rep. of a greater than tok.
	GT Type = ">"

	// EQ is the string rep. of an equal tok.
	EQ Type = "=="
	// NEQ is the string rep. of a not equal tok.
	NEQ Type = "!="

	// POW is the string rep. of a power tok. ie. (2 ** 3 == 8)
	POW Type = "**"
	// RANGE is the string rep. of a range tok.
	RANGE Type = ".."
	// FDIV is the string rep. of a floor division tok.
	FDIV Type = "//"
	// RSHIFT is the string rep. of a right shift tok.
	RSHIFT Type = ">>"
	// LSHIFT is the string rep. of a left shift tok.
	LSHIFT Type = "<<"

	// ATLBRACE is the string rep. of the @{ token used for struct literals
	ATLBRACE Type = "@{"

	// Delimeters

	// COMMA is the string rep. of a comma tok.
	COMMA Type = ","
	// SEMICOLON is the string rep. of a comma tok.
	SEMICOLON Type = ";"
	// COLON is the string rep. of a colon tok.
	COLON Type = ":"

	// BACKTICK is the string rep. of a backtick tok.
	BACKTICK Type = "`"

	// LPAREN is the string rep. of a left paren. tok.
	LPAREN Type = "("
	// RPAREN is the string rep. of a right paren. tok.
	RPAREN Type = ")"
	// LBRACE is the string rep. of a left brace tok.
	LBRACE Type = "{"
	// RBRACE is the string rep. of a right brace tok.
	RBRACE Type = "}"
	// LBRACKET is the string rep. of a left bracket tok.
	LBRACKET Type = "["
	// RBRACKET is the string rep. of a right bracket tok.
	RBRACKET Type = "]"
	// PIPE is the string rep. of the pipe tok.
	PIPE Type = "|"

	// POWEQ is the string rep. of the pow equal tok.
	POWEQ Type = "**="
	// ELLIPSE is the string rep. of the ellises tok.
	ELLIPSE Type = "..."
	// FDIVEQ is the string rep. of the floor div equal tok.
	FDIVEQ Type = "//="
	// RSHIFTEQ is the string rep. of the right shift equal tok.
	RSHIFTEQ Type = ">>="
	// LSHIFTEQ is the string rep. of the left shift equal tok.
	LSHIFTEQ Type = "<<="
	// NONINCRANGE is the string rep. of the non inclusive range token
	NONINCRANGE Type = "..<"

	// Keywords

	// FUNCTION is the string rep. of a function tok.
	FUNCTION Type = "FUNCTION"
	// VAR is the string rep. of a var tok.
	VAR Type = "VAR"
	// VAL is the string rep. of a val tok.
	VAL Type = "VAL"
	// TRUE is the string rep. of the `true` tok.
	TRUE Type = "TRUE"
	// FALSE is the string rep. of the `false` tok.
	FALSE Type = "FALSE"
	// IF is the string rep. of the `if` tok.
	IF Type = "IF"
	// ELSE is the string rep. of the `else` tok.
	ELSE Type = "ELSE"
	// RETURN is the string rep of the `return` tok.
	RETURN Type = "RETURN"
	// FOR is the string rep. of the `for` tok.
	FOR Type = "FOR"
	// IN is the string rep. of the `in` tok.
	IN Type = "IN"
	// NOTIN is the string rep. of the `notin` tok.
	NOTIN Type = "NOTIN"
	// AND is the string rep. of the `and` tok.
	AND Type = "AND"
	// OR is the string rep. of the `or` tok.
	OR Type = "OR"
	// NOT is the string rep. of the `not` tok.
	NOT Type = "NOT"
	// CONST is the string rep. of the `const` tok.
	CONST Type = "CONST"
	// MATCH is the string rep. of the `match` tok.
	MATCH Type = "MATCH"
	// NULL_KW is the string rep. of the `null` tok
	NULL_KW Type = "NULL_KW"

	// IMPORT is the string rep. of the import tok
	IMPORT Type = "IMPORT"
	// IMPORT_PATH is the string rep. of the import path tok
	IMPORT_PATH Type = "IMPORT_PATH"
	// FROM is the string rep. of the from import tok
	FROM Type = "FROM"
	// AS is the string rep. of the as tok
	AS Type = "AS"

	// TRY is the string rep. of the 'try' keyword token
	TRY Type = "TRY"
	// CATCH is the string rep. of the 'catch' keyword token
	CATCH Type = "CATCH"
	// FINALLY is the string rep. of the 'finally' keyword token
	FINALLY Type = "FINALLY"

	// EVAL is the string rep. of the 'eval' keyword token
	EVAL Type = "EVAL"
	// SPAWN is the string rep. of the 'spawn' keyword token
	SPAWN Type = "SPAWN"
	// DEFER is the string rep. of the 'defer' keyword token
	DEFER Type = "DEFER"
	// SELF is the string rep. of the 'self' keyword token
	SELF Type = "SELF"

	// BREAK is the string rep. of the 'break' keyword token
	BREAK Type = "BREAK"
	// CONTINUE is the string rep. of the 'continue' keyword token
	CONTINUE Type = "CONTINUE"
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
