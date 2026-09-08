package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

// Options configure a blue language server.
type Options struct {
	// Addr listens on TCP instead of stdio. One client is served at a time,
	// which is all an editor needs and keeps session state simple.
	Addr string

	// DiagnosticsDelay debounces republishing diagnostics after edits so that
	// typing does not re-analyze after every keystroke. Zero means publish
	// immediately, which tests rely on.
	DiagnosticsDelay time.Duration

	// Trace turns on logging of everything crossing the wire to stderr. It never
	// writes to stdout because that is the protocol stream.
	Trace bool
}

// Run starts a blue language server and blocks until ctx is cancelled or the
// client sends an exit notification. Without Options.Addr it speaks LSP over
// stdin and stdout, which is how editors launch servers.
func Run(ctx context.Context, opts Options) error {
	if opts.Addr != "" {
		return runTCP(ctx, opts)
	}
	return serveConn(ctx, stdioRWC{}, opts)
}

// stdioRWC adapts os.Stdin and os.Stdout to the ReadWriteCloser a session needs.
type stdioRWC struct{}

func (stdioRWC) Read(p []byte) (int, error)  { return os.Stdin.Read(p) }
func (stdioRWC) Write(p []byte) (int, error) { return os.Stdout.Write(p) }
func (stdioRWC) Close() error                { return nil }

// runTCP listens on opts.Addr and serves clients one at a time.
func runTCP(ctx context.Context, opts Options) error {
	listener, err := net.Listen("tcp", opts.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", opts.Addr, err)
	}
	defer func() {
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return nil
			}
			fmt.Fprintf(os.Stderr, "blue lsp: accept failed: %s\n", err.Error())
			continue
		}
		if err := serveConn(ctx, conn, opts); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			fmt.Fprintf(os.Stderr, "blue lsp: session ended: %s\n", err.Error())
		}
	}
}

// serveConn runs a single session over a connection. Cancelling ctx closes the
// connection, which unblocks the pending read.
func serveConn(ctx context.Context, rwc io.ReadWriteCloser, opts Options) error {
	s := newSession(rwc, opts)

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = rwc.Close()
		case <-done:
		}
	}()

	err := s.loop()
	close(done)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return err
}

// session is one LSP connection plus all of the editor state that comes with it.
type session struct {
	opts Options
	rwc  io.ReadWriteCloser
	in   *bufio.Reader
	out  *bufio.Writer

	writeMu sync.Mutex // only one goroutine may write to out at a time

	mu   sync.Mutex
	docs map[string]*document
	root string

	modules        *moduleCache
	clientSnippets bool
	started        bool
	shuttingDown   bool
	exited         bool
}

// document is an open buffer plus the analysis derived from its contents.
type document struct {
	uri     string
	name    string // name given to blue's lexer for error reporting
	src     *docSource
	index   *fileIndex
	version int
	timer   *time.Timer
}

func newSession(rwc io.ReadWriteCloser, opts Options) *session {
	return &session{
		opts:    opts,
		rwc:     rwc,
		in:      bufio.NewReader(rwc),
		out:     bufio.NewWriter(rwc),
		docs:    map[string]*document{},
		modules: newModuleCache(),
	}
}

// loop reads and answers messages until the transport closes or the client exits.
func (s *session) loop() error {
	for {
		msg, err := readMessage(s.in)
		if msg == nil {
			if err != nil && !errors.Is(err, io.EOF) && s.opts.Trace {
				fmt.Fprintf(os.Stderr, "blue lsp: read failed: %s\n", err.Error())
			}
			return nil
		}

		if msg.Method == methodParseFailure {
			s.writeResponse(msg.ID, nil, &responseError{Code: codeParseError, Message: "invalid json: " + err.Error()})
			continue
		}

		if s.opts.Trace {
			fmt.Fprintf(os.Stderr, "-> %s\n", msg.Method)
		}

		// Nothing may be answered before the handshake finishes: requests fail with
		// ServerNotInitialized and notifications other than exit are dropped, so an
		// early client never sees half built state.
		if !s.started {
			if !msg.isRequest() {
				if msg.Method != "exit" {
					continue
				}
			} else if msg.Method != "initialize" {
				s.writeResponse(msg.ID, nil, &responseError{Code: codeServerNotInitialized, Message: "server is not initialized"})
				continue
			}
		}

		if !msg.isRequest() {
			s.handleNotification(msg)
			if s.exited {
				return nil
			}
			continue
		}

		if s.shuttingDown {
			s.writeResponse(msg.ID, nil, &responseError{Code: codeInvalidRequest, Message: "server is shutting down"})
			continue
		}

		result, rpcErr := s.dispatch(msg.Method, msg.Params)
		s.writeResponse(msg.ID, result, rpcErr)

		if msg.Method == "initialize" {
			s.started = true
		}

		if s.exited {
			return nil
		}
	}
}

// writeResponse sends a result or an error back for a request id.
func (s *session) writeResponse(id json.RawMessage, result any, rpcErr *responseError) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if err := writeMessage(s.out, response{JSONRPC: "2.0", ID: id, Result: result, Error: rpcErr}); err != nil {
		fmt.Fprintf(os.Stderr, "blue lsp: failed to write response: %s\n", err.Error())
	}
}

// sendNotification writes a notification message to the client.
func (s *session) sendNotification(method string, params any) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if err := writeMessage(s.out, notification{Method: method, Params: params}); err != nil {
		fmt.Fprintf(os.Stderr, "blue lsp: failed to write notification: %s\n", err.Error())
	}
}

// handleNotification deals with messages that never get a reply.
func (s *session) handleNotification(msg *request) {
	switch msg.Method {
	case "exit":
		s.exited = true

	case "textDocument/didOpen":
		var params didOpenParams
		if err := json.Unmarshal(paramsOf(msg), &params); err != nil {
			return
		}
		s.openDocument(params)

	case "textDocument/didChange":
		var params didChangeParams
		if err := json.Unmarshal(paramsOf(msg), &params); err != nil {
			return
		}
		s.changeDocument(params)

	case "textDocument/didSave":
		var params didSaveParams
		if err := json.Unmarshal(paramsOf(msg), &params); err != nil {
			return
		}
		s.saveDocument(params)

	case "textDocument/didClose":
		var params closeParams
		if err := json.Unmarshal(paramsOf(msg), &params); err != nil {
			return
		}
		s.closeDocument(params)

	default:
		// Unknown notifications (progress, cancellation, configuration changes)
		// are safely ignored.
	}
}

// dispatch routes a request to its handler, turning bad parameters into protocol
// errors instead of panics.
func (s *session) dispatch(method string, params json.RawMessage) (any, *responseError) {
	if method == "shutdown" {
		s.shuttingDown = true
		return nil, nil
	}

	switch method {
	case "initialize":
		var p initializeParams
		if err := json.Unmarshal(paramsOf(&request{Params: params}), &p); err != nil {
			return nil, &responseError{Code: codeInvalidParams, Message: err.Error()}
		}
		return s.initialize(p), nil

	case "textDocument/completion":
		var p completionParams
		if err := json.Unmarshal(paramsOf(&request{Params: params}), &p); err != nil {
			return nil, &responseError{Code: codeInvalidParams, Message: err.Error()}
		}
		return s.completion(p), nil

	case "textDocument/hover":
		var p textDocumentHoverParams
		if err := json.Unmarshal(paramsOf(&request{Params: params}), &p); err != nil {
			return nil, &responseError{Code: codeInvalidParams, Message: err.Error()}
		}
		return s.hover(p), nil

	case "textDocument/definition":
		var p textDocumentPositionParams
		if err := json.Unmarshal(paramsOf(&request{Params: params}), &p); err != nil {
			return nil, &responseError{Code: codeInvalidParams, Message: err.Error()}
		}
		return s.definition(p), nil

	case "textDocument/references":
		var p referenceParams
		if err := json.Unmarshal(paramsOf(&request{Params: params}), &p); err != nil {
			return nil, &responseError{Code: codeInvalidParams, Message: err.Error()}
		}
		return s.references(p), nil

	case "textDocument/documentSymbol":
		var p textDocumentIdentifier
		if err := json.Unmarshal(paramsOf(&request{Params: params}), &p); err != nil {
			return nil, &responseError{Code: codeInvalidParams, Message: err.Error()}
		}
		return s.documentSymbol(p), nil

	case "textDocument/documentHighlight":
		var p documentHighlightParams
		if err := json.Unmarshal(paramsOf(&request{Params: params}), &p); err != nil {
			return nil, &responseError{Code: codeInvalidParams, Message: err.Error()}
		}
		return s.documentHighlight(p), nil

	case "textDocument/semanticTokens/full":
		var p textDocumentSemanticTokensParams
		if err := json.Unmarshal(paramsOf(&request{Params: params}), &p); err != nil {
			return nil, &responseError{Code: codeInvalidParams, Message: err.Error()}
		}
		return s.semanticTokensFull(p.TextDocument), nil

	case "workspace/symbol":
		var p workspaceSymbolParams
		if err := json.Unmarshal(paramsOf(&request{Params: params}), &p); err != nil {
			return nil, &responseError{Code: codeInvalidParams, Message: err.Error()}
		}
		return s.workspaceSymbol(p), nil

	default:
		return nil, &responseError{Code: codeMethodNotFound, Message: fmt.Sprintf("method not found: %s", method)}
	}
}

// paramsOf returns the raw parameters of a request, or an empty object when the
// client left them out.
func paramsOf(msg *request) json.RawMessage {
	if len(msg.Params) == 0 {
		return json.RawMessage("{}")
	}
	return msg.Params
}

// openDocument starts tracking a buffer the editor opened and analyzes it.
func (s *session) openDocument(params didOpenParams) {
	item := params.TextDocument
	name := item.URI
	if path, ok := uriToPath(item.URI); ok {
		name = path
	}

	s.mu.Lock()
	s.docs[item.URI] = &document{uri: item.URI, name: name, version: item.Version}
	s.mu.Unlock()

	s.refreshDocument(item.URI, item.Text)
}

// changeDocument applies content changes. The session advertises full sync, so a
// change without a range replaces the whole buffer; ranged changes are applied
// too for clients that send them anyway.
func (s *session) changeDocument(params didChangeParams) {
	uri := params.TextDocument.URI

	s.mu.Lock()
	doc := s.docs[uri]
	if doc == nil {
		s.mu.Unlock()
		return
	}
	doc.version = params.TextDocument.Version
	text := doc.src.text
	s.mu.Unlock()

	for _, change := range params.ContentChanges {
		if change.Range == nil {
			text = change.Text
			continue
		}
		from := doc.src.offsetOf(change.Range.Start)
		to := doc.src.offsetOf(change.Range.End)
		if from > to {
			from, to = to, from
		}
		runes := []rune(text)
		text = string(runes[:from]) + change.Text + string(runes[to:])
	}

	s.refreshDocument(uri, text)
}

// saveDocument re-analyzes a buffer after an explicit save.
func (s *session) saveDocument(params didSaveParams) {
	uri := params.TextDocument.URI

	s.mu.Lock()
	doc := s.docs[uri]
	if doc != nil && params.Text != nil {
		doc.src = newDocSource(doc.name, *params.Text)
	}
	s.mu.Unlock()

	if doc == nil {
		return
	}
	s.scheduleDiagnostics(uri)
}

// closeDocument stops tracking a buffer and clears its diagnostics.
func (s *session) closeDocument(params closeParams) {
	uri := params.TextDocument.URI

	s.mu.Lock()
	doc := s.docs[uri]
	if doc != nil && doc.timer != nil {
		doc.timer.Stop()
	}
	delete(s.docs, uri)
	s.mu.Unlock()

	if doc == nil {
		return
	}
	s.sendNotification("textDocument/publishDiagnostics", struct {
		URI         string       `json:"uri"`
		Diagnostics []diagnostic `json:"diagnostics"`
	}{URI: uri, Diagnostics: []diagnostic{}})
}

// refreshDocument rebuilds a buffer and its index, then schedules diagnostics.
func (s *session) refreshDocument(uri string, text string) {
	s.mu.Lock()
	doc := s.docs[uri]
	if doc == nil {
		s.mu.Unlock()
		return
	}
	doc.src = newDocSource(doc.name, text)
	doc.index = buildIndex(doc.src)
	doc.index.resolveExtents()
	// Anything answered from another file may be stale now.
	s.modules.dropInvalid()

	delay := s.opts.DiagnosticsDelay
	if doc.timer != nil {
		doc.timer.Stop()
	}
	if delay <= 0 {
		s.mu.Unlock()
		s.publishDiagnostics(uri)
		return
	}
	doc.timer = time.AfterFunc(delay, func() { s.publishDiagnostics(uri) })
	s.mu.Unlock()
}

// scheduleDiagnostics queues a diagnostics pass for a buffer.
func (s *session) scheduleDiagnostics(uri string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	doc := s.docs[uri]
	if doc == nil {
		return
	}
	if doc.timer != nil {
		doc.timer.Stop()
	}
	if s.opts.DiagnosticsDelay <= 0 {
		go s.publishDiagnostics(uri)
		return
	}
	doc.timer = time.AfterFunc(s.opts.DiagnosticsDelay, func() { s.publishDiagnostics(uri) })
}

// documentFor looks up an open buffer.
func (s *session) documentFor(uri string) *document {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.docs[uri]
}
