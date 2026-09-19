// Package main provides the Born ML Framework CLI.
package main

import (
	"fmt"
	"os"
)

const version = "v0.0.1-dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("Born ML Framework %s\n", version)
		return
	}

	fmt.Println("Born ML Framework - Production-Ready ML for Go")
	fmt.Printf("Version: %s\n\n", version)
	fmt.Println("Commands:")
	fmt.Println("  version    Show version")
	fmt.Println("")
	fmt.Println("Coming soon: train, infer, convert, serve")
}
