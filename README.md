# blue - a fun programming language

## Background

I started this project in 2020 and have been on and off adding features to it
and developing it since then. The language draws inspiration from many others
but mostly I just wanted a scripting language that was fun to use and fun to
develop.

Note: Its **not** _blazingly fast_ but that was never the point. It may be
practical to eventually compile the language to `go` which could improve its
speed?!

## Details

- recursive descent parser
- one pass (non-parallel) tokenizer
- interpreted
- 3rd party libs used liberally and appropriate licenses should be found within
  their respective vendored folder
- there are bugs!
- Fun! (imo)

## Building

- go1.25 required
  - `brew install go` or `scoop install go` or [here](https://go.dev/dl/)
- C Compiler
  - `brew install gcc` or `scoop install gcc`
- Install deps for [fyne](https://fyne.io)
- make sure no errors with `go build`
  - [had this error on mint](https://stackoverflow.com/questions/65387167/glfw-pkg-config-error-when-building-a-fyne-app)
    - added `export PKG_CONFIG_PATH=/usr/lib/x86_64-linux-gnu/pkgconfig` to
      `~/.bashrc`
  - windows has no requirements
  - tested on macos that still works on my old macbook (latest may or may not
    work)
- Install `upx` [here](https://upx.github.io/) (or other methods) to make the
  binary super small
  - small exe cmd =
    `go build -ldflags="-s -w -extldflags='-static'" && strip blue && upx blue`
- Static build now available (with no CGO) making it much easier to cross
  compile
  - See `make_release_static.*` to see how its being built and tested locally
- The `gg` builtins use raylib's RGFW backend via the `rgfw` build tag; without
  it raylib's bundled glfw collides with fyne's go-gl/glfw at link time
  (`multiple definition of _glfw...`). Set it once so all `go` commands pick it
  up automatically:
  - `go env -w GOFLAGS=-tags=rgfw` (undo with `go env -u GOFLAGS`)
  - note: this applies to every Go project on the machine; alternatively pass
    `-tags=rgfw` per command (`go build/test/vet -tags=rgfw ./...`)
  - static builds (`static` tag) don't need it but combine fine with it, tags
    from `GOFLAGS` and the command line are merged
- `govulncheck` via `go install golang.org/x/vuln/cmd/govulncheck@latest`

## Editor Integration

`blue lsp` speaks LSP 3.17 so any editor that can start a server process can use
it. Nothing extra has to be installed for it.

```
blue lsp                        # stdio, the default (the editor owns stdin/stdout)
blue lsp --addr 127.0.0.1:9257  # listen on TCP instead, handy for debugging
blue lsp --trace                # log every message to stderr
```

Stop it with `Ctrl+C` (`SIGINT`) or `SIGTERM`; both exit cleanly. Diagnostics,
log output and trace all go to stderr so the JSON-RPC channel on stdout stays
clean. `BLUE_INSTALL_PATH` matters for the server process as well: launch the editor
from an environment where it is already set, or put it in the launcher, so std module
lookups (`import math`, `import csv`, ...) resolve to your local checkout instead of
whatever was embedded at build time.

What works:

- diagnostics published on open, change and save, anchored at the offending token
  (a missing closing brace points at the line that needs the fix), re-published
  as an empty list so editors clear stale markers
- completion for keywords, buffers (`env`, `args`, ...), builtins, imported
  modules and their members after `.`, std modules already present in the source,
  and sibling `.b` files while completing an `import`
- snippets only when the client declares `snippetSupport`, otherwise plain text
  inserts are sent so no `${1:...}` leaks into editors that cannot handle it
- hover documentation above declarations, including function headers kept verbatim
  from the source (no invented signatures)
- go to definition and find references (`includeDeclaration` is honored), plus
  read/write highlights (`=` assignments, `for ... in ...` loop variables)
- document symbols nested under enclosing functions, and `workspace/symbol`
  covering open buffers plus `.b` files under the root folder

What is deliberately missing: formatting and rename. Both would need a
transformation that still compiles, which cannot be guaranteed for partial or
syntactically broken files. Only `.b` files are indexed by name since every blue
source file uses that extension.

Neovim example:

```lua
vim.filetype.add({ extension = { b = 'blue' } })
vim.api.nvim_create_autocmd('FileType', {
  pattern = 'blue',
  callback = function()
    vim.lsp.start({
      name = 'blue',
      cmd = { 'blue', 'lsp' },
      root_dir = vim.fs.dirname(vim.fs.find({ '.git', 'go.mod' }, { upward = true })[1]),
    })
  end,
})
```

### VS Code

The client lives in `editors/code` and needs no separate server install: it just
runs `blue lsp` from whatever binary you point it at.

1. Build the binary first, with any tags your machine needs:

   ```sh
   go build -o blue .
   ```

2. Package and install the extension:

   ```sh
   cd editors/code
   npm install
   npx @vscode/vsce package --allow-missing-repository
   code --install-extension blue-vscode-*.vsix
   ```

   Without packaging, run it straight from a clone in an Extension Development
   Host instead:

   ```sh
   cd editors/code && npm install
   code --extensionDevelopmentPath="$PWD"
   ```

3. Tell it where `blue` is. An absolute path is the reliable choice because the
   setting is passed through as typed, with no variable substitution:

   ```json
   {
     "blue.path": "/home/you/src/blue/blue"
   }
   ```

   Leave it empty to have the client search `PATH` in order and pick the first
   `blue` whose `help` output lists the `lsp` subcommand. That skip matters: an
   older install sitting earlier on `PATH` does not know `lsp`, treats it as a file
   name, and crash loops with "file not found: lsp". GUI launches often get a
   different PATH than a terminal, so pinning the absolute path is safest.

4. Open any folder with `.b` files. The extension registers the language id
   `blue` for that extension, so files open colored right away and the server
   starts on its own. If another extension already claimed `.b`, force it:

   ```json
   { "files.associations": { "*.b": "blue" } }
   ```

Settings and commands available after install:

- `blue.path`: executable used to launch `lsp`
- `blue.trace.server`: set to `verbose` to start the server with `--trace` so
  every protocol message lands in the "Blue Language Server" output channel
- `Blue: Restart language server` from the command palette, useful after moving
  or rebuilding the binary. Changing `blue.path` also restarts it without reloading the window

Colors come from two layers that are meant to work together. The TextMate grammar
(`editors/code/syntaxes/blue.tmLanguage.json`, scope `source.blue`) handles
comments including `### blocks ###`, every string form, numbers in hex, binary,
octal and floats, keywords, ranges (`..` and `..<`) and arrows. The server then
paints names through semantic tokens using the legend it advertises on initialize:

```
types:     comment keyword number string function variable constant parameter module property
modifiers: declaration readonly
```

That split is deliberate. Comments and strings are lexically obvious so the
grammar owns them, while whether a name is a parameter, a module member or a
read-only `val` requires the bindings only the server has.

Quick checks worth doing after install:

- type `fun helper(` in an empty buffer and confirm completion inserts a snippet
  with tab stops rather than literal `${1:...}` text
- break a file on purpose (drop a closing brace) and confirm the marker points at
  the line that needs it, then fix it and confirm markers clear
- hover `print` or `math.sqrt` where those are already imported, and hover an
  unknown name to confirm nothing is invented

If the server process dies, run `Blue: Restart language server` from the command
palette instead of reloading the window.
When something looks wrong at the protocol level, set `blue.trace.server` to
`verbose`, reproduce, and read the output channel before changing server code.

## Notes

- bundler will only work with ui deps installed (on linux/mac)
  - does not work cross-platform yet for building (gh actions handles it)
  - This will soon work with static builds (no ui or gg)
- set `BLUE_DISABLE_HTTP_SERVER_DEBUG` to `true` to disable http server route/welcome
  message printing
  - it will also prevent the stack trace from returning in http request failures
- set `BLUE_INSTALL_PATH` to the directory where `blue` is installed to
  - this is used for the bundler currently
  - if there are no files at the given path `git` will be used to clone the repo
    there once to cache it
- set `NO_COLOR` or `BLUE_NO_COLOR` to `true` to disable color printing in the
  terminal
- my `BLUE_INSTALL_PATH` is set as `export BLUE_INSTALL_PATH=~/.blue/src`
- my `blue` exe is located at `~/.blue/bin` with PATH set to
  `export PATH=$PATH:~/.blue/bin`

### Features

- builtin libs for ui, math, http, net, crypto, time, db, config, more to come!
- default args
- string interpolation
- list, set, map comprehensions
- first class functions
- return last expression
- "immutable" and mutable variables (`val` vs `var`)
- eval()
- processes (builtin on goroutine with channels)
- (x in y) - sets, lists, str, maps
- errors, assert, try-catch-finally
- match - basic sort of pattern matching
- if's are expressions

### Examples

- default args

```kotlin
fun main(name="You") {
    println('Hello #{name}!');
}

main() # "Hello You!"
main('me') # 'Hello me!' 
# also works with main(name='World')
```

- http client

```kotlin
import http

http.get("https://danluu.com/")
```

- matching (not quite pattern matching - more like switch)

```kotlin
# from core.b
fun send(obj, value) {
    return match obj {
        {t: "pid", v: _} => {
            _send(obj.v, value)
        },
        {t: "ws", v: _} => {
            import http
            http.ws_send(obj.v, value)
        },
        _ => {
            error("obj `#{obj}` is invalid type")
        },
    };
}
```

- check out files in `b_test_programs` for more

### Space Invaders Example

![SpaceInvaders](https://github.com/brice-v/assets/blob/133b60479f94302b8fb2078870a8dc738a8e4287/basic.gif)

### Usage

- Download the binary from the
  [latest release](https://github.com/brice-v/blue/releases)
  - only amd64 being built and tested
- Ensure the binary is executable
  - `chmod +x BINARY_NAME`
- For bundler
  - ensure `BLUE_INSTALL_PATH` is set to an empty dir
    - ex: `export BLUE_INSTALL_PATH=~/.blue/src`
  - `blue bundle my_prog.b` - files should all be in the same directory with 1
    file at the root level

```sh
blue is a tool for running blue source code

Usage:
    blue <command> [arguments]

The commands are:

    lex      start the lexer repl or lex the given file (converts the file to tokens and prints)

    parse    start the parser repl or parse the given file (converts the file to an inspectable AST without node names)
                                                                                              
             --all-parser-errors   show all parser errors instead of stopping at the first one

    bundle   bundle the given file into a go executable with the runtime included (bundle accepts a '-d' flag for debugging)
             (NOTE: currently non-functional)

    doc      print the help strings of all publicly accesible functions in the given filepath or module
                                                                                              
             note: the file/module will be compiled to gather all functions

    vm       run the given string or file through the VM
                                                                                              
             --all-parser-errors   show all parser errors instead of stopping at the first one
                                                                                              
             --no-exec             do not allow executing programs or scripts
                                                                                              
             -e, e, eval           alternative ways to trigger the vm evaluation

    compile  compiles the given string or file to bytecode
                                                                              
             -o <file>             write a compiled .bluec binary image instead of printing bytecode.
                                   run it with: blue vm out.bluec, bluerun out.bluec, or bundle it into an executable
                                                                              
             --no-tokens           strip the token table from the image (smaller file, error traces lose file/line info)
                                                                              
             --all-parser-errors   show all parser errors instead of stopping at the first one

    bundle   compile the given .b file and append it to a copy of the minimal bluerun runner template,
             producing a single self-contained executable. by default the template is built
             with the local go toolchain from the installed source (see blue install)
                                                                              
              -o <file>             path of the bundled executable to write
                                                                              
             --bluerun             use a prebuilt bluerun-<GOOS>-<GOARCH> (or bluerun) template
                                   found next to this executable instead of building one
                                                                              
             --all-parser-errors   show all parser errors instead of stopping at the first one

    help     prints this help message

    install  install blue source and binary to ~/.local/blue (src, bin)
             on windows the default root is %USERPROFILE%\.blue

             --prefix <dir>    install under <dir> instead of the default root
             -f, --force       overwrite an existing installed source tree
             --no-src          skip source extraction and go mod download
             --no-bin          skip binary copy

             re-running install refreshes the installed source when this binary
             carries newer source; blue and blues share one source tree

    version  prints the current version

             --full            include vcs revision and platform

The default behavior for no command/arguments will start an vm repl. (If given a file, the file will be evaluated with the vm)

Environment Variables:

BLUE_DISABLE_HTTP_SERVER_DEBUG   set to true to disable the gofiber http route path printing and message

BLUE_INSTALL_PATH                set to the path where the blue src is installed. ie. ~/.blue/src

NO_COLOR or BLUE_NO_COLOR        set to true (or any non empty string) to disable colored printing

BLUE_NO_CACHE                    set to true (or any non empty string) to disable the run cache
                                 (__blue_cache folders next to run programs)

PATH                             add blue to the path variable to access it anywhere. ie. ~/.blue/bin could be added to path with the blue exe inside of it
```

### Run cache (__blue_cache)

Running a program file (`blue prog.b` or `blue vm prog.b`) automatically caches its
compiled image in a `__blue_cache` folder created next to the program, so later runs
skip lexing, parsing and compiling entirely and only pay the VM startup cost:

```sh
blue prog.b    # first run: compiles and fills __blue_cache/prog.b.<key>.bluec
blue prog.b    # later runs: loads the cached image (much faster startup)
```

- entries are keyed by the binary's build fingerprint, the CLI path, parser flags and
  the SHA-256 of the main source; a rebuilt blue invalidates everything automatically
- imported module files are tracked too: editing any of them invalidates the entry on
  the next run
- a corrupt or stale entry is silently recompiled, caching can never break a run
- set `BLUE_NO_CACHE=1` to disable, or just delete `__blue_cache` to clear it

### Compiled programs (.bluec) and single executables

Programs can be precompiled once and reused, skipping the lexer/parser/compiler on
every subsequent run:

```sh
blue compile -o out.bluec main.b   # compile core + std + main.b into one binary image
blue vm out.bluec                  # run it with the full binary (no recompilation)
```

The `bluerun` runner is a minimal build that embeds NO lexer/parser/compiler and can
only execute `.bluec` images. It comes in two modes:

```sh
bluerun out.bluec arg1 arg2        # run a sidecar image, forwarding args to the program
```

Bundling produces a single self-contained executable by appending the compiled image
to a copy of the `bluerun` template:

```sh
blue bundle -o myapp main.b       # builds the template with go from the installed source
blue bundle --bluerun -o myapp main.b  # use a prebuilt bluerun-<GOOS>-<GOARCH> template next to blue
./myapp                          # run; all argv is forwarded to the program
```

Notes and limitations:

- images are fingerprinted: they only load into a binary built from the same source
  version, opcode set and flavor tags (`static`/`rgfw`). A mismatch prints an
  actionable error instead of misexecuting
- `eval(...)` requires the full toolchain, so it is unavailable inside `bluerun` or
  bundled executables (it returns a clear runtime error). The full `blue` binary is
  unaffected
- images compiled with `--no-tokens` are smaller but runtime error traces cannot show
  file/line pointers
