package lsp

import (
	"blue/token"
	"strings"
)

// Symbol kinds mirroring the LSP SymbolKind numeric codes so they can be sent
// straight out to clients.
const (
	symKindFile     = 1
	symKindModule   = 2
	symKindFunction = 12
	symKindVariable = 13
	symKindConstant = 14
)

// declKind is what bound a name in blue source.
type declKind int

const (
	declFunction declKind = iota
	declVariable
	declConstant
	declParam
	declLoopVar
	declCatchVar
)

// completionKind maps a declaration to a CompletionItemKind. Those numbers are
// not the SymbolKind ones, so completions must never use symbolKind().
func (k declKind) completionKind() int {
	switch k {
	case declFunction:
		return kindFunction
	case declConstant:
		return kindConstant
	default:
		return kindVariable
	}
}

func (k declKind) symbolKind() int {
	switch k {
	case declFunction:
		return symKindFunction
	case declConstant:
		return symKindConstant
	default:
		return symKindVariable
	}
}

// declaration is a name bound in blue source together with its extent.
type declaration struct {
	name      string
	kind      declKind
	start     int // rune offset of the binding keyword (fun, val, for, ...)
	nameStart int // rune offset of the name itself
	nameEnd   int // rune offset just past the name
	end       int // exclusive rune offset, -1 until extents are resolved
	bodyOpen  int // token index of the body '{' for functions, -1 otherwise
	detail    string
	params    []string
	doc       string
}

// importEntry is one `import`/`from ... import` statement.
type importEntry struct {
	path  string   // dotted module path such as "foo.bar"
	alias string   // name the module is reachable as (alias or last segment)
	names []string // explicit names for `from path import {a, b}`
	start int
	end   int
}

// fileIndex holds what the LSP knows about a document buffer. It comes from the
// light weight scanner instead of the real parser so that features keep working
// while a buffer has syntax errors, which is the normal state of code being
// typed. Call resolveExtents before querying it.
type fileIndex struct {
	src     *docSource
	tokens  []scanToken
	decls   []*declaration
	imports []importEntry
}

func isPunct(t scanToken, text string) bool {
	return t.kind == kPunct && t.text == text
}

// matchDelim returns the index of the token that closes the group opened at
// openIdx, or -1 when the group never closes.
func matchDelim(tokens []scanToken, openIdx int, open, close string) int {
	depth := 0
	for i := openIdx; i < len(tokens); i++ {
		if isPunct(tokens[i], open) {
			depth++
		} else if isPunct(tokens[i], close) {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// segments splits a token slice on commas that sit at nesting depth zero.
func segments(tokens []scanToken) [][]scanToken {
	out := [][]scanToken{}
	depth := 0
	current := []scanToken{}
	for _, t := range tokens {
		switch {
		case isPunct(t, "(") || isPunct(t, "[") || isPunct(t, "{"):
			depth++
		case isPunct(t, ")") || isPunct(t, "]") || isPunct(t, "}"):
			depth--
		case depth == 0 && isPunct(t, ","):
			out = append(out, current)
			current = []scanToken{}
			continue
		}
		current = append(current, t)
	}
	if len(current) > 0 {
		out = append(out, current)
	}
	return out
}

// bindingNames returns the names bound by one declaration fragment such as
// `x`, `x : int` or `[a, b]`. Type annotations and default values are dropped.
func bindingNames(tokens []scanToken) []scanToken {
	out := []scanToken{}
	for _, seg := range segments(tokens) {
		if len(seg) == 0 {
			continue
		}
		head := seg[0]
		if head.kind == kPunct && (head.text == "[" || head.text == "{") {
			// Destructuring, every identifier inside is bound.
			closeIdx := -1
			for i, t := range seg {
				if isPunct(t, "=") {
					closeIdx = i
					break
				}
			}
			if closeIdx < 0 {
				closeIdx = len(seg)
			}
			for _, t := range seg[1:closeIdx] {
				if t.kind == kIdent {
					out = append(out, t)
				}
			}
			continue
		}
		if head.kind != kIdent {
			continue
		}
		out = append(out, head)
	}
	return out
}

// functionDetail is the parameter list of a declaration taken straight out of the
// source, so defaults like `db_name=":memory:"` keep their exact text. Blue has
// no return types, so nothing else belongs in what follows a function name.
func functionDetail(src *docSource, tokens []scanToken, openParen, closeParen int) string {
	return string(src.runes[tokens[openParen].start:tokens[closeParen].end])
}

// docstringFor collects the `##` comment lines blue uses as a docstring. They
// are written either above the declaration or right after its opening brace, and
// both are picked up here.
// ownLine reports whether the rune at idx is the first non whitespace thing on
// its line. Only comments like that can document what follows them.
func ownLine(rs []rune, idx int) bool {
	for i := idx - 1; i >= 0; i-- {
		switch rs[i] {
		case '\n':
			return true
		case ' ', '\t':
			continue
		default:
			return false
		}
	}
	return true
}

// docstringFor collects the `##` comment lines that document a declaration.
// Only comments standing alone on their own line count, and collection stops at
// the first thing that is not such a comment. A trailing comment after code
// belongs to that code, not to whatever comes next, which matters because blue
// source puts notes like `# https://oeis.org/A000796` at the end of a line.
func docstringFor(rs []rune, tokens []scanToken, declStart int, bodyOpenTok int) string {
	lines := []string{}

	for i := declStart - 1; i >= 0; i-- {
		if tokens[i].kind != kComment {
			break
		}
		if !ownLine(rs, tokens[i].start) {
			break
		}
		text := strings.TrimRight(tokens[i].text, " \t")
		if !strings.HasPrefix(text, "#") {
			break
		}
		lines = append([]string{commentBody(text)}, lines...)
	}

	// A function literal can also carry its docs on the lines right after the
	// opening brace, which is the style blue's own standard library uses.
	if bodyOpenTok >= 0 && bodyOpenTok+1 < len(tokens) {
		for i := bodyOpenTok + 1; i < len(tokens); i++ {
			t := tokens[i]
			if t.kind != kComment {
				break
			}
			text := strings.TrimRight(t.text, " \t")
			if !strings.HasPrefix(text, "##") {
				continue
			}
			lines = append(lines, commentBody(text))
		}
	}

	picked := []string{}
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		picked = append(picked, l)
	}
	return strings.Join(picked, "\n")
}

// commentBody strips the leading comment markers from a comment token so the
// text can be shown on its own.
func commentBody(text string) string {
	return strings.TrimSpace(strings.TrimLeft(text, "#"))
}

// buildIndex scans a document buffer and records every name it binds.
func buildIndex(src *docSource) *fileIndex {
	ix := &fileIndex{src: src, tokens: tokenize(src.runes)}
	toks := ix.tokens

	// The loop advances by hand because several cases consume a whole statement.
	for i := 0; i < len(toks); {
		t := toks[i]
		if t.kind != kWord {
			i++
			continue
		}

		switch t.text {
		case "fun":
			j := i + 1
			if j >= len(toks) || toks[j].kind != kIdent {
				i++
				continue
			}
			nameTok := j
			j++
			if j >= len(toks) || !isPunct(toks[j], "(") {
				i++
				continue
			}
			openParen := j
			closeParen := matchDelim(toks, openParen, "(", ")")
			if closeParen < 0 {
				i++
				continue
			}
			k := closeParen + 1
			for k < len(toks) && !isPunct(toks[k], "{") {
				k++
			}
			if k >= len(toks) {
				i = closeParen + 1
				continue
			}

			paramTokens := toks[openParen+1 : closeParen]
			decl := &declaration{
				name:      toks[nameTok].text,
				kind:      declFunction,
				start:     t.start,
				nameStart: toks[nameTok].start,
				nameEnd:   toks[nameTok].end,
				end:       -1,
				bodyOpen:  k,
				detail:    functionDetail(src, toks, openParen, closeParen),
				params:    namesOf(paramTokens),
			}
			ix.decls = append(ix.decls, decl)

			for _, name := range bindingNames(paramTokens) {
				ix.decls = append(ix.decls, &declaration{
					name:      name.text,
					kind:      declParam,
					start:     decl.start,
					nameStart: name.start,
					nameEnd:   name.end,
				})
			}

			bodyClose := matchDelim(toks, k, "{", "}")
			if bodyClose < 0 {
				// The body never closes, which is the normal state of a buffer
				// someone is still typing. Everything after the brace belongs to it,
				// but scanning continues so bindings typed inside are still found.
				decl.end = len(src.runes)
				i = k + 1
				continue
			}
			decl.end = toks[bodyClose].end
			decl.doc = docstringFor(src.runes, toks, i, k)
			// Keep scanning inside the body so nested declarations are found.
			i = k + 1

		case "var", "val", "const":
			j := i + 1
			end := j
			for end < len(toks) && !isPunct(toks[end], "=") && !statementStartsHere(toks[end]) {
				end++
			}
			kind := declVariable
			if t.text != "var" {
				kind = declConstant
			}
			for _, name := range bindingNames(toks[j:end]) {
				ix.decls = append(ix.decls, &declaration{
					name:      name.text,
					kind:      kind,
					start:     t.start,
					nameStart: name.start,
					nameEnd:   name.end,
					end:       -1,
					doc:       docstringFor(src.runes, toks, i, -1),
				})
			}
			i++

		case "for":
			j := i + 1
			if j >= len(toks) || !isPunct(toks[j], "(") {
				i++
				continue
			}
			closeParen := matchDelim(toks, j, "(", ")")
			if closeParen < 0 {
				i++
				continue
			}
			for _, name := range loopBindingNames(toks[j+1 : closeParen]) {
				ix.decls = append(ix.decls, &declaration{
					name:      name.text,
					kind:      declLoopVar,
					start:     t.start,
					nameStart: name.start,
					nameEnd:   name.end,
				})
			}
			// The header has no other declarations in it.
			i = closeParen + 1

		case "catch":
			j := i + 1
			if j >= len(toks) || !isPunct(toks[j], "(") {
				i++
				continue
			}
			closeParen := matchDelim(toks, j, "(", ")")
			if closeParen < 0 {
				i++
				continue
			}
			for _, name := range bindingNames(toks[j+1 : closeParen]) {
				ix.decls = append(ix.decls, &declaration{
					name:      name.text,
					kind:      declCatchVar,
					start:     t.start,
					nameStart: name.start,
					nameEnd:   name.end,
				})
			}
			i = closeParen + 1

		case "import", "from":
			entry, next, ok := parseImport(toks, i)
			if !ok {
				i++
				continue
			}
			ix.imports = append(ix.imports, entry)
			// `next` is the first token that was not part of this statement, so
			// continue there without skipping it.
			i = next

		default:
			i++
		}
	}

	return ix
}

// namesOf maps tokens to their text.
func namesOf(tokens []scanToken) []string {
	out := []string{}
	for _, t := range bindingNames(tokens) {
		out = append(out, t.text)
	}
	return out
}

// loopBindingNames returns the identifiers bound before `in` in a for header.
func loopBindingNames(tokens []scanToken) []scanToken {
	for i, t := range tokens {
		if t.kind == kWord && t.text == "in" {
			return bindingNames(tokens[:i])
		}
	}
	return nil
}

// statementStartsHere reports whether a token begins a new statement, which ends
// the binding list of the preceding `var`/`val`.
func statementStartsHere(t scanToken) bool {
	if isPunct(t, ";") || isPunct(t, "}") || isPunct(t, ")") {
		return true
	}
	if t.kind == kWord {
		switch t.text {
		case "fun", "var", "val", "const", "import", "from", "for", "return", "if", "try", "match":
			return true
		}
	}
	return false
}

// parseImport parses an import statement starting at the `import` or `from`
// keyword and returns the entry plus the token index to continue scanning from.
func parseImport(tokens []scanToken, idx int) (importEntry, int, bool) {
	fromStmt := tokens[idx].text == "from"
	entry := importEntry{start: tokens[idx].start}

	i := idx + 1
	if i >= len(tokens) || tokens[i].kind != kIdent {
		return importEntry{}, idx, false
	}

	var path strings.Builder
	path.WriteString(tokens[i].text)
	i++
	for i+1 < len(tokens) && isPunct(tokens[i], ".") && tokens[i+1].kind == kIdent {
		path.WriteString(".")
		path.WriteString(tokens[i+1].text)
		i += 2
	}
	entry.path = path.String()

	if !fromStmt {
		// import path [as alias]
		if i+1 < len(tokens) && tokens[i].kind == kWord && tokens[i].text == "as" && tokens[i+1].kind == kIdent {
			entry.alias = tokens[i+1].text
			i += 2
		}
		if entry.alias == "" {
			entry.alias = lastSegment(entry.path)
		}
		if i < len(tokens) {
			entry.end = tokens[i].start
		} else {
			entry.end = tokens[len(tokens)-1].end
		}
		return entry, i, true
	}

	// from path import * | {names} | name, name ...
	for i < len(tokens) && tokens[i].kind == kPunct && !isPunct(tokens[i], "{") && !isPunct(tokens[i], ";") {
		i++
	}
	if i >= len(tokens) || tokens[i].kind != kWord || tokens[i].text != "import" {
		return importEntry{}, idx, false
	}
	i++

	switch {
	case i < len(tokens) && isPunct(tokens[i], "*"):
		entry.names = []string{"*"}
		i++
	case i < len(tokens) && isPunct(tokens[i], "{"):
		closeBrace := matchDelim(tokens, i, "{", "}")
		if closeBrace < 0 {
			return importEntry{}, idx, false
		}
		for _, t := range tokens[i+1 : closeBrace] {
			if t.kind == kIdent {
				entry.names = append(entry.names, t.text)
			}
		}
		i = closeBrace + 1
	default:
		for i < len(tokens) && (tokens[i].kind == kIdent || isPunct(tokens[i], ",")) {
			if tokens[i].kind == kIdent {
				entry.names = append(entry.names, tokens[i].text)
			}
			i++
		}
	}

	if entry.alias == "" {
		entry.alias = lastSegment(entry.path)
	}
	if i < len(tokens) {
		entry.end = tokens[i].start
	} else if len(tokens) > 0 {
		entry.end = tokens[len(tokens)-1].end
	}
	return entry, i, true
}

func lastSegment(dotted string) string {
	if idx := strings.LastIndex(dotted, "."); idx >= 0 {
		return dotted[idx+1:]
	}
	return dotted
}

// resolveExtents fills in the extent of every declaration that was left open:
// each one reaches to whatever comes next inside its enclosing function, or to
// the end of the buffer while at top level.
func (ix *fileIndex) resolveExtents() {
	var fnStack []int
	for i, d := range ix.decls {
		for len(fnStack) > 0 && ix.decls[fnStack[len(fnStack)-1]].end <= d.start {
			fnStack = fnStack[:len(fnStack)-1]
		}

		// Bindings only ever cover the name itself, which is everything hover and
		// highlight need, so they cannot stretch past it.
		switch d.kind {
		case declParam, declLoopVar, declCatchVar:
			d.end = d.nameEnd
			continue
		}

		if d.kind == declFunction {
			if d.end < 0 || d.end <= d.nameEnd {
				d.end = len(ix.src.runes)
			}
			fnStack = append(fnStack, i)
			continue
		}

		limit := len(ix.src.runes)
		if len(fnStack) > 0 {
			limit = ix.decls[fnStack[len(fnStack)-1]].end
		}
		end := limit
		if i+1 < len(ix.decls) && ix.decls[i+1].start < end {
			end = ix.decls[i+1].start
		}
		if end < d.nameEnd {
			end = d.nameEnd
		}
		d.end = end
	}
}

// indexAtOrAfter finds the first token whose start is at or after an offset.
func indexAtOrAfter(tokens []scanToken, off int) int {
	lo, hi := 0, len(tokens)
	for lo < hi {
		mid := (lo + hi) / 2
		if tokens[mid].start < off {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo
}

// definitionFor finds the declaration a use of a name refers to. It picks the
// nearest declaration above the use and otherwise falls back to any other
// declaration of that name so calls written before their definition resolve.
func (ix *fileIndex) definitionFor(name string, at int) *declaration {
	var above *declaration
	var after *declaration
	for _, d := range ix.decls {
		if d.name != name {
			continue
		}
		if d.nameStart <= at {
			if above == nil || d.nameStart > above.nameStart {
				above = d
			}
			continue
		}
		if after == nil {
			after = d
		}
	}
	if above != nil {
		return above
	}
	return after
}

// importedModule returns the import entry a module name refers to.
func (ix *fileIndex) importedModule(name string) (importEntry, bool) {
	for _, imp := range ix.imports {
		if imp.alias == name || imp.path == name {
			return imp, true
		}
	}
	return importEntry{}, false
}

// isKeyword reports whether a bare word is reserved by blue. It defers to the
// tokenizer so this package keeps no keyword list of its own.
func isKeyword(word string) bool {
	return token.LookupIdent(word) != token.IDENT
}
