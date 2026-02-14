package compiler

import (
	"blue/code"
	"blue/consts"
	"blue/lexer"
	"blue/lib"
	"blue/object"
	"blue/parser"
	"blue/token"
	"blue/utils"
	"os"

	"github.com/huandu/go-clone"
)

func (c *Compiler) compileCore() {
	if !c.coreCompiled || !utils.ENABLE_VM_CACHING {
		l := lexer.New(lib.CoreFile, consts.CORE_FILE_PATH)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			for _, msg := range p.Errors() {
				consts.ErrorPrinter("ParserError in core.b: %s\n", msg)
			}
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
	if _coreCompiler == nil || !utils.ENABLE_VM_CACHING {
		l := lexer.New(lib.CoreFile, consts.CORE_FILE_PATH)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			for _, msg := range p.Errors() {
				consts.ErrorPrinter("ParserError in core.b: %s\n", msg)
			}
			os.Exit(1)
		}
		constants := object.NewObjectConstants()
		symbolTable := NewSymbolTable()
		for i, v := range object.AllBuiltins[0].Builtins {
			symbolTable.DefineBuiltin(i, v.Name, 0)
		}
		for i, v := range object.BuiltinobjsList {
			symbolTable.DefineBuiltin(i, v.Name, object.BuiltinobjsModuleIndex)
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
	var (
		compilerConstants []object.Object
		constantFolds     map[uint64]int
		symbolTable       *SymbolTable
		scopes            []CompilationScope
		tokens            map[int][]token.Token
	)
	if utils.ENABLE_VM_CACHING {
		compilerConstants = clone.Clone(_coreCompiler.constants).([]object.Object)
		constantFolds = clone.Clone(_coreCompiler.constantFolds).(map[uint64]int)
		symbolTable = clone.Clone(_coreCompiler.symbolTable).(*SymbolTable)
		scopes = clone.Clone(_coreCompiler.scopes).([]CompilationScope)
		tokens = clone.Clone(_coreCompiler.Tokens).(map[int][]token.Token)
	} else {
		compilerConstants = _coreCompiler.constants
		constantFolds = _coreCompiler.constantFolds
		symbolTable = _coreCompiler.symbolTable
		scopes = _coreCompiler.scopes
		tokens = _coreCompiler.Tokens
	}
	return &Compiler{
		constants:        compilerConstants,
		constantFolds:    constantFolds,
		symbolTable:      symbolTable,
		scopes:           scopes,
		scopeIndex:       0,
		ErrorTrace:       []string{},
		currentPos:       _coreCompiler.currentPos,
		Tokens:           tokens,
		BlockNestLevel:   _coreCompiler.BlockNestLevel,
		forIndex:         _coreCompiler.forIndex,
		breakPos:         map[int][]int{},
		contPos:          map[int][]int{},
		importNestLevel:  _coreCompiler.importNestLevel,
		modName:          _coreCompiler.modName,
		CompilerBasePath: _coreCompiler.CompilerBasePath,
		ValidModuleNames: _coreCompiler.ValidModuleNames,

		listSetMapCompLiteralIndex: _coreCompiler.listSetMapCompLiteralIndex,
		coreCompiled:               true,
		inMatch:                    false,
	}
}
