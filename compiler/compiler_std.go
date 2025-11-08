package compiler

import (
	"blue/ast"
	"blue/code"
	"blue/consts"
	"blue/lexer"
	"blue/lib"
	"blue/object"
	"blue/parser"
	"fmt"
	"strings"
)

type StdModFile struct {
	File          string                     // File is the actual code used for the module
	Index         int                        // Index is the builtins index in AllBuiltins
	Builtins      object.NewBuiltinSliceType // Builtins is the builtins for the std module
	HelpStr       string                     // HelpStr is the help string for the std lib program
	ParsedProgram *ast.Program
}

var _std_mods = map[string]*StdModFile{
	"http":   {File: lib.ReadStdFileToString("http.b")},
	"time":   {File: lib.ReadStdFileToString("time.b")},
	"search": {File: lib.ReadStdFileToString("search.b")},
	"db":     {File: lib.ReadStdFileToString("db.b")},
	"math":   {File: lib.ReadStdFileToString("math.b")},
	"config": {File: lib.ReadStdFileToString("config.b")},
	"crypto": {File: lib.ReadStdFileToString("crypto.b")},
	"net":    {File: lib.ReadStdFileToString("net.b")},
	"color":  {File: lib.ReadStdFileToString("color.b")},
	"csv":    {File: lib.ReadStdFileToString("csv.b")},
	"psutil": {File: lib.ReadStdFileToString("psutil.b")},
	"wasm":   {File: lib.ReadStdFileToString("wasm.b")},
	"ui":     {File: lib.ReadStdFileToString("ui-static.b")},
	"gg":     {File: lib.ReadStdFileToString("gg-static.b")},
}

func (c *Compiler) IsStd(name string) bool {
	_, ok := _std_mods[name]
	return ok
}

func (c *Compiler) CompileStdModule(name string, nodeIdentsToImport []*ast.Identifier, shouldImportAll bool) error {
	if !c.IsStd(name) {
		return fmt.Errorf("failed to compile std module: '%s' is not in std lib map", name)
	}
	fb := _std_mods[name]
	if fb.ParsedProgram == nil {
		l := lexer.New(fb.File, "<std/"+name+".b>")
		p := parser.New(l)
		fb.ParsedProgram = p.ParseProgram()
		if len(p.Errors()) != 0 {
			for _, msg := range p.Errors() {
				splitMsg := strings.Split(msg, "\n")
				firstPart := fmt.Sprintf("%s%s\n", consts.PARSER_ERROR_PREFIX, splitMsg[0])
				consts.ErrorPrinter(firstPart)
				for i, s := range splitMsg {
					if i == 0 {
						continue
					}
					fmt.Println(s)
				}
			}
			return fmt.Errorf("%sFile '%s' contains Parser Errors", consts.PARSER_ERROR_PREFIX, name)
		}
	}
	if fb.Builtins == nil {
		i, b := object.GetIndexAndBuiltinsOf(name)
		fb.Index = i
		fb.Builtins = b
	}
	for i, stdBuiltin := range fb.Builtins {
		c.symbolTable.DefineBuiltin(i, stdBuiltin.Name, fb.Index)
	}
	defer func(builtinsToRemove object.NewBuiltinSliceType) {
		for _, stdBuiltin := range builtinsToRemove {
			c.symbolTable.RemoveBuiltin(stdBuiltin.Name)
		}
	}(fb.Builtins)
	if shouldImportAll {
		// Import All acts as if everything is in the current file
		return c.Compile(fb.ParsedProgram)
	}
	if len(nodeIdentsToImport) >= 1 {
		for _, ident := range nodeIdentsToImport {
			if strings.HasPrefix(ident.Value, "_") {
				return fmt.Errorf("imports must be public to import them. failed to import %s from %s", ident.Value, name)
			}
		}
		// TODO: Handle only importing these? Or only making them accessible during compiling
		// TODO: Add test case trying to call method such as abc._hello() => this should ideally fail to compile
		// when called from the file importing abc
	}
	c.importNestLevel++
	c.modName = append(c.modName, name)
	err := c.Compile(fb.ParsedProgram)
	if err != nil {
		return err
	}
	c.modName = c.modName[:c.importNestLevel]
	c.importNestLevel--
	c.ValidModuleNames = append(c.ValidModuleNames, name)
	// So the problem now is that index operator, needs to work based off available modules
	// while compiling, if we encounter a identifier that is a module, we must pull it in
	// TODO: Figure out help string later
	literal := &object.Module{Name: name, Env: nil, HelpStr: ""}
	c.emit(code.OpConstant, c.addConstant(literal))
	symbol := c.symbolTable.Define(name, true, c.BlockNestLevel)
	switch symbol.Scope {
	case GlobalScope:
		c.emit(code.OpSetGlobalImm, symbol.Index)
	case LocalScope:
		c.emit(code.OpSetLocalImm, symbol.Index)
	}
	return nil
	// 	NewEvaluatorLock.Lock()
	// 	fb.Env = newE.env.Clone()
	// 	// TODO: See if we can cache this somehow
	// 	pubFunHelpStr := fb.Env.GetOrderedPublicFunctionHelpString()
	// 	fb.HelpStr = CreateHelpStringFromProgramTokens(name, program.HelpStrTokens, pubFunHelpStr)
	// 	NewEvaluatorLock.Unlock()
	// }
}
