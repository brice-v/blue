package lsp

import (
	"strings"
	"testing"
)

// semToken is one decoded entry: where it lands and what it was called.
type semToken struct {
	line   int
	start  int // utf-16 code units from the start of the line
	length int
	typ    string
	mods   uint32
}

// decodeSemantic decodes the flat tuples exactly the way a client does: the line is
// a delta, and a start on the same line is relative to the previous token's start.
func decodeSemantic(t *testing.T, text string) []semToken {
	t.Helper()

	src := newDocSource("t.b", text)
	ix := buildIndex(src)
	ix.resolveExtents()
	data := encodeSemanticTokens(src, ix)

	if len(data)%5 != 0 {
		t.Fatalf("encoded data length %d is not a multiple of 5", len(data))
	}

	runes := []rune(text)
	out := []semToken{}
	prevLine := 0
	prevStart := 0

	for i := 0; i+4 < len(data); i += 5 {
		deltaLine := int(data[i])
		startChar := int(data[i+1])
		length := int(data[i+2])
		typeIdx := int(data[i+3])
		mods := data[i+4]

		if deltaLine < 0 {
			t.Fatalf("tuple %d: negative line delta %d", i/5, deltaLine)
		}
		if typeIdx >= len(semLegendTypes) {
			t.Fatalf("tuple %d: type index %d outside the legend of %d entries", i/5, typeIdx, len(semLegendTypes))
		}

		line := prevLine + deltaLine
		start := startChar
		if deltaLine == 0 {
			start = prevStart + startChar
			if startChar < 0 {
				t.Fatalf("tuple %d: negative same-line offset %d", i/5, startChar)
			}
		}

		lineStart := src.startOfLine(line)
		col := 0
		from := -1
		for j := lineStart; j < len(runes); j++ {
			if col == start {
				from = j
				break
			}
			col += utf16Len(runes[j : j+1])
		}
		if from < 0 {
			t.Fatalf("tuple %d: cannot map line %d column %d back to the source", i/5, line, start)
		}
		to := from
		for k := 0; k < length && to < len(runes); k++ {
			to += utf16Len(runes[to : to+1])
		}

		out = append(out, semToken{line: line, start: start, length: length, typ: semLegendTypes[typeIdx], mods: mods})
		prevLine = line
		prevStart = start
	}

	return out
}

func splitLines(text string) []string {
	return strings.Split(text, "\n")
}

// textAt resolves one decoded token back to the source text it covers.
func textAt(t *testing.T, text string, tok semToken) string {
	t.Helper()
	lines := splitLines(text)
	if tok.line < 0 || tok.line >= len(lines) {
		t.Fatalf("token on line %d but the buffer has %d lines", tok.line, len(lines))
	}
	runes := []rune(lines[tok.line])
	eaten := 0
	from := -1
	for i, r := range runes {
		if eaten == tok.start {
			from = i
			break
		}
		eaten += utf16Len([]rune{r})
	}
	if from < 0 {
		t.Fatalf("column %d is not on line %d (%q)", tok.start, tok.line, lines[tok.line])
	}
	to := from
	for k := 0; k < tok.length && to < len(runes); k++ {
		to += utf16Len([]rune{runes[to]})
	}
	return string(runes[from:to])
}

// tokensNamed returns every decoded token whose covered text equals name.
func tokensNamed(t *testing.T, text string, toks []semToken, name string) []semToken {
	t.Helper()
	out := []semToken{}
	for _, tok := range toks {
		if textAt(t, text, tok) == name {
			out = append(out, tok)
		}
	}
	return out
}

func TestSemanticTokensClassifyKinds(t *testing.T) {
	text := "import math\n\n## doc for pi\nval pi = 3.14159\nvar count = 0\ncount += 2\nprint(pi)\n"
	got := decodeSemantic(t, text)

	if len(got) == 0 {
		t.Fatal("nothing was encoded")
	}

	cases := []struct {
		name     string
		wantType string
		readonly bool
	}{
		{name: "import", wantType: "keyword"},
		{name: "math", wantType: "module"},
		{name: "pi", wantType: "constant", readonly: true},
		{name: "count", wantType: "variable"},
		{name: "print", wantType: "function"},
	}

	for _, tc := range cases {
		matches := tokensNamed(t, text, got, tc.name)
		if len(matches) == 0 {
			t.Errorf("no semantic token for %q in %+v", tc.name, got)
			continue
		}
		for _, m := range matches {
			if m.typ != tc.wantType {
				t.Errorf("%q on line %d typed %s, want %s", tc.name, m.line, m.typ, tc.wantType)
			}
			if got := m.mods&modReadonly != 0; got != tc.readonly {
				t.Errorf("%q on line %d readonly = %v, want %v", tc.name, m.line, got, tc.readonly)
			}
		}
	}

	// A comment on its own line is reported as a comment and nothing else.
	var comments []semToken
	for _, tok := range got {
		if tok.typ == "comment" {
			comments = append(comments, tok)
		}
	}
	if len(comments) != 1 {
		t.Fatalf("got %d comment tokens, want the one `## doc for pi` line", len(comments))
	}
	if comments[0].line != 2 {
		t.Errorf("comment reported on line %d, want line 2", comments[0].line)
	}
}

func TestSemanticTokensDeclarationSites(t *testing.T) {
	text := "fun helper(x) {\n\treturn x * 2\n}\n\nval total = helper(21)\n"
	got := decodeSemantic(t, text)

	// The name at its declaration carries the declaration modifier.
	defs := tokensNamed(t, text, got, "helper")
	if len(defs) != 2 {
		t.Fatalf("got %d tokens for `helper`, want the definition and the call: %+v", len(defs), got)
	}
	if defs[0].typ != "function" {
		t.Errorf("`helper` definition typed %s, want function", defs[0].typ)
	}
	if defs[0].mods&modDeclaration == 0 {
		t.Errorf("`helper` at its definition has no declaration modifier: %+v", defs[0])
	}

	// The call must not be marked as a definition.
	call := defs[1]
	if call.typ != "function" {
		t.Errorf("call to `helper` typed %s, want function", call.typ)
	}
	if call.mods&modDeclaration != 0 {
		t.Errorf("a plain call was marked as a declaration: %+v", call)
	}

	// Parameters bind in the signature and stay parameters where they are used.
	params := tokensNamed(t, text, got, "x")
	if len(params) != 2 {
		t.Fatalf("got %d tokens for `x`, want the binding and its use: %+v", len(params), got)
	}
	for i, p := range params {
		if p.typ != "parameter" {
			t.Errorf("`x` occurrence %d typed %s, want parameter", i, p.typ)
		}
	}
	if params[0].mods&modDeclaration == 0 {
		t.Errorf("the parameter binding lost its declaration modifier: %+v", params[0])
	}
	if params[1].mods&modDeclaration != 0 {
		t.Errorf("using a parameter is not itself a declaration: %+v", params[1])
	}

	for _, tok := range tokensNamed(t, text, got, "total") {
		if tok.typ != "constant" || tok.mods&modReadonly == 0 {
			t.Errorf("`total` typed %s with mods %d, want constant and readonly", tok.typ, tok.mods)
		}
	}
}

// Names that were never bound get nothing at all so the client's own base coloring
// stays in charge instead of every identifier being painted as something.
func TestSemanticTokensSkipUnknownNames(t *testing.T) {
	text := "print(zzz)\nval y = undefined_name + 1\n"
	got := decodeSemantic(t, text)

	for _, name := range []string{"zzz", "undefined_name"} {
		if matches := tokensNamed(t, text, got, name); len(matches) != 0 {
			t.Errorf("unbound name %q was typed as %q", name, matches[0].typ)
		}
	}

	// `y` is bound with val, so it has to be a read only constant.
	for _, tok := range tokensNamed(t, text, got, "y") {
		if tok.typ != "constant" || tok.mods&modReadonly == 0 {
			t.Errorf("`y` typed %s with mods %d, want constant and readonly", tok.typ, tok.mods)
		}
	}
}

// Loop variables bind inside blue's parentheses, and ranges must never be mistaken
// for member access even though both use dot characters.
func TestSemanticTokensLoopsAndRanges(t *testing.T) {
	text := "val xs = [1, 2, 3]\nfor (i in xs) {\n\tprint(i)\n}\nval a = 1\nval b = 5\nval r = a..b\nval q = 1..<5\n"
	got := decodeSemantic(t, text)

	is := tokensNamed(t, text, got, "i")
	if len(is) != 2 {
		t.Fatalf("got %d tokens for `i`, want the loop binding and its use: %+v", len(is), got)
	}
	if is[0].mods&modDeclaration == 0 {
		t.Errorf("the loop binding `i` has no declaration modifier: %+v", is[0])
	}
	if is[1].mods&modDeclaration != 0 {
		t.Errorf("using the loop variable is not itself a declaration: %+v", is[1])
	}
	for _, tok := range is {
		if tok.typ != "variable" {
			t.Errorf("loop variable typed %s, want variable", tok.typ)
		}
	}

	// Ranges must produce no member tokens.
	for _, line := range []int{6, 7} {
		for _, tok := range got {
			if tok.line != line {
				continue
			}
			if tok.typ == "property" || tok.typ == "function" {
				t.Errorf("line %d produced a %s token (%q); ranges must stay untouched", line, tok.typ, textAt(t, text, tok))
			}
		}
	}

	for _, name := range []string{"a", "b"} {
		for _, tok := range tokensNamed(t, text, got, name) {
			if tok.typ != "constant" {
				t.Errorf("`%s` on line %d typed %s, want constant", name, tok.line, tok.typ)
			}
		}
	}
}

// A dot that really is member access has to be reported as one.
func TestSemanticTokensMemberAccess(t *testing.T) {
	text := "import math\nimport json as j\nprint(math.sqrt(2))\nprint(math.pi)\nprint(j.encode({}))\n"
	got := decodeSemantic(t, text)

	mods := tokensNamed(t, text, got, "math")
	if len(mods) != 3 {
		t.Fatalf("got %d `math` tokens, want one per use: %+v", len(mods), got)
	}
	for _, tok := range mods {
		if tok.typ != "module" {
			t.Errorf("`math` on line %d typed %s, want module", tok.line, tok.typ)
		}
	}

	called := tokensNamed(t, text, got, "sqrt")
	if len(called) != 1 || called[0].typ != "function" {
		t.Errorf("`sqrt` (called) = %+v, want one function token", called)
	}
	prop := tokensNamed(t, text, got, "pi")
	if len(prop) != 1 || prop[0].typ != "property" {
		t.Errorf("`pi` after a dot = %+v, want one property token", prop)
	}
	// The alias appears at the import and again where the code reaches for it.
	alias := tokensNamed(t, text, got, "j")
	if len(alias) != 2 {
		t.Fatalf("got %d `j` tokens, want the import and the use", len(alias))
	}
	for _, tok := range alias {
		if tok.typ != "module" {
			t.Errorf("`j` on line %d typed %s, want module", tok.line, tok.typ)
		}
	}
	method := tokensNamed(t, text, got, "encode")
	if len(method) != 1 || method[0].typ != "function" {
		t.Errorf("`encode` (called) = %+v, want one function token", method)
	}
}

// Astral characters occupy two utf-16 code units each and clients count in code
// units, so everything after one has to stay aligned.
func TestSemanticTokensAstralColumns(t *testing.T) {
	text := "val s = \"a😀b\"\nval n = 7\nprint(s, n)\n"
	got := decodeSemantic(t, text)

	var str *semToken
	for i := range got {
		if got[i].typ == "string" && got[i].line == 0 {
			str = &got[i]
		}
	}
	if str == nil {
		t.Fatalf("no string token on line 0: %+v", got)
	}
	// quote, a, the emoji as two code units, b, closing quote
	if str.length != 6 {
		t.Errorf("string length = %d utf-16 units, want 6", str.length)
	}
	if textAt(t, text, *str) != "\"a😀b\"" {
		t.Errorf("string token covered %q", textAt(t, text, *str))
	}

	for _, name := range []string{"s", "n"} {
		matches := tokensNamed(t, text, got, name)
		if len(matches) == 0 {
			t.Errorf("names after an astral character were lost: %+v", got)
			continue
		}
		for _, m := range matches {
			if m.typ != "constant" || m.mods&modReadonly == 0 {
				t.Errorf("`%s` on line %d typed %s mods %d, want constant and readonly", name, m.line, m.typ, m.mods)
			}
		}
	}
}

// Unterminated strings, block comments and braces must not panic or run past the
// end of the buffer.
func TestSemanticTokensUnterminated(t *testing.T) {
	for _, text := range []string{
		"val s = \"never closed\nprint(1)\n",
		"### block comment never ended\n",
		"val x = `exec without backtick\n",
		"fun f( {\n\tprint(1)\n",
		"",
	} {
		got := decodeSemantic(t, text)
		for _, tok := range got {
			if tok.length <= 0 {
				t.Errorf("%q produced a token with length %d", text, tok.length)
			}
		}
	}
}

// The handler itself: open buffers get tuples, documents the session never saw get
// null so a client keeps whatever it already has instead of painting nothing.
func TestSemanticTokensFullHandler(t *testing.T) {
	s, uri := openDoc(t, "val pi = 3.14\nprint(pi)\n")

	res := s.semanticTokensFull(textDocumentIdentifier{URI: uri})
	list, ok := res.(semanticTokens)
	if !ok {
		t.Fatalf("handler returned %T, want semanticTokens", res)
	}
	if len(list.Data)%5 != 0 {
		t.Fatalf("data length %d is not a multiple of 5", len(list.Data))
	}
	if len(list.Data) < 5 {
		t.Fatalf("no tokens were returned for a buffer with code in it")
	}
	// The first tuple has no previous token, so both deltas must be zero and it
	// must land on the `val` keyword at column 0.
	if list.Data[0] != 0 || list.Data[1] != 0 {
		t.Errorf("first tuple starts at delta %d/%d, want 0/0", list.Data[0], list.Data[1])
	}
	if int(list.Data[3]) != semKeyword {
		t.Errorf("first token type = %d (%s), want keyword (%d)", list.Data[3], semLegendTypes[list.Data[3]], semKeyword)
	}

	if got := s.semanticTokensFull(textDocumentIdentifier{URI: "file:///tmp/nope.b"}); got != nil {
		t.Errorf("unknown document returned %v, want nil", got)
	}
}

// Every id and bit the encoder can emit must exist in the legend it advertises.
func TestSemanticTokensLegendCoversEverything(t *testing.T) {
	src := newDocSource("legend.b", "import math\n## c\nval v = \"s\" 42\nfun f(x) {\n\tprint(v)\n\tprint(math.pi)\n}\n")
	ix := buildIndex(src)
	ix.resolveExtents()
	data := encodeSemanticTokens(src, ix)

	for i := 0; i+4 < len(data); i += 5 {
		typeIdx := int(data[i+3])
		if typeIdx >= len(semLegendTypes) {
			t.Fatalf("tuple %d uses type id %d outside the legend (%d entries)", i/5, typeIdx, len(semLegendTypes))
		}
		mods := data[i+4]
		for bit := 0; bit < 32; bit++ {
			if mods&(1<<bit) != 0 && bit >= len(semLegendModifiers) {
				t.Fatalf("tuple %d sets modifier bit %d but the legend only has %d names", i/5, bit, len(semLegendModifiers))
			}
		}
	}

	// Sanity check that the well known ids line up with the advertised names.
	names := map[string]bool{}
	for _, n := range semLegendTypes {
		names[n] = true
	}
	for _, want := range []string{"comment", "keyword", "number", "string", "function", "variable", "constant", "parameter", "module", "property"} {
		if !names[want] {
			t.Errorf("legend is missing %q", want)
		}
	}
	if semString != 3 || semComment != 0 || semKeyword != 1 {
		t.Errorf("legend ids drifted from their declared order: comment=%d keyword=%d string=%d", semComment, semKeyword, semString)
	}
}
