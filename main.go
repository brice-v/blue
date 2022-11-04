package main

import (
	"blue/cmd"
	"blue/consts"
	"os"
)

func main() {
	cmd.Run(consts.VERSION, os.Args...)
	os.Exit(0)
}
