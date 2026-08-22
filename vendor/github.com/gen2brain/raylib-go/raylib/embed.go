//go:build !cgo && !raylib_no_embed && (windows || linux || darwin) && (amd64 || arm64)
// +build !cgo
// +build !raylib_no_embed
// +build windows linux darwin
// +build amd64 arm64

package rl

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	_ "embed"
	"io"
	"os"
	"path/filepath"
)

//go:embed libs/LICENSE
var embeddedLicense []byte

// extractLib extracts the embedded shared library and returns the path to it
func extractLib() string {
	if os.Getenv("RAYLIB_NO_EMBED") == "1" {
		return ""
	}

	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		panic(err)
	}

	// becomes "C:\Users\{username}\AppData\Local\github.com\gen2brain\raylib-go\6.0\"
	// or "$HOME/.cache/github.com/gen2brain/raylib-go/6.0/"
	outDir := filepath.Join(userCacheDir, "github.com", "gen2brain", "raylib-go", requiredVersion)

	// ensure directory exists
	if err := os.MkdirAll(outDir, 0755); err != nil {
		panic(err)
	}

	// write license file, if not already exists
	licenseFile := filepath.Join(outDir, "LICENSE")
	if _, err := os.Stat(licenseFile); os.IsNotExist(err) {
		if err := os.WriteFile(licenseFile, embeddedLicense, 0644); err != nil {
			panic(err)
		}
	}

	gzipReader, err := gzip.NewReader(bytes.NewReader(embeddedLib))
	if err != nil {
		panic(err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	// we do not iterate trough the .tar.gz, because it should only contain one file
	tarHeader, err := tarReader.Next()
	if err != nil {
		panic(err)
	}

	// get the name of the file inside the .tar.gz (e.g. libraylib.so.6.0.0 or raylib.dll)
	destLib := filepath.Join(outDir, tarHeader.Name)

	// write library, if not already exists
	if fileInfo, err := os.Stat(destLib); os.IsNotExist(err) {
		if file, err := os.OpenFile(destLib, os.O_RDWR|os.O_CREATE, 0644); err == nil {
			defer file.Close()
			if _, err := io.Copy(file, tarReader); err != nil {
				panic(err)
			}
			return destLib
		} else {
			panic(err)
		}
	} else {
		if fileInfo != nil && !fileInfo.IsDir() {
			return destLib
		}
	}

	return ""
}
