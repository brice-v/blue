package lsp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"blue/consts"
	"blue/token"
)

// clientCapabilities is the subset of what an editor declares that changes how
// blue answers. Only the parts used here are decoded.
type clientCapabilities struct {
	TextDocument struct {
		Completion *struct {
			CompletionItem *struct {
				SnippetSupport bool `json:"snippetSupport"`
			} `json:"completionItem"`
		} `json:"completion"`
	} `json:"textDocument"`
}

// initialize answers the client's initialize request with what this server can do.
func (s *session) initialize(p initializeParams) any {
	if path, ok := uriToPath(p.RootURI); ok && path != "" {
		s.root = path
	} else if p.RootPath != "" {
		s.root = p.RootPath
	}

	var caps clientCapabilities
	if len(p.Capabilities) > 0 {
		if err := json.Unmarshal(p.Capabilities, &caps); err == nil {
			completion := caps.TextDocument.Completion
			if completion != nil && completion.CompletionItem != nil {
				s.clientSnippets = completion.CompletionItem.SnippetSupport
			}
		}
	}

	return &initializeResult{
		Capabilities: serverCapabilities{
			TextDocumentSync: &textDocumentSyncOptions{
				OpenClose: true,
				Change:    syncKindFull,
				Save:      &saveOptions{IncludeText: true},
			},
			// Only "." is listed as a trigger character because nothing else in blue
			// identifiers needs an on-the-fly popup.
			CompletionProvider:        &completionOptions{TriggerCharacters: []string{"."}},
			HoverProvider:             true,
			DefinitionProvider:        true,
			ReferencesProvider:        true,
			DocumentSymbolProvider:    true,
			DocumentHighlightProvider: true,
			SemanticTokensProvider: &semanticTokensOptions{
				Legend: semanticTokensLegend{
					TokenTypes:     semLegendTypes,
					TokenModifiers: semLegendModifiers,
				},
				Range: false,
				Full:  true,
			},
		},
		ServerInfo: &serverInfo{Name: "blue-lsp", Version: consts.ShortVersion()},
	}
}

// maxCompletionItems caps a completion response so huge menus stay cheap.
const maxCompletionItems = 800

// completion offers keywords, builtins, imported module members and everything
// else declared in the buffer.
func (s *session) completion(p completionParams) any {
	c, ok := s.contextAt(p.TextDocument.URI, p.Position)
	if !ok {
		return &completionList{IsIncomplete: false, Items: []completionItem{}}
	}

	if c.index.insideStringOrComment(c.offset) {
		return &completionList{IsIncomplete: false, Items: []completionItem{}}
	}

	lower := strings.ToLower(c.prefix)
	b := newCompletionBuilder(lower, s.clientSnippets)
	// Which builtin groups this buffer can reach depends on what it imported.
	b.imported = map[string]bool{"core": true}
	for _, imp := range c.index.imports {
		b.imported[imp.path] = true
		if imp.alias != "" {
			b.imported[imp.alias] = true
		}
	}

	// `import <name>` offers std modules plus the .b files sitting next to this
	// buffer, which is the part people actually have to guess at.
	if importPathBeingTyped(c.src.textUpToPosition(p.Position)) {
		b.addImportCandidates(c)
		if len(b.items) > 0 {
			return b.list()
		}
	}

	// Member access after a dot. Modules get their own member list. A plain
	// value's members depend on its runtime type, which blue does not know at
	// this point, so the core builtins are offered instead of nothing.
	if c.module != "" {
		entry, found := s.resolveModuleEntry(c, c.module)
		if found && entry != nil {
			b.addModuleMembers(entry)
		} else {
			b.addBuiltinGroup("core")
		}
		return b.list()
	}

	b.addBufferDecls(c)
	b.addBuiltinGroup("core")
	b.addImportedModules(c)
	for _, word := range token.Keywords() {
		b.addKeyword(word)
	}
	return b.list()
}

// importPathBeingTyped reports whether the text so far is a bare `import` (or
// `from`) statement waiting for a module name.
func importPathBeingTyped(line string) bool {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return false
	}
	switch fields[0] {
	case "import":
		return true
	case "from":
		// `from x import {...}` takes member names, which the normal candidates
		// handle once the module itself is typed.
		return len(fields) == 1
	default:
		return false
	}
}

// completionBuilder assembles filtered and ranked completion items.
type completionBuilder struct {
	lower    string
	snippets bool
	items    []completionItem
	rank     int
	used     map[string]bool
	imported map[string]bool
}

func newCompletionBuilder(lower string, snippets bool) *completionBuilder {
	return &completionBuilder{lower: lower, snippets: snippets, used: map[string]bool{}}
}

// seen reports whether a label has already been offered so it is not repeated by
// a later group (a user declared `print` beats the builtin of the same name).
func (b *completionBuilder) seen(name string) bool { return b.used[name] }

// markSeen remembers that a label was offered.
func (b *completionBuilder) markSeen(name string) { b.used[name] = true }

func (b *completionBuilder) matches(label string) bool {
	if b.lower == "" {
		return true
	}
	return strings.HasPrefix(strings.ToLower(label), b.lower)
}

func (b *completionBuilder) add(item completionItem) {
	if len(b.items) >= maxCompletionItems {
		return
	}
	item.SortText = fmt.Sprintf("%03d:%s", b.rank, strings.TrimSuffix(item.Label, "()"))
	b.items = append(b.items, item)
}

func (b *completionBuilder) list() *completionList {
	if b.items == nil {
		b.items = []completionItem{}
	}
	return &completionList{IsIncomplete: false, Items: b.items}
}

// addKeyword adds a plain reserved word candidate.
func (b *completionBuilder) addKeyword(word string) {
	if !b.matches(word) {
		return
	}
	b.rank = 90
	b.add(completionItem{Label: word, Kind: kindKeyword})
}

// callInsertText builds the inserted text for a callable so that the cursor lands
// in its argument list.
func callInsertText(name string, params []string, snippets bool) string {
	if !snippets {
		return name + "()"
	}
	if len(params) == 0 {
		return name + "($0)"
	}
	parts := make([]string, 0, len(params))
	for i, p := range params {
		parts = append(parts, fmt.Sprintf("${%d:%s}", i+1, p))
	}
	return name + "(" + strings.Join(parts, ", ") + ")"
}

// addBufferDecls offers the names bound in the current buffer. Names already in
// scope (declared above the cursor) rank ahead of everything else.
func (b *completionBuilder) addBufferDecls(c editContext) {
	for _, d := range c.index.decls {
		if d.kind == declParam || d.kind == declLoopVar || d.kind == declCatchVar {
			continue
		}
		if b.seen(d.name) || !b.matches(d.name) {
			continue
		}
		b.markSeen(d.name)

		b.rank = 50
		if d.start <= c.offset {
			b.rank = 10
		}

		item := completionItem{Label: d.name, Kind: d.kind.completionKind()}
		if d.kind == declFunction {
			item.Detail = d.detail
			item.InsertText = callInsertText(d.name, d.params, b.snippets)
			if b.snippets {
				item.InsertTextFormat = completionSnippet
			}
		}
		b.add(item)
	}
}

// addBuiltinGroup offers the builtins of one group. The core group needs no
// import; every other group is only reachable once it has been imported.
func (b *completionBuilder) addBuiltinGroup(group string) {
	if group != "core" && !b.imported[group] {
		return
	}
	b.rank = 60
	for _, blt := range builtinGroupMembers(group) {
		if blt == nil || blt.Name == "" || !b.matches(blt.Name) || b.seen(blt.Name) {
			continue
		}
		b.markSeen(blt.Name)
		item := completionItem{Label: blt.Name, Kind: kindFunction}
		if help := strings.TrimSpace(blt.HelpStr); help != "" {
			item.Documentation = &markupContent{Kind: hoverMarkdown, Value: "```blue\n" + help + "\n```"}
		}
		b.add(item)
	}
}

// addImportedModules offers module names reachable through the buffer's imports.
func (b *completionBuilder) addImportedModules(c editContext) {
	b.rank = 70
	for _, imp := range c.index.imports {
		name := imp.alias
		if name == "" {
			name = lastSegment(imp.path)
		}
		if !b.matches(name) {
			continue
		}
		b.add(completionItem{Label: name, Kind: kindModule, Detail: "module " + imp.path})
	}
}

// addImportCandidates offers importable modules and files at an import site.
func (b *completionBuilder) addImportCandidates(c editContext) {
	b.rank = 5
	for _, name := range stdModuleNames() {
		if !b.matches(name) {
			continue
		}
		b.add(completionItem{Label: name, Kind: kindModule, Detail: "std module"})
	}

	dir := c.doc.src.dir()
	if dir == "" {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".b") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".b")
		if !b.matches(name) {
			continue
		}
		b.add(completionItem{Label: name, Kind: kindModule, Detail: filepath.Join(dir, entry.Name())})
	}
}

// addModuleMembers offers everything a module exposes after `module.`.
func (b *completionBuilder) addModuleMembers(entry *moduleEntry) {
	b.rank = 20
	for _, d := range entry.topLevelDecls() {
		if !b.matches(d.name) || b.seen(d.name) {
			continue
		}
		b.markSeen(d.name)

		item := completionItem{Label: d.name, Kind: d.kind.completionKind(), Detail: d.detail}
		if d.kind == declFunction {
			item.InsertText = callInsertText(d.name, d.params, b.snippets)
			if b.snippets {
				item.InsertTextFormat = completionSnippet
			}
		}
		if d.doc != "" {
			item.Documentation = &markupContent{Kind: hoverMarkdown, Value: d.doc}
		} else if target := aliasedBuiltin(entry.src.runes, entry.ix, d.name); target != "" {
			if builtin, ok := builtinHelp(target); ok && strings.TrimSpace(builtin.HelpStr) != "" {
				help := strings.TrimSpace(builtin.HelpStr)
				item.Detail = firstLine(help)
				item.Documentation = &markupContent{Kind: hoverMarkdown, Value: "```blue\n" + help + "\n```"}
			}
		}
		b.add(item)
	}

	// Some std modules expose members that only exist as go builtins.
	if group := stdGroupName(entry); group != "" {
		for _, blt := range builtinGroupMembers(group) {
			if !b.matches(blt.Name) || b.seen(blt.Name) {
				continue
			}
			b.markSeen(blt.Name)
			item := completionItem{Label: blt.Name, Kind: kindFunction}
			if help := strings.TrimSpace(blt.HelpStr); help != "" {
				item.Documentation = &markupContent{Kind: hoverMarkdown, Value: "```blue\n" + help + "\n```"}
			}
			b.add(item)
		}
	}
}

func firstLine(s string) string {
	if idx := strings.Index(s, "\n"); idx >= 0 {
		return strings.TrimSpace(s[:idx])
	}
	return s
}
