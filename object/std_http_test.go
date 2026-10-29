package object

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/sse"
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
	vmImplemented := map[string]bool{"_handle": true, "_handle_use": true, "_handle_ws": true, "_handle_sse": true}
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
		"scheme":       "https",
		"username":     "user",
		"password":     "secret",
		"host":         "go.dev:8080",
		"path":         "/path/x",
		"raw_query":    "q=1",
		"fragment":     "frag",
		"raw_path":     "",
		"raw_fragment": "",
		"opaque":       "",
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
		if _, writeErr := w.Write([]byte(payload)); writeErr != nil {
			log.Printf("Failed to write test payload, error: %s", writeErr.Error())
		}
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

func newHttpTestApp(t *testing.T) *GoObj[*fiber.App] {
	t.Helper()
	res := httpBuiltinFn(t, "_new_server")(&Stringo{Value: "tcp"})
	app, ok := res.(*GoObj[*fiber.App])
	if !ok {
		t.Fatalf("new_server returned %T, want GoObj[*fiber.App]", res)
	}
	return app
}

func newHttpTestOptions(values map[string]Object) *Map {
	ordered := NewOrderedMap[string, Object]()
	for k, v := range values {
		ordered.Set(k, v)
	}
	return CreateMapObjectForGoMap(*ordered)
}

func TestHelmetMiddleware(t *testing.T) {
	helmetFn := httpBuiltinFn(t, "_helmet")
	app := newHttpTestApp(t)
	res := helmetFn(app, newHttpTestOptions(map[string]Object{
		"x_frame_options": &Stringo{Value: "DENY"},
		"hsts_max_age":    &Integer{Value: 31536000},
	}))
	if res != NULL {
		t.Fatalf("_helmet = %s, want NULL", res.Inspect())
	}
	app.Value.Get("/", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})
	resp, err := app.Value.Test(httptest.NewRequest(http.MethodGet, "https://example.com/", nil))
	if err != nil {
		t.Fatalf("request failed: %s", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Logf("failed to close response body: %s", closeErr.Error())
		}
	}()
	if got := resp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
	}
	if got := resp.Header.Get("X-Frame-Options"); got != "DENY" {
		t.Errorf("X-Frame-Options = %q, want DENY", got)
	}
	runBuiltinTestsFor(t, HttpBuiltins, "_helmet", []builtinTestCase{
		{name: "no args", args: []Object{}, err: "InvalidArgCountError"},
		{name: "app wrong goobj", args: []Object{NewGoObj("str"), newHttpTestOptions(nil)}, err: "PositionalTypeError"},
		{name: "options not map", args: []Object{app, &Stringo{Value: "x"}}, err: "PositionalTypeError"},
		{name: "unknown option", args: []Object{app, newHttpTestOptions(map[string]Object{"nope": TRUE})}, err: `unknown option "nope"`},
		{name: "bad option type", args: []Object{app, newHttpTestOptions(map[string]Object{"x_frame_options": TRUE})}, err: "expects STRING"},
		{name: "negative hsts", args: []Object{app, newHttpTestOptions(map[string]Object{"hsts_max_age": in(-1)})}, err: "hsts_max_age must be >= 0"},
		{name: "preload excludes subdomains", args: []Object{app, newHttpTestOptions(map[string]Object{
			"hsts_preload_enabled":    TRUE,
			"hsts_exclude_subdomains": TRUE,
		})}, err: "hsts_preload_enabled requires hsts_exclude_subdomains=false"},
	})
}

func TestCompressMiddleware(t *testing.T) {
	compressFn := httpBuiltinFn(t, "_compress")
	app := newHttpTestApp(t)
	if res := compressFn(app, in(0)); res != NULL {
		t.Fatalf("_compress = %s, want NULL", res.Inspect())
	}
	app.Value.Get("/", func(c fiber.Ctx) error {
		return c.SendString(strings.Repeat("blue ", 500))
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := app.Value.Test(req)
	if err != nil {
		t.Fatalf("request failed: %s", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Logf("failed to close response body: %s", closeErr.Error())
		}
	}()
	if got := resp.Header.Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", got)
	}
	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %s", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			t.Logf("failed to close gzip reader: %s", closeErr.Error())
		}
	}()
	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read compressed body: %s", err)
	}
	if len(body) != len("blue ")*500 {
		t.Errorf("decompressed body length = %d, want %d", len(body), len("blue ")*500)
	}

	runBuiltinTestsFor(t, HttpBuiltins, "_compress", []builtinTestCase{
		{name: "no args", args: []Object{}, err: "InvalidArgCountError"},
		{name: "app wrong goobj", args: []Object{NewGoObj("str"), in(0)}, err: "PositionalTypeError"},
		{name: "level not int", args: []Object{app, &Stringo{Value: "1"}}, err: "PositionalTypeError"},
		{name: "level too low", args: []Object{app, in(-2)}, err: "level must be between"},
		{name: "level too high", args: []Object{app, in(3)}, err: "level must be between"},
	})
}

func TestLimiterMiddleware(t *testing.T) {
	limiterFn := httpBuiltinFn(t, "_limiter")
	app := newHttpTestApp(t)
	if res := limiterFn(app, in(1), in(60), FALSE, FALSE, FALSE, FALSE); res != NULL {
		t.Fatalf("_limiter = %s, want NULL", res.Inspect())
	}
	app.Value.Get("/", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})
	for i, want := range []int{http.StatusOK, http.StatusTooManyRequests} {
		resp, err := app.Value.Test(httptest.NewRequest(http.MethodGet, "/", nil))
		if err != nil {
			t.Fatalf("request %d failed: %s", i, err)
		}
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Logf("failed to close response body: %s", closeErr.Error())
		}
		if resp.StatusCode != want {
			t.Errorf("request %d status = %d, want %d", i, resp.StatusCode, want)
		}
	}

	runBuiltinTestsFor(t, HttpBuiltins, "_limiter", []builtinTestCase{
		{name: "no args", args: []Object{}, err: "InvalidArgCountError"},
		{name: "app wrong goobj", args: []Object{NewGoObj("str"), in(1), in(60), FALSE, FALSE, FALSE, FALSE}, err: "PositionalTypeError"},
		{name: "max not int", args: []Object{app, &Stringo{Value: "1"}, in(60), FALSE, FALSE, FALSE, FALSE}, err: "PositionalTypeError"},
		{name: "expiration not int", args: []Object{app, in(1), &Stringo{Value: "60"}, FALSE, FALSE, FALSE, FALSE}, err: "PositionalTypeError"},
		{name: "skip failed not bool", args: []Object{app, in(1), in(60), in(1), FALSE, FALSE, FALSE}, err: "PositionalTypeError"},
		{name: "max too low", args: []Object{app, in(0), in(60), FALSE, FALSE, FALSE, FALSE}, err: "max must be > 0"},
		{name: "expiration too low", args: []Object{app, in(1), in(0), FALSE, FALSE, FALSE, FALSE}, err: "expiration must be > 0"},
	})
}

func TestEncryptCookieMiddleware(t *testing.T) {
	generateKeyFn := httpBuiltinFn(t, "_generate_cookie_key")
	encryptCookieFn := httpBuiltinFn(t, "_encrypt_cookie")
	keyRes := generateKeyFn(in(16))
	key, ok := keyRes.(*Stringo)
	if !ok {
		t.Fatalf("generate_cookie_key returned %T, want *Stringo", keyRes)
	}
	app := newHttpTestApp(t)
	if res := encryptCookieFn(app, key, &List{Elements: []Object{}}); res != NULL {
		t.Fatalf("_encrypt_cookie = %s, want NULL", res.Inspect())
	}
	app.Value.Get("/", func(c fiber.Ctx) error {
		if cookie := c.Cookies("session"); cookie != "" {
			return c.SendString(cookie)
		}
		c.Cookie(&fiber.Cookie{Name: "session", Value: "secret"})
		return c.SendString("set")
	})

	resp, err := app.Value.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("first request failed: %s", err)
	}
	if closeErr := resp.Body.Close(); closeErr != nil {
		t.Logf("failed to close first response body: %s", closeErr.Error())
	}
	cookies := resp.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("got %d response cookies, want 1", len(cookies))
	}
	if cookies[0].Value == "secret" {
		t.Error("response cookie was not encrypted")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookies[0])
	resp, err = app.Value.Test(req)
	if err != nil {
		t.Fatalf("second request failed: %s", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Logf("failed to close second response body: %s", closeErr.Error())
		}
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read second response body: %s", err)
	}
	if string(body) != "secret" {
		t.Errorf("decrypted request cookie = %q, want secret", string(body))
	}

	runBuiltinTestsFor(t, HttpBuiltins, "_generate_cookie_key", []builtinTestCase{
		{name: "no args", args: []Object{}, err: "InvalidArgCountError"},
		{name: "not an int", args: []Object{&Stringo{Value: "32"}}, err: "PositionalTypeError"},
		{name: "invalid length", args: []Object{in(8)}, err: "length must be 16, 24, or 32"},
	})
	runBuiltinTestsFor(t, HttpBuiltins, "_encrypt_cookie", []builtinTestCase{
		{name: "no args", args: []Object{}, err: "InvalidArgCountError"},
		{name: "app wrong goobj", args: []Object{NewGoObj("str"), key, &List{Elements: []Object{}}}, err: "PositionalTypeError"},
		{name: "key not string", args: []Object{app, in(1), &List{Elements: []Object{}}}, err: "PositionalTypeError"},
		{name: "invalid key encoding", args: []Object{app, &Stringo{Value: "%%%"}, &List{Elements: []Object{}}}, err: "failed to base64-decode key"},
		{name: "invalid key length", args: []Object{app, &Stringo{Value: "nope"}, &List{Elements: []Object{}}}, err: "key must decode to 16, 24, or 32 bytes"},
		{name: "except not list", args: []Object{app, key, &Stringo{Value: "x"}}, err: "PositionalTypeError"},
		{name: "except element not string", args: []Object{app, key, &List{Elements: []Object{in(1)}}}, err: "must be STRING"},
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

type fakeSSEStream struct {
	events []sse.Event
	lastID string
	err    error
}

func (f *fakeSSEStream) Event(event sse.Event) error {
	if f.err != nil {
		return f.err
	}
	f.events = append(f.events, event)
	return nil
}

func (f *fakeSSEStream) LastEventID() string {
	return f.lastID
}

func TestSendSSE(t *testing.T) {
	stream := &fakeSSEStream{}
	res := sendSSE(stream, "send", 1, &Stringo{Value: "line1\r\nline2"}, &Stringo{Value: "update"}, &Stringo{Value: "42"}, &Integer{Value: 3000})
	if res != NULL {
		t.Fatalf("sendSSE = %s, want NULL", res.Inspect())
	}
	if len(stream.events) != 1 {
		t.Fatalf("got %d events, want 1", len(stream.events))
	}
	got := stream.events[0]
	if string(got.Data.([]byte)) != "line1\r\nline2" || got.Name != "update" || got.ID != "42" || got.Retry != 3*time.Second {
		t.Errorf("sendSSE event = %#v, want data/name/id/retry to match", got)
	}

	negativeRetry := sendSSE(stream, "send", 1, &Stringo{Value: "x"}, NULL, NULL, &Integer{Value: -1})
	if _, ok := negativeRetry.(*Error); !ok {
		t.Errorf("sendSSE(negative retry) should error, got %v", negativeRetry.Inspect())
	}
}

func TestNewSSEConnection(t *testing.T) {
	conn := NewSSEConnection((*sse.Stream)(nil))
	tVal, ok := getMapValueByString(conn, "t")
	if !ok || tVal.Inspect() != "sse" {
		t.Fatalf("sse connection t = %v, want sse", tVal)
	}
	vVal, ok := getMapValueByString(conn, "v")
	if !ok {
		t.Fatal("sse connection is missing v")
	}
	if _, ok := vVal.(*GoObj[*sse.Stream]); !ok {
		t.Fatalf("sse connection v is %T, want *GoObj[*sse.Stream]", vVal)
	}
	lastVal, ok := getMapValueByString(conn, "last_event_id")
	if !ok {
		t.Fatal("sse connection is missing last_event_id")
	}
	if _, ok := lastVal.(*Builtin); !ok {
		t.Fatalf("sse connection last_event_id is %T, want *Builtin", lastVal)
	}
}

func TestSSELastEventID(t *testing.T) {
	runBuiltinTestsFor(t, HttpBuiltins, "_sse_last_event_id", []builtinTestCase{
		{name: "not a goobj", args: []Object{in(1)}, err: "PositionalTypeError"},
		{name: "wrong goobj", args: []Object{NewGoObj("not a stream")}, err: "PositionalTypeError"},
		{name: "no args", args: []Object{}, err: "InvalidArgCountError"},
	})
}

func TestSSESendBuiltin(t *testing.T) {
	writer := NewGoObj((*sse.Stream)(nil))
	fn := httpBuiltinFn(t, "_sse_send")

	conn := NewSSEConnection((*sse.Stream)(nil))
	mapWrongValue := fn(conn, in(1), NULL, NULL, NULL)
	if _, ok := mapWrongValue.(*Error); !ok {
		t.Errorf("sse_send(map, wrong value) should error, got %v", mapWrongValue.Inspect())
	}
	badWriter := fn(in(1), &Stringo{Value: "x"}, NULL, NULL, NULL)
	if _, ok := badWriter.(*Error); !ok {
		t.Errorf("sse_send(non-goobj) should error, got %v", badWriter.Inspect())
	}
	wrongWriter := fn(NewGoObj("not a stream"), &Stringo{Value: "x"}, NULL, NULL, NULL)
	if _, ok := wrongWriter.(*Error); !ok {
		t.Errorf("sse_send(wrong goobj) should error, got %v", wrongWriter.Inspect())
	}
	negativeRetry := fn(writer, &Stringo{Value: "x"}, NULL, NULL, &Integer{Value: -1})
	if _, ok := negativeRetry.(*Error); !ok {
		t.Errorf("sse_send(negative retry) should error, got %v", negativeRetry.Inspect())
	}

	tests := []builtinTestCase{
		{name: "not enough args", args: []Object{writer, &Stringo{Value: "x"}}, err: "InvalidArgCountError"},
		{name: "value wrong type", args: []Object{writer, in(1), NULL, NULL, NULL}, err: "PositionalTypeError"},
		{name: "event wrong type", args: []Object{writer, &Stringo{Value: "x"}, TRUE, NULL, NULL}, err: "PositionalTypeError"},
		{name: "id wrong type", args: []Object{writer, &Stringo{Value: "x"}, NULL, TRUE, NULL}, err: "PositionalTypeError"},
		{name: "retry wrong type", args: []Object{writer, &Stringo{Value: "x"}, NULL, NULL, &Stringo{Value: "1"}}, err: "PositionalTypeError"},
	}
	runBuiltinTestsFor(t, HttpBuiltins, "_sse_send", tests)
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
