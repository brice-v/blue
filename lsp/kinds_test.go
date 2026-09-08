package lsp

import (
	"strings"
	"testing"
)

func findSymbol(syms []documentSymbol, name string) (documentSymbol, bool) {
	for _, sym := range syms {
		if sym.Name == name {
			return sym, true
		}
		if len(sym.Children) > 0 {
			if found, ok := findSymbol(sym.Children, name); ok {
				return found, true
			}
		}
	}
	return documentSymbol{}, false
}

// CompletionItemKind and SymbolKind share nothing but a name. Completion has to
// use the CompletionItemKind numbers or editors render the wrong icon.
func TestCompletionKindsMatchSpec(t *testing.T) {
	text := strings.Join([]string{
		"import math",
		"fun greet(name) {",
		"\tprint(name)",
		"}",
		"var count = 0",
		"val total = 7",
	}, "\n")
	s, uri := openDoc(t, text)

	res := s.completion(completionParams{TextDocument: textDocumentIdentifier{URI: uri}, Position: position{Line: 4, Character: 0}})
	list, ok := res.(*completionList)
	if !ok {
		t.Fatalf("completion returned %T", res)
	}

	checks := []struct {
		label string
		kind  int
	}{
		{"greet", kindFunction},
		{"count", kindVariable},
		{"total", kindConstant},
		{"math", kindModule},
	}
	for _, c := range checks {
		item, found := findLabel(list, c.label)
		if !found {
			t.Errorf("%q missing from completion list", c.label)
			continue
		}
		if item.Kind != c.kind {
			t.Errorf("%q kind = %d, want CompletionItemKind %d", c.label, item.Kind, c.kind)
		}
	}

	for _, kw := range []string{"val", "var", "fun", "import", "for", "if"} {
		item, found := findLabel(list, kw)
		if !found {
			t.Errorf("keyword %q missing from completion list", kw)
			continue
		}
		if item.Kind != kindKeyword {
			t.Errorf("keyword %q kind = %d, want %d", kw, item.Kind, kindKeyword)
		}
	}
}

// Document symbols keep the SymbolKind numbering.
func TestSymbolKindsMatchSpec(t *testing.T) {
	text := strings.Join([]string{
		"import math",
		"var count = 0",
		"val total = 7",
		"fun greet(name) {",
		"\tprint(name)",
		"}",
	}, "\n")
	s, uri := openDoc(t, text)

	out := s.documentSymbol(textDocumentIdentifier{URI: uri})
	syms, ok := out.([]documentSymbol)
	if !ok {
		t.Fatalf("documentSymbol returned %T", out)
	}

	checks := []struct {
		name string
		kind int
	}{
		{"import math", symKindModule},
		{"count", symKindVariable},
		{"total", symKindConstant},
		{"greet", symKindFunction},
	}
	for _, c := range checks {
		sym, found := findSymbol(syms, c.name)
		if !found {
			t.Errorf("symbol %q missing from outline: %v", c.name, symbolNames(syms))
			continue
		}
		if sym.Kind != c.kind {
			t.Errorf("symbol %q kind = %d, want SymbolKind %d", c.name, sym.Kind, c.kind)
		}
	}
}

func symbolNames(syms []documentSymbol) []string {
	out := []string{}
	for _, sym := range syms {
		out = append(out, sym.Name)
		if len(sym.Children) > 0 {
			for _, child := range symbolNames(sym.Children) {
				out = append(out, "\t"+child)
			}
		}
	}
	return out
}
