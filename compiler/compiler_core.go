package compiler

import (
	"blue/code"
	"blue/consts"
	"blue/lexer"
	"blue/lib"
	"blue/object"
	"blue/parser"
	"os"
)

func (c *Compiler) compileCore() {
	// if !c.coreCompiled {
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
	// }
}

var _coreCompiler *Compiler = nil

func newFromCore() *Compiler {
	// if _coreCompiler == nil {
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
	return c
	// }
	// return &Compiler{
	// 	constants:                  _coreCompiler.constants,
	// 	constantFolds:              _coreCompiler.constantFolds,
	// 	symbolTable:                _coreCompiler.symbolTable,
	// 	scopes:                     _coreCompiler.scopes,
	// 	scopeIndex:                 0,
	// 	ErrorTrace:                 []string{},
	// 	currentPos:                 _coreCompiler.currentPos,
	// 	Tokens:                     _coreCompiler.Tokens,
	// 	BlockNestLevel:             _coreCompiler.BlockNestLevel,
	// 	forIndex:                   _coreCompiler.forIndex,
	// 	breakPos:                   map[int][]int{},
	// 	contPos:                    map[int][]int{},
	// 	importNestLevel:            _coreCompiler.importNestLevel,
	// 	modName:                    _coreCompiler.modName,
	// 	CompilerBasePath:           _coreCompiler.CompilerBasePath,
	// 	ValidModuleNames:           _coreCompiler.ValidModuleNames,
	// 	listSetMapCompLiteralIndex: _coreCompiler.listSetMapCompLiteralIndex,
	// 	coreCompiled:               true,
	// 	inMatch:                    false,
	// }
}
