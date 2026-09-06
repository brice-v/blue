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

    lex      start the lexer repl or lex the given file (converts the file to tokens and prints)

    parse    start the parser repl or parse the given file (converts the file to an inspectable AST without node names)
                                                                                              
             --all-parser-errors   show all parser errors instead of stopping at the first one

    doc      print the help strings of all publicly accesible functions in the given filepath or module
                                                                                              
             note: the file/module will be compiled to gather all functions

    vm       run the given string or file through the VM (a .bluec binary image is run directly without recompiling)
                                                                                              
             --all-parser-errors   show all parser errors instead of stopping at the first one
                                                                                              
             --no-exec             do not allow executing programs or scripts
                                                                                              
             -e, e, eval           alternative ways to trigger the vm evaluation

    compile  compiles the given string or file to bytecode
                                                                              
             -o <file>             write a compiled .bluec binary image instead of printing bytecode.
                                   run it with: blue vm out.bluec, bluerun out.bluec, or bundle it into an executable
                                                                              
             --no-tokens           strip the token table from the image (smaller file, error traces lose file/line info)
                                                                              
             --all-parser-errors   show all parser errors instead of stopping at the first one

    bundle   compile the given .b file and append it to a copy of the minimal bluerun runner template,
             producing a single self-contained executable. by default the template is built
             with the local go toolchain from the installed source (see blue install)

              -o <file>             path of the bundled executable to write

             --bluerun             use a prebuilt bluerun-<GOOS>-<GOARCH> (or bluerun) template
                                   found next to this executable instead of building one

             --all-parser-errors   show all parser errors instead of stopping at the first one

    help     prints this help message

    install  install blue source and binary to ~/.local/blue (src, bin)
             on windows the default root is %USERPROFILE%\.blue

             --prefix <dir>    install under <dir> instead of the default root
             -f, --force       overwrite an existing installed source tree
             --no-src          skip source extraction and go mod download
             --no-bin          skip binary copy

             re-running install refreshes the source when the embedded
             bundle changed; blue and blues share one source tree

    version  prints the current version

The default behavior for no command/arguments will start an vm repl. (If given a file, the file will be evaluated with the vm)

Run Cache:

    running a program file (blue prog.b or blue vm prog.b) caches its compiled image
    in a __blue_cache folder next to the program so later runs skip lexing, parsing
    and compiling while the sources stay unchanged. imported module files are tracked,
    editing any of them invalidates the entry automatically. set BLUE_NO_CACHE to any
    non empty value to disable the cache entirely.

A '-' can be given in place of a file to read the program from STDIN. When no
command is given and stdin is piped or redirected, the piped program is evaluated:

    echo 'println(1 + 2)' | blue
    blue - < program.b
    cat program.b | blue vm -

Errors are printed to STDERR so that program output and errors can be separated.

Environment Variables:

BLUE_DISABLE_HTTP_SERVER_DEBUG   set to true to disable the http server route path printing and message

BLUE_INSTALL_PATH                set to the path where the blue src is installed. ie. ~/.blue/src

NO_COLOR or BLUE_NO_COLOR        set to true (or any non empty string) to disable colored printing

BLUE_NO_CACHE                    set to true (or any non empty string) to disable the run cache
                                 (__blue_cache folders next to run programs)

PATH                             add blue to the path variable to access it anywhere. ie. ~/.blue/bin could be added to path with the blue exe inside of it
`

// Run runs the cmd line parsing of arguments and kicks off blue
func Run(args ...string) {
	if os.Getenv(consts.BLUE_NO_COLOR) != "" {
		color.Disable()
	}
	installFullBuildHooks()
	arguments := args[1:]
	argc := len(arguments)
	if argc == 0 {
		// This means there was no command given so perform the default
		// behavior. If stdin is piped or redirected (not a terminal) then
		// evaluate it as a program, otherwise start a vm repl.
		if !stdinIsTerminal() {
			vmFileOrString(STDIN_ARG, true, false, false, false)
			os.Exit(0)
		}
		repl.StartVmRepl()
		os.Exit(0)
	}
	command := strings.ToLower(arguments[0])
	switch command {
	case "version", "--version", "-version":
		printVersion()
	case "help", "--help", "-h":
		printUsage()
	case "lex":
		handleLexCommand(argc, arguments)
	case "parse":
		handleParseCommand(argc, arguments)
	case "vm", "eval", "-e", "e":
		handleVmCommand(argc, arguments)
	case "compile", "-c", "c":
		handleCompileCommand(argc, arguments)
	case "doc":
		handleDocCommand(argc, arguments)
	case "install":
		handleInstallCommand(argc, arguments)
	case "bundle":
		handleBundleCommand(argc, arguments)
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
		if fpath == STDIN_ARG || isFile(fpath) {
			vmFileOrString(fpath, true, noExec, allErrors, false)
		} else {
			consts.ErrorPrinter("error: file not found: %s (run 'blue help' for usage)\n", fpath)
			os.Exit(1)
		}
	}
}

// printVersion prints the version of the executable
func printVersion() {
	fmt.Printf("blue v%s\n", consts.VERSION)
}

// printUsage prints the USAGE string
func printUsage() {
	_, _ = os.Stdout.WriteString(USAGE)
}

func handleLexCommand(argc int, arguments []string) {
	if argc == 1 {
		repl.StartLexerRepl()
	} else {
		// Check if the file exists and if so, run the lexer on it
		fpath := arguments[1]
		if fpath == STDIN_ARG || isFile(fpath) {
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
		if fpath == STDIN_ARG || isFile(fpath) {
			parseFile(fpath, allErrors)
		} else {
			consts.ErrorPrinter("`parse` command expects valid file as argument. got=%s\n", fpath)
			os.Exit(1)
		}
	}
}

func handleVmCommand(argc int, arguments []string) {
	switch argc {
	case 1:
		if !stdinIsTerminal() {
			// stdin is piped or redirected so evaluate it as a program
			vmFileOrString(STDIN_ARG, true, false, false, true)
			return
		}
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
		vmFileOrString(strToEval, strToEval == STDIN_ARG || isFile(strToEval), flagNoExec, allErrors, true)
	default:
		consts.ErrorPrinter("unexpected `vm` arguments. got=%+v\n", arguments)
		os.Exit(1)
	}
}

func handleCompileCommand(argc int, arguments []string) {
	if argc < 2 || argc > 6 {
		consts.ErrorPrinter("unexpected `compile` arguments. got=%+v\n", arguments)
		os.Exit(1)
	}
	strToEval := ""
	allErrors := false
	noTokens := false
	outPath := ""
	for i, arg := range arguments[1:] {
		switch arg {
		case "--all-parser-errors":
			allErrors = true
		case "--no-tokens":
			noTokens = true
		case "-o":
			if i+2 >= argc {
				consts.ErrorPrinter("`compile` flag -o requires a file path\n")
				os.Exit(1)
			}
			outPath = arguments[i+2]
		default:
			if arg == outPath && outPath != "" {
				continue
			}
			strToEval = arg
		}
	}
	isFpath := strToEval == STDIN_ARG || isFile(strToEval)
	if outPath == "" {
		// Keep the historical debug-print behavior when no -o is given
		compileFileOrString(strToEval, isFpath, allErrors)
		return
	}
	bc, err := compileFileOrStringToImage(strToEval, isFpath, allErrors)
	if err != nil {
		consts.ErrorPrinter("%s%s\n", consts.COMPILER_ERROR_PREFIX, err.Error())
		os.Exit(1)
	}
	saveImageFile(bc, outPath, noTokens)
}

// bundleOptions is the parsed form of `blue bundle` arguments.
type bundleOptions struct {
	fpath       string
	outPath     string
	allErrors   bool
	usePrebuilt bool
}

// parseBundleArgs parses `blue bundle` arguments. The template is built
// with the local go toolchain by default; --bluerun selects a prebuilt
// template next to the executable instead. --go-build is kept as a
// deprecated alias for the default.
func parseBundleArgs(argc int, arguments []string) (bundleOptions, error) {
	if argc < 3 || argc > 6 {
		return bundleOptions{}, fmt.Errorf("unexpected `bundle` arguments. got=%+v", arguments)
	}
	var opts bundleOptions
	sawGoBuild := false
	for i, arg := range arguments[1:] {
		switch arg {
		case "--all-parser-errors":
			opts.allErrors = true
		case "--bluerun":
			opts.usePrebuilt = true
		case "--go-build":
			sawGoBuild = true
		case "-o":
			if i+2 >= argc {
				return bundleOptions{}, fmt.Errorf("`bundle` flag -o requires a file path")
			}
			opts.outPath = arguments[i+2]
		default:
			if arg == opts.outPath && opts.outPath != "" {
				continue
			}
			opts.fpath = arg
		}
	}
	if opts.usePrebuilt && sawGoBuild {
		return bundleOptions{}, fmt.Errorf("`bundle` flags --bluerun and --go-build conflict; --go-build is the default")
	}
	if opts.fpath == "" || !isFile(opts.fpath) || opts.fpath == STDIN_ARG {
		return bundleOptions{}, fmt.Errorf("`bundle` expects a valid .b source file as argument. got=%s", opts.fpath)
	}
	if opts.outPath == "" {
		return bundleOptions{}, fmt.Errorf("`bundle` requires -o <output-executable>")
	}
	return opts, nil
}

func handleBundleCommand(argc int, arguments []string) {
	opts, err := parseBundleArgs(argc, arguments)
	if err != nil {
		consts.ErrorPrinter("%s\n", err.Error())
		os.Exit(1)
	}
	bundleProgram(opts.fpath, opts.outPath, opts.allErrors, opts.usePrebuilt)
}

func handleDocCommand(argc int, arguments []string) {
	if argc != 2 {
		consts.ErrorPrinter("unexpected `doc` arguments. got=%+v\n", arguments)
		os.Exit(1)
	}
	name := arguments[1]
	fmt.Print(getDocStringFor(name))
}
