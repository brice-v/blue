package object

import (
	"log"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

func netBuiltinFn(t *testing.T, name string) BuiltinFunction {
	t.Helper()
	for _, b := range NetBuiltins {
		if b.Name == name {
			if b.Fun == nil {
				t.Fatalf("net builtin %q has nil Fun", name)
			}
			return b.Fun
		}
	}
	t.Fatalf("net builtin %q not found", name)
	return nil
}

func goObjValue[T any](t *testing.T, res Object) T {
	t.Helper()
	m, ok := res.(*Map)
	if !ok {
		t.Fatalf("expected MAP result, got %T %v", res, res.Inspect())
	}
	v := mapGetString(t, m, "v")
	g, ok := v.(*GoObj[T])
	if !ok {
		t.Fatalf("expected GoObj value field, got %T", v)
	}
	return g.Value
}

func TestNetRegistry(t *testing.T) {
	seen := make(map[string]bool)
	for _, b := range NetBuiltins {
		if b.Name == "" || b.HelpStr == "" {
			t.Fatalf("net builtin %q missing Name or HelpStr", b.Name)
		}
		if seen[b.Name] {
			t.Errorf("duplicate net builtin name %q", b.Name)
		}
		seen[b.Name] = true
		if b.Fun == nil {
			t.Errorf("net builtin %q has nil Fun", b.Name)
		}
	}
}

func TestConnectBuiltin(t *testing.T) {
	connectFn := netBuiltinFn(t, "_connect")
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if closeErr := l.Close(); closeErr != nil {
			log.Printf("Failed to close listener, error: %s", closeErr.Error())
		}
	}()
	go func() {
		c, aerr := l.Accept()
		if aerr == nil {
			if closeErr := c.Close(); closeErr != nil {
				log.Printf("Failed to close accepted connection, error: %s", closeErr.Error())
			}
		}
	}()
	host, port, _ := net.SplitHostPort(l.Addr().String())

	res := connectFn(&Stringo{Value: "tcp"}, &Stringo{Value: host}, &Stringo{Value: port})
	conn := goObjValue[net.Conn](t, res)
	if conn == nil {
		t.Fatal("connect returned nil conn")
	}
	if ts := mapGetString(t, res.(*Map), "t").(*Stringo).Value; ts != "net" {
		t.Errorf("connect 't' field = %q, want 'net'", ts)
	}
	if res2 := netBuiltinFn(t, "_net_close")(NewGoObj(conn), &Stringo{Value: "net"}); res2 != NULL {
		t.Errorf("net_close(conn) = %s, want NULL", res2.Inspect())
	}

	refused := connectFn(&Stringo{Value: "tcp"}, &Stringo{Value: "127.0.0.1"}, &Stringo{Value: "1"})
	errObj, ok := refused.(*Error)
	if !ok || !strings.Contains(errObj.Message, "`connect` error") {
		t.Errorf("connect to closed port should error, got %#v", refused)
	}

	tests := []builtinTestCase{
		{name: "transport not string", args: []Object{in(1), &Stringo{Value: host}, &Stringo{Value: port}}, err: "PositionalTypeError"},
		{name: "addr not string", args: []Object{&Stringo{Value: "tcp"}, in(1), &Stringo{Value: port}}, err: "PositionalTypeError"},
		{name: "port not string", args: []Object{&Stringo{Value: "tcp"}, &Stringo{Value: host}, in(1)}, err: "PositionalTypeError"},
		{name: "two args", args: []Object{&Stringo{Value: "tcp"}, &Stringo{Value: host}}, err: "InvalidArgCountError"},
	}
	runBuiltinTestsFor(t, NetBuiltins, "_connect", tests)
}

func TestListenAcceptReadWriteCloseTcp(t *testing.T) {
	connectFn := netBuiltinFn(t, "_connect")
	listenFn := netBuiltinFn(t, "_listen")
	acceptFn := netBuiltinFn(t, "_accept")
	readFn := netBuiltinFn(t, "_net_read")
	writeFn := netBuiltinFn(t, "_net_write")
	closeFn := netBuiltinFn(t, "_net_close")
	inspectFn := netBuiltinFn(t, "_inspect")

	lMap := listenFn(&Stringo{Value: "tcp"}, &Stringo{Value: "127.0.0.1"}, &Stringo{Value: "0"})
	listener := goObjValue[net.Listener](t, lMap)
	if ts := mapGetString(t, lMap.(*Map), "t").(*Stringo).Value; ts != "net/tcp" {
		t.Errorf("listen 't' field = %q, want 'net/tcp'", ts)
	}

	info := inspectFn(NewGoObj(net.Listener(listener)), &Stringo{Value: "net/tcp"}).(*Map)
	if addr := mapGetString(t, info, "addr").(*Stringo).Value; addr != listener.Addr().String() {
		t.Errorf("listener addr = %q, want %q", addr, listener.Addr().String())
	}
	if netw := mapGetString(t, info, "addr_network").(*Stringo).Value; netw != "tcp" {
		t.Errorf("listener network = %q, want tcp", netw)
	}

	host, port, _ := net.SplitHostPort(listener.Addr().String())
	dialCh := make(chan Object, 1)
	go func() {
		dialCh <- connectFn(&Stringo{Value: "tcp"}, &Stringo{Value: host}, &Stringo{Value: port})
	}()

	acceptRes := acceptFn(NewGoObj(net.Listener(listener)))
	clientRes := <-dialCh
	if _, isErr := clientRes.(*Error); isErr {
		t.Fatalf("background dial failed: %s", clientRes.Inspect())
	}
	serverConn := goObjValue[net.Conn](t, acceptRes)
	clientConn := goObjValue[net.Conn](t, clientRes)
	deadline := time.Now().Add(10 * time.Second)
	if err := serverConn.SetDeadline(deadline); err != nil {
		log.Printf("Failed to set server deadline, error: %s", err.Error())
	}
	if err := clientConn.SetDeadline(deadline); err != nil {
		log.Printf("Failed to set client deadline, error: %s", err.Error())
	}

	if res := writeFn(NewGoObj(clientConn), &Stringo{Value: "net"}, &Stringo{Value: "hello"}, &Stringo{Value: "\n"}); res != NULL {
		t.Errorf("net_write with end byte = %s, want NULL", res.Inspect())
	}
	rd := readFn(NewGoObj(serverConn), &Stringo{Value: "net"}, NULL, FALSE)
	if s, ok := rd.(*Stringo); !ok || s.Value != "hello" {
		t.Errorf("net_read(end byte) = %#v, want 'hello'", rd.Inspect())
	}

	if res := writeFn(NewGoObj(serverConn), &Stringo{Value: "net"}, &Bytes{Value: []byte("abcd")}); res != NULL {
		t.Errorf("3-arg net_write with BYTES = %s, want NULL", res.Inspect())
	}
	rdb := readFn(NewGoObj(clientConn), &Stringo{Value: "net"}, in(4), TRUE)
	if bs, ok := rdb.(*Bytes); !ok || string(bs.Value) != "abcd" {
		t.Errorf("net_read(len=4, as_bytes) = %#v, want 'abcd' bytes", rdb.Inspect())
	}

	writeFn(NewGoObj(clientConn), &Stringo{Value: "net"}, &Stringo{Value: "xyz"})
	mismatch := readFn(NewGoObj(serverConn), &Stringo{Value: "net"}, in(5), FALSE)
	errObj, ok := mismatch.(*Error)
	if !ok || !strings.Contains(errObj.Message, "does not match buffer length") {
		t.Errorf("length mismatch should error, got %#v", mismatch)
	}

	for _, tc := range []builtinTestCase{
		{name: "len zero", args: []Object{NewGoObj(serverConn), &Stringo{Value: "net"}, in(0), FALSE}, err: "must not be 0"},
		{name: "end byte too long", args: []Object{NewGoObj(serverConn), &Stringo{Value: "net"}, &Stringo{Value: "ab"}, FALSE}, err: "not length 1"},
	} {
		res := readFn(tc.args...)
		if _, isErr := res.(*Error); !isErr {
			t.Errorf("%s should error, got %v", tc.name, res.Inspect())
		}
	}
	runBuiltinTestsFor(t, NetBuiltins, "_net_read", []builtinTestCase{
		{name: "conn not goobj", args: []Object{in(1), &Stringo{Value: "net"}, NULL, FALSE}, err: "PositionalTypeError"},
		{name: "conn_t not string", args: []Object{NewGoObj(serverConn), in(1), NULL, FALSE}, err: "PositionalTypeError"},
		{name: "end byte wrong type", args: []Object{NewGoObj(serverConn), &Stringo{Value: "net"}, fl(1), FALSE}, err: "PositionalTypeError"},
		{name: "as_bytes wrong type", args: []Object{NewGoObj(serverConn), &Stringo{Value: "net"}, NULL, in(1)}, err: "PositionalTypeError"},
		{name: "three args", args: []Object{NewGoObj(serverConn), &Stringo{Value: "net"}, NULL}, err: "InvalidArgCountError"},
	})
	runBuiltinTestsFor(t, NetBuiltins, "_net_write", []builtinTestCase{
		{name: "value wrong type", args: []Object{NewGoObj(serverConn), &Stringo{Value: "net"}, TRUE}, err: "PositionalTypeError"},
		{name: "end byte wrong type", args: []Object{NewGoObj(serverConn), &Stringo{Value: "net"}, &Stringo{Value: "x"}, in(1)}, err: "PositionalTypeError"},
		{name: "five args", args: []Object{NewGoObj(serverConn), &Stringo{Value: "net"}, &Stringo{Value: "x"}, NULL, NULL}, err: "InvalidArgCountError"},
	})

	if res := closeFn(NewGoObj(clientConn), &Stringo{Value: "net"}); res != NULL {
		t.Errorf("net_close client = %s, want NULL", res.Inspect())
	}
	if res := closeFn(NewGoObj(serverConn), &Stringo{Value: "net"}); res != NULL {
		t.Errorf("net_close server = %s, want NULL", res.Inspect())
	}
	if res := closeFn(NewGoObj(net.Listener(listener)), &Stringo{Value: "net/tcp"}); res != NULL {
		t.Errorf("net_close listener = %s, want NULL", res.Inspect())
	}
	doubleClose := closeFn(NewGoObj(net.Listener(listener)), &Stringo{Value: "net/tcp"})
	if _, isErr := doubleClose.(*Error); !isErr {
		t.Errorf("double close should error, got %v", doubleClose.Inspect())
	}
	badType := closeFn(NewGoObj(clientConn), &Stringo{Value: "bogus"})
	if _, isErr := badType.(*Error); !isErr {
		t.Error("net_close unknown type should error")
	}
	badInspect := inspectFn(NewGoObj(clientConn), &Stringo{Value: "bogus"})
	if _, isErr := badInspect.(*Error); !isErr {
		t.Error("inspect unknown type should error")
	}
	acceptBad := acceptFn(NewGoObj(net.Listener(listener)))
	if _, isErr := acceptBad.(*Error); !isErr {
		t.Error("accept on closed listener should error")
	}
	runBuiltinTestsFor(t, NetBuiltins, "_accept", []builtinTestCase{
		{name: "not goobj", args: []Object{in(1)}, err: "PositionalTypeError"},
	})
	runBuiltinTestsFor(t, NetBuiltins, "_inspect", []builtinTestCase{
		{name: "not goobj", args: []Object{in(1), &Stringo{Value: "net"}}, err: "PositionalTypeError"},
		{name: "wrong inner type", args: []Object{NewGoObj("str"), &Stringo{Value: "net"}}, err: "PositionalTypeError"},
		{name: "one arg", args: []Object{NewGoObj("str")}, err: "InvalidArgCountError"},
	})
	runBuiltinTestsFor(t, NetBuiltins, "_net_close", []builtinTestCase{
		{name: "not goobj", args: []Object{in(1), &Stringo{Value: "net"}}, err: "PositionalTypeError"},
		{name: "type not string", args: []Object{NewGoObj("str"), in(1)}, err: "PositionalTypeError"},
		{name: "one arg", args: []Object{NewGoObj("str")}, err: "InvalidArgCountError"},
	})
}

func TestUdpListenReadWriteClose(t *testing.T) {
	listenFn := netBuiltinFn(t, "_listen")
	readFn := netBuiltinFn(t, "_net_read")
	writeFn := netBuiltinFn(t, "_net_write")
	closeFn := netBuiltinFn(t, "_net_close")
	inspectFn := netBuiltinFn(t, "_inspect")

	lMap := listenFn(&Stringo{Value: "udp"}, &Stringo{Value: "127.0.0.1"}, &Stringo{Value: "0"})
	serverUDP := goObjValue[*net.UDPConn](t, lMap)
	if ts := mapGetString(t, lMap.(*Map), "t").(*Stringo).Value; ts != "net/udp" {
		t.Errorf("udp listen 't' field = %q, want 'net/udp'", ts)
	}
	if err := serverUDP.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		log.Printf("Failed to set udp server deadline, error: %s", err.Error())
	}
	udpAddr := serverUDP.LocalAddr().(*net.UDPAddr)

	clientUDP, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if closeErr := clientUDP.Close(); closeErr != nil {
			log.Printf("Failed to close udp client, error: %s", closeErr.Error())
		}
	}()
	if err := clientUDP.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		log.Printf("Failed to set udp client deadline, error: %s", err.Error())
	}

	winfo := inspectFn(NewGoObj(clientUDP), &Stringo{Value: "net/udp"}).(*Map)
	ra := mapGetString(t, winfo, "remote_addr").(*Stringo).Value
	if !strings.HasSuffix(ra, ":"+strconv.Itoa(udpAddr.Port)) {
		t.Errorf("udp remote_addr = %q, want port %d", ra, udpAddr.Port)
	}

	if res := writeFn(NewGoObj(clientUDP), &Stringo{Value: "net/udp"}, &Stringo{Value: "hi"}, &Stringo{Value: "\n"}); res != NULL {
		t.Fatalf("udp net_write = %s, want NULL", res.Inspect())
	}
	rd := readFn(NewGoObj(serverUDP), &Stringo{Value: "net/udp"}, &Stringo{Value: "\n"}, FALSE)
	if s, ok := rd.(*Stringo); !ok || s.Value != "hi" {
		t.Errorf("udp net_read = %#v, want 'hi'", rd.Inspect())
	}

	if res := closeFn(NewGoObj(serverUDP), &Stringo{Value: "net/udp"}); res != NULL {
		t.Errorf("udp net_close = %s, want NULL", res.Inspect())
	}

	badPort := listenFn(&Stringo{Value: "tcp"}, &Stringo{Value: "127.0.0.1"}, &Stringo{Value: "not-a-port"})
	if _, isErr := badPort.(*Error); !isErr {
		t.Errorf("listen with bad port should error, got %v", badPort.Inspect())
	}
	runBuiltinTestsFor(t, NetBuiltins, "_listen", []builtinTestCase{
		{name: "transport not string", args: []Object{in(1), &Stringo{Value: "h"}, &Stringo{Value: "0"}}, err: "PositionalTypeError"},
		{name: "addr not string", args: []Object{&Stringo{Value: "tcp"}, in(1), &Stringo{Value: "0"}}, err: "PositionalTypeError"},
		{name: "port not string", args: []Object{&Stringo{Value: "tcp"}, &Stringo{Value: "h"}, in(1)}, err: "PositionalTypeError"},
		{name: "four args", args: []Object{&Stringo{Value: "tcp"}, &Stringo{Value: "h"}, &Stringo{Value: "0"}, NULL}, err: "InvalidArgCountError"},
	})
}
