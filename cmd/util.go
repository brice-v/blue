package cmd

import (
	"blue/ast"
	"blue/bluec"
	"blue/blueutil"
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
	"errors"
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
const ImageFileExtension = ".bluec"

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
	return "", fmt.Errorf("runner template not found: looked for %s and bluerun next to %s\nbuild one with: go build -tags \"minivm,<your-flavor-tags>\" -o %s ./cmd/bluerun\nor run `blue bundle` without --bluerun to build it from source automatically", runnerTemplateName(), dir, runnerTemplateName())
}

// runnerPackageRelPath is where the minimal runner lives inside the blue
// source tree.
const runnerPackageRelPath = "cmd/bluerun"

// findBlueSourceDir locates the blue module root so the default bundle
// template build can compile the runner package no matter where blue was
// invoked from. It walks up from the working directory, then consults
// BLUE_INSTALL_PATH, then the default install roots (see blue install).
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
	for _, candidate := range defaultSourceCandidates() {
		if isDir(filepath.Join(candidate, runnerPackageRelPath)) {
			return candidate, true
		}
	}
	return "", false
}

func isDir(fpath string) bool {
	info, err := os.Stat(fpath)
	return err == nil && info.IsDir()
}

// runningBuildTags returns the -tags value the CURRENT executable was
// built with (from build info), so the bundle template build can
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

// buildRunnerWithGo shells out to the go toolchain to obtain a runner
// template. The template is built with the same flavor tags as the
// running blue binary (plus the structural minivm tag) so that bundled
// images match the bundler's fingerprint.
func buildRunnerWithGo(outPath string) error {
	sourceDir, ok := findBlueSourceDir()
	if !ok {
		return fmt.Errorf("cannot find the blue source tree (looked up from the working directory, $BLUE_INSTALL_PATH and the default install roots); run `blue install` or place a %s template next to the blue executable or run bundle from inside the blue repository", runnerTemplateName())
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

// runnerTempPath returns the temp runner template path for a bundle
// output. It is absolute so the go build, which runs inside the source
// directory, writes where the bundler reads.
func runnerTempPath(outPath string) string {
	if abs, err := filepath.Abs(outPath); err == nil {
		return abs + ".bluerun-tmp"
	}
	return outPath + ".bluerun-tmp"
}

// bundleProgram compiles source through the normal pipeline, encodes it as a
// binary image, and appends it to a copy of the minimal runner template,
// producing a single self-contained executable. By default the template is
// built with the local go toolchain; usePrebuilt selects a prebuilt
// template next to the executable instead.
func bundleProgram(sourcePath string, outPath string, allErrors bool, usePrebuilt bool) error {
	bc, err := compileFileOrStringToImage(sourcePath, true, allErrors)
	if err != nil {
		consts.ErrorPrinter("%s%s\n", consts.COMPILER_ERROR_PREFIX, err.Error())
		return err
	}
	payload, err := bluec.Encode(bc, bluec.EncodeOptions{})
	if err != nil {
		return failf("error encoding program: %w", err)
	}

	var templateBytes []byte
	if !usePrebuilt {
		tmpTemplate := runnerTempPath(outPath)
		if err := buildRunnerWithGo(tmpTemplate); err != nil {
			return failf("%s", err.Error())
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
			return failf("error bundling: %w", terr)
		}
		templateBytes, err = os.ReadFile(templatePath)
	}
	if err != nil {
		return failf("error reading runner template: %w", err)
	}

	out := make([]byte, 0, len(templateBytes)+len(payload))
	out = append(out, templateBytes...)
	out = append(out, payload...)
	if err := os.WriteFile(outPath, out, 0o755); err != nil {
		return failf("error trying to write `%s`: %w", outPath, err)
	}
	fmt.Printf("bundled %s into %s (%d bytes)\nrun it with ./%s\n", sourcePath, outPath, len(out), outPath)
	return nil
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

// failf formats an error, prints it to stderr, and returns it so commands can
// report failure without exiting the process. The returned error carries the
// same message that was printed, keeping the string in one place.
func failf(format string, args ...any) error {
	err := fmt.Errorf(format, args...)
	consts.ErrorPrinter("%s\n", err.Error())
	return err
}

// isFile checks whether fpath exists and is not a directory.
func isFile(fpath string) bool {
	info, err := os.Stat(fpath)
	return !os.IsNotExist(err) && !info.IsDir()
}

// lexFile tokenizes and lexically analyzes the given file
func lexFile(fpath string) error {
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
		return failf("`lexFile` error trying to read file `%s`: %w", fpath, err)
	}

	l := lexer.New(string(data), fname)

	for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
		fmt.Printf("%+v\n", tok)
	}
	return nil
}

// parseFile parses the given file and writes the resulting program to out
func parseFile(fpath string, allErrors bool) error {
	program, err := lexAndParse(fpath, true, allErrors)
	if err != nil {
		return err
	}
	for _, text := range []string{program.String(), "\n"} {
		if _, err := io.WriteString(out, text); err != nil {
			return fmt.Errorf("failed to write program to output: %w", err)
		}
	}
	return nil
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

// ErrProgramFailed is returned by Run when an evaluated program failed at
// runtime; the run itself already printed its error to stderr.
var ErrProgramFailed = errors.New("blue program exited with an error")

func lexAndParse(inputOrFpath string, isFpath bool, allErrors bool) (*ast.Program, error) {
	var l *lexer.Lexer
	if inputOrFpath == STDIN_ARG {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, failf("error trying to read from stdin: %w", err)
		}
		l = lexer.New(string(data), STDIN_NAME)
	} else if isFpath {
		data, err := os.ReadFile(inputOrFpath)
		if err != nil {
			return nil, failf("error trying to read file `%s`: %w", inputOrFpath, err)
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
		return nil, fmt.Errorf("%d parser error(s) found", len(p.ErrorMessages()))
	}
	return program, nil
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

func compileProgram(c *compiler.Compiler, program *ast.Program) error {
	if err := c.Compile(program); err != nil {
		errToPrint, _, _ := strings.Cut(err.Error(), "\n"+consts.INTERNAL_ERROR_PATTERN)
		consts.ErrorPrinter("%s%s\n", consts.COMPILER_ERROR_PREFIX, errToPrint)
		c.PrintStackTrace()
		return err
	}
	return nil
}

func instantiateCompiler(inputOrFpath string, isFpath bool, allErrors bool) (*compiler.Compiler, error) {
	program, err := lexAndParse(inputOrFpath, isFpath, allErrors)
	if err != nil {
		return nil, err
	}
	c := newCompiler(isFpath, inputOrFpath)
	if err := compileProgram(c, program); err != nil {
		return nil, err
	}
	return c, nil
}

func instantiateCompilerForDoc(fpath string) string {
	modName := strings.ReplaceAll(filepath.Base(fpath), ".b", "")
	program, err := lexAndParse(fpath, true, false)
	if err != nil {
		return ""
	}
	c := newCompiler(true, fpath)
	c.SetDocModName(modName)
	if err := compileProgram(c, program); err != nil {
		return ""
	}
	pubFunHelpStr := c.GetDocOrderedPublicFunctionHelpString(modName)
	return object.CreateHelpStringFromProgramTokens(modName, program.HelpStrTokens, pubFunHelpStr) + "\n"
}

func compileFileOrString(inputOrFpath string, isFpath bool, allErrors bool) error {
	c, err := instantiateCompiler(inputOrFpath, isFpath, allErrors)
	if err != nil {
		return err
	}
	offset := 0
	for i, ins := range c.Bytecode().Instructions {
		if ins == byte(code.OpCoreCompiled) {
			offset = i
		}
	}
	fmt.Print(blueutil.BytecodeDebugStringWithOffset(offset, c.Bytecode().Instructions[offset:], c.Bytecode().Constants))
	return nil
}

// compileFileOrStringToImage compiles like compileFileOrString and returns
// the merged program image ready to be encoded into a .bluec container.
func compileFileOrStringToImage(inputOrFpath string, isFpath bool, allErrors bool) (*bluec.Bytecode, error) {
	c, err := instantiateCompiler(inputOrFpath, isFpath, allErrors)
	if err != nil {
		return nil, err
	}
	bc := c.Bytecode()
	if idx, err := object.FindUnserializableConstant(bc.Constants); err != nil {
		return nil, fmt.Errorf("constant %d cannot be stored in a binary image: %w\n%s", idx, err, object.DebugDumpConstants(bc.Constants))
	}
	return bc, nil
}

// saveImageFile encodes an image and writes it to fpath.
func saveImageFile(bc *bluec.Bytecode, fpath string, noTokens bool) error {
	data, err := bluec.Encode(bc, bluec.EncodeOptions{NoTokens: noTokens})
	if err != nil {
		return failf("error encoding `%s`: %w", fpath, err)
	}
	if err := os.WriteFile(fpath, data, 0o755); err != nil {
		return failf("error trying to write file `%s`: %w", fpath, err)
	}
	return nil
}

// loadImageFile reads a .bluec container from disk. It sniffs the magic so
// files with any extension (or none) are supported.
func loadImageFile(fpath string) (*bluec.Bytecode, error) {
	data, err := os.ReadFile(fpath)
	if err != nil {
		return nil, err
	}
	return bluec.Decode(data, true)
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
	header := make([]byte, len(bluec.Magic))
	n, _ := io.ReadFull(f, header)
	return n == len(header) && bluec.SniffMagic(header[:n])
}

func vmFileOrString(inputOrFpath string, isFpath, noExec, allErrors, printResult bool) error {
	var bc *bluec.Bytecode
	if looksLikeImage(inputOrFpath) {
		img, err := loadImageFile(inputOrFpath)
		if err != nil {
			return failf("error loading binary image `%s`: %w", inputOrFpath, err)
		}
		bc = img
	} else if cached := lookupCachedProgram(inputOrFpath, allErrors); cached != nil {
		bc = cached
	} else {
		c, err := instantiateCompiler(inputOrFpath, isFpath, allErrors)
		if err != nil {
			return err
		}
		storeCachedProgram(c, inputOrFpath, allErrors)
		bc = c.Bytecode()
	}
	return runBytecode(bc, noExec, printResult)
}

// runBytecode runs a program image and handles exit-code/error semantics.
// It delegates to the shared runner package so the minimal standalone
// runner behaves identically.
func runBytecode(bc *bluec.Bytecode, noExec, printResult bool) error {
	if runner.RunBytecode(bc, noExec, printResult) != 0 {
		return ErrProgramFailed
	}
	return nil
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
