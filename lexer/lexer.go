package lexer

import (
	"blue/consts"
	"blue/lib"
	"blue/token"
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Lexer is the struct that contains members for
// lexing needs
type Lexer struct {
	input        string
	inputAsRunes []rune
	position     int  // current pos. in input (points to current char)
	readPosition int  // current reading pos. in input (after current char)
	ch           rune // current char under examination
	prevCh       rune // previous char read

	// For error printing
	lineNo    int
	posInLine int
	fname     string

	prevTokType token.Type
}

// New returns a pointer to the lexer struct
func New(input string, fname string) *Lexer {
	l := &Lexer{input: input, inputAsRunes: []rune(input), fname: fname, prevTokType: token.ILLEGAL}
	l.readChar()
	return l
}

// TODO: Cache the split lines per fpath
func GetErrorLineMessageForJson(tok token.Token) (string, string) {
	lineNo := tok.LineNumber
	posInLine := tok.PositionInLine

	fpath := tok.Filepath
	var fullFile string
	if fpath == consts.CORE_FILE_PATH {
		fullFile = lib.CoreFile
	} else if strings.HasPrefix(fpath, "<std/") {
		// This shouldnt fail because we set the filepaths
		file := strings.ReplaceAll(strings.Split(fpath, "<std/")[1], ">", "")
		fullFile = lib.ReadStdFileToString(file)
	} else {
		fdata, err := os.ReadFile(fpath)
		if err != nil {
			// fallback option if the filepath doesnt exist or if its internal
			return fmt.Sprint(tok.DisplayForErrorLine()), ""
		}
		fullFile = string(fdata)
	}

	// TODO: This will be VERY SLOW but I just want to test it out
	lines := strings.Split(fullFile, "\n")
	var out bytes.Buffer
	// Fist write the filename of the input
	fileErrorLine := fmt.Sprintf("%s:%d:%d ", fpath, lineNo+1, posInLine+1)
	out.WriteString(fileErrorLine)
	// Write the line
	out.WriteString(lines[lineNo])
	// Then a newline (so we can play a ^ on the following line)
	firstStr := out.String()
	out = bytes.Buffer{}
	// Write Spaces to pad the following line
	for range fileErrorLine {
		out.WriteByte(' ')
	}
	for i := range lines[lineNo] {
		if i == posInLine {
			out.WriteByte('^')
		} else {
			out.WriteByte(' ')
		}
	}
	return firstStr, out.String()
}

// TODO: Cache the split lines per fpath
func GetErrorLineMessage(tok token.Token) string {
	lineNo := tok.LineNumber
	posInLine := tok.PositionInLine

	fpath := tok.Filepath
	var fullFile string
	if fpath == consts.CORE_FILE_PATH {
		fullFile = lib.CoreFile
	} else if strings.HasPrefix(fpath, "<std/") {
		// This shouldnt fail because we set the filepaths
		file := strings.ReplaceAll(strings.Split(fpath, "<std/")[1], ">", "")
		fullFile = lib.ReadStdFileToString(file)
	} else {
		fdata, err := os.ReadFile(fpath)
		if err != nil {
			// fallback option if the filepath doesnt exist or if its internal
			return fmt.Sprint(tok.DisplayForErrorLine())
		}
		fullFile = string(fdata)
	}

	// TODO: This will be VERY SLOW but I just want to test it out
	lines := strings.Split(fullFile, "\n")
	var out bytes.Buffer
	// Fist write the filename of the input
	fileErrorLine := fmt.Sprintf("%s:%d:%d ", fpath, lineNo+1, posInLine+1)
	out.WriteString(fileErrorLine)
	// Write the line
	out.WriteString(lines[lineNo])
	// Then a newline (so we can play a ^ on the following line)
	out.WriteByte('\n')
	// Write Spaces to pad the following line
	for range fileErrorLine {
		out.WriteByte(' ')
	}
	for i := range lines[lineNo] {
		if i == posInLine {
			out.WriteByte('^')
		} else {
			out.WriteByte(' ')
		}
	}
	return out.String()
}

// readChar gives us the next character and advances out position
// in the input string
func (l *Lexer) readChar() {
	if l.ch == '\n' {
		l.lineNo++
		l.posInLine = 0
	} else {
		l.posInLine++
	}
	l.prevCh = l.ch
	if l.readPosition >= len(l.inputAsRunes) {
		l.ch = 0
	} else {
		l.ch = l.inputAsRunes[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += utf8.RuneCountInString(string(l.ch))
}

// peekChar will return the rune that is in the readPosition without consuming any input
func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.inputAsRunes) {
		return 0
	}
	return l.inputAsRunes[l.readPosition]
}

// peekSecondChar will return the rune right after the readPosition without consuming any input
func (l *Lexer) peekSecondChar() rune {
	if l.readPosition >= len(l.inputAsRunes) || l.readPosition+1 >= len(l.inputAsRunes) {
		return 0
	}
	return l.inputAsRunes[l.readPosition+1]
}

// readHexNumber will read the hex number as a helper
func (l *Lexer) readHexNumber(ogPos int) (token.Type, string) {
	// consume the 0 and x and continue to the number
	l.readChar()
	l.readChar()
	for isHexChar(l.ch) || (l.ch == '_' && isHexChar(l.peekChar())) {
		l.readChar()
	}
	return token.HEX, string(l.inputAsRunes[ogPos:l.position])
}

// readOctalNumber will read the hex number as a helper
func (l *Lexer) readOctalNumber(ogPos int) (token.Type, string) {
	// consume the 0 and the o and continue to the number
	l.readChar()
	l.readChar()
	for isOctalChar(l.ch) || (l.ch == '_' && isOctalChar(l.peekChar())) {
		l.readChar()
	}
	return token.OCTAL, string(l.inputAsRunes[ogPos:l.position])
}

func (l *Lexer) readBinaryNumber(ogPos int) (token.Type, string) {
	// consume the 0 and the b and continue to the number
	l.readChar()
	l.readChar()
	for isBinaryChar(l.ch) || (l.ch == '_' && isBinaryChar(l.peekChar())) {
		l.readChar()
	}
	return token.BINARY, string(l.inputAsRunes[ogPos:l.position])
}

// readNumber will keep consuming valid digits of the input according to `isDigit`
// and return the string
func (l *Lexer) readNumber() (token.Type, string) {
	position := l.position
	if l.ch == '0' {
		if l.peekChar() == 'x' && isHexChar(l.peekSecondChar()) {
			return l.readHexNumber(position)
		} else if l.peekChar() == 'o' && isOctalChar(l.peekSecondChar()) {
			return l.readOctalNumber(position)
		} else if l.peekChar() == 'b' && isBinaryChar(l.peekSecondChar()) {
			return l.readBinaryNumber(position)
		}
	}
	dotFlag := false
	for isDigit(l.ch) || (l.ch == '_' && isDigit(l.peekChar())) {
		if l.peekChar() == '.' && !dotFlag && l.peekSecondChar() != '.' {
			dotFlag = true
			l.readChar()
			l.readChar()
		}
		l.readChar()
	}
	if dotFlag {
		return token.FLOAT, string(l.inputAsRunes[position:l.position])
	}
	return token.INT, string(l.inputAsRunes[position:l.position])
}

// readIdentifier will keep consuming valid letters out of the input according to `isLetter`
// and return the string
func (l *Lexer) readIdentifier() string {
	position := l.position
	// Note: We can only do this because we check if the first char is a 'letter'
	// That includes underscores which is why 1 of the lexer tests changes to accomodate that
	for isLetter(l.ch) || unicode.IsNumber(l.ch) {
		l.readChar()
	}
	return string(l.inputAsRunes[position:l.position])
}

// readImportPath reads the following import chars and returns the accumulated string
func (l *Lexer) readImportPath() string {
	position := l.position
	for isImportChar(l.ch) {
		l.readChar()
	}
	return string(l.inputAsRunes[position:l.position])
}

// readMultiLineComment will continue to consume input until the end multiline token is reached
func (l *Lexer) readMultiLineComment() {
	for l.ch != 0 {
		if l.ch == 0 {
			break // break on EOF
		}
		if l.ch == '#' && l.peekChar() == '#' && l.peekSecondChar() == '#' {
			l.readChar()
			l.readChar()
			l.readChar()
			// fmt.Println(l.ch)
			break
		}
		l.readChar()
	}
}

// readSingleLineComment will continue to consume input until the EOL is reached
func (l *Lexer) readSingleLineComment() {
	for l.ch != 0 {
		if l.ch == 0 {
			break
		}
		if l.ch == '#' {
			l.readChar()
			for l.ch != '\n' {
				l.readChar()
				if l.ch == 0 {
					break
				}
			}
			break
		}
		l.readChar()
	}
}

func (l *Lexer) readExecString() string {
	b := strings.Builder{}
	for {
		l.readChar()
		if l.ch == '`' || l.ch == 0 {
			l.readChar()
			break
		}
		b.WriteRune(l.ch)
	}
	return b.String()
}

func (l *Lexer) readRawString() string {
	b := &strings.Builder{}
	// Skip the first 2 " chars
	l.readChar()
	l.readChar()
	for {
		l.readChar()
		if (l.ch == '"' && l.peekChar() == '"' && l.peekSecondChar() == '"') || l.ch == 0 {
			l.readChar()
			l.readChar()
			break
		}
		b.WriteRune(l.ch)
	}
	return b.String()
}

// readString will consume tokens until the string is fully read
func (l *Lexer) readString() (string, error) {
	b := &strings.Builder{}

	stringStart := string(l.ch)
	for {
		l.readChar()

		// Support some basic escapes like \"
		if l.ch == '\\' {
			switch l.peekChar() {
			case '"':
				b.WriteByte('"')
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			case '\\':
				b.WriteByte('\\')
			case 'x':
				// Skip over the the '\\', 'x' and the next two bytes (hex)
				l.readChar()
				l.readChar()
				l.readChar()
				src := string([]rune{l.prevCh, l.ch})
				dst, err := hex.DecodeString(src)
				if err != nil {
					return "", err
				}
				b.Write(dst)
				continue
			}

			// Skip over the '\\' and the matched single escape char
			l.readChar()
			continue
		} else {
			if string(l.ch) == stringStart || l.ch == 0 {
				break
			}
		}

		b.WriteRune(l.ch)
	}

	if l.ch != '"' && l.ch != '\'' {
		return "", fmt.Errorf("string is not ended")
	}

	return b.String(), nil
}

// skipWhitespace will continue to advance if the current byte is considered
// a whitespace character such as ' ', '\t', '\n', '\r'
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// NextToken matches against a byte and if it succeeds it will
// read the next char and return a token struct
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			// Note: cant use newToken here because it is not 1 byte long
			tok = l.makeTwoCharToken(token.EQ)
		} else if l.peekChar() == '>' {
			tok = l.makeTwoCharToken(token.RARROW)
		} else {
			tok = l.newToken(token.ASSIGN, l.ch)
		}
	case ';':
		tok = l.newToken(token.SEMICOLON, l.ch)
	case '(':
		tok = l.newToken(token.LPAREN, l.ch)
	case ')':
		tok = l.newToken(token.RPAREN, l.ch)
	case '{':
		tok = l.newToken(token.LBRACE, l.ch)
	case '}':
		tok = l.newToken(token.RBRACE, l.ch)
	case '[':
		tok = l.newToken(token.LBRACKET, l.ch)
	case ']':
		tok = l.newToken(token.RBRACKET, l.ch)
	case ',':
		tok = l.newToken(token.COMMA, l.ch)
	case '+':
		if l.peekChar() == '=' {
			tok = l.makeTwoCharToken(token.PLUSEQ)
		} else {
			tok = l.newToken(token.PLUS, l.ch)
		}
	case '!':
		if l.peekChar() == '=' {
			tok = l.makeTwoCharToken(token.NEQ)
		} else {
			tok = l.newToken(token.BANG, l.ch)
		}
	case '-':
		if l.peekChar() == '=' {
			tok = l.makeTwoCharToken(token.MINUSEQ)
		} else {
			tok = l.newToken(token.MINUS, l.ch)
		}
	case '/':
		if l.peekChar() == '=' {
			tok = l.makeTwoCharToken(token.DIVEQ)
		} else if l.peekChar() == '/' && l.peekSecondChar() != '=' {
			tok = l.makeTwoCharToken(token.FDIV)
		} else if l.peekChar() == '/' && l.peekSecondChar() == '=' {
			tok = l.makeThreeCharToken(token.FDIVEQ)
		} else {
			tok = l.newToken(token.FSLASH, l.ch)
		}
	case '*':
		if l.peekChar() == '=' {
			tok = l.makeTwoCharToken(token.MULEQ)
		} else if l.peekChar() == '*' && l.peekSecondChar() != '=' {
			tok = l.makeTwoCharToken(token.POW)
		} else if l.peekChar() == '*' && l.peekSecondChar() == '=' {
			tok = l.makeThreeCharToken(token.POWEQ)
		} else {
			tok = l.newToken(token.STAR, l.ch)
		}
	case '<':
		if l.peekChar() == '<' {
			if l.peekSecondChar() == '=' {
				tok = l.makeThreeCharToken(token.LSHIFTEQ)
			} else {
				tok = l.makeTwoCharToken(token.LSHIFT)
			}
		} else if l.peekChar() == '=' {
			tok = l.makeTwoCharToken(token.LTEQ)
		} else {
			tok = l.newToken(token.LT, l.ch)
		}
	case '>':
		if l.peekChar() == '>' {
			if l.peekSecondChar() == '=' {
				tok = l.makeThreeCharToken(token.RSHIFTEQ)
			} else {
				tok = l.makeTwoCharToken(token.RSHIFT)
			}
		} else if l.peekChar() == '=' {
			tok = l.makeTwoCharToken(token.GTEQ)
		} else {
			tok = l.newToken(token.GT, l.ch)
		}
	case '|':
		if l.peekChar() == '=' {
			tok = l.makeTwoCharToken(token.OREQ)
		} else {
			tok = l.newToken(token.PIPE, l.ch)
		}
	case '&':
		if l.peekChar() == '=' {
			tok = l.makeTwoCharToken(token.ANDEQ)
		} else {
			tok = l.newToken(token.AMPERSAND, l.ch)
		}
	case '^':
		if l.peekChar() == '=' {
			tok = l.makeTwoCharToken(token.XOREQ)
		} else {
			tok = l.newToken(token.HAT, l.ch)
		}
	case '#':
		if l.peekChar() == '{' {
			tok = l.makeTwoCharToken(token.STRINGINTERP)
		} else if l.peekChar() == '#' && l.peekSecondChar() == '#' {
			tok = l.makeThreeCharToken(token.MULTLINE_COMMENT)
			l.readMultiLineComment()
		} else {
			tok = l.newToken(token.HASH, l.ch)
			l.readSingleLineComment()
		}
	case '%':
		if l.peekChar() == '=' {
			tok = l.makeTwoCharToken(token.PERCENTEQ)
		} else {
			tok = l.newToken(token.PERCENT, l.ch)
		}
	case '.':
		if l.peekChar() == '.' {
			if l.peekSecondChar() == '.' {
				tok = l.makeThreeCharToken(token.ELLIPSE)
			} else if l.peekSecondChar() == '<' {
				tok = l.makeThreeCharToken(token.NONINCRANGE)
			} else {
				tok = l.makeTwoCharToken(token.RANGE)
			}
		} else {
			tok = l.newToken(token.DOT, l.ch)
		}
	case '~':
		if l.peekChar() == '=' {
			tok = l.makeTwoCharToken(token.BINNOTEQ)
		} else {
			tok = l.newToken(token.TILDE, l.ch)
		}
	case '`':
		tok.Filepath = l.fname
		tok.LineNumber = l.lineNo
		tok.PositionInLine = l.posInLine
		tok.Type = token.BACKTICK
		tok.Literal = l.readExecString()
		return tok
	case ':':
		tok = l.newToken(token.COLON, l.ch)
	case 0:
		tok.Filepath = l.fname
		tok.LineNumber = l.lineNo
		tok.PositionInLine = l.posInLine
		tok.Literal = ""
		tok.Type = token.EOF
	case '"':
		if l.peekChar() == '"' && l.peekSecondChar() == '"' {
			tok.Filepath = l.fname
			tok.LineNumber = l.lineNo
			tok.PositionInLine = l.posInLine
			str := l.readRawString()
			tok.Type = token.RAW_STRING
			tok.Literal = str
		} else {
			str, err := l.readString()
			if err != nil {
				tok = l.newToken(token.ILLEGAL, l.prevCh)
			} else {
				tok.Filepath = l.fname
				tok.LineNumber = l.lineNo
				tok.PositionInLine = l.posInLine
				tok.Type = token.STRING_DOUBLE_QUOTE
				tok.Literal = str
			}
		}
	case '\'':
		str, err := l.readString()
		if err != nil {
			tok = l.newToken(token.ILLEGAL, l.prevCh)
		} else {
			tok.Filepath = l.fname
			tok.Type = token.STRING_SINGLE_QUOTE
			tok.Literal = str
			tok.LineNumber = l.lineNo
			tok.PositionInLine = l.posInLine
		}
	default:
		if l.prevTokType == token.IMPORT {
			l.prevTokType = token.ILLEGAL
			tok.Filepath = l.fname
			tok.LineNumber = l.lineNo
			tok.PositionInLine = l.posInLine
			tok.Literal = l.readImportPath()
			tok.Type = token.IMPORT_PATH
			return tok
		} else if isLetter(l.ch) {
			tok.Filepath = l.fname
			tok.LineNumber = l.lineNo
			tok.PositionInLine = l.posInLine
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			// This is only used to determine that we need to read an import path
			if tok.Type == token.IMPORT {
				l.prevTokType = token.IMPORT
			}
			return tok
		} else if isDigit(l.ch) {
			tok.Filepath = l.fname
			tok.LineNumber = l.lineNo
			tok.PositionInLine = l.posInLine
			tok.Type, tok.Literal = l.readNumber()
			return tok
		}
		tok = l.newToken(token.ILLEGAL, l.ch)
	}

	l.readChar()
	return tok
}
