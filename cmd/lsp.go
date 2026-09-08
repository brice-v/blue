package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"blue/consts"
	"blue/lsp"
)

// handleLspCommand starts a Language Server Protocol server for editor
// integration. Without --addr it speaks the protocol over stdin and stdout,
// which is how editors launch language servers.
func handleLspCommand(argc int, arguments []string) {
	opts := lsp.Options{}
	args := arguments[1:]

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--addr":
			if i+1 >= len(args) {
				consts.ErrorPrinter("error: `lsp --addr` requires a host:port value\n")
				os.Exit(1)
			}
			opts.Addr = args[i+1]
			i++
		case "--trace":
			opts.Trace = true
		default:
			consts.ErrorPrinter("unexpected `lsp` argument. got=%s\n", args[i])
			os.Exit(1)
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if opts.Addr != "" {
		fmt.Fprintf(os.Stderr, "blue language server listening on %s\n", opts.Addr)
	} else {
		fmt.Fprint(os.Stderr, "blue language server reading stdin\n")
	}

	if err := lsp.Run(ctx, opts); err != nil && ctx.Err() == nil {
		consts.ErrorPrinter("language server error: %s\n", err.Error())
		os.Exit(1)
	}
}
