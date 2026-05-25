package compiler

import (
	"blue/ast"
	"blue/blueutil"
	"blue/code"
	"blue/consts"
	"blue/lexer"
	"blue/lib"
	"blue/object"
	"blue/parser"
	"fmt"
	"os"
	"strings"
)

type StdModFile struct {
	File          string            // File is the actual code used for the module
	Index         int               // Index is the builtins index in AllBuiltins
	Builtins      []*object.Builtin // Builtins is the builtins for the std module
	HelpStr       string            // HelpStr is the help string for the std lib program
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

func IsStd(name string) bool {
	_, ok := _std_mods[name]
	return ok
}

func StdModuleNames() []string {
	names := make([]string, 0, len(_std_mods))
	for n := range _std_mods {
		names = append(names, n)
	}
	return names
}

func (c *Compiler) GetStdModuleDocString(name string) string {
	if err := c.CompileStdModule(name, nil, false); err != nil {
		return ""
	}
	for i := len(c.constants) - 1; i >= 0; i-- {
		if mod, ok := c.constants[i].(*object.Module); ok && mod.Name == name {
			return mod.Help() + "\n"
		}
	}
	return ""
}

func (c *Compiler) CompileStdModule(name string, nodeIdentsToImport []*ast.Identifier, shouldImportAll bool) error {
	if !IsStd(name) {
		return fmt.Errorf("failed to compile std module: '%s' is not in std lib map", name)
	}
	fb := _std_mods[name]
	if fb.ParsedProgram == nil || !blueutil.ENABLE_VM_CACHING {
		l := lexer.New(fb.File, "<std/"+name+".b>")
		p := parser.New(l)
		fb.ParsedProgram = p.ParseProgram()
		if p.HasErrors() {
			p.PrintParserErrors(os.Stdout)
			return fmt.Errorf("%sFile '%s' contains Parser Errors", consts.PARSER_ERROR_PREFIX, name)
		}
	}
	if fb.Builtins == nil || !blueutil.ENABLE_VM_CACHING {
		i, b := object.GetIndexAndBuiltinsOf(name)
		fb.Index = i
		fb.Builtins = b
	}
	for i, stdBuiltin := range fb.Builtins {
		c.symbolTable.DefineBuiltin(i, stdBuiltin.Name, fb.Index, stdBuiltin.Help())
	}
	defer func(builtinsToRemove []*object.Builtin) {
		for _, stdBuiltin := range builtinsToRemove {
			c.symbolTable.Remove(stdBuiltin.Name)
		}
	}(fb.Builtins)
	if shouldImportAll {
		// Import All acts as if everything is in the current file
		return c.Compile(fb.ParsedProgram)
	}
	checkNodeIdentsToImport := len(nodeIdentsToImport) > 0
	if checkNodeIdentsToImport {
		for _, ident := range nodeIdentsToImport {
			if strings.HasPrefix(ident.Value, "_") {
				return fmt.Errorf("imports must be public to import them. failed to import %s from %s", ident.Value, name)
			}
		}
		// TODO: Add test case trying to call method such as abc._hello() => this should ideally fail to compile
		// when called from the file importing abc
	}
	c.importNestLevel++
	c.modName = append(c.modName, name)
	err := c.Compile(fb.ParsedProgram)
	if err != nil {
		return err
	}
	if checkNodeIdentsToImport {
		for _, ident := range nodeIdentsToImport {
			err := c.symbolTable.UpdateName(fmt.Sprintf("%s.%s", name, ident.Value), ident.Value)
			if err != nil {
				return err
			}
		}
	}
	c.modName = c.modName[:c.importNestLevel]
	c.importNestLevel--
	c.ValidModuleNames = append(c.ValidModuleNames, name)
	// So the problem now is that index operator, needs to work based off available modules
	// while compiling, if we encounter a identifier that is a module, we must pull it in
	pubFunHelpStr := c.symbolTable.GetOrderedPublicFunctionHelpString(name)
	literal := &object.Module{Name: name, Env: nil, HelpStr: object.CreateHelpStringFromProgramTokens(name, fb.ParsedProgram.HelpStrTokens, pubFunHelpStr)}
	c.emit(code.OpConstant, c.addConstant(literal))
	symbol := c.symbolTable.Define(name, true)
	switch symbol.Scope {
	case GlobalScope:
		c.emit(code.OpSetGlobalImm, symbol.Index)
	case LocalScope:
		c.emit(code.OpSetLocalImm, symbol.Index)
	}
	return nil
}
