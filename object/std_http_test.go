package object

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func httpBuiltinFn(t *testing.T, name string) BuiltinFunction {
	t.Helper()
	for _, b := range HttpBuiltins {
		if b.Name == name {
			if b.Fun == nil {
				t.Fatalf("http builtin %q has nil Fun", name)
			}
			return b.Fun
		}
	}
	t.Fatalf("http builtin %q not found", name)
	return nil
}

func TestHttpRegistry(t *testing.T) {
	vmImplemented := map[string]bool{"_handle": true, "_handle_use": true, "_handle_ws": true}
	seen := make(map[string]bool)
	for _, b := range HttpBuiltins {
		if b.Name == "" || b.HelpStr == "" {
			t.Fatalf("http builtin %q missing Name or HelpStr", b.Name)
		}
		if seen[b.Name] {
			t.Errorf("duplicate http builtin name %q", b.Name)
		}
		seen[b.Name] = true
		if b.Fun == nil && !vmImplemented[b.Name] {
			t.Errorf("http builtin %q has nil Fun but is not known to be VM-implemented", b.Name)
		}
	}
}

func TestUrlParseBuiltin(t *testing.T) {
	fn := httpBuiltinFn(t, "_url_parse")
	res := fn(&Stringo{Value: "https://user:secret@go.dev:8080/path/x?q=1#frag"})
	bs, ok := res.(*BlueStruct)
	if !ok {
		t.Fatalf("url_parse returned %T, want *BlueStruct", res)
	}
	want := map[string]string{
		"scheme":     "https",
		"username":   "user",
		"password":   "secret",
		"host":       "go.dev:8080",
		"path":       "/path/x",
		"raw_query":  "q=1",
		"fragment":   "frag",
		"raw_path":   "",
		"raw_fragment": "",
		"opaque":     "",
	}
	for field, expected := range want {
		v, idx := bs.Get(field)
		if idx == -1 {
			t.Errorf("url_parse result missing field %q", field)
			continue
		}
		s, ok := v.(*Stringo)
		if !ok || s.Value != expected {
			t.Errorf("url_parse %q = %#v, want %q", field, v.Inspect(), expected)
		}
	}
	if v, _ := bs.Get("force_query"); v != FALSE {
		t.Error("force_query should be false")
	}
	res2 := fn(&Stringo{Value: "https://go.dev"}).(*BlueStruct)
	if pw, _ := res2.Get("password"); pw != NULL {
		t.Errorf("password without credentials should be NULL, got %v", pw.Inspect())
	}

	tests := []builtinTestCase{
		{name: "invalid url", args: []Object{&Stringo{Value: "ht\x7ftp://["}}, err: "`url_parse` error"},
		{name: "not a string", args: []Object{in(1)}, err: "PositionalTypeError"},
		{name: "no args", args: []Object{}, err: "InvalidArgCountError"},
	}
	runBuiltinTestsFor(t, HttpBuiltins, "_url_parse", tests)
}

func TestUrlEncodeEscapeUnescapeBuiltins(t *testing.T) {
	runBuiltinTestsFor(t, HttpBuiltins, "_url_encode", []builtinTestCase{
		{name: "encodes path spaces", args: []Object{&Stringo{Value: "https://go.dev/a b"}}, want: "https://go.dev/a%20b"},
		{name: "already clean stays same", args: []Object{&Stringo{Value: "https://go.dev/x?a=b"}}, want: "https://go.dev/x?a=b"},
		{name: "not a string", args: []Object{in(1)}, err: "PositionalTypeError"},
		{name: "two args", args: []Object{&Stringo{Value: "a"}, &Stringo{Value: "b"}}, err: "InvalidArgCountError"},
	})
	runBuiltinTestsFor(t, HttpBuiltins, "_url_escape", []builtinTestCase{
		{name: "escape space", args: []Object{&Stringo{Value: "hello world"}}, want: "hello+world"},
		{name: "escape special", args: []Object{&Stringo{Value: "a&b=c"}}, want: "a%26b%3Dc"},
		{name: "not a string", args: []Object{fl(1)}, err: "PositionalTypeError"},
		{name: "no args", args: []Object{}, err: "InvalidArgCountError"},
	})
	runBuiltinTestsFor(t, HttpBuiltins, "_url_unescape", []builtinTestCase{
		{name: "unescape plus", args: []Object{&Stringo{Value: "hello+world"}}, want: "hello world"},
		{name: "unescape percent", args: []Object{&Stringo{Value: "a%26b%3Dc"}}, want: "a&b=c"},
		{name: "invalid escape", args: []Object{&Stringo{Value: "%zz"}}, err: "`url_unescape` error"},
		{name: "not a string", args: []Object{in(1)}, err: "PositionalTypeError"},
	})
}

func TestDownloadBuiltin(t *testing.T) {
	downloadFn := httpBuiltinFn(t, "_download")
	tmp := t.TempDir()

	payload := "blue download payload"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(payload))
	}))
	defer srv.Close()

	dest := filepath.Join(tmp, "out.txt")
	if res := downloadFn(&Stringo{Value: srv.URL + "/file.txt"}, &Stringo{Value: dest}); res != NULL {
		t.Fatalf("download = %s, want NULL", res.Inspect())
	}
	data, err := os.ReadFile(dest)
	if err != nil || string(data) != payload {
		t.Errorf("downloaded file content = %q (err=%v), want %q", data, err, payload)
	}

	srv.Close()
	unreachable := filepath.Join(tmp, "never.txt")
	res := downloadFn(&Stringo{Value: srv.URL + "/x.txt"}, &Stringo{Value: unreachable})
	errObj, ok := res.(*Error)
	if !ok || !strings.Contains(errObj.Message, "`download` error") {
		t.Errorf("download from dead server should error, got %#v", res)
	}

	tests := []builtinTestCase{
		{name: "empty url", args: []Object{&Stringo{Value: ""}, &Stringo{Value: dest}}, err: "argument 1 to `download` is ''"},
		{name: "url not string", args: []Object{in(1), &Stringo{Value: dest}}, err: "PositionalTypeError"},
		{name: "fpath not string", args: []Object{&Stringo{Value: srv.URL}, in(1)}, err: "PositionalTypeError"},
		{name: "one arg", args: []Object{&Stringo{Value: srv.URL}}, err: "InvalidArgCountError"},
	}
	runBuiltinTestsFor(t, HttpBuiltins, "_download", tests)
}

func TestNewServerStaticMonitor(t *testing.T) {
	newServerFn := httpBuiltinFn(t, "_new_server")
	staticFn := httpBuiltinFn(t, "_static")
	monitorFn := httpBuiltinFn(t, "_handle_monitor")

	appRes := newServerFn(&Stringo{Value: "tcp"})
	app, ok := appRes.(*GoObj[*fiber.App])
	if !ok {
		t.Fatalf("new_server returned %T, want GoObj[*fiber.App]", appRes)
	}

	tmp := t.TempDir()
	writeTempFile(t, tmp, "index.html", "<h1>blue</h1>")
	if res := staticFn(app, &Stringo{Value: "/"}, &Stringo{Value: tmp}, FALSE); res != NULL {
		t.Errorf("_static = %s, want NULL", res.Inspect())
	}
	if res := monitorFn(app, &Stringo{Value: "/monitor"}, TRUE); res != NULL {
		t.Errorf("_handle_monitor = %s, want NULL", res.Inspect())
	}

	shutdownFn := httpBuiltinFn(t, "_shutdown_server")
	if res := shutdownFn(); res != NULL {
		t.Errorf("_shutdown_server with no server running = %s, want NULL", res.Inspect())
	}

	runBuiltinTestsFor(t, HttpBuiltins, "_new_server", []builtinTestCase{
		{name: "not a string", args: []Object{in(1)}, err: "PositionalTypeError"},
		{name: "no args", args: []Object{}, err: "InvalidArgCountError"},
	})
	runBuiltinTestsFor(t, HttpBuiltins, "_static", []builtinTestCase{
		{name: "app wrong goobj", args: []Object{NewGoObj("str"), &Stringo{Value: "/"}, &Stringo{Value: tmp}, FALSE}, err: "PositionalTypeError"},
		{name: "prefix not string", args: []Object{app, in(1), &Stringo{Value: tmp}, FALSE}, err: "PositionalTypeError"},
		{name: "path not string", args: []Object{app, &Stringo{Value: "/"}, in(1), FALSE}, err: "PositionalTypeError"},
		{name: "browse not bool", args: []Object{app, &Stringo{Value: "/"}, &Stringo{Value: tmp}, in(0)}, err: "PositionalTypeError"},
		{name: "three args", args: []Object{app, &Stringo{Value: "/"}, &Stringo{Value: tmp}}, err: "InvalidArgCountError"},
	})
	runBuiltinTestsFor(t, HttpBuiltins, "_handle_monitor", []builtinTestCase{
		{name: "app wrong goobj", args: []Object{NewGoObj("str"), &Stringo{Value: "/m"}, TRUE}, err: "PositionalTypeError"},
		{name: "path not string", args: []Object{app, in(1), TRUE}, err: "PositionalTypeError"},
		{name: "flag not bool", args: []Object{app, &Stringo{Value: "/m"}, fl(1)}, err: "PositionalTypeError"},
		{name: "two args", args: []Object{app, &Stringo{Value: "/m"}}, err: "InvalidArgCountError"},
	})
	runBuiltinTestsFor(t, HttpBuiltins, "_serve", []builtinTestCase{
		{name: "app wrong goobj", args: []Object{NewGoObj("str"), &Stringo{Value: "127.0.0.1:0"}, FALSE}, err: "PositionalTypeError"},
		{name: "addr not string", args: []Object{app, in(1), FALSE}, err: "PositionalTypeError"},
		{name: "embed flag not bool", args: []Object{app, &Stringo{Value: "127.0.0.1:0"}, in(1)}, err: "PositionalTypeError"},
	})
}

func TestMarkdownHtmlConversions(t *testing.T) {
	mdFn := httpBuiltinFn(t, "_md_to_html")
	htmlOut := mdFn(&Stringo{Value: "# Hello World"})
	if s, ok := htmlOut.(*Stringo); !ok || s.Value != "<h1>Hello World</h1>\n" {
		t.Errorf("md_to_html = %#v, want '<h1>Hello World</h1>\\n'", htmlOut.Inspect())
	}
	runBuiltinTestsFor(t, HttpBuiltins, "_md_to_html", []builtinTestCase{
		{name: "not a string", args: []Object{in(1)}, err: "PositionalTypeError"},
		{name: "two args", args: []Object{&Stringo{Value: "a"}, &Stringo{Value: "b"}}, err: "InvalidArgCountError"},
	})

	toMdFn := httpBuiltinFn(t, "_html_to_md")
	mdOut := toMdFn(&Stringo{Value: "<h1>Hello World</h1>"}, &Stringo{Value: ""})
	if s, ok := mdOut.(*Stringo); !ok || s.Value != "# Hello World" {
		t.Errorf("html_to_md = %#v, want '# Hello World'", mdOut.Inspect())
	}
	linkOut := toMdFn(&Stringo{Value: `<a href="/docs">docs</a>`}, &Stringo{Value: "https://go.dev"})
	if ls, ok := linkOut.(*Stringo); !ok || !strings.Contains(ls.Value, "https://go.dev/docs") {
		t.Errorf("html_to_md with domain should absolutize links, got %#v", linkOut.Inspect())
	}
	runBuiltinTestsFor(t, HttpBuiltins, "_html_to_md", []builtinTestCase{
		{name: "html not string", args: []Object{in(1), &Stringo{Value: ""}}, err: "PositionalTypeError"},
		{name: "domain not string", args: []Object{&Stringo{Value: "<p/>"}, in(1)}, err: "PositionalTypeError"},
		{name: "one arg", args: []Object{&Stringo{Value: "<p/>"}}, err: "InvalidArgCountError"},
	})
}

func TestSanitizeAndMinifyBuiltin(t *testing.T) {
	fn := httpBuiltinFn(t, "_sanitize_and_minify")

	dirty := "<p>hi<script>alert(1)</script></p>"
	clean := fn(&Stringo{Value: dirty}, TRUE, FALSE).(*Stringo).Value
	if strings.Contains(clean, "script") {
		t.Errorf("sanitize did not strip script tag: %q", clean)
	}
	if !strings.Contains(clean, "hi") {
		t.Errorf("sanitize stripped safe content: %q", clean)
	}

	spaced := "<div>\n  <p>x</p>\n</div>"
	minified := fn(&Stringo{Value: spaced}, FALSE, TRUE).(*Stringo).Value
	if len(minified) >= len(spaced) {
		t.Errorf("minify did not shrink output: %q", minified)
	}

	passthrough := fn(&Stringo{Value: spaced}, FALSE, FALSE).(*Stringo).Value
	if passthrough != spaced {
		t.Errorf("sanitize=false,minify=false should pass through unchanged, got %q", passthrough)
	}

	tests := []builtinTestCase{
		{name: "content not string", args: []Object{in(1), TRUE, TRUE}, err: "PositionalTypeError"},
		{name: "sanitize flag not bool", args: []Object{&Stringo{Value: "x"}, in(1), TRUE}, err: "PositionalTypeError"},
		{name: "minify flag not bool", args: []Object{&Stringo{Value: "x"}, TRUE, fl(1)}, err: "PositionalTypeError"},
		{name: "two args", args: []Object{&Stringo{Value: "x"}, TRUE}, err: "InvalidArgCountError"},
	}
	runBuiltinTestsFor(t, HttpBuiltins, "_sanitize_and_minify", tests)
}

func TestWsAndMiscTypeErrors(t *testing.T) {
	newServerApp := httpBuiltinFn(t, "_new_server")(&Stringo{Value: "tcp"}).(*GoObj[*fiber.App])

	for _, name := range []string{"_ws_send", "_ws_recv", "_ws_client_send", "_ws_client_recv"} {
		fn := httpBuiltinFn(t, name)
		res := fn(in(1))
		if _, isErr := res.(*Error); !isErr {
			t.Errorf("%s(non-goobj) should error, got %v", name, res.Inspect())
		}
		res2 := fn(NewGoObj(newServerApp.Value))
		if _, isErr := res2.(*Error); !isErr {
			t.Errorf("%s(wrong inner type) should error, got %v", name, res2.Inspect())
		}
	}
	runBuiltinTestsFor(t, HttpBuiltins, "_ws_send", []builtinTestCase{
		{name: "value wrong type", args: []Object{NewGoObj((*fiber.App)(nil)), TRUE}, err: "PositionalTypeError"},
	})
	runBuiltinTestsFor(t, HttpBuiltins, "_ws_client_send", []builtinTestCase{
		{name: "value wrong type", args: []Object{NewGoObj((*fiber.App)(nil)), in(5)}, err: "PositionalTypeError"},
	})
	badDial := httpBuiltinFn(t, "_new_ws")(&Stringo{Value: "http://127.0.0.1:1/ws"})
	if _, isErr := badDial.(*Error); !isErr {
		t.Errorf("new_ws to dead endpoint should error, got %v", badDial.Inspect())
	}
	runBuiltinTestsFor(t, HttpBuiltins, "_new_ws", []builtinTestCase{
		{name: "not a string", args: []Object{in(1)}, err: "PositionalTypeError"},
	})
	runBuiltinTestsFor(t, HttpBuiltins, "_inspect", []builtinTestCase{
		{name: "unknown type", args: []Object{NewGoObj((*fiber.App)(nil)), &Stringo{Value: "nope"}}, err: "expects type of 'ws'|'ws/client'"},
		{name: "not goobj", args: []Object{in(1), &Stringo{Value: "ws"}}, err: "PositionalTypeError"},
	})
	runBuiltinTestsFor(t, HttpBuiltins, "_shutdown_server", []builtinTestCase{
		{name: "takes no args", args: []Object{NULL}, err: "InvalidArgCountError"},
	})
	runBuiltinTestsFor(t, HttpBuiltins, "_open_browser", []builtinTestCase{
		{name: "not a string", args: []Object{in(1)}, err: "PositionalTypeError"},
		{name: "no args", args: []Object{}, err: "InvalidArgCountError"},
	})
}
