package object

import (
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestShutdownWaitsForServe(t *testing.T) {
	s := NewServer()
	errCh := make(chan error, 1)
	go func() {
		time.Sleep(50 * time.Millisecond)
		ln, err := net.Listen("tcp", "localhost:0")
		if err != nil {
			errCh <- err
			return
		}
		errCh <- s.Serve(ln)
	}()
	time.Sleep(10 * time.Millisecond)
	if err := s.Shutdown(); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}
	if err := <-errCh; err != nil && err.Error() != "http: Server closed" {
		t.Fatalf("Serve returned unexpected error: %v", err)
	}
}

func TestShutdownBeforeServeTimesOut(t *testing.T) {
	s := NewServer()
	start := time.Now()
	if err := s.Shutdown(); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}
	if elapsed := time.Since(start); elapsed >= 15*time.Second {
		t.Fatalf("Shutdown blocked for %v, expected timeout at 10s", elapsed)
	}
}

func TestSetSrvNilUnblocksShutdown(t *testing.T) {
	s := NewServer()
	go func() {
		time.Sleep(20 * time.Millisecond)
		s.setSrv(nil)
	}()
	start := time.Now()
	if err := s.Shutdown(); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}
	if elapsed := time.Since(start); elapsed >= 5*time.Second {
		t.Fatalf("Shutdown blocked for %v after setSrv(nil), expected prompt return", elapsed)
	}
}

func TestRouteParamsCaptured(t *testing.T) {
	s := NewServer()
	var name, id string
	s.Add("GET", "/hello/:name", func(c *Ctx) {
		name = c.Params("name")
		_ = c.SendString(name)
	}, false)
	s.Add("", "/pre/:id", func(c *Ctx) {
		id = c.Params("id")
		_ = c.Next()
	}, true)

	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/hello/Brice", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "Brice" {
		t.Fatalf("exact route param: got code=%d body=%q, want 200 %q", rec.Code, rec.Body.String(), "Brice")
	}

	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/pre/42", nil))
	if id != "42" {
		t.Fatalf("middleware route param: got %q, want %q", id, "42")
	}
}

func TestAddStaticFilesScoping(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<h1>hi</h1>"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "app.js"), []byte("console.log(1)"), 0644); err != nil {
		t.Fatal(err)
	}

	s := NewServer()
	apiCalled := false
	s.Add("GET", "/api", func(c *Ctx) {
		apiCalled = true
		_ = c.SendString("API_OK")
	}, false)
	s.Add("GET", "/index.html", func(c *Ctx) {
		_ = c.SendString("API_OK")
	}, false)
	addStaticFiles(s, "/static", http.Dir(dir), false)

	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api", nil))
	if !apiCalled || rec.Code != http.StatusOK || rec.Body.String() != "API_OK" {
		t.Fatalf("endpoint shadowed by static middleware: called=%v code=%d body=%q", apiCalled, rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/index.html", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "API_OK" {
		t.Fatalf("exact route /index.html shadowed by static middleware: code=%d body=%q", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/static/assets/app.js", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "console.log(1)" {
		t.Fatalf("static file under prefix: got code=%d body=%q", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/missing.html", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("path outside prefix should fall through to 404, got code=%d body=%q", rec.Code, rec.Body.String())
	}
}
