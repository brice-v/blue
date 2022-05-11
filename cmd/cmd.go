package cmd

import (
	"blue/repl"
	"flag"
	"fmt"
	"log"
)

// Run runs the cmd line parsing of arguments and kicks off the Bee language
func Run(version string, args ...string) {
	// TODO: Handle command line better, maybe use external package

	lexFlag := flag.Bool("lex", false, "Start the lexer REPL or lex the given file path")
	parseFlag := flag.Bool("parse", false, "Start the parser REPL or parse the given file path")
	evalFlag := flag.Bool("eval", false, "Start the eval REPL or eval the given file path")
	// TODO: See if we can build a whole directory with imports (embedding it all with the interpreter and running)
	// just missing the import part but single scripts should be fine
	bundleFlag := flag.Bool("b", false, "Bundle the script into a go executable")
	versionFlag := flag.Bool("v", false, "Prints the version of "+args[0])

	flag.Parse()
	argc := len(args)
	switch {
	case argc == 2 && *versionFlag:
		fmt.Println(args[0] + " v" + version)
		return
	case argc == 2 && !(*lexFlag || *parseFlag) && isValidFileForEval():
		evalFile()
	case argc == 2 && *lexFlag:
		repl.StartLexerRepl(version)
	case argc == 2 && *parseFlag:
		repl.StartParserRepl(version)
	case argc == 2 && *evalFlag:
		repl.StartEvalRepl(version)
	case argc == 3 && isValidFile() && *lexFlag:
		lexCurrentFile()
	case argc == 3 && isValidFile() && *parseFlag:
		parseCurrentFile()
	case argc == 3 && isValidFile() && *evalFlag:
		evalCurrentFile()
	case argc == 3 && isValidFile() && *bundleFlag:
		bundleCurrentFile()
	case argc == 1:
		repl.StartEvalRepl(version)
	default:
		log.Fatal("Invalid command line options")
	}
}
