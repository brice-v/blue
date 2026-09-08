package lsp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

// wireMsg is any framed message, request or response, for assertions that have
// to look at raw protocol fields.
type wireMsg struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *responseError  `json:"error,omitempty"`
}

// pipeSession runs a real session over an in-memory pipe pair so framing, the
// handshake and notifications are all exercised the way an editor uses them.
// Writes go through one ordered writer goroutine because net.Pipe is unbuffered:
// a write only completes once the server reads it, so the test must be free to
// read replies at the same time.
type pipeSession struct {
	t      *testing.T
	conn   net.Conn
	in     *bufio.Reader
	served chan error
	out    chan []byte
}

func startPipeSession(t *testing.T, opts Options) *pipeSession {
	t.Helper()
	client, server := net.Pipe()
	ps := &pipeSession{
		t:      t,
		conn:   client,
		in:     bufio.NewReader(client),
		served: make(chan error, 1),
		out:    make(chan []byte, 256),
	}

	go func() { ps.served <- serveConn(context.Background(), server, opts) }()
	go ps.writeLoop()
	return ps
}

// writeLoop pushes queued frames out in order.
func (ps *pipeSession) writeLoop() {
	for frame := range ps.out {
		if _, err := ps.conn.Write(frame); err != nil {
			return
		}
	}
}

// send queues one framed message to the server.
func (ps *pipeSession) send(method string, id any, params any) {
	ps.t.Helper()
	msg := map[string]any{"jsonrpc": "2.0", "method": method}
	if id != nil {
		msg["id"] = id
	}
	if params != nil {
		msg["params"] = params
	}
	body, err := json.Marshal(msg)
	if err != nil {
		ps.t.Fatalf("marshal %s: %s", method, err)
	}
	frame := []byte(fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(body), body))
	select {
	case ps.out <- frame:
	case <-time.After(5 * time.Second):
		ps.t.Fatalf("write queue stuck sending %s", method)
	}
}

// read waits for the next framed message, returning false when nothing arrives.
func (ps *pipeSession) read(timeout time.Duration) (wireMsg, bool) {
	ps.t.Helper()
	var msg wireMsg
	if err := ps.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		ps.t.Fatalf("set read deadline: %s", err)
	}
	raw, err := readFrame(ps.in)
	if err != nil {
		return msg, false
	}
	if err := json.Unmarshal(raw, &msg); err != nil {
		ps.t.Fatalf("undecodable message %q", string(raw))
	}
	return msg, true
}

// responseFor reads messages until the reply with the wanted id shows up,
// handing back anything else (notifications) that arrived along the way.
func (ps *pipeSession) responseFor(id any, timeout time.Duration) (wireMsg, []wireMsg) {
	ps.t.Helper()
	others := []wireMsg{}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		msg, ok := ps.read(time.Until(deadline))
		if !ok {
			break
		}
		if msg.Method != "" {
			others = append(others, msg)
			continue
		}
		if string(msg.ID) == fmt.Sprintf("%v", id) {
			return msg, others
		}
		others = append(others, msg)
	}
	return wireMsg{}, others
}

// readFrame reads one `Content-Length` framed message body off r.
func readFrame(r *bufio.Reader) ([]byte, error) {
	contentLength := -1
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		trimmed := bytes.TrimRight([]byte(line), "\r\n")
		if len(trimmed) == 0 {
			break
		}
		name, value, found := bytes.Cut(trimmed, []byte(":"))
		if !found {
			continue
		}
		if bytes.EqualFold(bytes.TrimSpace(name), []byte("Content-Length")) {
			n, convErr := atoiTrimmed(value)
			if convErr != nil {
				return nil, convErr
			}
			contentLength = n
		}
	}
	if contentLength < 0 {
		return nil, fmt.Errorf("missing Content-Length header")
	}
	body := make([]byte, contentLength)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, err
	}
	return body, nil
}

func atoiTrimmed(b []byte) (int, error) {
	s := string(bytes.TrimSpace(b))
	n := 0
	if s == "" {
		return 0, fmt.Errorf("empty content length")
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("bad content length %q", s)
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

func (ps *pipeSession) initialize(t *testing.T, extra map[string]any) map[string]any {
	t.Helper()
	params := map[string]any{
		"processId":    nil,
		"rootUri":      nil,
		"capabilities": map[string]any{},
	}
	for k, v := range extra {
		params[k] = v
	}
	ps.send("initialize", 1, params)
	reply, _ := ps.responseFor(1, 3*time.Second)
	if reply.Error != nil {
		t.Fatalf("initialize failed: %s", reply.Error.Message)
	}
	var res struct {
		Capabilities map[string]any `json:"capabilities"`
	}
	if err := json.Unmarshal(reply.Result, &res); err != nil {
		t.Fatalf("unmarshal initialize result: %s", err)
	}
	return res.Capabilities
}

func TestEndToEndSession(t *testing.T) {
	ps := startPipeSession(t, Options{DiagnosticsDelay: 0})
	const uri = "file:///tmp/e2e.b"

	caps := ps.initialize(t, nil)
	if _, ok := caps["textDocumentSync"]; !ok {
		t.Error("textDocumentSync missing from initialize result")
	}
	comp, _ := caps["completionProvider"].(map[string]any)
	if comp == nil {
		t.Error("completionProvider missing from initialize result")
	} else if _, ok := comp["resolveProvider"]; ok {
		t.Error("completionProvider advertises resolveProvider although there is no resolver")
	}

	ps.send("initialized", nil, nil)
	ps.send("textDocument/didOpen", nil, map[string]any{
		"textDocument": map[string]any{"uri": uri, "languageId": "blue", "version": 1, "text": "val x =\n"},
	})

	// The broken buffer has to produce exactly one error, on the offending line.
	diagMsg, ok := ps.read(3 * time.Second)
	if !ok {
		t.Fatalf("no diagnostics published for a broken buffer")
	}
	if diagMsg.Method != "textDocument/publishDiagnostics" {
		t.Fatalf("first message = %q, want textDocument/publishDiagnostics", diagMsg.Method)
	}
	var params struct {
		URI         string       `json:"uri"`
		Diagnostics []diagnostic `json:"diagnostics"`
	}
	if err := json.Unmarshal(diagMsg.Params, &params); err != nil {
		t.Fatalf("unmarshal publishDiagnostics: %s", err)
	}
	if params.URI != uri {
		t.Errorf("diagnostics for %q, want %q", params.URI, uri)
	}
	if len(params.Diagnostics) != 1 {
		t.Fatalf("got %d diagnostics, want exactly one: %+v", len(params.Diagnostics), params.Diagnostics)
	}
	if params.Diagnostics[0].Severity != sevError {
		t.Errorf("severity = %d, want error", params.Diagnostics[0].Severity)
	}

	// Fixing the buffer must push an empty array so the editor clears the marker.
	ps.send("textDocument/didChange", nil, map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 2},
		"contentChanges": []map[string]any{
			{"text": "val x = 1\nprint(x)\n"},
		},
	})
	cleared, ok := ps.read(3 * time.Second)
	if !ok {
		t.Fatalf("no diagnostics published after the buffer was fixed")
	}
	var clearedParams struct {
		Diagnostics []diagnostic `json:"diagnostics"`
	}
	if err := json.Unmarshal(cleared.Params, &clearedParams); err != nil {
		t.Fatalf("unmarshal cleared diagnostics: %s", err)
	}
	var diagCheck map[string]json.RawMessage
	if err := json.Unmarshal(cleared.Params, &diagCheck); err != nil {
		t.Fatalf("unmarshal: %s", err)
	}
	if got, present := diagCheck["diagnostics"]; !present {
		t.Error("cleared report has no diagnostics field")
	} else if string(got) != "[]" {
		t.Errorf("cleared report diagnostics = %s, want [] (an empty array clears markers, null does not)", string(got))
	}
	if len(clearedParams.Diagnostics) != 0 {
		t.Errorf("cleared report still has %d diagnostics", len(clearedParams.Diagnostics))
	}

	ps.send("textDocument/didSave", nil, map[string]any{"textDocument": map[string]any{"uri": uri}})

	// Requests over the wire must answer with protocol shaped results.
	ps.send("textDocument/documentSymbol", 2, map[string]any{"textDocument": map[string]any{"uri": uri}})
	symReply, _ := ps.responseFor(2, 3*time.Second)
	if symReply.Error != nil {
		t.Errorf("documentSymbol error: %s", symReply.Error.Message)
	}

	ps.send("workspace/symbol", 3, map[string]any{"query": ""})
	symAll, _ := ps.responseFor(3, 3*time.Second)
	if symAll.Error != nil {
		t.Errorf("workspace/symbol with an empty query errored: %s", symAll.Error.Message)
	}

	ps.send("textDocument/formatting", 4, map[string]any{"textDocument": map[string]any{"uri": uri}})
	fmtReply, _ := ps.responseFor(4, 3*time.Second)
	if fmtReply.Error == nil {
		t.Error("formatting answered although no formatter is registered")
	} else if fmtReply.Error.Code != codeMethodNotFound {
		t.Errorf("formatting error code = %d, want MethodNotFound %d", fmtReply.Error.Code, codeMethodNotFound)
	}

	ps.send("completionItem/resolve", 5, map[string]any{"label": "print"})
	resolveReply, _ := ps.responseFor(5, 3*time.Second)
	if resolveReply.Error == nil {
		t.Error("completionItem/resolve answered although no resolver is registered")
	} else if resolveReply.Error.Code != codeMethodNotFound {
		t.Errorf("resolve error code = %d, want MethodNotFound %d", resolveReply.Error.Code, codeMethodNotFound)
	}

	ps.send("shutdown", 6, nil)
	shut, _ := ps.responseFor(6, 3*time.Second)
	if shut.Error != nil {
		t.Errorf("shutdown errored: %s", shut.Error.Message)
	}
	if string(shut.Result) != "null" {
		t.Errorf("shutdown result = %s, want null", string(shut.Result))
	}

	// Anything sent after shutdown must be rejected rather than processed.
	ps.send("textDocument/completion", 7, map[string]any{
		"textDocument": map[string]any{"uri": uri},
		"position":     map[string]any{"line": 1, "character": 6},
	})
	late, _ := ps.responseFor(7, 2*time.Second)
	if late.Error == nil {
		t.Error("a request after shutdown was answered instead of rejected")
	} else if late.Error.Code != codeInvalidRequest {
		t.Errorf("post shutdown code = %d, want InvalidRequest %d", late.Error.Code, codeInvalidRequest)
	}

	ps.send("exit", nil, nil)
	select {
	case err := <-ps.served:
		if err != nil {
			t.Errorf("serveConn returned error after exit: %s", err)
		}
	case <-time.After(3 * time.Second):
		t.Error("session did not end after the exit notification")
	}
}

// Requests before the handshake must be refused, and notifications other than
// exit dropped instead of crashing.
func TestRequestsBeforeInitialize(t *testing.T) {
	ps := startPipeSession(t, Options{DiagnosticsDelay: 0})

	ps.send("textDocument/completion", 1, map[string]any{
		"textDocument": map[string]any{"uri": "file:///tmp/none.b"},
		"position":     map[string]any{"line": 0, "character": 0},
	})
	reply, _ := ps.responseFor(1, 3*time.Second)
	if reply.Error == nil {
		t.Fatalf("a request before initialize was answered normally: %s", string(reply.Result))
	}
	if reply.Error.Code != codeServerNotInitialized {
		t.Errorf("code = %d, want ServerNotInitialized %d", reply.Error.Code, codeServerNotInitialized)
	}

	ps.send("textDocument/didOpen", nil, map[string]any{
		"textDocument": map[string]any{"uri": "file:///tmp/none.b", "languageId": "blue", "version": 1, "text": "val x =\n"},
	})
	if msg, ok := ps.read(500 * time.Millisecond); ok && msg.Method != "" {
		t.Errorf("a notification before initialize produced %q, want it dropped", msg.Method)
	}

	ps.send("exit", nil, nil)
	select {
	case err := <-ps.served:
		if err != nil {
			t.Errorf("serveConn returned error after exit: %s", err)
		}
	case <-time.After(3 * time.Second):
		t.Error("session did not end after the exit notification")
	}
}

// Garbage on the wire must come back as a parse error without killing the loop.
func TestGarbageOnTheWire(t *testing.T) {
	ps := startPipeSession(t, Options{DiagnosticsDelay: 0})
	ps.initialize(t, nil)

	body := []byte("this is not json")
	frame := []byte(fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(body), body))
	select {
	case ps.out <- frame:
	case <-time.After(5 * time.Second):
		t.Fatal("write queue stuck sending garbage")
	}
	msg, ok := ps.read(3 * time.Second)
	if !ok {
		t.Fatalf("no reply to undecodable input")
	}
	if msg.Error == nil {
		t.Fatalf("undecodable input produced no error: %v", msg)
	}
	if msg.Error.Code != codeParseError {
		t.Errorf("code = %d, want ParseError %d", msg.Error.Code, codeParseError)
	}

	// The session has to keep working afterwards.
	ps.send("textDocument/completion", 2, map[string]any{
		"textDocument": map[string]any{"uri": "file:///tmp/none.b"},
		"position":     map[string]any{"line": 0, "character": 0},
	})
	reply, _ := ps.responseFor(2, 3*time.Second)
	if reply.Error != nil {
		t.Errorf("completion after garbage errored: %s", reply.Error.Message)
	}

	ps.send("exit", nil, nil)
	select {
	case err := <-ps.served:
		if err != nil {
			t.Errorf("serveConn returned error after exit: %s", err)
		}
	case <-time.After(3 * time.Second):
		t.Error("session did not end after the exit notification")
	}
}

// Snippets must only be produced when the client said it can handle them.
func TestEndToEndSnippetsFollowClientCapability(t *testing.T) {
	ps := startPipeSession(t, Options{DiagnosticsDelay: 0})
	const uri = "file:///tmp/snips.b"

	caps := map[string]any{"textDocument": map[string]any{
		"completion": map[string]any{
			"completionItem": map[string]any{"snippetSupport": true},
			"contextSupport": true,
		},
	}}
	ps.initialize(t, map[string]any{"capabilities": caps})
	_ = caps

	ps.send("textDocument/didOpen", nil, map[string]any{
		"textDocument": map[string]any{"uri": uri, "languageId": "blue", "version": 1,
			"text": "fun greet(name) {\n\tprint(name)\n}\n\ngre\n"},
	})

	ps.send("textDocument/completion", 3, map[string]any{
		"textDocument": map[string]any{"uri": uri},
		"position":     map[string]any{"line": 4, "character": 3},
		"context":      map[string]any{"triggerKind": 1},
	})
	reply, _ := ps.responseFor(3, 3*time.Second)
	if reply.Error != nil {
		t.Fatalf("completion errored: %s", reply.Error.Message)
	}

	var list struct {
		IsIncomplete bool `json:"isIncomplete"`
		Items        []struct {
			Label            string `json:"label"`
			InsertText       string `json:"insertText"`
			InsertTextFormat int    `json:"insertTextFormat"`
		} `json:"items"`
	}
	if err := json.Unmarshal(reply.Result, &list); err != nil {
		t.Fatalf("unmarshal completion result: %s", err)
	}
	found := false
	for _, it := range list.Items {
		if it.Label != "greet" {
			continue
		}
		found = true
		if it.InsertTextFormat != completionSnippet {
			t.Errorf("%q insertTextFormat = %d, want snippet", it.Label, it.InsertTextFormat)
		}
		if it.InsertText != "greet(${1:name})" {
			t.Errorf("%q insertText = %q, want %q", it.Label, it.InsertText, "greet(${1:name})")
		}
	}
	if !found {
		t.Errorf("no completion item for `gre` among %d items", len(list.Items))
	}

	ps.send("exit", nil, nil)
	select {
	case err := <-ps.served:
		if err != nil {
			t.Errorf("serveConn returned error after exit: %s", err)
		}
	case <-time.After(3 * time.Second):
		t.Error("session did not end after the exit notification")
	}
}
