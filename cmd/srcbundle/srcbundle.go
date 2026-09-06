package srcbundle

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// tree is the source tree to install from. Embed patterns can only name files
// below their own package directory, so the tree is declared in the module
// root's main package and handed over with SetFS before anything else runs
// (see main.go).
var tree fs.FS

// isTestScratch reports whether an entry of the tree never belongs in an
// installed source tree. An install only needs what it takes to build and run
// blue, so test files, testdata folders and files named like tests stay out.
func isTestScratch(name string) bool {
	if name == "testdata" {
		return true
	}
	return strings.Contains(name, "_test")
}

// SetFS hands this package the source tree to install from. It must be called
// before any other function in this package is used.
func SetFS(fsys fs.FS) {
	tree = fsys
}

// Available reports whether this build carries a source tree.
func Available() bool {
	return tree != nil
}

// List returns the slash separated paths of the files that Extract writes,
// in walk order.
func List() ([]string, error) {
	if !Available() {
		return nil, io.ErrUnexpectedEOF
	}
	var names []string
	err := fs.WalkDir(tree, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == "testdata" {
				return fs.SkipDir
			}
			return nil
		}
		if isTestScratch(entry.Name()) {
			return nil
		}
		names = append(names, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return names, nil
}

// TreeSHA256 returns a hex SHA-256 over the paths and contents of the files
// Extract writes. Install records it so a later run can tell whether the
// running binary carries the same source as the tree already on disk.
func TreeSHA256() string {
	if !Available() {
		return ""
	}
	names, err := List()
	if err != nil {
		return ""
	}
	hash := sha256.New()
	for _, name := range names {
		data, err := fs.ReadFile(tree, name)
		if err != nil {
			return ""
		}
		hash.Write([]byte(name))
		hash.Write(data)
	}
	return hex.EncodeToString(hash.Sum(nil))
}

// Extract writes the source tree to destDir, skipping test scratch files.
func Extract(destDir string) error {
	if !Available() {
		return io.ErrUnexpectedEOF
	}
	names, err := List()
	if err != nil {
		return err
	}
	for _, name := range names {
		target := filepath.Join(destDir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		data, err := fs.ReadFile(tree, name)
		if err != nil {
			return err
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}
