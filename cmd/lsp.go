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
// parseLspArgs extracts the LSP options from `blue lsp` arguments.
func parseLspArgs(args []string) (lsp.Options, error) {
	opts := lsp.Options{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--addr":
			if i+1 >= len(args) {
				return lsp.Options{}, fmt.Errorf("`lsp --addr` requires a host:port value")
			}
			opts.Addr = args[i+1]
			i++
		case "--trace":
			opts.Trace = true
		default:
			return lsp.Options{}, fmt.Errorf("unexpected `lsp` argument. got=%s", args[i])
		}
	}
	return opts, nil
}

func handleLspCommand(argc int, arguments []string) error {
	opts, err := parseLspArgs(arguments[1:])
	if err != nil {
		return failf("error: %s", err.Error())
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
		return err
	}
	return nil
}
