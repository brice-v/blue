package lexer

import (
	"blue/token"
	"unicode"
)

func (l *Lexer) newToken(tokenType token.Type, ch rune) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch), LineNumber: l.lineNo, PositionInLine: l.posInLine, Filepath: l.fname}
}

// makeTwoCharToken takes a tokens type and returns the new token
// while advancing the readPosition and current char
func (l *Lexer) makeTwoCharToken(typ token.Type) token.Token {
	ch := l.ch
	lineNo := l.lineNo
	posInLine := l.posInLine
	// consume next char because we know it is an =
	l.readChar()
	// TODO: Could add length here of 2?
	return token.Token{Type: typ, Literal: string(ch) + string(l.ch), LineNumber: lineNo, PositionInLine: posInLine, Filepath: l.fname}
}

// makeThreeCharToken takes a tokens type and returns the new token
// while advancing the readPosition and current char to the proper position
func (l *Lexer) makeThreeCharToken(typ token.Type) token.Token {
	ch := l.ch
	lineNo := l.lineNo
	posInLine := l.posInLine
	l.readChar()
	ch1 := l.ch
	l.readChar()
	// TODO: Could add length here of 3?
	return token.Token{Type: typ, Literal: string(ch) + string(ch1) + string(l.ch), LineNumber: lineNo, PositionInLine: posInLine, Filepath: l.fname}
}

// isLetter will return true if the rune given matches the pattern below
// 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
// TOOD: Update isLetter to include more valid identifier chars such as ! and numbers
func isLetter(ch rune) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_' || ch == '?'
	// return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch == '?'
}

// isImportChar will return true if the rune given is allowed as part of an import path
//
// Note: numbers are not allowed in the filename because they are not allowed in identifiers
// this is a design decision and prevents issues.  The reason why '.' is allowed is because
// that will signifiy the path separation in the import path.
//
// We could just use a basic string which would solve most of these problems but i like
// the look of python's import syntax :)
func isImportChar(ch rune) bool {
	return isLetter(ch) || ch == '.'
}

// isDigit will return true if the rune give is 0-9
// TODO: Support isDigit for unicode values
func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

// isHexChar will return true if the rune given is a hex character
func isHexChar(ch rune) bool {
	return 'a' <= ch && ch <= 'f' || 'A' <= ch && ch <= 'F' || '0' <= ch && ch <= '9'
}

// isOctalChar will return true if the rune given is an octal character
func isOctalChar(ch rune) bool {
	return '0' <= ch && ch <= '7'
}

// isBinaryChar will return true if the rune given is a binary character
func isBinaryChar(ch rune) bool {
	return '0' == ch || '1' == ch
}
