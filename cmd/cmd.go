package cmd

import (
	"blue/consts"
	"blue/repl"
	"fmt"
	"os"
	"strings"

	"github.com/gookit/color"
)

const USAGE = `blue is a tool for running blue source code

Usage:
    blue <command> [arguments]

The commands are:

    lex     start the lexer repl or lex the given file
            (converts the file to tokens and prints)
    parse   start the parser repl or parse the given file
            (converts the file to an inspectable AST
            without node names)
    bundle  bundle the given file into a go executable
            with the runtime included
            (bundle accepts a '-d' flag for debugging)
    eval    eval the given string
    doc     print the help strings of all publicly accesible
            functions in the given filepath or module
            note: the file/module will be evaluated to gather
            all functions - so any side effects may take place
    help    prints this help message
    version prints the current version

The default behavior for no command/arguments will start
an evaluator repl. (If given a file, the file will be 
evaluated)

Environment Variables:
DISABLE_HTTP_SERVER_DEBUG   set to true to disable the gofiber
                            http route path printing and message
BLUE_INSTALL_PATH           set to the path where the blue src is
                            installed. ie. ~/.blue/src
NO_COLOR or BLUE_NO_COLOR   set to true (or any non empty string)
                            to disable colored printing
PATH                        add blue to the path variable to access
                            it anywhere. ie. ~/.blue/bin
                            could be added to path with the blue exe
                            inside of it
`

// Run runs the cmd line parsing of arguments and kicks off blue
func Run(args ...string) {
	if os.Getenv(consts.BLUE_NO_COLOR) != "" {
		color.Disable()
	}
	arguments := args[1:]
	argc := len(arguments)
	if argc == 0 {
		// This means there was no command given
		// so perform the default behavior of starting
		// an evaluator repl.
		repl.StartEvalRepl()
	}
	command := strings.ToLower(arguments[0])
	switch command {
	case "version":
		printVersion()
	case "help":
		printUsage()
	case "lex":
		handleLexCommand(argc, arguments)
	case "parse":
		handleParseCommand(argc, arguments)
	case "bundle":
		handleBundleCommand(argc, arguments)
	case "eval":
		handleEvalCommand(argc, arguments)
	case "doc":
		handleDocCommand(argc, arguments)
	default:
		if isFile(command) {
			// Eval the file
			evalFile(command)
		} else {
			printUsage()
		}
	}
}

// printVersion prints the version of the executable
func printVersion() {
	fmt.Printf("blue v%s\n", consts.VERSION)
}

// printUsage prints the USAGE string
func printUsage() {
	fmt.Print(USAGE)
}

func handleLexCommand(argc int, arguments []string) {
	if argc == 1 {
		repl.StartLexerRepl()
	} else {
		// Check if the file exists and if so, run the lexer on it
		fpath := arguments[1]
		if isFile(fpath) {
			lexFile(fpath)
		} else {
			fmt.Printf("`lex` command expects valid file as argument. got=%s\n", fpath)
			os.Exit(1)
		}
	}
}

func handleParseCommand(argc int, arguments []string) {
	if argc == 1 {
		repl.StartParserRepl()
	} else {
		// Check if the file exists and if so, run the parser on it
		fpath := arguments[1]
		if isFile(fpath) {
			parseFile(fpath)
		} else {
			fmt.Printf("`parse` command expects valid file as argument. got=%s\n", fpath)
			os.Exit(1)
		}
	}
}

func handleBundleCommand(argc int, arguments []string) {
	if argc == 2 {
		fpath := arguments[1]
		if isFile(fpath) {
			err := bundleFile(fpath)
			if err != nil {
				fmt.Printf("`bundle` error: %s\n", err.Error())
				os.Exit(1)
			}
		} else {
			fmt.Printf("`bundle` command expects valid file as argument. got=%s\n", fpath)
			os.Exit(1)
		}
	} else {
		fmt.Printf("unexpected `bundle` arguments. got=%+v\n", arguments)
		os.Exit(1)
	}
}

func handleEvalCommand(argc int, arguments []string) {
	if argc == 2 {
		strToEval := arguments[1]
		evalString(strToEval)
	} else {
		fmt.Printf("unexpected `eval` arguments. got=%+v\n", arguments)
		os.Exit(1)
	}
}

func handleDocCommand(argc int, arguments []string) {
	if argc != 2 {
		fmt.Printf("unexpected `doc` arguments. got=%+v\n", arguments)
	}
	name := arguments[1]
	fmt.Print(getDocStringFor(name))
}
