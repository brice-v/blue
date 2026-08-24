package cmd

import (
	"blue/ast"
	"blue/blueutil"
	"blue/binc"
	"blue/code"
	"blue/compiler"
	"blue/consts"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"blue/runner"
	"blue/token"
	"blue/vm"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
)

// ImageFileExtension is the conventional extension for compiled blue images.
const ImageFileExtension = ".bbc"

// runnerTemplateName returns the expected filename of the minimal runner
// template for the host platform.
func runnerTemplateName() string {
	return "bluerun-" + runtime.GOOS + "-" + runtime.GOARCH
}

// findRunnerTemplate looks for a prebuilt bluerun template next to the blue
// executable: first the platform-suffixed name, then a plain `bluerun`.
func findRunnerTemplate() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(exePath)
	for _, name := range []string{runnerTemplateName(), "bluerun"} {
		candidate := filepath.Join(dir, name)
		if isFile(candidate) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("runner template not found: looked for %s and bluerun next to %s\nbuild one with: go build -tags \"minivm,<your-flavor-tags>\" -o %s ./cmd/bluerun\nor pass --go-build to build it with the same flavor automatically", runnerTemplateName(), dir, runnerTemplateName())
}

// runnerPackageRelPath is where the minimal runner lives inside the blue
// source tree.
const runnerPackageRelPath = "cmd/bluerun"

// findBlueSourceDir locates the blue module root so the --go-build fallback
// can compile the runner package no matter where blue was invoked from. It
// walks up from the working directory, then consults BLUE_INSTALL_PATH.
func findBlueSourceDir() (string, bool) {
	if dir, err := os.Getwd(); err == nil {
		for {
			if isFile(filepath.Join(dir, "go.mod")) && isDir(filepath.Join(dir, runnerPackageRelPath)) {
				return dir, true
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	if install := os.Getenv(consts.BLUE_INSTALL_PATH); install != "" {
		if isDir(filepath.Join(install, runnerPackageRelPath)) {
			return install, true
		}
	}
	return "", false
}

func isDir(fpath string) bool {
	info, err := os.Stat(fpath)
	return err == nil && info.IsDir()
}

// runningBuildTags returns the -tags value the CURRENT executable was
// built with (from build info), so the --go-build template fallback can
// reproduce the exact same runtime flavor.
func runningBuildTags() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "-tags" {
				return setting.Value
			}
		}
	}
	return ""
}

// buildRunnerWithGo shells out to the go toolchain as a fallback way of
// obtaining a runner template. The template is built with the same flavor
// tags as the running blue binary (plus the structural minivm tag) so that
// packed images match the packer's fingerprint.
func buildRunnerWithGo(outPath string) error {
	sourceDir, ok := findBlueSourceDir()
	if !ok {
		return fmt.Errorf("cannot find the blue source tree (looked up from the working directory and $BLUE_INSTALL_PATH); place a %s template next to the blue executable or run pack from inside the blue repository", runnerTemplateName())
	}
	tags := []string{"minivm"}
	if t := runningBuildTags(); t != "" {
		tags = append(tags, strings.Split(t, ",")...)
	}
	cmd := exec.Command("go", "build", "-ldflags=-s -w", "-tags", strings.Join(tags, ","), "-o", outPath, "./"+runnerPackageRelPath)
	cmd.Dir = sourceDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build runner template with go: %w", err)
	}
	return nil
}

// packProgram compiles source through the normal pipeline, encodes it as a
// binary image, and appends it to a copy of the minimal runner template,
// producing a single self-contained executable.
func packProgram(sourcePath string, outPath string, allErrors bool, useGoBuild bool) {
	bc, err := compileFileOrStringToImage(sourcePath, true, allErrors)
	if err != nil {
		consts.ErrorPrinter("%s%s\n", consts.COMPILER_ERROR_PREFIX, err.Error())
		os.Exit(1)
	}
	payload, err := binc.Encode(bc, binc.EncodeOptions{})
	if err != nil {
		consts.ErrorPrinter("error encoding program: %s\n", err.Error())
		os.Exit(1)
	}

	var templateBytes []byte
	if useGoBuild {
		tmpTemplate := outPath + ".bluerun-tmp"
		if err := buildRunnerWithGo(tmpTemplate); err != nil {
			consts.ErrorPrinter("%s\n", err.Error())
			os.Exit(1)
		}
		defer func() {
			err := os.Remove(tmpTemplate)
			if err != nil {
				log.Printf("Failed to remove temporary bluerun template, error: %s", err.Error())
			}
		}()
		templateBytes, err = os.ReadFile(tmpTemplate)
	} else {
		templatePath, terr := findRunnerTemplate()
		if terr != nil {
			consts.ErrorPrinter("error packing: %s\n", terr.Error())
			os.Exit(1)
		}
		templateBytes, err = os.ReadFile(templatePath)
	}
	if err != nil {
		consts.ErrorPrinter("error reading runner template: %s\n", err.Error())
		os.Exit(1)
	}

	out := make([]byte, 0, len(templateBytes)+len(payload))
	out = append(out, templateBytes...)
	out = append(out, payload...)
	if err := os.WriteFile(outPath, out, 0o755); err != nil {
		consts.ErrorPrinter("error trying to write `%s`. error: %s\n", outPath, err.Error())
		os.Exit(1)
	}
	fmt.Printf("packed %s into %s (%d bytes)\nrun it with ./%s\n", sourcePath, outPath, len(out), outPath)
}

// installFullBuildHooks wires up runtime features that require the full
// lexer/parser/compiler toolchain (the `eval` keyword and builtins like
// `to_num`). Minimal VM-only builds never call this.
func installFullBuildHooks() {
	vm.EvalHook = evalSourceString
}

// evalSourceString is the production implementation of vm.EvalHook.
func evalSourceString(src string) object.Object {
	l := lexer.New(src, "<internal:string>")
	p := parser.New(l)
	prog := p.ParseProgram()
	if p.HasErrors() {
		return &object.Error{Message: fmt.Sprintf("failed to `eval` string, found '%d' parser errors", len(p.ErrorMessages()))}
	}
	c := compiler.New()
	if err := c.Compile(prog); err != nil {
		return &object.Error{Message: "compiler error in `eval` string: " + err.Error()}
	}
	vmInstance := vm.New(c.Bytecode())
	if err := vmInstance.Run(); err != nil {
		return &object.Error{Message: "vm error in `eval` string: " + err.Error()}
	}
	return vmInstance.LastPoppedStackElem()
}

// out is where normal program and command output is written
var out = os.Stdout

// isFile checks whether fpath exists and is not a directory.
func isFile(fpath string) bool {
	info, err := os.Stat(fpath)
	return !os.IsNotExist(err) && !info.IsDir()
}

// lexFile tokenizes and lexically analyzes the given file
func lexFile(fpath string) {
	var data []byte
	var err error
	fname := fpath
	if fpath == STDIN_ARG {
		data, err = io.ReadAll(os.Stdin)
		fname = STDIN_NAME
	} else {
		data, err = os.ReadFile(fpath)
	}
	if err != nil {
		consts.ErrorPrinter("`lexFile` error trying to read file `%s`. error: %s\n", fpath, err.Error())
		os.Exit(1)
	}

	l := lexer.New(string(data), fname)

	for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
		fmt.Printf("%+v\n", tok)
	}
}

// parseFile parses the given file
func parseFile(fpath string, allErrors bool) {
	program := lexAndParse(fpath, true, allErrors)
	_, err := io.WriteString(out, program.String())
	if err != nil {
		log.Printf("Failed to write string to out parameter, error: %s", err.Error())
	}
	_, err = io.WriteString(out, "\n")
	if err != nil {
		log.Printf("Failed to write string to out parameter, error: %s", err.Error())
	}
}

// STDIN_ARG is the conventional argument that means read the program from STDIN
const STDIN_ARG = "-"

// STDIN_NAME is the name reported in error traces for programs read from STDIN
const STDIN_NAME = "<stdin>"

// stdinIsTerminal reports whether stdin is attached to a terminal as
// opposed to being piped or redirected
func stdinIsTerminal() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

func lexAndParse(inputOrFpath string, isFpath bool, allErrors bool) *ast.Program {
	var l *lexer.Lexer
	if inputOrFpath == STDIN_ARG {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			consts.ErrorPrinter("error trying to read from stdin. error: %s\n", err.Error())
			os.Exit(1)
		}
		l = lexer.New(string(data), STDIN_NAME)
	} else if isFpath {
		data, err := os.ReadFile(inputOrFpath)
		if err != nil {
			consts.ErrorPrinter("error trying to read file `%s`. error: %s\n", inputOrFpath, err.Error())
			os.Exit(1)
		}
		l = lexer.New(string(data), inputOrFpath)
	} else {
		l = lexer.New(inputOrFpath, "<stdin>")
	}

	var p *parser.Parser
	if allErrors {
		p = parser.New(l)
	} else {
		p = parser.NewWithStopAfterFirst(l)
	}
	program := p.ParseProgram()
	if p.HasErrors() {
		p.PrintParserErrors(os.Stderr)
		os.Exit(1)
	}
	return program
}

func newCompiler(isFpath bool, fpath string) *compiler.Compiler {
	constants := object.NewObjectConstants()
	symbolTable := compiler.NewSymbolTable()
	for i, v := range object.AllBuiltins[0].Builtins {
		symbolTable.DefineBuiltin(i, v.Name, 0, v.Help())
	}
	for i, v := range object.BuiltinobjsList {
		symbolTable.DefineBuiltin(i, v.Name, object.BuiltinobjsModuleIndex, v.Builtin.Help())
	}
	c := compiler.NewWithStateAndCore(symbolTable, constants)
	if isFpath {
		c.CompilerBasePath = filepath.Dir(fpath)
	}
	return c
}

func compileProgram(c *compiler.Compiler, program *ast.Program) {
	if err := c.Compile(program); err != nil {
		errToPrint, _, _ := strings.Cut(err.Error(), "\n"+consts.INTERNAL_ERROR_PATTERN)
		consts.ErrorPrinter("%s%s\n", consts.COMPILER_ERROR_PREFIX, errToPrint)
		c.PrintStackTrace()
		os.Exit(1)
	}
}

func instantiateCompiler(inputOrFpath string, isFpath bool, allErrors bool) *compiler.Compiler {
	program := lexAndParse(inputOrFpath, isFpath, allErrors)
	c := newCompiler(isFpath, inputOrFpath)
	compileProgram(c, program)
	return c
}

func instantiateCompilerForDoc(fpath string) string {
	modName := strings.ReplaceAll(filepath.Base(fpath), ".b", "")
	program := lexAndParse(fpath, true, false)
	c := newCompiler(true, fpath)
	c.SetDocModName(modName)
	compileProgram(c, program)
	pubFunHelpStr := c.GetDocOrderedPublicFunctionHelpString(modName)
	return object.CreateHelpStringFromProgramTokens(modName, program.HelpStrTokens, pubFunHelpStr) + "\n"
}

func compileFileOrString(inputOrFpath string, isFpath bool, allErrors bool) {
	c := instantiateCompiler(inputOrFpath, isFpath, allErrors)
	offset := 0
	for i, ins := range c.Bytecode().Instructions {
		if ins == byte(code.OpCoreCompiled) {
			offset = i
		}
	}
	fmt.Print(blueutil.BytecodeDebugStringWithOffset(offset, c.Bytecode().Instructions[offset:], c.Bytecode().Constants))
	os.Exit(0)
}

// compileFileOrStringToImage compiles like compileFileOrString and returns
// the merged program image ready to be encoded into a .bbc container.
func compileFileOrStringToImage(inputOrFpath string, isFpath bool, allErrors bool) (*binc.Bytecode, error) {
	c := instantiateCompiler(inputOrFpath, isFpath, allErrors)
	bc := c.Bytecode()
	if idx, err := object.FindUnserializableConstant(bc.Constants); err != nil {
		return nil, fmt.Errorf("constant %d cannot be stored in a binary image: %w\n%s", idx, err, object.DebugDumpConstants(bc.Constants))
	}
	return bc, nil
}

// saveImageFile encodes an image and writes it to fpath.
func saveImageFile(bc *binc.Bytecode, fpath string, noTokens bool) {
	data, err := binc.Encode(bc, binc.EncodeOptions{NoTokens: noTokens})
	if err != nil {
		consts.ErrorPrinter("error encoding `%s`: %s\n", fpath, err.Error())
		os.Exit(1)
	}
	if err := os.WriteFile(fpath, data, 0o755); err != nil {
		consts.ErrorPrinter("error trying to write file `%s`. error: %s\n", fpath, err.Error())
		os.Exit(1)
	}
}

// loadImageFile reads a .bbc container from disk. It sniffs the magic so
// files with any extension (or none) are supported.
func loadImageFile(fpath string) (*binc.Bytecode, error) {
	data, err := os.ReadFile(fpath)
	if err != nil {
		return nil, err
	}
	return binc.Decode(data, true)
}

// looksLikeImage reports whether input should be treated as a compiled
// binary image rather than blue source.
func looksLikeImage(inputOrFpath string) bool {
	if !isFile(inputOrFpath) && inputOrFpath != STDIN_ARG {
		return false
	}
	if strings.HasSuffix(strings.ToLower(inputOrFpath), ImageFileExtension) {
		return true
	}
	f, err := os.Open(inputOrFpath)
	if err != nil {
		return false
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Printf("Failed to close file with path: %s, error: %s", inputOrFpath, err.Error())
		}
	}()
	header := make([]byte, len(binc.Magic))
	n, _ := io.ReadFull(f, header)
	return n == len(header) && binc.SniffMagic(header[:n])
}

func vmFileOrString(inputOrFpath string, isFpath, noExec, allErrors, printResult bool) {
	var bc *binc.Bytecode
	if looksLikeImage(inputOrFpath) {
		img, err := loadImageFile(inputOrFpath)
		if err != nil {
			consts.ErrorPrinter("error loading binary image `%s`:\n%s\n", inputOrFpath, err.Error())
			os.Exit(1)
		}
		bc = img
	} else {
		c := instantiateCompiler(inputOrFpath, isFpath, allErrors)
		bc = c.Bytecode()
	}
	runBytecode(bc, noExec, printResult)
}

// runBytecode runs a program image and handles exit-code/error semantics.
// It delegates to the shared runner package so the minimal standalone
// runner behaves identically.
func runBytecode(bc *binc.Bytecode, noExec, printResult bool) {
	os.Exit(runner.RunBytecode(bc, noExec, printResult))
}

func getBuiltinHelpIfExists(name string) string {
	var out bytes.Buffer
	found := false
	// Look through modules
	for _, builtins := range object.AllBuiltins {
		if builtins.Name == name {
			found = true
			fmt.Fprintf(&out, "MODULE: %s\n", name)
			for _, b := range builtins.Builtins {
				fmt.Fprintf(&out, "%s\n", b.HelpStr)
			}
		}
	}
	// Look through builtins individually
	if !found {
		for _, builtins := range object.AllBuiltins {
			for _, b := range builtins.Builtins {
				if b.Name == name || b.Name[1:] == name {
					fmt.Fprintf(&out, "%s", b.HelpStr)
				}
			}
		}
	}
	return out.String()
}

func getDocStringFor(name string) string {
	builtinHelpStr := getBuiltinHelpIfExists(name)
	if builtinHelpStr != "" {
		return builtinHelpStr
	}
	if name == "std" {
		mods := compiler.StdModuleNames()
		sort.Strings(mods)
		var out bytes.Buffer
		for i, mod := range mods {
			c := compiler.NewFromCore()
			out.WriteString(c.GetStdModuleDocString(mod))
			if i != len(mods)-1 {
				out.WriteByte('\n')
			}
		}
		return out.String()
	}
	if compiler.IsStd(name) {
		c := compiler.NewFromCore()
		return c.GetStdModuleDocString(name)
	}
	if isFile(name) {
		return instantiateCompilerForDoc(name)
	}
	return ""
}
