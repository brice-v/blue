package vm

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"blue/compiler"
	"blue/lexer"
	"blue/object"
	"blue/parser"

	"github.com/gofiber/fiber/v3"
	ws "github.com/gorilla/websocket"
)

func compileClosure(t *testing.T, input string) (*object.Closure, *VM) {
	t.Helper()
	l := lexer.New(input, "<test>")
	p := parser.New(l)
	program := p.ParseProgram()
	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	vm := New(comp.Bytecode())
	if err := vm.Run(); err != nil {
		t.Fatalf("vm error: %s", err)
	}
	cl, ok := vm.LastPoppedStackElem().(*object.Closure)
	if !ok {
		t.Fatalf("expected closure, got %T", vm.LastPoppedStackElem())
	}
	return cl, vm
}

func compileClosureWithCore(t *testing.T, input string) (*object.Closure, *VM) {
	t.Helper()
	l := lexer.New(input, "<test>")
	p := parser.New(l)
	program := p.ParseProgram()
	comp := compiler.NewFromCore()
	if err := comp.Compile(program); err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	vm := New(comp.Bytecode())
	if err := vm.Run(); err != nil {
		t.Fatalf("vm error: %s", err)
	}
	cl, ok := vm.LastPoppedStackElem().(*object.Closure)
	if !ok {
		t.Fatalf("expected closure, got %T", vm.LastPoppedStackElem())
	}
	return cl, vm
}

func TestHandleSseStreamsEvents(t *testing.T) {
	app := fiber.New()
	cl, vm := compileClosureWithCore(t, `import http
	fun(sse, room, query_params=["name"]) {
		from http import {sse_send, sse_last_event_id}
		from time import {sleep}
		sse.send("hello");
		val msg = "#{room}:#{name}:#{sse.last_event_id()}:#{sse_last_event_id(sse.v)}";
		sse.send(msg);
		sleep(500);
		sse_send(sse.v, "line1\nline2", event="update", id="7", retry=3000);
	}`)
	normalCl, normalVm := compileClosureWithCore(t, `fun() { "ok" }`)

	builtin := createHttpHandleSSEBuiltin(vm)
	res := builtin.Fun(object.NewGoObj(app), &object.Stringo{Value: "/events/:room"}, cl)
	if isError(res) {
		t.Fatalf("handle_sse error: %s", res.(*object.Error).Message)
	}
	normalBuiltin := createHttpHandleBuiltin(normalVm, false)
	res = normalBuiltin.Fun(object.NewGoObj(app), &object.Stringo{Value: "/normal"}, normalCl, &object.Stringo{Value: "GET"})
	if isError(res) {
		t.Fatalf("handle error: %s", res.(*object.Error).Message)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	listenerDone := make(chan struct{})
	go func() {
		defer close(listenerDone)
		_ = app.Listener(ln)
	}()
	t.Cleanup(func() {
		_ = app.Shutdown()
		<-listenerDone
	})

	client := &http.Client{Timeout: 5 * time.Second}
	start := time.Now()
	req, err := http.NewRequest(http.MethodGet, "http://"+ln.Addr().String()+"/events/lobby?name=brice", http.NoBody)
	if err != nil {
		t.Fatalf("create sse request: %s", err)
	}
	req.Header.Set("Last-Event-ID", "abc")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("get sse: %s", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Logf("failed to close sse response body: %s", closeErr.Error())
		}
	}()
	if got := resp.Header.Get("Content-Type"); got != "text/event-stream" {
		t.Errorf("Content-Type = %q, want text/event-stream", got)
	}
	if got := resp.Header.Get("Cache-Control"); got != "no-cache" {
		t.Errorf("Cache-Control = %q, want no-cache", got)
	}

	reader := bufio.NewScanner(resp.Body)
	if !reader.Scan() {
		t.Fatalf("expected first sse event, scan error: %v", reader.Err())
	}
	if elapsed := time.Since(start); elapsed > 300*time.Millisecond {
		t.Errorf("first event took %s, should be streamed before the 500ms sleep", elapsed)
	}
	first := reader.Text()
	if first != "data: hello" {
		t.Errorf("first event = %q, want %q", first, "data: hello")
	}

	normalResp, err := client.Get("http://" + ln.Addr().String() + "/normal")
	if err != nil {
		t.Fatalf("get normal while sse open: %s", err)
	}
	normalBody, err := io.ReadAll(normalResp.Body)
	_ = normalResp.Body.Close()
	if err != nil {
		t.Fatalf("read normal response: %s", err)
	}
	if normalResp.StatusCode != http.StatusOK || !strings.Contains(string(normalBody), "ok") {
		t.Errorf("normal response = %d %q, want 200 containing ok", normalResp.StatusCode, string(normalBody))
	}

	lines := []string{first}
	for reader.Scan() {
		lines = append(lines, reader.Text())
	}
	if err := reader.Err(); err != nil {
		t.Fatalf("scan sse: %s", err)
	}
	want := []string{"data: hello", "", "data: lobby:brice:abc:abc", "", "id: 7", "event: update", "retry: 3000", "data: line1", "data: line2", ""}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Errorf("sse events = %q, want %q", lines, want)
	}
}

func TestHandleSseValidation(t *testing.T) {
	cl, vm := compileClosureWithCore(t, `fun(sse) { }`)
	builtin := createHttpHandleSSEBuiltin(vm)
	app := object.NewGoObj(fiber.New())

	noParams, _ := compileClosureWithCore(t, `fun() { }`)
	res := builtin.Fun(app, &object.Stringo{Value: "/events"}, noParams)
	if !isError(res) {
		t.Errorf("handle_sse with no params should error, got %v", res.Inspect())
	}
	ctxParam, _ := compileClosureWithCore(t, `fun(sse, ctx) { }`)
	res = builtin.Fun(app, &object.Stringo{Value: "/events"}, ctxParam)
	if !isError(res) {
		t.Errorf("handle_sse with ctx param should error, got %v", res.Inspect())
	}
	res = builtin.Fun(object.NewGoObj("not an app"), &object.Stringo{Value: "/events"}, cl)
	if !isError(res) {
		t.Errorf("handle_sse with wrong app should error, got %v", res.Inspect())
	}
	res = builtin.Fun(app, &object.Integer{Value: 1}, cl)
	if !isError(res) {
		t.Errorf("handle_sse with non-string pattern should error, got %v", res.Inspect())
	}
	res = builtin.Fun(app, &object.Stringo{Value: "/events"}, &object.Integer{Value: 1})
	if !isError(res) {
		t.Errorf("handle_sse with non-closure should error, got %v", res.Inspect())
	}
	res = builtin.Fun(app)
	if !isError(res) {
		t.Errorf("handle_sse with no args should error, got %v", res.Inspect())
	}
}

func TestHandleWsPassesRouteParams(t *testing.T) {
	app := fiber.New()
	cl, vm := compileClosure(t, `fun(ws, room) { room }`)

	builtin := createHttpHandleWSBuiltin(vm)
	res := builtin.Fun(object.NewGoObj(app), &object.Stringo{Value: "/ws/:room"}, cl)
	if isError(res) {
		t.Fatalf("handle_ws error: %s", res.(*object.Error).Message)
	}

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		os.Stdout = oldStdout
		_ = w.Close()
		_ = r.Close()
		t.Fatal(err)
	}

	listenerDone := make(chan struct{})
	go func() {
		defer close(listenerDone)
		_ = app.Listener(ln)
	}()

	t.Cleanup(func() {
		_ = app.Shutdown()
		<-listenerDone
		os.Stdout = oldStdout
		_ = w.Close()
		_ = r.Close()
	})

	conn, _, err := ws.DefaultDialer.Dial("ws://"+ln.Addr().String()+"/ws/lobby", nil)
	if err != nil {
		t.Fatalf("dial error: %s", err)
	}
	_, _, _ = conn.ReadMessage()
	_ = conn.Close()

	_ = w.Close()
	out, _ := io.ReadAll(r)
	if !strings.Contains(string(out), "lobby") {
		t.Fatalf("ws handler did not receive room param, output: %q", string(out))
	}
}

func TestHttpRequestHandlersAreSerialized(t *testing.T) {
	app := fiber.New()
	cl, vm := compileClosure(t, `fun() { 123 }`)

	builtin := createHttpHandleBuiltin(vm, false)
	res := builtin.Fun(object.NewGoObj(app), &object.Stringo{Value: "/n"}, cl, &object.Stringo{Value: "GET"})
	if isError(res) {
		t.Fatalf("handle error: %s", res.(*object.Error).Message)
	}

	const n = 64
	codes := make([]int, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/n", nil))
			if err != nil {
				t.Errorf("request %d errored: %s", i, err)
				return
			}
			codes[i] = resp.StatusCode
		}(i)
	}
	wg.Wait()
	for i, c := range codes {
		if c != http.StatusOK {
			t.Fatalf("request %d got code %d, want %d", i, c, http.StatusOK)
		}
	}
}

func TestCloneHandlerClosureIndependence(t *testing.T) {
	cl, _ := compileClosure(t, `fun(a, b) { a }`)
	key := object.NameIndexKey{Name: "query_params", Index: 0}
	cl.Fun.SpecialFunctionParameters = map[object.NameIndexKey]map[object.NameIndexKey]object.Object{
		key: {object.NameIndexKey{Name: "x", Index: 0}: &object.Stringo{Value: "orig"}},
	}

	c1 := cloneHandlerClosure(cl)
	if c1 == cl || c1.Fun == cl.Fun {
		t.Fatal("clone did not produce a new closure and compiled function")
	}
	if &c1.Fun.Instructions[0] != &cl.Fun.Instructions[0] {
		t.Fatal("bytecode is immutable and should be shared with the original")
	}
	inner := c1.Fun.SpecialFunctionParameters[key]
	inner[object.NameIndexKey{Name: "x", Index: 0}] = &object.Stringo{Value: "changed"}
	if cl.Fun.SpecialFunctionParameters[key][object.NameIndexKey{Name: "x", Index: 0}].(*object.Stringo).Value != "orig" {
		t.Fatal("mutating clone special params affected the original")
	}
}
