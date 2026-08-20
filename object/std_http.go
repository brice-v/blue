package object

import (
	"blue/consts"
	"blue/lib"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"strings"

	ws "github.com/gorilla/websocket"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/strikethrough"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
)

// Used to catch interrupt to shutdown server
var interruptCh = make(chan os.Signal, 1)

// createMonitorHandler returns a handler serving the self contained monitor
// dashboard. When shouldShow is false the handler falls through to the next
// handler so the dashboard is hidden.
func createMonitorHandler(path string, shouldShow bool) func(*Ctx) {
	dataPath := strings.TrimSuffix(path, "/") + "/data"
	return func(c *Ctx) {
		if !shouldShow {
			_ = c.Next()
			return
		}
		html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>Blue HTTP Monitor</title>
<style>
body { font-family: sans-serif; margin: 2rem; }
pre { background: #f5f5f5; padding: 1rem; border-radius: 4px; overflow-x: auto; }
h1 { font-size: 1.4rem; }
</style>
</head>
<body>
<h1>Blue HTTP Monitor</h1>
<pre id="stats">loading...</pre>
<script>
async function refresh() {
  try {
    const r = await fetch('%s');
    const data = await r.json();
    document.getElementById('stats').textContent = JSON.stringify(data, null, 2);
  } catch (e) {
    document.getElementById('stats').textContent = 'error fetching stats: ' + e;
  }
}
refresh();
setInterval(refresh, 1000);
</script>
</body>
</html>`, dataPath)
		c.Set("Content-Type", "text/html; charset=utf-8")
		_ = c.SendString(html)
	}
}

// collectMonitorStats gathers runtime plus gopsutil cpu and memory stats for
// the monitor json endpoint.
func collectMonitorStats() map[string]any {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	stats := map[string]any{
		"goroutines": runtime.NumGoroutine(),
		"mem_alloc":  memStats.Alloc,
		"mem_total":  memStats.TotalAlloc,
		"num_gc":     memStats.NumGC,
	}
	cpuPercents, err := cpu.Percent(0, false)
	if err == nil && len(cpuPercents) > 0 {
		stats["cpu_percent"] = cpuPercents[0]
	}
	if vm, err := mem.VirtualMemory(); err == nil {
		stats["mem_used"] = vm.Used
		stats["mem_total_bytes"] = vm.Total
		stats["mem_used_percent"] = vm.UsedPercent
	}
	return stats
}

var HttpBuiltins = []*Builtin{
	{
		Name: "_url_parse",
		Fun: func(args ...Object) Object {
			e := checkArgCount("url_parse", 1, args)
			if e != nil {
				return e
			}
			e = checkArgType("url_parse", 1, STRING_OBJ, args)
			if e != nil {
				return e
			}
			u, err := url.Parse(args[0].(*Stringo).Value)
			if err != nil {
				return newError("`url_parse` error: %s", err.Error())
			}
			var passwordField Object
			if pw, ok := u.User.Password(); ok {
				passwordField = &Stringo{Value: pw}
			} else {
				passwordField = NULL
			}
			fields := []string{
				"scheme",
				"opaque",
				"username",
				"password",
				"host",
				"path",
				"fragment",
				"raw_query",
				"raw_path",
				"raw_fragment",
				"force_query",
				"omit_host",
			}
			values := []Object{&Stringo{Value: u.Scheme},
				&Stringo{Value: u.Opaque},
				&Stringo{Value: u.User.Username()},
				passwordField,
				&Stringo{Value: u.Host},
				&Stringo{Value: u.Path},
				&Stringo{Value: u.Fragment},
				&Stringo{Value: u.RawQuery},
				&Stringo{Value: u.RawPath},
				&Stringo{Value: u.RawFragment},
				nativeToBooleanObject(u.ForceQuery),
				nativeToBooleanObject(u.OmitHost),
			}
			bs, err := NewBlueStruct(fields, values)
			if err != nil {
				return newError("`url_parse` error: %s", err.Error())
			}
			return bs
		},
		HelpStr: helpStrArgs{
			explanation: "`url_parse` returns the url as it was parsed with different components in a blue struct",
			signature:   "url_parse(arg: str) -> struct",
			errors:      "InvalidArgCount,PositionalType,Custom",
			example:     "url_parse('https://go.dev') => @{scheme: 'https', opaque: '', username: '', password: null, host: 'go.dev', path: '', fragment: '', raw_query: '', raw_path: '', raw_fragment: '', force_query: false, omit_host: false}",
		}.String(),
	},
	{
		Name: "_url_encode",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("url_encode", len(args), 1, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("url_encode", 1, STRING_OBJ, args[0].Type())
			}
			s := args[0].(*Stringo).Value
			u, err := url.Parse(s)
			if err != nil {
				return newError("`url_encode` error: %s", err.Error())
			}
			return &Stringo{Value: u.String()}
		},
		HelpStr: helpStrArgs{
			explanation: "`url_encode` returns the STRING encoded as a valid URL",
			signature:   "url_encode(arg: str) -> str",
			errors:      "InvalidArgCount,PositionalType",
			example:     "url_encode('hello world') => 'hello%20world'",
		}.String(),
	},
	{
		Name: "_url_escape",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("url_escape", len(args), 1, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("url_escape", 1, STRING_OBJ, args[0].Type())
			}
			s := args[0].(*Stringo).Value
			return &Stringo{Value: url.QueryEscape(s)}
		},
		HelpStr: helpStrArgs{
			explanation: "`url_escape` returns the STRING encoded as a valid value to be passed through a URL",
			signature:   "url_escape(arg: str) -> str",
			errors:      "InvalidArgCount,PositionalType",
			example:     "url_escape('hello world') => 'hello+world'",
		}.String(),
	},
	{
		Name: "_url_unescape",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("url_unescape", len(args), 1, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("url_unescape", 1, STRING_OBJ, args[1].Type())
			}
			s := args[0].(*Stringo).Value
			urlUnescaped, err := url.QueryUnescape(s)
			if err != nil {
				return newError("`url_unescape` error: %s", err.Error())
			}
			return &Stringo{Value: urlUnescaped}
		},
		HelpStr: helpStrArgs{
			explanation: "`url_unescape` returns the STRING encoded as a valid value to be passed through a URL",
			signature:   "url_unescape(arg: str) -> str",
			errors:      "InvalidArgCount,PositionalType",
			example:     "url_unescape('hello+world') => 'hello world'",
		}.String(),
	},
	{
		Name: "_download",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("download", len(args), 2, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("download", 1, STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("download", 2, STRING_OBJ, args[1].Type())
			}
			urlS := args[0].(*Stringo).Value
			fname := args[1].(*Stringo).Value
			if urlS == "" {
				return newError("argument 1 to `download` is ''")
			}
			if fname == "" {
				// Build fileName from fullPath
				fileURL, err := url.Parse(urlS)
				if err != nil {
					return newError("`download` error: %s", err.Error())
				}
				path := fileURL.Path
				segments := strings.Split(path, "/")
				fname = segments[len(segments)-1]
			}
			resp, err := http.Get(urlS)
			if err != nil {
				return newError("`download` error: %s", err.Error())
			}
			defer func() {
				err := resp.Body.Close()
				if err != nil {
					log.Printf("Failed to close response body, error: %s", err.Error())
				}
			}()
			f, err := os.Create(fname)
			if err != nil {
				return newError("`download` error: %s", err.Error())
			}
			defer func() {
				err := f.Close()
				if err != nil {
					log.Printf("Failed to close file %s, error: %s", fname, err.Error())
				}
			}()

			_, err = io.Copy(f, resp.Body)
			if err != nil {
				log.Printf("Failed to copy response body to file, error: %s", err.Error())
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`download` copys the file at the URL to the given file path. if the fpath is empty, then the URL is used to determine the name",
			signature:   "download(url: str, fpath: str='') -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "download('http://example.com/test.txt') => null => writes test.txt to current directory",
		}.String(),
	},
	{
		Name: "_new_server",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("new_server", len(args), 1, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("new_server", 1, STRING_OBJ, args[0].Type())
			}
			network := args[0].(*Stringo).Value
			app := NewServer()
			app.Network = network
			return NewGoObj(app)
		},
		HelpStr: helpStrArgs{
			explanation: "`new_server` returns a new server object",
			signature:   "new_server(network: str('tcp','tcp4','tcp6')='tcp4') -> GoObj[*Server]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "new_server() => server obj",
		}.String(),
	},
	{
		Name: "_serve",
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("serve", len(args), 3, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("serve", 1, GO_OBJ, args[0].Type())
			}
			app, ok := args[0].(*GoObj[*Server])
			if !ok {
				return newPositionalTypeErrorForGoObj("serve", 1, "*Server", args[0])
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("serve", 2, STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("seve", 3, BOOLEAN_OBJ, args[2].Type())
			}
			useEmbeddedLibWeb := args[2].(*Boolean).Value
			addrPort := args[1].(*Stringo).Value
			signal.Notify(interruptCh, os.Interrupt)
			go func() {
				<-interruptCh
				fmt.Println("Interupt... Shutting down http server")
				_ = app.Value.Shutdown()
			}()
			if useEmbeddedLibWeb {
				sub, err := fs.Sub(lib.WebEmbedFiles, "web")
				if err != nil {
					return newError("`serve` error: %s", err.Error())
				}
				addStaticFiles(app.Value, "/", http.FS(sub), false)
			}
			ln, err := net.Listen(app.Value.Network, addrPort)
			if err != nil {
				return newError("`serve` error: %s", err.Error())
			}
			err = app.Value.Serve(ln)
			if err != nil && err != http.ErrServerClosed {
				return newError("`serve` error: %s", err.Error())
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`serve` starts the http server listener at the given address/port with the embedded lib web files included if set to true",
			signature:   "serve(server: GoObj[*Server], addr_port: str='localhost:3001', use_embedded_lib_web: bool=true) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "serve() => null => starts server",
		}.String(),
	},
	{
		Name: "_shutdown_server",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("shutdown_server", len(args), 0, "")
			}
			interruptCh <- os.Interrupt
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`shutdown_server` shuts down the given http server cleanly. it does not need to happen in the same process",
			signature:   "shutdown_server() -> null",
			errors:      "InvalidArgCount",
			example:     "shutdown_server() => null",
		}.String(),
	},
	{
		Name: "_static",
		Fun: func(args ...Object) Object {
			if len(args) != 4 {
				return newInvalidArgCountError("static", len(args), 4, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("static", 1, GO_OBJ, args[0].Type())
			}
			app, ok := args[0].(*GoObj[*Server])
			if !ok {
				return newPositionalTypeErrorForGoObj("static", 1, "*Server", args[0])
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("static", 2, STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != STRING_OBJ {
				return newPositionalTypeError("static", 3, STRING_OBJ, args[2].Type())
			}
			if args[3].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("static", 4, BOOLEAN_OBJ, args[3].Type())
			}
			prefix := args[1].(*Stringo).Value
			fpath := args[2].(*Stringo).Value
			shouldBrowse := args[3].(*Boolean).Value
			if IsEmbed {
				if strings.HasPrefix(fpath, "./") {
					fpath = strings.TrimLeft(fpath, "./")
				}
				sub, err := fs.Sub(Files, consts.EMBED_FILES_PREFIX+fpath)
				if err != nil {
					return newError("`static` error: %s", err.Error())
				}
				addStaticFiles(app.Value, prefix, http.FS(sub), shouldBrowse)
			} else {
				addStaticFiles(app.Value, prefix, http.Dir(fpath), shouldBrowse)
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`static` serves the given directory as static files for the http server",
			signature:   "static(server: GoObj[*Server], prefix: str='/', dir_path: str='.', browse: bool=false) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "static() => null => current directory served at addr:port/",
		}.String(),
	},
	{
		Name: "_ws_send",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("ws_send", len(args), 2, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("ws_send", 1, GO_OBJ, args[0].Type())
			}
			c, ok := args[0].(*GoObj[*ws.Conn])
			if !ok {
				return newPositionalTypeErrorForGoObj("ws_send", 1, "*ws.Conn", args[0])
			}
			if args[1].Type() != STRING_OBJ && args[1].Type() != BYTES_OBJ {
				return newPositionalTypeError("ws_send", 2, "STRING or BYTES", args[1].Type())
			}
			var err error
			if args[1].Type() == STRING_OBJ {
				err = c.Value.WriteMessage(ws.TextMessage, []byte(args[1].(*Stringo).Value))
			} else {
				err = c.Value.WriteMessage(ws.BinaryMessage, args[1].(*Bytes).Value)
			}
			if err != nil {
				return newError("`ws_send` error: %s", err.Error())
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`ws_send` sends the given value on the websocket connection, if the value is a string the websocket message type is TextMessage, otherwise if bytes BinaryMessage",
			signature:   "ws_send(c: GoObj[*ws.Conn], value: str|bytes) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "ws_send(c, '1') => null",
		}.String(),
	},
	{
		Name: "_ws_recv",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("ws_recv", len(args), 1, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("ws_recv", 1, GO_OBJ, args[0].Type())
			}
			c, ok := args[0].(*GoObj[*ws.Conn])
			if !ok {
				return newPositionalTypeErrorForGoObj("ws_send", 1, "*ws.Conn", args[0])
			}
			mt, msg, err := c.Value.ReadMessage()
			if err != nil {
				// If its closed we still want to return an error so that the handler fn wont try to send NULL
				return newError("`ws_recv` error: %s", err.Error())
			}
			switch mt {
			case ws.BinaryMessage:
				return &Bytes{Value: msg}
			case ws.TextMessage:
				return &Stringo{Value: string(msg)}
			case ws.PingMessage:
				return newError("`ws_recv` error: ping message type not supported.")
			case ws.PongMessage:
				return newError("`ws_recv` error: pong message type not supported.")
			default:
				// If its closed we still want to return an error so that the handler fn wont try to send NULL
				return newError("`ws_recv` error: websocket closed.")
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`ws_recv` receives a websocket message on the given websocket connection",
			signature:   "ws_recv(c: GoObj[*ws.Conn]) -> str|bytes",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "ws_recv(c) => str|bytes",
		}.String(),
	},
	{
		Name: "_new_ws",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("new_ws", len(args), 1, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("new_ws", 1, STRING_OBJ, args[0].Type())
			}
			url := args[0].(*Stringo).Value
			conn, _, err := ws.DefaultDialer.Dial(url, nil)
			if err != nil {
				return newError("`new_ws` error: %s", err.Error())
			}
			return CreateBasicMapObjectForGoObj("ws/client", NewGoObj(conn))
		},
		HelpStr: helpStrArgs{
			explanation: "`new_ws` returns a new websocket client object",
			signature:   "new_ws(url: str) -> {t: 'ws/client', v: GoObj[*ws.Conn]}",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "new_ws('http://localhost:3001/ws') => ws client obj",
		}.String(),
	},
	{
		Name: "_ws_client_send",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("ws_client_send", len(args), 2, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("ws_client_send", 1, GO_OBJ, args[0].Type())
			}
			c, ok := args[0].(*GoObj[*ws.Conn])
			if !ok {
				return newPositionalTypeErrorForGoObj("ws_client_send", 1, "*ws.Conn", args[0])
			}
			if args[1].Type() != STRING_OBJ && args[1].Type() != BYTES_OBJ {
				return newPositionalTypeError("ws_client_send", 2, "STRING or BYTES", args[1].Type())
			}
			var err error
			if args[1].Type() == STRING_OBJ {
				err = c.Value.WriteMessage(ws.TextMessage, []byte(args[1].(*Stringo).Value))
			} else {
				err = c.Value.WriteMessage(ws.BinaryMessage, args[1].(*Bytes).Value)
			}
			if err != nil {
				return newError("`ws_send` error: %s", err.Error())
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`ws_client_send` sends the given value on the websocket client connection, if the value is a string the websocket message type is TextMessage, otherwise if bytes BinaryMessage",
			signature:   "ws_client_send(c: GoObj[*ws.Conn], value: str|bytes) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "ws_client_send(c, '1') => null",
		}.String(),
	},
	{
		Name: "_ws_client_recv",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("ws_client_recv", len(args), 1, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("ws_client_recv", 1, GO_OBJ, args[0].Type())
			}
			c, ok := args[0].(*GoObj[*ws.Conn])
			if !ok {
				return newPositionalTypeErrorForGoObj("ws_client_recv", 1, "*ws.Conn", args[0])
			}
			mt, msg, err := c.Value.ReadMessage()
			if err != nil {
				// If its closed we still want to return an error so that the handler fn wont try to send NULL
				return newError("`ws_client_recv` error: %s", err.Error())
			}
			switch mt {
			case ws.BinaryMessage:
				return &Bytes{Value: msg}
			case ws.TextMessage:
				return &Stringo{Value: string(msg)}
			case ws.PingMessage:
				return newError("`ws_client_recv` error: ping message type not supported.")
			case ws.PongMessage:
				return newError("`ws_client_recv` error: pong message type not supported.")
			default:
				// If its closed we still want to return an error so that the handler fn wont try to send NULL
				return newError("`ws_client_recv` error: websocket closed.")
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`ws_client_recv` receives a value on the websocket client connection",
			signature:   "ws_client_recv(c: GoObj[*ws.Conn]) -> str|bytes",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "ws_client_recv(c) => str|bytes",
		}.String(),
	},
	{
		Name: "_handle_monitor",
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("handle_monitor", len(args), 3, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("handle_monitor", 1, GO_OBJ, args[0].Type())
			}
			app, ok := args[0].(*GoObj[*Server])
			if !ok {
				return newPositionalTypeErrorForGoObj("handle_monitor", 1, "*Server", args[0])
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("handle_monitor", 2, STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("handle_monitor", 3, BOOLEAN_OBJ, args[2].Type())
			}
			path := args[1].(*Stringo).Value
			shouldShow := args[2].(*Boolean).Value
			monitor := createMonitorHandler(path, shouldShow)
			app.Value.Add("GET", path, func(c *Ctx) {
				monitor(c)
			}, false)
			app.Value.Add("GET", strings.TrimSuffix(path, "/")+"/data", func(c *Ctx) {
				if !shouldShow {
					_ = c.Next()
					return
				}
				c.JSON(collectMonitorStats())
			}, false)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`handle_monitor` creates a monitor handler on the given http server at the given path a boolean that determines when it should show",
			signature:   "handle_monitor(s: GoObj[*Server], path: str, should_show: bool) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "handle_monitor(s, '/monitor', true) => null",
		}.String(),
	},
	{
		Name: "_md_to_html",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("md_to_html", len(args), 1, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("md_to_html", 1, STRING_OBJ, args[0].Type())
			}
			bs := []byte(args[0].(*Stringo).Value)
			output := blackfriday.Run(bs)
			return &Stringo{Value: string(output)}
		},
		HelpStr: helpStrArgs{
			explanation: "`md_to_html` converts a given markdown string to valid html",
			signature:   "md_to_html(s: str) -> str",
			errors:      "InvalidArgCount,PositionalType",
			example:     "md_to_html('# Hello World') => '<h1>Hello World</h1>'",
		}.String(),
	},
	{
		Name: "_html_to_md",
		Fun: func(args ...Object) Object {
			e := checkArgCount("html_to_md", 2, args)
			if e != nil {
				return e
			}
			e = checkArgType("html_to_md", 1, STRING_OBJ, args)
			if e != nil {
				return e
			}
			e = checkArgType("html_to_md", 2, STRING_OBJ, args)
			if e != nil {
				return e
			}
			htmlString := args[0].(*Stringo).Value
			domain := args[1].(*Stringo).Value
			conv := converter.NewConverter(
				converter.WithPlugins(
					base.NewBasePlugin(),
					commonmark.NewCommonmarkPlugin(),
					table.NewTablePlugin(),
					strikethrough.NewStrikethroughPlugin(),
				),
			)
			var markdown string
			if domain == "" {
				m, err := conv.ConvertString(htmlString)
				if err != nil {
					return newError("`html_to_md` error: %s", err.Error())
				}
				markdown = m
			} else {
				m, err := conv.ConvertString(htmlString, converter.WithDomain(domain))
				if err != nil {
					return newError("`html_to_md` error: %s", err.Error())
				}
				markdown = m
			}
			return &Stringo{Value: markdown}
		},
		HelpStr: helpStrArgs{
			explanation: "`html_to_md` converts html string to markdown. If domain is populate, links will use absolute link with domain name.",
			signature:   "html_to_md(s: str, domain='') -> str",
			errors:      "InvalidArgCount,PositionalType,Custom",
			example:     "html_to_md('<h1>Hello World</h1>') => '# Hello World'",
		}.String(),
	},
	{
		Name: "_sanitize_and_minify",
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("sanitize_and_minify", len(args), 3, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("sanitize_and_minify", 1, STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("sanitize_and_minify", 2, BOOLEAN_OBJ, args[1].Type())
			}
			if args[2].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("sanitize_and_minify", 3, BOOLEAN_OBJ, args[2].Type())
			}
			bs := []byte(args[0].(*Stringo).Value)
			shouldSanitize := args[1].(*Boolean).Value
			shouldMinify := args[2].(*Boolean).Value
			htmlContent := bs
			if shouldSanitize {
				p := bluemonday.UGCPolicy()
				// allow code to still get syntax highlighting
				p.AllowAttrs("class").Matching(regexp.MustCompile("^language-[a-zA-Z0-9]+$")).OnElements("code")
				htmlContent = p.SanitizeBytes(htmlContent)
			}
			if shouldMinify {
				m := minify.New()
				m.Add("text/html", &html.Minifier{
					KeepWhitespace:   true,
					KeepDocumentTags: true,
				})
				htmlContent1, err := m.Bytes("text/html", htmlContent)
				if err != nil {
					return newError("`sanitize_and_minify` error: %s", err.Error())
				}
				htmlContent = htmlContent1
			}
			return &Stringo{Value: string(htmlContent)}
		},
		HelpStr: helpStrArgs{
			explanation: "`sanitize_and_minify` santizes and/or minifies the given content",
			signature:   "sanitize_and_minify(content: str, should_sanitize: bool=true, should_minify: bool=true) -> str",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "sanitize_and_minify('<script></script>') => ''",
		}.String(),
	},
	{
		Name: "_inspect",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("inspect", len(args), 2, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("inspect", 1, GO_OBJ, args[0].Type())
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("inspect", 2, STRING_OBJ, args[1].Type())
			}
			t := args[1].(*Stringo).Value
			switch t {
			case "ws":
				c, ok := args[0].(*GoObj[*ws.Conn])
				if !ok {
					return newPositionalTypeErrorForGoObj("inspect", 1, "*ws.Conn", args[0])
				}
				mapObj := NewOrderedMap[string, Object]()
				mapObj.Set("remote_addr", &Stringo{Value: c.Value.RemoteAddr().String()})
				mapObj.Set("local_addr", &Stringo{Value: c.Value.LocalAddr().String()})
				mapObj.Set("remote_addr_network", &Stringo{Value: c.Value.RemoteAddr().Network()})
				mapObj.Set("local_addr_network", &Stringo{Value: c.Value.LocalAddr().Network()})
				return CreateMapObjectForGoMap(*mapObj)
			case "ws/client":
				c, ok := args[0].(*GoObj[*ws.Conn])
				if !ok {
					return newPositionalTypeErrorForGoObj("inspect", 1, "*ws.Conn", args[0])
				}
				mapObj := NewOrderedMap[string, Object]()
				mapObj.Set("remote_addr", &Stringo{Value: c.Value.RemoteAddr().String()})
				mapObj.Set("local_addr", &Stringo{Value: c.Value.LocalAddr().String()})
				mapObj.Set("remote_addr_network", &Stringo{Value: c.Value.RemoteAddr().Network()})
				mapObj.Set("local_addr_network", &Stringo{Value: c.Value.LocalAddr().Network()})
				return CreateMapObjectForGoMap(*mapObj)
			default:
				return newError("`inspect` error: expects type of 'ws'|'ws/client'")
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`inspect` will return a map of info for the given ws connection",
			signature:   "inspect(c: GoObj[*ws.Conn]|GoObj[*ws.Connection]) -> map[str]str",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "inspect(c) => {remote_addr: ...}",
		}.String(),
	},
	{
		Name: "_open_browser",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("open_browser", len(args), 1, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("open_browser", 1, STRING_OBJ, args[0].Type())
			}
			url := args[0].(*Stringo).Value
			var err error
			switch runtime.GOOS {
			case "linux":
				err = exec.Command("xdg-open", url).Start()
			case "windows":
				err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
			case "darwin":
				err = exec.Command("open", url).Start()
			default:
				err = fmt.Errorf("unsupported platform")
			}
			if err != nil {
				return newError("`open_browser` error: %s", err.Error())
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`open_browser` will open the user's default browser with the given URL",
			signature:   "open_browser(url: str) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "open_browser('http://localhost:3000/') => null -> open's browser (side effect)",
		}.String(),
	},
	{
		Name: "_handle",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`handle` puts a handler on the server for a given pattern and method, `handle_use` also can use this function if no method is provided",
			signature:   "handle(server: GoObj[*Server], pattern: str, fn: fun, method: str='GET') -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "handle(s, '/', fn) => null",
		}.String(),
	},
	{
		Name: "_handle_use",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`handle` puts a handler on the server for a given pattern and method, `handle_use` also can use this function if no method is provided",
			signature:   "handle(server: GoObj[*Server], pattern: str, fn: fun, method: str='GET') -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "handle(s, '/', fn) => null",
		}.String(),
	},
	{
		Name: "_handle_ws",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`handle_ws` puts a websocket handler on the server for a given pattern and method",
			signature:   "handle_ws(server: GoObj[*Server], pattern: str, fn: fun) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "handle_ws(s, '/ws', fn) => null",
		}.String(),
	},
}
