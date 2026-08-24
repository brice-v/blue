package compiler

import (
	"blue/code"
	"blue/consts"
	"blue/lexer"
	"blue/lib"
	"blue/object"
	"blue/parser"
	"os"

	clone "github.com/huandu/go-clone/generic"
)

func (c *Compiler) compileCore() {
	if !c.coreCompiled {
		l := lexer.New(lib.CoreFile, consts.CORE_FILE_PATH)
		p := parser.New(l)
		program := p.ParseProgram()
		if p.HasErrors() {
			p.PrintParserErrors(os.Stderr)
			os.Exit(1)
		}
		err := c.Compile(program)
		if err != nil {
			consts.ErrorPrinter("Failed to compile core.b: %s\n", err.Error())
		}
		c.coreCompiled = true
	}
}

var _coreCompiler *Compiler = nil

func newFromCore() *Compiler {
	if _coreCompiler == nil {
		l := lexer.New(lib.CoreFile, consts.CORE_FILE_PATH)
		p := parser.New(l)
		program := p.ParseProgram()
		if p.HasErrors() {
			p.PrintParserErrors(os.Stderr)
			os.Exit(1)
		}
		constants := object.NewObjectConstants()
		symbolTable := NewSymbolTable()
		for i, v := range object.AllBuiltins[0].Builtins {
			symbolTable.DefineBuiltin(i, v.Name, 0, v.Help())
		}
		for i, v := range object.BuiltinobjsList {
			symbolTable.DefineBuiltin(i, v.Name, object.BuiltinobjsModuleIndex, v.Builtin.Help())
		}
		c := NewWithState(symbolTable, constants)
		err := c.Compile(program)
		if err != nil {
			consts.ErrorPrinter("Failed to compile core.b: %s\n", err.Error())
			os.Exit(1)
		}
		c.emit(code.OpCoreCompiled)
		// log.Printf("COMPILER: %s", c.DebugString())
		_coreCompiler = c
	}
	// Cloned (not shared) so every compilation gets isolated state: sharing
	// the cached compiler's mutable slices/maps would let one compilation's
	// symbols, constants and fold indices leak into another's.
	compilerConstants := clone.Clone(_coreCompiler.constants)
	constantFolds := clone.Clone(_coreCompiler.constantFolds)
	symbolTable := clone.Clone(_coreCompiler.symbolTable)
	scopes := clone.Clone(_coreCompiler.scopes)
	return &Compiler{
		constants:        compilerConstants,
		constantFolds:    constantFolds,
		symbolTable:      symbolTable,
		scopes:           scopes,
		scopeIndex:       0,
		ErrorTrace:       []string{},
		currentPos:       _coreCompiler.currentPos,
		forIndex:         _coreCompiler.forIndex,
		breakPos:         map[int][]int{},
		contPos:          map[int][]int{},
		inTry:            map[int]struct{}{},
		importNestLevel:  _coreCompiler.importNestLevel,
		modName:          _coreCompiler.modName,
		CompilerBasePath: _coreCompiler.CompilerBasePath,
		ValidModuleNames: _coreCompiler.ValidModuleNames,

		listSetMapCompLiteralIndex: _coreCompiler.listSetMapCompLiteralIndex,
		coreCompiled:               true,
		inMatch:                    false,

		// Per-compilation dependency tracking, always fresh: core.b is
		// embedded so no file reads belong to the cached core compiler.
		ReadFiles:  []string{},
		tokens:     clone.Clone(_coreCompiler.tokens),
		tokenFolds: clone.Clone(_coreCompiler.tokenFolds),
	}
}
