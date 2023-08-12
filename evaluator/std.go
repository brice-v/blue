package evaluator

import (
	"blue/ast"
	"blue/consts"
	"blue/lexer"
	"blue/lib"
	"blue/object"
	"blue/parser"
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"database/sql"
	"encoding/base32"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"math"
	mr "math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/antchfx/htmlquery"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/websocket/v2"
	"github.com/golang-module/carbon/v2"
	"github.com/gookit/color"
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/ini"
	"github.com/gookit/config/v2/properties"
	"github.com/gookit/config/v2/toml"
	"github.com/gookit/config/v2/yamlv3"
	"github.com/gookit/ini/v2/dotenv"
	ws "github.com/gorilla/websocket"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	psutilnet "github.com/shirou/gopsutil/v3/net"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/scrypt"
	_ "modernc.org/sqlite"
)

// StdModFileAndBuiltins keeps the file and builtins together for each std lib module
type StdModFileAndBuiltins struct {
	File     string              // File is the actual code used for the module
	Builtins BuiltinMapType      // Builtins is the map of functions to be used by the module
	Env      *object.Environment // Env is the environment to pull the lib functions/variables from
	HelpStr  string              // HelpStr is the help string for the std lib program
}

var _std_mods = map[string]StdModFileAndBuiltins{
	"http":   {File: lib.ReadStdFileToString("http.b"), Builtins: _http_builtin_map},
	"time":   {File: lib.ReadStdFileToString("time.b"), Builtins: _time_builtin_map},
	"search": {File: lib.ReadStdFileToString("search.b"), Builtins: _search_builtin_map},
	"db":     {File: lib.ReadStdFileToString("db.b"), Builtins: _db_builtin_map},
	"math":   {File: lib.ReadStdFileToString("math.b"), Builtins: _math_builtin_map},
	"config": {File: lib.ReadStdFileToString("config.b"), Builtins: _config_builtin_map},
	"crypto": {File: lib.ReadStdFileToString("crypto.b"), Builtins: _crypto_builtin_map},
	"net":    {File: lib.ReadStdFileToString("net.b"), Builtins: _net_builtin_map},
	"ui":     {File: lib.ReadStdFileToString("ui.b"), Builtins: _ui_builtin_map},
	"color":  {File: lib.ReadStdFileToString("color.b"), Builtins: _color_builtin_map},
	"csv":    {File: lib.ReadStdFileToString("csv.b"), Builtins: _csv_builtin_map},
	"psutil": {File: lib.ReadStdFileToString("psutil.b"), Builtins: _psutil_builtin_map},
	"gg":     {File: lib.ReadStdFileToString("gg.b"), Builtins: _gg_builtin_map},
}

func (e *Evaluator) IsStd(name string) bool {
	_, ok := _std_mods[name]
	return ok
}

func (e *Evaluator) AddStdLibToEnv(name string, nodeIdentsToImport []*ast.Identifier, shouldImportAll bool) object.Object {
	if !e.IsStd(name) {
		fmt.Printf("AddStdLibToEnv: '%s' is not in std lib map\n", name)
		os.Exit(1)
	}
	fb := _std_mods[name]
	if fb.Env == nil {
		l := lexer.New(fb.File, "<std/"+name+".b>")
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			for _, msg := range p.Errors() {
				splitMsg := strings.Split(msg, "\n")
				firstPart := fmt.Sprintf("%smodule `%s`: %s\n", consts.PARSER_ERROR_PREFIX, name, splitMsg[0])
				consts.ErrorPrinter(firstPart)
				for i, s := range splitMsg {
					if i == 0 {
						continue
					}
					fmt.Println(s)
				}
			}
			os.Exit(1)
		}
		newE := New()
		newE.Builtins.PushBack(fb.Builtins)
		val := newE.Eval(program)
		if isError(val) {
			errorObj := val.(*object.Error)
			var buf bytes.Buffer
			buf.WriteString(errorObj.Message)
			buf.WriteByte('\n')
			for newE.ErrorTokens.Len() > 0 {
				buf.WriteString(lexer.GetErrorLineMessage(newE.ErrorTokens.PopBack()))
				buf.WriteByte('\n')
			}
			msg := fmt.Sprintf("%smodule `%s`: %s", consts.EVAL_ERROR_PREFIX, name, buf.String())
			splitMsg := strings.Split(msg, "\n")
			for i, s := range splitMsg {
				if i == 0 {
					consts.ErrorPrinter(s + "\n")
					continue
				}
				delimeter := ""
				if i != len(splitMsg)-1 {
					delimeter = "\n"
				}
				fmt.Printf("%s%s", s, delimeter)
			}
			os.Exit(1)
		}
		NewEvaluatorLock.Lock()
		fb.Env = newE.env.Clone()
		pubFunHelpStr := fb.Env.GetPublicFunctionHelpString()
		fb.HelpStr = CreateHelpStringFromProgramTokens(name, program.HelpStrTokens, pubFunHelpStr)
		NewEvaluatorLock.Unlock()
	}

	if len(nodeIdentsToImport) >= 1 {
		for _, ident := range nodeIdentsToImport {
			if strings.HasPrefix(ident.Value, "_") {
				return newError("ImportError: imports must be public to import them. failed to import %s from %s", ident.Value, name)
			}
			o, ok := fb.Env.Get(ident.Value)
			if !ok {
				return newError("ImportError: failed to import %s from %s", ident.Value, name)
			}
			e.env.Set(ident.Value, o)
		}
		// return early if we specifically import some objects
		return NULL
	} else if shouldImportAll {
		// Here we want to import everything from the module
		for k, v := range fb.Env.GetAll() {
			if !strings.HasPrefix(k, "_") {
				e.env.Set(k, v)
			}
		}
		return NULL
	}

	mod := &object.Module{Name: name, Env: fb.Env, HelpStr: fb.HelpStr}
	e.env.Set(name, mod)
	return nil
}

// Used to catch Interupt to shutdown server
var c = make(chan os.Signal, 1)

// Note: Look at how we import the get function in http.b
var _http_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_url_encode": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("url_encode", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("url_encode", 1, object.STRING_OBJ, args[0].Type())
			}
			s := args[0].(*object.Stringo).Value
			u, err := url.Parse(s)
			if err != nil {
				return newError("`url_encode` error: %s", err.Error())
			}
			return &object.Stringo{Value: u.String()}
		},
	},
	"_url_escape": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("url_escape", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("url_escape", 1, object.STRING_OBJ, args[0].Type())
			}
			s := args[0].(*object.Stringo).Value
			return &object.Stringo{Value: url.QueryEscape(s)}
		},
	},
	"_url_unescape": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("url_unescape", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("url_unescape", 1, object.STRING_OBJ, args[1].Type())
			}
			s := args[0].(*object.Stringo).Value
			urlUnescaped, err := url.QueryUnescape(s)
			if err != nil {
				return newError("`url_unescape` error: %s", err.Error())
			}
			return &object.Stringo{Value: urlUnescaped}
		},
	},
	"_download": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("download", len(args), 2, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("download", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("download", 2, object.STRING_OBJ, args[1].Type())
			}
			urlS := args[0].(*object.Stringo).Value
			fname := args[1].(*object.Stringo).Value
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
	},
	"_new_server": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("new_server", len(args), 0, "")
			}
			var disableStartupDebug bool
			disableStartupMessageStr := os.Getenv(consts.DISABLE_HTTP_SERVER_DEBUG)
			disableStartupDebug, err := strconv.ParseBool(disableStartupMessageStr)
			if err != nil {
				disableStartupDebug = false
			}
			app := fiber.New(fiber.Config{
				Immutable:             true,
				EnablePrintRoutes:     !disableStartupDebug,
				DisableStartupMessage: disableStartupDebug,
			})
			return &object.GoObj[*fiber.App]{Value: app, Id: GoObjId.Add(1)}
		},
	},
	"_serve": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("serve", len(args), 3, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("serve", 1, object.GO_OBJ, args[0].Type())
			}
			app, ok := args[0].(*object.GoObj[*fiber.App])
			if !ok {
				return newPositionalTypeErrorForGoObj("serve", 1, "*fiber.App", app)
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("serve", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("seve", 3, object.BOOLEAN_OBJ, args[2].Type())
			}
			useEmbeddedTwindAndPreact := args[2].(*object.Boolean).Value
			addrPort := args[1].(*object.Stringo).Value
			signal.Notify(c, os.Interrupt)
			go func() {
				<-c
				fmt.Println("Interupt... Shutting down http server")
				_ = app.Value.Shutdown()
			}()
			if useEmbeddedTwindAndPreact {
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
	},
	"_shutdown_server": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("shutdown_server", len(args), 0, "")
			}
			c <- os.Interrupt
			return NULL
		},
	},
	"_static": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 4 {
				return newInvalidArgCountError("static", len(args), 4, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("static", 1, object.GO_OBJ, args[0].Type())
			}
			app, ok := args[0].(*object.GoObj[*fiber.App])
			if !ok {
				return newPositionalTypeErrorForGoObj("static", 1, "*fiber.App", app)
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("static", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.STRING_OBJ {
				return newPositionalTypeError("static", 3, object.STRING_OBJ, args[2].Type())
			}
			if args[3].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("static", 4, object.BOOLEAN_OBJ, args[3].Type())
			}
			prefix := args[1].(*object.Stringo).Value
			fpath := args[2].(*object.Stringo).Value
			shouldBrowse := args[3].(*object.Boolean).Value
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
	},
	"_ws_send": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("ws_send", len(args), 2, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("ws_send", 1, object.GO_OBJ, args[0].Type())
			}
			c, ok := args[0].(*object.GoObj[*websocket.Conn])
			if !ok {
				return newPositionalTypeErrorForGoObj("ws_send", 1, "*websocket.Conn", c)
			}
			if args[1].Type() != object.STRING_OBJ && args[1].Type() != object.BYTES_OBJ {
				return newPositionalTypeError("ws_send", 2, "STRING or BYTES", args[1].Type())
			}
			var err error
			if args[1].Type() == object.STRING_OBJ {
				err = c.Value.WriteMessage(websocket.TextMessage, []byte(args[1].(*object.Stringo).Value))
			} else {
				err = c.Value.WriteMessage(websocket.BinaryMessage, args[1].(*object.Bytes).Value)
			}
			if err != nil {
				return newError("`ws_send` error: %s", err.Error())
			}
			return NULL
		},
	},
	"_ws_recv": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("ws_recv", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("ws_recv", 1, object.GO_OBJ, args[0].Type())
			}
			c, ok := args[0].(*object.GoObj[*websocket.Conn])
			if !ok {
				return newPositionalTypeErrorForGoObj("ws_send", 1, "*websocket.Conn", c)
			}
			mt, msg, err := c.Value.ReadMessage()
			if err != nil {
				// If its closed we still want to return an error so that the handler fn wont try to send NULL
				return newError("`ws_recv`: %s", err.Error())
			}
			switch mt {
			case websocket.BinaryMessage:
				return &object.Bytes{Value: msg}
			case websocket.TextMessage:
				return &object.Stringo{Value: string(msg)}
			case websocket.PingMessage:
				return newError("`ws_recv`: ping message type not supported.")
			case websocket.PongMessage:
				return newError("`ws_recv`: pong message type not supported.")
			default:
				// If its closed we still want to return an error so that the handler fn wont try to send NULL
				return newError("`ws_recv`: websocket closed.")
			}
		},
	},
	"_new_ws": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("new_ws", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("new_ws", 1, object.STRING_OBJ, args[0].Type())
			}
			url := args[0].(*object.Stringo).Value
			conn, _, err := ws.DefaultDialer.Dial(url, nil)
			if err != nil {
				return newError("`new_ws` error: %s", err.Error())
			}
			return object.CreateBasicMapObjectForGoObj("ws/client", &object.GoObj[*ws.Conn]{Value: conn, Id: GoObjId.Add(1)})
		},
	},
	"_ws_client_send": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("ws_client_send", len(args), 2, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("ws_client_send", 1, object.GO_OBJ, args[0].Type())
			}
			c, ok := args[0].(*object.GoObj[*ws.Conn])
			if !ok {
				return newPositionalTypeErrorForGoObj("ws_client_send", 1, "*ws.Conn", c)
			}
			if args[1].Type() != object.STRING_OBJ && args[1].Type() != object.BYTES_OBJ {
				return newPositionalTypeError("ws_client_send", 2, "STRING or BYTES", args[1].Type())
			}
			var err error
			if args[1].Type() == object.STRING_OBJ {
				err = c.Value.WriteMessage(websocket.TextMessage, []byte(args[1].(*object.Stringo).Value))
			} else {
				err = c.Value.WriteMessage(websocket.BinaryMessage, args[1].(*object.Bytes).Value)
			}
			if err != nil {
				return newError("`ws_send` error: %s", err.Error())
			}
			return NULL
		},
	},
	"_ws_client_recv": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("ws_client_recv", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("ws_client_recv", 1, object.GO_OBJ, args[0].Type())
			}
			c, ok := args[0].(*object.GoObj[*ws.Conn])
			if !ok {
				return newPositionalTypeErrorForGoObj("ws_client_recv", 1, "*ws.Conn", c)
			}
			mt, msg, err := c.Value.ReadMessage()
			if err != nil {
				// If its closed we still want to return an error so that the handler fn wont try to send NULL
				return newError("`ws_recv`: %s", err.Error())
			}
			switch mt {
			case websocket.BinaryMessage:
				return &object.Bytes{Value: msg}
			case websocket.TextMessage:
				return &object.Stringo{Value: string(msg)}
			case websocket.PingMessage:
				return newError("`ws_recv`: ping message type not supported.")
			case websocket.PongMessage:
				return newError("`ws_recv`: pong message type not supported.")
			default:
				// If its closed we still want to return an error so that the handler fn wont try to send NULL
				return newError("`ws_recv`: websocket closed.")
			}
		},
	},
	"_handle_monitor": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("handle_monitor", len(args), 3, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("handle_monitor", 1, object.GO_OBJ, args[0].Type())
			}
			app, ok := args[0].(*object.GoObj[*fiber.App])
			if !ok {
				return newPositionalTypeErrorForGoObj("handle_monitor", 1, "*fiber.App", app)
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("handle_monitor", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("handle_monitor", 3, object.BOOLEAN_OBJ, args[2].Type())
			}
			path := args[1].(*object.Stringo).Value
			shouldShow := args[2].(*object.Boolean).Value
			app.Value.Get(path, monitor.New(monitor.Config{
				Next: func(c *fiber.Ctx) bool {
					return !shouldShow
				},
			}))
			return NULL
		},
	},
	"_md_to_html": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("md_to_html", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("md_to_html", 1, object.STRING_OBJ, args[0].Type())
			}
			bs := []byte(args[0].(*object.Stringo).Value)
			output := blackfriday.Run(bs)
			return &object.Stringo{Value: string(output)}
		},
	},
	"_sanitize_and_minify": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("sanitize_and_minify", len(args), 3, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("sanitize_and_minify", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("sanitize_and_minify", 2, object.BOOLEAN_OBJ, args[1].Type())
			}
			if args[2].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("sanitize_and_minify", 3, object.BOOLEAN_OBJ, args[2].Type())
			}
			bs := []byte(args[0].(*object.Stringo).Value)
			shouldSanitize := args[1].(*object.Boolean).Value
			shouldMinify := args[2].(*object.Boolean).Value
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
			return &object.Stringo{Value: string(htmlContent)}
		},
	},
	"_inspect": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("inspect", len(args), 2, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("inspect", 1, object.GO_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("inspect", 2, object.STRING_OBJ, args[1].Type())
			}
			t := args[1].(*object.Stringo).Value
			switch t {
			case "ws":
				c, ok := args[0].(*object.GoObj[*websocket.Conn])
				if !ok {
					return newPositionalTypeErrorForGoObj("inspect", 1, "*websocket.Conn", c)
				}
				mapObj := object.NewOrderedMap[string, object.Object]()
				mapObj.Set("remote_addr", &object.Stringo{Value: c.Value.RemoteAddr().String()})
				mapObj.Set("local_addr", &object.Stringo{Value: c.Value.LocalAddr().String()})
				mapObj.Set("remote_addr_network", &object.Stringo{Value: c.Value.RemoteAddr().Network()})
				mapObj.Set("local_addr_network", &object.Stringo{Value: c.Value.LocalAddr().Network()})
				return object.CreateMapObjectForGoMap(*mapObj)
			case "ws/client":
				c, ok := args[0].(*object.GoObj[*ws.Conn])
				if !ok {
					return newPositionalTypeErrorForGoObj("inspect", 1, "*ws.Conn", c)
				}
				mapObj := object.NewOrderedMap[string, object.Object]()
				mapObj.Set("remote_addr", &object.Stringo{Value: c.Value.RemoteAddr().String()})
				mapObj.Set("local_addr", &object.Stringo{Value: c.Value.LocalAddr().String()})
				mapObj.Set("remote_addr_network", &object.Stringo{Value: c.Value.RemoteAddr().Network()})
				mapObj.Set("local_addr_network", &object.Stringo{Value: c.Value.LocalAddr().Network()})
				return object.CreateMapObjectForGoMap(*mapObj)
			default:
				return newError("`inspect` expects type of 'ws'")
			}
		},
	},
})

var _time_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_sleep": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("sleep", len(args), 1, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("sleep", 1, object.INTEGER_OBJ, args[0].Type())
			}
			i := args[0].(*object.Integer).Value
			if i < 0 {
				return newError("INTEGER argument to `sleep` must be > 0, got=%d", i)
			}
			time.Sleep(time.Duration(i) * time.Millisecond)
			return NULL
		},
	},
	"_now": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("now", len(args), 0, "")
			}
			return &object.Integer{Value: time.Now().UnixMilli()}
		},
	},
	"_parse": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("parse", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("parse", 1, object.STRING_OBJ, args[0].Type())
			}
			s := args[0].(*object.Stringo).Value
			return &object.Integer{Value: carbon.Parse(s).ToStdTime().UnixMilli()}
		},
	},
	"_to_str": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("to_str", len(args), 2, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("to_str", 1, object.INTEGER_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ && args[1].Type() != object.NULL_OBJ {
				return newPositionalTypeError("to_str", 2, "STRING or NULL", args[1].Type())
			}
			i := args[0].(*object.Integer).Value
			tm := time.UnixMilli(i)
			if args[1].Type() == object.STRING_OBJ {
				tz := args[1].(*object.Stringo).Value
				return &object.Stringo{Value: carbon.FromStdTime(tm).ToDateTimeMilliString(tz)}
			} else {
				return &object.Stringo{Value: carbon.FromStdTime(tm).ToDateTimeMilliString()}
			}
		},
	},
})

var _search_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_by_xpath": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("by_xpath", len(args), 3, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("by_xpath", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("by_xpath", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("by_xpath", 3, object.BOOLEAN_OBJ, args[2].Type())
			}
			strToSearch := args[0].(*object.Stringo).Value
			if strToSearch == "" {
				return newError("`by_xpath` error: str_to_search argument is empty")
			}
			strQuery := args[1].(*object.Stringo).Value
			if strQuery == "" {
				return newError("`by_xpath` error: query argument is empty")
			}
			shouldFindOne := args[2].(*object.Boolean).Value
			doc, err := htmlquery.Parse(strings.NewReader(strToSearch))
			if err != nil {
				return newError("`by_xpath` failed to parse document as html: error %s", err.Error())
			}
			if !shouldFindOne {
				listToReturn := &object.List{Elements: []object.Object{}}
				for _, e := range htmlquery.Find(doc, strQuery) {
					result := htmlquery.OutputHTML(e, true)
					listToReturn.Elements = append(listToReturn.Elements, &object.Stringo{Value: result})
				}
				return listToReturn
			} else {
				e := htmlquery.FindOne(doc, strQuery)
				result := htmlquery.OutputHTML(e, true)
				return &object.Stringo{Value: result}
			}
		},
	},
	"_by_regex": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("by_regex", len(args), 3, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("by_regex", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("by_regex", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("by_regex", 3, object.BOOLEAN_OBJ, args[2].Type())
			}
			strToSearch := args[0].(*object.Stringo).Value
			if strToSearch == "" {
				return newError("`by_regex` error: str_to_search argument is empty")
			}
			strQuery := args[1].(*object.Stringo).Value
			if strQuery == "" {
				return newError("`by_regex` error: query argument is empty")
			}
			shouldFindOne := args[2].(*object.Boolean).Value
			re, err := regexp.Compile(strQuery)
			if err != nil {
				return newError("`by_regex` error: failed to compile regexp %q", strQuery)
			}
			if !shouldFindOne {
				listToReturn := &object.List{Elements: []object.Object{}}
				results := re.FindAllString(strToSearch, -1)
				for _, str := range results {
					listToReturn.Elements = append(listToReturn.Elements, &object.Stringo{Value: str})
				}
				return listToReturn
			} else {
				result := re.FindString(strToSearch)
				return &object.Stringo{Value: result}
			}
		},
	},
})

var _db_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_db_open": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("db_open", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("db_open", 1, object.STRING_OBJ, args[0].Type())
			}
			dbName := args[0].(*object.Stringo).Value
			if dbName == "" {
				return newError("`db_open` error: db_name argument is empty")
			}
			db, err := sql.Open("sqlite", dbName)
			if err != nil {
				return newError("`db_open` error: %s", err.Error())
			}
			return &object.GoObj[*sql.DB]{Value: db, Id: GoObjId.Add(1)}
		},
	},
	"_db_ping": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("db_ping", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("db_ping", 1, object.GO_OBJ, args[0].Type())
			}
			db, ok := args[0].(*object.GoObj[*sql.DB])
			if !ok {
				return newPositionalTypeErrorForGoObj("db_ping", 1, "*sql.DB", db)
			}
			err := db.Value.Ping()
			if err != nil {
				return &object.Stringo{Value: err.Error()}
			}
			return NULL
		},
	},
	"_db_close": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("db_close", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("db_close", 1, object.GO_OBJ, args[0].Type())
			}
			db, ok := args[0].(*object.GoObj[*sql.DB])
			if !ok {
				return newPositionalTypeErrorForGoObj("db_close", 1, "*sql.DB", db)
			}
			err := db.Value.Close()
			if err != nil {
				return newError("`db_close` error: %s", err.Error())
			}
			return NULL
		},
	},
	"_db_exec": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("db_exec", len(args), 3, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("db_exec", 1, object.GO_OBJ, args[0].Type())
			}
			db, ok := args[0].(*object.GoObj[*sql.DB])
			if !ok {
				return newPositionalTypeErrorForGoObj("db_exec", 1, "*sql.DB", db)
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("db_exec", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.LIST_OBJ {
				return newPositionalTypeError("db_exec", 3, object.LIST_OBJ, args[2].Type())
			}
			s := args[1].(*object.Stringo).Value
			l := args[2].(*object.List).Elements

			var result sql.Result
			var err error
			if len(l) > 1 {
				execArgs := make([]any, len(l))
				for idx, e := range l {
					switch e.Type() {
					case object.STRING_OBJ:
						execArgs[idx] = e.(*object.Stringo).Value
					case object.INTEGER_OBJ:
						execArgs[idx] = e.(*object.Integer).Value
					case object.FLOAT_OBJ:
						execArgs[idx] = e.(*object.Float).Value
					case object.NULL_OBJ:
						execArgs[idx] = nil
					case object.BOOLEAN_OBJ:
						execArgs[idx] = e.(*object.Boolean).Value
					case object.BYTES_OBJ:
						execArgs[idx] = e.(*object.Bytes).Value
					default:
						return newError("argument list to `db_exec` included invalid type. got=%s", e.Type())
					}
				}
				result, err = db.Value.Exec(s, execArgs...)
			} else {
				result, err = db.Value.Exec(s)
			}
			if err != nil {
				return newError("`db_exec` error: %s", err.Error())
			}
			lastInsertId, err := result.LastInsertId()
			if err != nil {
				return newError("`db_exec` error: %s", err.Error())
			}
			rowsAffected, err := result.RowsAffected()
			if err != nil {
				return newError("`db_exec` error: %s", err.Error())
			}
			mapToConvert := object.NewOrderedMap[string, object.Object]()
			mapToConvert.Set("last_insert_id", &object.Integer{Value: lastInsertId})
			mapToConvert.Set("rows_affected", &object.Integer{Value: rowsAffected})
			return object.CreateMapObjectForGoMap(*mapToConvert)
		},
	},
	"_db_query": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 4 {
				return newInvalidArgCountError("db_query", len(args), 4, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("db_query", 1, object.GO_OBJ, args[0].Type())
			}
			db, ok := args[0].(*object.GoObj[*sql.DB])
			if !ok {
				return newPositionalTypeErrorForGoObj("db_query", 1, "*sql.DB", db)
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("db_query", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.LIST_OBJ {
				return newPositionalTypeError("db_query", 3, object.LIST_OBJ, args[2].Type())
			}
			if args[3].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("db_query", 4, object.BOOLEAN_OBJ, args[3].Type())
			}
			s := args[1].(*object.Stringo).Value
			l := args[2].(*object.List).Elements
			isNamedCols := args[3].(*object.Boolean).Value
			var rows *sql.Rows
			var err error
			if len(l) >= 1 {
				execArgs := make([]any, len(l))
				for idx, e := range l {
					switch e.Type() {
					case object.STRING_OBJ:
						execArgs[idx] = e.(*object.Stringo).Value
					case object.INTEGER_OBJ:
						execArgs[idx] = e.(*object.Integer).Value
					case object.FLOAT_OBJ:
						execArgs[idx] = e.(*object.Float).Value
					case object.NULL_OBJ:
						execArgs[idx] = nil
					case object.BOOLEAN_OBJ:
						execArgs[idx] = e.(*object.Boolean).Value
					case object.BYTES_OBJ:
						execArgs[idx] = e.(*object.Bytes).Value
					default:
						return newError("argument list to `db_query` included invalid type. got=%s", e.Type())
					}
				}
				rows, err = db.Value.Query(s, execArgs...)
			} else {
				rows, err = db.Value.Query(s)
			}
			if rows != nil {
				defer rows.Close()
			}
			if err != nil {
				return newError("`db_query` error: %s", err.Error())
			}
			colNames, err := rows.Columns()
			if err != nil {
				return newError("`db_query` error: %s", err.Error())
			}
			// Get Types to properly scan
			// https://www.sqlite.org/datatype3.html
			// NULL. The value is a NULL value.
			// INTEGER. The value is a signed integer, stored in 0, 1, 2, 3, 4, 6, or 8 bytes depending on the magnitude of the value.
			// REAL. The value is a floating point value, stored as an 8-byte IEEE floating point number.
			// TEXT. The value is a text string, stored using the database encoding (UTF-8, UTF-16BE or UTF-16LE).
			// BLOB. The value is a blob of data, stored exactly as it was input.
			cols := make([]interface{}, len(colNames))
			colPtrs := make([]interface{}, len(colNames))
			for i := 0; i < len(colNames); i++ {
				colPtrs[i] = &cols[i]
			}
			returnList := &object.List{
				Elements: []object.Object{},
			}
			for rows.Next() {
				err = rows.Scan(colPtrs...)
				if err != nil {
					return newError("`db_query` error: %s", err.Error())
				}
				rowList := &object.List{
					Elements: make([]object.Object, len(cols)),
				}
				rowMap := object.NewOrderedMap[string, object.Object]()
				for idx, e := range cols {
					obj := object.CreateObjectFromDbInterface(e)
					if obj == nil {
						obj = NULL
					}
					if !isNamedCols {
						rowList.Elements[idx] = obj
					} else {
						rowMap.Set(colNames[idx], obj)
					}
				}
				if !isNamedCols {
					returnList.Elements = append(returnList.Elements, rowList)
				} else {
					returnList.Elements = append(returnList.Elements, object.CreateMapObjectForGoMap(*rowMap))
				}
			}
			return returnList
		},
	},
})

var _math_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_rand": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("rand", len(args), 0, "")
			}
			return &object.Float{Value: mr.Float64()}
		},
	},
	"_NaN": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("NaN", len(args), 0, "")
			}
			return &object.Float{Value: math.NaN()}
		},
	},
	"acos": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("acos", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("acos", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Acos(x)}
		},
	},
	"acosh": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("acosh", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("acosh", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Acosh(x)}
		},
	},
	"asin": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("asin", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("asin", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Asin(x)}
		},
	},
	"asinh": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("asinh", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("asinh", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Asinh(x)}
		},
	},
	"atan": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("atan", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("atan", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Atan(x)}
		},
	},
	"atan2": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("atan2", len(args), 2, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("atan2", 1, object.FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("atan2", 2, object.FLOAT_OBJ, args[1].Type())
			}
			x := args[0].(*object.Float).Value
			y := args[1].(*object.Float).Value
			return &object.Float{Value: math.Atan2(x, y)}
		},
	},
	"atanh": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("atanh", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("atanh", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Atanh(x)}
		},
	},
	"cbrt": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("cbrt", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("cbrt", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Cbrt(x)}
		},
	},
	"ceil": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("ceil", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("ceil", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Ceil(x)}
		},
	},
	"copysign": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("copysign", len(args), 2, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("copysign", 1, object.FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("copysign", 2, object.FLOAT_OBJ, args[1].Type())
			}
			f := args[0].(*object.Float).Value
			sign := args[1].(*object.Float).Value
			return &object.Float{Value: math.Copysign(f, sign)}
		},
	},
	"cos": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("cos", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("cos", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Cos(x)}
		},
	},
	"cosh": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("cosh", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("cosh", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Cosh(x)}
		},
	},
	"dim": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("dim", len(args), 2, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("dim", 1, object.FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("dim", 2, object.FLOAT_OBJ, args[1].Type())
			}
			x := args[0].(*object.Float).Value
			y := args[1].(*object.Float).Value
			return &object.Float{Value: math.Dim(x, y)}
		},
	},
	"erf": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("erf", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("erf", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Erf(x)}
		},
	},
	"erfc": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("erfc", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("erfc", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Erfc(x)}
		},
	},
	"erfcinv": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("erfcinv", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("erfcinv", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Erfcinv(x)}
		},
	},
	"erfinv": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("erfinv", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("erfinv", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Erfinv(x)}
		},
	},
	"exp": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("exp", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("exp", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Exp(x)}
		},
	},
	"exp2": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("exp2", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("exp2", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Exp2(x)}
		},
	},
	"expm1": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("expm1", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("expm1", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Expm1(x)}
		},
	},
	"fma": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("fma", len(args), 3, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("fma", 1, object.FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("fma", 2, object.FLOAT_OBJ, args[1].Type())
			}
			if args[2].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("fma", 3, object.FLOAT_OBJ, args[2].Type())
			}
			x := args[0].(*object.Float).Value
			y := args[1].(*object.Float).Value
			z := args[2].(*object.Float).Value
			return &object.Float{Value: math.FMA(x, y, z)}
		},
	},
	"floor": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("floor", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("floor", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Floor(x)}
		},
	},
	"frexp": {Fun: func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return newInvalidArgCountError("frexp", len(args), 1, "")
		}
		if args[0].Type() != object.FLOAT_OBJ {
			return newPositionalTypeError("frexp", 1, object.FLOAT_OBJ, args[0].Type())
		}
		frac, exp := math.Frexp(args[0].(*object.Float).Value)
		mapObj := object.NewOrderedMap[string, object.Object]()
		mapObj.Set("frac", &object.Float{Value: frac})
		mapObj.Set("exp", &object.Integer{Value: int64(exp)})
		return object.CreateMapObjectForGoMap(*mapObj)
	}},
	"gamma": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("gamma", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("gamma", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Gamma(x)}
		},
	},
	"hypot": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("hypot", len(args), 2, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("hypot", 1, object.FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("hypot", 2, object.FLOAT_OBJ, args[1].Type())
			}
			p := args[0].(*object.Float).Value
			q := args[1].(*object.Float).Value
			return &object.Float{Value: math.Hypot(p, q)}
		},
	},
	"ilogb": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("ilogb", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("ilogb", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Integer{Value: int64(math.Ilogb(x))}
		},
	},
	"inf": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("inf", len(args), 1, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("inf", 1, object.INTEGER_OBJ, args[0].Type())
			}
			sign := args[0].(*object.Integer).Value
			return &object.Float{Value: math.Inf(int(sign))}
		},
	},
	"is_inf": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("is_inf", len(args), 2, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("is_inf", 1, object.FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("is_inf", 2, object.INTEGER_OBJ, args[1].Type())
			}
			f := args[0].(*object.Float).Value
			sign := int(args[1].(*object.Integer).Value)
			return nativeToBooleanObject(math.IsInf(f, sign))
		},
	},
	"is_NaN": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("is_NaN", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("is_NaN", 1, object.FLOAT_OBJ, args[0].Type())
			}
			f := args[0].(*object.Float).Value
			return nativeToBooleanObject(math.IsNaN(f))
		},
	},
	"j0": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("j0", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("j0", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.J0(x)}
		},
	},
	"j1": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("j1", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("j1", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.J1(x)}
		},
	},
	"jn": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("jn", len(args), 2, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("jn", 1, object.INTEGER_OBJ, args[0].Type())
			}
			if args[1].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("jn", 2, object.FLOAT_OBJ, args[1].Type())
			}
			n := int(args[0].(*object.Integer).Value)
			x := args[1].(*object.Float).Value
			return &object.Float{Value: math.Jn(n, x)}
		},
	},
	"ldexp": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("ldexp", len(args), 2, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("ldexp", 1, object.FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("ldexp", 2, object.INTEGER_OBJ, args[1].Type())
			}
			frac := args[0].(*object.Float).Value
			exp := int(args[1].(*object.Integer).Value)
			return &object.Float{Value: math.Ldexp(frac, exp)}
		},
	},
	"lgamma": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("lgamma", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("lgamma", 1, object.FLOAT_OBJ, args[0].Type())
			}
			lgamma, sign := math.Lgamma(args[0].(*object.Float).Value)
			mapObj := object.NewOrderedMap[string, object.Object]()
			mapObj.Set("lgamma", &object.Float{Value: lgamma})
			mapObj.Set("sign", &object.Integer{Value: int64(sign)})
			return object.CreateMapObjectForGoMap(*mapObj)
		},
	},
	"log": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("log", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("log", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Log(x)}
		},
	},
	"log10": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("log10", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("log10", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Log10(x)}
		},
	},
	"log1p": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("log1p", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("log1p", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Log1p(x)}
		},
	},
	"log2": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("log2", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("log2", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Log2(x)}
		},
	},
	"logb": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("logb", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("logb", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Logb(x)}
		},
	},
	"mod": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("mod", len(args), 2, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("mod", 1, object.FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("mod", 2, object.FLOAT_OBJ, args[1].Type())
			}
			x := args[0].(*object.Float).Value
			y := args[1].(*object.Float).Value
			return &object.Float{Value: math.Mod(x, y)}
		},
	},
	"modf": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("modf", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("modf", 1, object.FLOAT_OBJ, args[0].Type())
			}
			i, frac := math.Modf(args[0].(*object.Float).Value)
			mapObj := object.NewOrderedMap[string, object.Object]()
			mapObj.Set("i", &object.Integer{Value: int64(i)})
			mapObj.Set("frac", &object.Float{Value: frac})
			return object.CreateMapObjectForGoMap(*mapObj)
		},
	},
	"next_after": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("next_after", len(args), 2, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("next_after", 1, object.FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("next_after", 2, object.FLOAT_OBJ, args[1].Type())
			}
			x := args[0].(*object.Float).Value
			y := args[1].(*object.Float).Value
			return &object.Float{Value: math.Nextafter(x, y)}
		},
	},
	"remainder": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("remainder", len(args), 2, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("remainder", 1, object.FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("remainder", 2, object.FLOAT_OBJ, args[1].Type())
			}
			x := args[0].(*object.Float).Value
			y := args[1].(*object.Float).Value
			return &object.Float{Value: math.Remainder(x, y)}
		},
	},
	"round": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("round", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("round", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Round(x)}
		},
	},
	"round_to_even": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("round_to_even", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("round_to_even", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.RoundToEven(x)}
		},
	},
	"signbit": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("signbit", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("signbit", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return nativeToBooleanObject(math.Signbit(x))
		},
	},
	"sin": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("sin", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("sin", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Sin(x)}
		},
	},
	"sincos": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("sincos", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("sincos", 1, object.FLOAT_OBJ, args[0].Type())
			}
			sin, cos := math.Sincos(args[0].(*object.Float).Value)
			mapObj := object.NewOrderedMap[string, object.Object]()
			mapObj.Set("sin", &object.Float{Value: sin})
			mapObj.Set("cos", &object.Float{Value: cos})
			return object.CreateMapObjectForGoMap(*mapObj)
		},
	},
	"sinh": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("sinh", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("sinh", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Sinh(x)}
		},
	},
	"tan": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("tan", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("tan", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Tan(x)}
		},
	},
	"tanh": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("tanh", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("tanh", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Tanh(x)}
		},
	},
	"trunc": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("trunc", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("trunc", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Trunc(x)}
		},
	},
	"y0": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("y0", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("y0", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Y0(x)}
		},
	},
	"y1": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("y1", len(args), 1, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("y1", 1, object.FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Y1(x)}
		},
	},
	"yn": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("yn", len(args), 2, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("yn", 1, object.INTEGER_OBJ, args[0].Type())
			}
			if args[1].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("yn", 2, object.FLOAT_OBJ, args[1].Type())
			}
			n := int(args[0].(*object.Integer).Value)
			x := args[1].(*object.Float).Value
			return &object.Float{Value: math.Yn(n, x)}
		},
	},
})

var _config_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_load_file": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("load_file", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("load_file", 1, object.STRING_OBJ, args[0].Type())
			}
			c := config.NewWithOptions("config", config.ParseEnv, config.Readonly)
			c.WithDriver(yamlv3.Driver, ini.Driver, toml.Driver, properties.Driver)
			fpath := args[0].(*object.Stringo).Value
			err := c.LoadFiles(fpath)
			if err != nil {
				if err.Error() == "not register decoder for the format: env" {
					err = dotenv.LoadFiles(fpath)
					builtinobjs["ENV"] = &object.BuiltinObj{
						Obj: populateENVObj(),
					}
					if err != nil {
						return newError("`load_file` error: %s", err.Error())
					} else {
						// Need to return a valid JSON value
						return &object.Stringo{Value: "{}"}
					}
				}
				return newError("`load_file` error: %s", err.Error())
			}
			return &object.Stringo{Value: c.ToJSON()}
		},
	},
	"_dump_config": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("dump_config", len(args), 3, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("dump_config", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("dump_config", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.STRING_OBJ {
				return newPositionalTypeError("dump_config", 3, object.STRING_OBJ, args[2].Type())
			}
			c := config.New("config")
			configAsJson := args[0].(*object.Stringo).Value
			c.LoadStrings(config.JSON, configAsJson)
			fpath := args[1].(*object.Stringo).Value
			format := args[2].(*object.Stringo).Value
			out := new(bytes.Buffer)
			switch format {
			case "JSON":
				config.DumpTo(out, config.JSON)
			case "TOML":
				config.DumpTo(out, config.Toml)
			case "YAML":
				config.DumpTo(out, config.Yaml)
			case "INI":
				config.DumpTo(out, config.Ini)
			case "PROPERTIES":
				config.DumpTo(out, config.Prop)
			}
			err := os.WriteFile(fpath, out.Bytes(), 0755)
			if err != nil {
				return newError("`dump_config` error: %s", err.Error())
			}
			return NULL
		},
	},
})

var _crypto_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_sha": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("sha", len(args), 2, "")
			}
			if args[0].Type() != object.STRING_OBJ && args[0].Type() != object.BYTES_OBJ {
				return newPositionalTypeError("sha", 1, "STRING or BYTES", args[0].Type())
			}
			if args[1].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("sha", 2, object.INTEGER_OBJ, args[1].Type())
			}
			var bs []byte
			if args[0].Type() == object.STRING_OBJ {
				bs = []byte(args[0].(*object.Stringo).Value)
			} else {
				bs = args[0].(*object.Bytes).Value
			}
			i := args[1].(*object.Integer).Value
			var hasher hash.Hash
			switch i {
			case 1:
				hasher = sha1.New()
			case 256:
				hasher = sha256.New()
			case 512:
				hasher = sha512.New()
			default:
				return newError("argument 2 to `sha` should be 1, 256, or 512. got=%d", i)
			}
			hasher.Write(bs)
			return &object.Stringo{Value: fmt.Sprintf("%x", hasher.Sum(nil))}
		},
	},
	"_md5": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("md5", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ && args[0].Type() != object.BYTES_OBJ {
				return newPositionalTypeError("md5", 1, "STRING or BYTES", args[0].Type())
			}
			var bs []byte
			if args[0].Type() == object.STRING_OBJ {
				bs = []byte(args[0].(*object.Stringo).Value)
			} else {
				bs = args[0].(*object.Bytes).Value
			}
			hasher := md5.New()
			hasher.Write(bs)
			return &object.Stringo{Value: fmt.Sprintf("%x", hasher.Sum(nil))}
		},
	},
	"_generate_from_password": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("generate_from_password", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("generate_from_password", 1, object.STRING_OBJ, args[0].Type())
			}
			pw := []byte(args[0].(*object.Stringo).Value)
			hashedPw, err := bcrypt.GenerateFromPassword(pw, bcrypt.DefaultCost)
			if err != nil {
				return newError("bcrypt error: %s", err.Error())
			}
			return &object.Stringo{Value: string(hashedPw)}
		},
	},
	"_compare_hash_and_password": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("compare_hash_and_password", len(args), 2, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("compare_hash_and_password", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("compare_hash_and_password", 2, object.STRING_OBJ, args[1].Type())
			}
			hashedPw := []byte(args[0].(*object.Stringo).Value)
			pw := []byte(args[1].(*object.Stringo).Value)
			err := bcrypt.CompareHashAndPassword(hashedPw, pw)
			if err != nil {
				return newError("bcrypt error: %s", err.Error())
			}
			return TRUE
		},
	},
	"_encrypt": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("encrypt", len(args), 2, "")
			}
			if args[0].Type() != object.STRING_OBJ && args[0].Type() != object.BYTES_OBJ {
				return newPositionalTypeError("encrypt", 1, "STRING or BYTES", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ && args[1].Type() != object.BYTES_OBJ {
				return newPositionalTypeError("encrypt", 2, "STRING or BYTES", args[1].Type())
			}
			var pw []byte
			if args[0].Type() == object.STRING_OBJ {
				pw = []byte(args[0].(*object.Stringo).Value)
			} else {
				pw = args[0].(*object.Bytes).Value
			}
			var data []byte
			if args[1].Type() == object.STRING_OBJ {
				data = []byte(args[1].(*object.Stringo).Value)
			} else {
				data = args[1].(*object.Bytes).Value
			}

			// Deriving key from pw as it needs to be 32 bytes
			salt := make([]byte, 32)
			if _, err := rand.Read(salt); err != nil {
				return newError("encrypt error: %s", err.Error())
			}
			key, err := scrypt.Key(pw, salt, 1048576, 8, 1, 32)
			if err != nil {
				return newError("encrypt error: %s", err.Error())
			}
			// Done Deriving key

			blockCipher, err := aes.NewCipher(key)
			if err != nil {
				return newError("encrypt error: %s", err.Error())
			}
			gcm, err := cipher.NewGCM(blockCipher)
			if err != nil {
				return newError("encrypt error: %s", err.Error())
			}
			nonce := make([]byte, gcm.NonceSize())
			if _, err = rand.Read(nonce); err != nil {
				return newError("encrypt error: %s", err.Error())
			}
			ciphertext := gcm.Seal(nonce, nonce, data, nil)
			ciphertext = append(ciphertext, salt...)
			return &object.Bytes{Value: ciphertext}
		},
	},
	"_decrypt": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("decrypt", len(args), 3, "")
			}
			if args[0].Type() != object.STRING_OBJ && args[0].Type() != object.BYTES_OBJ {
				return newPositionalTypeError("decrypt", 1, "STRING or BYTES", args[0].Type())
			}
			if args[1].Type() != object.BYTES_OBJ {
				return newPositionalTypeError("decrypt", 2, object.BYTES_OBJ, args[1].Type())
			}
			if args[2].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("decrypt", 3, object.BOOLEAN_OBJ, args[2].Type())
			}
			var pw []byte
			if args[0].Type() == object.STRING_OBJ {
				pw = []byte(args[0].(*object.Stringo).Value)
			} else {
				pw = args[0].(*object.Bytes).Value
			}
			data := args[1].(*object.Bytes).Value
			getDataAsBytes := args[2].(*object.Boolean).Value

			// Deriving key from pw as it needs to be 32 bytes
			salt, data := data[len(data)-32:], data[:len(data)-32]
			key, err := scrypt.Key(pw, salt, 1048576, 8, 1, 32)
			if err != nil {
				return newError("decrypt error: %s", err.Error())
			}
			// Done Deriving key

			blockCipher, err := aes.NewCipher(key)
			if err != nil {
				return newError("decrypt error: %s", err.Error())
			}
			gcm, err := cipher.NewGCM(blockCipher)
			if err != nil {
				return newError("decrypt error: %s", err.Error())
			}
			nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
			plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
			if err != nil {
				return newError("decrypt error: %s", err.Error())
			}
			if getDataAsBytes {
				return &object.Bytes{Value: plaintext}
			} else {
				return &object.Stringo{Value: string(plaintext)}
			}
		},
	},
	"_encode_base_64_32": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("encode_base_64_32", len(args), 3, "")
			}
			if args[0].Type() != object.STRING_OBJ && args[0].Type() != object.BYTES_OBJ {
				return newPositionalTypeError("encode_base_64_32", 1, "STRING or BYTES", args[0].Type())
			}
			if args[1].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("encode_base_64_32", 2, object.BOOLEAN_OBJ, args[1].Type())
			}
			if args[2].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("encode_base_64_32", 3, object.BOOLEAN_OBJ, args[2].Type())
			}
			useBase64 := args[2].(*object.Boolean).Value
			var bs []byte
			if args[0].Type() == object.STRING_OBJ {
				bs = []byte(args[0].(*object.Stringo).Value)
			} else {
				bs = args[0].(*object.Bytes).Value
			}
			asBytes := args[1].(*object.Boolean).Value
			var encoded string
			if useBase64 {
				encoded = base64.StdEncoding.EncodeToString(bs)
			} else {
				encoded = base32.StdEncoding.EncodeToString(bs)
			}
			if asBytes {
				return &object.Bytes{Value: []byte(encoded)}
			}
			return &object.Stringo{Value: encoded}
		},
	},
	"_decode_base_64_32": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("decode_base_64_32", len(args), 3, "")
			}
			if args[0].Type() != object.STRING_OBJ && args[0].Type() != object.BYTES_OBJ {
				return newPositionalTypeError("decode_base_64_32", 1, "STRING or BYTES", args[0].Type())
			}
			if args[1].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("decode_base_64_32", 2, object.BOOLEAN_OBJ, args[1].Type())
			}
			if args[2].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("decode_base_64_32", 3, object.BOOLEAN_OBJ, args[2].Type())
			}
			useBase64 := args[2].(*object.Boolean).Value
			var s string
			if args[0].Type() == object.STRING_OBJ {
				s = args[0].(*object.Stringo).Value
			} else {
				s = string(args[0].(*object.Bytes).Value)
			}
			asBytes := args[1].(*object.Boolean).Value
			var decoded []byte
			var err error
			if useBase64 {
				decoded, err = base64.StdEncoding.DecodeString(s)
			} else {
				decoded, err = base32.StdEncoding.DecodeString(s)
			}
			if err != nil {
				return newError("`decode_base_64_32` error: %s", err.Error())
			}
			if !asBytes {
				return &object.Stringo{Value: string(decoded)}
			}
			return &object.Bytes{Value: decoded}
		},
	},
	"_decode_hex": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("decode_hex", len(args), 2, "")
			}
			if args[0].Type() != object.STRING_OBJ && args[0].Type() != object.BYTES_OBJ {
				return newPositionalTypeError("decode_hex", 1, "STRING or BYTES", args[0].Type())
			}
			if args[1].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("encode_hex", 2, object.BOOLEAN_OBJ, args[1].Type())
			}
			asBytes := args[1].(*object.Boolean).Value
			var bs []byte
			if args[0].Type() == object.STRING_OBJ {
				s := args[0].(*object.Stringo).Value
				data, err := hex.DecodeString(s)
				if err != nil {
					return newError("`decode_hex` error: %s", err.Error())
				}
				bs = data
			} else if args[0].Type() == object.BYTES_OBJ {
				b := args[0].(*object.Bytes).Value
				bs = make([]byte, hex.DecodedLen(len(b)))
				l, err := hex.Decode(bs, b)
				if err != nil {
					return newError("`decode_hex` error: %s", err.Error())
				}
				if l != len(b) {
					return newError("`decode_hex` error: length of bytes does not match bytes written. got=%d, want=%d", l, len(b))
				}
			}
			if !asBytes {
				return &object.Stringo{Value: string(bs)}
			}
			return &object.Bytes{Value: bs}
		},
	},
	"_encode_hex": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("encode_hex", len(args), 2, "")
			}
			if args[0].Type() != object.STRING_OBJ && args[0].Type() != object.BYTES_OBJ {
				return newPositionalTypeError("encode_hex", 1, "STRING or BYTES", args[0].Type())
			}
			if args[1].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("encode_hex", 2, object.BOOLEAN_OBJ, args[1].Type())
			}
			asBytes := args[1].(*object.Boolean).Value
			var s string
			if args[0].Type() == object.BYTES_OBJ {
				b := args[0].(*object.Bytes).Value
				s = hex.EncodeToString(b)
			} else if args[0].Type() == object.STRING_OBJ {
				b := args[0].(*object.Stringo).Value
				bs := make([]byte, hex.EncodedLen(len(b)))
				hex.Encode(bs, []byte(b))
				// if l != len(b) {
				// 	return newError("`encode_hex` error: length of bytes does not match bytes written. got=%d, want=%d", l, len(b))
				// }
				s = string(bs)
			}
			if asBytes {
				return &object.Bytes{Value: []byte(s)}
			}
			return &object.Stringo{Value: s}
		},
	},
})

var _net_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_connect": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("connect", len(args), 3, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("connect", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("connect", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.STRING_OBJ {
				return newPositionalTypeError("connect", 3, object.STRING_OBJ, args[2].Type())
			}
			transport := strings.ToLower(args[0].(*object.Stringo).Value)
			addr := args[1].(*object.Stringo).Value
			port := args[2].(*object.Stringo).Value
			addrStr := fmt.Sprintf("%s:%s", addr, port)
			conn, err := net.Dial(transport, addrStr)
			if err != nil {
				return newError("`connect` error: %s", err.Error())
			}
			return object.CreateBasicMapObjectForGoObj("net", &object.GoObj[net.Conn]{Value: conn, Id: GoObjId.Add(1)})
		},
	},
	"_listen": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("listen", len(args), 3, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("listen", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("listen", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.STRING_OBJ {
				return newPositionalTypeError("listen", 3, object.STRING_OBJ, args[2].Type())
			}
			transport := strings.ToLower(args[0].(*object.Stringo).Value)
			addr := args[1].(*object.Stringo).Value
			port := args[2].(*object.Stringo).Value
			addrStr := fmt.Sprintf("%s:%s", addr, port)
			if strings.Contains(transport, "udp") {
				s, err := net.ResolveUDPAddr(transport, ":"+port)
				if err != nil {
					return newError("`listen` udp error: %s", err.Error())
				}
				l, err := net.ListenUDP(transport, s)
				if err != nil {
					return newError("`listen` udp error: %s", err.Error())
				}
				return object.CreateBasicMapObjectForGoObj("net/udp", &object.GoObj[*net.UDPConn]{Value: l, Id: GoObjId.Add(1)})
			}
			l, err := net.Listen(transport, addrStr)
			if err != nil {
				return newError("`listen` error: %s", err.Error())
			}
			return object.CreateBasicMapObjectForGoObj("net/tcp", &object.GoObj[net.Listener]{Value: l, Id: GoObjId.Add(1)})
		},
	},
	"_accept": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("accept", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("accept", 1, object.GO_OBJ, args[0].Type())
			}
			l, ok := args[0].(*object.GoObj[net.Listener])
			if !ok {
				return newPositionalTypeErrorForGoObj("accept", 1, "net.Listener", l)
			}
			conn, err := l.Value.Accept()
			if err != nil {
				return newError("`accept` error: %s", err.Error())
			}
			return object.CreateBasicMapObjectForGoObj("net", &object.GoObj[net.Conn]{Value: conn, Id: GoObjId.Add(1)})
		},
	},
	"_net_close": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("net_close", len(args), 2, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("net_close", 1, object.GO_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("net_close", 2, object.STRING_OBJ, args[1].Type())
			}
			t := args[1].(*object.Stringo).Value
			switch t {
			case "net/udp":
				c, ok := args[0].(*object.GoObj[*net.UDPConn])
				if !ok {
					return newPositionalTypeErrorForGoObj("net_close", 1, "*net.UDPConn", c)
				}
				c.Value.Close()
			case "net/tcp":
				listener, ok := args[0].(*object.GoObj[net.Listener])
				if !ok {
					return newPositionalTypeErrorForGoObj("net_close", 1, "net.Listener", listener)
				}
				listener.Value.Close()
			case "net":
				conn, ok := args[0].(*object.GoObj[net.Conn])
				if !ok {
					return newPositionalTypeErrorForGoObj("net_close", 1, "net.Conn", conn)
				}
				conn.Value.Close()
			default:
				return newError("`net_close` expects type of 'net/tcp', 'net/udp', or 'net'")
			}
			return NULL
		},
	},
	"_net_read": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 4 {
				return newInvalidArgCountError("net_read", len(args), 4, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("net_read", 1, object.GO_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("net_read", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.NULL_OBJ && args[2].Type() != object.STRING_OBJ && args[2].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("net_read", 3, "NULL or STRING or INTEGER", args[2].Type())
			}
			if args[3].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("net_read", 4, object.BOOLEAN_OBJ, args[3].Type())
			}
			t := args[1].(*object.Stringo).Value
			var conn net.Conn
			if t == "net/udp" {
				c, ok := args[0].(*object.GoObj[*net.UDPConn])
				if !ok {
					return newPositionalTypeErrorForGoObj("net_read", 1, "*net.UDPConn", c)
				}
				conn = c.Value
			} else {
				c, ok := args[0].(*object.GoObj[net.Conn])
				if !ok {
					return newPositionalTypeErrorForGoObj("net_read", 1, "net.Conn", c)
				}
				conn = c.Value
			}
			if args[2].Type() == object.INTEGER_OBJ {
				asBytes := args[3].(*object.Boolean).Value
				bufLen := args[2].(*object.Integer).Value
				if bufLen == 0 {
					return newError("`net_read` error: len must not be 0")
				}
				buf := make([]byte, bufLen)
				readLen, err := bufio.NewReader(conn).Read(buf)
				if err != nil {
					return newError("`net_read` error: %s", err.Error())
				}
				if readLen != int(bufLen) {
					return newError("`net_read` error: read length (%d) does not match buffer length (%d)", readLen, bufLen)
				}
				if asBytes {
					return &object.Bytes{Value: buf}
				} else {
					return &object.Stringo{Value: string(buf)}
				}
			}
			var endByte byte
			if args[2].Type() == object.NULL_OBJ {
				endByte = '\n'
			} else {
				endByteString := args[2].(*object.Stringo).Value
				if len(endByteString) != 1 {
					return newError("`net_read` error: end byte given was not length 1, got=%d", len(endByteString))
				}
				endByte = []byte(endByteString)[0]
			}
			s, err := bufio.NewReader(conn).ReadString(endByte)
			if err != nil {
				return newError("`net_read` error: %s", err.Error())
			}
			return &object.Stringo{Value: s[:len(s)-1]}
		},
	},
	"_net_write": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("net_write", len(args), 3, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("net_write", 1, object.GO_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("net_write", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.STRING_OBJ && args[2].Type() != object.BYTES_OBJ {
				return newPositionalTypeError("net_write", 3, "STRING or BYTES", args[2].Type())
			}
			if args[3].Type() != object.NULL_OBJ && args[3].Type() != object.STRING_OBJ {
				return newPositionalTypeError("net_write", 4, "NULL or STRING", args[3].Type())
			}
			t := args[1].(*object.Stringo).Value
			var conn net.Conn
			if t == "net/udp" {
				c, ok := args[0].(*object.GoObj[*net.UDPConn])
				if !ok {
					return newPositionalTypeErrorForGoObj("net_write", 1, "*net.UDPConn", c)
				}
				conn = c.Value
			} else {
				c, ok := args[0].(*object.GoObj[net.Conn])
				if !ok {
					return newPositionalTypeErrorForGoObj("net_write", 1, "net.Conn", c)
				}
				conn = c.Value
			}
			var appendByte *byte = nil
			if args[3].Type() == object.STRING_OBJ {
				endByteString := args[3].(*object.Stringo).Value
				if len(endByteString) != 1 {
					return newError("`net_read` error: end byte given was not length 1, got=%d", len(endByteString))
				}
				appendByte = &[]byte(endByteString)[0]
			}
			buf := bytes.Buffer{}
			if args[2].Type() == object.STRING_OBJ {
				s := args[2].(*object.Stringo).Value
				n, err := buf.WriteString(s)
				if err != nil {
					return newError("`net_write` error: failed writing to internal buffer. %s", err.Error())
				}
				if n != len(s) {
					return newError("`net_write` error: failed writing string of len %d to internal buffer, wrote=%d", len(s), n)
				}
			} else {
				bs := args[2].(*object.Bytes).Value
				n, err := buf.Write(bs)
				if err != nil {
					return newError("`net_write` error: failed writing to internal buffer. %s", err.Error())
				}
				if n != len(bs) {
					return newError("`net_write` error: failed writing bytes of len %d to internal buffer, wrote=%d", len(bs), n)
				}
			}
			if appendByte != nil {
				err := buf.WriteByte(*appendByte)
				if err != nil {
					return newError("`net_write` error: failed writing end byte %#+v to internal buffer. %s", *appendByte, err.Error())
				}
			}
			bs := buf.Bytes()
			n, err := conn.Write(bs)
			if err != nil {
				return newError("`net_write` error: %s", err.Error())
			}
			if n != len(bs) {
				return newError("`net_write` error: did not write the entire length got=%d, want=%d", n, len(bs))
			}
			return NULL
		},
	},
	"_inspect": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("inspect", len(args), 2, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("inspect", 1, object.GO_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("inspect", 2, object.STRING_OBJ, args[1].Type())
			}
			t := args[1].(*object.Stringo).Value
			switch t {
			case "net/udp":
				c, ok := args[0].(*object.GoObj[*net.UDPConn])
				if !ok {
					return newPositionalTypeErrorForGoObj("inspect", 1, "*net.UDPConn", c)
				}
				mapObj := object.NewOrderedMap[string, object.Object]()
				mapObj.Set("remote_addr", &object.Stringo{Value: c.Value.RemoteAddr().String()})
				mapObj.Set("local_addr", &object.Stringo{Value: c.Value.LocalAddr().String()})
				mapObj.Set("remote_addr_network", &object.Stringo{Value: c.Value.RemoteAddr().Network()})
				mapObj.Set("local_addr_network", &object.Stringo{Value: c.Value.LocalAddr().Network()})
				return object.CreateMapObjectForGoMap(*mapObj)
			case "net/tcp":
				l, ok := args[0].(*object.GoObj[net.Listener])
				if !ok {
					return newPositionalTypeErrorForGoObj("inspect", 1, "net.Listener", l)
				}
				mapObj := object.NewOrderedMap[string, object.Object]()
				mapObj.Set("addr", &object.Stringo{Value: l.Value.Addr().String()})
				mapObj.Set("addr_network", &object.Stringo{Value: l.Value.Addr().Network()})
				return object.CreateMapObjectForGoMap(*mapObj)
			case "net":
				c, ok := args[0].(*object.GoObj[net.Conn])
				if !ok {
					return newPositionalTypeErrorForGoObj("inspect", 1, "net.Conn", c)
				}
				mapObj := object.NewOrderedMap[string, object.Object]()
				mapObj.Set("remote_addr", &object.Stringo{Value: c.Value.RemoteAddr().String()})
				mapObj.Set("local_addr", &object.Stringo{Value: c.Value.LocalAddr().String()})
				mapObj.Set("remote_addr_network", &object.Stringo{Value: c.Value.RemoteAddr().Network()})
				mapObj.Set("local_addr_network", &object.Stringo{Value: c.Value.LocalAddr().Network()})
				return object.CreateMapObjectForGoMap(*mapObj)
			default:
				return newError("`inspect` expects type of 'net/tcp', 'net/udp', or 'net'")
			}
		},
	},
})

var _ui_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_new_app": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("new_app", len(args), 0, "")
			}
			app := app.New()
			return &object.GoObj[fyne.App]{Value: app, Id: GoObjId.Add(1)}
		},
	},
	"_window": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 5 {
				return newInvalidArgCountError("window", len(args), 5, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("window", 1, object.GO_OBJ, args[0].Type())
			}
			app, ok := args[0].(*object.GoObj[fyne.App])
			if !ok {
				return newPositionalTypeErrorForGoObj("window", 1, "fyne.App", app)
			}
			if args[1].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("window", 2, object.INTEGER_OBJ, args[1].Type())
			}
			if args[2].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("window", 3, object.INTEGER_OBJ, args[2].Type())
			}
			if args[3].Type() != object.STRING_OBJ {
				return newPositionalTypeError("window", 4, object.STRING_OBJ, args[3].Type())
			}
			if args[4].Type() != object.GO_OBJ {
				return newPositionalTypeError("window", 5, object.GO_OBJ, args[4].Type())
			}
			content, ok := args[4].(*object.GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("window", 4, "fyne.CanvasObject", app)
			}
			width := args[1].(*object.Integer).Value
			height := args[2].(*object.Integer).Value
			title := args[3].(*object.Stringo).Value
			w := app.Value.NewWindow(title)
			w.Resize(fyne.Size{Width: float32(width), Height: float32(height)})
			w.SetContent(content.Value)
			w.ShowAndRun()
			return NULL
		},
	},
	"_label": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("label", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("label", 1, object.STRING_OBJ, args[0].Type())
			}
			label := args[0].(*object.Stringo).Value
			l := widget.NewLabel(label)
			return object.CreateBasicMapObjectForGoObj("ui", &object.GoObj[fyne.CanvasObject]{Value: l, Id: GoObjId.Add(1)})
		},
	},
	"_row": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("row", len(args), 1, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("row", 1, object.LIST_OBJ, args[0].Type())
			}
			elements := args[0].(*object.List).Elements
			canvasObjects := make([]fyne.CanvasObject, len(elements))
			for i, e := range elements {
				if e.Type() != object.GO_OBJ {
					return newError("`row` error: all children should be GO_OBJ[fyne.CanvasObject]. found=%s", e.Type())
				}
				o, ok := e.(*object.GoObj[fyne.CanvasObject])
				if !ok {
					return newPositionalTypeErrorForGoObj("row", i+1, "fyne.CanvasObject", o)
				}
				canvasObjects[i] = o.Value
			}
			vbox := container.NewVBox(canvasObjects...)
			return object.CreateBasicMapObjectForGoObj("ui", &object.GoObj[fyne.CanvasObject]{Value: vbox, Id: GoObjId.Add(1)})
		},
	},
	"_col": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("col", len(args), 1, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("col", 1, object.LIST_OBJ, args[0].Type())
			}
			elements := args[0].(*object.List).Elements
			canvasObjects := make([]fyne.CanvasObject, len(elements))
			for i, e := range elements {
				if e.Type() != object.GO_OBJ {
					return newError("`col` error: all children should be GO_OBJ[fyne.CanvasObject]. found=%s", e.Type())
				}
				o, ok := e.(*object.GoObj[fyne.CanvasObject])
				if !ok {
					return newPositionalTypeErrorForGoObj("col", i+1, "fyne.CanvasObject", o)
				}
				canvasObjects[i] = o.Value
			}
			hbox := container.NewHBox(canvasObjects...)
			return object.CreateBasicMapObjectForGoObj("ui", &object.GoObj[fyne.CanvasObject]{Value: hbox, Id: GoObjId.Add(1)})
		},
	},
	"_grid": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("grid", len(args), 2, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("grid", 1, object.INTEGER_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("grid", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.LIST_OBJ {
				return newPositionalTypeError("grid", 3, object.LIST_OBJ, args[2].Type())
			}
			rowsOrCols := int(args[0].(*object.Integer).Value)
			gridType := args[1].(*object.Stringo).Value
			if gridType != "COLS" && gridType != "ROWS" {
				return newError("`grid` error: type must be COLS or ROWS. got=%s", gridType)
			}
			elements := args[2].(*object.List).Elements
			canvasObjects := make([]fyne.CanvasObject, len(elements))
			for i, e := range elements {
				if e.Type() != object.GO_OBJ {
					return newError("`grid` error: all children should be GO_OBJ[fyne.CanvasObject]. found=%s", e.Type())
				}
				o, ok := e.(*object.GoObj[fyne.CanvasObject])
				if !ok {
					return newPositionalTypeErrorForGoObj("grid", i+1, "fyne.CanvasObject", o)
				}
				canvasObjects[i] = o.Value
			}
			var grid *fyne.Container
			if gridType == "ROWS" {
				grid = container.NewGridWithRows(rowsOrCols, canvasObjects...)
			} else {
				grid = container.NewGridWithColumns(rowsOrCols, canvasObjects...)
			}
			return object.CreateBasicMapObjectForGoObj("ui", &object.GoObj[fyne.CanvasObject]{Value: grid, Id: GoObjId.Add(1)})
		},
	},
	"_entry": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("entry", len(args), 1, "")
			}
			if args[0].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("entry", 1, object.BOOLEAN_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("entry", 2, object.STRING_OBJ, args[1].Type())
			}
			isMultiline := args[0].(*object.Boolean).Value
			placeholderText := args[1].(*object.Stringo).Value
			var entry *widget.Entry
			if isMultiline {
				entry = widget.NewMultiLineEntry()
			} else {
				entry = widget.NewEntry()
			}
			entry.SetPlaceHolder(placeholderText)
			return object.CreateBasicMapObjectForGoObj("ui/entry", &object.GoObj[fyne.CanvasObject]{Value: entry, Id: GoObjId.Add(1)})
		},
	},
	"_entry_get_text": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("entry_get_text", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("entry_get_text", 1, object.GO_OBJ, args[0].Type())
			}
			entry, ok := args[0].(*object.GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("entry_get_text", 1, "fyne.CanvasObject", entry)
			}
			switch x := entry.Value.(type) {
			case *widget.Entry:
				return &object.Stringo{Value: x.Text}
			default:
				return newError("`entry_get_text` error: entry id did not match entry. got=%T", x)
			}
		},
	},
	"_entry_set_text": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("entry_set_text", len(args), 2, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("entry_set_text", 1, object.GO_OBJ, args[0].Type())
			}
			entry, ok := args[0].(*object.GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("entry_set_text", 1, "fyne.CanvasObject", entry)
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("entry_set_text", 2, object.STRING_OBJ, args[1].Type())
			}
			value := args[1].(*object.Stringo).Value
			switch x := entry.Value.(type) {
			case *widget.Entry:
				x.SetText(value)
				return NULL
			default:
				return newError("`entry_set_text` error: entry id did not match entry. got=%T", x)
			}
		},
	},
	"_append_form": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("append_form", len(args), 3, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("append_form", 1, object.GO_OBJ, args[0].Type())
			}
			maybeForm, ok := args[0].(*object.GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("append_form", 1, "fyne.CanvasObject", maybeForm)
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("append_form", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.GO_OBJ {
				return newPositionalTypeError("append_form", 3, object.GO_OBJ, args[2].Type())
			}
			w, ok := args[3].(*object.GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("append_form", 3, "fyne.CanvasObject", w)
			}
			var form *widget.Form
			switch x := maybeForm.Value.(type) {
			case *widget.Form:
				form = x
			default:
				return newError("`append_form` error: id used for form is not form. got=%T", x)
			}
			form.Append(args[1].(*object.Stringo).Value, w.Value)
			return NULL
		},
	},
})

var _color_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_style": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("style", len(args), 3, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("style", 1, object.INTEGER_OBJ, args[0].Type())
			}
			if args[1].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("style", 2, object.INTEGER_OBJ, args[1].Type())
			}
			if args[2].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("style", 3, object.INTEGER_OBJ, args[2].Type())
			}
			arg1, arg2, arg3 := args[0].(*object.Integer).Value, args[1].(*object.Integer).Value, args[2].(*object.Integer).Value
			textStyle := color.Color(arg1)
			fgActualColor := color.Color(arg2)
			fgColor := fgActualColor.ToFg()
			bgActualColor := color.Color(arg3)
			bgColor := bgActualColor.ToBg()
			textStyleName := textStyle.Name()
			fgColorName := fgColor.Name()
			bgColorName := bgColor.Name()
			fgActualColorName := fgActualColor.Name()
			bgActualColorName := bgActualColor.Name()
			s := color.New()
			unknown := "unknown"
			if textStyleName != unknown {
				s.Add(textStyle)
			}
			if fgColorName != unknown || fgActualColorName != unknown {
				s.Add(fgColor)
			}
			if bgColorName != unknown || bgActualColorName != unknown {
				s.Add(bgColor)
			}
			return object.CreateBasicMapObjectForGoObj("color", &object.GoObj[color.Style]{Value: s, Id: GoObjId.Add(1)})
		},
	},
	"_normal": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("normal", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.Normal)}
		},
	},
	"_red": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("red", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.Red)}
		},
	},
	"_cyan": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("cyan", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.Cyan)}
		},
	},
	"_gray": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("gray", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.Gray)}
		},
	},
	"_blue": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("blue", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.Blue)}
		},
	},
	"_black": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("black", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.Black)}
		},
	},
	"_green": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("green", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.Green)}
		},
	},
	"_white": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("white", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.White)}
		},
	},
	"_yellow": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("yellow", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.Yellow)}
		},
	},
	"_magenta": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("magenta", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.Magenta)}
		},
	},
	"_bold": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("bold", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.Bold)}
		},
	},
	"_italic": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("italic", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.OpItalic)}
		},
	},
	"_underlined": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("underlined", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.OpUnderscore)}
		},
	},
})

var _csv_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_parse": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 6 {
				return newInvalidArgCountError("parse", len(args), 6, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("parse", 1, object.STRING_OBJ, args[0].Type())
			}
			// parse(data, delimeter=',', named_fields=false, comment=null, lazy_quotes=false, trim_leading_space=false) {
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("parse", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("parse", 3, object.BOOLEAN_OBJ, args[2].Type())
			}
			if args[3].Type() != object.NULL_OBJ && args[3].Type() != object.STRING_OBJ {
				return newPositionalTypeError("parse", 4, "NULL or STRING", args[3].Type())
			}
			if args[4].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("parse", 5, object.BOOLEAN_OBJ, args[4].Type())
			}
			if args[5].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("parse", 6, object.BOOLEAN_OBJ, args[5].Type())
			}
			data := args[0].(*object.Stringo).Value
			delimeter := args[1].(*object.Stringo).Value
			namedFields := args[2].(*object.Boolean).Value
			useComment := false
			var comment rune
			if args[3].Type() == object.NULL_OBJ {
				useComment = true
			} else {
				c := args[3].(*object.Stringo).Value
				if runeLen(c) != 1 {
					return newError("parse error: comment length is not 1. got=%d '%s'", runeLen(c), c)
				}
				comment = []rune(c)[0]
			}
			lazyQuotes := args[4].(*object.Boolean).Value
			trimLeadingSpace := args[5].(*object.Boolean).Value
			if runeLen(delimeter) != 1 {
				return newError("parse error: delimeter length is not 1. got=%d '%s'", runeLen(delimeter), delimeter)
			}
			dRune := []rune(delimeter)[0]

			reader := csv.NewReader(strings.NewReader(data))
			reader.Comma = dRune
			if useComment {
				reader.Comment = comment
			}
			reader.LazyQuotes = lazyQuotes
			reader.TrimLeadingSpace = trimLeadingSpace

			rows, err := reader.ReadAll()
			if err != nil {
				return newError("parse error: %s", err.Error())
			}
			if !namedFields {
				// Here we are just returning a list of lists
				allRows := &object.List{
					Elements: make([]object.Object, len(rows)),
				}
				for i, row := range rows {
					rowList := &object.List{
						Elements: make([]object.Object, len(row)),
					}
					for j, e := range row {
						rowList.Elements[j] = &object.Stringo{Value: e}
					}
					allRows.Elements[i] = rowList
				}
				return allRows
			}

			if len(rows) < 1 {
				return newError("parse error: named fields requires at least 1 row in the csv to act as the header")
			}
			headerRow := rows[0]
			rows = rows[1:]
			allRows := &object.List{
				Elements: make([]object.Object, len(rows)),
			}
			for i, row := range rows {
				if len(row) != len(headerRow) {
					return newError("parse error: row length did not match header row length. got=%d, want=%d", len(row), len(headerRow))
				}
				m := object.NewOrderedMap[string, object.Object]()
				for i, v := range row {
					m.Set(headerRow[i], &object.Stringo{Value: v})
				}
				allRows.Elements[i] = object.CreateMapObjectForGoMap(*m)
			}
			return allRows
		},
	},
	"_dump": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("dump", len(args), 3, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("dump", 1, object.LIST_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("dump", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("dump", 3, object.BOOLEAN_OBJ, args[2].Type())
			}
			l := args[0].(*object.List).Elements
			comma := args[1].(*object.Stringo).Value
			if runeLen(comma) != 1 {
				return newError("dump error: comma needs to be 1 character long. got=%d", runeLen(comma))
			}
			c := []rune(comma)[0]
			useCrlf := args[2].(*object.Boolean).Value
			if len(l) < 1 {
				return newError("dump error: list was empty. got=%d", len(l))
			}
			if l[0].Type() != object.MAP_OBJ && l[0].Type() != object.LIST_OBJ {
				return newError("dump error: list should be a list of maps, or list of lists. got=%s", l[0].Type())
			}
			offset := 0
			if l[0].Type() == object.MAP_OBJ {
				// Account for headers
				offset = 1
			}
			allRows := make([][]string, len(l)+offset)

			// checking types and info
			if l[0].Type() == object.MAP_OBJ {
				var keys []object.HashKey
				for i, e := range l {
					if e.Type() != object.MAP_OBJ {
						return newError("dump error: invalid data. for rows that should be MAPs, found %s", e.Type())
					}
					// Validate that all the keys are at least the same - then we can use inspect
					// to get the actual keys and also use inspect for all the values
					// May just want to use a separate loops
					mps := e.(*object.Map).Pairs
					if keys == nil && i == 0 {
						keys = append(keys, mps.Keys...)
						for _, k := range mps.Keys {
							mp, _ := mps.Get(k)
							// This is for the headers
							allRows[i] = append(allRows[i], mp.Key.Inspect())
							allRows[i+offset] = append(allRows[i+offset], mp.Value.Inspect())
						}
					} else {
						if len(keys) != len(mps.Keys) {
							return newError("dump error: invalid data. found a row where number of keys did not match")
						}
						for j, k := range mps.Keys {
							if keys[j] != k {
								return newError("dump error: invalid data. found a row where the key at a certain position did not match the expected")
							}
							mp, _ := mps.Get(k)
							allRows[i+offset] = append(allRows[i+offset], mp.Value.Inspect())
						}
					}
				}
			} else {
				for i, e := range l {
					if e.Type() != object.LIST_OBJ {
						return newError("dump error: invalid data. for rows that should be LISTs, found %s", e.Type())
					}
					rowL := e.(*object.List).Elements
					for _, elem := range rowL {
						// No offset should be needed here (but if we added it, it would just be 0)
						allRows[i] = append(allRows[i], elem.Inspect())
					}
				}
			}
			sb := &strings.Builder{}
			w := csv.NewWriter(sb)
			w.Comma = c
			w.UseCRLF = useCrlf
			err := w.WriteAll(allRows)
			if err != nil {
				return newError("dump error: csv writer error: %s", err.Error())
			}
			return &object.Stringo{Value: sb.String()}
		},
	},
})

var _psutil_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_cpu_usage_percent": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("cpu_usage_percent", len(args), 0, "")
			}
			usages, err := cpu.Percent(0, true)
			if err != nil {
				return newError("`cpu_usage_percent` error: %s", err.Error())
			}
			l := &object.List{Elements: make([]object.Object, len(usages))}
			for i, v := range usages {
				l.Elements[i] = &object.Float{Value: v}
			}
			return l
		},
	},
	"_cpu_info": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("cpu_info", len(args), 0, "")
			}
			infos, err := cpu.Info()
			if err != nil {
				return newError("`cpu_info` error: %s", err.Error())
			}
			l := &object.List{Elements: make([]object.Object, len(infos))}
			for i, v := range infos {
				l.Elements[i] = &object.Stringo{Value: v.String()}
			}
			return l
		},
	},
	"_cpu_time_info": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("cpu_time_info", len(args), 0, "")
			}
			infos, err := cpu.Times(true)
			if err != nil {
				return newError("`cpu_time_info` error: %s", err.Error())
			}
			l := &object.List{Elements: make([]object.Object, len(infos))}
			for i, v := range infos {
				l.Elements[i] = &object.Stringo{Value: v.String()}
			}
			return l
		},
	},
	"_cpu_count": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("cpu_count", len(args), 0, "")
			}
			count, err := cpu.Counts(true)
			if err != nil {
				return newError("`cpu_count` error: %s", err.Error())
			}
			return &object.Integer{Value: int64(count)}
		},
	},
	"_mem_virt_info": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("mem_virt_info", len(args), 0, "")
			}
			v, err := mem.VirtualMemory()
			if err != nil {
				return newError("`mem_virt_info` error: %s", err.Error())
			}
			return &object.Stringo{Value: v.String()}
		},
	},
	"_mem_swap_info": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("mem_swap_info", len(args), 0, "")
			}
			v, err := mem.SwapMemory()
			if err != nil {
				return newError("`mem_swap_info` error: %s", err.Error())
			}
			return &object.Stringo{Value: v.String()}
		},
	},
	"_host_info": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("host_info", len(args), 0, "")
			}
			i, err := host.Info()
			if err != nil {
				return newError("`host_info` error: %s", err.Error())
			}
			return &object.Stringo{Value: i.String()}
		},
	},
	"_host_temps_info": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("host_temps_info", len(args), 0, "")
			}
			temps, err := host.SensorsTemperatures()
			if err != nil {
				if !strings.Contains(err.Error(), "warnings") {
					return newError("`host_temps_info` error: %s", err.Error())
				}
			}
			l := &object.List{Elements: make([]object.Object, len(temps))}
			for i, t := range temps {
				l.Elements[i] = &object.Stringo{Value: t.String()}
			}
			return l
		},
	},
	"_net_connections": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("net_connections", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("net_connections", 1, object.STRING_OBJ, args[0].Type())
			}
			option := args[0].(*object.Stringo).Value
			conns, err := psutilnet.Connections(option)
			if err != nil {
				return newError("`net_connections` error: %s", err.Error())
			}
			l := &object.List{Elements: make([]object.Object, len(conns))}
			for i, c := range conns {
				l.Elements[i] = &object.Stringo{Value: c.String()}
			}
			return l
		},
	},
	"_net_io_info": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("net_io_info", len(args), 0, "")
			}
			ioc, err := psutilnet.IOCounters(true)
			if err != nil {
				return newError("`net_io_info` error: %s", err.Error())
			}
			l := &object.List{Elements: make([]object.Object, len(ioc))}
			for i, oc := range ioc {
				l.Elements[i] = &object.Stringo{Value: oc.String()}
			}
			return l
		},
	},
	"_disk_partitions": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("disk_partitions", len(args), 0, "")
			}
			parts, err := disk.Partitions(true)
			if err != nil {
				return newError("`disk_partitions` error: %s", err.Error())
			}
			l := &object.List{Elements: make([]object.Object, len(parts))}
			for i, p := range parts {
				l.Elements[i] = &object.Stringo{Value: p.String()}
			}
			return l
		},
	},
	"_disk_io_info": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("disk_io_info", len(args), 0, "")
			}
			ioc, err := disk.IOCounters()
			if err != nil {
				return newError("`disk_io_info` error: %s", err.Error())
			}
			m := object.NewOrderedMap[string, object.Object]()
			for k, v := range ioc {
				m.Set(k, &object.Stringo{Value: v.String()})
			}
			return object.CreateMapObjectForGoMap(*m)
		},
	},
	"_disk_usage": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("disk_usage", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("disk_usage", 1, object.STRING_OBJ, args[0].Type())
			}
			path := args[0].(*object.Stringo).Value
			usage, err := disk.Usage(path)
			if err != nil {
				return newError("`disk_usage` error: %s", err.Error())
			}
			return &object.Stringo{Value: usage.String()}
		},
	},
})

var _gg_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_init_window": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("init_window", len(args), 3, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("init_window", 1, object.INTEGER_OBJ, args[0].Type())
			}
			if args[1].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("init_window", 2, object.INTEGER_OBJ, args[1].Type())
			}
			if args[2].Type() != object.STRING_OBJ {
				return newPositionalTypeError("init_window", 3, object.STRING_OBJ, args[2].Type())
			}
			width := int32(args[0].(*object.Integer).Value)
			height := int32(args[1].(*object.Integer).Value)
			title := args[2].(*object.Stringo).Value
			rl.InitWindow(width, height, title)
			return NULL
		},
	},
	"_close_window": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("close_window", len(args), 0, "")
			}
			rl.CloseWindow()
			return NULL
		},
	},
	"_window_should_close": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("window_should_close", len(args), 0, "")
			}
			return nativeToBooleanObject(rl.WindowShouldClose())
		},
	},
	"_begin_drawing": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("begin_drawing", len(args), 0, "")
			}
			rl.BeginDrawing()
			return NULL
		},
	},
	"_end_drawing": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("end_drawing", len(args), 0, "")
			}
			rl.EndDrawing()
			return NULL
		},
	},
	"_clear_background": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("clear_background", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("clear_background", 1, object.GO_OBJ, args[0].Type())
			}
			goObj, ok := args[0].(*object.GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("clear_background", 1, "rl.Color", args[0])
			}
			rl.ClearBackground(goObj.Value)
			return NULL
		},
	},
	"_color_map": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("color_map", len(args), 0, "")
			}
			mapObj := object.NewOrderedMap[string, object.Object]()
			lightGray := &object.GoObj[rl.Color]{Value: rl.LightGray, Id: GoObjId.Add(1)}
			gray := &object.GoObj[rl.Color]{Value: rl.Gray, Id: GoObjId.Add(1)}
			darkGray := &object.GoObj[rl.Color]{Value: rl.DarkGray, Id: GoObjId.Add(1)}
			yellow := &object.GoObj[rl.Color]{Value: rl.Yellow, Id: GoObjId.Add(1)}
			gold := &object.GoObj[rl.Color]{Value: rl.Gold, Id: GoObjId.Add(1)}
			orange := &object.GoObj[rl.Color]{Value: rl.Orange, Id: GoObjId.Add(1)}
			pink := &object.GoObj[rl.Color]{Value: rl.Pink, Id: GoObjId.Add(1)}
			red := &object.GoObj[rl.Color]{Value: rl.Red, Id: GoObjId.Add(1)}
			maroon := &object.GoObj[rl.Color]{Value: rl.Maroon, Id: GoObjId.Add(1)}
			green := &object.GoObj[rl.Color]{Value: rl.Green, Id: GoObjId.Add(1)}
			lime := &object.GoObj[rl.Color]{Value: rl.Lime, Id: GoObjId.Add(1)}
			darkGreen := &object.GoObj[rl.Color]{Value: rl.DarkGreen, Id: GoObjId.Add(1)}
			skyBlue := &object.GoObj[rl.Color]{Value: rl.SkyBlue, Id: GoObjId.Add(1)}
			blue := &object.GoObj[rl.Color]{Value: rl.Blue, Id: GoObjId.Add(1)}
			darkBlue := &object.GoObj[rl.Color]{Value: rl.DarkBlue, Id: GoObjId.Add(1)}
			purple := &object.GoObj[rl.Color]{Value: rl.Purple, Id: GoObjId.Add(1)}
			violet := &object.GoObj[rl.Color]{Value: rl.Violet, Id: GoObjId.Add(1)}
			darkPurple := &object.GoObj[rl.Color]{Value: rl.DarkPurple, Id: GoObjId.Add(1)}
			beige := &object.GoObj[rl.Color]{Value: rl.Beige, Id: GoObjId.Add(1)}
			brown := &object.GoObj[rl.Color]{Value: rl.Brown, Id: GoObjId.Add(1)}
			darkBrown := &object.GoObj[rl.Color]{Value: rl.DarkBrown, Id: GoObjId.Add(1)}
			white := &object.GoObj[rl.Color]{Value: rl.White, Id: GoObjId.Add(1)}
			black := &object.GoObj[rl.Color]{Value: rl.Black, Id: GoObjId.Add(1)}
			blank := &object.GoObj[rl.Color]{Value: rl.Blank, Id: GoObjId.Add(1)}
			magenta := &object.GoObj[rl.Color]{Value: rl.Magenta, Id: GoObjId.Add(1)}
			rayWhite := &object.GoObj[rl.Color]{Value: rl.RayWhite, Id: GoObjId.Add(1)}
			newColor := &object.Builtin{
				Fun: func(args ...object.Object) object.Object {
					if len(args) != 4 {
						return newInvalidArgCountError("new_color", len(args), 4, "")
					}
					if args[0].Type() != object.INTEGER_OBJ {
						return newPositionalTypeError("new_color", 1, object.INTEGER_OBJ, args[0].Type())
					}
					if args[1].Type() != object.INTEGER_OBJ {
						return newPositionalTypeError("new_color", 2, object.INTEGER_OBJ, args[1].Type())
					}
					if args[2].Type() != object.INTEGER_OBJ {
						return newPositionalTypeError("new_color", 3, object.INTEGER_OBJ, args[2].Type())
					}
					if args[3].Type() != object.INTEGER_OBJ {
						return newPositionalTypeError("new_color", 4, object.INTEGER_OBJ, args[3].Type())
					}
					return &object.GoObj[rl.Color]{
						Value: rl.NewColor(
							uint8(args[0].(*object.Integer).Value),
							uint8(args[1].(*object.Integer).Value),
							uint8(args[2].(*object.Integer).Value),
							uint8(args[3].(*object.Integer).Value)),
						Id: GoObjId.Add(1),
					}
				},
			}
			mapObj.Set("light_gray", lightGray)
			mapObj.Set("gray", gray)
			mapObj.Set("dark_gray", darkGray)
			mapObj.Set("yellow", yellow)
			mapObj.Set("gold", gold)
			mapObj.Set("orange", orange)
			mapObj.Set("pink", pink)
			mapObj.Set("red", red)
			mapObj.Set("maroon", maroon)
			mapObj.Set("green", green)
			mapObj.Set("lime", lime)
			mapObj.Set("dark_green", darkGreen)
			mapObj.Set("sky_blue", skyBlue)
			mapObj.Set("blue", blue)
			mapObj.Set("dark_blue", darkBlue)
			mapObj.Set("purple", purple)
			mapObj.Set("violet", violet)
			mapObj.Set("dark_purple", darkPurple)
			mapObj.Set("beige", beige)
			mapObj.Set("brown", brown)
			mapObj.Set("dark_brown", darkBrown)
			mapObj.Set("white", white)
			mapObj.Set("black", black)
			mapObj.Set("blank", blank)
			mapObj.Set("magenta", magenta)
			mapObj.Set("ray_white", rayWhite)
			mapObj.Set("new", newColor)
			return object.CreateMapObjectForGoMap(*mapObj)
		},
	},
	"_draw_text": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 5 {
				return newInvalidArgCountError("draw_text", len(args), 5, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("draw_text", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("draw_text", 2, object.INTEGER_OBJ, args[1].Type())
			}
			if args[2].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("draw_text", 3, object.INTEGER_OBJ, args[2].Type())
			}
			if args[3].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("draw_text", 4, object.INTEGER_OBJ, args[3].Type())
			}
			if args[4].Type() != object.GO_OBJ {
				return newPositionalTypeError("draw_text", 5, object.GO_OBJ, args[4].Type())
			}
			goObj, ok := args[4].(*object.GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_text", 5, "rl.Color", args[4])
			}
			text := args[0].(*object.Stringo).Value
			posX := int32(args[1].(*object.Integer).Value)
			posY := int32(args[2].(*object.Integer).Value)
			fontSize := int32(args[3].(*object.Integer).Value)
			rl.DrawText(text, posX, posY, fontSize, goObj.Value)
			return NULL
		},
	},
	"_draw_texture": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 4 {
				return newInvalidArgCountError("draw_texture", len(args), 4, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("draw_texture", 1, object.GO_OBJ, args[0].Type())
			}
			tex, ok := args[0].(*object.GoObj[rl.Texture2D])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_texture", 1, "rl.Texture2D", tex)
			}
			if args[1].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("draw_texture", 2, object.INTEGER_OBJ, args[1].Type())
			}
			if args[2].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("draw_texture", 3, object.INTEGER_OBJ, args[2].Type())
			}
			if args[3].Type() != object.GO_OBJ {
				return newPositionalTypeError("draw_texture", 4, object.GO_OBJ, args[3].Type())
			}
			tint, ok := args[3].(*object.GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_texture", 4, "rl.Color", tint)
			}
			posX := int32(args[1].(*object.Integer).Value)
			posY := int32(args[2].(*object.Integer).Value)
			rl.DrawTexture(tex.Value, posX, posY, tint.Value)
			return NULL
		},
	},
	"_draw_texture_pro": {
		Fun: func(args ...object.Object) object.Object {
			// draw_texture_pro(texture: GO_OBJ[rl.Texture2D], source_rec: GO_OBJ[rl.Rectangle]=Rectangle(), dest_rec: GO_OBJ[rl.Rectangle]=Rectangle(), origin: GO_OBJ[rl.Vector2]=Vector2(), rotation: float=0.0, tint=color.white)
			if len(args) != 6 {
				return newInvalidArgCountError("draw_texture_pro", len(args), 4, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("draw_texture_pro", 1, object.GO_OBJ, args[0].Type())
			}
			tex, ok := args[0].(*object.GoObj[rl.Texture2D])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_texture_pro", 1, "rl.Texture2D", tex)
			}
			if args[1].Type() != object.GO_OBJ {
				return newPositionalTypeError("draw_texture_pro", 2, object.GO_OBJ, args[1].Type())
			}
			srcRect, ok := args[1].(*object.GoObj[rl.Rectangle])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_texture_pro", 2, "rl.Rectangle", srcRect)
			}
			if args[2].Type() != object.GO_OBJ {
				return newPositionalTypeError("draw_texture_pro", 3, object.GO_OBJ, args[2].Type())
			}
			dstRect, ok := args[2].(*object.GoObj[rl.Rectangle])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_texture_pro", 3, "rl.Rectangle", dstRect)
			}
			if args[3].Type() != object.GO_OBJ {
				return newPositionalTypeError("draw_texture_pro", 4, object.GO_OBJ, args[3].Type())
			}
			origin, ok := args[3].(*object.GoObj[rl.Vector2])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_texture_pro", 4, "rl.Rectangle", origin)
			}
			if args[4].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("draw_texture_pro", 5, object.FLOAT_OBJ, args[4].Type())
			}
			rotation := float32(args[4].(*object.Float).Value)
			if args[5].Type() != object.GO_OBJ {
				return newPositionalTypeError("draw_texture_pro", 6, object.GO_OBJ, args[5].Type())
			}
			tint, ok := args[5].(*object.GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_texture_pro", 6, "rl.Color", tint)
			}
			rl.DrawTexturePro(tex.Value, srcRect.Value, dstRect.Value, origin.Value, rotation, tint.Value)
			return NULL
		},
	},
	"_set_target_fps": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("set_target_fps", len(args), 1, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("set_target_fps", 1, object.INTEGER_OBJ, args[0].Type())
			}
			fps := int32(args[0].(*object.Integer).Value)
			rl.SetTargetFPS(fps)
			return NULL
		},
	},
	"_set_exit_key": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("set_exit_key", len(args), 1, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("set_exit_key", 1, object.INTEGER_OBJ, args[0].Type())
			}
			key := int32(args[0].(*object.Integer).Value)
			rl.SetExitKey(key)
			return NULL
		},
	},
	"_is_key_up": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("is_key_up", len(args), 1, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("is_key_up", 1, object.INTEGER_OBJ, args[0].Type())
			}
			key := int32(args[0].(*object.Integer).Value)
			return nativeToBooleanObject(rl.IsKeyUp(key))
		},
	},
	"_is_key_down": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("is_key_down", len(args), 1, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("is_key_down", 1, object.INTEGER_OBJ, args[0].Type())
			}
			key := int32(args[0].(*object.Integer).Value)
			return nativeToBooleanObject(rl.IsKeyDown(key))
		},
	},
	"_is_key_pressed": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("is_key_pressed", len(args), 1, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("is_key_pressed", 1, object.INTEGER_OBJ, args[0].Type())
			}
			key := int32(args[0].(*object.Integer).Value)
			return nativeToBooleanObject(rl.IsKeyPressed(key))
		},
	},
	"_is_key_released": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("is_key_released", len(args), 1, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("is_key_released", 1, object.INTEGER_OBJ, args[0].Type())
			}
			key := int32(args[0].(*object.Integer).Value)
			return nativeToBooleanObject(rl.IsKeyReleased(key))
		},
	},
	"_load_texture": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("load_texture", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("load_texture", 1, object.STRING_OBJ, args[0].Type())
			}
			fname := args[0].(*object.Stringo).Value
			return &object.GoObj[rl.Texture2D]{Value: rl.LoadTexture(fname), Id: GoObjId.Add(1)}
		},
	},
	"_rectangle": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 4 {
				return newInvalidArgCountError("Rectangle", len(args), 4, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Rectangle", 1, object.FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Rectangle", 2, object.FLOAT_OBJ, args[1].Type())
			}
			if args[2].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Rectangle", 3, object.FLOAT_OBJ, args[2].Type())
			}
			if args[3].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Rectangle", 4, object.FLOAT_OBJ, args[3].Type())
			}
			x := float32(args[0].(*object.Float).Value)
			y := float32(args[1].(*object.Float).Value)
			width := float32(args[2].(*object.Float).Value)
			height := float32(args[3].(*object.Float).Value)
			return &object.GoObj[rl.Rectangle]{Value: rl.NewRectangle(x, y, width, height), Id: GoObjId.Add(1)}
		},
	},
	"_vector2": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("Vector2", len(args), 2, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Vector2", 1, object.FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Vector2", 2, object.FLOAT_OBJ, args[1].Type())
			}
			x := float32(args[0].(*object.Float).Value)
			y := float32(args[1].(*object.Float).Value)
			return &object.GoObj[rl.Vector2]{Value: rl.NewVector2(x, y), Id: GoObjId.Add(1)}
		},
	},
	"_vector3": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("Vector3", len(args), 3, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Vector3", 1, object.FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Vector3", 2, object.FLOAT_OBJ, args[1].Type())
			}
			if args[2].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Vector3", 3, object.FLOAT_OBJ, args[2].Type())
			}
			x := float32(args[0].(*object.Float).Value)
			y := float32(args[1].(*object.Float).Value)
			z := float32(args[2].(*object.Float).Value)
			return &object.GoObj[rl.Vector3]{Value: rl.NewVector3(x, y, z), Id: GoObjId.Add(1)}
		},
	},
	"_vector4": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 4 {
				return newInvalidArgCountError("Vector4", len(args), 4, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Vector4", 1, object.FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Vector4", 2, object.FLOAT_OBJ, args[1].Type())
			}
			if args[2].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Vector4", 3, object.FLOAT_OBJ, args[2].Type())
			}
			if args[3].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Vector4", 4, object.FLOAT_OBJ, args[3].Type())
			}
			x := float32(args[0].(*object.Float).Value)
			y := float32(args[1].(*object.Float).Value)
			z := float32(args[2].(*object.Float).Value)
			w := float32(args[3].(*object.Float).Value)
			return &object.GoObj[rl.Vector4]{Value: rl.NewVector4(x, y, z, w), Id: GoObjId.Add(1)}
		},
	},
	"_camera2d": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 4 {
				return newInvalidArgCountError("Camera2D", len(args), 4, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("Camera2D", 1, object.GO_OBJ, args[0].Type())
			}
			offset, ok := args[0].(*object.GoObj[rl.Vector2])
			if !ok {
				return newPositionalTypeErrorForGoObj("Camera2D", 1, "rl.Vector2", offset)
			}
			if args[1].Type() != object.GO_OBJ {
				return newPositionalTypeError("Camera2D", 2, object.GO_OBJ, args[1].Type())
			}
			target, ok := args[1].(*object.GoObj[rl.Vector2])
			if !ok {
				return newPositionalTypeErrorForGoObj("Camera2D", 2, "rl.Vector2", target)
			}
			if args[2].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Camera2D", 3, object.FLOAT_OBJ, args[2].Type())
			}
			if args[3].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Camera2D", 4, object.FLOAT_OBJ, args[3].Type())
			}
			rotation := float32(args[2].(*object.Float).Value)
			zoom := float32(args[3].(*object.Float).Value)
			return &object.GoObj[rl.Camera2D]{Value: rl.NewCamera2D(offset.Value, target.Value, rotation, zoom), Id: GoObjId.Add(1)}
		},
	},
	"_begin_mode2d": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("begin_mode2d", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("begin_mode2d", 1, object.GO_OBJ, args[0].Type())
			}
			cam, ok := args[0].(*object.GoObj[rl.Camera2D])
			if !ok {
				return newPositionalTypeErrorForGoObj("begin_mode2d", 1, "rl.Camera2D", cam)
			}
			rl.BeginMode2D(cam.Value)
			return NULL
		},
	},
	"_end_mode2d": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("end_mode2d", len(args), 0, "")
			}
			rl.EndMode2D()
			return NULL
		},
	},
	"_camera3d": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 5 {
				return newInvalidArgCountError("Camera3D", len(args), 5, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("Camera3D", 1, object.GO_OBJ, args[0].Type())
			}
			position, ok := args[0].(*object.GoObj[rl.Vector3])
			if !ok {
				return newPositionalTypeErrorForGoObj("Camera3D", 1, "rl.Vector3", position)
			}
			if args[1].Type() != object.GO_OBJ {
				return newPositionalTypeError("Camera3D", 2, object.GO_OBJ, args[1].Type())
			}
			target, ok := args[1].(*object.GoObj[rl.Vector3])
			if !ok {
				return newPositionalTypeErrorForGoObj("Camera3D", 2, "rl.Vector3", target)
			}
			if args[2].Type() != object.GO_OBJ {
				return newPositionalTypeError("Camera3D", 3, object.GO_OBJ, args[2].Type())
			}
			up, ok := args[2].(*object.GoObj[rl.Vector3])
			if !ok {
				return newPositionalTypeErrorForGoObj("Camera3D", 3, "rl.Vector3", up)
			}
			if args[3].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Camera3D", 4, object.FLOAT_OBJ, args[3].Type())
			}
			if args[4].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("Camera3D", 5, object.INTEGER_OBJ, args[4].Type())
			}
			fovy := float32(args[3].(*object.Float).Value)
			projection := rl.CameraProjection(args[4].(*object.Integer).Value)
			return &object.GoObj[rl.Camera3D]{Value: rl.NewCamera3D(position.Value, target.Value, up.Value, fovy, projection), Id: GoObjId.Add(1)}
		},
	},
	"_begin_mode3d": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("begin_mode3d", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("begin_mode3d", 1, object.GO_OBJ, args[0].Type())
			}
			cam, ok := args[0].(*object.GoObj[rl.Camera3D])
			if !ok {
				return newPositionalTypeErrorForGoObj("begin_mode3d", 1, "rl.Camera3D", cam)
			}
			rl.BeginMode3D(cam.Value)
			return NULL
		},
	},
	"_end_mode3d": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("end_mode3d", len(args), 0, "")
			}
			rl.EndMode3D()
			return NULL
		},
	},
	"_init_audio_device": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("init_audio_device", len(args), 0, "")
			}
			rl.InitAudioDevice()
			return NULL
		},
	},
	"_close_audio_device": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("close_audio_device", len(args), 0, "")
			}
			rl.CloseAudioDevice()
			return NULL
		},
	},
	"_load_music": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("load_music", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("load_music", 1, object.STRING_OBJ, args[0].Type())
			}
			fname := args[0].(*object.Stringo).Value
			return &object.GoObj[rl.Music]{Value: rl.LoadMusicStream(fname), Id: GoObjId.Add(1)}
		},
	},
	"_update_music": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("update_music", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("update_music", 1, object.GO_OBJ, args[0].Type())
			}
			music, ok := args[0].(*object.GoObj[rl.Music])
			if !ok {
				return newPositionalTypeErrorForGoObj("update_music", 1, "rl.Music", music)
			}
			rl.UpdateMusicStream(music.Value)
			return NULL
		},
	},
	"_play_music": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("play_music", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("play_music", 1, object.GO_OBJ, args[0].Type())
			}
			music, ok := args[0].(*object.GoObj[rl.Music])
			if !ok {
				return newPositionalTypeErrorForGoObj("play_music", 1, "rl.Music", music)
			}
			rl.PlayMusicStream(music.Value)
			return NULL
		},
	},
	"_stop_music": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("stop_music", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("stop_music", 1, object.GO_OBJ, args[0].Type())
			}
			music, ok := args[0].(*object.GoObj[rl.Music])
			if !ok {
				return newPositionalTypeErrorForGoObj("stop_music", 1, "rl.Music", music)
			}
			rl.StopMusicStream(music.Value)
			return NULL
		},
	},
	"_resume_music": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("resume_music", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("resume_music", 1, object.GO_OBJ, args[0].Type())
			}
			music, ok := args[0].(*object.GoObj[rl.Music])
			if !ok {
				return newPositionalTypeErrorForGoObj("resume_music", 1, "rl.Music", music)
			}
			rl.ResumeMusicStream(music.Value)
			return NULL
		},
	},
	"_pause_music": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("pause_music", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("pause_music", 1, object.GO_OBJ, args[0].Type())
			}
			music, ok := args[0].(*object.GoObj[rl.Music])
			if !ok {
				return newPositionalTypeErrorForGoObj("pause_music", 1, "rl.Music", music)
			}
			rl.PauseMusicStream(music.Value)
			return NULL
		},
	},
	"_load_sound": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("load_sound", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("load_sound", 1, object.STRING_OBJ, args[0].Type())
			}
			fname := args[0].(*object.Stringo).Value
			return &object.GoObj[rl.Sound]{Value: rl.LoadSound(fname), Id: GoObjId.Add(1)}
		},
	},
	"_play_sound": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("play_sound", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("play_sound", 1, object.GO_OBJ, args[0].Type())
			}
			sound, ok := args[0].(*object.GoObj[rl.Sound])
			if !ok {
				return newPositionalTypeErrorForGoObj("play_sound", 1, "rl.Sound", sound)
			}
			rl.PlaySound(sound.Value)
			return NULL
		},
	},
	"_stop_sound": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("stop_sound", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("stop_sound", 1, object.GO_OBJ, args[0].Type())
			}
			sound, ok := args[0].(*object.GoObj[rl.Sound])
			if !ok {
				return newPositionalTypeErrorForGoObj("stop_sound", 1, "rl.Sound", sound)
			}
			rl.StopSound(sound.Value)
			return NULL
		},
	},
	"_resume_sound": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("resume_sound", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("resume_sound", 1, object.GO_OBJ, args[0].Type())
			}
			sound, ok := args[0].(*object.GoObj[rl.Sound])
			if !ok {
				return newPositionalTypeErrorForGoObj("resume_sound", 1, "rl.Sound", sound)
			}
			rl.ResumeSound(sound.Value)
			return NULL
		},
	},
	"_pause_sound": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("pause_sound", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("pause_sound", 1, object.GO_OBJ, args[0].Type())
			}
			sound, ok := args[0].(*object.GoObj[rl.Sound])
			if !ok {
				return newPositionalTypeErrorForGoObj("pause_sound", 1, "rl.Sound", sound)
			}
			rl.PauseSound(sound.Value)
			return NULL
		},
	},
	"_unload": {
		Fun: func(args ...object.Object) object.Object {
			for i, arg := range args {
				// If the arg is a list go through the list and check every arg to remove
				if arg.Type() == object.LIST_OBJ {
					l := arg.(*object.List).Elements
					for _, e := range l {
						maybeErr := unloadFromRaylib(e, i)
						if isError(maybeErr) {
							return maybeErr
						}
					}
				} else {
					maybeErr := unloadFromRaylib(arg, i)
					if isError(maybeErr) {
						return maybeErr
					}
				}
			}
			return NULL
		},
	},
})

func unloadFromRaylib(arg object.Object, pos int) object.Object {
	if arg.Type() != object.GO_OBJ {
		return newPositionalTypeError("unload", pos, object.GO_OBJ, arg.Type())
	}
	if tex, ok := arg.(*object.GoObj[rl.Texture2D]); ok {
		rl.UnloadTexture(tex.Value)
		return NULL
	} else if music, ok := arg.(*object.GoObj[rl.Music]); ok {
		rl.UnloadMusicStream(music.Value)
		return NULL
	} else if sound, ok := arg.(*object.GoObj[rl.Sound]); ok {
		rl.UnloadSound(sound.Value)
		return NULL
	}
	return newError("`unload` error: Failed to find gg object to unload, expected any GO_OBJ of [rl.Texture2D, rl.Music, rl.Sound]")
}
