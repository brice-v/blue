package lsp

import (
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

// docSource is a document buffer plus the line index needed to translate between
// LSP positions (line, UTF-16 code unit) and rune offsets into the text.
type docSource struct {
	name       string // name given to blue's lexer for this buffer
	text       string
	runes      []rune
	lineStarts []int // rune offset of every line start (always begins with 0)
}

func newDocSource(name, text string) *docSource {
	d := &docSource{name: name, text: text, runes: []rune(text)}
	d.lineStarts = []int{0}
	for i, r := range d.runes {
		if r == '\n' {
			d.lineStarts = append(d.lineStarts, i+1)
		}
	}
	return d
}

// dir returns the directory the buffer's file lives in, which is what relative
// import paths resolve against (same rule the compiler uses).
func (d *docSource) dir() string {
	if d.name == "" || strings.HasPrefix(d.name, "<") {
		return "."
	}
	return filepath.Dir(d.name)
}

// lineCount returns the number of lines in the buffer. A buffer that does not
// end with a newline still counts its trailing partial line.
func (d *docSource) lineCount() int {
	return len(d.lineStarts)
}

// lineRunes returns the runes of one line without its trailing newline.
func (d *docSource) lineRunes(line int) []rune {
	if line < 0 || line >= len(d.lineStarts) {
		return nil
	}
	start := d.lineStarts[line]
	end := len(d.runes)
	if line+1 < len(d.lineStarts) {
		end = d.lineStarts[line+1] - 1
	}
	return d.runes[start:end]
}

// lineText returns one line without its trailing newline. Out of range lines are
// empty rather than fatal, because parser error positions can point past the end
// of a buffer that has since been edited.
func (d *docSource) lineText(line int) string {
	return string(d.lineRunes(line))
}

// startOfLine returns the rune offset a line begins at.
func (d *docSource) startOfLine(line int) int {
	if line < 0 {
		return 0
	}
	if line >= len(d.lineStarts) {
		return len(d.runes)
	}
	return d.lineStarts[line]
}

// positionOf converts a rune offset into an LSP position. Offsets past the end
// of the buffer clamp to the end of the last line.
func (d *docSource) positionOf(runeIdx int) position {
	if runeIdx <= 0 {
		return position{}
	}
	if runeIdx > len(d.runes) {
		runeIdx = len(d.runes)
	}
	line := sort.Search(len(d.lineStarts), func(i int) bool { return d.lineStarts[i] > runeIdx }) - 1
	if line < 0 {
		line = 0
	}
	return position{Line: line, Character: utf16Len(d.runes[d.lineStarts[line]:runeIdx])}
}

// offsetOf converts an LSP position into a rune offset, clamped to the buffer.
func (d *docSource) offsetOf(p position) int {
	if p.Line < 0 {
		return 0
	}
	if p.Line >= len(d.lineStarts) {
		return len(d.runes)
	}
	idx := d.lineStarts[p.Line]
	end := len(d.runes)
	if p.Line+1 < len(d.lineStarts) {
		end = d.lineStarts[p.Line+1] - 1
	}

	remaining := p.Character
	for i := idx; i < end; {
		r := d.runes[i]
		units := 1
		if r >= 0x10000 {
			// Astral characters take two UTF-16 code units.
			units = 2
		}
		if remaining < units {
			return i
		}
		remaining -= units
		i++
	}
	return end
}

// textUpToPosition returns the text of a line up to a UTF-16 column, which is
// what completion needs in order to filter candidates.
func (d *docSource) textUpToPosition(p position) string {
	if p.Line < 0 || p.Line >= len(d.lineStarts) {
		return ""
	}
	return string(d.runes[d.startOfLine(p.Line):d.offsetOf(p)])
}

// utf16Len counts the UTF-16 code units a rune slice would occupy, which is how
// LSP measures columns. Runes above U+FFFF take two code units.
func utf16Len(rs []rune) int {
	total := 0
	for _, r := range rs {
		if r >= 0x10000 {
			total += 2
			continue
		}
		total++
	}
	return total
}

// tokenKind classifies the things the LSP cares about while scanning blue source
// for names, comment regions and string regions.
type tokenKind uint8

const (
	kPunct tokenKind = iota
	kIdent
	kWord // identifier that is also one of blue's reserved words
	kString
	kComment
)

// scanToken is one scanned token with exact rune offsets into its source.
type scanToken struct {
	kind  tokenKind
	text  string
	start int
	end   int
}

// wordAt returns the identifier or keyword token covering a rune offset. When the
// offset falls between tokens it falls back to the word ending exactly there so
// that hovering right after an identifier still resolves it.
func wordAt(tokens []scanToken, idx int) (scanToken, bool) {
	i := sort.Search(len(tokens), func(i int) bool { return tokens[i].end > idx })
	if i >= len(tokens) {
		return scanToken{}, false
	}
	tok := tokens[i]
	if tok.kind == kIdent || tok.kind == kWord {
		return tok, true
	}
	if i > 0 {
		prev := tokens[i-1]
		if prev.kind == kIdent || prev.kind == kWord {
			return prev, true
		}
	}
	return scanToken{}, false
}

// wordAroundRune expands from a rune offset over the identifier touching it. The
// token stream cannot answer for positions inside comments or raw strings since
// those are consumed whole, and an editor reports the cursor position between two
// characters so both sides have to be checked.
func wordAroundRune(rs []rune, idx int) string {
	if idx < 0 || idx > len(rs) {
		return ""
	}
	start := idx
	for start > 0 && isIdentRune(rs[start-1]) {
		start--
	}
	end := idx
	for end < len(rs) && isIdentRune(rs[end]) {
		end++
	}
	if end-start == 0 {
		return ""
	}
	return string(rs[start:end])
}

func isIdentStartRune(r rune) bool { return unicode.IsLetter(r) || r == '_' }

func isIdentRune(r rune) bool {
	return isIdentStartRune(r) || unicode.IsNumber(r) || r == '?' || r == '!'
}

// tokenize splits blue source into tokens with exact rune offsets.
//
// It is deliberately lighter weight than package lexer: it only has to get
// identifier boundaries, comments and strings right because every feature below
// works from those offsets, and it must never fail on half typed code. The
// delimiter rules mirror the real lexer: '#' line comments, '###' block comments
// ended by another '###', '##' docstring comments up to end of line, "..." and
// '...' strings with backslash escapes, `"""` raw strings, backtick exec
// strings and r/.../ regex literals.
func tokenize(rs []rune) []scanToken {
	out := []scanToken{}
	i := 0

	for i < len(rs) {
		r := rs[i]
		if unicode.IsSpace(r) {
			i++
			continue
		}

		switch {
		case r == '#':
			start := i
			switch {
			case i+2 < len(rs) && rs[i+1] == '#' && rs[i+2] == '#':
				j := i + 3
				for j < len(rs) {
					if j+2 < len(rs) && rs[j] == '#' && rs[j+1] == '#' && rs[j+2] == '#' {
						j += 3
						break
					}
					j++
				}
				i = j
			case i+1 < len(rs) && rs[i+1] == '{':
				// String interpolation marker, the braces balance on their own.
				out = append(out, scanToken{kind: kPunct, text: "#{", start: start, end: i + 2})
				i += 2
			case i+1 < len(rs) && rs[i+1] == '#':
				j := i + 2
				for j < len(rs) && rs[j] != '\n' {
					j++
				}
				out = append(out, scanToken{kind: kComment, text: string(rs[start:j]), start: start, end: j})
				i = j
			default:
				j := i + 1
				for j < len(rs) && rs[j] != '\n' {
					j++
				}
				out = append(out, scanToken{kind: kComment, text: string(rs[start:j]), start: start, end: j})
				i = j
			}

		case r == '"':
			start := i
			if i+2 < len(rs) && rs[i+1] == '"' && rs[i+2] == '"' {
				j := i + 3
				closed := false
				for j+2 < len(rs) {
					if rs[j] == '"' && rs[j+1] == '"' && rs[j+2] == '"' {
						j += 3
						closed = true
						break
					}
					j++
				}
				if !closed {
					j = len(rs)
				}
				out = append(out, scanToken{kind: kString, text: string(rs[start:j]), start: start, end: j})
				i = j
				continue
			}
			j := i + 1
			for j < len(rs) && rs[j] != '"' {
				if rs[j] == '\\' && j+1 < len(rs) {
					j += 2
					continue
				}
				j++
			}
			if j < len(rs) {
				j++ // closing quote
			}
			out = append(out, scanToken{kind: kString, text: string(rs[start:j]), start: start, end: j})
			i = j

		case r == '\'':
			start := i
			j := i + 1
			for j < len(rs) && rs[j] != '\'' {
				if rs[j] == '\\' && j+1 < len(rs) {
					j += 2
					continue
				}
				j++
			}
			if j < len(rs) {
				j++
			}
			out = append(out, scanToken{kind: kString, text: string(rs[start:j]), start: start, end: j})
			i = j

		case r == '`':
			start := i
			j := i + 1
			for j < len(rs) && rs[j] != '`' {
				j++
			}
			if j < len(rs) {
				j++
			}
			out = append(out, scanToken{kind: kString, text: string(rs[start:j]), start: start, end: j})
			i = j

		case r == 'r' && i+1 < len(rs) && rs[i+1] == '/':
			start := i
			j := i + 2
			for j < len(rs) {
				if rs[j] == '\\' && j+1 < len(rs) {
					j += 2
					continue
				}
				if rs[j] == '/' {
					j++
					break
				}
				j++
			}
			out = append(out, scanToken{kind: kString, text: string(rs[start:j]), start: start, end: j})
			i = j

		case isIdentStartRune(r):
			start := i
			j := i + 1
			for j < len(rs) && isIdentRune(rs[j]) {
				j++
			}
			word := string(rs[start:j])
			kind := kIdent
			if isKeyword(word) {
				kind = kWord
			}
			out = append(out, scanToken{kind: kind, text: word, start: start, end: j})
			i = j

		case unicode.IsDigit(r):
			start := i
			j := numericEnd(rs, i)
			out = append(out, scanToken{kind: kString, text: string(rs[start:j]), start: start, end: j})
			i = j

		default:
			out = append(out, scanToken{kind: kPunct, text: string(r), start: i, end: i + 1})
			i++
		}
	}

	return out
}

// numericEnd consumes a numeric literal starting at i, including hex, binary and
// octal prefixes, float dots and exponents, while leaving ranges such as 1..5
// and 1..<10 alone.
func numericEnd(rs []rune, i int) int {
	j := i + 1
	if rs[i] == '0' && j < len(rs) {
		switch rs[j] {
		case 'x', 'X', 'o', 'O', 'b', 'B':
			j++
			for j < len(rs) && (isDigitOrUnderscore(rs[j]) || isHexLetter(rs[j])) {
				j++
			}
			return j
		}
	}
	for j < len(rs) && (isDigitOrUnderscore(rs[j])) {
		j++
	}
	if j+1 < len(rs) && rs[j] == '.' && unicode.IsDigit(rs[j+1]) {
		j++
		for j < len(rs) && (isDigitOrUnderscore(rs[j])) {
			j++
		}
	}
	if j+1 < len(rs) && (rs[j] == 'e' || rs[j] == 'E') {
		k := j + 1
		if k < len(rs) && (rs[k] == '+' || rs[k] == '-') {
			k++
		}
		if k < len(rs) && unicode.IsDigit(rs[k]) {
			for j = k; j < len(rs) && (isDigitOrUnderscore(rs[j])); j++ {
			}
		}
	}
	return j
}

func isDigitOrUnderscore(r rune) bool { return unicode.IsDigit(r) || r == '_' }

func isHexLetter(r rune) bool {
	return (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}
