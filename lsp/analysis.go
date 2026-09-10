package lsp

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"blue/lexer"
	"blue/parser"
)

// dotCallFollows reports whether an opening parenthesis follows an offset, which is what turns `value.name` into a call of the plain function or builtin name with the value as first argument. blue compiles such calls exactly that way, so the docs to show are the ones of that name alone.
func dotCallFollows(src *docSource, offset int) bool {
	i := offset
	for i < len(src.runes) {
		r := src.runes[i]
		if r == ' ' || r == '\t' || r == '\n' {
			i++
			continue
		}
		return r == '('
	}
	return false
}

// maxWorkspaceFiles caps how many files a workspace wide search walks so a huge
// directory cannot stall the editor.
const maxWorkspaceFiles = 400

// maxWorkspaceSymbols caps how many symbols a workspace symbol answer returns.
const maxWorkspaceSymbols = 250

// analyze runs blue's own lex plus parse over a buffer and turns the parser
// errors into diagnostics. It is guarded because partially typed input must
// never take the server down.
func analyze(src *docSource) (diags []diagnostic) {
	defer func() {
		if rec := recover(); rec != nil {
			diags = []diagnostic{{
				Range:    rangeStruct{Start: position{}, End: position{}},
				Severity: sevError,
				Source:   "blue",
				Message:  fmt.Sprintf("internal error while analyzing %s: %v", src.name, rec),
			}}
		}
	}()

	tokens := tokenize(src.runes)

	l := lexer.New(src.text, src.name)
	p := parser.New(l)
	_ = p.ParseProgram()

	seen := map[string]bool{}
	for _, detail := range p.ErrorDetails() {
		message := strings.TrimSpace(detail.Message)
		if message == "" {
			continue
		}

		line := detail.Line
		if line < 0 {
			line = 0
		}
		col := detail.Column
		if col < 0 {
			col = 0
		}

		key := fmt.Sprintf("%s|%d|%d", message, line, col)
		if seen[key] {
			continue
		}
		seen[key] = true

		lineRunes := src.lineRunes(line)
		lineLen := utf16Len(lineRunes)
		literal := detail.TokenLiteral

		// PositionInLine cannot be used as a column on its own: the lexer records
		// it one past the first character for identifiers, keywords and literals,
		// while punctuation lands exactly. Matching the literal against our own scan
		// gives an exact span in both cases, which is what the highlight needs.
		start, end := col, col+1
		if literal != "" {
			base := src.startOfLine(line)
			if span, found := spanOfLiteral(tokens, base, col, literal); found {
				start, end = span[0], span[1]
			}
		}
		if end > lineLen {
			end = lineLen
		}
		if end < start+1 {
			end = start + 1
		}

		full := message
		for _, hint := range detail.Hints {
			hint = strings.TrimSpace(hint)
			if hint != "" {
				full += "\n" + hint
			}
		}

		diags = append(diags, diagnostic{
			Range: rangeStruct{
				Start: position{Line: line, Character: start},
				End:   position{Line: line, Character: end},
			},
			Severity: sevError,
			Code:     "parse",
			Source:   "blue",
			Message:  full,
		})
	}

	return diags
}

// spanOfLiteral finds the span of an offending token on one line of a buffer.
//
// The parser reports the column of the token it stumbled on together with that
// token's literal text. Depending on how the token was produced the recorded
// column is either exact or sits one before or after the character itself, so
// this looks for a scanned token matching the literal right around the reported
// column instead of trusting the number blindly. `base` is the rune offset where
// the line begins and col is relative to it.
func spanOfLiteral(tokens []scanToken, base, col int, literal string) ([2]int, bool) {
	target := base + col

	// Prefer, in order: the token that actually contains the reported column, then
	// the next one after it, then the closest one before it. Reported columns can
	// land inside a token or one place before it, and identical literals right
	// next to each other must never steal the match from the intended one.
	var containing, next, prev *scanToken
	for i := indexAtOrAfter(tokens, base); i < len(tokens); i++ {
		tok := tokens[i]
		if tok.start > target+1 {
			break
		}
		if tok.text != literal {
			continue
		}
		switch {
		case tok.start <= target && target < tok.end:
			if containing == nil {
				containing = &tokens[i]
			}
		case tok.start > target:
			if next == nil {
				next = &tokens[i]
			}
		default:
			prev = &tokens[i]
		}
	}

	switch {
	case containing != nil:
		return [2]int{containing.start - base, containing.end - base}, true
	case next != nil && int(next.start-target) <= 1:
		return [2]int{next.start - base, next.end - base}, true
	case prev != nil:
		return [2]int{prev.start - base, prev.end - base}, true
	}
	return [2]int{}, false
}

// publishDiagnostics sends the diagnostics of one buffer to the client.
func (s *session) publishDiagnostics(uri string) {
	s.mu.Lock()
	doc := s.docs[uri]
	s.mu.Unlock()

	if doc == nil || doc.src == nil {
		return
	}

	out := analyze(doc.src)
	if out == nil {
		out = []diagnostic{}
	}
	s.sendNotification("textDocument/publishDiagnostics", struct {
		URI         string       `json:"uri"`
		Diagnostics []diagnostic `json:"diagnostics"`
	}{URI: uri, Diagnostics: out})
}

// editContext is everything the cursor features need to know about a position in
// an open buffer.
type editContext struct {
	doc    *document
	index  *fileIndex
	src    *docSource
	offset int // rune offset of the cursor
	word   string
	tok    scanToken
	prefix string // identifier characters typed immediately before the cursor
	module string // module name when the cursor follows `module.`
}

func (s *session) contextAt(uri string, pos position) (editContext, bool) {
	doc := s.documentFor(uri)
	if doc == nil || doc.src == nil || doc.index == nil {
		return editContext{}, false
	}

	c := editContext{doc: doc, index: doc.index, src: doc.src, offset: doc.src.offsetOf(pos)}
	if tok, ok := wordAt(c.index.tokens, c.offset); ok {
		c.word, c.tok = tok.text, tok
	} else {
		// The scanner swallows comments and exec strings whole, so an identifier
		// living inside one never shows up as its own token. Expand over the runes
		// around the cursor instead so names mentioned in comments still resolve.
		c.word = wordAroundRune(c.src.runes, c.offset)
	}

	lineStart := c.src.startOfLine(pos.Line)
	start := c.offset
	for start > lineStart && isIdentRune(c.src.runes[start-1]) {
		start--
	}
	c.prefix = string(c.src.runes[start:c.offset])

	// Look for `<name>.` right before what is being typed.
	j := start
	for j > lineStart && (c.src.runes[j-1] == ' ' || c.src.runes[j-1] == '\t') {
		j--
	}
	if j > lineStart && c.src.runes[j-1] == '.' {
		end := j - 1
		k := end
		for k > lineStart && isIdentRune(c.src.runes[k-1]) {
			k--
		}
		c.module = string(c.src.runes[k:end])
	}

	return c, true
}

// insideStringOrComment reports whether an offset sits inside something that is
// not code so completion stays out of the way.
func (ix *fileIndex) insideStringOrComment(offset int) bool {
	i := indexAtOrAfter(ix.tokens, offset)
	for j := 0; j < i; j++ {
		t := ix.tokens[j]
		if (t.kind == kString || t.kind == kComment) && t.start <= offset && offset < t.end {
			return true
		}
	}
	return false
}

// resolveModuleEntry finds the source of a module that a buffer can reach. std
// modules come from blue's embedded library, local modules are read relative to
// the buffer and open buffers win over what is on disk.
func (s *session) resolveModuleEntry(c editContext, name string) (*moduleEntry, bool) {
	if isStdModule(name) {
		entry := s.modules.std(name)
		return entry, entry != nil
	}

	entry, ok := c.index.importedModule(name)
	if !ok {
		// Referencing something that was never imported is an error in blue, so
		// there is nothing sensible to resolve here.
		return nil, false
	}
	found := s.modules.get(c.doc.src.dir(), entry.path, func(p string) *docSource {
		return s.openSourceAt(p)
	})
	return found, found != nil
}

// openSourceAt returns the live text of an already open buffer for a path.
func (s *session) openSourceAt(path string) *docSource {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, doc := range s.docs {
		if doc.name == "" || doc.src == nil {
			continue
		}
		if samePath(doc.name, path) {
			return doc.src
		}
	}
	return nil
}

func samePath(a string, b string) bool {
	if a == b {
		return true
	}
	aa, errA := filepath.Abs(a)
	bb, errB := filepath.Abs(b)
	if errA != nil || errB != nil {
		return false
	}
	return aa == bb
}

// hover explains whatever word the cursor is sitting on.
func (s *session) hover(p textDocumentHoverParams) any {
	c, ok := s.contextAt(p.TextDocument.URI, p.Position)
	if !ok {
		return nil
	}

	word := c.word
	if word == "" {
		word = c.prefix
	}
	if word == "" || isKeyword(word) {
		return nil
	}

	wordRange := &rangeStruct{Start: c.src.positionOf(c.tok.start), End: c.src.positionOf(c.tok.end)}

	// `module.member` references explain the member inside the module.
	if c.module != "" {
		entry, found := s.resolveModuleEntry(c, c.module)
		if found && entry != nil {
			value := moduleMemberMarkdown(entry, c.module, word)
			if value == "" {
				return nil
			}
			return &hoverResult{Contents: markupContent{Kind: hoverMarkdown, Value: value}, Range: wordRange}
		}
		// A dot call on a plain value resolves its name in scope the way a bare call does, so only when a call follows do those names explain it. A bare `map.key` index has no docs to show here.
		end := c.tok.end
		if end == 0 {
			end = c.offset
		}
		if dotCallFollows(c.src, end) {
			return s.dotCallHover(c, word)
		}
		return nil
	}

	// Builtins such as `println` or `type_of`.
	if b, found := builtinHelp(word); found && c.index.definitionFor(word, c.offset) == nil {
		return &hoverResult{
			Contents: markupContent{Kind: hoverMarkdown, Value: "```blue\n" + strings.TrimSpace(b.HelpStr) + "\n```"},
			Range:    wordRange,
		}
	}

	// Names declared in this buffer.
	decl := c.index.definitionFor(word, c.offset)
	if decl == nil {
		return nil
	}
	return &hoverResult{
		Contents: markupContent{Kind: hoverMarkdown, Value: declarationMarkdown(c.src.runes, decl)},
		Range:    wordRange,
	}
}

// dotCallHover explains `value.name` when blue cannot tell that value is a module. It looks up name as a plain function or builtin first, which is what such a call compiles to: the receiver simply becomes the first argument.
func (s *session) dotCallHover(c editContext, word string) any {
	wordRange := &rangeStruct{Start: c.src.positionOf(c.tok.start), End: c.src.positionOf(c.tok.end)}

	// A local declaration of that name wins over a builtin, exactly like the bare-name hover below does.
	decl := c.index.definitionFor(word, c.offset)
	if decl != nil {
		return &hoverResult{
			Contents: markupContent{Kind: hoverMarkdown, Value: declarationMarkdown(c.src.runes, decl)},
			Range:    wordRange,
		}
	}
	if b, found := looseBuiltinHelp(word); found && strings.TrimSpace(b.HelpStr) != "" {
		return &hoverResult{
			Contents: markupContent{Kind: hoverMarkdown, Value: "```blue\n" + strings.TrimSpace(b.HelpStr) + "\n```"},
			Range:    wordRange,
		}
	}
	return nil
}

// declarationMarkdown renders a declaration of the current buffer for hover.
func declarationMarkdown(runes []rune, d *declaration) string {
	var b strings.Builder
	b.WriteString("```blue\n")
	b.WriteString(declarationSignature(runes, d))
	b.WriteString("\n```")
	if d.doc != "" {
		b.WriteString("\n\n")
		b.WriteString(d.doc)
	}
	return b.String()
}

// declarationSignature renders a declaration back into readable source text.
func declarationSignature(runes []rune, d *declaration) string {
	end := d.end
	if end < 0 || end > len(runes) {
		end = len(runes)
	}
	text := strings.TrimSpace(string(runes[d.start:end]))
	text = strings.TrimRight(text, "{ \t\n")

	switch d.kind {
	case declVariable:
		return text
	case declConstant:
		return text
	case declParam:
		return "param " + d.name
	case declLoopVar:
		if text == "" {
			return "for ... {"
		}
		return text + " ... {"
	default:
		return text
	}
}

// moduleMemberMarkdown explains `<module>.<member>` from the module's source.
// Members that are nothing but a wrapper around a go builtin also show that
// builtin's help text, which carries the signature and examples.
func moduleMemberMarkdown(entry *moduleEntry, moduleName string, member string) string {
	index := entry.ix
	var b strings.Builder
	fmt.Fprintf(&b, "**module** `%s`\n\n", moduleName)

	decl := index.definitionFor(member, 0)
	if decl != nil {
		b.WriteString("```blue\n")
		b.WriteString(declarationSignature(entry.src.runes, decl))
		b.WriteString("\n```")
		if decl.doc != "" {
			b.WriteString("\n\n")
			b.WriteString(decl.doc)
		}
	} else if target := aliasedBuiltin(entry.src.runes, index, member); target == "" {
		return ""
	}

	if target := aliasedBuiltin(entry.src.runes, index, member); target != "" {
		if builtin, ok := builtinHelp(target); ok && strings.TrimSpace(builtin.HelpStr) != "" {
			b.WriteString("\n\n```blue\n")
			b.WriteString(strings.TrimSpace(builtin.HelpStr))
			b.WriteString("\n```")
		}
	}
	return b.String()
}

// aliasedBuiltin finds the builtin a module level name simply wraps, such as the
// `_acos` behind `val acos = _acos;`.
func aliasedBuiltin(runes []rune, index *fileIndex, name string) string {
	for _, d := range index.decls {
		if d.name != name {
			continue
		}
		end := d.end
		if end < 0 || end > len(runes) {
			end = len(runes)
		}
		fields := strings.Fields(string(runes[d.nameEnd:end]))
		if len(fields) == 0 || fields[0] != "=" {
			continue
		}
		if len(fields) < 2 {
			continue
		}
		candidate := strings.TrimRight(fields[1], ";")
		if isBuiltinName(candidate) {
			return candidate
		}
	}
	return ""
}

// definition jumps to where a name is bound.
func (s *session) definition(p textDocumentPositionParams) any {
	c, ok := s.contextAt(p.TextDocument.URI, p.Position)
	if !ok {
		return nil
	}

	word := c.word
	if word == "" {
		word = c.prefix
	}

	// Import statements jump to the module file.
	if entry, found := importEntryAt(c.index, c.offset); found {
		if isStdModule(entry.path) {
			return nil
		}
		path := moduleFilePath(c.doc.src.dir(), entry.path)
		if !fileExists(path) {
			return nil
		}
		return []location{{URI: pathToURI(path), Range: rangeStruct{Start: position{}, End: position{}}}}
	}

	if word == "" || isKeyword(word) {
		return nil
	}

	// `module.member` jumps into the module's file when it has one. When the
	// receiver is not a resolvable module, such as a dot call on a plain value,
	// the name resolves in this buffer instead so the fallthrough below handles it.
	if c.module != "" {
		entry, found := s.resolveModuleEntry(c, c.module)
		if found && entry != nil && entry.path != "" {
			decl := entry.ix.definitionFor(word, 0)
			if decl == nil {
				return nil
			}
			return []location{{
				URI:   pathToURI(entry.path),
				Range: rangeStruct{Start: entry.src.positionOf(decl.nameStart), End: entry.src.positionOf(decl.nameEnd)},
			}}
		}
	}

	if decl := c.index.definitionFor(word, c.offset); decl != nil {
		return []location{{
			URI:   p.TextDocument.URI,
			Range: rangeStruct{Start: c.src.positionOf(decl.nameStart), End: c.src.positionOf(decl.nameEnd)},
		}}
	}

	// Nothing local, so look through the other open buffers and the workspace.
	return s.findGlobalDefinition(word, p.TextDocument.URI)
}

// importEntryAt returns the import statement covering a rune offset.
func importEntryAt(index *fileIndex, off int) (importEntry, bool) {
	for _, entry := range index.imports {
		if entry.start <= off && off <= entry.end {
			return entry, true
		}
	}
	return importEntry{}, false
}

// findGlobalDefinition searches other open buffers and then blue files in the
// workspace root for a declaration of a name.
func (s *session) findGlobalDefinition(name string, skipURI string) []location {
	if name == "" {
		return nil
	}

	s.mu.Lock()
	others := make([]*document, 0, len(s.docs))
	for uri, doc := range s.docs {
		if uri == skipURI || doc.index == nil {
			continue
		}
		others = append(others, doc)
	}
	root := s.root
	s.mu.Unlock()

	for _, doc := range others {
		decl := doc.index.definitionFor(name, 0)
		if decl == nil {
			continue
		}
		return []location{{
			URI:   doc.uri,
			Range: rangeStruct{Start: doc.src.positionOf(decl.nameStart), End: doc.src.positionOf(decl.nameEnd)},
		}}
	}

	if root == "" {
		return nil
	}
	for _, path := range blueFilesIn(root, maxWorkspaceFiles) {
		if s.openSourceAt(path) != nil {
			continue
		}
		data, err := readCapped(path)
		if err != nil {
			continue
		}
		src := newDocSource(path, string(data))
		index := buildIndex(src)
		index.resolveExtents()
		decl := index.definitionFor(name, 0)
		if decl == nil {
			continue
		}
		return []location{{
			URI:   pathToURI(path),
			Range: rangeStruct{Start: src.positionOf(decl.nameStart), End: src.positionOf(decl.nameEnd)},
		}}
	}
	return nil
}

// blueFilesIn collects .b files under a root directory, skipping the usual noise.
func blueFilesIn(root string, limit int) []string {
	found := []string{}
	errStopWalk := fmt.Errorf("walk stopped")

	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			if entry != nil && entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			name := entry.Name()
			if path != root && (strings.HasPrefix(name, ".") || name == "__blue_cache" || name == "vendor" || name == "node_modules") {
				return fs.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(entry.Name(), ".b") {
			found = append(found, path)
			if limit > 0 && len(found) >= limit {
				return errStopWalk
			}
		}
		return nil
	})

	return found
}

// references lists everywhere a name is used inside its buffer.
func (s *session) references(p referenceParams) any {
	c, ok := s.contextAt(p.TextDocument.URI, p.Position)
	if !ok || c.word == "" {
		return nil
	}

	// The client says whether the declaration itself belongs in the answer.
	include := true
	if p.Context != nil {
		include = p.Context.IncludeDeclaration
	}

	onDecl := map[int]bool{}
	if !include {
		for _, d := range c.index.decls {
			if d.name == c.word {
				onDecl[d.nameStart] = true
			}
		}
	}

	out := []location{}
	for _, tok := range c.index.tokens {
		if tok.kind != kIdent || tok.text != c.word {
			continue
		}
		if onDecl[tok.start] {
			continue
		}
		out = append(out, location{
			URI:   p.TextDocument.URI,
			Range: rangeStruct{Start: c.src.positionOf(tok.start), End: c.src.positionOf(tok.end)},
		})
	}
	return out
}

// documentSymbol builds the outline of a buffer, nesting declarations inside the
// functions that contain them.
// symNode keeps a stable address for every symbol while the tree is assembled.
// The protocol struct stores children by value, so appending into one while later
// declarations are still arriving would silently lose them. Nodes hold pointers
// and are flattened back into values once everything is in place.
type symNode struct {
	sym  *documentSymbol
	end  int
	kids []*symNode
}

// flatten turns assembled nodes into protocol symbols, depth first.
func flatten(nodes []*symNode) []documentSymbol {
	out := []documentSymbol{}
	for _, n := range nodes {
		sym := *n.sym
		sym.Children = flatten(n.kids)
		out = append(out, sym)
	}
	return out
}

func (s *session) documentSymbol(p textDocumentIdentifier) any {
	doc := s.documentFor(p.URI)
	if doc == nil || doc.index == nil {
		return nil
	}

	index := doc.index
	src := doc.src
	limit := len(src.runes)
	root := &symNode{sym: &documentSymbol{Name: src.name, Kind: symKindFile}}
	stack := []*symNode{}

	// add drops any scopes that have already closed and attaches the symbol to
	// whatever scope is open at that point.
	add := func(sym *documentSymbol, start int, end int, pushScope bool) {
		if end < 0 || end > limit {
			end = limit
		}
		for len(stack) > 0 && stack[len(stack)-1].end <= start {
			stack = stack[:len(stack)-1]
		}

		parent := root
		if len(stack) > 0 {
			parent = stack[len(stack)-1]
		}

		// The node put on the stack must be the very same one that was attached to
		// the parent, otherwise children land on an orphan and never reach the
		// client. Sharing one pointer keeps `detail` as written on source too.
		node := &symNode{sym: sym, end: end}
		parent.kids = append(parent.kids, node)
		if pushScope {
			stack = append(stack, node)
		}
	}

	for _, entry := range index.imports {
		label := "import " + entry.path
		if entry.alias != "" && entry.alias != lastSegment(entry.path) {
			label += " as " + entry.alias
		}
		if len(entry.names) > 0 {
			label += " import {" + strings.Join(entry.names, ", ") + "}"
		}
		end := entry.end
		if end < 0 || end > limit {
			end = limit
		}
		add(&documentSymbol{
			Name:           label,
			Kind:           symKindModule,
			Range:          rangeStruct{Start: src.positionOf(entry.start), End: src.positionOf(end)},
			SelectionRange: rangeStruct{Start: src.positionOf(entry.start), End: src.positionOf(entry.start)},
		}, entry.start, end, false)
	}

	for _, d := range index.decls {
		if d.kind == declParam || d.kind == declLoopVar || d.kind == declCatchVar {
			continue
		}
		sym := &documentSymbol{
			Name:           d.name,
			Kind:           d.kind.symbolKind(),
			Range:          rangeStruct{Start: src.positionOf(d.start), End: src.positionOf(d.end)},
			SelectionRange: rangeStruct{Start: src.positionOf(d.nameStart), End: src.positionOf(d.nameEnd)},
		}
		if d.kind == declFunction && d.detail != "" {
			sym.Detail = d.detail
		}
		add(sym, d.start, d.end, d.kind == declFunction)
	}

	return flatten(root.kids)
}

// documentHighlight marks every occurrence of the identifier under the cursor.
func (s *session) documentHighlight(p documentHighlightParams) any {
	c, ok := s.contextAt(p.TextDocument.URI, p.Position)
	if !ok || c.word == "" {
		return nil
	}

	declaredHere := map[int]string{}
	for _, d := range c.index.decls {
		if d.name == c.word {
			declaredHere[d.nameStart] = d.name
		}
	}

	toks := c.index.tokens
	out := []documentHighlight{}
	for i, tok := range toks {
		if tok.kind != kIdent || tok.text != c.word {
			continue
		}
		kind := int(highlightRead)
		if _, isDecl := declaredHere[tok.start]; isDecl {
			kind = int(highlightWrite)
		} else if isAssignOp(operatorAfter(toks, i)) {
			kind = int(highlightWrite)
		}
		out = append(out, documentHighlight{
			Range: rangeStruct{Start: c.src.positionOf(tok.start), End: c.src.positionOf(tok.end)},
			Kind:  kind,
		})
	}
	return out
}

// operatorAfter joins the punctuation tokens that make up the operator applied
// right after token i. tokenize emits operators one rune at a time, so `+=` and
// `==` only exist once those runes are put back together.
func operatorAfter(tokens []scanToken, i int) string {
	run := ""
	for j := i + 1; j < len(tokens); j++ {
		t := tokens[j]
		if t.kind != kPunct || !isOperatorRune([]rune(t.text)[0]) {
			break
		}
		run += t.text
	}
	return run
}

// isAssignOp reports whether an operator binds a value to the name before it.
// Comparisons such as `==` and `<=` are not assignments even though they end in
// '=', and `=>` is an arrow, not an assignment either.
func isAssignOp(op string) bool {
	switch op {
	case "=", "+=", "-=", "*=", "/=", "%=", "&=", "|=", "^=", "<<=", ">>=", "&&=", "||=":
		return true
	}
	return false
}

func isOperatorRune(r rune) bool {
	switch r {
	case '=', '+', '-', '*', '/', '%', '&', '|', '^', '<', '>':
		return true
	}
	return false
}

// workspaceSymbol answers a symbol search across open buffers and the workspace.
func (s *session) workspaceSymbol(p workspaceSymbolParams) any {
	query := strings.ToLower(strings.TrimSpace(p.Query))
	matches := []symbolInformation{}
	seen := map[string]bool{}

	add := func(name string, kind int, uri string, start position, end position) {
		if query != "" && !strings.Contains(strings.ToLower(name), query) {
			return
		}
		key := uri + ":" + name
		if seen[key] {
			return
		}
		seen[key] = true
		matches = append(matches, symbolInformation{
			Name:     name,
			Kind:     kind,
			Location: location{URI: uri, Range: rangeStruct{Start: start, End: end}},
		})
	}

	s.mu.Lock()
	docs := make([]*document, 0, len(s.docs))
	for _, doc := range s.docs {
		if doc.index != nil {
			docs = append(docs, doc)
		}
	}
	root := s.root
	s.mu.Unlock()

	for _, doc := range docs {
		for _, d := range doc.index.decls {
			add(d.name, d.kind.symbolKind(), doc.uri, doc.src.positionOf(d.nameStart), doc.src.positionOf(d.nameEnd))
		}
	}

	if root != "" {
		for _, path := range blueFilesIn(root, maxWorkspaceFiles) {
			if s.openSourceAt(path) != nil {
				continue
			}
			data, err := readCapped(path)
			if err != nil {
				continue
			}
			src := newDocSource(path, string(data))
			index := buildIndex(src)
			index.resolveExtents()
			for _, d := range index.decls {
				add(d.name, d.kind.symbolKind(), pathToURI(path), src.positionOf(d.nameStart), src.positionOf(d.nameEnd))
			}
			if len(matches) >= maxWorkspaceSymbols {
				break
			}
		}
	}

	sort.SliceStable(matches, func(i, j int) bool { return matches[i].Name < matches[j].Name })
	return matches
}
