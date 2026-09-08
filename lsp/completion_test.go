package lsp

import (
	"fmt"
	"testing"
)

// nopRW discards everything written so a session can be driven in memory.
type nopRW struct{}

func (nopRW) Read(p []byte) (int, error)  { return 0, fmt.Errorf("eof") }
func (nopRW) Write(p []byte) (int, error) { return len(p), nil }
func (nopRW) Close() error                { return nil }

// openDoc puts one buffer into a session the way didOpen does, without any
// transport, and hands back the session and the document uri.
func openDoc(t *testing.T, text string) (*session, string) {
	t.Helper()
	s := newSession(nopRW{}, Options{DiagnosticsDelay: 0})
	src := newDocSource("s.b", text)
	idx := buildIndex(src)
	idx.resolveExtents()
	uri := "file:///tmp/s.b"
	s.docs[uri] = &document{uri: uri, name: "s.b", src: src, index: idx}
	return s, uri
}

// labelItems lists completion labels.
func labelItems(list *completionList) []string {
	out := []string{}
	for _, it := range list.Items {
		out = append(out, it.Label)
	}
	return out
}

func hasLabel(list *completionList, want string) bool {
	for _, it := range list.Items {
		if it.Label == want {
			return true
		}
	}
	return false
}

func findLabel(list *completionList, label string) (completionItem, bool) {
	for _, it := range list.Items {
		if it.Label == label {
			return it, true
		}
	}
	return completionItem{}, false
}

// Completion must never offer names blue does not have. blue's standard library
// has no strings module and nothing called query, so neither may appear.
func TestCompletionOnlyRealNames(t *testing.T) {
	s, uri := openDoc(t, "import math\nprint(2 + 3)\n")

	res := s.completion(completionParams{TextDocument: textDocumentIdentifier{URI: uri}, Position: position{Line: 1, Character: 6}})
	list, ok := res.(*completionList)
	if !ok {
		t.Fatalf("completion returned %T, want *completionList", res)
	}

	if !hasLabel(list, "math") {
		t.Errorf("imported module math missing from completion: %v", labelItems(list))
	}
	for _, banned := range []string{"strings", "query"} {
		if hasLabel(list, banned) {
			t.Errorf("%q offered although blue has no such thing", banned)
		}
	}
}

// After `<module>.` only that module's real members may be offered, filtered by
// whatever has been typed after the dot.
func TestCompletionModuleMembers(t *testing.T) {
	text := "import math\nval r = 2\nmath.sq\nprint(r)\n"
	s, uri := openDoc(t, text)

	res := s.completion(completionParams{TextDocument: textDocumentIdentifier{URI: uri}, Position: position{Line: 2, Character: 8}})
	list, ok := res.(*completionList)
	if !ok {
		t.Fatalf("completion returned %T", res)
	}

	got := labelItems(list)
	for _, want := range []string{"sqrt", "sqrt2", "sqrt_e"} {
		if !hasLabel(list, want) {
			t.Errorf("missing std member %q of module math, got %v", want, got)
		}
	}
	for _, banned := range []string{"pi", "e"} {
		if hasLabel(list, banned) {
			t.Errorf("%q does not match the typed prefix %q but was offered: %v", banned, "sq", got)
		}
	}

	// A module that was never imported cannot be referenced at all in blue, so
	// nothing from it may be suggested.
	s2, uri2 := openDoc(t, "print(db.\n")
	res2 := s2.completion(completionParams{TextDocument: textDocumentIdentifier{URI: uri2}, Position: position{Line: 0, Character: 8}})
	list2 := res2.(*completionList)
	for _, banned := range []string{"exec", "query", "ping"} {
		if hasLabel(list2, banned) {
			t.Errorf("member %q of module `db` offered without an import: %v", banned, labelItems(list2))
		}
	}
}

// An import statement offers the modules that can actually be imported.
func TestCompletionImportCandidates(t *testing.T) {
	s, uri := openDoc(t, "import ma")
	res := s.completion(completionParams{TextDocument: textDocumentIdentifier{URI: uri}, Position: position{Line: 0, Character: 9}})
	list, ok := res.(*completionList)
	if !ok {
		t.Fatalf("completion returned %T", res)
	}
	if len(list.Items) != 1 || list.Items[0].Label != "math" {
		t.Errorf("`import ma` candidates = %v, want exactly [math]", labelItems(list))
	}

	s2, uri2 := openDoc(t, "import ")
	res2 := s2.completion(completionParams{TextDocument: textDocumentIdentifier{URI: uri2}, Position: position{Line: 0, Character: 7}})
	list2 := res2.(*completionList)
	for _, want := range []string{"math", "csv", "http"} {
		if !hasLabel(list2, want) {
			t.Errorf("std module %q missing from bare import candidates: %v", want, labelItems(list2))
		}
	}
	if hasLabel(list2, "strings") {
		t.Error(`"strings" offered although lib/std has no strings.b`)
	}
}

// Nothing may be offered while the cursor sits inside a string or a comment,
// but positions outside them still have to complete normally.
func TestCompletionSuppressedInLiterals(t *testing.T) {
	text := "val s = \"ab cd\"\n# comment word here\nprint(1)\n"
	s, uri := openDoc(t, text)

	for _, pos := range []position{{0, 9}, {0, 12}, {0, 14}, {1, 1}, {1, 5}, {1, 16}} {
		res := s.completion(completionParams{TextDocument: textDocumentIdentifier{URI: uri}, Position: pos})
		list, ok := res.(*completionList)
		if !ok {
			t.Fatalf("pos %v returned %T", pos, res)
		}
		if len(list.Items) != 0 {
			t.Errorf("completion inside literal/comment at %v = %v, want none", pos, labelItems(list))
		}
	}

	for _, pos := range []position{{0, 0}, {0, 8}, {1, 0}, {2, 6}} {
		res := s.completion(completionParams{TextDocument: textDocumentIdentifier{URI: uri}, Position: pos})
		list, ok := res.(*completionList)
		if !ok {
			t.Fatalf("pos %v returned %T", pos, res)
		}
		if len(list.Items) == 0 {
			t.Errorf("completion outside literals at %v returned nothing, want candidates", pos)
		}
	}

	// An unterminated string must not swallow the rest of the buffer either.
	s2, uri2 := openDoc(t, "val t = \"unclosed\nprint(2)\n")
	res2 := s2.completion(completionParams{TextDocument: textDocumentIdentifier{URI: uri2}, Position: position{Line: 1, Character: 6}})
	if _, ok := res2.(*completionList); !ok {
		t.Fatalf("returned %T", res2)
	}
}

// Callables declared in the buffer get argument placeholders only when the
// client declared snippet support, and plain text otherwise.
func TestCompletionSnippetGating(t *testing.T) {
	text := "fun greet(name, greeting) {\n\tprint(name)\n}\n\ngre\n"

	run := func(snippets bool) *completionList {
		s := newSession(nopRW{}, Options{DiagnosticsDelay: 0})
		s.clientSnippets = snippets
		src := newDocSource("greet.b", text)
		uri := "file:///tmp/greet.b"
		s.docs[uri] = &document{uri: uri, name: "greet.b", src: src, index: buildIndex(src)}
		return s.completion(completionParams{TextDocument: textDocumentIdentifier{URI: uri}, Position: position{Line: 4, Character: 3}}).(*completionList)
	}

	plain := run(false)
	item, ok := findLabel(plain, "greet")
	if !ok {
		t.Fatalf("no completion for `gre`, got %v", labelItems(plain))
	}
	if item.InsertText != "greet()" {
		t.Errorf("without snippet support insertText = %q, want %q", item.InsertText, "greet()")
	}
	if item.InsertTextFormat == completionSnippet {
		t.Error("snippet format advertised although snippets are disabled")
	}

	snips := run(true)
	item2, ok := findLabel(snips, "greet")
	if !ok {
		t.Fatalf("no completion for `gre` with snippets on, got %v", labelItems(snips))
	}
	want := "greet(${1:name}, ${2:greeting})"
	if item2.InsertText != want {
		t.Errorf("with snippet support insertText = %q, want %q", item2.InsertText, want)
	}
	if item2.InsertTextFormat != completionSnippet {
		t.Errorf("with snippet support insertTextFormat = %d, want %d", item2.InsertTextFormat, completionSnippet)
	}
}
