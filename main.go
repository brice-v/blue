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
// The ML/GPU packages must ship too: object imports ml, and the GPU backend is
// built from the vendored WebGPU binding. borncgo is listed directory by
// directory so its examples folder, whose MNIST sample data is tens of
// megabytes, stays out of the binary. The WebGPU native libraries themselves are
// not embedded; `blue install` runs `go mod download`, which fetches the
// libs-<system> modules that provide the .a files in module mode.
//
//go:embed go.mod go.sum LICENSE main.go ast bd bluec blueutil cmd code compiler consts lexer lib lsp ml object parser repl runner token util vm wasmmain borncgo/autodiff borncgo/backend borncgo/internal borncgo/loader borncgo/models borncgo/nn borncgo/onnx borncgo/optim borncgo/tensor borncgo/tokenizer
var srcTree embed.FS

func main() {
	srcbundle.SetFS(srcTree)
	if err := cmd.Run(os.Args...); err != nil {
		os.Exit(1)
	}
}
