package compiler

import (
	"blue/consts"
	"blue/lexer"
	"blue/lib"
	"blue/object"
	"blue/parser"
	"blue/token"
	"os"

	"github.com/huandu/go-clone"
)

func (c *Compiler) compileCore() {
	if !c.coreCompiled {
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
	if _coreCompiler == nil {
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
		c := NewWithState(symbolTable, constants)
		err := c.Compile(program)
		if err != nil {
			consts.ErrorPrinter("Failed to compile core.b: %s\n", err.Error())
			os.Exit(1)
		}
		// log.Printf("COMPILER: %s", c.DebugString())
		_coreCompiler = c
	}
	return &Compiler{
		constants:                  clone.Clone(_coreCompiler.constants).([]object.Object),
		constantFolds:              clone.Clone(_coreCompiler.constantFolds).(map[uint64]int),
		symbolTable:                clone.Clone(_coreCompiler.symbolTable).(*SymbolTable),
		scopes:                     clone.Clone(_coreCompiler.scopes).([]CompilationScope),
		scopeIndex:                 0,
		ErrorTrace:                 []string{},
		currentPos:                 _coreCompiler.currentPos,
		Tokens:                     clone.Clone(_coreCompiler.Tokens).(map[int][]token.Token),
		BlockNestLevel:             _coreCompiler.BlockNestLevel,
		forIndex:                   _coreCompiler.forIndex,
		breakPos:                   map[int][]int{},
		contPos:                    map[int][]int{},
		importNestLevel:            _coreCompiler.importNestLevel,
		modName:                    _coreCompiler.modName,
		CompilerBasePath:           _coreCompiler.CompilerBasePath,
		ValidModuleNames:           _coreCompiler.ValidModuleNames,
		listSetMapCompLiteralIndex: _coreCompiler.listSetMapCompLiteralIndex,
		coreCompiled:               true,
		inMatch:                    false,
	}
}
