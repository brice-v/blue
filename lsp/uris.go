package lsp

import (
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
)

// uriToPath turns an LSP document URI into a local file path. It reports false
// for URIs that have no local representation, such as untitled buffers.
func uriToPath(uri string) (string, bool) {
	if uri == "" {
		return "", false
	}
	parsed, err := url.Parse(uri)
	if err != nil {
		return "", false
	}
	if parsed.Scheme != "" && parsed.Scheme != "file" {
		return "", false
	}
	path := parsed.Path
	if path == "" {
		return "", false
	}
	if runtime.GOOS == "windows" {
		// file:///C:/foo/bar.b has a leading slash that Windows paths do not want.
		if len(path) >= 3 && path[0] == '/' && path[1] >= 'A' && path[1] <= 'Z' && path[2] == ':' {
			path = path[1:]
		}
	}
	return path, true
}

// pathToURI turns a local file path into a file:// URI.
func pathToURI(path string) string {
	if path == "" {
		return ""
	}
	if p, err := url.Parse(path); err == nil && p.Scheme != "" && p.Scheme != "file" {
		// Already a URI of some kind, hand it back untouched.
		return path
	}
	slash := filepath.ToSlash(path)
	if !strings.HasPrefix(slash, "/") {
		// Relative paths stay unusable as document identifiers.
		slash = "/" + slash
	}
	u := &url.URL{Scheme: "file", Path: slash}
	return u.String()
}
