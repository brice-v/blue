package repl

import (
	"blue/vm"
	"bytes"
	"fmt"
	"io"
	"log"
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
	_, err := io.WriteString(out, helpCommandUsage)
	if err != nil {
		log.Printf("Failed to write help to repl output, error: %s", err.Error())
	}
}

func handleSaveCommand(out io.Writer, filebuf *bytes.Buffer, filename string) error {
	err := os.WriteFile(filename, filebuf.Bytes(), 0666)
	if err != nil {
		return err
	}
	_, errr := fmt.Fprintf(out, "file `%s` saved\n", filename)
	if errr != nil {
		log.Printf("Failed to write to repl output, error: %s", errr.Error())
	}
	return nil
}

func handleVmLoadCommand(out io.Writer, filebuf *bytes.Buffer, filename string, vm *vm.VM) error {
	return fmt.Errorf("vm load not yet supported")
}
