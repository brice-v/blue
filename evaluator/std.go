package evaluator

import (
	"blue/consts"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"database/sql"
	"embed"
	"fmt"
	"hash"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/antchfx/htmlquery"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/websocket/v2"
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/ini"
	"github.com/gookit/config/v2/properties"
	"github.com/gookit/config/v2/toml"
	"github.com/gookit/config/v2/yamlv3"
	"github.com/gookit/ini/v2/dotenv"
	ws "github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

// StdModFileAndBuiltins keeps the file and builtins together for each std lib module
type StdModFileAndBuiltins struct {
	File     string         // File is the actual code used for the module
	Builtins BuiltinMapType // Builtins is the map of functions to be used by the module
}

//go:embed std/*
var stdFs embed.FS

func readStdFileToString(fname string) string {
	bs, err := stdFs.ReadFile("std/" + fname)
	if err != nil {
		panic(err)
	}
	return string(bs)
}

var _std_mods = map[string]StdModFileAndBuiltins{
	"http":   {File: readStdFileToString("http.b"), Builtins: _http_builtin_map},
	"time":   {File: readStdFileToString("time.b"), Builtins: _time_builtin_map},
	"search": {File: readStdFileToString("search.b"), Builtins: _search_builtin_map},
	"db":     {File: readStdFileToString("db.b"), Builtins: _db_builtin_map},
	"math":   {File: readStdFileToString("math.b"), Builtins: _math_builtin_map},
	"config": {File: readStdFileToString("config.b"), Builtins: _config_builtin_map},
	"crypto": {File: readStdFileToString("crypto.b"), Builtins: _crypto_builtin_map},
	"net":    {File: readStdFileToString("net.b"), Builtins: _net_builtin_map},
	"ui":     {File: readStdFileToString("ui.b"), Builtins: _ui_builtin_map},
}

func (e *Evaluator) IsStd(name string) bool {
	_, ok := _std_mods[name]
	return ok
}

func (e *Evaluator) AddStdLibToEnv(name string) {
	if !e.IsStd(name) {
		fmt.Printf("AddStdLibToEnv: '%s' is not in std lib map\n", name)
		os.Exit(1)
	}
	fb := _std_mods[name]
	l := lexer.New(fb.File, "<std/"+name+".b>")
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, msg := range p.Errors() {
			fmt.Printf("ParserError in `%s` module: %s\n", name, msg)
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
			buf.WriteString(l.GetErrorLineMessage(newE.ErrorTokens.PopBack()))
			buf.WriteByte('\n')
		}
		fmt.Printf("EvaluatorError in `%s` module: %s", name, buf.String())
		os.Exit(1)
	}
	mod := &object.Module{Name: name, Env: newE.env}
	e.env.Set(name, mod)
}

// Note: Look at how we import the get function in http.b
var _http_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_fetch": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 4 {
				return newError("`fetch` expects 4 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument 1 to `fetch` should be STRING. got=%s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newError("argument 2 to `fetch` should be STRING. got=%s", args[1].Type())
			}
			if args[2].Type() != object.MAP_OBJ {
				return newError("argument 3 to `fetch` should be MAP. got=%s", args[2].Type())
			}
			if args[3].Type() != object.NULL_OBJ && args[3].Type() != object.STRING_OBJ {
				return newError("argument 4 to `fetch` should be NULL or STRING. got=%s", args[3].Type())
			}
			resource := args[0].(*object.Stringo).Value
			method := args[1].(*object.Stringo).Value
			headersMap := args[2].(*object.Map).Pairs
			var body io.Reader
			if args[3].Type() == object.NULL_OBJ {
				body = nil
			} else {
				body = strings.NewReader(args[3].(*object.Stringo).Value)
			}
			request, err := http.NewRequest(method, resource, body)
			if err != nil {
				return newError("`fetch` error: %s", err.Error())
			}
			// Add User Agent always and then it can be overwritten
			request.Header.Add("user-agent", "blue/v"+consts.VERSION)
			for _, k := range headersMap.Keys {
				mp, _ := headersMap.Get(k)
				if key, ok := mp.Key.(*object.Stringo); ok {
					if val, ok := mp.Value.(*object.Stringo); ok {
						request.Header.Add(key.Value, val.Value)
					}
				}
			}
			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				return newError("`fetch` error: %s", err.Error())
			}
			defer resp.Body.Close()
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return newError("`fetch` error: %s", err.Error())
			}
			return &object.Stringo{Value: string(respBody)}
		},
	},
	"_get": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`get` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument to `get` must be STRING. got=%s", args[0].Type())
			}
			resp, err := http.Get(args[0].(*object.Stringo).Value)
			if err != nil {
				return newError("`get` failed: %s", err.Error())
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return newError("`get` failed: %s", err.Error())
			}
			return &object.Stringo{Value: string(body)}
		},
	},
	"_post": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("`post` expects 3 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument 1 to `post` must be STRING. got=%s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newError("argument 2 to `post` must be STRING. got=%s", args[1].Type())
			}
			if args[2].Type() != object.STRING_OBJ {
				return newError("argument 3 to `post` must be STRING. got=%s", args[2].Type())
			}
			urlInput := args[0].(*object.Stringo).Value
			mimeTypeInput := args[1].(*object.Stringo).Value
			bodyInput := args[2].(*object.Stringo).Value
			resp, err := http.Post(urlInput, mimeTypeInput, strings.NewReader(bodyInput))
			if err != nil {
				return newError("`post` failed: %s", err.Error())
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return newError("`post` failed: %s", err.Error())
			}
			return &object.Stringo{Value: string(body)}
		},
	},
	"_new_server": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newError("`new_server` expects 0 args. got=%d", len(args))
			}
			var disableStartupDebug bool
			disableStartupMessageStr := os.Getenv("DISABLE_HTTP_SERVER_DEBUG")
			disableStartupDebug, err := strconv.ParseBool(disableStartupMessageStr)
			if err != nil {
				disableStartupDebug = false
			}
			app := fiber.New(fiber.Config{
				Immutable:             true,
				EnablePrintRoutes:     !disableStartupDebug,
				DisableStartupMessage: disableStartupDebug,
			})
			curServer := serverCount.Add(1)
			ServerMap.Put(curServer, app)
			return &object.UInteger{Value: curServer}
		},
	},
	"_serve": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("`serve` expects 2 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newError("argument 1 to `serve` should be UINTEGER. got=%s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newError("argument 2 to `serve` should be STRING. got=%s", args[1].Type())
			}
			app, ok := ServerMap.Get(args[0].(*object.UInteger).Value)
			if !ok {
				return newError("`serve` could not find Server Object")
			}
			addrPort := args[1].(*object.Stringo).Value
			// nil here means use the default server mux (ie. things that were http.HandleFunc's)
			err := app.Listen(addrPort)
			if err != nil {
				return newError("`serve` error: %s", err.Error())
			}
			return NULL
		},
	},
	"_static": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 4 {
				return newError("`static` expects 4 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newError("argument 1 to `static` should be UINTEGER. got=%s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newError("argument 2 to `static` should be STRING. got=%s", args[1].Type())
			}
			if args[2].Type() != object.STRING_OBJ {
				return newError("argument 3 to `static` should be STRING. got=%s", args[2].Type())
			}
			if args[3].Type() != object.BOOLEAN_OBJ {
				return newError("argument 4 to `static` should be BOOLEAN. got=%s", args[3].Type())
			}
			app, ok := ServerMap.Get(args[0].(*object.UInteger).Value)
			if !ok {
				return newError("`static` could not find Server Object")
			}
			prefix := args[1].(*object.Stringo).Value
			fpath := args[2].(*object.Stringo).Value
			shouldBrowse := args[3].(*object.Boolean).Value
			app.Static(prefix, fpath, fiber.Static{
				Browse: shouldBrowse,
			})
			return NULL
		},
	},
	"_ws_send": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("`ws_send` expects 2 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newError("argument 1 to `ws_send` should be UINTEGER. got=%s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ && args[1].Type() != object.BYTES_OBJ {
				return newError("argument 2 to `ws_send` should be STRING or BYTES. got=%s", args[1].Type())
			}
			connId := args[0].(*object.UInteger).Value
			c, ok := WSConnMap.Get(connId)
			if !ok {
				return newError("`ws_send` could not find ws object")
			}
			var err error
			if args[1].Type() == object.STRING_OBJ {
				err = c.WriteMessage(websocket.TextMessage, []byte(args[1].(*object.Stringo).Value))
			} else {
				err = c.WriteMessage(websocket.BinaryMessage, args[1].(*object.Bytes).Value)
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
				return newError("`ws_recv` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newError("argument 1 to `ws_recv` should be UINTEGER. got=%s", args[0].Type())
			}
			connId := args[0].(*object.UInteger).Value
			c, ok := WSConnMap.Get(connId)
			if !ok {
				return newError("`ws_recv` could not find ws object")
			}
			mt, msg, err := c.ReadMessage()
			if err != nil {
				// Remove this conn
				WSConnMap.Remove(connId)
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
				// Remove the conn
				WSConnMap.Remove(connId)
				// If its closed we still want to return an error so that the handler fn wont try to send NULL
				return newError("`ws_recv`: websocket closed.")
			}
		},
	},
	"_new_ws": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`new_ws` expects 1 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument 1 to `new_ws` should be STRING. got=%s", args[0].Type())
			}
			url := args[0].(*object.Stringo).Value

			conn, _, err := ws.DefaultDialer.Dial(url, nil)
			if err != nil {
				return newError("`new_ws` error: %s", err.Error())
			}
			// log.Printf("resp = %#v", resp)
			connId := wsClientConnCount.Add(1)
			WSClientConnMap.Put(connId, conn)
			return object.CreateBasicMapObject("ws/client", connId)
		},
	},
	"_ws_client_send": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("`ws_send` expects 2 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newError("argument 1 to `ws_send` should be UINTEGER. got=%s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ && args[1].Type() != object.BYTES_OBJ {
				return newError("argument 2 to `ws_send` should be STRING or BYTES. got=%s", args[1].Type())
			}
			connId := args[0].(*object.UInteger).Value
			c, ok := WSClientConnMap.Get(connId)
			if !ok {
				return newError("`ws_send` could not find ws object")
			}
			var err error
			if args[1].Type() == object.STRING_OBJ {
				err = c.WriteMessage(websocket.TextMessage, []byte(args[1].(*object.Stringo).Value))
			} else {
				err = c.WriteMessage(websocket.BinaryMessage, args[1].(*object.Bytes).Value)
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
				return newError("`ws_recv` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newError("argument 1 to `ws_recv` should be UINTEGER. got=%s", args[0].Type())
			}
			connId := args[0].(*object.UInteger).Value
			c, ok := WSClientConnMap.Get(connId)
			if !ok {
				return newError("`ws_recv` could not find ws object")
			}
			mt, msg, err := c.ReadMessage()
			if err != nil {
				// Remove this conn
				WSClientConnMap.Remove(connId)
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
				// Remove the conn
				WSClientConnMap.Remove(connId)
				// If its closed we still want to return an error so that the handler fn wont try to send NULL
				return newError("`ws_recv`: websocket closed.")
			}
		},
	},
	"_handle_monitor": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("`handle_monitor` expects 3 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newError("argument 1 to `handle_monitor` should be UINTEGER. got=%s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newError("argument 2 to `handle_monitor` should be STRING. got=%s", args[1].Type())
			}
			if args[2].Type() != object.BOOLEAN_OBJ {
				return newError("argument 3 to `handle_monitor` should be BOOLEAN got=%s", args[2].Type())
			}
			serverId := args[0].(*object.UInteger).Value
			path := args[1].(*object.Stringo).Value
			shouldShow := args[2].(*object.Boolean).Value
			app, ok := ServerMap.Get(serverId)
			if !ok {
				return newError("`handle_monitor` could not find server object")
			}
			app.Get(path, monitor.New(monitor.Config{
				Next: func(c *fiber.Ctx) bool {
					return !shouldShow
				},
			}))
			return NULL
		},
	},
})

var _time_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_sleep": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`sleep` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newError("argument to `sleep` must be INTEGER, got=%s", args[0].Type())
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
				return newError("`now` expects 0 arguments. got=%d", len(args))
			}
			return &object.Integer{Value: time.Now().UnixNano()}
		},
	},
})

var _search_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_by_xpath": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("`by_xpath` expects 3 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument 1 to `by_xpath` should be STRING. got=%s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newError("argument 2 to `by_xpath` should be STRING. got=%s", args[1].Type())
			}
			if args[2].Type() != object.BOOLEAN_OBJ {
				return newError("argument 3 to `by_xpath` should be BOOLEAN. got=%s", args[2].Type())
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
				listToReturn := &object.List{Elements: make([]object.Object, 0)}
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
			return NULL
		},
	},
})

var _db_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_db_open": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`open` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument 1 to `open` should be STRING. got=%s", args[0].Type())
			}
			dbName := args[0].(*object.Stringo).Value
			if dbName == "" {
				return newError("`open` error: db_name argument is empty")
			}
			db, err := sql.Open("sqlite", dbName)
			if err != nil {
				return newError("`open` error: %s", err.Error())
			}
			curDB := dbCount.Add(1)
			DBMap.Put(curDB, db)
			return object.CreateBasicMapObject("db", curDB)
		},
	},
	"_db_ping": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`ping` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newError("argument 1 to `ping` should be UINTEGER. got=%s", args[0].Type())
			}
			i := args[0].(*object.UInteger).Value
			if db, ok := DBMap.Get(i); ok {
				err := db.Ping()
				if err != nil {
					return &object.Stringo{Value: err.Error()}
				}
				return NULL
			}
			return newError("`ping` error: DB not found")
		},
	},
	"_db_close": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`close` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newError("argument 1 to `close` should be UINTEGER. got=%s", args[0].Type())
			}
			i := args[0].(*object.UInteger).Value
			if db, ok := DBMap.Get(i); ok {
				err := db.Close()
				if err != nil {
					return newError("`close` error: %s", err.Error())
				}
				DBMap.Remove(i)
				return NULL
			}
			return newError("`close` error: DB not found")
		},
	},
	"_db_exec": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("`exec` expects 3 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newError("argument 1 to `exec` should be UINTEGER. got=%s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newError("argument 2 to `exec` should be STRING. got=%s", args[1].Type())
			}
			if args[2].Type() != object.LIST_OBJ {
				return newError("argument 3 to `exec` should be LIST. got=%s", args[2].Type())
			}
			i := args[0].(*object.UInteger).Value
			s := args[1].(*object.Stringo).Value
			l := args[2].(*object.List).Elements
			if db, ok := DBMap.Get(i); ok {
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
							return newError("argument list to `exec` included invalid type. got=%s", e.Type())
						}
					}
					result, err = db.Exec(s, execArgs...)
				} else {
					result, err = db.Exec(s)
				}
				if err != nil {
					return newError("`exec` error: %s", err.Error())
				}
				lastInsertId, err := result.LastInsertId()
				if err != nil {
					return newError("`exec` error: %s", err.Error())
				}
				rowsAffected, err := result.RowsAffected()
				if err != nil {
					return newError("`exec` error: %s", err.Error())
				}
				mapToConvert := object.NewOrderedMap[string, object.Object]()
				mapToConvert.Set("last_insert_id", &object.Integer{Value: lastInsertId})
				mapToConvert.Set("rows_affected", &object.Integer{Value: rowsAffected})
				return object.CreateMapObjectForGoMap(*mapToConvert)
			}
			return newError("`exec` error: DB not found")
		},
	},
	"_db_query": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 4 {
				return newError("`query` expects 4 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newError("argument 1 to `query` should be UINTEGER. got=%s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newError("argument 2 to `query` should be STRING. got=%s", args[1].Type())
			}
			if args[2].Type() != object.LIST_OBJ {
				return newError("argument 3 to `query` should be LIST. got=%s", args[2].Type())
			}
			if args[3].Type() != object.BOOLEAN_OBJ {
				return newError("argument 4 to `query` should be BOOLEAN. got=%s", args[3].Type())
			}
			i := args[0].(*object.UInteger).Value
			s := args[1].(*object.Stringo).Value
			l := args[2].(*object.List).Elements
			isNamedCols := args[3].(*object.Boolean).Value
			if db, ok := DBMap.Get(i); ok {
				var rows *sql.Rows
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
							return newError("argument list to `query` included invalid type. got=%s", e.Type())
						}
					}
					rows, err = db.Query(s, execArgs...)
				} else {
					rows, err = db.Query(s)
				}
				defer rows.Close()
				if err != nil {
					return newError("`query` error: %s", err.Error())
				}
				colNames, err := rows.Columns()
				if err != nil {
					return newError("`query` error: %s", err.Error())
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
						return newError("`query` error: %s", err.Error())
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
			}
			return newError("`exec` error: DB not found")
		},
	},
})

var _math_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	// "_abs": {},
	// TODO: Do we need to / want to support BigFloat/BigDecimal as well? with whatever is there
	// func Abs(x float64) float64
	// func Acos(x float64) float64
	// func Acosh(x float64) float64
	// func Asin(x float64) float64
	// func Asinh(x float64) float64
	// func Atan(x float64) float64
	// func Atan2(y, x float64) float64
	// func Atanh(x float64) float64
	// func Cbrt(x float64) float64
	// func Ceil(x float64) float64
	// func Copysign(f, sign float64) float64
	// func Cos(x float64) float64
	// func Cosh(x float64) float64
	// func Dim(x, y float64) float64
	// func Erf(x float64) float64
	// func Erfc(x float64) float64
	// func Erfcinv(x float64) float64
	// func Erfinv(x float64) float64
	// func Exp(x float64) float64
	// func Exp2(x float64) float64
	// func Expm1(x float64) float64
	// func FMA(x, y, z float64) float64
	// func Float32bits(f float32) uint32
	// func Float32frombits(b uint32) float32
	// func Float64bits(f float64) uint64
	// func Float64frombits(b uint64) float64
	// func Floor(x float64) float64
	// func Frexp(f float64) (frac float64, exp int)
	// func Gamma(x float64) float64
	// func Hypot(p, q float64) float64
	// func Ilogb(x float64) int
	// func Inf(sign int) float64
	// func IsInf(f float64, sign int) bool
	// func IsNaN(f float64) (is bool)
	// func J0(x float64) float64
	// func J1(x float64) float64
	// func Jn(n int, x float64) float64
	// func Ldexp(frac float64, exp int) float64
	// func Lgamma(x float64) (lgamma float64, sign int)
	// func Log(x float64) float64
	// func Log10(x float64) float64
	// func Log1p(x float64) float64
	// func Log2(x float64) float64
	// func Logb(x float64) float64
	// func Max(x, y float64) float64
	// func Min(x, y float64) float64
	// func Mod(x, y float64) float64
	// func Modf(f float64) (int float64, frac float64)
	// func NaN() float64
	// func Nextafter(x, y float64) (r float64)
	// func Nextafter32(x, y float32) (r float32)
	// func Pow(x, y float64) float64
	// func Pow10(n int) float64
	// func Remainder(x, y float64) float64
	// func Round(x float64) float64
	// func RoundToEven(x float64) float64
	// func Signbit(x float64) bool
	// func Sin(x float64) float64
	// func Sincos(x float64) (sin, cos float64)
	// func Sinh(x float64) float64
	// func Sqrt(x float64) float64
	// func Tan(x float64) float64
	// func Tanh(x float64) float64
	// func Trunc(x float64) float64
	// func Y0(x float64) float64
	// func Y1(x float64) float64
	// func Yn(n int, x float64) float64
})

var _config_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_load_file": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`load_file` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument 1 to `load_file` should be STRING. got=%s", args[0].Type())
			}
			c := config.NewWithOptions("config", config.ParseEnv, config.Readonly)
			c.WithDriver(yamlv3.Driver, ini.Driver, toml.Driver, properties.Driver)
			fpath := args[0].(*object.Stringo).Value
			err := c.LoadFiles(fpath)
			if err != nil {
				if err.Error() == "not exists or not register decoder for the format: env" {
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
				return newError("`dump_config` expects 3 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument 1 to `dump_config` should be STRING. got=%s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newError("argument 2 to `dump_config` should be STRING. got=%s", args[1].Type())
			}
			if args[2].Type() != object.STRING_OBJ {
				return newError("argument 3 to `dump_config` should be STRING. got=%s", args[2].Type())
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
				return newError("`sha` expects 2 arguments. got=%d", len(args))
			}
			// TODO: Support Bytes object?
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument 1 to `sha` should be STRING. got=%s", args[0].Type())
			}
			if args[1].Type() != object.INTEGER_OBJ {
				return newError("argument 2 to `sha` should be INTEGER. got=%s", args[1].Type())
			}
			s := args[0].(*object.Stringo).Value
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
			hasher.Write([]byte(s))
			return &object.Stringo{Value: fmt.Sprintf("%x", hasher.Sum(nil))}
		},
	},
	"_md5": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`md5` expects 1 argument. got=%d", len(args))
			}
			// TODO: Support Bytes object?
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument 1 to `sha` should be STRING. got=%s", args[0].Type())
			}
			s := args[0].(*object.Stringo).Value
			hasher := md5.New()
			hasher.Write([]byte(s))
			return &object.Stringo{Value: fmt.Sprintf("%x", hasher.Sum(nil))}
		},
	},
	"_generate_from_password": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`generate_from_password` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument 1 to `generate_from_password` should be STRING. got=%s", args[0].Type())
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
				return newError("`compare_hash_and_password` expects 2 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument 1 to `compare_hash_and_password` should be STRING. got=%s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newError("argument 2 to `compare_hash_and_password` should be STRING. got=%s", args[1].Type())
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
})

var _net_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_connect": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("`connect` expects 3 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument 1 to `connect` should be STRING. got=%s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newError("argument 2 to `connect` should be STRING. got=%s", args[1].Type())
			}
			if args[2].Type() != object.STRING_OBJ {
				return newError("argument 3 to `connect` should be STRING. got=%s", args[2].Type())
			}
			transport := strings.ToLower(args[0].(*object.Stringo).Value)
			addr := args[1].(*object.Stringo).Value
			port := args[2].(*object.Stringo).Value
			addrStr := fmt.Sprintf("%s:%s", addr, port)
			conn, err := net.Dial(transport, addrStr)
			if err != nil {
				return newError("`connect` error: %s", err.Error())
			}
			curConn := netConnCount.Add(1)
			NetConnMap.Put(curConn, conn)
			return object.CreateBasicMapObject("net", curConn)
		},
	},
	"_listen": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("`listen` expects 3 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument 1 to `listen` should be STRING. got=%s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newError("argument 2 to `listen` should be STRING. got=%s", args[1].Type())
			}
			if args[2].Type() != object.STRING_OBJ {
				return newError("argument 3 to `listen` should be STRING. got=%s", args[2].Type())
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
				curServer := netUDPServerCount.Add(1)
				NetUDPServerMap.Put(curServer, l)
				return object.CreateBasicMapObject("net/udp", curServer)
			}
			l, err := net.Listen(transport, addrStr)
			if err != nil {
				return newError("`listen` error: %s", err.Error())
			}
			curServer := netTCPServerCount.Add(1)
			NetTCPServerMap.Put(curServer, l)
			return object.CreateBasicMapObject("net/tcp", curServer)
		},
	},
	"_accept": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`accept` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newError("argument 1 to `accept` should be UINTEGER. got=%s", args[0].Type())
			}
			l, ok := NetTCPServerMap.Get(args[0].(*object.UInteger).Value)
			if !ok {
				return newError("`accept` error: listener not found")
			}
			conn, err := l.Accept()
			if err != nil {
				return newError("`accept` error: %s", err.Error())
			}
			curConn := netConnCount.Add(1)
			NetConnMap.Put(curConn, conn)
			return object.CreateBasicMapObject("net", curConn)
		},
	},
	"_net_close": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("`net_close` expects 2 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newError("argument 1 to `net_close` should be UINTEGER. got=%s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newError("argument 2 to `net_close` should be STRING. got=%s", args[1].Type())
			}
			id := args[0].(*object.UInteger).Value
			t := args[1].(*object.Stringo).Value
			switch t {
			case "net/udp":
				c, ok := NetUDPServerMap.Get(id)
				if !ok {
					return NULL
				}
				c.Close()
				NetUDPServerMap.Remove(id)
			case "net/tcp":
				listener, ok := NetTCPServerMap.Get(id)
				if !ok {
					// Dont error out if were just trying to close
					return NULL
				}
				listener.Close()
				NetTCPServerMap.Remove(id)
			case "net":
				conn, ok := NetConnMap.Get(id)
				if !ok {
					// Dont error out if were just trying to close
					return NULL
				}
				conn.Close()
				NetConnMap.Remove(id)
			default:
				return newError("`net_close` expects type of 'net/tcp', 'net/udp', or 'net'")
			}
			return NULL
		},
	},
	"_net_read": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("`net_read` expects 2 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newError("argument 1 to `net_read` should be UINTEGER. got=%s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newError("argument 2 to `net_read` should be STRING. got=%s", args[1].Type())
			}
			connId := args[0].(*object.UInteger).Value
			t := args[1].(*object.Stringo).Value
			var conn net.Conn
			if t == "net/udp" {
				c, ok := NetUDPServerMap.Get(connId)
				if !ok {
					return newError("`net_read` udp error: connection not found")
				}
				conn = c
			} else {
				c, ok := NetConnMap.Get(connId)
				if !ok {
					return newError("`net_read` error: connection not found")
				}
				conn = c
			}
			s, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				return newError("`net_read` error: %s", err.Error())
			}
			return &object.Stringo{Value: s[:len(s)-1]}
		},
	},
	"_net_write": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("`net_write` expects 3 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newError("argument 1 to `net_write` should be UINTEGER. got=%s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newError("argument 2 to `net_write` should be STRING. got=%s", args[1].Type())
			}
			// TODO: Support bytes object (but then we cant always use '\n' by default)
			if args[2].Type() != object.STRING_OBJ {
				return newError("argument 3 to `net_write` should be STRING. got=%s", args[2].Type())
			}
			connId := args[0].(*object.UInteger).Value
			t := args[1].(*object.Stringo).Value
			var conn net.Conn
			if t == "net/udp" {
				c, ok := NetUDPServerMap.Get(connId)
				if !ok {
					return newError("`net_write` udp error: connection not found")
				}
				conn = c
			} else {
				c, ok := NetConnMap.Get(connId)
				if !ok {
					return newError("`net_write` error: connection not found")
				}
				conn = c
			}
			s := args[2].(*object.Stringo).Value
			bs := []byte(s)
			n, err := conn.Write(bs)
			if n != len(bs) {
				return newError("`net_write` error: did not write the entire string")
			}
			if err != nil {
				return newError("`net_write` error: %s", err.Error())
			}
			// If the string contains a \n its going to break off anyways
			// TODO: FIXME - allow user to decide there cutoff byte? or
			if !strings.Contains(s, "\n") {
				n, err = conn.Write([]byte{'\n'})
				if n != 1 {
					return newError("`net_write` error: did not write the last byte")
				}
				if err != nil {
					return newError("`net_write` error: %s", err.Error())
				}
			}
			return NULL
		},
	},
	"_inspect": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("`inspect` expects 2 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newError("argument 1 to `inspect` should be UINTEGER. got=%s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newError("argument 2 to `inspect` should be STRING. got=%s", args[1].Type())
			}
			id := args[0].(*object.UInteger).Value
			t := args[1].(*object.Stringo).Value
			switch t {
			case "net/udp":
				c, ok := NetUDPServerMap.Get(id)
				if !ok {
					return newError("`inspect` udp server connection not found")
				}
				mapObj := object.NewOrderedMap[string, object.Object]()
				mapObj.Set("remote_addr", &object.Stringo{Value: c.RemoteAddr().String()})
				mapObj.Set("local_addr", &object.Stringo{Value: c.LocalAddr().String()})
				mapObj.Set("remote_addr_network", &object.Stringo{Value: c.RemoteAddr().Network()})
				mapObj.Set("local_addr_network", &object.Stringo{Value: c.LocalAddr().Network()})
				return object.CreateMapObjectForGoMap(*mapObj)
			case "net/tcp":
				l, ok := NetTCPServerMap.Get(id)
				if !ok {
					return newError("`inspect` tcp server connection not found")
				}
				mapObj := object.NewOrderedMap[string, object.Object]()
				mapObj.Set("addr", &object.Stringo{Value: l.Addr().String()})
				mapObj.Set("addr_network", &object.Stringo{Value: l.Addr().Network()})
				return object.CreateMapObjectForGoMap(*mapObj)
			case "net":
				c, ok := NetConnMap.Get(id)
				if !ok {
					return newError("`inspect` connection not found")
				}
				mapObj := object.NewOrderedMap[string, object.Object]()
				mapObj.Set("remote_addr", &object.Stringo{Value: c.RemoteAddr().String()})
				mapObj.Set("local_addr", &object.Stringo{Value: c.LocalAddr().String()})
				mapObj.Set("remote_addr_network", &object.Stringo{Value: c.RemoteAddr().Network()})
				mapObj.Set("local_addr_network", &object.Stringo{Value: c.LocalAddr().Network()})
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
				return newError("`new_app` expects 0 arguments. got=%d", len(args))
			}
			app := app.New()
			appId := uiAppCount.Add(1)
			UIAppMap.Put(appId, app)
			return &object.UInteger{Value: appId}
		},
	},
	"_window": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 5 {
				return newError("`window` expects 5 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newError("argument 1 to `window` should be UINTEGER. got=%s", args[0].Type())
			}
			if args[1].Type() != object.INTEGER_OBJ {
				return newError("argument 2 to `window` should be INTEGER. got=%s", args[1].Type())
			}
			if args[2].Type() != object.INTEGER_OBJ {
				return newError("argument 3 to `window` should be INTEGER. got=%s", args[2].Type())
			}
			if args[3].Type() != object.STRING_OBJ {
				return newError("argument 4 to `window` should be STRING. got=%s", args[3].Type())
			}
			if args[4].Type() != object.UINTEGER_OBJ {
				return newError("argument 5 to `window` should be UINTEGER. got=%s", args[4].Type())
			}
			appId := args[0].(*object.UInteger).Value
			width := args[1].(*object.Integer).Value
			height := args[2].(*object.Integer).Value
			title := args[3].(*object.Stringo).Value
			contentId := args[4].(*object.UInteger).Value
			app, ok := UIAppMap.Get(appId)
			if !ok {
				return newError("`window` could not find app object")
			}
			content, ok := UICanvasObjectMap.Get(contentId)
			if !ok {
				return newError("`window` could not find content object")
			}
			w := app.NewWindow(title)
			w.Resize(fyne.Size{Width: float32(width), Height: float32(height)})
			w.SetContent(content)
			w.ShowAndRun()
			return NULL
		},
	},
	"_label": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`label` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument 1 to `label` should be STRING. got=%s", args[0].Type())
			}
			label := args[0].(*object.Stringo).Value
			labelId := uiCanvasObjectCount.Add(1)
			l := widget.NewLabel(label)
			UICanvasObjectMap.Put(labelId, l)
			return object.CreateBasicMapObject("ui", labelId)
		},
	},
	"_row": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`row` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.LIST_OBJ {
				return newError("argument 1 to `row` should be LIST. got=%s", args[0].Type())
			}
			elements := args[0].(*object.List).Elements
			canvasObjects := make([]fyne.CanvasObject, len(elements))
			for i, e := range elements {
				if e.Type() != object.UINTEGER_OBJ {
					return newError("`row` error: all children should be UINTEGER. found=%s", e.Type())
				}
				elemId := e.(*object.UInteger).Value
				o, ok := UICanvasObjectMap.Get(elemId)
				if !ok {
					return newError("`row` error: could not find ui element")
				}
				canvasObjects[i] = o
			}
			vboxId := uiCanvasObjectCount.Add(1)
			vbox := container.NewVBox(canvasObjects...)
			UICanvasObjectMap.Put(vboxId, vbox)
			return object.CreateBasicMapObject("ui", vboxId)
		},
	},
	"_col": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`col` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.LIST_OBJ {
				return newError("argument 1 to `col` should be LIST. got=%s", args[0].Type())
			}
			elements := args[0].(*object.List).Elements
			canvasObjects := make([]fyne.CanvasObject, len(elements))
			for i, e := range elements {
				if e.Type() != object.UINTEGER_OBJ {
					return newError("`col` error: all children should be UINTEGER. found=%s", e.Type())
				}
				elemId := e.(*object.UInteger).Value
				o, ok := UICanvasObjectMap.Get(elemId)
				if !ok {
					return newError("`col` error: could not find ui element")
				}
				canvasObjects[i] = o
			}
			vboxId := uiCanvasObjectCount.Add(1)
			vbox := container.NewHBox(canvasObjects...)
			UICanvasObjectMap.Put(vboxId, vbox)
			return object.CreateBasicMapObject("ui", vboxId)
		},
	},
	"_entry": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`entry` expects 1 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.BOOLEAN_OBJ {
				return newError("argument 1 to `entry` should be BOOLEAN. got=%s", args[0].Type())
			}
			isMultiline := args[0].(*object.Boolean).Value
			var entry *widget.Entry
			if isMultiline {
				entry = widget.NewMultiLineEntry()
			} else {
				entry = widget.NewEntry()
			}
			entryId := uiCanvasObjectCount.Add(1)
			UICanvasObjectMap.Put(entryId, entry)
			return object.CreateBasicMapObject("ui/entry", entryId)
		},
	},
	"_entry_get_text": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`entry_get_text` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newError("argument 1 to `entry_get_text` should be UINTEGER. got=%s", args[0].Type())
			}
			entryId := args[0].(*object.UInteger).Value
			entry, ok := UICanvasObjectMap.Get(entryId)
			if !ok {
				return newError("`entry_get_text` error: could not find ui element")
			}
			switch x := entry.(type) {
			case *widget.Entry:
				return &object.Stringo{Value: x.Text}
			default:
				return newError("`entry_get_text` error: entry id did not match entry")
			}
		},
	},
	"_append_form": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("`append_form` expects 3 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newError("argument 1 to `append_form` should be UINTEGER. got=%s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newError("argument 2 to `append_form` should be STRING. got=%s", args[1].Type())
			}
			if args[2].Type() != object.UINTEGER_OBJ {
				return newError("argument 3 to `append_form` should be UINTEGER. got=%s", args[2].Type())
			}
			formId := args[0].(*object.UInteger).Value
			maybeForm, ok := UICanvasObjectMap.Get(formId)
			if !ok {
				return newError("`append_form` error: form not found")
			}
			var form *widget.Form
			switch x := maybeForm.(type) {
			case *widget.Form:
				form = x
			default:
				return newError("`append_form` error: id used for form is not form. got=%T", x)
			}
			wId := args[2].(*object.UInteger).Value
			w, ok := UICanvasObjectMap.Get(wId)
			if !ok {
				return newError("`append_form` error: widget not found")
			}
			form.Append(args[1].(*object.Stringo).Value, w)
			return NULL
		},
	},
})
