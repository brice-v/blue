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

- go1.21 required
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

## Notes

- bundler will only work with ui deps installed (on linux/mac)
  - does not work cross-platform yet for building (gh actions handles it)
  - This will soon work with static builds (no ui or gg)
- set `DISABLE_HTTP_SERVER_DEBUG` to `true` to disable http server route/welcome
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

    lex     start the lexer repl or lex the given file
            (converts the file to tokens and prints)
    parse   start the parser repl or parse the given file
            (converts the file to an inspectable AST
            without node names)
    bundle  bundle the given file into a go executable
            with the runtime included
            (bundle accepts a '-d' flag for debugging)
    eval    eval the given string
    doc     print the help strings of all publicly accesible
            functions in the given filepath or module
            note: the file/module will be evaluated to gather
            all functions - so any side effects may take place
    help    prints this help message
    version prints the current version

The default behavior for no command/arguments will start
an evaluator repl. (If given a file, the file will be 
evaluated)

Environment Variables:
DISABLE_HTTP_SERVER_DEBUG   set to true to disable the gofiber
                            http route path printing and message
BLUE_INSTALL_PATH           set to the path where the blue src is
                            installed. ie. ~/.blue/src
NO_COLOR or BLUE_NO_COLOR   set to true (or any non empty string)
                            to disable colored printing
PATH                        add blue to the path variable to access
                            it anywhere. ie. ~/.blue/bin
                            could be added to path with the blue exe
                            inside of it
```
