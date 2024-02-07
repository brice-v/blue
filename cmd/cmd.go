package cmd

import (
	"blue/consts"
	"blue/repl"
	"blue/ws_vendor/gops"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/google/gops/agent"

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
    ps      list the 'blue' (and other gops agent) listeners
            which have commands that can be run
    help    prints this help message
    version prints the current version

The default behavior for no command/arguments will start
an evaluator repl. (If given a file, the file will be 
evaluated)

Environment Variables:
DISABLE_AGENT_LISTENER      set to true to disable gops listener from
                            running this makes the 'ps' command non
                            functional
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
	case "ps":
		handlePsCommand(argc, arguments)
	default:
		if isFile(command) {
			// Eval the file
			evalFile(command)
		} else {
			printUsage()
		}
	}
}

// RunAgentIfEnabled will run the gops agent listener if DISABLE_AGENT_LISTENER is set to false, or ”
func RunAgentIfEnabled() {
	isDisabled := os.Getenv(consts.DISABLE_AGENT_LISTENER)
	ok, err := strconv.ParseBool(isDisabled)
	if isDisabled == "" || (err == nil && !ok) {
		if err := agent.Listen(agent.Options{}); err != nil {
			consts.ErrorPrinter("`ps` listener error: %s", err.Error())
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
			consts.ErrorPrinter("`lex` command expects valid file as argument. got=%s\n", fpath)
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
			consts.ErrorPrinter("`parse` command expects valid file as argument. got=%s\n", fpath)
			os.Exit(1)
		}
	}
}

func detectAllCommandsNeeded() error {
	consts.InfoPrinter("Detecting git and go are present...\n")
	_, err := exec.LookPath("git")
	if err != nil {
		return err
	}
	_, err = exec.LookPath("go")
	if err != nil {
		return err
	}
	if _, err = exec.LookPath("upx"); err != nil {
		color.FgYellow.Println("    bundler::detectAllCommands: upx not present so packing will not happen")
	}
	if _, err = exec.LookPath("strip"); err != nil {
		color.FgYellow.Println("    bundler::detectAllCommands: strip not present so stripping will not happen")
	}
	return nil
}

func handleBundleCommand(argc int, arguments []string) {
	err := detectAllCommandsNeeded()
	if err != nil {
		consts.ErrorPrinter("`bundle` error: %s\n", err.Error())
		os.Exit(1)
	}
	if argc == 2 || argc == 3 || argc == 4 || argc == 5 {
		isStatic := false
		oos := runtime.GOOS
		arch := runtime.GOARCH
		fpath := ""
		for _, arg := range arguments[1:] {
			if strings.HasPrefix(arg, "--static") {
				isStatic = true
			} else if strings.HasPrefix(arg, "--os=") {
				newOs := strings.Split(arg, "--os=")[1]
				if newOs != oos {
					isStatic = true
				}
				oos = newOs
			} else if strings.HasPrefix(arg, "--arch=") {
				arch = strings.Split(arg, "--arch=")[1]
			} else {
				fpath = arg
			}
		}
		if isFile(fpath) {
			err := bundleFile(fpath, isStatic, oos, arch)
			if err != nil {
				consts.ErrorPrinter("`bundle` error: %s\n", err.Error())
				os.Exit(1)
			}
		} else {
			consts.ErrorPrinter("`bundle` command expects valid file as argument. got=%s\n", fpath)
			os.Exit(1)
		}
	} else {
		consts.ErrorPrinter("unexpected `bundle` arguments. got=%+v\n", arguments)
		os.Exit(1)
	}
}

func handleEvalCommand(argc int, arguments []string) {
	if argc == 2 {
		strToEval := arguments[1]
		evalString(strToEval)
	} else {
		consts.ErrorPrinter("unexpected `eval` arguments. got=%+v\n", arguments)
		os.Exit(1)
	}
}

func handleDocCommand(argc int, arguments []string) {
	if argc != 2 {
		consts.ErrorPrinter("unexpected `doc` arguments. got=%+v\n", arguments)
		os.Exit(1)
	}
	name := arguments[1]
	fmt.Print(getDocStringFor(name))
}

func handlePsCommand(argc int, arguments []string) {
	// Have to trim off 'blue ps' for cobra as well
	for i := 0; i < len(os.Args); i++ {
		if strings.Contains(os.Args[i], "blue") && os.Args[i+1] == "ps" {
			os.Args = os.Args[i+1:]
			break
		}
	}
	err := gops.Gops_main(arguments[1:]) // trim off 'ps' and pass to Gops_main
	if err != nil {
		os.Exit(1)
	}
}
