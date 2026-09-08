package main

import (
	"embed"
	"os"

	"blue/cmd"
	"blue/cmd/srcbundle"
)

// srcTree is the blue source that ships inside the executable so that
// `blue install` can lay down a tree for `blue bundle` to rebuild the minimal
// runner from (see cmd/util.go and package cmd/srcbundle).
//
// It is declared here, in the module root, because embed patterns cannot reach
// above their own package directory. Everything needed to build every flavor of
// blue and the runner is listed; vendor is left out (go resolves dependencies
// from the module cache when installing), as are tests, docs, the playground and
// tooling. No generated artifact is committed, so a binary always carries the
// exact source of the checkout it was built from.
//
//go:embed go.mod go.sum LICENSE main.go ast bd bluec blueutil cmd code compiler consts lexer lib lsp object parser repl runner token util vm wasmmain
var srcTree embed.FS

func main() {
	srcbundle.SetFS(srcTree)
	cmd.Run(os.Args...)
	os.Exit(0)
}
