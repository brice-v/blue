package object

import (
	"blue/consts"
	"blue/lib"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/websocket/v2"
	ws "github.com/gorilla/websocket"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

// Used to catch interrupt to shutdown server
var interruptCh = make(chan os.Signal, 1)

var HttpBuiltins = NewBuiltinSliceType{
	{Name: "_url_encode", Builtin: &Builtin{
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
	}},
	{Name: "_url_escape", Builtin: &Builtin{
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
	}},
	{Name: "_url_unescape", Builtin: &Builtin{
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
	}},
	{Name: "_download", Builtin: &Builtin{
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
			defer resp.Body.Close()
			f, err := os.Create(fname)
			if err != nil {
				return newError("`download` error: %s", err.Error())
			}
			defer f.Close()

			io.Copy(f, resp.Body)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`download` copys the file at the URL to the given file path. if the fpath is empty, then the URL is used to determine the name",
			signature:   "download(url: str, fpath: str='') -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "download('http://example.com/test.txt') => null => writes test.txt to current directory",
		}.String(),
	}},
	{Name: "_new_server", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("new_server", len(args), 1, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("new_server", 1, STRING_OBJ, args[0].Type())
			}
			network := args[0].(*Stringo).Value
			var disableStartupDebug bool
			disableStartupMessageStr := os.Getenv(consts.BLUE_DISABLE_HTTP_SERVER_DEBUG)
			disableStartupDebug, err := strconv.ParseBool(disableStartupMessageStr)
			if err != nil {
				disableStartupDebug = false
			}
			app := fiber.New(fiber.Config{
				Immutable:             true,
				EnablePrintRoutes:     !disableStartupDebug,
				DisableStartupMessage: disableStartupDebug,
				Network:               network,
			})
			return NewGoObj(app)
		},
		HelpStr: helpStrArgs{
			explanation: "`new_server` returns a new server object",
			signature:   "new_server(network: str('tcp','tcp4','tcp6')='tcp4') -> GoObj[*fiber.App]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "new_server() => server obj",
		}.String(),
	}},
	{Name: "_serve", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("serve", len(args), 3, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("serve", 1, GO_OBJ, args[0].Type())
			}
			app, ok := args[0].(*GoObj[*fiber.App])
			if !ok {
				return newPositionalTypeErrorForGoObj("serve", 1, "*fiber.App", args[0])
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
				app.Value.Use(filesystem.New(filesystem.Config{Root: http.FS(sub)}))
			}
			// nil here means use the default server mux (ie. things that were http.HandleFunc's)
			err := app.Value.Listen(addrPort)
			if err != nil {
				return newError("`serve` error: %s", err.Error())
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`serve` starts the http server listener at the given address/port with the embedded lib web files included if set to true",
			signature:   "serve(server: GoObj[*fiber.App], addr_port: str='localhost:3001', use_embedded_lib_web: bool=true) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "serve() => null => starts server",
		}.String(),
	}},
	{Name: "_shutdown_server", Builtin: &Builtin{
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
	}},
	{Name: "_static", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 4 {
				return newInvalidArgCountError("static", len(args), 4, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("static", 1, GO_OBJ, args[0].Type())
			}
			app, ok := args[0].(*GoObj[*fiber.App])
			if !ok {
				return newPositionalTypeErrorForGoObj("static", 1, "*fiber.App", args[0])
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
				app.Value.Use(prefix, filesystem.New(filesystem.Config{
					Root:   http.FS(sub),
					Browse: shouldBrowse,
				}))
			} else {
				app.Value.Static(prefix, fpath, fiber.Static{
					Browse: shouldBrowse,
				})
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`static` serves the given directory as static files for the http server",
			signature:   "static(server: GoObj[*fiber.App], prefix: str='/', dir_path: str='.', browse: bool=false) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "static() => null => current directory served at addr:port/",
		}.String(),
	}},
	{Name: "_ws_send", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("ws_send", len(args), 2, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("ws_send", 1, GO_OBJ, args[0].Type())
			}
			c, ok := args[0].(*GoObj[*websocket.Conn])
			if !ok {
				return newPositionalTypeErrorForGoObj("ws_send", 1, "*websocket.Conn", args[0])
			}
			if args[1].Type() != STRING_OBJ && args[1].Type() != BYTES_OBJ {
				return newPositionalTypeError("ws_send", 2, "STRING or BYTES", args[1].Type())
			}
			var err error
			if args[1].Type() == STRING_OBJ {
				err = c.Value.WriteMessage(websocket.TextMessage, []byte(args[1].(*Stringo).Value))
			} else {
				err = c.Value.WriteMessage(websocket.BinaryMessage, args[1].(*Bytes).Value)
			}
			if err != nil {
				return newError("`ws_send` error: %s", err.Error())
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`ws_send` sends the given value on the websocket connection, if the value is a string the websocket message type is TextMessage, otherwise if bytes BinaryMessage",
			signature:   "ws_send(c: GoObj[*websocket.Conn], value: str|bytes) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "ws_send(c, '1') => null",
		}.String(),
	}},
	{Name: "_ws_recv", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("ws_recv", len(args), 1, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("ws_recv", 1, GO_OBJ, args[0].Type())
			}
			c, ok := args[0].(*GoObj[*websocket.Conn])
			if !ok {
				return newPositionalTypeErrorForGoObj("ws_send", 1, "*websocket.Conn", args[0])
			}
			mt, msg, err := c.Value.ReadMessage()
			if err != nil {
				// If its closed we still want to return an error so that the handler fn wont try to send NULL
				return newError("`ws_recv` error: %s", err.Error())
			}
			switch mt {
			case websocket.BinaryMessage:
				return &Bytes{Value: msg}
			case websocket.TextMessage:
				return &Stringo{Value: string(msg)}
			case websocket.PingMessage:
				return newError("`ws_recv` error: ping message type not supported.")
			case websocket.PongMessage:
				return newError("`ws_recv` error: pong message type not supported.")
			default:
				// If its closed we still want to return an error so that the handler fn wont try to send NULL
				return newError("`ws_recv` error: websocket closed.")
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`ws_recv` receives a websocket message on the given websocket connection",
			signature:   "ws_recv(c: GoObj[*websocket.Conn]) -> str|bytes",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "ws_recv(c) => str|bytes",
		}.String(),
	}},
	{Name: "_new_ws", Builtin: &Builtin{
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
	}},
	{Name: "_ws_client_send", Builtin: &Builtin{
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
				err = c.Value.WriteMessage(websocket.TextMessage, []byte(args[1].(*Stringo).Value))
			} else {
				err = c.Value.WriteMessage(websocket.BinaryMessage, args[1].(*Bytes).Value)
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
	}},
	{Name: "_ws_client_recv", Builtin: &Builtin{
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
			case websocket.BinaryMessage:
				return &Bytes{Value: msg}
			case websocket.TextMessage:
				return &Stringo{Value: string(msg)}
			case websocket.PingMessage:
				return newError("`ws_client_recv` error: ping message type not supported.")
			case websocket.PongMessage:
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
	}},
	{Name: "_handle_monitor", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("handle_monitor", len(args), 3, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("handle_monitor", 1, GO_OBJ, args[0].Type())
			}
			app, ok := args[0].(*GoObj[*fiber.App])
			if !ok {
				return newPositionalTypeErrorForGoObj("handle_monitor", 1, "*fiber.App", args[0])
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("handle_monitor", 2, STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("handle_monitor", 3, BOOLEAN_OBJ, args[2].Type())
			}
			path := args[1].(*Stringo).Value
			shouldShow := args[2].(*Boolean).Value
			app.Value.Get(path, monitor.New(monitor.Config{
				Next: func(c *fiber.Ctx) bool {
					return !shouldShow
				},
			}))
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`handle_monitor` creates a monitor handler on the given http server at the given path a boolean that determines when it should show",
			signature:   "handle_monitor(s: GoObj[*fiber.App], path: str, should_show: bool) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "handle_monitor(s, '/monitor', true) => null",
		}.String(),
	}},
	{Name: "_md_to_html", Builtin: &Builtin{
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
	}},
	{Name: "_sanitize_and_minify", Builtin: &Builtin{
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
			var htmlContent []byte = bs
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
	}},
	{Name: "_inspect", Builtin: &Builtin{
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
				c, ok := args[0].(*GoObj[*websocket.Conn])
				if !ok {
					return newPositionalTypeErrorForGoObj("inspect", 1, "*websocket.Conn", args[0])
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
	}},
	{Name: "_open_browser", Builtin: &Builtin{
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
	}},
	// TODO: Figure out how to handle these
	{Name: "_handle", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			return newError("not supported")
		},
		HelpStr: "",
	}},
	{Name: "_handle_use", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			return newError("not supported")
		},
		HelpStr: "",
	}},
	{Name: "_handle_ws", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			return newError("not supported")
		},
		HelpStr: "",
	}},
}
