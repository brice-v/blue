package cmd

import (
	"blue/consts"
	"blue/repl"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/gookit/color"
)

const USAGE = `blue is a tool for running blue source code

Usage:
    blue <command> [arguments]

The commands are:

    lex      start the lexer repl or lex the given file (converts the file to tokens and prints)

    parse    start the parser repl or parse the given file (converts the file to an inspectable AST without node names)
                                                                                              
             --all-parser-errors   show all parser errors instead of stopping at the first one

    bundle   bundle the given file into a go executable with the runtime included (bundle accepts a '-d' flag for debugging)
             (NOTE: currently non-functional)

    doc      print the help strings of all publicly accesible functions in the given filepath or module
                                                                                              
             note: the file/module will be compiled to gather all functions

    vm       run the given string or file through the VM
                                                                                              
             --all-parser-errors   show all parser errors instead of stopping at the first one
                                                                                              
             --no-exec             do not allow executing programs or scripts
                                                                                              
             -e, e, eval           alternative ways to trigger the vm evaluation

    compile  compiles the given string or file to bytecode
                                                                                              
             --all-parser-errors   show all parser errors instead of stopping at the first one

    help     prints this help message

    version  prints the current version

The default behavior for no command/arguments will start an vm repl. (If given a file, the file will be evaluated with the vm)

Environment Variables:

BLUE_DISABLE_HTTP_SERVER_DEBUG   set to true to disable the gofiber http route path printing and message

BLUE_INSTALL_PATH                set to the path where the blue src is installed. ie. ~/.blue/src

NO_COLOR or BLUE_NO_COLOR        set to true (or any non empty string) to disable colored printing

PATH                             add blue to the path variable to access it anywhere. ie. ~/.blue/bin could be added to path with the blue exe inside of it
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
		// an vm repl.
		repl.StartVmRepl()
		os.Exit(0)
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
	case "vm", "eval", "-e", "e":
		handleVmCommand(argc, arguments)
	case "compile", "-c", "c":
		handleCompileCommand(argc, arguments)
	case "doc":
		handleDocCommand(argc, arguments)
	default:
		// Check for flags before the filename
		fpath := ""
		noExec := false
		allErrors := false
		for _, arg := range arguments {
			if arg == "--no-exec" {
				noExec = true
			} else if arg == "--all-parser-errors" {
				allErrors = true
			} else if fpath == "" {
				fpath = arg
			}
		}
		if isFile(fpath) {
			vmFileOrString(fpath, true, noExec, allErrors)
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
		fpath := ""
		allErrors := false
		for _, arg := range arguments[1:] {
			if arg == "--all-parser-errors" {
				allErrors = true
			} else {
				fpath = arg
			}
		}
		if isFile(fpath) {
			parseFile(fpath, allErrors)
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
	if argc == 2 || argc == 3 || argc == 4 || argc == 5 || argc == 6 {
		isStatic := false
		oos := runtime.GOOS
		arch := runtime.GOARCH
		fpath := ""
		outputFileName := ""
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
			} else if strings.HasPrefix(arg, "--o=") {
				outputFileName = strings.Split(arg, "--o=")[1]
			} else {
				fpath = arg
			}
		}
		if isFile(fpath) {
			err := bundleFile(fpath, isStatic, oos, arch, outputFileName)
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

func handleVmCommand(argc int, arguments []string) {
	switch argc {
	case 1:
		repl.StartVmRepl()
	case 2, 3:
		strToEval := ""
		flagNoExec := false
		allErrors := false
		for _, arg := range arguments[1:] {
			switch arg {
			case "--no-exec":
				flagNoExec = true
			case "--all-parser-errors":
				allErrors = true
			default:
				strToEval = arg
			}
		}
		vmFileOrString(strToEval, isFile(strToEval), flagNoExec, allErrors)
	default:
		consts.ErrorPrinter("unexpected `vm` arguments. got=%+v\n", arguments)
		os.Exit(1)
	}
}

func handleCompileCommand(argc int, arguments []string) {
	if argc == 2 || argc == 3 {
		strToEval := ""
		allErrors := false
		for _, arg := range arguments[1:] {
			if arg == "--all-parser-errors" {
				allErrors = true
			} else {
				strToEval = arg
			}
		}
		compileFileOrString(strToEval, isFile(strToEval), allErrors)
	} else {
		consts.ErrorPrinter("unexpected `compile` arguments. got=%+v\n", arguments)
		os.Exit(1)
	}
}

func printNodeErrorUsageAndExit() {
	consts.ErrorPrinter("`node` incorrect usage: example: `node --name \"n1@localhost\"`\n" +
		"                                 (name here requires identifier and address separated with @)")
	os.Exit(1)
}

func handleDocCommand(argc int, arguments []string) {
	if argc != 2 {
		consts.ErrorPrinter("unexpected `doc` arguments. got=%+v\n", arguments)
		os.Exit(1)
	}
	name := arguments[1]
	fmt.Print(getDocStringFor(name))
}
