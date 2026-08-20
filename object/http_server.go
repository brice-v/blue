package object

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// segment is a single path segment in a Route pattern. Either literal or a
// blue style :name wildcard is set, never both.
type segment struct {
	literal string
	param   string
}

// Route is a single registered route or middleware on a Server.
// Middleware routes are prefix matched and run before the matched endpoint
// route. Endpoint routes are exact path matched and method restricted.
type Route struct {
	isMiddleware bool
	method       string
	pattern      string
	segments     []segment
	handler      func(*Ctx)
}

// Server is the stdlib http replacement for the fiber server. It owns the
// ordered route list and dispatches requests through a Next style chain.
type Server struct {
	mu      sync.Mutex
	srv     *http.Server
	routes  []*Route
	Network string
}

// NewServer returns an empty Server with no routes registered.
func NewServer() *Server {
	return &Server{}
}

// Serve starts the http server on the given listener.
func (s *Server) Serve(ln net.Listener) error {
	s.srv = &http.Server{Handler: s}
	return s.srv.Serve(ln)
}

// Shutdown gracefully shuts down the running server if one is active.
func (s *Server) Shutdown() error {
	if s.srv == nil {
		return nil
	}
	return s.srv.Shutdown(context.Background())
}

// Add registers a new route. isMiddleware routes are prefix matched for all
// methods, otherwise method and pattern must both match the request exactly.
func (s *Server) Add(method, pattern string, h func(*Ctx), isMiddleware bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.routes = append(s.routes, &Route{
		isMiddleware: isMiddleware,
		method:       method,
		pattern:      pattern,
		segments:     compilePattern(pattern),
		handler:      h,
	})
}

// addStaticFiles registers a file serving middleware at the given prefix. The
// prefix is stripped from the request path before the filesystem is consulted
// so routes map under the served directory. Directories are only listed when
// browse is true, otherwise requests for directories without an index file
// fall through to the rest of the handler chain.
func addStaticFiles(s *Server, prefix string, httpFS http.FileSystem, browse bool) {
	fileServer := http.FileServer(httpFS)
	if prefix != "" && prefix != "/" {
		fileServer = http.StripPrefix(prefix, fileServer)
	}
	s.Add(prefix, "", func(c *Ctx) {
		rel := c.R.URL.Path
		if prefix != "" && prefix != "/" {
			rel = strings.TrimPrefix(rel, prefix)
			if rel == "" {
				rel = "/"
			}
		}
		f, err := httpFS.Open(rel)
		if err != nil {
			c.Next()
			return
		}
		stat, err := f.Stat()
		f.Close()
		if err != nil {
			c.Next()
			return
		}
		if stat.IsDir() && !browse {
			index, err := httpFS.Open(strings.TrimSuffix(rel, "/") + "/index.html")
			if err != nil {
				c.Next()
				return
			}
			index.Close()
		}
		fileServer.ServeHTTP(c.W, c.R)
	}, true)
}

// ServeHTTP dispatches one request through the registered middleware chain,
// then the matched endpoint route, ending with a 404 handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	routes := make([]*Route, len(s.routes))
	copy(routes, s.routes)
	s.mu.Unlock()

	params := make(map[string]string)
	handlers := make([]func(*Ctx), 0, len(routes)+1)
	for _, route := range routes {
		if !route.isMiddleware {
			continue
		}
		p, ok := route.matchPrefix(r.URL.Path)
		if !ok {
			continue
		}
		for k, v := range p {
			params[k] = v
		}
		handlers = append(handlers, route.handler)
	}
	for _, route := range routes {
		if route.isMiddleware {
			continue
		}
		if route.method != "" && route.method != r.Method {
			continue
		}
		p, ok := route.matchParams(r.URL.Path)
		if !ok {
			continue
		}
		for k, v := range p {
			params[k] = v
		}
		handlers = append(handlers, route.handler)
		break
	}

	handlers = append(handlers, func(c *Ctx) {
		http.NotFound(c.W, c.R)
	})

	c := &Ctx{W: w, R: r, params: params, locals: make(map[any]any)}
	i := 0
	var next func()
	next = func() {
		if i >= len(handlers) {
			return
		}
		h := handlers[i]
		i++
		h(c)
	}
	c.next = next
	next()
}

// compilePattern converts a blue pattern string into a list of segments.
// The leading slash is dropped and each :name becomes a wildcard segment.
func compilePattern(pattern string) []segment {
	segs := make([]segment, 0)
	if pattern == "" {
		return segs
	}
	for _, p := range strings.Split(pattern, "/") {
		if p == "" {
			continue
		}
		if strings.HasPrefix(p, ":") {
			segs = append(segs, segment{param: p[1:]})
		} else {
			segs = append(segs, segment{literal: p})
		}
	}
	return segs
}

func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	return strings.TrimSuffix(p, "/")
}

// pathSegments returns the non-root path segments of a request path.
func pathSegments(p string) []string {
	cp := cleanPath(p)
	if cp == "/" {
		return nil
	}
	parts := strings.Split(cp, "/")
	return parts[1:]
}

// matchParams checks an exact path match and returns the captured params.
func (r *Route) matchParams(path string) (map[string]string, bool) {
	segs := pathSegments(path)
	if len(segs) != len(r.segments) {
		return nil, false
	}
	params := make(map[string]string)
	for i, seg := range r.segments {
		if seg.param != "" {
			params[seg.param] = segs[i]
		} else if segs[i] != seg.literal {
			return nil, false
		}
	}
	return params, true
}

// matchPrefix checks a prefix match and returns any captured params. An empty
// pattern matches every request.
func (r *Route) matchPrefix(path string) (map[string]string, bool) {
	if r.pattern == "" {
		return nil, true
	}
	segs := pathSegments(path)
	if len(segs) < len(r.segments) {
		return nil, false
	}
	params := make(map[string]string)
	for i, seg := range r.segments {
		if seg.param != "" {
			params[seg.param] = segs[i]
		} else if segs[i] != seg.literal {
			return nil, false
		}
	}
	return params, true
}

// Ctx wraps a single http request/response pair plus the state needed for the
// handler chain. It is the replacement for fiber.Ctx.
type Ctx struct {
	W          http.ResponseWriter
	R          *http.Request
	params     map[string]string
	locals     map[any]any
	body       []byte
	next       func()
	statusCode int
}

// Query returns the first value for the named query param.
func (c *Ctx) Query(name string) string {
	return c.R.URL.Query().Get(name)
}

// Cookies returns the value of the named cookie, or "" if absent.
func (c *Ctx) Cookies(name string) string {
	cookie, err := c.R.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// FormValue returns the first value for the named form field.
func (c *Ctx) FormValue(name string) string {
	return c.R.FormValue(name)
}

// Body returns the request body, cached for the lifetime of the request. The
// request body stream is replaced with a fresh reader so later reads (eg.
// form parsing) still see the full body.
func (c *Ctx) Body() []byte {
	if c.body != nil {
		return c.body
	}
	body, err := io.ReadAll(c.R.Body)
	if err != nil {
		return nil
	}
	c.body = body
	c.R.Body = io.NopCloser(bytes.NewReader(body))
	return body
}

// Get returns the value of the named request header.
func (c *Ctx) Get(key string) string {
	return c.R.Header.Get(key)
}

// GetReqHeaders returns all request headers, joining multi values with ", ".
// Host is added from the request Host field since Go keeps it there, and
// Content-Length is reconstructed for non-GET requests since Go keeps it in
// r.ContentLength.
func (c *Ctx) GetReqHeaders() map[string]string {
	headers := make(map[string]string)
	for k, v := range c.R.Header {
		headers[k] = strings.Join(v, ", ")
	}
	if c.R.Host != "" {
		headers["Host"] = c.R.Host
	}
	if c.R.Method != http.MethodGet && c.R.Method != http.MethodHead && c.R.ContentLength >= 0 {
		headers["Content-Length"] = strconv.FormatInt(c.R.ContentLength, 10)
	}
	return headers
}

// Params returns the value of the named route param.
func (c *Ctx) Params(name string) string {
	return c.params[name]
}

// Method returns the request http method.
func (c *Ctx) Method() string {
	return c.R.Method
}

// Protocol returns "http" or "https" depending on whether the connection used
// TLS, mirroring fiber's Protocol method.
func (c *Ctx) Protocol() string {
	if c.R.TLS != nil {
		return "https"
	}
	return "http"
}

// IP returns the client ip extracted from RemoteAddr.
func (c *Ctx) IP() string {
	host, _, err := net.SplitHostPort(c.R.RemoteAddr)
	if err != nil {
		return c.R.RemoteAddr
	}
	return host
}

// IsFromLocal returns true when the client ip is a loopback or private ip.
func (c *Ctx) IsFromLocal() bool {
	ip := net.ParseIP(c.IP())
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast()
}

// Secure returns true when the request arrived over TLS.
func (c *Ctx) Secure() bool {
	return c.R.TLS != nil
}

// Hostname returns the request host without any port.
func (c *Ctx) Hostname() string {
	host, _, err := net.SplitHostPort(c.R.Host)
	if err != nil {
		return c.R.Host
	}
	return host
}

// ClearCookie expires the named cookies so the client drops them.
func (c *Ctx) ClearCookie(names ...string) {
	for _, name := range names {
		http.SetCookie(c.W, &http.Cookie{
			Name:    name,
			Value:   "",
			Path:    "/",
			MaxAge:  -1,
			Expires: time.Unix(1, 0),
		})
	}
}

// Locals gets or sets a per request value keyed by the given value.
func (c *Ctx) Locals(key any, value ...any) any {
	if len(value) > 0 {
		c.locals[key] = value[0]
		return value[0]
	}
	return c.locals[key]
}

// Status stores the response status code for the next write.
func (c *Ctx) Status(code int) *Ctx {
	c.statusCode = code
	return c
}

// JSON writes the value as an application/json response.
func (c *Ctx) JSON(v any) error {
	c.W.Header().Set("Content-Type", "application/json")
	c.W.WriteHeader(c.statusCodeOr(http.StatusOK))
	return json.NewEncoder(c.W).Encode(v)
}

// Send writes the given bytes as the response body.
func (c *Ctx) Send(bs []byte) error {
	c.W.WriteHeader(c.statusCodeOr(http.StatusOK))
	_, err := c.W.Write(bs)
	return err
}

// SendString writes the given string as the response body.
func (c *Ctx) SendString(s string) error {
	if c.W.Header().Get("Content-Type") == "" {
		c.W.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
	return c.Send([]byte(s))
}

// SendStatus writes the response with the given status code. The status text
// (eg. "I'm a teapot" for 418) is written as the body when one exists, which
// matches fiber's behavior.
func (c *Ctx) SendStatus(code int) error {
	msg := http.StatusText(code)
	if msg != "" {
		if c.W.Header().Get("Content-Type") == "" {
			c.W.Header().Set("Content-Type", "text/plain; charset=utf-8")
		}
		c.W.WriteHeader(code)
		_, err := c.W.Write([]byte(msg))
		return err
	}
	c.W.WriteHeader(code)
	return nil
}

// SendFile serves the file at the given path with the correct content type.
func (c *Ctx) SendFile(path string, _ ...bool) error {
	http.ServeFile(c.W, c.R, path)
	return nil
}

// Redirect sends an http redirect to the given location.
func (c *Ctx) Redirect(location string, code int) error {
	http.Redirect(c.W, c.R, location, code)
	return nil
}

// Next invokes the next handler in the chain.
func (c *Ctx) Next() error {
	c.next()
	return nil
}

// Set sets a response header value.
func (c *Ctx) Set(key, val string) {
	c.W.Header().Set(key, val)
}

// Format writes the string with a content type negotiated from the Accept
// header, mirroring fiber. HTML responses are wrapped in <p> tags.
func (c *Ctx) Format(s string) error {
	accept := c.R.Header.Get("Accept")
	// fiber's Accepts defaults to the first preference when the header is
	// missing or a wildcard, so html wins unless the client explicitly asks
	// for another format.
	if accept == "" || strings.Contains(accept, "*/*") || strings.Contains(accept, "text/html") {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.Send([]byte("<p>" + s + "</p>"))
	}
	c.Set("Content-Type", "text/plain; charset=utf-8")
	return c.Send([]byte(s))
}

func (c *Ctx) statusCodeOr(def int) int {
	if c.statusCode != 0 {
		return c.statusCode
	}
	return def
}