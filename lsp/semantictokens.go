package lsp

import "unicode"

// semLegendTypes are the semantic token type names this server emits, in the exact
// order their ids are used by the encoder below. Only LSP standard names are used
// so editor themes color them without any extra configuration.
var semLegendTypes = []string{
	"comment",
	"keyword",
	"number",
	"string",
	"function",
	"variable",
	"constant",
	"parameter",
	"module",
	"property",
}

const (
	semComment = iota
	semKeyword
	semNumber
	semString
	semFunction
	semVariable
	semConstant
	semParameter
	semModule
	semProperty
)

// semLegendModifiers is indexed the same way, so bit positions must stay in sync.
var semLegendModifiers = []string{
	"declaration",
	"readonly",
}

const (
	modDeclaration uint32 = 1 << iota
	modReadonly
)

// semanticTokens holds one textDocument/semanticTokens result: five integers per
// token, as documented in the protocol.
type semanticTokens struct {
	Data []uint32 `json:"data"`
}

// semanticTokensFull answers a full range request. Documents that are not open, or
// that have nothing worth painting, get null rather than an empty list so a client
// does not clear tokens it could still show.
func (s *session) semanticTokensFull(p textDocumentIdentifier) any {
	doc := s.documentFor(p.URI)
	if doc == nil || doc.src == nil || doc.index == nil {
		return nil
	}

	data := encodeSemanticTokens(doc.src, doc.index)
	if len(data) == 0 {
		return nil
	}
	return semanticTokens{Data: data}
}

// encodeSemanticTokens walks an already indexed buffer and emits the relative
// tuples clients decode. Every offset is derived from docSource so columns are
// UTF-16 code units even where blue's own lexer positions would not be.
func encodeSemanticTokens(src *docSource, ix *fileIndex) []uint32 {
	if src == nil || ix == nil || len(ix.tokens) == 0 {
		return nil
	}

	sites := map[int]*declaration{}
	byName := map[string][]*declaration{}
	for _, d := range ix.decls {
		sites[d.nameStart] = d
		byName[d.name] = append(byName[d.name], d)
	}

	modules := map[string]bool{}
	for _, imp := range ix.imports {
		if imp.alias != "" {
			modules[imp.alias] = true
		}
	}

	out := make([]uint32, 0, len(ix.tokens)*5)
	prevLine := 0
	prevChar := 0

	for i, tok := range ix.tokens {
		typ, mods, ok := classifyToken(ix.tokens, i, sites, byName, modules)
		if !ok {
			continue
		}

		length := utf16Len(src.runes[tok.start:tok.end])
		if length == 0 {
			continue
		}

		pos := src.positionOf(tok.start)
		startChar := int(pos.Character)
		if int(pos.Line) == prevLine {
			startChar -= prevChar
		}

		out = append(out, uint32(int(pos.Line)-prevLine), uint32(startChar), uint32(length), uint32(typ), mods)
		prevLine = int(pos.Line)
		prevChar = int(pos.Character)
	}

	return out
}

// classifyToken maps one token onto the legend. Tokens that carry no meaning of
// their own are skipped so the client's base coloring stays in charge, which is why
// punctuation and unbound names produce nothing here.
func classifyToken(tokens []scanToken, i int, sites map[int]*declaration, byName map[string][]*declaration, modules map[string]bool) (int, uint32, bool) {
	tok := tokens[i]

	switch tok.kind {
	case kComment:
		return semComment, 0, true

	case kWord:
		return semKeyword, 0, true

	case kString:
		runes := []rune(tok.text)
		if len(runes) > 0 && unicode.IsDigit(runes[0]) {
			return semNumber, 0, true
		}
		return semString, 0, true

	case kIdent:
		return classifyIdent(tokens, i, sites, byName, modules)
	}

	return 0, 0, false
}

// classifyIdent decides between a binding site, a member, a call and a plain use of
// an already bound name. Every decision comes from tokens actually in the buffer;
// no names, types or signatures are invented here.
func classifyIdent(tokens []scanToken, i int, sites map[int]*declaration, byName map[string][]*declaration, modules map[string]bool) (int, uint32, bool) {
	tok := tokens[i]
	name := tok.text

	// A binding site keeps the kind blue used to bind it.
	if d, ok := sites[tok.start]; ok {
		switch d.kind {
		case declFunction:
			return semFunction, modDeclaration, true
		case declConstant:
			return semConstant, modDeclaration | modReadonly, true
		case declParam:
			return semParameter, modDeclaration, true
		default:
			return semVariable, modDeclaration, true
		}
	}

	// The scanner emits dots one at a time, so member access is only what follows a
	// run of exactly one dot. Ranges such as a..b or 1..<5 leave two or three dots
	// behind and must not turn the name after them into a property.
	if dots := dotRunBefore(tokens, i); dots == 1 {
		if nextToken(tokens, i+1) == "(" {
			return semFunction, 0, true
		}
		return semProperty, 0, true
	}

	// A name reachable as a module is painted as one wherever the code uses it that
	// way, so `import json as j` colors the alias rather than the original name.
	if modules[name] {
		return semModule, 0, true
	}

	if nextToken(tokens, i+1) == "(" {
		return semFunction, 0, true
	}

	d := visibleDeclaration(byName[name], tok.start)
	if d == nil {
		return 0, 0, false
	}

	switch d.kind {
	case declParam:
		return semParameter, 0, true
	case declConstant:
		return semConstant, modReadonly, true
	default:
		return semVariable, 0, true
	}
}

// dotRunBefore counts consecutive '.' punctuation tokens immediately before index i.
func dotRunBefore(tokens []scanToken, i int) int {
	n := 0
	for j := i - 1; j >= 0 && tokens[j].kind == kPunct && tokens[j].text == "."; j-- {
		n++
	}
	return n
}

// nextToken returns the text of the token at i, or "" past the end.
func nextToken(tokens []scanToken, i int) string {
	if i >= len(tokens) {
		return ""
	}
	return tokens[i].text
}

// visibleDeclaration picks the binding a use refers to. Candidates are in source
// order, so walking backwards finds the closest binding first: parameters belong to
// their own function, and once that function is over the file scope binding of the
// same name takes over again. Names that were never bound return nil.
func visibleDeclaration(candidates []*declaration, offset int) *declaration {
	if len(candidates) == 0 {
		return nil
	}

	// The closest preceding binding wins. That is what makes a parameter used inside
	// its own body resolve, and what makes a later top level rebinding take over.
	for i := len(candidates) - 1; i >= 0; i-- {
		d := candidates[i]
		if d.nameStart >= offset {
			continue
		}
		return d
	}
	return nil
}
