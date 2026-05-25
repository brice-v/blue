package repl

import (
	"blue/vm"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

func handleVmDotCommand(line string, out io.Writer, fileBuf *bytes.Buffer, vm *vm.VM) error {
	cmdAndArg := strings.Split(line, " ")
	if len(cmdAndArg) == 1 {
		handleHelpCommand(out)
	}
	cmd := cmdAndArg[0]
	switch cmd {
	case ".save":
		return handleSaveCommand(out, fileBuf, cmdAndArg[1])
	case ".load":
		return handleVmLoadCommand(out, fileBuf, cmdAndArg[1], vm)
	}
	return nil
}

const helpCommandUsage = `.exit           exits the repl
.help           prints this message
.save <fname>   saves the successfully evaluated commands
                in the repl session to a file
.load <fname>   loads the given file into the repl session
`

func handleHelpCommand(out io.Writer) {
	io.WriteString(out, helpCommandUsage)
}

func handleSaveCommand(out io.Writer, filebuf *bytes.Buffer, filename string) error {
	err := os.WriteFile(filename, filebuf.Bytes(), 0666)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "file `%s` saved\n", filename)
	return nil
}

func handleVmLoadCommand(out io.Writer, filebuf *bytes.Buffer, filename string, vm *vm.VM) error {
	return fmt.Errorf("vm load not yet supported")
}
