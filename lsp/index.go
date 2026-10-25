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
	name          string
	kind          declKind
	start         int // rune offset of the binding keyword (fun, val, for, ...)
	nameStart     int // rune offset of the name itself
	nameEnd       int // rune offset just past the name
	end           int // exclusive rune offset, -1 until extents are resolved
	bodyOpen      int // token index of the body '{' for functions, -1 otherwise
	detail        string
	params        []string
	doc           string
	members       []*mapMember
	initCall      []string
	resultMembers []*mapMember
}

type mapMember struct {
	name          string
	kind          declKind
	keyStart      int
	keyEnd        int
	bodyOpen      int
	detail        string
	doc           string
	ref           string
	src           *docSource
	ix            *fileIndex
	members       []*mapMember
	resultMembers []*mapMember
}

// importEntry is one `import`/`from ... import` statement.
type importEntry struct {
	path  string   // dotted module path such as "foo.bar"
	alias string   // name the module is reachable as (alias or last segment)
	names []string // explicit names for `from path import {a, b}`
	start int
	end   int
	// Extent of the part of the statement that names the module, which is the
	// dotted path plus any alias after it. The names inside `from x import
	// {a, b}` are members of the module rather than the module itself, so they
	// fall outside it.
	nameStart int
	nameEnd   int
}

// coversModuleName reports whether an offset sits on the part of an import
// statement that names the module.
func (e importEntry) coversModuleName(off int) bool {
	return e.nameStart <= off && off <= e.nameEnd
}

// fileIndex holds what the LSP knows about a document buffer. It comes from the
// light weight scanner instead of the real parser so that features keep working
// while a buffer has syntax errors, which is the normal state of code being
// typed. Call resolveExtents before querying it.
type fileIndex struct {
	src        *docSource
	tokens     []scanToken
	decls      []*declaration
	imports    []importEntry
	mapMembers []*mapMember
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

func mapKeyName(t scanToken) (string, bool) {
	switch t.kind {
	case kIdent, kWord:
		return t.text, true
	case kString:
		rs := []rune(t.text)
		if len(rs) < 2 {
			return "", false
		}
		quote := rs[0]
		if quote != '\'' && quote != '"' {
			return "", false
		}
		if rs[len(rs)-1] != quote {
			return "", false
		}
		body := string(rs[1 : len(rs)-1])
		if strings.Contains(body, "#{") {
			return "", false
		}
		return body, true
	}
	return "", false
}

func indexMapLiteral(src *docSource, ix *fileIndex, toks []scanToken, openIdx int) []*mapMember {
	if openIdx < 0 || openIdx >= len(toks) || !isPunct(toks[openIdx], "{") {
		return nil
	}
	closeIdx := matchDelim(toks, openIdx, "{", "}")
	limit := len(toks)
	if closeIdx >= 0 {
		limit = closeIdx
	}

	members := []*mapMember{}
	start := openIdx + 1
	depth := 0
	for i := start; i <= limit; i++ {
		if i == limit || (depth == 0 && isPunct(toks[i], ",")) {
			if i > start {
				if member := indexMapMember(src, ix, toks, start, i); member != nil {
					members = append(members, member)
					ix.mapMembers = append(ix.mapMembers, member)
				}
			}
			start = i + 1
			continue
		}
		switch {
		case isPunct(toks[i], "(") || isPunct(toks[i], "[") || isPunct(toks[i], "{"):
			depth++
		case isPunct(toks[i], ")") || isPunct(toks[i], "]") || isPunct(toks[i], "}"):
			depth--
		}
	}
	return members
}

func indexMapMember(src *docSource, ix *fileIndex, toks []scanToken, start, end int) *mapMember {
	keyIdx := start
	for keyIdx < end && toks[keyIdx].kind == kComment {
		keyIdx++
	}
	if keyIdx >= end {
		return nil
	}
	name, ok := mapKeyName(toks[keyIdx])
	if !ok || keyIdx+1 >= end || !isPunct(toks[keyIdx+1], ":") {
		return nil
	}
	valueStart := keyIdx + 2
	if valueStart >= end {
		return nil
	}
	return newMapMember(src, ix, toks, name, keyIdx, valueStart, end)
}

func newMapMember(src *docSource, ix *fileIndex, toks []scanToken, name string, keyIdx, valueStart, end int) *mapMember {
	member := &mapMember{
		name:     name,
		kind:     declConstant,
		keyStart: toks[keyIdx].start,
		keyEnd:   toks[keyIdx].end,
		src:      src,
		ix:       ix,
	}

	value := toks[valueStart]
	switch {
	case value.kind == kWord && value.text == "fun" && valueStart+1 < end && isPunct(toks[valueStart+1], "("):
		openParen := valueStart + 1
		closeParen := matchDelim(toks, openParen, "(", ")")
		if closeParen < 0 || closeParen >= end {
			member.doc = docstringFor(src.runes, toks, keyIdx, -1)
			return member
		}
		bodyOpen := closeParen + 1
		for bodyOpen < end && !isPunct(toks[bodyOpen], "{") {
			bodyOpen++
		}
		if bodyOpen >= end {
			member.doc = docstringFor(src.runes, toks, keyIdx, -1)
			return member
		}
		member.kind = declFunction
		member.detail = functionDetail(src, toks, openParen, closeParen)
		member.bodyOpen = bodyOpen
		member.doc = docstringFor(src.runes, toks, keyIdx, bodyOpen)

	case isPunct(value, "{"):
		member.members = indexMapLiteral(src, ix, toks, valueStart)
		member.doc = docstringFor(src.runes, toks, keyIdx, -1)

	default:
		if ref, ok := aliasRef(src, toks, valueStart, end); ok {
			member.ref = ref
		}
		member.doc = docstringFor(src.runes, toks, keyIdx, -1)
	}
	return member
}

func (ix *fileIndex) mapMemberAt(off int) *mapMember {
	for _, member := range ix.mapMembers {
		if member.keyStart <= off && off <= member.keyEnd {
			return member
		}
	}
	return nil
}

func findMapMember(members []*mapMember, name string) *mapMember {
	for _, member := range members {
		if member.name == name {
			return member
		}
	}
	return nil
}

func aliasRef(src *docSource, toks []scanToken, start, end int) (string, bool) {
	if start >= end {
		return "", false
	}
	t := toks[start]
	if t.kind != kIdent && t.kind != kWord {
		return "", false
	}
	if start+1 < end {
		next := toks[start+1]
		separator := isPunct(next, ";") || isPunct(next, "}") || isPunct(next, ",") || isPunct(next, ")")
		if next.kind != kComment && !separator && sameTokenLine(src, t, next) {
			return "", false
		}
	}
	return t.text, true
}

func sameTokenLine(src *docSource, a, b scanToken) bool {
	return src.positionOf(a.end).Line == src.positionOf(b.start).Line
}

func declarationFunctionLiteral(ix *fileIndex, d *declaration) *mapMember {
	toks := ix.tokens
	nameIdx := indexAtOrAfter(toks, d.nameStart)
	if nameIdx >= len(toks) || toks[nameIdx].start != d.nameStart {
		return nil
	}
	for i := nameIdx + 1; i+1 < len(toks) && toks[i].start < d.end; i++ {
		if isPunct(toks[i], "=") && isFunctionLiteralAt(toks, i+1, len(toks)) {
			lit := newMapMember(ix.src, ix, toks, d.name, nameIdx, i+1, len(toks))
			if lit == nil {
				return nil
			}
			if declIdx := indexAtOrAfter(toks, d.start); declIdx < len(toks) {
				lit.doc = docstringFor(ix.src.runes, toks, declIdx, lit.bodyOpen)
			}
			return lit
		}
	}
	return nil
}

func isMapMemberValue(src *docSource, toks []scanToken, idx, end int) bool {
	if idx >= end {
		return false
	}
	if isPunct(toks[idx], "{") || isFunctionLiteralAt(toks, idx, end) {
		return true
	}
	_, ok := aliasRef(src, toks, idx, end)
	return ok
}

func callPathAfter(toks []scanToken, start int) []string {
	path := []string{}
	i := start
	for i < len(toks) {
		if toks[i].kind != kIdent && toks[i].kind != kWord {
			break
		}
		path = append(path, toks[i].text)
		i++
		if i < len(toks) && isPunct(toks[i], ".") {
			i++
			continue
		}
		break
	}
	if len(path) == 0 || i >= len(toks) || !isPunct(toks[i], "(") {
		return nil
	}
	if closeIdx := matchDelim(toks, i, "(", ")"); closeIdx >= 0 && closeIdx+1 < len(toks) {
		next := toks[closeIdx+1]
		if next.kind == kPunct && !isPunct(next, ";") && !isPunct(next, "}") && !isPunct(next, ")") {
			return nil
		}
	}
	return path
}

func isFunctionLiteralAt(toks []scanToken, idx, end int) bool {
	return idx+1 < end && toks[idx].kind == kWord && toks[idx].text == "fun" && isPunct(toks[idx+1], "(")
}

func mergeMapMembers(existing []*mapMember, adding []*mapMember) []*mapMember {
	out := append([]*mapMember{}, existing...)
	for _, member := range adding {
		if findMapMember(out, member.name) == nil {
			out = append(out, member)
		}
	}
	return out
}

func returnedNames(toks []scanToken, bodyOpen int, limit int) []string {
	names := []string{}
	seen := map[string]bool{}
	depth := 0
	for i := bodyOpen + 1; i < limit; i++ {
		t := toks[i]
		switch {
		case isPunct(t, "{"):
			depth++
			continue
		case isPunct(t, "}"):
			depth--
			continue
		}
		if depth != 0 || t.kind != kWord || t.text != "return" {
			continue
		}
		if i+1 >= limit || toks[i+1].kind != kIdent {
			continue
		}
		name := toks[i+1].text
		if !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}
	return names
}

func assignedMapMembers(src *docSource, ix *fileIndex, toks []scanToken, bodyOpen int, limit int) map[string][]*mapMember {
	assigned := map[string][]*mapMember{}
	depth := 0
	for i := bodyOpen + 1; i < limit; i++ {
		t := toks[i]
		switch {
		case isPunct(t, "{"):
			depth++
			continue
		case isPunct(t, "}"):
			depth--
			continue
		}
		if depth != 0 || t.kind != kIdent {
			continue
		}

		if i+3 < limit && isPunct(toks[i+1], ".") && toks[i+2].kind == kIdent && isPunct(toks[i+3], "=") {
			keyIdx := i + 2
			valueStart := i + 4
			if isMapMemberValue(src, toks, valueStart, limit) {
				member := newMapMember(src, ix, toks, toks[keyIdx].text, keyIdx, valueStart, limit)
				assigned[t.text] = append(assigned[t.text], member)
				ix.mapMembers = append(ix.mapMembers, member)
			}
			continue
		}

		if i+5 < limit && isPunct(toks[i+1], "[") && isPunct(toks[i+3], "]") && isPunct(toks[i+4], "=") {
			name, ok := mapKeyName(toks[i+2])
			valueStart := i + 5
			if ok && isMapMemberValue(src, toks, valueStart, limit) {
				member := newMapMember(src, ix, toks, name, i+2, valueStart, limit)
				assigned[t.text] = append(assigned[t.text], member)
				ix.mapMembers = append(ix.mapMembers, member)
			}
		}
	}
	return assigned
}

func indexFunctionResultMembers(src *docSource, ix *fileIndex) {
	toks := ix.tokens
	for _, fn := range ix.decls {
		if fn.kind != declFunction || fn.bodyOpen < 0 || fn.bodyOpen >= len(toks) {
			continue
		}
		limit := len(toks)
		if closeIdx := matchDelim(toks, fn.bodyOpen, "{", "}"); closeIdx >= 0 {
			limit = closeIdx
		}
		fn.resultMembers = indexBuilderBody(src, ix, fn.bodyOpen, limit, fn.start, fn.end)
	}

	for i := 0; i < len(ix.mapMembers); i++ {
		member := ix.mapMembers[i]
		if member.kind != declFunction || member.bodyOpen < 0 || member.bodyOpen >= len(toks) {
			continue
		}
		limit := len(toks)
		if closeIdx := matchDelim(toks, member.bodyOpen, "{", "}"); closeIdx >= 0 {
			limit = closeIdx
		}
		attachEnd := len(src.runes)
		if limit < len(toks) {
			attachEnd = toks[limit].end
		}
		member.resultMembers = indexBuilderBody(src, ix, member.bodyOpen, limit, member.keyStart, attachEnd)
	}
}

func indexBuilderBody(src *docSource, ix *fileIndex, bodyOpen int, limit int, attachStart int, attachEnd int) []*mapMember {
	toks := ix.tokens
	names := returnedNames(toks, bodyOpen, limit)
	if len(names) == 0 {
		return nil
	}
	assigned := assignedMapMembers(src, ix, toks, bodyOpen, limit)
	var result []*mapMember
	for _, name := range names {
		members := assigned[name]
		if len(members) == 0 {
			continue
		}
		result = mergeMapMembers(result, members)
		var target *declaration
		for _, d := range ix.decls {
			if d.name != name || (d.kind != declVariable && d.kind != declConstant) {
				continue
			}
			if d.start < attachStart || d.start >= attachEnd {
				continue
			}
			if target == nil || d.start < target.start {
				target = d
			}
		}
		if target != nil {
			target.members = mergeMapMembers(target.members, members)
		}
	}
	return result
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

	return cleanDocstring(lines)
}

// cleanDocstring tidies collected docstring lines: compiler directives are not
// prose and never belong in what a reader sees, and blank `##` lines that only
// pad the ends of the block are dropped while the ones separating a description
// from its signature are kept.
func cleanDocstring(lines []string) string {
	picked := []string{}
	for _, l := range lines {
		if isHelpDirective(strings.TrimSpace(l)) {
			continue
		}
		picked = append(picked, l)
	}
	for len(picked) > 0 && strings.TrimSpace(picked[0]) == "" {
		picked = picked[1:]
	}
	for len(picked) > 0 && strings.TrimSpace(picked[len(picked)-1]) == "" {
		picked = picked[:len(picked)-1]
	}
	return strings.Join(picked, "\n")
}

// moduleDocstring collects the `##` comment lines a module opens its source
// with, which document the module as a whole. Blue's standard library writes one
// like that in every module, and it is what a reader wants when the module name
// itself is hovered. Only the comment block ahead of the module's first
// declaration is considered, so the docstring of some later declaration can
// never be mistaken for the module's own description. Plain `#` lines sitting
// in that block, such as a shebang or a licence header, are stepped over rather
// than ending it, since a `##` block below one is still the module's docs.
func moduleDocstring(tokens []scanToken) string {
	lines := []string{}
	for _, t := range tokens {
		if t.kind != kComment {
			break
		}
		text := strings.TrimRight(t.text, " \t")
		if !strings.HasPrefix(text, "##") {
			continue
		}
		lines = append(lines, commentBody(text))
	}
	return cleanDocstring(lines)
}

// isHelpDirective reports whether a docstring line is one of blue's compiler
// help directives, like `std:this,__style` or `core:ignore`.
func isHelpDirective(line string) bool {
	return strings.HasPrefix(line, "std:") || strings.HasPrefix(line, "core:")
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
			var members []*mapMember
			var initCall []string
			if end < len(toks) && isPunct(toks[end], "=") && end+1 < len(toks) {
				if isPunct(toks[end+1], "{") {
					members = indexMapLiteral(src, ix, toks, end+1)
				}
				initCall = callPathAfter(toks, end+1)
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
					members:   members,
					initCall:  initCall,
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

	indexFunctionResultMembers(src, ix)
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
	entry.nameStart = tokens[i].start
	i++
	for i+1 < len(tokens) && isPunct(tokens[i], ".") && tokens[i+1].kind == kIdent {
		path.WriteString(".")
		path.WriteString(tokens[i+1].text)
		i += 2
	}
	entry.path = path.String()
	entry.nameEnd = tokens[i-1].end

	if !fromStmt {
		// import path [as alias]
		if i+1 < len(tokens) && tokens[i].kind == kWord && tokens[i].text == "as" && tokens[i+1].kind == kIdent {
			entry.alias = tokens[i+1].text
			entry.nameEnd = tokens[i+1].end
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
