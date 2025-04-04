package evaluator

import (
	"blue/ast"
	"blue/consts"
	"blue/evaluator/wazm"
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
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"

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
	"github.com/tetratelabs/wazero/api"
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
	"color":  {File: lib.ReadStdFileToString("color.b"), Builtins: _color_builtin_map},
	"csv":    {File: lib.ReadStdFileToString("csv.b"), Builtins: _csv_builtin_map},
	"psutil": {File: lib.ReadStdFileToString("psutil.b"), Builtins: _psutil_builtin_map},
	"wasm":   {File: lib.ReadStdFileToString("wasm.b"), Builtins: _wasm_builtin_map},
	"ui":     {File: lib.ReadStdFileToString("ui-static.b"), Builtins: NewBuiltinObjMap(BuiltinMapTypeInternal{})},
	"gg":     {File: lib.ReadStdFileToString("gg-static.b"), Builtins: NewBuiltinObjMap(BuiltinMapTypeInternal{})},
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
		newE.Builtins = append(newE.Builtins, fb.Builtins)
		setupBuiltinsWithEvaluator(name, newE)
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
		// TODO: See if we can cache this somehow
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
		fb.Env.SetAllPublicOnEnv(e.env)
		return NULL
	}

	mod := &object.Module{Name: name, Env: fb.Env, HelpStr: fb.HelpStr}
	e.env.Set(name, mod)
	return nil
}

// var goObjDecoders = map[string]any{}

func NewGoObj[T any](obj T) *object.GoObj[T] {
	gob := &object.GoObj[T]{Value: obj, Id: GoObjId.Add(1)}
	// Note: This is disabled for now due to the complexity of handling all Go Object Types supported by blue
	// t := fmt.Sprintf("%T", gob)
	// if _, ok := goObjDecoders[t]; !ok {
	// 	goObjDecoders[t] = gob.Decoder
	// }
	return gob
}

// Used to catch interrupt to shutdown server
var interruptCh = make(chan os.Signal, 1)

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
		HelpStr: helpStrArgs{
			explanation: "`url_encode` returns the STRING encoded as a valid URL",
			signature:   "url_encode(arg: str) -> str",
			errors:      "InvalidArgCount,PositionalType",
			example:     "url_encode('hello world') => 'hello%20world'",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`url_escape` returns the STRING encoded as a valid value to be passed through a URL",
			signature:   "url_escape(arg: str) -> str",
			errors:      "InvalidArgCount,PositionalType",
			example:     "url_escape('hello world') => 'hello+world'",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`url_unescape` returns the STRING encoded as a valid value to be passed through a URL",
			signature:   "url_unescape(arg: str) -> str",
			errors:      "InvalidArgCount,PositionalType",
			example:     "url_unescape('hello+world') => 'hello world'",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`download` copys the file at the URL to the given file path. if the fpath is empty, then the URL is used to determine the name",
			signature:   "download(url: str, fpath: str='') -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "download('http://example.com/test.txt') => null => writes test.txt to current directory",
		}.String(),
	},
	"_new_server": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("new_server", len(args), 0, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("new_server", 1, object.STRING_OBJ, args[0].Type())
			}
			network := args[0].(*object.Stringo).Value
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
				return newPositionalTypeErrorForGoObj("serve", 1, "*fiber.App", args[0])
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("serve", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("seve", 3, object.BOOLEAN_OBJ, args[2].Type())
			}
			useEmbeddedLibWeb := args[2].(*object.Boolean).Value
			addrPort := args[1].(*object.Stringo).Value
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
	},
	"_shutdown_server": {
		Fun: func(args ...object.Object) object.Object {
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
				return newPositionalTypeErrorForGoObj("static", 1, "*fiber.App", args[0])
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
		HelpStr: helpStrArgs{
			explanation: "`static` serves the given directory as static files for the http server",
			signature:   "static(server: GoObj[*fiber.App], prefix: str='/', dir_path: str='.', browse: bool=false) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "static() => null => current directory served at addr:port/",
		}.String(),
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
				return newPositionalTypeErrorForGoObj("ws_send", 1, "*websocket.Conn", args[0])
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
		HelpStr: helpStrArgs{
			explanation: "`ws_send` sends the given value on the websocket connection, if the value is a string the websocket message type is TextMessage, otherwise if bytes BinaryMessage",
			signature:   "ws_send(c: GoObj[*websocket.Conn], value: str|bytes) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "ws_send(c, '1') => null",
		}.String(),
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
				return newPositionalTypeErrorForGoObj("ws_send", 1, "*websocket.Conn", args[0])
			}
			mt, msg, err := c.Value.ReadMessage()
			if err != nil {
				// If its closed we still want to return an error so that the handler fn wont try to send NULL
				return newError("`ws_recv` error: %s", err.Error())
			}
			switch mt {
			case websocket.BinaryMessage:
				return &object.Bytes{Value: msg}
			case websocket.TextMessage:
				return &object.Stringo{Value: string(msg)}
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
			return object.CreateBasicMapObjectForGoObj("ws/client", NewGoObj(conn))
		},
		HelpStr: helpStrArgs{
			explanation: "`new_ws` returns a new websocket client object",
			signature:   "new_ws(url: str) -> {t: 'ws/client', v: GoObj[*ws.Conn]}",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "new_ws('http://localhost:3001/ws') => ws client obj",
		}.String(),
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
				return newPositionalTypeErrorForGoObj("ws_client_send", 1, "*ws.Conn", args[0])
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
		HelpStr: helpStrArgs{
			explanation: "`ws_client_send` sends the given value on the websocket client connection, if the value is a string the websocket message type is TextMessage, otherwise if bytes BinaryMessage",
			signature:   "ws_client_send(c: GoObj[*ws.Conn], value: str|bytes) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "ws_client_send(c, '1') => null",
		}.String(),
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
				return newPositionalTypeErrorForGoObj("ws_client_recv", 1, "*ws.Conn", args[0])
			}
			mt, msg, err := c.Value.ReadMessage()
			if err != nil {
				// If its closed we still want to return an error so that the handler fn wont try to send NULL
				return newError("`ws_client_recv` error: %s", err.Error())
			}
			switch mt {
			case websocket.BinaryMessage:
				return &object.Bytes{Value: msg}
			case websocket.TextMessage:
				return &object.Stringo{Value: string(msg)}
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
				return newPositionalTypeErrorForGoObj("handle_monitor", 1, "*fiber.App", args[0])
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
		HelpStr: helpStrArgs{
			explanation: "`handle_monitor` creates a monitor handler on the given http server at the given path a boolean that determines when it should show",
			signature:   "handle_monitor(s: GoObj[*fiber.App], path: str, should_show: bool) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "handle_monitor(s, '/monitor', true) => null",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`md_to_html` converts a given markdown string to valid html",
			signature:   "md_to_html(s: str) -> str",
			errors:      "InvalidArgCount,PositionalType",
			example:     "md_to_html('# Hello World') => '<h1>Hello World</h1>'",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`sanitize_and_minify` santizes and/or minifies the given content",
			signature:   "sanitize_and_minify(content: str, should_sanitize: bool=true, should_minify: bool=true) -> str",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "sanitize_and_minify('<script></script>') => ''",
		}.String(),
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
					return newPositionalTypeErrorForGoObj("inspect", 1, "*websocket.Conn", args[0])
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
					return newPositionalTypeErrorForGoObj("inspect", 1, "*ws.Conn", args[0])
				}
				mapObj := object.NewOrderedMap[string, object.Object]()
				mapObj.Set("remote_addr", &object.Stringo{Value: c.Value.RemoteAddr().String()})
				mapObj.Set("local_addr", &object.Stringo{Value: c.Value.LocalAddr().String()})
				mapObj.Set("remote_addr_network", &object.Stringo{Value: c.Value.RemoteAddr().Network()})
				mapObj.Set("local_addr_network", &object.Stringo{Value: c.Value.LocalAddr().Network()})
				return object.CreateMapObjectForGoMap(*mapObj)
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
	"_open_browser": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("open_browser", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("open_browser", 1, object.STRING_OBJ, args[0].Type())
			}
			url := args[0].(*object.Stringo).Value
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
		HelpStr: helpStrArgs{
			explanation: "`sleep` will sleep and block for the given INTEGER by milliseconds",
			signature:   "sleep(i: int) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "sleep(100) => null",
		}.String(),
	},
	"_now": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("now", len(args), 0, "")
			}
			return &object.Integer{Value: time.Now().UnixMilli()}
		},
		HelpStr: helpStrArgs{
			explanation: "`now` returns the current unix timestamp in milliseconds as an INTEGER",
			signature:   "now() -> int",
			errors:      "InvalidArgCount",
			example:     "now(100) => 1703479130205",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`parse` returns the parsed timestamp a unix timestamp in milliseconds as an INTEGER",
			signature:   "parse(s: str) -> int",
			errors:      "InvalidArgCount,PositionalType",
			example:     "parse('now') => 1703479130205",
		}.String(),
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
				return &object.Stringo{Value: carbon.CreateFromStdTime(tm).ToDateTimeMilliString(tz)}
			} else {
				return &object.Stringo{Value: carbon.CreateFromStdTime(tm).ToDateTimeMilliString()}
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`to_str` returns the string fomratted version of a unix timestamp value",
			signature:   "to_str(i: int) -> str",
			errors:      "InvalidArgCount,PositionalType",
			example:     "to_str(1703479130205) => '2023-12-24 23:42:28.144'",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`by_xpath` finds the string based on an xpath query from the given html",
			signature:   "by_xpath(str_to_search: str, str_query: str, should_find_one: bool) -> list[str]|str",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "by_xpath('<html><div id='abc'>123</div></html>', '//*[@id='abc']', true) => '<div id='abc'>123</div>'",
		}.String(),
	},
	"_by_regex": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("by_regex", len(args), 3, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("by_regex", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ && args[1].Type() != object.REGEX_OBJ {
				return newPositionalTypeError("by_regex", 2, object.STRING_OBJ+" or REGEX", args[1].Type())
			}
			if args[2].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("by_regex", 3, object.BOOLEAN_OBJ, args[2].Type())
			}
			strToSearch := args[0].(*object.Stringo).Value

			var re *regexp.Regexp
			if args[1].Type() == object.STRING_OBJ {
				strQuery := args[1].(*object.Stringo).Value
				re1, err := regexp.Compile(strQuery)
				if err != nil {
					return newError("`by_regex` error: failed to compile regexp %q", strQuery)
				}
				re = re1
			} else {
				re = args[1].(*object.Regex).Value
			}
			shouldFindOne := args[2].(*object.Boolean).Value

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
		HelpStr: helpStrArgs{
			explanation: "`by_regex` finds the string given a regex or string to search with",
			signature:   "by_regex(str_to_search: str, query: str|regex, should_find_one: bool) -> list[str]|str",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "by_regex('abc', r/abc/, true) => 'abc'",
		}.String(),
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
			return NewGoObj(db)
		},
		HelpStr: helpStrArgs{
			explanation: "`db_open` opens a connection to the builtin sqlite db and returns the DB obj",
			signature:   "db_open(db_name: str=':memory:') -> GoObj[*sql.DB]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "db_open() => GoObj[*sql.DB]",
		}.String(),
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
				return newPositionalTypeErrorForGoObj("db_ping", 1, "*sql.DB", args[0])
			}
			err := db.Value.Ping()
			if err != nil {
				return &object.Stringo{Value: err.Error()}
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`db_ping` pings the connection to the DB to verify connectivity. if no error, null is returned",
			signature:   "db_ping(db: GoObj[*sql.DB]) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "db_ping(db) => null",
		}.String(),
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
				return newPositionalTypeErrorForGoObj("db_close", 1, "*sql.DB", args[0])
			}
			err := db.Value.Close()
			if err != nil {
				return newError("`db_close` error: %s", err.Error())
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`db_close` closes the connection to the DB",
			signature:   "db_close(db: GoObj[*sql.DB]) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "db_close(db) => null",
		}.String(),
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
				return newPositionalTypeErrorForGoObj("db_exec", 1, "*sql.DB", args[0])
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
		HelpStr: helpStrArgs{
			explanation: "`db_exec` is used to execute queries against the DB that affect rows (ie. INSERT statments)",
			signature:   "db_exec(db: GoObj[*sql.DB], exec_query: str, exec_query_args: list[any]) -> {last_insert_id: _, rows_affected: _}",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "db_exec(db, 'CREATE TABLE ABC;', []) => {last_insert_id: 1, rows_affected: 1}",
		}.String(),
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
				return newPositionalTypeErrorForGoObj("db_query", 1, "*sql.DB", args[0])
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
		HelpStr: helpStrArgs{
			explanation: "`db_query` is used to query the DB (ie. SELECT)",
			signature:   "db_query(db: GoObj[*sql.DB], query: str, query_args: list[any], named_cols: bool) -> list[any]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "db_query(db, 'SELECT * FROM ABC;', [], false) => list[any]",
		}.String(),
	},
})

// greatest common divisor (GCD) via Euclidean algorithm
func gcd(a, b int64) int64 {
	for b != 0 {
		t := b
		b = a % b
		a = t
	}
	return a
}

// find Least Common Multiple (LCM) via GCD
func lcm(a, b int64, integers ...int64) int64 {
	result := a * b / gcd(a, b)
	for i := 0; i < len(integers); i++ {
		result = lcm(result, integers[i])
	}
	return result
}

var _math_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_rand": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("rand", len(args), 0, "")
			}
			return &object.Float{Value: mr.Float64()}
		},
		HelpStr: helpStrArgs{
			explanation: "`rand` returns a FLOAT a pseudo-random number in the half-open interval [0.0,1.0)",
			signature:   "rand() -> float",
			errors:      "InvalidArgCount",
			example:     "rand() => 0.125215",
		}.String(),
	},
	"_NaN": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("NaN", len(args), 0, "")
			}
			return &object.Float{Value: math.NaN()}
		},
		HelpStr: helpStrArgs{
			explanation: "`NaN` is the representation of NaN",
			signature:   "NaN() -> NaN",
			errors:      "InvalidArgCount",
			example:     "NaN() => NaN",
		}.String(),
	},
	"_acos": {
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
		HelpStr: helpStrArgs{
			explanation: "`acos` returns the arccosine, in radians, of x",
			signature:   "acos(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "acos(0.5) => 1.047198",
		}.String(),
	},
	"_acosh": {
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
		HelpStr: helpStrArgs{
			explanation: "`acosh` returns the inverse hyperbolic cosine of x",
			signature:   "acosh(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "acosh(1.04) => 0.281908",
		}.String(),
	},
	"_asin": {
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
		HelpStr: helpStrArgs{
			explanation: "`asin` returns the arcsine, in radians, of x",
			signature:   "asin(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "asin(0.4) => 0.411517",
		}.String(),
	},
	"_asinh": {
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
		HelpStr: helpStrArgs{
			explanation: "`asinh` returns the inverse hyperbolic sine of x",
			signature:   "asinh(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "asinh(0.4) => 0.390035",
		}.String(),
	},
	"_atan": {
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
		HelpStr: helpStrArgs{
			explanation: "`atan` returns the arctangent, in radians, of x",
			signature:   "atan(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "atan(0.4) => 0.380506",
		}.String(),
	},
	"_atan2": {
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
		HelpStr: helpStrArgs{
			explanation: "`atan2` returns the arc tangent of y/x, using the signs of the two to determine the quadrant of the return value",
			signature:   "atan2(x: float, y: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "atan2(0.4,0.4) => 0.785398",
		}.String(),
	},
	"_atanh": {
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
		HelpStr: helpStrArgs{
			explanation: "`atanh` returns the inverse hyperbolic tangent of x",
			signature:   "atanh(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "atanh(0.4) => 0.423649",
		}.String(),
	},
	"_cbrt": {
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
		HelpStr: helpStrArgs{
			explanation: "`cbrt` returns the cube root of x",
			signature:   "cbrt(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "cbrt(8.0) => 2.0",
		}.String(),
	},
	"_ceil": {
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
		HelpStr: helpStrArgs{
			explanation: "`ceil` returns the least integer value greater than or equal to x",
			signature:   "ceil(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "ceil(1.2) => 2.0",
		}.String(),
	},
	"_copysign": {
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
		HelpStr: helpStrArgs{
			explanation: "`copysign` returns a value with the magnitude of f and the sign of sign",
			signature:   "copysign(f: float, sign: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "copysign(1.2, -2.8) => -1.2",
		}.String(),
	},
	"_cos": {
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
		HelpStr: helpStrArgs{
			explanation: "`cos` returns the cosine of the radian argument x",
			signature:   "cos(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "cos(1.20) => 0.362358",
		}.String(),
	},
	"_cosh": {
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
		HelpStr: helpStrArgs{
			explanation: "`cosh` returns the hyperbolic cosine of x",
			signature:   "cosh(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "cosh(1.2) => 1.810656",
		}.String(),
	},
	"_dim": {
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
		HelpStr: helpStrArgs{
			explanation: "`dim` returns the maximum of x-y or 0",
			signature:   "dim(x: float, y: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "dim(3.4, 1.2) => 2.2",
		}.String(),
	},
	"_erf": {
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
		HelpStr: helpStrArgs{
			explanation: "`erf` returns the error function of x",
			signature:   "erf(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "erf(1.234567) => 0.919179",
		}.String(),
	},
	"_erfc": {
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
		HelpStr: helpStrArgs{
			explanation: "`erfc` returns the complementary error function of x",
			signature:   "erfc(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "erfc(1.234567) => 0.080821",
		}.String(),
	},
	"_erfcinv": {
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
		HelpStr: helpStrArgs{
			explanation: "`erfcinv` returns the inverse of erfc(x)",
			signature:   "erfcinv(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "erfcinv(1.234567) => -0.210968",
		}.String(),
	},
	"_erfinv": {
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
		HelpStr: helpStrArgs{
			explanation: "`erfinv` returns the inverse error function of x",
			signature:   "erfinv(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "erfinv(0.234567) => 0.210968",
		}.String(),
	},
	"_fma": {
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
		HelpStr: helpStrArgs{
			explanation: "`fma` returns x * y + z, computed with only one rounding. fma returns the fused multiply-add of x, y, and z",
			signature:   "fma(x: float, y: float, z: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "fma(2.0, 3.0, 4.0) => 10.0",
		}.String(),
	},
	"_floor": {
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
		HelpStr: helpStrArgs{
			explanation: "`floor` returns the greatest integer value less than or equal to x",
			signature:   "floor(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "floor(1.2) => 1.0",
		}.String(),
	},
	"_frexp": {
		Fun: func(args ...object.Object) object.Object {
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
		},
		HelpStr: helpStrArgs{
			explanation: "`frexp` breaks f into a normalized fraction and an integral power of two. it returns frac and exp satisfying f == frac x 2**exp, with the absolute value of frac in the interval [1/2, 1)",
			signature:   "frexp(x: float) -> {frac: float, exp: int}",
			errors:      "InvalidArgCount,PositionalType",
			example:     "frexp(3.0) => {frac: 0.750000, exp: 2}",
		}.String(),
	},
	"_gamma": {
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
		HelpStr: helpStrArgs{
			explanation: "`gamma` returns the Gamma function of x",
			signature:   "gamma(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "gamma(2.0) => 1.0",
		}.String(),
	},
	"_gcd": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("gcd", len(args), 2, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("gcd", 1, object.INTEGER_OBJ, args[0].Type())
			}
			if args[1].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("gcd", 2, object.INTEGER_OBJ, args[1].Type())
			}
			a, b := args[0].(*object.Integer).Value, args[1].(*object.Integer).Value
			return &object.Integer{Value: gcd(a, b)}
		},
		HelpStr: helpStrArgs{
			explanation: "`gcd` returns the greatest common divisor (GCD) via Euclidean algorithm",
			signature:   "gcd(a: int, b: int) -> int",
			errors:      "InvalidArgCount,PositionalType",
			example:     "gcd(10,20) => 10",
		}.String(),
	},
	"_hypot": {
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
		HelpStr: helpStrArgs{
			explanation: "`hypot` returns sqrt(p*p + q*q), taking care to avoid unnecessary overflow and underflow",
			signature:   "hypot(p: float, q: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "hypot(3.0,4.0) => 5.0",
		}.String(),
	},
	"_ilogb": {
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
		HelpStr: helpStrArgs{
			explanation: "`ilogb` returns the binary exponent of x as an INTEGER",
			signature:   "ilogb(x: float) -> int",
			errors:      "InvalidArgCount,PositionalType",
			example:     "ilogb(203.0) => 7",
		}.String(),
	},
	"_inf": {
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
		HelpStr: helpStrArgs{
			explanation: "`inf` returns positive infinity if sign >= 0, negative infinity if sign < 0",
			signature:   "inf(sign: int) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "inf(1) => +Inf",
		}.String(),
	},
	"_is_inf": {
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
		HelpStr: helpStrArgs{
			explanation: "`is_inf` reports whether f is an infinity, according to sign. if sign > 0 { f == +Inf } else if sign < 0 { f == -Inf } else if sign == 0 { f == +Inf || f == -Inf}",
			signature:   "is_inf(x: float, sign: int) -> bool",
			errors:      "InvalidArgCount,PositionalType",
			example:     "is_inf(inf(1), 0) => true",
		}.String(),
	},
	"_is_NaN": {
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
		HelpStr: helpStrArgs{
			explanation: "`is_NaN` reports whether f is not-a-number value",
			signature:   "is_NaN(x: float) -> bool",
			errors:      "InvalidArgCount,PositionalType",
			example:     "is_NaN(NaN) => true",
		}.String(),
	},
	"_j0": {
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
		HelpStr: helpStrArgs{
			explanation: "`j0` returns the order-zero Bessel function of the first kind",
			signature:   "j0(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "j0(1.2) => 0.671133",
		}.String(),
	},
	"_j1": {
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
		HelpStr: helpStrArgs{
			explanation: "`j1` returns the order-one Bessel function of the first kind",
			signature:   "j1(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "j1(1.2) => 0.498289",
		}.String(),
	},
	"_jn": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("jn", len(args), 2, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("jn", 1, object.FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("jn", 2, object.INTEGER_OBJ, args[1].Type())
			}
			n := int(args[1].(*object.Integer).Value)
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Jn(n, x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`jn` returns the order-n Bessel function of the first kind",
			signature:   "jn(x: float, n: int) -> float",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "jn(1.2, 3) => 0.032874",
		}.String(),
	},
	"_lcm": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) < 1 {
				return newInvalidArgCountError("lcm", len(args), 1, "as a list, or 2 or more")
			}
			if args[0].Type() == object.LIST_OBJ {
				l := args[0].(*object.List)
				ints := make([]int64, len(l.Elements))
				if len(l.Elements) < 2 {
					return newError("`lcm` error: list must be at least 2 elements long")
				}
				for i, e := range l.Elements {
					if e.Type() != object.INTEGER_OBJ {
						return newError("`lcm` error: all elements in list need to be INTEGER. got=%s", e.Type())
					}
					ints[i] = e.(*object.Integer).Value
				}
				if len(ints) > 2 {
					return &object.Integer{Value: lcm(ints[0], ints[1], ints[2:]...)}
				}
				return &object.Integer{Value: lcm(ints[0], ints[1])}
			}
			if len(args) < 2 {
				return newInvalidArgCountError("lcm", len(args), 2, "or more")
			}
			if len(args) == 2 {
				if args[0].Type() != object.INTEGER_OBJ {
					return newPositionalTypeError("lcm", 1, object.INTEGER_OBJ, args[0].Type())
				}
				if args[1].Type() != object.INTEGER_OBJ {
					return newPositionalTypeError("lcm", 2, object.INTEGER_OBJ, args[1].Type())
				}
				return &object.Integer{Value: lcm(args[0].(*object.Integer).Value, args[1].(*object.Integer).Value)}
			} else {
				ints := make([]int64, len(args))
				for i, e := range args {
					if e.Type() != object.INTEGER_OBJ {
						return newPositionalTypeError("lcm", i+1, object.INTEGER_OBJ, e.Type())
					}
					ints[i] = e.(*object.Integer).Value
				}
				return &object.Integer{Value: lcm(ints[0], ints[1], ints[2:]...)}
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`lcm` finds the Least Common Multiple (LCM) via GCD",
			signature:   "lcm(a: int, b: int, args: int) -> int || lcm(arg: list[int]) -> int",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "lcm(1,2,3,4) => 12",
		}.String(),
	},
	"_ldexp": {
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
		HelpStr: helpStrArgs{
			explanation: "`ldexp` is the inverse of frexp, returns frac x 2**exp.",
			signature:   "ldexp(frac: float, exp: int) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "ldexp(0.75, 2) => 3.0",
		}.String(),
	},
	"_lgamma": {
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
		HelpStr: helpStrArgs{
			explanation: "`lgamma` returns the natural logarithm and sign (-1 or +1) of gamma(x)",
			signature:   "lgamma(x: float) -> {lgamma: float, sign: int}",
			errors:      "InvalidArgCount,PositionalType",
			example:     "lgamma(2.3) => {lgamma: 0.154189, sign: 1}",
		}.String(),
	},
	"_log": {
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
		HelpStr: helpStrArgs{
			explanation: "`log` returns the natural logarithm of x",
			signature:   "log(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "log(120.0) => 4.787492",
		}.String(),
	},
	"_log10": {
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
		HelpStr: helpStrArgs{
			explanation: "`log10` returns the decimal logarithm of x",
			signature:   "log10(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "log10(120.0) => 2.079181",
		}.String(),
	},
	"_log1p": {
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
		HelpStr: helpStrArgs{
			explanation: "`log1p` returns the natural logarithm of 1 plus its argument x. it is more accurate than log(1 + x) when x is near zero",
			signature:   "log1p(x: float) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "log1p(0.2) => 0.182322",
		}.String(),
	},
	"_log2": {
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
		HelpStr: helpStrArgs{
			explanation: "`log2` returns the binary logarithm of x",
			signature:   "log2(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "log2(0.2) => -2.321928",
		}.String(),
	},
	"_logb": {
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
		HelpStr: helpStrArgs{
			explanation: "`logb` returns the binary exponent of x",
			signature:   "logb(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "logb(0.2) => -3.0",
		}.String(),
	},
	"_modf": {
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
		HelpStr: helpStrArgs{
			explanation: "`modf` returns INTEGER and fractional FLOAT numbers that sum to f. both values have the same sign as f",
			signature:   "modf(x: float) -> {i: int, frac: float}",
			errors:      "InvalidArgCount,PositionalType",
			example:     "modf(10.1) => {i: 10, frac: 0.1}",
		}.String(),
	},
	"_next_after": {
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
		HelpStr: helpStrArgs{
			explanation: "`next_after` returns the next representable FLOAT value after x towards y",
			signature:   "next_after(x: float, y: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "next_after(3.1, 5.0) => 3.1",
		}.String(),
	},
	"_remainder": {
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
		HelpStr: helpStrArgs{
			explanation: "`remainder` returns the FLOAT remainder of x/y",
			signature:   "remainder(x: float, y: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "remainder(98.2,38.3) => -16.7",
		}.String(),
	},
	"_round": {
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
		HelpStr: helpStrArgs{
			explanation: "`round` returns the nearest integer as a float, rounding half away from zero",
			signature:   "round(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "round(3.5) => 4.0",
		}.String(),
	},
	"_round_to_even": {
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
		HelpStr: helpStrArgs{
			explanation: "`round_to_even` returns the nearest integer as a float, rounding ties to even",
			signature:   "round_to_even(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "round_to_even(3.2) => 3.0",
		}.String(),
	},
	"_signbit": {
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
		HelpStr: helpStrArgs{
			explanation: "`signbit` reports whether x is negative or negative zero",
			signature:   "signbit(x: float) -> bool",
			errors:      "InvalidArgCount,PositionalType",
			example:     "signbit(-3.0) => true",
		}.String(),
	},
	"_sin": {
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
		HelpStr: helpStrArgs{
			explanation: "`sin` returns the sine of the radian argument x",
			signature:   "sin(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "sin(0.5) => 0.479426",
		}.String(),
	},
	"_sincos": {
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
		HelpStr: helpStrArgs{
			explanation: "`sincos` returns sin(x), cos(x)",
			signature:   "sincos(x: float) -> {sin: float, cos: float}",
			errors:      "InvalidArgCount,PositionalType",
			example:     "sincos(0.5) => {sin: 0.479426, cos: 0.877583}",
		}.String(),
	},
	"_sinh": {
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
		HelpStr: helpStrArgs{
			explanation: "`sinh` returns the hyperbolic sine of x",
			signature:   "sinh(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "sinh(0.5) => 0.521095",
		}.String(),
	},
	"_tan": {
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
		HelpStr: helpStrArgs{
			explanation: "`tan` returns the tangent of the radian argument x",
			signature:   "tan(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "tan(0.5) => 0.546302",
		}.String(),
	},
	"_tanh": {
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
		HelpStr: helpStrArgs{
			explanation: "`tanh` returns the hyperbolic tangent of x",
			signature:   "tanh(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "tanh(0.5) => 0.462117",
		}.String(),
	},
	"_trunc": {
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
		HelpStr: helpStrArgs{
			explanation: "`trunc` returns the integer value of x as a FLOAT",
			signature:   "trunc(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "trunc(2.5) => 2.0",
		}.String(),
	},
	"_y0": {
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
		HelpStr: helpStrArgs{
			explanation: "`y0` returns the order-zero Bessel function of the second kind",
			signature:   "y0(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "y0(2.0) => 0.510376",
		}.String(),
	},
	"_y1": {
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
		HelpStr: helpStrArgs{
			explanation: "`y1` returns the order-one Bessel function of the second kind",
			signature:   "y1(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "y1(2.0) => -0.107032",
		}.String(),
	},
	"_yn": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("yn", len(args), 2, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("yn", 1, object.FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("yn", 2, object.INTEGER_OBJ, args[1].Type())
			}
			n := int(args[1].(*object.Integer).Value)
			x := args[0].(*object.Float).Value
			return &object.Float{Value: math.Yn(n, x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`yn` returns the order-n Bessel function of the second kind",
			signature:   "yn(x: float, n: int) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "yn(3.0, 5) => -1.905946",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`load_file` returns the object version of the parsed config file (yaml, ini, toml, properties, json)",
			signature:   "load_file(fpath: str) -> str(json)",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "load_file(fpath) => {}",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`dump_config` takes the config map and writes it to a file in the given format",
			signature:   "dump_config(c: str(json), fpath: str, format: str('JSON'|'TOML'|'YAML'|'INI'|'PROPERTIES')='JSON') -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "dump_config(c, 'test.json') => null",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`sha` returns the sha 1, 256, or 512 sum of the given content as a STRING",
			signature:   "sha(content: str|bytes, type: int(1|256|512)) -> str",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "sha('a',1) => '86f7e437faa5a7fce15d1ddcb9eaeaea377667b8'",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`md5` returns the md5 sum of the given content as a STRING",
			signature:   "md5(content: str|bytes) -> str",
			errors:      "InvalidArgCount,PositionalType",
			example:     "md5('a') => '0cc175b9c0f1b6a831c399e269772661'",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`generate_from_password` returns a bcyrpt STRING for the given password STRING",
			signature:   "generate_from_password(pw: str) -> str",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "generate_from_password('a') => '$2a$10$4GjpUS8/60qPsxFtPbo.3e5ueULg4Llk0iCwVsGAV9LBDuw2FkSa2'",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`compare_hash_and_password` returns a true if the given hashed password matches the given password",
			signature:   "compare_hash_and_password(hashed_pw: str, pw: str) -> bool",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "compare_hash_and_password('$2a$10$4GjpUS8/60qPsxFtPbo.3e5ueULg4Llk0iCwVsGAV9LBDuw2FkSa2', 'a') => true",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`encrypt` encrypts the data given with the password given",
			signature:   "encrypt(pw: str|bytes, data: str|bytes) -> bytes",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "encrypt('a','test') => bytes",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`decrypt` decrypts the data given with the password given, bytes are returned if as_bytes is set to true",
			signature:   "decrypt(pw: str|bytes, data: bytes, as_bytes: bool=false) -> str|bytes",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "decrypt('a',bs) => 'test'",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`encode_base_64_32` encodes the data given in base64 if true, else base32, bytes are returned if as_bytes is set to true. Note: this function should only be called from encode",
			signature:   "encode_base_64_32(data: str|bytes, is_base_64: bool=false, as_bytes: bool=false) -> str|bytes",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "encode_base_64_32('a', true, false) => 'YQ=='",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`decode_base_64_32` decodes the data given in base64 if true, else base32, bytes are returned if as_bytes is set to true. Note: this function should only be called from decode",
			signature:   "decode_base_64_32(data: str|bytes, is_base_64: bool=false, as_bytes: bool=false) -> str|bytes",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "decode_base_64_32('YQ==', true, false) => 'a'",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`decode_hex` decodes the data given in hex, bytes are returned if as_bytes is set to true. Note: this function should only be called from decode",
			signature:   "decode_hex(data: str|bytes, as_bytes: bool=false) -> str|bytes",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "decode_hex('61') => 'a'",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`encode_hex` encodes the data given as hex, bytes are returned if as_bytes is set to true. Note: this function should only be called from encode",
			signature:   "encode_hex(data: str|bytes, as_bytes: bool=false) -> str|bytes",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "encode_hex('a') => '61'",
		}.String(),
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
			addrStr := net.JoinHostPort(addr, port)
			conn, err := net.Dial(transport, addrStr)
			if err != nil {
				return newError("`connect` error: %s", err.Error())
			}
			return object.CreateBasicMapObjectForGoObj("net", NewGoObj(conn))
		},
		HelpStr: helpStrArgs{
			explanation: "`connect` connects to the given transport://addr:port",
			signature: `connect(transport: str('tcp'|'tcp4'|'tcp6'|'udp'|'udp4'|'udp6'|'ip'|'ip4'|'ip6'|'unix'|'unixgram'|'unixpacket')='tcp',
			addr: str='localhost', port: str='18650') -> {t: 'net', v: GoObj[net.Conn]}`,
			errors:  "InvalidArgCount,PositionalType,CustomError",
			example: "connect() => {t: 'net', v: GoObj[net.Conn]}",
		}.String(),
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
				return object.CreateBasicMapObjectForGoObj("net/udp", NewGoObj(l))
			}
			l, err := net.Listen(transport, addrStr)
			if err != nil {
				return newError("`listen` error: %s", err.Error())
			}
			return object.CreateBasicMapObjectForGoObj("net/tcp", NewGoObj(l))
		},
		HelpStr: helpStrArgs{
			explanation: "`listen` listens for connections on the given transport://addr:port",
			signature: `listen(transport: str('tcp'|'tcp4'|'tcp6'|'udp'|'udp4'|'udp6'|'ip'|'ip4'|'ip6'|'unix'|'unixgram'|'unixpacket')='tcp',
			addr: str='localhost', port: str='18650') -> {t: 'net/tcp'|'net/udp', v: GoObj[net.Listener]|GoObj[*net.UDPConn]}`,
			errors:  "InvalidArgCount,PositionalType,CustomError",
			example: "listen() => {t: 'net/tcp', v: GoObj[net.Listener]}",
		}.String(),
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
				return newPositionalTypeErrorForGoObj("accept", 1, "net.Listener", args[0])
			}
			conn, err := l.Value.Accept()
			if err != nil {
				return newError("`accept` error: %s", err.Error())
			}
			return object.CreateBasicMapObjectForGoObj("net", NewGoObj(conn))
		},
		HelpStr: helpStrArgs{
			explanation: "`accept` accepts connections on the given listener",
			signature:   "accept(l: {t: 'net/tcp', v: GoObj[net.Listener]}) -> {t: 'net', v: GoObj[net.Conn]}",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "accept(l) => {t: 'net', v: GoObj[net.Conn]}",
		}.String(),
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
					return newPositionalTypeErrorForGoObj("net_close", 1, "*net.UDPConn", args[0])
				}
				c.Value.Close()
			case "net/tcp":
				listener, ok := args[0].(*object.GoObj[net.Listener])
				if !ok {
					return newPositionalTypeErrorForGoObj("net_close", 1, "net.Listener", args[0])
				}
				listener.Value.Close()
			case "net":
				conn, ok := args[0].(*object.GoObj[net.Conn])
				if !ok {
					return newPositionalTypeErrorForGoObj("net_close", 1, "net.Conn", args[0])
				}
				conn.Value.Close()
			default:
				return newError("`net_close` expects type of 'net/tcp', 'net/udp', or 'net'")
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`net_close` closes the given connection/listener",
			signature:   "net_close(c: {t: 'net/tcp', v: GoObj[net.Listener]}|{t: 'net/udp', v: GoObj[*net.UDPConn]}|{t: 'net', v: GoObj[net.Conn]}) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "net_close(c) => null",
		}.String(),
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
					return newPositionalTypeErrorForGoObj("net_read", 1, "*net.UDPConn", args[0])
				}
				conn = c.Value
			} else {
				c, ok := args[0].(*object.GoObj[net.Conn])
				if !ok {
					return newPositionalTypeErrorForGoObj("net_read", 1, "net.Conn", args[0])
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
		HelpStr: helpStrArgs{
			explanation: "`net_read` reads on the given connection to end_byte (default '\\n') or the buffer length, returning a string or bytes if as_bytes is true",
			signature:   "net_read(conn_v: GoObj[*net.UDPConn]|GoObj[net.Conn], conn_t: 'net/tcp'|'net/udp'|'net', end_byte_or_len: str|int|null=null, as_bytes: bool=false) -> str|bytes",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "net_read(c.v, c.t) => 'test'",
		}.String(),
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
					return newPositionalTypeErrorForGoObj("net_write", 1, "*net.UDPConn", args[0])
				}
				conn = c.Value
			} else {
				c, ok := args[0].(*object.GoObj[net.Conn])
				if !ok {
					return newPositionalTypeErrorForGoObj("net_write", 1, "net.Conn", args[0])
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
		HelpStr: helpStrArgs{
			explanation: "`net_write` writes the string/bytes on the given connection in full or to the end_byte (default null)",
			signature:   "net_write(conn_v: GoObj[*net.UDPConn]|GoObj[net.Conn], conn_t: 'net/tcp'|'net/udp'|'net', value: str|bytes, end_byte: str|null=null) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "net_write(c.v, c.t, 'test') => null",
		}.String(),
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
					return newPositionalTypeErrorForGoObj("inspect", 1, "*net.UDPConn", args[0])
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
					return newPositionalTypeErrorForGoObj("inspect", 1, "net.Listener", args[0])
				}
				mapObj := object.NewOrderedMap[string, object.Object]()
				mapObj.Set("addr", &object.Stringo{Value: l.Value.Addr().String()})
				mapObj.Set("addr_network", &object.Stringo{Value: l.Value.Addr().Network()})
				return object.CreateMapObjectForGoMap(*mapObj)
			case "net":
				c, ok := args[0].(*object.GoObj[net.Conn])
				if !ok {
					return newPositionalTypeErrorForGoObj("inspect", 1, "net.Conn", args[0])
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
		HelpStr: helpStrArgs{
			explanation: "`inspect` returns a map of info for the given net object",
			signature:   "inspect(conn_v: GoObj[*net.UDPConn]|GoObj[net.Conn]|GoObj[net.Listener], conn_t: 'net/tcp'|'net/udp'|'net') -> map[str:str]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "inspect(c.v, c.t) => {'addr': '127.0.0.1', 'addr_network': 'tcp'}",
		}.String(),
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
			return object.CreateBasicMapObjectForGoObj("color", NewGoObj(s))
		},
		HelpStr: helpStrArgs{
			explanation: "`style` returns an object to be used in printing that affects the stylized output",
			signature:   "style(text: int=normal, fg_color: int=normal, bg_color: int=normal) -> {t: 'color', v: GoObj[color.Style]}",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "style(fg_color=magenta, bg_color=white) => color style object",
		}.String(),
	},
	"_normal": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("normal", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.Normal)}
		},
		HelpStr: helpStrArgs{
			explanation: "`normal` returns the int version of the normal color",
			signature:   "normal() -> int",
			errors:      "InvalidArgCount",
			example:     "normal() -> int",
		}.String(),
	},
	"_red": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("red", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.Red)}
		},
		HelpStr: helpStrArgs{
			explanation: "`red` returns the int version of the red color",
			signature:   "red() -> int",
			errors:      "InvalidArgCount",
			example:     "red() -> int",
		}.String(),
	},
	"_cyan": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("cyan", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.Cyan)}
		},
		HelpStr: helpStrArgs{
			explanation: "`cyan` returns the int version of the cyan color",
			signature:   "cyan() -> int",
			errors:      "InvalidArgCount",
			example:     "cyan() -> int",
		}.String(),
	},
	"_gray": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("gray", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.Gray)}
		},
		HelpStr: helpStrArgs{
			explanation: "`gray` returns the int version of the gray color",
			signature:   "gray() -> int",
			errors:      "InvalidArgCount",
			example:     "gray() -> int",
		}.String(),
	},
	"_blue": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("blue", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.Blue)}
		},
		HelpStr: helpStrArgs{
			explanation: "`blue` returns the int version of the blue color",
			signature:   "blue() -> int",
			errors:      "InvalidArgCount",
			example:     "blue() -> int",
		}.String(),
	},
	"_black": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("black", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.Black)}
		},
		HelpStr: helpStrArgs{
			explanation: "`black` returns the int version of the black color",
			signature:   "black() -> int",
			errors:      "InvalidArgCount",
			example:     "black() -> int",
		}.String(),
	},
	"_green": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("green", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.Green)}
		},
		HelpStr: helpStrArgs{
			explanation: "`green` returns the int version of the green color",
			signature:   "green() -> int",
			errors:      "InvalidArgCount",
			example:     "green() -> int",
		}.String(),
	},
	"_white": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("white", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.White)}
		},
		HelpStr: helpStrArgs{
			explanation: "`white` returns the int version of the white color",
			signature:   "white() -> int",
			errors:      "InvalidArgCount",
			example:     "white() -> int",
		}.String(),
	},
	"_yellow": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("yellow", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.Yellow)}
		},
		HelpStr: helpStrArgs{
			explanation: "`yellow` returns the int version of the yellow color",
			signature:   "yellow() -> int",
			errors:      "InvalidArgCount",
			example:     "yellow() -> int",
		}.String(),
	},
	"_magenta": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("magenta", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.Magenta)}
		},
		HelpStr: helpStrArgs{
			explanation: "`magenta` returns the int version of the magenta color",
			signature:   "magenta() -> int",
			errors:      "InvalidArgCount",
			example:     "magenta() -> int",
		}.String(),
	},
	"_bold": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("bold", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.Bold)}
		},
		HelpStr: helpStrArgs{
			explanation: "`bold` returns the int version of the bold color",
			signature:   "bold() -> int",
			errors:      "InvalidArgCount",
			example:     "bold() -> int",
		}.String(),
	},
	"_italic": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("italic", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.OpItalic)}
		},
		HelpStr: helpStrArgs{
			explanation: "`italic` returns the int version of the italic color",
			signature:   "italic() -> int",
			errors:      "InvalidArgCount",
			example:     "italic() -> int",
		}.String(),
	},
	"_underlined": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("underlined", len(args), 0, "")
			}
			return &object.Integer{Value: int64(color.OpUnderscore)}
		},
		HelpStr: helpStrArgs{
			explanation: "`underlined` returns the int version of the underlined color",
			signature:   "underlined() -> int",
			errors:      "InvalidArgCount",
			example:     "underlined() -> int",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`parse` parses the string or bytes as a CSV and returns the data as a list of objects",
			signature:   "parse(data: str|bytes, delimeter: str=',', named_fields: bool=false, comment: str|null=null, lazy_quotes: bool=false, trim_leading_space: bool=false) -> list[any]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "parse(data) => list[any]",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`dump` dumps the data to a CSV",
			signature:   "dump(data: list[any], comma: str=',', use_crlf: bool=false) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "dump(data) => null",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`cpu_usage_percent` returns a list of cpu usages as floats per core",
			signature:   "cpu_usage_percent() -> list[float]",
			errors:      "InvalidArgCount,CustomError",
			example:     "cpu_usage_percent() => [1.0,0.4,0.2,0.6]",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`cpu_info` returns a list of json strings of cpu info per prcoessor",
			signature:   "cpu_info() -> list[str]",
			errors:      "InvalidArgCount,CustomError",
			example:     "cpu_info() => [json_with_keys('cpu','vendorId','family','model','stepping','physicalId','coreId','cores','modelName','mhz','cacheSize','flags','microcode')]",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`cpu_time_info` returns a list of json strings of cpu time stat info per prcoessor",
			signature:   "cpu_time_info() -> list[str]",
			errors:      "InvalidArgCount,CustomError",
			example:     "cpu_time_info() => [json_with_keys('cpu','user','system','idle','nice','iowait','irq','softirq','steal','guest','guestNice')]",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`cpu_count` returns the number of cores as an INTEGER",
			signature:   "cpu_count() -> int",
			errors:      "InvalidArgCount,CustomError",
			example:     "cpu_count() => 4",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`mem_virt_info` returns a json string of virtual memory info",
			signature:   "mem_virt_info() -> str",
			errors:      "InvalidArgCount,CustomError",
			example:     "mem_virt_info() => json_with_keys('total','available','used','usedPercent','free','active','inactive','wired','laundry','buffers','cached','writeBack','dirty','writeBackTmp','shared','slab','sreclaimable','sunreclaim','pageTables','swapCached','commitLimit','committedAS','highTotal','highFree','lowTotal','lowFree','swapTotal','swapFree','mapped','vmallocTotal','vmallocUsed','vmallocChunk','hugePagesTotal','hugePagesFree','hugePagesRsvd','hugePagesSurp','hugePageSize')",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`mem_swap_info` returns a json string of swap memory info",
			signature:   "mem_swap_info() -> str",
			errors:      "InvalidArgCount,CustomError",
			example:     "mem_swap_info() => json_with_keys('total','used','free','usedPercent','sin','sout','pgIn','pgOut','pgFault','pgMajFault')",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`host_info` returns a json string of host info",
			signature:   "host_info() -> str",
			errors:      "InvalidArgCount,CustomError",
			example:     "host_info() => json_with_keys('hostname','uptime','bootTime','procs','os','platform','platformFamily','platformVersion','kernelVersion','kernelArch','virtualizationSystem','virtualizationRole','hostId')",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`host_temps_info` returns a list of json strings of host sensor temperature info",
			signature:   "host_temps_info() -> list[str]",
			errors:      "InvalidArgCount,CustomError",
			example:     "host_temps_info() => [json_with_keys('sensorKey','temperature','sensorHigh','sensorCritical')]",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`net_connections` returns a list of json strings of host network connection stats for the given option",
			signature:   "net_connections(option: str('all'|'tcp'|'tcp4'|'tcp6'|'udp'|'udp4'|'udp6'|'inet'|'inet4'|'inet6')='all') -> list[str]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "net_connections() => [json_with_keys('fd','family','type','localaddr','remoteaddr','status','uids','pid')]",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`net_io_info` returns a list of json strings of network io stat info",
			signature:   "net_io_info() -> list[str]",
			errors:      "InvalidArgCount,CustomError",
			example:     "net_io_info() => [json_with_keys('name','bytesSent','bytesRecv','packetsSent','packetsRecv','errin','errout','dropin','dropout','fifoin','fifoout')]",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`disk_partitions` returns a list of json strings of disk partition info",
			signature:   "disk_partitions() -> list[str]",
			errors:      "InvalidArgCount,CustomError",
			example:     "disk_partitions() => [json_with_keys('device','mountpoint','fstype','opts')]",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`disk_io_info` returns a map of drive to json string of disk io info",
			signature:   "disk_io_info() -> map[str:str]",
			errors:      "InvalidArgCount,CustomError",
			example:     "disk_io_info() => {'drive': json_with_keys('readCount','mergedReadCount','writeCount','mergedWriteCount','readBytes','writeBytes','readTime','writeTime','iopsInProgress','ioTime','weightedIO','name','serialNumber','label')...}",
		}.String(),
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
		HelpStr: helpStrArgs{
			explanation: "`disk_usage` returns a json string of disk usage for the given path",
			signature:   "disk_usage(path: str) -> str",
			errors:      "InvalidArgCount,CustomError",
			example:     "disk_usage(root_path) => json_with_keys('path','fstype','total','free','used','usedPercent','inodesTotal','inodesUsed','inodesFree','inodesUsedPercent')",
		}.String(),
	},
})

var _wasm_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_wasm_init": {
		Fun: func(args ...object.Object) object.Object {
			//_wasm_init(wasm_code_path, args, mounts, stdout, stderr, stdin, envs, enable_rand, enable_time_and_sleep_precision, host_logging, listens, timeout)
			//(wasm_code_path, args=ARGV, mounts={'.':'/'}, stdout=FSTDOUT, stderr=FSTDERR, stdin=FSTDIN,
			//envs=ENV, enable_rand=true, enable_time_and_sleep_precision=true, host_logging='', listens=[], timeout=0) {
			if len(args) != 12 {
				return newInvalidArgCountError("wasm_init", len(args), 12, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("wasm_init", 1, object.STRING_OBJ, args[0].Type())
			}
			wasmCodePath := args[0].(*object.Stringo).Value
			wasmArgs := []string{}
			if args[1].Type() != object.LIST_OBJ && args[1].Type() != object.NULL_OBJ {
				return newPositionalTypeError("wasm_init", 2, "list[str] or null", args[1].Type())
			}
			if args[1].Type() == object.LIST_OBJ {
				l := args[1].(*object.List).Elements
				wasmArgs = make([]string, len(l))
				for i, e := range l {
					if e.Type() != object.STRING_OBJ {
						return newError("`wasm_init` error: found non-string element in 'args' list")
					}
					wasmArgs[i] = e.(*object.Stringo).Value
				}
			}
			mounts := make(map[string]string)
			if args[2].Type() != object.MAP_OBJ && args[2].Type() != object.NULL_OBJ {
				return newPositionalTypeError("wasm_init", 3, "map[str]str or null", args[2].Type())
			}
			if args[2].Type() == object.MAP_OBJ {
				m := args[2].(*object.Map).Pairs
				for _, k := range m.Keys {
					mp, _ := m.Get(k)
					if mp.Key.Type() != object.STRING_OBJ {
						return newError("`wasm_init` error: found non-string key in 'mounts' map")
					}
					if mp.Value.Type() != object.STRING_OBJ {
						return newError("`wasm_init` error: found non-string key in 'mounts' map")
					}
					mounts[mp.Key.(*object.Stringo).Value] = mp.Value.(*object.Stringo).Value
				}
			}
			if args[3].Type() != object.GO_OBJ && args[3].Type() != object.NULL_OBJ {
				return newPositionalTypeError("wasm_init", 4, "GO_OBJ[*os.File] or null", args[3].Type())
			}
			var stdout io.Writer = nil
			var stdin io.Reader = nil
			var stderr *os.File
			if args[3].Type() == object.GO_OBJ {
				sout, ok := args[3].(*object.GoObj[*os.File])
				if !ok {
					return newPositionalTypeErrorForGoObj("wasm_init", 4, "*os.File", args[3])
				}
				stdout = sout.Value
			} else {
				stdout = nil
			}
			if args[4].Type() != object.GO_OBJ && args[4].Type() != object.NULL_OBJ {
				return newPositionalTypeError("wasm_init", 5, "GO_OBJ[*os.File] or null", args[4].Type())
			}
			if args[4].Type() == object.GO_OBJ {
				serr, ok := args[4].(*object.GoObj[*os.File])
				if !ok {
					return newPositionalTypeErrorForGoObj("wasm_init", 5, "*os.File", args[4])
				}
				stderr = serr.Value
			} else {
				stderr = nil
			}
			if args[5].Type() != object.GO_OBJ && args[5].Type() != object.NULL_OBJ {
				return newPositionalTypeError("wasm_init", 6, "GO_OBJ[*os.File] or null", args[5].Type())
			}
			if args[5].Type() == object.GO_OBJ {
				sin, ok := args[5].(*object.GoObj[*os.File])
				if !ok {
					return newPositionalTypeErrorForGoObj("wasm_init", 6, "*os.File", args[5])
				}
				stdin = sin.Value
			} else {
				stdin = nil
			}
			envs := make(map[string]string)
			if args[6].Type() != object.MAP_OBJ && args[6].Type() != object.NULL_OBJ {
				return newPositionalTypeError("wasm_init", 7, "map[str]str or null", args[6].Type())
			}
			if args[6].Type() == object.MAP_OBJ {
				m := args[6].(*object.Map).Pairs
				for _, k := range m.Keys {
					mp, _ := m.Get(k)
					if mp.Key.Type() != object.STRING_OBJ {
						return newError("`wasm_init` error: found non-string key in 'envs' map")
					}
					if mp.Value.Type() != object.STRING_OBJ {
						return newError("`wasm_init` error: found non-string value in 'envs' map")
					}
					envs[mp.Key.(*object.Stringo).Value] = mp.Value.(*object.Stringo).Value
				}
			}
			if args[7].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("wasm_init", 8, object.BOOLEAN_OBJ, args[7].Type())
			}
			enableRand := args[7].(*object.Boolean).Value
			if args[8].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("wasm_init", 9, object.BOOLEAN_OBJ, args[8].Type())
			}
			enableTimeAndSleepPrecision := args[8].(*object.Boolean).Value
			if args[9].Type() != object.STRING_OBJ {
				return newPositionalTypeError("wasm_init", 10, object.STRING_OBJ, args[9].Type())
			}
			hostLogging := args[9].(*object.Stringo).Value
			listens := []string{}
			if args[10].Type() != object.LIST_OBJ && args[10].Type() != object.NULL_OBJ {
				return newPositionalTypeError("wasm_init", 11, "list[str] or null", args[10].Type())
			}
			if args[10].Type() == object.LIST_OBJ {
				l := args[10].(*object.List).Elements
				listens = make([]string, len(l))
				for i, e := range l {
					if e.Type() != object.STRING_OBJ {
						return newError("`wasm_init` error: found non-string element in 'listens' list")
					}
					listens[i] = e.(*object.Stringo).Value
				}
			}
			if args[11].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("wasm_init", 12, object.INTEGER_OBJ, args[11].Type())
			}
			timeoutDuration := time.Duration(args[11].(*object.Integer).Value)

			var bs []byte
			if IsEmbed {
				s := wasmCodePath
				if strings.HasPrefix(s, "./") {
					s = strings.TrimLeft(s, "./")
				}
				fileData, err := Files.ReadFile(consts.EMBED_FILES_PREFIX + s)
				if err != nil {
					// Fallback option for reading when in embedded context
					fileData, err := os.ReadFile(wasmCodePath)
					if err != nil {
						return newError("`wasm_init` error reading wasm_code_path `%s`: %s", wasmCodePath, err.Error())
					}
					bs = fileData
				} else {
					bs = fileData
				}
			} else {
				fileData, err := os.ReadFile(wasmCodePath)
				if err != nil {
					return newError("`wasm_init` error reading wasm_code_path `%s`: %s", wasmCodePath, err.Error())
				}
				bs = fileData
			}
			wc := wazm.Config{
				WasmExe: bs,
				StdIn:   stdin,
				StdOut:  stdout,
				StdErr:  stderr,
				Args:    wasmArgs,
				Envs:    envs,
				Mounts:  mounts,
				Listens: listens,

				EnableRandSource:            enableRand,
				EnableTimeAndSleepPrecision: enableTimeAndSleepPrecision,

				HostLogging: hostLogging,
				Timeout:     timeoutDuration,
			}
			wm, err := wazm.WazmInit(wc)
			if err != nil {
				return newError("`wasm_init` error: failed initalizing %s", err.Error())
			}
			return NewGoObj(wm)
		},
		HelpStr: helpStrArgs{
			explanation: "`wasm_init` initalizes a wasm module with all the necessary parameters to interact with it. Note: the module should be built with wasi_preview1 ie. GOOS=wasip1 GOARCH=wasm go build -o cat.wasm",
			signature: `wasm_init(wasm_code_path: str, args: list[str], mounts: map[str:str], stdout: GoObj[*os.File],
			stderr: GoObj[*os.File], stdin: GoObj[*os.File], envs: map[str:str], enable_rand: bool=true
			enable_time_and_sleep_precision: bool=true, host_logging: str='', listens: list[str]|null=[], timeout: int=0) -> GoObj[*wazm.Module]`,
			errors:  "InvalidArgCount,PositionalType,CustomError",
			example: "wasm_init('wasm_test_files/cat.wasm', args=['wasm_test_files/cat.go.tmp']) => GoObj[*wazm.Module]",
		}.String(),
	},
	"_wasm_get_functions": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("wasm_get_functions", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("wasm_get_functions", 1, object.GO_OBJ, "")
			}
			wm, ok := args[0].(*object.GoObj[*wazm.Module])
			if !ok {
				return newPositionalTypeErrorForGoObj("wasm_get_functions", 1, "*wazm.Module", args[0])
			}
			funs := wazm.GetFunctions(wm.Value)
			l := &object.List{
				Elements: make([]object.Object, len(funs)),
			}
			for i, fun := range funs {
				l.Elements[i] = &object.Stringo{Value: fun}
			}
			return l
		},
		HelpStr: helpStrArgs{
			explanation: "`wasm_get_functions` returns the available functions on the wasm module and works closely with wasm_get_exported_functions",
			signature:   "wasm_get_functions(mod: GoObj[*wazm.Module])",
			errors:      "InvalidArgCount,PositionalType",
			example:     "wasm_get_functions(add_mod) => ['realloc', '_start', 'add', 'asyncify_start_unwind', 'asyncify_stop_unwind', 'asyncify_start_rewind', 'free', 'calloc', 'asyncify_stop_rewind', 'malloc', 'asyncify_get_state']",
		}.String(),
	},
	"_wasm_get_exported_function": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("wasm_get_exported_function", len(args), 2, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("wasm_get_exported_function", 1, object.GO_OBJ, args[0].Type())
			}
			wm, ok := args[0].(*object.GoObj[*wazm.Module])
			if !ok {
				return newPositionalTypeErrorForGoObj("wasm_get_exported_function", 1, "*wazm.Module", args[0])
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("wasm_get_exported_function", 2, object.STRING_OBJ, args[1].Type())
			}
			fnName := args[1].(*object.Stringo).Value
			if _, ok := wm.Value.Module.ExportedFunctions()[fnName]; !ok {
				return newError("`wasm_get_exported_function` error: function '%s' not found", fnName)
			}
			// Return a builtin function to call the wasm function
			return &object.Builtin{
				Fun: func(args ...object.Object) object.Object {
					argsForCall := make([]uint64, len(args))
					for i, arg := range args {
						if arg.Type() != object.UINTEGER_OBJ {
							return newPositionalTypeError("wasm_call", i+1, object.UINTEGER_OBJ, arg.Type())
						}
						argsForCall[i] = arg.(*object.UInteger).Value
					}
					var mod api.Module
					// TODO: Figure out timeout stuff
					// if wm.Value.CancelFun != nil {
					// 	defer wm.Value.CancelFun()
					// }
					if !wm.Value.IsInstantiated {
						module, _, err := wazm.WazmRun(wm.Value)
						if err != nil {
							return newError("`wasm_call` error: instantiating failed %s", err.Error())
						}
						wm.Value = module
						mod = wm.Value.ApiMod
					} else {
						mod = wm.Value.ApiMod
					}
					fn := mod.ExportedFunction(fnName)
					var err error
					var retVal []uint64
					if len(argsForCall) == 0 {
						retVal, err = fn.Call(wm.Value.Ctx)
					} else {
						retVal, err = fn.Call(wm.Value.Ctx, argsForCall...)
					}
					if err != nil {
						return newError("`wasm_call` error: calling '%s' failed with params %v. %s", fnName, argsForCall, err.Error())
					}
					returnValue := &object.List{
						Elements: make([]object.Object, len(retVal)),
					}
					for i, e := range retVal {
						returnValue.Elements[i] = &object.UInteger{Value: e}
					}
					return returnValue
				},
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`wasm_get_exported_functions` returns the available function on the wasm module to be callable (via a BUILTIN) and works closely with wasm_get_functions",
			signature:   "wasm_get_exported_functions(mod: GoObj[*wazm.Module], func: str) -> (fn(any...) -> any)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "wasm_get_exported_functions(add_mod, 'add')(0x3, 0x7) => 0u10",
		}.String(),
	},
	"_wasm_run": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("wasm_run", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("wasm_run", 1, object.GO_OBJ, args[0].Type())
			}
			wm, ok := args[0].(*object.GoObj[*wazm.Module])
			if !ok {
				return newPositionalTypeErrorForGoObj("wasm_run", 1, "*wazm.Module", args[0])
			}
			if wm.Value.CancelFun != nil {
				defer wm.Value.CancelFun()
			}
			defer wm.Value.Runtime.Close(wm.Value.Ctx)
			module, rc, err := wazm.WazmRun(wm.Value)
			if err != nil {
				return newError("`wasm_run` error: %s", err.Error())
			}
			wm.Value = module
			return &object.Integer{Value: int64(rc)}
		},
		HelpStr: helpStrArgs{
			explanation: "`wasm_run` runs the main or _start of the wasm module and returns its return code as an integer",
			signature:   "wasm_run(mod: GoObj[*wazm.Module]) -> int",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "wasm_run(cat_mod) => 0 (side-effects may happen such as writing to stdout)",
		}.String(),
	},
	"_wasm_close": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("wasm_close", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("wasm_close", 1, object.GO_OBJ, args[0].Type())
			}
			wm, ok := args[0].(*object.GoObj[*wazm.Module])
			if !ok {
				return newPositionalTypeErrorForGoObj("wasm_close", 1, "*wazm.Module", args[0])
			}
			err := wm.Value.Runtime.Close(wm.Value.Ctx)
			if err != nil {
				return newError("`wasm_close` error: %s", err.Error())
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`wasm_close` closes the wasm module and disposes of the resource, currently if an error occurs a string is returned with the error",
			signature:   "wasm_close(mod: GoObj[*wazm.Module]) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "wasm_close(cat_mod) => null",
		}.String(),
	},
})
