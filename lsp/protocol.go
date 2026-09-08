// Package lsp implements a Language Server Protocol server for blue source,
// started with `blue lsp`. It speaks LSP 3.17 over stdio (or TCP) so editors
// such as Neovim, VS Code and Emacs can get diagnostics, completion, hover
// info, symbol outlines and go to definition for .b files.
//
// The wire protocol is implemented directly on top of encoding/json instead of
// pulling in an LSP library so that blue keeps building in vendor mode.
package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// JSON-RPC 2.0 error codes used by this server.
const (
	codeParseError     = -32700
	codeInvalidRequest = -32600
	codeMethodNotFound = -32601
	codeInvalidParams  = -32602
	codeInternalError  = -32603

	// LSP specific codes. Requests must not be answered before initialize, and
	// nothing may be answered after shutdown except the exit notification.
	codeServerNotInitialized = -32002
	codeRequestCancelled     = -32800
)

// request is a JSON-RPC 2.0 request or notification as sent by the client.
type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

// isRequest reports whether the message expects a response (as opposed to
// being a notification, which never gets one).
func (r *request) isRequest() bool {
	return len(r.ID) > 0 && string(r.ID) != "null"
}

// response is a JSON-RPC 2.0 response written by the server.
type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result"`
	Error   *responseError  `json:"error,omitempty"`
}

type responseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// notification builds a JSON-RPC notification message body.
type notification struct {
	Method string `json:"method"`
	Params any    `json:"params,omitempty"`
}

// Positions and ranges. LSP counts lines from 0 and characters within a line
// as UTF-16 code units, which is what these structs carry.

type position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type rangeStruct struct {
	Start position `json:"start"`
	End   position `json:"end"`
}

type textDocumentIdentifier struct {
	URI string `json:"uri"`
}

type versionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

type textDocumentPositionParams struct {
	TextDocument textDocumentIdentifier `json:"textDocument"`
	Position     position               `json:"position"`
}

// Text document synchronization.

type textDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

type didOpenParams struct {
	TextDocument textDocumentItem `json:"textDocument"`
}

type didChangeParams struct {
	TextDocument   versionedTextDocumentIdentifier `json:"textDocument"`
	ContentChanges []contentChange                 `json:"contentChanges"`
}

// contentChange carries the new text. When Range is nil the change replaces the
// whole document (TextDocumentSyncKind.Full).
type contentChange struct {
	Range *rangeStruct `json:"range,omitempty"`
	Text  string       `json:"text"`
}

type didSaveParams struct {
	TextDocument textDocumentIdentifier `json:"textDocument"`
	Text         *string                `json:"text,omitempty"`
}

type closeParams struct {
	TextDocument textDocumentIdentifier `json:"textDocument"`
}

// Initialize.

type initializeParams struct {
	ProcessID             int             `json:"processId"`
	RootURI               string          `json:"rootUri"`
	RootPath              string          `json:"rootPath"`
	Capabilities          json.RawMessage `json:"capabilities"`
	InitializationOptions json.RawMessage `json:"initializationOptions"`
}

type serverCapabilities struct {
	TextDocumentSync          *textDocumentSyncOptions `json:"textDocumentSync,omitempty"`
	CompletionProvider        *completionOptions       `json:"completionProvider,omitempty"`
	HoverProvider             bool                     `json:"hoverProvider"`
	DefinitionProvider        bool                     `json:"definitionProvider"`
	ReferencesProvider        bool                     `json:"referencesProvider"`
	DocumentSymbolProvider    bool                     `json:"documentSymbolProvider"`
	DocumentHighlightProvider bool                     `json:"documentHighlightProvider"`
	SemanticTokensProvider    *semanticTokensOptions   `json:"semanticTokensProvider,omitempty"`
}

// semanticTokensLegend lists the type and modifier names the server may emit. The
// client indexes into these, so the encoder has to use exactly these positions.
type semanticTokensLegend struct {
	TokenTypes     []string `json:"tokenTypes"`
	TokenModifiers []string `json:"tokenModifiers"`
}

type semanticTokensOptions struct {
	Legend semanticTokensLegend `json:"legend"`
	// Range stays false because only whole document tokens are produced.
	Range bool `json:"range"`
	// Full is true without delta support: the client asks again after each edit
	// instead of sending a previously returned id.
	Full bool `json:"full"`
}

type textDocumentSemanticTokensParams struct {
	TextDocument textDocumentIdentifier `json:"textDocument"`
}

const (
	syncKindNone      = 0
	syncKindFull      = 1
	syncKindIncr      = 2
	hoverPlainText    = "plaintext"
	hoverMarkdown     = "markdown"
	completionSnippet = 2
)

type textDocumentSyncOptions struct {
	OpenClose bool         `json:"openClose"`
	Change    int          `json:"change"`
	Save      *saveOptions `json:"save,omitempty"`
}

type saveOptions struct {
	IncludeText bool `json:"includeText"`
}

// completionOptions advertises completion support. There is deliberately no
// resolveProvider because items already carry their documentation inline, so a
// second round trip per item would buy nothing.
type completionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

type initializeResult struct {
	Capabilities serverCapabilities `json:"capabilities"`
	ServerInfo   *serverInfo        `json:"serverInfo,omitempty"`
}

type serverInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Hover.

type textDocumentHoverParams struct {
	TextDocument textDocumentIdentifier `json:"textDocument"`
	Position     position               `json:"position"`
}

type markupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type hoverResult struct {
	Contents markupContent `json:"contents"`
	Range    *rangeStruct  `json:"range,omitempty"`
}

// Completion.

const (
	// CompletionItemKind values from the LSP spec. These are deliberately kept
	// apart from the SymbolKind codes used by document symbols because the two
	// enumerations share nothing but a name: Function is 3 here while it is 12
	// in SymbolKind, and Constant is 21 versus 14.
	kindFunction = 3
	kindVariable = 6
	kindModule   = 9
	kindKeyword  = 14
	kindConstant = 21
)

type completionParams struct {
	TextDocument textDocumentIdentifier `json:"textDocument"`
	Position     position               `json:"position"`
	Context      *completionContext     `json:"context,omitempty"`
}

type completionContext struct {
	TriggerKind      int    `json:"triggerKind"`
	TriggerCharacter string `json:"triggerCharacter,omitempty"`
}

type completionItem struct {
	Label            string         `json:"label"`
	Kind             int            `json:"kind,omitempty"`
	Detail           string         `json:"detail,omitempty"`
	Documentation    *markupContent `json:"documentation,omitempty"`
	SortText         string         `json:"sortText,omitempty"`
	FilterText       string         `json:"filterText,omitempty"`
	InsertText       string         `json:"insertText,omitempty"`
	InsertTextFormat int            `json:"insertTextFormat,omitempty"`
}

type completionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []completionItem `json:"items"`
}

// Locations and symbols.

type location struct {
	URI   string      `json:"uri"`
	Range rangeStruct `json:"range"`
}

type documentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           int              `json:"kind"`
	Range          rangeStruct      `json:"range"`
	SelectionRange rangeStruct      `json:"selectionRange"`
	Children       []documentSymbol `json:"children,omitempty"`
}

type symbolInformation struct {
	Name     string   `json:"name"`
	Kind     int      `json:"kind"`
	Location location `json:"location"`
}

type workspaceSymbolParams struct {
	Query string `json:"query"`
}

type referenceParams struct {
	TextDocument textDocumentIdentifier `json:"textDocument"`
	Position     position               `json:"position"`
	Context      *referenceContext      `json:"context,omitempty"`
}

type referenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}

type documentHighlightParams = textDocumentPositionParams

type documentHighlightKind int

const (
	highlightText  documentHighlightKind = 1
	highlightRead  documentHighlightKind = 2
	highlightWrite documentHighlightKind = 3
)

type documentHighlight struct {
	Range rangeStruct `json:"range"`
	Kind  int         `json:"kind,omitempty"`
}

// Diagnostics.

const (
	sevError       = 1
	sevWarning     = 2
	sevInformation = 3
	sevHint        = 4
)

type diagnostic struct {
	Range    rangeStruct `json:"range"`
	Severity int         `json:"severity,omitempty"`
	Code     any         `json:"code,omitempty"`
	Source   string      `json:"source,omitempty"`
	Message  string      `json:"message"`
}

// readMessage reads one framed LSP message off r. Framing is the standard
// "Content-Length: N\r\n\r\n" followed by N bytes of JSON.
func readMessage(r *bufio.Reader) (*request, error) {
	contentLength := -1
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		name, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
			n, convErr := strconv.Atoi(strings.TrimSpace(value))
			if convErr != nil || n < 0 {
				return nil, fmt.Errorf("bad Content-Length header %q", value)
			}
			contentLength = n
		}
	}

	if contentLength < 0 {
		return nil, fmt.Errorf("missing Content-Length header")
	}
	if contentLength == 0 {
		return nil, fmt.Errorf("empty message body")
	}
	if contentLength > maxMessageBytes {
		return nil, fmt.Errorf("message too large (%d bytes)", contentLength)
	}

	body := make([]byte, contentLength)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, err
	}

	msg := &request{}
	if err := json.Unmarshal(body, msg); err != nil {
		return &request{JSONRPC: "2.0", Method: methodParseFailure}, err
	}
	return msg, nil
}

// methodParseFailure is an internal marker method used when the client sends
// invalid JSON so the dispatch loop can reply with a parse error.
const methodParseFailure = "$/parse-failure"

const maxMessageBytes = 64 << 20 // 64 MiB guard against runaway messages

// writeMessage writes one framed message to w.
func writeMessage(w *bufio.Writer, v any) error {
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Content-Length: %d\r\n\r\n", len(body)); err != nil {
		return err
	}
	if _, err := w.Write(body); err != nil {
		return err
	}
	return w.Flush()
}
