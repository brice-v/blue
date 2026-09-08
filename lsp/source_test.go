package lsp

import (
	"strings"
	"testing"
)

func TestDocSourceLineIndexAndPositions(t *testing.T) {
	src := newDocSource("test.b", "val x = 1\nfun y() {\n  return 2\n}\n")

	if got := src.lineCount(); got != 5 {
		t.Fatalf("expected 5 lines, got %d", got)
	}
	if got := src.lineText(1); got != "fun y() {" {
		t.Fatalf("unexpected line 1: %q", got)
	}
	// "val x = 1\n" is ten runes long, so line 2 begins right after it.
	if got := src.startOfLine(2); got != 20 {
		t.Fatalf("expected line 2 to start at rune 20, got %d", got)
	}
	if got := src.lineText(99); got != "" {
		t.Fatalf("out of range line should be empty, got %q", got)
	}

	pos := src.positionOf(22)
	if pos.Line != 2 || pos.Character != 2 {
		t.Fatalf("unexpected position for rune 22: %+v", pos)
	}
	if got := src.offsetOf(pos); got != 22 {
		t.Fatalf("expected offsetOf to round trip back to 22, got %d", got)
	}

	// Positions past the end of the buffer clamp instead of panicking.
	last := src.positionOf(len(src.runes) + 50)
	if last.Line != 4 {
		t.Fatalf("expected clamping to the last line, got %+v", last)
	}
	if got := src.offsetOf(position{Line: 99, Character: 99}); got != len(src.runes) {
		t.Fatalf("expected offsetOf to clamp to the buffer length, got %d", got)
	}
}

func TestDocSourceCountsColumnsInUtf16CodeUnits(t *testing.T) {
	text := "val emoji = \"a😀b\"\nprintln(emoji)\n"
	src := newDocSource("unicode.b", text)

	// 'a', one astral emoji (two code units) and 'b'.
	if got := utf16Len([]rune("a😀b")); got != 4 {
		t.Fatalf("expected 4 UTF-16 code units, got %d", got)
	}

	line := []rune(src.lineText(0))
	// 17 runes where the emoji is one rune occupying two code units. If a client
	// asks for any column up to that width it must map inside the line.
	full := utf16Len(line)
	if full != 18 {
		t.Fatalf("expected 18 columns on line 0, got %d", full)
	}
	if full == 0 {
		t.Fatal("line is unexpectedly empty")
	}

	// Columns must map to non-decreasing rune offsets and never overshoot.
	previous := -1
	for col := 0; col <= full; col++ {
		offset := src.offsetOf(position{Line: 0, Character: col})
		if offset < previous {
			t.Fatalf("offset went backwards at column %d: %d after %d", col, offset, previous)
		}
		previous = offset
		back := src.positionOf(offset)
		if back.Line != 0 || back.Character > col {
			t.Fatalf("column %d round tripped to %+v", col, back)
		}
	}

	// The full width round trips exactly.
	if got := src.offsetOf(position{Line: 0, Character: full}); got != src.startOfLine(1)-1 {
		t.Fatalf("expected the last column of line 0 to be rune %d, got %d", src.startOfLine(1)-1, got)
	}
}

func TestTokenizeSkipsCommentsAndStrings(t *testing.T) {
	text := `val a = 1 # not ## a comment
###
block comment with "quotes" and fun(x)
###
## docstring line
val b = "a#{interp}y"
val c = 'sq'
val d = """raw
stuff"""
` + "val e = `exec me #{x}`\n" + "val f = r/a\\/b/\n"

	src := newDocSource("t.b", text)
	tokens := tokenize(src.runes)

	for _, tok := range tokens {
		if tok.kind == kIdent && (tok.text == "quotes" || tok.text == "interp" || tok.text == "stuff") {
			t.Errorf("text inside strings and comments must never become an identifier, got %q", tok.text)
		}
		if tok.kind == kWord && tok.text == "fun" {
			t.Errorf("the fun inside the block comment was tokenized as a keyword")
		}
	}

	var comments []string
	var literals []string
	for _, tok := range tokens {
		switch tok.kind {
		case kComment:
			comments = append(comments, tok.text)
		case kString:
			literals = append(literals, tok.text)
		}
	}

	// The `###` block comment is swallowed whole, so only the line comment and
	// the docstring survive as tokens.
	if len(comments) != 2 {
		t.Fatalf("expected 2 comment tokens, got %d: %#v", len(comments), comments)
	}
	if comments[0] != "# not ## a comment" {
		t.Fatalf("unexpected first comment: %q", comments[0])
	}
	if comments[1] != "## docstring line" {
		t.Fatalf("unexpected second comment: %q", comments[1])
	}

	wantLiterals := []string{"1", "\"a#{interp}y\"", "'sq'", "\"\"\"raw\nstuff\"\"\"", "`exec me #{x}`", "r/a\\/b/"}
	if len(literals) != len(wantLiterals) {
		t.Fatalf("expected %d literals, got %d: %#v", len(wantLiterals), len(literals), literals)
	}
	for i, want := range wantLiterals {
		if literals[i] != want {
			t.Fatalf("literal %d: expected %q, got %q", i, want, literals[i])
		}
	}

	// Everything after the last string still tokenizes.
	foundPrintln := false
	for _, tok := range tokens {
		if tok.kind == kIdent && tok.text == "f" {
			foundPrintln = true
		}
	}
	if !foundPrintln {
		t.Error("expected identifiers after the regex literal to still be tokenized")
	}
	if !strings.Contains(text, "r/a\\/b/") {
		t.Error("test input lost the regex literal")
	}
}

func TestTokenizeNumbersKeepRangesSeparate(t *testing.T) {
	src := newDocSource("t.b", "x = 0x1F\ny = 1..5\nz = 2..<10\nw = 1.5e3\n")
	tokens := tokenize(src.runes)

	got := []string{}
	for _, tok := range tokens {
		if tok.kind == kString || (tok.kind == kPunct && (tok.text == "." || tok.text == "<")) {
			got = append(got, tok.text)
		}
	}

	want := []string{"0x1F", "1", ".", ".", "5", "2", ".", ".", "<", "10", "1.5e3"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("token %d: expected %q, got %q (all: %v)", i, want[i], got[i], got)
		}
	}
}

func TestWordAtFindsIdentifierAndNeighbour(t *testing.T) {
	src := newDocSource("t.b", "val hello = world\n")
	tokens := tokenize(src.runes)

	// In the middle of `world`.
	if tok, ok := wordAt(tokens, 14); !ok || tok.text != "world" {
		t.Fatalf("expected `world`, got %q ok=%v", tok.text, ok)
	}
	// Directly after `hello` still resolves that word.
	if tok, ok := wordAt(tokens, 10); !ok || tok.text != "hello" {
		t.Fatalf("expected `hello` right after the word, got %q ok=%v", tok.text, ok)
	}
	// Past the end of the token list there is nothing.
	if _, ok := wordAt(tokens, 23); ok {
		t.Fatal("expected no word at the end of the buffer")
	}
}

func TestInsideStringOrComment(t *testing.T) {
	src := newDocSource("t.b", "val s = \"hello there\"\n# comment here\nval x = 1\n")
	index := buildIndex(src)
	index.resolveExtents()

	// Inside the string literal on line 0.
	if !index.insideStringOrComment(14) {
		t.Error("expected offset 14 (inside the string) to be reported")
	}
	// Inside the comment on line 1.
	if !index.insideStringOrComment(30) {
		t.Error("expected offset 30 (inside the comment) to be reported")
	}
	// Real code after the comment ended.
	if index.insideStringOrComment(40) {
		t.Error("code after the comment should not be reported as inside a comment")
	}
}
