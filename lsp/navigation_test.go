package lsp

import (
	"strings"
	"testing"
)

// colOf returns the column of the first occurrence of word on a line.
func colOf(text string, line int, word string) int {
	return strings.Index(strings.Split(text, "\n")[line], word)
}

// TestHighlightKinds checks that reads and writes are told apart. Operators are
// emitted one rune at a time by the scanner, so `==` must not be mistaken for
// an assignment and `+=` must not be mistaken for a read.
func TestHighlightKinds(t *testing.T) {
	text := strings.Join([]string{
		"var total = 0",
		"print(total)",
		"total = 5",
		"total += 2",
		"total == 9",
		"total <= 3",
		"total != 4",
	}, "\n")
	s, uri := openDoc(t, text)

	res := s.documentHighlight(documentHighlightParams{
		TextDocument: textDocumentIdentifier{URI: uri},
		Position:     position{Line: 1, Character: colOf(text, 1, "total") + 2},
	})
	list, ok := res.([]documentHighlight)
	if !ok {
		t.Fatalf("documentHighlight returned %T", res)
	}

	want := map[int]int{
		0: int(highlightWrite), // declaration binds
		1: int(highlightRead),  // plain read
		2: int(highlightWrite), // assignment
		3: int(highlightWrite), // compound assignment
		4: int(highlightRead),  // comparison, not a write
		5: int(highlightRead),  // comparison, not a write
		6: int(highlightRead),  // comparison, not a write
	}
	if len(list) != len(want) {
		t.Fatalf("got %d highlights, want %d: %+v", len(list), len(want), list)
	}
	lines := strings.Split(text, "\n")
	for _, h := range list {
		line := int(h.Range.Start.Line)
		if h.Range.Start.Line != h.Range.End.Line {
			t.Errorf("highlight %v spans lines, want a single line", h.Range)
			continue
		}
		raw := []rune(lines[line])
		if h.Range.Start.Character > h.Range.End.Character || h.Range.End.Character > len(raw) {
			t.Errorf("highlight %v out of range on line %d", h.Range, line)
			continue
		}
		if got := string(raw[h.Range.Start.Character:h.Range.End.Character]); got != "total" {
			t.Errorf("highlight %v covers %q, want the name itself", h.Range, got)
		}
		if wantKind, known := want[line]; !known {
			t.Errorf("unexpected highlight on line %d: %v", line, h.Range)
			continue
		} else if h.Kind != wantKind {
			t.Errorf("line %d kind = %d, want %d", line, h.Kind, wantKind)
		}
		delete(want, line)
	}
	for line := range want {
		t.Errorf("missing highlight for line %d", line)
	}
}

// A name that is only ever read must never be reported as a write.
func TestHighlightReadOnly(t *testing.T) {
	text := "val pi = 3.14\nprint(pi)\npi == 3\n"
	s, uri := openDoc(t, text)
	res := s.documentHighlight(documentHighlightParams{
		TextDocument: textDocumentIdentifier{URI: uri},
		Position:     position{Line: 1, Character: colOf(text, 1, "pi") + 1},
	})
	list, ok := res.([]documentHighlight)
	if !ok {
		t.Fatalf("documentHighlight returned %T", res)
	}
	for _, h := range list {
		if h.Range.Start.Line == 2 && h.Kind != int(highlightRead) {
			t.Errorf("`pi == 3` reported as kind %d, want read", h.Kind)
		}
	}
}

func TestDefinitionAndReferences(t *testing.T) {
	text := strings.Join([]string{
		"import math",
		"",
		"fun helper(x) {",
		"\treturn x * 2",
		"}",
		"",
		"print(helper(1))",
		"print(zzz)",
	}, "\n")
	s, uri := openDoc(t, text)

	t.Run("definition of a called function", func(t *testing.T) {
		res := s.definition(textDocumentPositionParams{
			TextDocument: textDocumentIdentifier{URI: uri},
			Position:     position{Line: 6, Character: colOf(text, 6, "helper") + 2},
		})
		locs, ok := res.([]location)
		if !ok {
			t.Fatalf("definition returned %T", res)
		}
		if len(locs) != 1 {
			t.Fatalf("got %d locations, want exactly one: %+v", len(locs), locs)
		}
		if locs[0].URI != uri {
			t.Errorf("uri = %q, want %q", locs[0].URI, uri)
		}
		if locs[0].Range.Start.Line != 2 {
			t.Errorf("range start line = %d, want the declaration on line 2", locs[0].Range.Start.Line)
		}
	})

	t.Run("references with includeDeclaration true keep the declaration", func(t *testing.T) {
		res := s.references(referenceParams{
			TextDocument: textDocumentIdentifier{URI: uri},
			Position:     position{Line: 6, Character: colOf(text, 6, "helper") + 2},
			Context:      &referenceContext{IncludeDeclaration: true},
		})
		locs, ok := res.([]location)
		if !ok {
			t.Fatalf("references returned %T", res)
		}
		if len(locs) != 2 {
			t.Fatalf("got %d references, want declaration plus call: %+v", len(locs), locs)
		}
		if locs[0].Range.Start.Line != 2 {
			t.Errorf("first reference should be the declaration on line 2, got %v", locs[0].Range)
		}
	})

	t.Run("references with includeDeclaration false drop the declaration", func(t *testing.T) {
		res := s.references(referenceParams{
			TextDocument: textDocumentIdentifier{URI: uri},
			Position:     position{Line: 6, Character: colOf(text, 6, "helper") + 2},
			Context:      &referenceContext{IncludeDeclaration: false},
		})
		locs, ok := res.([]location)
		if !ok {
			t.Fatalf("references returned %T", res)
		}
		if len(locs) != 1 {
			t.Fatalf("got %d references, want only the call site: %+v", len(locs), locs)
		}
		if locs[0].Range.Start.Line != 6 {
			t.Errorf("remaining reference is on line %d, want the call on line 6", locs[0].Range.Start.Line)
		}
	})

	t.Run("repeated names report every occurrence in the buffer", func(t *testing.T) {
		text := "fun f(alpha) {\n\treturn alpha + 1\n}\nval alpha = 9\nprint(alpha)\n"
		s2, uri2 := openDoc(t, text)
		res := s2.references(referenceParams{
			TextDocument: textDocumentIdentifier{URI: uri2},
			Position:     position{Line: 1, Character: 9},
		})
		locs, ok := res.([]location)
		if !ok {
			t.Fatalf("references returned %T", res)
		}
		// Matching is by name within the buffer, so all four spellings of `alpha`
		// come back. Blue's lexical scoping is not resolved here on purpose since
		// rename, which would need it, is not offered.
		if len(locs) != 4 {
			t.Fatalf("got %d occurrences of `alpha`, want 4: %+v", len(locs), locs)
		}
		for _, l := range locs {
			if l.URI != uri2 {
				t.Errorf("occurrence reported in a different document: %v", l)
			}
		}
	})

	t.Run("references include the declaration and the call (legacy no context)", func(t *testing.T) {
		res := s.references(referenceParams{
			TextDocument: textDocumentIdentifier{URI: uri},
			Position:     position{Line: 6, Character: colOf(text, 6, "helper") + 2},
		})
		locs, ok := res.([]location)
		if !ok {
			t.Fatalf("references returned %T", res)
		}
		if len(locs) != 2 {
			t.Fatalf("got %d references, want declaration plus call: %+v", len(locs), locs)
		}
		for _, l := range locs {
			if l.URI != uri {
				t.Errorf("reference in unknown document %q", l.URI)
			}
		}
	})

	t.Run("undeclared name has no definition", func(t *testing.T) {
		res := s.definition(textDocumentPositionParams{
			TextDocument: textDocumentIdentifier{URI: uri},
			Position:     position{Line: 7, Character: colOf(text, 7, "zzz") + 2},
		})
		switch v := res.(type) {
		case nil:
			return
		case []location:
			if len(v) != 0 {
				t.Errorf("got %d locations for an undeclared name: %+v", len(v), v)
			}
		default:
			t.Errorf("got %T, want nil or empty", res)
		}
	})

	t.Run("nothing is invented for names that do not exist", func(t *testing.T) {
		s2, uri2 := openDoc(t, "# just prose here\nprint(1)\n")
		for _, pos := range []position{{0, 3}, {0, 9}} {
			got := s2.definition(textDocumentPositionParams{
				TextDocument: textDocumentIdentifier{URI: uri2},
				Position:     pos,
			})
			if list, ok := got.([]location); ok && len(list) > 0 {
				t.Errorf("definition at %v returned locations %v for prose", pos, list)
			}
		}
	})

	t.Run("a name mentioned in a comment resolves when it exists", func(t *testing.T) {
		s3, uri3 := openDoc(t, "fun helper(x) {\n\tprint(x)\n}\n\n# see helper for doubling\n")
		line := 4
		col := strings.Index(strings.Split("fun helper(x) {\n\tprint(x)\n}\n\n# see helper for doubling\n", "\n")[line], "helper") + 2
		got := s3.definition(textDocumentPositionParams{
			TextDocument: textDocumentIdentifier{URI: uri3},
			Position:     position{Line: line, Character: col},
		})
		locs, ok := got.([]location)
		if !ok || len(locs) != 1 {
			t.Fatalf("definition of a name mentioned in a comment = %v, want the one declaration", got)
		}
		if locs[0].Range.Start.Line != 0 {
			t.Errorf("range start line = %d, want 0 where helper is declared", locs[0].Range.Start.Line)
		}
	})
}

func TestWorkspaceSymbol(t *testing.T) {
	s, _ := openDoc(t, "fun Alpha() {\n\tprint(1)\n}\n\nval beta = 2\n")

	names := func(query string) []string {
		res := s.workspaceSymbol(workspaceSymbolParams{Query: query})
		list, ok := res.([]symbolInformation)
		if !ok {
			t.Fatalf("workspaceSymbol returned %T", res)
		}
		out := []string{}
		for _, it := range list {
			out = append(out, it.Name)
		}
		return out
	}

	if got := names(""); len(got) == 0 {
		t.Errorf("empty query returned nothing, want every symbol")
	}
	if got := names("alph"); len(got) != 1 || got[0] != "Alpha" {
		t.Errorf("query %q = %v, want [Alpha]", "alph", got)
	}
	if got := names("BETA"); len(got) != 1 || got[0] != "beta" {
		t.Errorf("query %q = %v, want [beta] (matching is case insensitive)", "BETA", got)
	}
	if got := names("zzz"); len(got) != 0 {
		t.Errorf("query %q = %v, want nothing", "zzz", got)
	}
}

// The outline must nest declarations that live inside a function body rather than
// flattening everything to the top level.
func TestDocumentSymbolOutlineNesting(t *testing.T) {
	text := "import math\nimport json\n\nval total = 0\n\nfun outer(a, b) {\n\tval inner = a + b\n\tprint(inner)\n}\n\nprint(outer(1, 2))\n"
	s, uri := openDoc(t, text)

	out := s.documentSymbol(textDocumentIdentifier{URI: uri})
	syms, ok := out.([]documentSymbol)
	if !ok {
		t.Fatalf("documentSymbol returned %T, want []documentSymbol", out)
	}

	names := map[string]documentSymbol{}
	for _, sym := range syms {
		names[sym.Name] = sym
	}

	rootCount := len(syms)
	if rootCount != 4 {
		t.Fatalf("got %d top level symbols, want import math, import json, total and outer: %+v", rootCount, syms)
	}

	outer, ok := names["outer"]
	if !ok {
		t.Fatalf("no `outer` symbol in %+v", syms)
	}
	if outer.Kind != symKindFunction {
		t.Errorf("`outer` kind = %d, want Function (%d)", outer.Kind, symKindFunction)
	}
	if len(outer.Children) != 1 {
		t.Fatalf("`outer` has %d children, want the `inner` binding inside its body", len(outer.Children))
	}
	child := outer.Children[0]
	if child.Name != "inner" {
		t.Errorf("child = %q, want inner", child.Name)
	}
	if child.Kind != symKindConstant {
		t.Errorf("`inner` kind = %d, want Constant (%d)", child.Kind, symKindConstant)
	}

	// Selection ranges are the name itself so editors highlight the right span.
	sel := child.SelectionRange
	if sel.Start.Line != 6 || sel.End.Line != 6 {
		t.Errorf("`inner` selection spans lines %d-%d, want all on line 6", sel.Start.Line, sel.End.Line)
	}
	if int(sel.End.Character-sel.Start.Character) != len("inner") {
		t.Errorf("`inner` selection width = %d, want 5", int(sel.End.Character-sel.Start.Character))
	}

	// A module import is reported as a Module symbol.
	mod, ok := names["import math"]
	if !ok {
		t.Fatalf("no `import math` symbol in %+v", syms)
	}
	if mod.Kind != symKindModule {
		t.Errorf("`import math` kind = %d, want Module (%d)", mod.Kind, symKindModule)
	}
}
