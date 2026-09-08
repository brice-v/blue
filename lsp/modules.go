package lsp

import (
	"blue/compiler"
	"blue/object"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// maxReadFileBytes caps how much of a file is read when another module's source
// is pulled in for hover, completion or go to definition.
const maxReadFileBytes = 512 << 10 // 512 KiB

// builtinsIndex gives O(1) lookups of builtin help text by name. It is built
// once, lazily, from blue's own builtin tables so this package never keeps its
// own list of builtins.
var builtinsIndex = sync.OnceValue(func() map[string]*object.Builtin {
	index := map[string]*object.Builtin{}
	for _, group := range object.AllBuiltins {
		for _, b := range group.Builtins {
			if b == nil || b.Name == "" {
				continue
			}
			if _, exists := index[b.Name]; !exists {
				index[b.Name] = b
			}
		}
	}
	return index
})

// builtinHelp returns the help text blue prints for a builtin function.
func builtinHelp(name string) (*object.Builtin, bool) {
	b, ok := builtinsIndex()[name]
	return b, ok
}

// isBuiltinName reports whether a bare identifier is a builtin in any group.
func isBuiltinName(name string) bool {
	_, ok := builtinHelp(name)
	return ok
}

// stdModuleNames are the modules of blue's standard library.
func stdModuleNames() []string { return compiler.StdModuleNames() }

// isStdModule reports whether a name refers to a standard library module.
func isStdModule(name string) bool { return compiler.IsStd(name) }

// stdSource returns the embedded blue source of a standard library module,
// which is always available regardless of where on disk blue was invoked from.
func stdSource(name string) (*docSource, bool) {
	source, ok := compiler.StdModuleSource(name)
	if !ok {
		return nil, false
	}
	return newDocSource("<std/"+name+".b>", source), true
}

// moduleEntry is a resolved module: its source plus the index built from it.
type moduleEntry struct {
	path string // backing file path, empty for embedded std modules
	src  *docSource
	ix   *fileIndex
}

// moduleCache caches parsed module sources so repeated completion and hover
// requests inside one editing session stay cheap. Everything in it is thrown away
// whenever the owning buffer changes so nothing stale can be answered.
type moduleCache struct {
	mu      sync.Mutex
	entries map[string]*moduleEntry
}

func newModuleCache() *moduleCache {
	return &moduleCache{entries: map[string]*moduleEntry{}}
}

// std returns an embedded standard library module by name.
func (mc *moduleCache) std(name string) *moduleEntry {
	key := "std:" + name
	mc.mu.Lock()
	if entry, ok := mc.entries[key]; ok {
		mc.mu.Unlock()
		return entry
	}
	mc.mu.Unlock()

	source, ok := stdSource(name)
	if !ok {
		return nil
	}
	entry := &moduleEntry{src: source, ix: buildIndex(source)}
	entry.ix.resolveExtents()

	mc.mu.Lock()
	mc.entries[key] = entry
	mc.mu.Unlock()
	return entry
}

// get resolves a local (non std) module from its dotted import path. Open
// buffers win over what is on disk so unsaved edits to imported files are
// visible through this cache.
func (mc *moduleCache) get(baseDir string, importPath string, open func(path string) *docSource) *moduleEntry {
	path := moduleFilePath(baseDir, importPath)
	key := "file:" + path

	mc.mu.Lock()
	if entry, ok := mc.entries[key]; ok {
		mc.mu.Unlock()
		return entry
	}
	mc.mu.Unlock()

	src := open(path)
	if src == nil {
		data, err := readCapped(path)
		if err != nil {
			return nil
		}
		src = newDocSource(path, string(data))
	}

	entry := &moduleEntry{path: path, src: src, ix: buildIndex(src)}
	entry.ix.resolveExtents()

	mc.mu.Lock()
	mc.entries[key] = entry
	mc.mu.Unlock()
	return entry
}

// dropInvalid forgets every cached file backed module. It is called whenever the
// owning buffer changes because any of those answers may now be stale.
func (mc *moduleCache) dropInvalid() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	for key := range mc.entries {
		if strings.HasPrefix(key, "file:") {
			delete(mc.entries, key)
		}
	}
}

// moduleFilePath mirrors how the compiler turns a dotted import path into a file:
// dots become path separators and ".b" is appended when missing.
func moduleFilePath(baseDir string, importPath string) string {
	path := strings.ReplaceAll(importPath, ".", string(filepath.Separator))
	if !strings.HasSuffix(path, ".b") {
		path += ".b"
	}
	if baseDir != "" && baseDir != "." {
		path = filepath.Join(baseDir, path)
	}
	return path
}

// topLevelDecls returns the declarations a module exposes at its top level,
// which is what completion offers after `module.`.
func (me *moduleEntry) topLevelDecls() []*declaration {
	out := []*declaration{}
	for _, d := range me.ix.decls {
		switch d.kind {
		case declParam, declLoopVar, declCatchVar:
			continue
		}
		out = append(out, d)
	}
	return out
}

// builtinGroupMembers returns the builtins defined in one group.
func builtinGroupMembers(group string) []*object.Builtin {
	_, builtins := object.GetIndexAndBuiltinsOf(group)
	return builtins
}

// stdGroupName returns the builtin group an embedded std module is backed by,
// for example "math" for <std/math.b>. Empty when the entry is not a std module.
func stdGroupName(entry *moduleEntry) string {
	if entry == nil || entry.src == nil {
		return ""
	}
	name := entry.src.name
	if !strings.HasPrefix(name, "<std/") || !strings.HasSuffix(name, ">") {
		return ""
	}
	return strings.TrimSuffix(strings.TrimPrefix(name, "<std/"), ">")
}

func readCapped(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()
	data := make([]byte, maxReadFileBytes)
	n, _ := f.Read(data)
	return data[:n], nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
