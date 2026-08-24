package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"

	"blue/bluec"
	"blue/compiler"
	"blue/consts"
	"blue/object"
)

// The run cache stores compiled program images so repeated runs of the same
// program skip lexing, parsing and compiling entirely.
//
// Layout, created next to the program being run:
//
//	<dir>/__blue_cache/<key>.bluec   compiled program image
//	<dir>/__blue_cache/<key>.json    manifest describing how it was built
//
// <key> is derived from the running binary's identity (build fingerprint +
// blue version, so stale entries self-invalidate after a rebuild), the exact
// CLI path string, the --all-parser-errors flag and the SHA-256 of the main
// source file. Because imported user modules are NOT part of the key, every
// entry carries a manifest listing each dependency file with the content
// hash seen at compile time; a hit is only served when every dependency
// still hashes identically. Anything unexpected (missing files, bad JSON,
// decode/fingerprint failure) degrades silently to a normal uncached run.

const cacheManifestVersion = 1

// cacheDep is one dependency file recorded in a cache manifest.
type cacheDep struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

// cacheManifest describes how a cached image was produced and which files
// must remain unchanged for the entry to stay valid.
type cacheManifest struct {
	ManifestVersion int        `json:"manifest_version"`
	MainPath        string     `json:"main_path"`
	CliPath         string     `json:"cli_path"`
	AllErrors       bool       `json:"all_parser_errors"`
	BlueVersion     string     `json:"blue_version"`
	Fingerprint     string     `json:"fingerprint"`
	Deps            []cacheDep `json:"deps"`
}

// cachingEnabled reports whether the run cache may be used at all. It is on
// by default and disabled by setting BLUE_NO_CACHE to any non empty value.
func cachingEnabled() bool {
	return os.Getenv(consts.BLUE_NO_CACHE) == ""
}

// cacheDirFor returns the cache directory for a program whose absolute path
// is mainAbsPath: a __blue_cache folder next to the program file.
func cacheDirFor(mainAbsPath string) string {
	return filepath.Join(filepath.Dir(mainAbsPath), consts.CACHE_DIR_NAME)
}

// cacheEntryKey derives the cache key for one run request. The CLI path
// string participates in the key because token table entries bake it in and
// runtime error traces display it.
func cacheEntryKey(cliPath string, source []byte, allErrors bool) string {
	sourceSum := sha256.Sum256(source)
	h := sha256.New()
	h.Write([]byte(bluec.Fingerprint()))
	h.Write([]byte{0})
	h.Write([]byte(bluec.BlueVersion()))
	h.Write([]byte{0})
	h.Write([]byte(cliPath))
	h.Write([]byte{0})
	if allErrors {
		h.Write([]byte("all-parser-errors=1"))
	} else {
		h.Write([]byte("all-parser-errors=0"))
	}
	h.Write(sourceSum[:])
	return hex.EncodeToString(h.Sum(nil))
}

// hashFile returns the hex encoded SHA-256 of a file's contents.
func hashFile(fpath string) (string, error) {
	data, err := os.ReadFile(fpath)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// writeFileAtomic writes data to fpath via a temporary file and rename so a
// concurrent or crashed process never observes a half written cache entry.
func writeFileAtomic(fpath string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(fpath), filepath.Base(fpath)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, fpath)
}

// lookupCachedProgram returns the cached image for the given program file,
// or nil when the cache cannot serve this run (disabled, never compiled,
// any dependency changed, or anything failed to load). It never fails the
// run: a miss simply means the caller compiles normally.
func lookupCachedProgram(cliPath string, allErrors bool) *bluec.Bytecode {
	if !cachingEnabled() || cliPath == STDIN_ARG || !isFile(cliPath) {
		return nil
	}
	absPath, err := filepath.Abs(cliPath)
	if err != nil {
		return nil
	}
	source, err := os.ReadFile(cliPath)
	if err != nil {
		return nil
	}
	key := cacheEntryKey(cliPath, source, allErrors)
	dir := cacheDirFor(absPath)

	manData, err := os.ReadFile(filepath.Join(dir, key+".json"))
	if err != nil {
		return nil
	}
	var man cacheManifest
	if err := json.Unmarshal(manData, &man); err != nil {
		return nil
	}
	if man.ManifestVersion != cacheManifestVersion ||
		man.MainPath != absPath ||
		man.CliPath != cliPath ||
		man.AllErrors != allErrors ||
		man.Fingerprint != bluec.Fingerprint() ||
		man.BlueVersion != bluec.BlueVersion() {
		return nil
	}
	for _, dep := range man.Deps {
		sum, err := hashFile(dep.Path)
		if err != nil || sum != dep.SHA256 {
			return nil
		}
	}

	imageData, err := os.ReadFile(filepath.Join(dir, key+".bluec"))
	if err != nil {
		return nil
	}
	bc, err := bluec.Decode(imageData, true)
	if err != nil {
		return nil
	}
	return bc
}

// storeCachedProgram writes the compiled image of a just finished
// compilation to the run cache along with its dependency manifest. Every
// failure is swallowed: caching is best effort and must never turn a working
// run into an error.
func storeCachedProgram(c *compiler.Compiler, cliPath string, allErrors bool) {
	if !cachingEnabled() || cliPath == STDIN_ARG || !isFile(cliPath) {
		return
	}
	absPath, err := filepath.Abs(cliPath)
	if err != nil {
		return
	}
	source, err := os.ReadFile(cliPath)
	if err != nil {
		return
	}
	bc := c.Bytecode()
	if _, err := object.FindUnserializableConstant(bc.Constants); err != nil {
		return
	}
	imageData, err := bluec.Encode(bc, bluec.EncodeOptions{})
	if err != nil {
		return
	}

	deps := make([]cacheDep, 0, len(c.ReadFiles))
	seen := make(map[string]struct{}, len(c.ReadFiles))
	for _, dep := range c.ReadFiles {
		if _, ok := seen[dep]; ok {
			continue
		}
		seen[dep] = struct{}{}
		sum, err := hashFile(dep)
		if err != nil {
			return
		}
		deps = append(deps, cacheDep{Path: dep, SHA256: sum})
	}

	man := cacheManifest{
		ManifestVersion: cacheManifestVersion,
		MainPath:        absPath,
		CliPath:         cliPath,
		AllErrors:       allErrors,
		BlueVersion:     bluec.BlueVersion(),
		Fingerprint:     bluec.Fingerprint(),
		Deps:            deps,
	}
	manData, err := json.MarshalIndent(&man, "", "  ")
	if err != nil {
		return
	}

	dir := cacheDirFor(absPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	key := cacheEntryKey(cliPath, source, allErrors)
	if err := writeFileAtomic(filepath.Join(dir, key+".bluec"), imageData); err != nil {
		return
	}
	_ = writeFileAtomic(filepath.Join(dir, key+".json"), manData)
}
