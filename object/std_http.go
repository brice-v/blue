package object

import (
	"blue/consts"
	"blue/lib"
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"log"
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

	"github.com/gofiber/contrib/v3/monitor"
	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/encryptcookie"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/sse"
	"github.com/gofiber/fiber/v3/middleware/static"
	ws "github.com/gorilla/websocket"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/strikethrough"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
)

const fiberNetworkStateKey = "blue.network"

// Used to catch interrupt to shutdown server
var interruptCh = make(chan os.Signal, 1)

type sseStream interface {
	Event(event sse.Event) error
	LastEventID() string
}

type wsMessageWriter interface {
	WriteMessage(messageType int, data []byte) error
}

func sendWSMessage(funName string, w wsMessageWriter, value Object) Object {
	if value.Type() != STRING_OBJ && value.Type() != BYTES_OBJ {
		return newPositionalTypeError(funName, 2, "STRING or BYTES", value.Type())
	}
	var err error
	if value.Type() == STRING_OBJ {
		err = w.WriteMessage(ws.TextMessage, []byte(value.(*Stringo).Value))
	} else {
		err = w.WriteMessage(ws.BinaryMessage, value.(*Bytes).Value)
	}
	if err != nil {
		return newError("`%s` error: %s", funName, err.Error())
	}
	return NULL
}

func sendObject(obj, value Object) Object {
	if obj.Type() == PROCESS_OBJ {
		p, ok := obj.(*Process)
		if !ok {
			return newError("obj `%s` is invalid type. got=`%s` (%s)", obj.Inspect(), obj.Inspect(), obj.Type())
		}
		p.Ch <- value
		return NULL
	}
	if obj.Type() == GO_OBJ {
		if w, ok := obj.(*GoObj[*sse.Stream]); ok {
			return sendSSE(w.Value, "sse_send", 2, value, NULL, NULL, NULL)
		}
	}
	if obj.Type() == MAP_OBJ {
		m := obj.(*Map)
		if tObj, ok := getMapValueByString(m, "t"); ok {
			if t, ok := tObj.(*Stringo); ok {
				vObj, _ := getMapValueByString(m, "v")
				switch t.Value {
				case "ws":
					if c, ok := vObj.(*GoObj[*websocket.Conn]); ok {
						return sendWSMessage("ws_send", c.Value, value)
					}
				case "ws/client":
					if c, ok := vObj.(*GoObj[*ws.Conn]); ok {
						return sendWSMessage("ws_client_send", c.Value, value)
					}
				case "sse":
					if c, ok := vObj.(*GoObj[*sse.Stream]); ok {
						return sendSSE(c.Value, "sse_send", 2, value, NULL, NULL, NULL)
					}
				}
			}
		}
	}
	return newError("obj `%s` is invalid type. got=`%s` (%s)", obj.Inspect(), obj.Inspect(), obj.Type())
}

func getMapValueByString(m *Map, key string) (Object, bool) {
	pair, ok := m.Pairs.Get(HashKey{Type: STRING_OBJ, Value: HashObject(&Stringo{Value: key})})
	if !ok {
		return nil, false
	}
	return pair.Value, true
}

func getFiberAppArg(funName string, obj Object) (*fiber.App, Object) {
	if obj.Type() != GO_OBJ {
		return nil, newPositionalTypeError(funName, 1, GO_OBJ, obj.Type())
	}
	app, ok := obj.(*GoObj[*fiber.App])
	if !ok {
		return nil, newPositionalTypeErrorForGoObj(funName, 1, "*fiber.App", obj)
	}
	return app.Value, nil
}

func validateMapOptionKeys(funName string, m *Map, allowed map[string]struct{}) Object {
	for _, hashKey := range m.Pairs.Keys {
		pair, ok := m.Pairs.Get(hashKey)
		if !ok {
			continue
		}
		key, ok := pair.Key.(*Stringo)
		if !ok {
			return newError("`%s` error: option keys must be STRING. got=%s", funName, pair.Key.Type())
		}
		if _, ok := allowed[key.Value]; !ok {
			return newError("`%s` error: unknown option %q", funName, key.Value)
		}
	}
	return nil
}

func setStringOption(funName string, m *Map, key string, dst *string) Object {
	v, ok := getMapValueByString(m, key)
	if !ok {
		return nil
	}
	s, ok := v.(*Stringo)
	if !ok {
		return newError("`%s` error: option %q expects STRING. got=%s", funName, key, v.Type())
	}
	*dst = s.Value
	return nil
}

func setBoolOption(funName string, m *Map, key string, dst *bool) Object {
	v, ok := getMapValueByString(m, key)
	if !ok {
		return nil
	}
	b, ok := v.(*Boolean)
	if !ok {
		return newError("`%s` error: option %q expects BOOL. got=%s", funName, key, v.Type())
	}
	*dst = b.Value
	return nil
}

func setIntOption(funName string, m *Map, key string, dst *int) Object {
	v, ok := getMapValueByString(m, key)
	if !ok {
		return nil
	}
	i, ok := v.(*Integer)
	if !ok {
		return newError("`%s` error: option %q expects INT. got=%s", funName, key, v.Type())
	}
	*dst = int(i.Value)
	return nil
}

var helmetOptionKeys = map[string]struct{}{
	"xss_protection":                    {},
	"content_type_nosniff":              {},
	"x_frame_options":                   {},
	"content_security_policy":           {},
	"referrer_policy":                   {},
	"permission_policy":                 {},
	"cross_origin_embedder_policy":      {},
	"cross_origin_opener_policy":        {},
	"cross_origin_resource_policy":      {},
	"origin_agent_cluster":              {},
	"x_dns_prefetch_control":            {},
	"x_download_options":                {},
	"x_permitted_cross_domain_policies": {},
	"hsts_max_age":                      {},
	"hsts_exclude_subdomains":           {},
	"csp_report_only":                   {},
	"hsts_preload_enabled":              {},
}

func helmetConfigFromMap(m *Map) (helmet.Config, Object) {
	var cfg helmet.Config
	if errObj := validateMapOptionKeys("helmet", m, helmetOptionKeys); errObj != nil {
		return cfg, errObj
	}
	stringOptions := []struct {
		key string
		dst *string
	}{
		{"xss_protection", &cfg.XSSProtection},
		{"content_type_nosniff", &cfg.ContentTypeNosniff},
		{"x_frame_options", &cfg.XFrameOptions},
		{"content_security_policy", &cfg.ContentSecurityPolicy},
		{"referrer_policy", &cfg.ReferrerPolicy},
		{"permission_policy", &cfg.PermissionPolicy},
		{"cross_origin_embedder_policy", &cfg.CrossOriginEmbedderPolicy},
		{"cross_origin_opener_policy", &cfg.CrossOriginOpenerPolicy},
		{"cross_origin_resource_policy", &cfg.CrossOriginResourcePolicy},
		{"origin_agent_cluster", &cfg.OriginAgentCluster},
		{"x_dns_prefetch_control", &cfg.XDNSPrefetchControl},
		{"x_download_options", &cfg.XDownloadOptions},
		{"x_permitted_cross_domain_policies", &cfg.XPermittedCrossDomain},
	}
	for _, opt := range stringOptions {
		if errObj := setStringOption("helmet", m, opt.key, opt.dst); errObj != nil {
			return cfg, errObj
		}
	}
	boolOptions := []struct {
		key string
		dst *bool
	}{
		{"hsts_exclude_subdomains", &cfg.HSTSExcludeSubdomains},
		{"csp_report_only", &cfg.CSPReportOnly},
		{"hsts_preload_enabled", &cfg.HSTSPreloadEnabled},
	}
	for _, opt := range boolOptions {
		if errObj := setBoolOption("helmet", m, opt.key, opt.dst); errObj != nil {
			return cfg, errObj
		}
	}
	if errObj := setIntOption("helmet", m, "hsts_max_age", &cfg.HSTSMaxAge); errObj != nil {
		return cfg, errObj
	}
	if cfg.HSTSMaxAge < 0 {
		return cfg, newError("`helmet` error: hsts_max_age must be >= 0. got=%d", cfg.HSTSMaxAge)
	}
	if cfg.HSTSPreloadEnabled && cfg.HSTSExcludeSubdomains {
		return cfg, newError("`helmet` error: hsts_preload_enabled requires hsts_exclude_subdomains=false")
	}
	return cfg, nil
}

func validateCookieEncryptionKey(key string) error {
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return fmt.Errorf("failed to base64-decode key: %w", err)
	}
	switch len(decoded) {
	case 16, 24, 32:
		return nil
	default:
		return fmt.Errorf("key must decode to 16, 24, or 32 bytes. got=%d", len(decoded))
	}
}

func stringListArg(funName string, pos int, obj Object) ([]string, Object) {
	if obj.Type() != LIST_OBJ {
		return nil, newPositionalTypeError(funName, pos, LIST_OBJ, obj.Type())
	}
	list := obj.(*List)
	result := make([]string, 0, len(list.Elements))
	for i, e := range list.Elements {
		s, ok := e.(*Stringo)
		if !ok {
			return nil, newError("`%s` error: argument %d element %d must be STRING. got=%s", funName, pos, i+1, e.Type())
		}
		result = append(result, s.Value)
	}
	return result, nil
}

func getSSEStream(obj Object) (*sse.Stream, bool) {
	if w, ok := obj.(*GoObj[*sse.Stream]); ok {
		return w.Value, true
	}
	if obj.Type() != MAP_OBJ {
		return nil, false
	}
	m := obj.(*Map)
	tObj, ok := getMapValueByString(m, "t")
	if !ok {
		return nil, false
	}
	t, ok := tObj.(*Stringo)
	if !ok || t.Value != "sse" {
		return nil, false
	}
	vObj, ok := getMapValueByString(m, "v")
	if !ok {
		return nil, false
	}
	w, ok := vObj.(*GoObj[*sse.Stream])
	if !ok {
		return nil, false
	}
	return w.Value, true
}

func sseLastEventID(stream *sse.Stream) *Stringo {
	return &Stringo{Value: stream.LastEventID()}
}

// NewSSEConnection returns the map passed to a server-sent events handler.
func NewSSEConnection(stream *sse.Stream) *Map {
	m := NewOrderedMap[string, Object]()
	m.Set("t", &Stringo{Value: "sse"})
	m.Set("v", NewGoObj(stream))
	m.Set("last_event_id", &Builtin{
		Name: "last_event_id",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("last_event_id", len(args), 0, "")
			}
			return sseLastEventID(stream)
		},
		HelpStr: helpStrArgs{
			explanation: "`last_event_id` returns the Last-Event-ID sent by a reconnecting sse client, or an empty string",
			signature:   "last_event_id() -> str",
			errors:      "InvalidArgCount",
			example:     "sse.last_event_id() => '42'",
		}.String(),
	})
	return CreateMapObjectForGoMap(*m)
}

func sendSSE(stream sseStream, funName string, valuePos int, valueArg, eventArg, idArg, retryArg Object) Object {
	if valueArg.Type() != STRING_OBJ && valueArg.Type() != BYTES_OBJ {
		return newPositionalTypeError(funName, valuePos, "STRING or BYTES", valueArg.Type())
	}
	if eventArg.Type() != STRING_OBJ && eventArg.Type() != NULL_OBJ {
		return newPositionalTypeError(funName, valuePos+1, "STRING or NULL", eventArg.Type())
	}
	if idArg.Type() != STRING_OBJ && idArg.Type() != NULL_OBJ {
		return newPositionalTypeError(funName, valuePos+2, "STRING or NULL", idArg.Type())
	}
	if retryArg.Type() != INTEGER_OBJ && retryArg.Type() != NULL_OBJ {
		return newPositionalTypeError(funName, valuePos+3, "INTEGER or NULL", retryArg.Type())
	}
	var data []byte
	if valueArg.Type() == STRING_OBJ {
		data = []byte(valueArg.(*Stringo).Value)
	} else {
		data = valueArg.(*Bytes).Value
	}
	event := ""
	if eventArg.Type() == STRING_OBJ {
		event = eventArg.(*Stringo).Value
	}
	id := ""
	if idArg.Type() == STRING_OBJ {
		id = idArg.(*Stringo).Value
	}
	var retryDuration time.Duration
	if retryArg.Type() == INTEGER_OBJ {
		retry := retryArg.(*Integer).Value
		if retry < 0 {
			return newError("`sse_send` error: retry must be >= 0")
		}
		retryDuration = time.Duration(retry) * time.Millisecond
	}
	if err := stream.Event(sse.Event{
		Data:  data,
		ID:    id,
		Name:  event,
		Retry: retryDuration,
	}); err != nil {
		return newError("`sse_send` error: %s", err.Error())
	}
	return NULL
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
				return newPositionalTypeError("url_unescape", 1, STRING_OBJ, args[0].Type())
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
			app := fiber.New(fiber.Config{
				Immutable: true,
			})
			app.State().Set(fiberNetworkStateKey, network)
			return NewGoObj(app)
		},
		HelpStr: helpStrArgs{
			explanation: "`new_server` returns a new server object",
			signature:   "new_server(network: str('tcp','tcp4','tcp6')='tcp4') -> GoObj[*fiber.App]",
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
				app.Value.Use(static.New(".", static.Config{FS: sub}))
			}
			network := "tcp4"
			if v, ok := app.Value.State().Get(fiberNetworkStateKey); ok {
				network = v.(string)
			}
			var disableStartupDebug bool
			disableStartupMessageStr := os.Getenv(consts.BLUE_DISABLE_HTTP_SERVER_DEBUG)
			disableStartupDebug, err := strconv.ParseBool(disableStartupMessageStr)
			if err != nil {
				disableStartupDebug = false
			}
			err = app.Value.Listen(addrPort, fiber.ListenConfig{
				DisableStartupMessage: disableStartupDebug,
				EnablePrintRoutes:     !disableStartupDebug,
				ListenerNetwork:       network,
			})
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
				app.Value.Use(prefix, static.New(".", static.Config{
					FS:     sub,
					Browse: shouldBrowse,
				}))
			} else {
				app.Value.Use(prefix, static.New(fpath, static.Config{
					Browse: shouldBrowse,
				}))
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
	{
		Name: "_helmet",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("helmet", len(args), 2, "")
			}
			app, errObj := getFiberAppArg("helmet", args[0])
			if errObj != nil {
				return errObj
			}
			if args[1].Type() != MAP_OBJ {
				return newPositionalTypeError("helmet", 2, MAP_OBJ, args[1].Type())
			}
			cfg, errObj := helmetConfigFromMap(args[1].(*Map))
			if errObj != nil {
				return errObj
			}
			app.Use(helmet.New(cfg))
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`helmet` adds security related http headers to every response from the given server. options is a map using the snake_case names of the fiber helmet config",
			signature:   "helmet(server: GoObj[*fiber.App], options: map[str]any={}) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "helmet(s, {x_frame_options: 'DENY', hsts_max_age: 31536000}) => null",
		}.String(),
	},
	{
		Name: "_compress",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("compress", len(args), 2, "")
			}
			app, errObj := getFiberAppArg("compress", args[0])
			if errObj != nil {
				return errObj
			}
			if args[1].Type() != INTEGER_OBJ {
				return newPositionalTypeError("compress", 2, INTEGER_OBJ, args[1].Type())
			}
			level := int(args[1].(*Integer).Value)
			if level < int(compress.LevelDisabled) || level > int(compress.LevelBestCompression) {
				return newError("`compress` error: level must be between %d and %d. got=%d", compress.LevelDisabled, compress.LevelBestCompression, level)
			}
			app.Use(compress.New(compress.Config{Level: compress.Level(level)}))
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`compress` compresses http responses. level is -1 (disabled), 0 (default), 1 (best speed), or 2 (best compression)",
			signature:   "compress(server: GoObj[*fiber.App], level: int=0) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "compress(s, 1) => null",
		}.String(),
	},
	{
		Name: "_limiter",
		Fun: func(args ...Object) Object {
			if len(args) != 7 {
				return newInvalidArgCountError("limiter", len(args), 7, "")
			}
			app, errObj := getFiberAppArg("limiter", args[0])
			if errObj != nil {
				return errObj
			}
			if args[1].Type() != INTEGER_OBJ {
				return newPositionalTypeError("limiter", 2, INTEGER_OBJ, args[1].Type())
			}
			if args[2].Type() != INTEGER_OBJ {
				return newPositionalTypeError("limiter", 3, INTEGER_OBJ, args[2].Type())
			}
			var skipFailedRequests, skipSuccessfulRequests, disableHeaders, disableValueRedaction bool
			for i, dst := range []*bool{&skipFailedRequests, &skipSuccessfulRequests, &disableHeaders, &disableValueRedaction} {
				arg := args[i+3]
				if arg.Type() != BOOLEAN_OBJ {
					return newPositionalTypeError("limiter", i+4, BOOLEAN_OBJ, arg.Type())
				}
				*dst = arg.(*Boolean).Value
			}
			max := int(args[1].(*Integer).Value)
			if max <= 0 {
				return newError("`limiter` error: max must be > 0. got=%d", max)
			}
			expiration := int(args[2].(*Integer).Value)
			if expiration <= 0 {
				return newError("`limiter` error: expiration must be > 0 seconds. got=%d", expiration)
			}
			app.Use(limiter.New(limiter.Config{
				Max:                    max,
				Expiration:             time.Duration(expiration) * time.Second,
				SkipFailedRequests:     skipFailedRequests,
				SkipSuccessfulRequests: skipSuccessfulRequests,
				DisableHeaders:         disableHeaders,
				DisableValueRedaction:  disableValueRedaction,
			}))
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`limiter` limits requests per client. max is the request count allowed during expiration seconds",
			signature:   "limiter(server: GoObj[*fiber.App], max: int=5, expiration: int=60, skip_failed_requests: bool=false, skip_successful_requests: bool=false, disable_headers: bool=false, disable_value_redaction: bool=false) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "limiter(s, 100, 60) => null",
		}.String(),
	},
	{
		Name: "_encrypt_cookie",
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("encrypt_cookie", len(args), 3, "")
			}
			app, errObj := getFiberAppArg("encrypt_cookie", args[0])
			if errObj != nil {
				return errObj
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("encrypt_cookie", 2, STRING_OBJ, args[1].Type())
			}
			key := args[1].(*Stringo).Value
			if err := validateCookieEncryptionKey(key); err != nil {
				return newError("`encrypt_cookie` error: %s", err.Error())
			}
			except, errObj := stringListArg("encrypt_cookie", 3, args[2])
			if errObj != nil {
				return errObj
			}
			app.Use(encryptcookie.New(encryptcookie.Config{
				Key:    key,
				Except: except,
			}))
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`encrypt_cookie` encrypts response cookies and decrypts request cookies using an AES key. except is a list of cookie names that should not be encrypted",
			signature:   "encrypt_cookie(server: GoObj[*fiber.App], key: str, except: list[str]=[]) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "encrypt_cookie(s, generate_cookie_key()) => null",
		}.String(),
	},
	{
		Name: "_generate_cookie_key",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("generate_cookie_key", len(args), 1, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("generate_cookie_key", 1, INTEGER_OBJ, args[0].Type())
			}
			length := int(args[0].(*Integer).Value)
			switch length {
			case 16, 24, 32:
			default:
				return newError("`generate_cookie_key` error: length must be 16, 24, or 32. got=%d", length)
			}
			return &Stringo{Value: encryptcookie.GenerateKey(length)}
		},
		HelpStr: helpStrArgs{
			explanation: "`generate_cookie_key` returns a random base64 encoded key for `encrypt_cookie`. length must be 16, 24, or 32",
			signature:   "generate_cookie_key(length: int=32) -> str",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "generate_cookie_key() => 'base64 key'",
		}.String(),
	},
	{
		Name: "_sse_send",
		Fun: func(args ...Object) Object {
			if len(args) != 5 {
				return newInvalidArgCountError("sse_send", len(args), 5, "")
			}
			stream, ok := getSSEStream(args[0])
			if !ok {
				return newPositionalTypeErrorForGoObj("sse_send", 1, "GoObj[*sse.Stream]|map", args[0])
			}
			return sendSSE(stream, "sse_send", 2, args[1], args[2], args[3], args[4])
		},
		HelpStr: helpStrArgs{
			explanation: "`sse_send` sends a server-sent event on the given sse connection. event, id, and retry are optional",
			signature:   "sse_send(s: GoObj[*sse.Stream]|map, value: str|bytes, event: str|null=null, id: str|null=null, retry: int|null=null) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "sse_send(s, 'hello', event='message', id='1', retry=3000) => null",
		}.String(),
	},
	{
		Name: "_sse_last_event_id",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("sse_last_event_id", len(args), 1, "")
			}
			stream, ok := getSSEStream(args[0])
			if !ok {
				return newPositionalTypeErrorForGoObj("sse_last_event_id", 1, "GoObj[*sse.Stream]|map", args[0])
			}
			return sseLastEventID(stream)
		},
		HelpStr: helpStrArgs{
			explanation: "`sse_last_event_id` returns the Last-Event-ID sent by a reconnecting sse client, or an empty string",
			signature:   "sse_last_event_id(s: GoObj[*sse.Stream]|map) -> str",
			errors:      "InvalidArgCount,PositionalType",
			example:     "sse_last_event_id(s) => '42'",
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
			c, ok := args[0].(*GoObj[*websocket.Conn])
			if !ok {
				return newPositionalTypeErrorForGoObj("ws_send", 1, "*websocket.Conn", args[0])
			}
			return sendWSMessage("ws_send", c.Value, args[1])
		},
		HelpStr: helpStrArgs{
			explanation: "`ws_send` sends the given value on the websocket connection, if the value is a string the websocket message type is TextMessage, otherwise if bytes BinaryMessage",
			signature:   "ws_send(c: GoObj[*websocket.Conn], value: str|bytes) -> null",
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
			return sendWSMessage("ws_client_send", c.Value, args[1])
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
				Next: func(c fiber.Ctx) bool {
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
			signature:   "inspect(c: GoObj[*websocket.Conn]|GoObj[*ws.Conn]) -> map[str]str",
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
			signature:   "handle(server: GoObj[*fiber.App], pattern: str, fn: fun, method: str='GET') -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "handle(s, '/', fn) => null",
		}.String(),
	},
	{
		Name: "_handle_use",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`handle` puts a handler on the server for a given pattern and method, `handle_use` also can use this function if no method is provided",
			signature:   "handle(server: GoObj[*fiber.App], pattern: str, fn: fun, method: str='GET') -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "handle(s, '/', fn) => null",
		}.String(),
	},
	{
		Name: "_handle_ws",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`handle_ws` puts a websocket handler on the server for a given pattern and method",
			signature:   "handle_ws(server: GoObj[*fiber.App], pattern: str, fn: fun) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "handle_ws(s, '/ws', fn) => null",
		}.String(),
	},
	{
		Name: "_handle_sse",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`handle_sse` puts a server-sent events handler on the server for a given pattern",
			signature:   "handle_sse(server: GoObj[*fiber.App], pattern: str, fn: fun) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "handle_sse(s, '/events', fn) => null",
		}.String(),
	},
}
