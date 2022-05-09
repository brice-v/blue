package main

import (
	"blue/cmd"
	"os"
)

// VERSION is the version number of the blang repl and language
// it will be incremented as seen fit
const VERSION = "0.0.16"

func main() {
	cmd.Run(VERSION, os.Args...)
	os.Exit(0)
}
