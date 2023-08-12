package main

import (
	"blue/cmd"
	"os"
)

func main() {
	cmd.RunAgentIfEnabled()
	cmd.Run(os.Args...)
	os.Exit(0)
}
