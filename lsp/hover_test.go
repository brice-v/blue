package lsp

import (
	"strings"
	"testing"
)

// hoverAt runs hover on the nth occurrence of word in the given line.
func hoverAt(t *testing.T, s *session, uri, text string, line int, word string) any {
	t.Helper()
	col := strings.Index(strings.Split(text, "\n")[line], word)
	if col < 0 {
		t.Fatalf("word %q not found on line %d", word, line)
	}
	return s.hover(textDocumentHoverParams{
		TextDocument: textDocumentIdentifier{URI: uri},
		Position:     position{Line: line, Character: col},
	})
}

func hoverValue(res any) (string, bool) {
	hr, ok := res.(*hoverResult)
	if !ok || hr == nil {
		return "", false
	}
	return hr.Contents.Value, true
}

// Hovering a plain name shows the statement that bound it.
func TestHoverDeclaration(t *testing.T) {
	text := "val pi = 3.14\nvar count = 0\nprint(pi)\nprint(count)\n"
	s, uri := openDoc(t, text)

	got, ok := hoverValue(hoverAt(t, s, uri, text, 2, "pi"))
	if !ok {
		t.Fatalf("hover over a used constant returned nothing")
	}
	if !strings.Contains(got, "val pi = 3.14") {
		t.Errorf("hover = %q, want the binding statement", got)
	}

	got2, ok := hoverValue(hoverAt(t, s, uri, text, 3, "count"))
	if !ok {
		t.Fatalf("hover over a used variable returned nothing")
	}
	if !strings.Contains(got2, "var count = 0") {
		t.Errorf("hover = %q, want the binding statement", got2)
	}
}

// A function header must be shown exactly as written, because blue has no
// return types and hand written parameter types would be made up.
func TestHoverFunctionHeaderVerbatim(t *testing.T) {
	text := "fun connect(host, port = 5432, dbname = \":memory:\") {\n\tprint(host)\n}\n\nprint(connect)\n"
	s, uri := openDoc(t, text)

	got, ok := hoverValue(hoverAt(t, s, uri, text, 4, "connect"))
	if !ok {
		t.Fatalf("hover over a declared function returned nothing")
	}
	want := "fun connect(host, port = 5432, dbname = \":memory:\")"
	if !strings.Contains(got, want) {
		t.Errorf("hover = %q, want the verbatim header %q", got, want)
	}
	if strings.Contains(got, "->") {
		t.Errorf("hover invented a return type: %q", got)
	}

	// Hovering on the declaration itself must work as well.
	if _, ok := hoverValue(hoverAt(t, s, uri, text, 0, "connect")); !ok {
		t.Error("hover over the function's own name returned nothing")
	}
}

// Docstrings are `##` lines standing alone above the declaration or right after
// its opening brace. A trailing comment on another statement's line belongs to
// that statement and must not leak into whatever comes next.
func TestHoverDocstringRules(t *testing.T) {
	text := strings.Join([]string{
		"val e = 2.718 # note about e",
		"## documents pi",
		"## second doc line",
		"val pi = 3.14",
		"print(pi)",
		"",
		"fun damped(x) {",
		"\t## decays with noise",
		"\tx",
		"}",
		"print(damped)",
	}, "\n")
	s, uri := openDoc(t, text)

	got, ok := hoverValue(hoverAt(t, s, uri, text, 4, "pi"))
	if !ok {
		t.Fatalf("hover over pi returned nothing")
	}
	if !strings.Contains(got, "documents pi") || !strings.Contains(got, "second doc line") {
		t.Errorf("own-line docstrings above a val were not shown: %q", got)
	}

	got2, ok := hoverValue(hoverAt(t, s, uri, text, 10, "damped"))
	if !ok {
		t.Fatalf("hover over damped returned nothing")
	}
	if !strings.Contains(got2, "decays with noise") {
		t.Errorf("in-body docstring was not shown: %q", got2)
	}
	if strings.Contains(got2, "note about e") {
		t.Errorf("trailing comment of an earlier statement leaked into %q", got2)
	}
}

// Nothing may be invented: unknown names and members blue does not have get no
// hover at all.
func TestHoverUnknown(t *testing.T) {
	text := "import math\nprint(nope)\nprint(math.PI)\n"
	s, uri := openDoc(t, text)

	if res := hoverAt(t, s, uri, text, 1, "nope"); res != nil {
		t.Errorf("hover over an undeclared name returned %v, want nil", res)
	}

	// math.b defines pi in lower case only, so hovering PI must find nothing.
	if res := hoverAt(t, s, uri, text, 2, "PI"); res != nil {
		t.Errorf("hover over math.PI returned %v, want nil because blue has no PI", res)
	}
}
