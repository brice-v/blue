package srcbundle

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//go:embed blue-src.tar.gz
var archive []byte

// Available reports whether an embedded source archive is present.
func Available() bool {
	return len(archive) > 0
}

// ArchiveSHA256 returns the hex SHA-256 of the embedded source archive.
// Install records this alongside the extracted tree to detect updates.
func ArchiveSHA256() string {
	sum := sha256.Sum256(archive)
	return hex.EncodeToString(sum[:])
}

// WriteArchive writes a copy of the embedded archive to path so later
// installs can tell whether the embedded bundle changed.
func WriteArchive(path string) error {
	return os.WriteFile(path, archive, 0o644)
}

// List returns the file names contained in the embedded archive.
func List() ([]string, error) {
	if !Available() {
		return nil, io.ErrUnexpectedEOF
	}
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = gz.Close()
	}()
	var names []string
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		names = append(names, hdr.Name)
	}
	return names, nil
}

// Extract writes the embedded source tree into destDir.
func Extract(destDir string) error {
	if !Available() {
		return io.ErrUnexpectedEOF
	}
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return err
	}
	defer func() {
		_ = gz.Close()
	}()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		name := filepath.Clean(hdr.Name)
		if name == "." || strings.HasPrefix(name, "..") {
			continue
		}
		target := filepath.Join(destDir, name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		default:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				_ = f.Close()
				return err
			}
			if err := f.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}
