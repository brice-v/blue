package vm

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"blue/compiler"
	"blue/lexer"
	"blue/object"
	"blue/parser"

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

func TestHandleWsPassesRouteParams(t *testing.T) {
	s := object.NewServer()
	cl, vm := compileClosure(t, `fun(ws, room) { room }`)

	builtin := createHttpHandleWSBuiltin(vm)
	res := builtin.Fun(object.NewGoObj(s), &object.Stringo{Value: "/ws/:room"}, cl)
	if isError(res) {
		t.Fatalf("handle_ws error: %s", res.(*object.Error).Message)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := &http.Server{Handler: s}
	go func() { _ = srv.Serve(ln) }()
	defer func() { _ = srv.Close() }()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	conn, _, err := ws.DefaultDialer.Dial("ws://"+ln.Addr().String()+"/ws/lobby", nil)
	if err != nil {
		t.Fatalf("dial error: %s", err)
	}
	_, _, _ = conn.ReadMessage()
	_ = conn.Close()

	_ = w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)
	if !strings.Contains(string(out), "lobby") {
		t.Fatalf("ws handler did not receive room param, output: %q", string(out))
	}
}

func TestHttpRequestHandlersAreSerialized(t *testing.T) {
	s := object.NewServer()
	cl, vm := compileClosure(t, `fun() { 123 }`)

	builtin := createHttpHandleBuiltin(vm, false)
	res := builtin.Fun(object.NewGoObj(s), &object.Stringo{Value: "/n"}, cl, &object.Stringo{Value: "GET"})
	if isError(res) {
		t.Fatalf("handle error: %s", res.(*object.Error).Message)
	}

	const n = 64
	codes := make([]int, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			rec := httptest.NewRecorder()
			s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/n", nil))
			codes[i] = rec.Code
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
